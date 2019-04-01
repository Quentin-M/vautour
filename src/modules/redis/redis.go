// Vautour - A distributed & extensible web hunter
// Copyright (C) 2019 Quentin Machu & Vautour contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package redis

import (
	"fmt"
	lib "github.com/go-redis/redis"
	"github.com/quentin-m/vautour/src/modules"
	"github.com/quentin-m/vautour/src/pkg/vautour"
	"strconv"
	"time"
	log "github.com/sirupsen/logrus"
)

const (
	processingSuffix = ":processing"
	lockSuffix = ":locks"
	cacheSuffix = ":cache"
)

type redis struct {
	lib.Options  `yaml:",inline"`
	c *lib.Client
}

func init() {
	modules.Register("redis", &redis{})
}

func (q *redis) Configure(moduleConfig *modules.ModuleConfig) error {
	// Parse parameters.
	q.Addr = "localhost:6379"

	if err := modules.ParseParams(moduleConfig.Params, &q); err != nil {
		return fmt.Errorf("invalid configuration: %v", err)
	}

	// Connect to Redis.
	q.c = lib.NewClient(&q.Options)

	// Ping it to verify the connection.
	if _, err := q.c.Ping().Result(); err !=nil {
		return err
	}
	return nil
}

func (q *redis) AddDocument(queue string, d *vautour.Document, cacheTTL time.Duration) error {
	// Cache the added document.
	if cacheTTL > 0 {
		isNew, err := q.c.ZAddNX(queue + cacheSuffix, lib.Z{Member: d.ID, Score: float64(time.Now().Add(cacheTTL).Unix())}).Result()
		if err != nil {
			return fmt.Errorf("(ZAddNX) %s", err)
		}
		if isNew == 0 {
			return vautour.ErrAlreadyExists
		}
	}

	// Add the document.
	if _, err := q.c.LPush(queue, d.JSON()).Result(); err != nil {
		q.c.ZRem(d.ID)
		return fmt.Errorf("(LPush) %s", err)
	}

	return nil
}

func (q *redis) GetDocument(queue string, ttl time.Duration) (string, *vautour.Document, func(time.Duration) error, error) {
	// Get the document, while keeping it into a processing queue.
	json, err := q.c.BRPopLPush(queue, queue + processingSuffix, 0).Result()
	if err != nil {
		return "", nil, nil, fmt.Errorf("(RPopLPush) %s", err)
	}

	// Parse the document.
	d, err := vautour.NewDocumentFromJSON(json)
	if err != nil {
		return "", nil, nil, fmt.Errorf("(NewDocumentFromJSON) %s", err)
	}

	// Lock the document.
	lock := func(ttl time.Duration) error {
		return q.c.Set(queue + lockSuffix + ":" + d.ID, "", ttl).Err()
	}
	if err := lock(ttl); err != nil {
		return "", nil, nil, fmt.Errorf("(Lock) %s", err)
	}

	return json, d, lock, nil
}

func (q *redis) ReleaseDocument(queue string, d *vautour.Document) error {
	if err := q.DeleteDocument(queue + processingSuffix, d); err != nil {
		return err
	}
	q.c.Del(queue + lockSuffix + ":" + d.ID)
	return nil
}

func (q *redis) DeleteDocument(queue string, d *vautour.Document) error {
	if c, err := q.c.LRem(queue, -1, d.JSON()).Result(); c <= 0 || err != nil {
		return fmt.Errorf("(LRem) removed: %d, expected 1, err: %v", c, err)
	}
	return nil
}

func (q *redis) Bookkeep(queues []string) {
	for _, queue := range queues {
		// Remove outdated cached document IDs.
		if c, err := q.c.ZRemRangeByScore(queue + cacheSuffix, "-inf", strconv.FormatInt(time.Now().Unix(), 10)).Result(); err != nil {
			log.WithField("role", "bookkeep").WithField("module", "redis").WithError(err).Warnf("failed to prune cached document IDs from queue %s", queue + cacheSuffix)
		} else if c > 0 {
			log.WithField("role", "bookkeep").WithField("module", "redis").Debugf("pruned %d cached document IDs from queue %s", c, queue + cacheSuffix)
		}

		// Publish back processing documents which locks have expired.
		djs, err := q.c.LRange(queue + processingSuffix, 0, -1).Result()
		if err != nil {
			log.WithField("role", "bookkeep").WithField("module", "redis").WithError(err).Warn("failed to list processing documents from queue %s", queue + processingSuffix)
			continue
		}
		for _, dj := range djs {
			d, err := vautour.NewDocumentFromJSON(dj)
			if err != nil {
				log.WithField("role", "bookkeep").WithField("module", "redis").WithError(err).Warn("failed to parse processing document from queue %s", queue + processingSuffix)
				continue
			}
			if err := q.c.Get(queue + lockSuffix + ":" + d.ID).Err(); err != nil && err != lib.Nil {
				log.WithField("role", "bookkeep").WithField("module", "redis").WithError(err).Warn("failed to lookup lock for processing document from queue %s", queue + processingSuffix)
				continue
			} else if err == nil {
				continue
			}
			if err := q.AddDocument(queue, d, 0); err != nil {
				log.WithField("role", "bookkeep").WithField("module", "redis").WithField("item_id", d.ID).WithError(err).Warn("failed to re-publish expired processing document from queue %s", queue + processingSuffix)
				continue
			}
			if err := q.c.LRem(queue, -1, dj).Err(); err != nil {
				log.WithField("role", "bookkeep").WithField("module", "redis").WithField("item_id", d.ID).WithError(err).Warn("failed to remove expired processing document from queue %s", queue + processingSuffix)
			}
		}
	}
}

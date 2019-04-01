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

package elasticsearch

import (
	"context"
	"errors"
	"fmt"
	"github.com/olivere/elastic"
	"github.com/olivere/elastic/config"
	"github.com/quentin-m/vautour/src/modules"
	"github.com/quentin-m/vautour/src/pkg/vautour"
	"sync"
	"time"
)

const (
	indexMapping = `
{
	"settings":{
		"number_of_shards": %d,
		"number_of_replicas": %d
	}
}
`
)

type elasticsearch struct{
	config.Config `yaml:",inline"`
	Timeout time.Duration

	client *elastic.Client
	clientM sync.Mutex
}

func init() {
	modules.Register("elasticsearch", &elasticsearch{})
}

func (e *elasticsearch) Configure(cfg *modules.ModuleConfig) error {
	// Default configuration.
	e.Timeout = 3 * time.Second
	e.URL = "http://127.0.0.1:9200"
	e.Index = "vautour"
	e.Replicas = 0
	e.Shards = 1
	e.Healthcheck = func() *bool { b := false; return &b }()

	// Parse parameters.
	if err := modules.ParseParams(cfg.Params, e); err != nil {
		return err
	}

	// Connect
	if err := e.connect(); err != nil {
		return err
	}

	return nil
}

func (e *elasticsearch) Send(d *vautour.Document) error {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(e.Timeout))
	defer cancel()

	if _, err := e.client.Index().Index(e.Index).Type("document").Id(d.ID).BodyJson(d.JSON()).Do(ctx); err != nil {
		return err
	}

	return nil
}

func (e *elasticsearch) connect() error {
	e.clientM.Lock()
	defer e.clientM.Unlock()

	// Connect.
	client, err := elastic.NewClientFromConfig(&e.Config);
	if err != nil {
		return err
	}
	e.client = client

	// Create index if necessary.
	return e.createIndex()
}

func (e *elasticsearch) createIndex() error {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(e.Timeout))
	defer cancel()

	indexExists, err := e.client.IndexExists(e.Index).Do(ctx)
	if err != nil {
		return err
	}
	if !indexExists {
		ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(e.Timeout))
		defer cancel()

		createIndex, err := e.client.CreateIndex(e.Index).Body(fmt.Sprintf(indexMapping, e.Shards, e.Replicas)).Do(ctx)
		if err != nil {
			return err
		}
		if !createIndex.Acknowledged {
			return errors.New("index creation not acknowledge")
		}
	}

	return nil
}
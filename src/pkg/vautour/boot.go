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

package vautour

import (
	"errors"
	"fmt"
	"github.com/coreos/pkg/timeutil"
	"github.com/quentin-m/vautour/src/modules"
	"github.com/quentin-m/vautour/src/pkg/stopper"
	log "github.com/sirupsen/logrus"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	// Durations / Periods.
	lockDuration = 3 * 60 * time.Second
	listedCacheDuration = 10 * time.Minute
	backoffSimpleDuration = 5 * time.Second
	backoffMaxDuration = 15 * time.Second
	bookkeepPeriod = 1 * time.Minute

	// Queues.
	queueDocumentsListed         = "vautour:listed"
	queueDocumentsScraped 		 = "vautour:scraped"
	queueDocumentsParsed 		 = "vautour:parsed"
)
var (
	// Durations / Periods.
	relockPeriod = time.Duration(math.Round(float64(lockDuration) * 0.9))

	// Queues.
	queues = []string{queueDocumentsListed, queueDocumentsScraped, queueDocumentsParsed}
)

func Boot(cfg Config) {
	rand.Seed(time.Now().UnixNano())
	st := stopper.NewStopper()

	// Print registered modules.
	for modS := range modules.Modules {
		log.Debugf("module %s registered", modS)
	}

	// Pre-configure modules.
	for modS, modC := range cfg.Modules {
		log.WithField("module", modS).Debug("configuring module")

		mod, err := mod(cfg, modS)
		if err != nil {
			log.WithField("module", modS).WithError(err).Fatal("failed to configure module")
		}
		if modT, ok := mod.(InputModule); ok {
			err = modT.Configure(modC)
		} else if modT, ok := mod.(ProcessorModule); ok {
			err = modT.Configure(modC)
		}  else if modT, ok := mod.(OutputModule); ok {
			err = modT.Configure(modC)
		}  else if modT, ok := mod.(QueueModule); ok {
			err = modT.Configure(modC)
		} else {
			err = fmt.Errorf("unexpected module type %T", mod)
		}
		if err != nil {
			log.WithField("module", modS).WithError(err).Fatal("failed to configure module")
		}
	}

	// Get the queue module.
	qModT, err := queueMod(cfg, cfg.Queues.Module)
	if err != nil {
		log.Fatalf("failed to find queue module: %s", err)
	}

	// Run listers.
	for _, iModS := range cfg.Inputs.Modules {
		st.Begin()
		go input(st, qModT, cfg, iModS)
	}

	// Run scrapers.
	for i := 0; i < cfg.Scrapers.Threads; i++ {
		st.Begin()
		go do(st, qModT, queueDocumentsListed, queueDocumentsScraped, log.WithField("role", "scraper"), func(d *Document) error {return scrape(cfg, d)})
	}

	// Run processors.
	for i := 0; i < cfg.Processors.Threads; i++ {
		st.Begin()
		go do(st, qModT, queueDocumentsScraped, queueDocumentsParsed, log.WithField("role", "processor"), func(d *Document) error {return process(cfg, d)})
	}

	// Run outputters.
	for i := 0; i < cfg.Outputs.Threads; i++ {
		st.Begin()
		go do(st, qModT, queueDocumentsParsed, "", log.WithField("role", "output"), func(d *Document) error {return output(cfg, d)})
	}

	// Run bookkeeping job.
	st.Begin()
	go bookkeep(st, qModT)

	// Wait for interruption and shutdown gracefully.
	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, syscall.SIGINT, syscall.SIGTERM)
	<-interrupts
	log.Info("Received interruption, gracefully stopping ...")
	st.Stop()
}

func input(st *stopper.Stopper, q QueueModule, cfg Config, iModS string) error {
	defer st.End()

	// Get the module.
	lModT, err := inputMod(cfg, iModS)
	if err != nil {
		log.WithField("role", "lister").WithField("module", iModS).Warn(err)
		return err
	}
	
	// Capture yielded documents & add them to the listed queue.
	done := make(chan bool, 1)
	ch := make(chan *Document)
	go func() {
		defer close(done)
		for {
			select {
			case d := <-ch:
				d.InputModuleName = iModS

				var backOff time.Duration
				for {
					if !st.Sleep(backOff) {
						return
					}
					if err := q.AddDocument(queueDocumentsListed, d, listedCacheDuration); err != nil && err != ErrAlreadyExists {
						backOff = timeutil.ExpBackoff(backOff, backoffMaxDuration)
						log.WithField("role", "lister").WithField("module", iModS).WithField("item_id", d.ID).WithField("duration", backOff).WithError(err).Warn("failed to add document to queue (backing off)")
						continue
					} else if err != ErrAlreadyExists {
						log.WithField("role", "lister").WithField("module", iModS).WithField("item_id", d.ID).Debug("listed new document")
					}
					break
				}
			case <-st.Chan():
				return
			}
		}
	}()

	// Run the lister, restart it if necessary.
	for st.IsRunning() {
		if err := lModT.List(st, ch); err != nil {
			log.WithField("role", "lister").WithField("module", iModS).WithError(err).Warn("input failed")
		}
	}

	// Wait for all the yielded documents to be published.
	<- done
	log.WithField("role", "lister").WithField("module", iModS).Debug("routine stopped")
	return nil
}

func scrape(cfg Config, d *Document) error {
	// Get the module.
	sModT, err := inputMod(cfg, d.InputModuleName)
	if err != nil {
		log.WithField("role", "scraper").WithField("module", d.InputModuleName).WithField("item_id", d.ID).Warn(err)
		return err
	}

	// Scrape
	if err := sModT.Scrape(d); err != nil {
		log.WithField("role", "scraper").WithField("module", d.InputModuleName).WithField("item_id", d.ID).WithError(err).Error("scraping failed")
		return errors.New("scraping failed")
	}
	log.WithField("role", "scraper").WithField("module", d.InputModuleName).WithField("item_id", d.ID).Debug("scraped document")

	return nil
}

func process(cfg Config, d *Document) error {
	for _, pModN := range cfg.Processors.Modules {
		// Get the module.
		pModT, err := processorMod(cfg, pModN)
		if err != nil {
			log.WithField("role", "processor").WithField("module", pModN).WithField("item_id", d.ID).Warn(err)
			return err
		}

		// Process the item.
		if err := pModT.Process(d); err != nil {
			log.WithField("role", "processor").WithField("module", pModN).WithField("item_id", d.ID).WithError(err).Error("processing failed")
			return errors.New("processing failed")
		}
		log.WithField("role", "processor").WithField("module", pModN).WithField("item_id", d.ID).Debug("processed document")
	}

	return nil
}

func output(cfg Config, d *Document) error {
	for _, oModN := range cfg.Outputs.Modules {
		// Get the module.
		oModT, err := outputMod(cfg, oModN)
		if err != nil {
			log.WithField("role", "output").WithField("module", oModN).WithField("item_id", d.ID).Warn(err)
			return err
		}

		// Send the item.
		if err := oModT.Send(d); err != nil {
			log.WithField("role", "output").WithField("module", oModN).WithField("item_id", d.ID).WithError(err).Error("output failed")
			return errors.New("processing failed")
		}
		log.WithField("role", "outout").WithField("module", oModN).WithField("item_id", d.ID).Debug("sent document")
	}

	return nil
}

func do(st *stopper.Stopper, q QueueModule, srcQueue, dstQueue string, logger *log.Entry, f func(d *Document) error) {
	defer st.End()

	for {
		done := make(chan bool, 1)
		var j string
		var d *Document
		var dm sync.Mutex
		var lock func(duration time.Duration) error
		var err error

		go func() {
			defer close(done)

			// Get document.
			j, d, lock, err = q.GetDocument(srcQueue, lockDuration)
			if err != nil {
				logger.WithField("item_id", d.ID).WithError(err).Warn("failed to add document to queue")
				time.Sleep(backoffSimpleDuration)
				return
			}
			dO, _ := NewDocumentFromJSON(j)

			// Run function
			if err := f(d); err != nil {
				return
			}

			// From that point, prevent the early termination of the routine until we're done.
			dm.Lock()
			defer dm.Unlock()

			// Move document in the queues.
			if dstQueue != "" {
				if err := q.AddDocument(dstQueue, d, 0); err != nil {
					logger.WithField("item_id", d.ID).WithError(err).Warn("failed to add document to queue")
					return
				}
			}
			if err := q.ReleaseDocument(srcQueue, dO); err != nil {
				fmt.Println(srcQueue)
				fmt.Println(dO.JSON())
				logger.WithField("item_id", d.ID).WithError(err).Warn("failed to release document")
				return
			}
		}()

		// Refresh task lock until done.
	outer:
		for {
			select {
			case <-done:
				break outer
			case <-time.After(relockPeriod):
				if lock != nil {
					if err := lock(lockDuration); err != nil {
						logger.WithField("item_id", d.ID).WithError(err).Warn("failed to renew document lock")
					}
				}
			case <-st.Chan():
				dm.Lock()
				logger.Debug("routine stopped")
				return
			}
		}
	}
}

func bookkeep(st *stopper.Stopper, q QueueModule) {
	defer st.End()

	t := time.NewTicker(bookkeepPeriod)
	for {
		select {
		case <-t.C:
			q.Bookkeep(queues)
		case <-st.Chan():
			log.WithField("role", "bookkeep").Debug("routine stopped")
			return
		}
	}
}

func mod(cfg Config, modS string) (interface{}, error) {
	modC := cfg.Modules[modS]
	if modC == nil {
		return nil, errors.New("module configuration is missing")
	}
	mod := modules.Modules[modC.Driver]
	if mod == nil {
		return nil, errors.New("undefined module")
	}
	return mod, nil
}

func queueMod(cfg Config, qModS string) (QueueModule, error) {
	qMod, err := mod(cfg, qModS)
	if err != nil {
		return nil, err
	}
	qModT, ok := qMod.(QueueModule)
	if !ok {
		return nil, errors.New("module is of wrong type")
	}
	return qModT, nil
}

func inputMod(cfg Config, iModS string) (InputModule, error) {
	iMod, err := mod(cfg, iModS)
	if err != nil {
		return nil, err
	}
	lModT, ok := iMod.(InputModule)
	if !ok {
		return nil, errors.New("module is of wrong type")
	}
	return lModT, nil
}

func processorMod(cfg Config, pModS string) (ProcessorModule, error) {
	pMod, err := mod(cfg, pModS)
	if err != nil {
		return nil, err
	}
	pModT, ok := pMod.(ProcessorModule)
	if !ok {
		return nil, errors.New("module is of wrong type")
	}
	return pModT, nil
}

func outputMod(cfg Config, oModS string) (OutputModule, error) {
	oMod, err := mod(cfg, oModS)
	if err != nil {
		return nil, err
	}
	oModT, ok := oMod.(OutputModule)
	if !ok {
		return nil, errors.New("module is of wrong type")
	}
	return oModT, nil
}

//func printStruct(s interface{}) {
//	enc := json.NewEncoder(os.Stdout)
//	enc.SetIndent("", "    ")
//	if err := enc.Encode(s); err != nil {
//		fmt.Println(err)
//	}
//}

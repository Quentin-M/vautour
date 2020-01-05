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

package pastebin

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/quentin-m/vautour/src/modules"
	"github.com/quentin-m/vautour/src/pkg/stopper"
	"github.com/quentin-m/vautour/src/pkg/vautour"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

const (
	listingURL = "https://scrape.pastebin.com/api_scraping.php?limit=250"
	scrapingURL = "https://scrape.pastebin.com/api_scrape_item.php?i=%s"
)

var (
	noAcccessREx = regexp.MustCompile("YOUR IP: .* DOES NOT HAVE ACCESS.")
)

type pastebin struct{
	Interval time.Duration
}

func init() {
	modules.Register("pastebin", &pastebin{})
}

func (p *pastebin) Configure(cfg *modules.ModuleConfig) error {
	p.Interval = 15 * time.Second
	return modules.ParseParams(cfg.Params, p)
}

func (p *pastebin) List(st *stopper.Stopper, ch chan *vautour.Document) error {
	if p.Interval <= 0 {
		<-st.Chan()
		return nil
	}

	t := time.NewTicker(p.Interval)
	for {
		// Wait for next loop.
		select {
		case <-st.Chan():
			return nil
		case <-t.C:
		}

		// Scrape.
		client := &http.Client{Timeout: time.Second * 5}
		res, err := client.Get(listingURL)
		if err != nil {
			log.WithField("role", "lister").WithField("module", "pastebin").WithError(err).Warn("failed to list new pastes")
			continue
		}

		// Decode & Add.
		var ps pastes
		if err := json.NewDecoder(res.Body).Decode(&ps); err != nil {
			res.Body.Close()
			log.WithField("role", "lister").WithField("module", "pastebin").WithError(err).Warn("failed to parse new list of pastes")
			continue
		}
		for _, p := range ps {
			ch <- &p.Document
		}
		res.Body.Close()
	}
}

func (p *pastebin) Scrape(d *vautour.Document) error {
	client := &http.Client{Timeout: time.Second * 5}
	res, err := client.Get(fmt.Sprintf(scrapingURL, d.ID))
	if err != nil {
		log.WithField("role", "scraper").WithField("module", "pastebin").WithField("item_id", d.ID).WithError(err).Warn("failed to scrape paste")
		return err
	}
	defer res.Body.Close()

	d.Content, err = ioutil.ReadAll(res.Body)
	if err != nil {
		log.WithField("role", "scraper").WithField("module", "pastebin").WithField("item_id", d.ID).WithError(err).Warn("failed to scrape paste")
		return err
	}

	// Sometimes, the scraping API will return unauthorized regardless of whether the IP is indeed authorized.
	// However, retries will eventually succeed.
	if err := errors.New(noAcccessREx.FindString(string(d.Content))); err.Error() != "" {
		log.WithField("role", "scraper").WithField("module", "pastebin").WithField("item_id", d.ID).WithError(err).Warn("failed to scrape paste")
		return err
	}

	return nil
}
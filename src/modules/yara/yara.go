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

package yara

import (
	"encoding/json"
	"fmt"
	"github.com/quentin-m/vautour/src/modules"
	"github.com/quentin-m/vautour/src/pkg/vautour"
	lib "github.com/hillu/go-yara"
	"os"
	"time"
	log "github.com/sirupsen/logrus"
)

type yara struct{
	Path string
	Timeout time.Duration

	c *lib.Compiler
	r *lib.Rules
}

func init() {
	modules.Register("yara", &yara{})
}

func (y *yara) Configure(moduleConfig *modules.ModuleConfig) error {
	// Default configuration.
	y.Path = "rules/_index.yar"
	y.Timeout = 15 * time.Second

	// Parse parameters.
	if err := modules.ParseParams(moduleConfig.Params, &y); err != nil {
		return fmt.Errorf("invalid configuration: %v", err)
	}

	// Create compiler.
	c, err := lib.NewCompiler()
	if err != nil {
		return err
	}
	y.c = c

	// Read rules file.
	f, err := os.Open(y.Path)
	if err != nil {
		return fmt.Errorf("could not open yara file %s: %s", y.Path, err)
	}
	defer f.Close()

	// Compile rules.
	if err := c.AddFile(f, ""); err != nil {
		return fmt.Errorf("could not compile yara rules: %s", err)
	}

	r, err := c.GetRules()
	if err != nil {
		return fmt.Errorf("could not read the compiled rules back: %s", err)
	}
	for _, r := range r.GetRules() {
		log.WithField("role", "processor").WithField("module", "yara").Debugf("compiled rule %s", r.Identifier())
	}
	y.r = r

	return nil
}

func (y *yara) Process(d *vautour.Document) error {
	matches, err := y.r.ScanMem(d.Content, 0, y.Timeout)
	if err != nil {
		return err
	}

	for _, match := range matches {
		j, err := json.Marshal(match)
		if err != nil {
			log.WithField("role", "processor").WithField("module", "yara").WithField("item_id", d.ID).WithField("rule", match.Rule).Warn("failed to marshal match result")
		}
		d.Processed = append(d.Processed, vautour.ProcessedData{
			Module: "yara", // TODO: This is not the actual module name, but the driver name - but I am feeling lazy.
			RawMessage: j,
		})
		if score, ok := match.Meta["score"].(int32); ok && int(score) > d.Score {
			d.Score = int(score)
		}
		log.WithField("role", "processor").WithField("module", "yara").WithField("item_id", d.ID).WithField("rule", match.Rule).Debug("matched rule")
	}

	return nil
}
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

package mailer

import (
	"encoding/json"
	"fmt"
	"github.com/quentin-m/vautour/src/modules"
	"github.com/quentin-m/vautour/src/pkg/vautour"

	mail "github.com/kataras/go-mailer"
)

type mailer struct {
	SMTP mail.Config
	Recipients []string
	MinScore int
}

func init() {
	modules.Register("mailer", &mailer{})
}

func (e *mailer) Configure(cfg *modules.ModuleConfig) error {
	// Default configuration.
	e.SMTP.UseCommand = true

	// Parse parameters.
	if err := modules.ParseParams(cfg.Params, e); err != nil {
		return err
	}

	return nil
}

func (e *mailer) Send(d *vautour.Document) error {
	// Skip if the score is too low.
	if d.Score < e.MinScore {
		return nil
	}

	// Format the document.
	dj, err := json.MarshalIndent(d, "", "    ")
	if err != nil {
		return err
	}

	// Send the e-mail.
	return mail.New(e.SMTP).Send(
		fmt.Sprintf("[Vautour] An item from %s matched with score %d", d.InputModuleName, d.Score),
		string(dj),
		e.Recipients...
	)

	return nil
}

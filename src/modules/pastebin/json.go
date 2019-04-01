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
	"github.com/quentin-m/vautour/src/pkg/vautour"
	"strconv"
	"strings"
	"time"
)

type pastes []*paste

type paste struct {
	vautour.Document
}

func (p *paste) UnmarshalJSON(b []byte) (err error) {
	v := map[string]interface{}{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	p.ID = v["key"].(string)
	p.Title = v["title"].(string)
	p.User = v["user"].(string)
	p.URL = v["full_url"].(string)

	if v, err := string2time(v["date"].(string)); err == nil {
		p.CreatedAt = v
	}
	if v, err := string2int(v["size"].(string)); err == nil {
		p.Size = v
	}
	if v, err := string2time(v["expire"].(string)); err == nil {
		p.ExpireAt = v
	}

	return
}

func string2time(s string) (time.Time, error) {
	r := strings.Replace(s, `"`, ``, -1)

	q, err := strconv.ParseInt(r, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	if q == 0 {
		return time.Time{}, nil
	}

	t := time.Unix(q, 0)
	return t, nil
}

func string2int(s string) (int, error) {
	r := strings.Replace(s, `"`, ``, -1)

	q, err := strconv.ParseInt(r, 10, 64)
	if err != nil {
		return 0, err
	}

	return int(q), nil
}

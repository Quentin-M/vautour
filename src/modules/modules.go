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

package modules

import (
	"gopkg.in/yaml.v2"
	"reflect"
	"sync"
	"unsafe"
)

var (
	mu      sync.RWMutex
	Modules = make(map[string]interface{})
)

type ModuleConfig struct {
	Driver string
	Params map[string]interface{} `yaml:",inline"`
}

func Register(name string, e interface{}) {
	if name == ""  {
		panic("could not register a module with an empty name")
	}
	if e == nil {
		panic("could not register a nil module")
	}
	if _, ok := Modules[name]; ok {
		panic("Register called twice for module " + name)
	}

	mu.Lock()
	defer mu.Unlock()
	Modules[name] = e
}

func ParseParams(params map[string]interface{}, cfg interface{}) error {
	yConfig, err := yaml.Marshal(params)
	if err != nil {
		return err
	}

	// We could use `reflect.New()` and avoid using the unsafe package if we are
	// ok with returning a new config object (obj), and therefore using
	// `p.config = *cfg.(*providerXConfig)` in the caller function
	// `func (p *providerXConfig) Configure(GenericProvider). However, this means
	// that defaults cannot be set.
	obj := reflect.NewAt(reflect.TypeOf(cfg).Elem(), unsafe.Pointer(reflect.ValueOf(cfg).Pointer())).Interface()
	if err := yaml.Unmarshal(yConfig, obj); err != nil {
		return err
	}

	return nil
}
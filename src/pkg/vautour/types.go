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
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/quentin-m/vautour/src/modules"
	"github.com/quentin-m/vautour/src/pkg/stopper"
	"time"
)

// Configuration

type Config struct {
	Modules map[string]*modules.ModuleConfig

	Inputs InputsConfig
	Queues QueuesConfig
	Scrapers ScrapersConfig
	Processors ProcessorsConfig
	Outputs OutputsConfig
}

type InputsConfig struct {
	Modules []string
}

type QueuesConfig struct {
	Module string
}

type ScrapersConfig struct {
	Modules []string
	Threads int
}

type ProcessorsConfig struct {
	Modules []string
	Threads int
}

type OutputsConfig struct {
	Modules []string
	Threads int
}

// Document

type Document struct {
	ID string
	Title string
	User string  `json:",omitempty"`
	Size int
	URL string
	Content []byte

	CreatedAt time.Time
	ExpireAt time.Time

	Score int
	Processed []ProcessedData `json:",omitempty"`

	// Internally managed //
	InputModuleName string
}

func NewDocumentFromJSON(s string) (*Document, error) {
	var d Document
	if err := json.Unmarshal([]byte(s), &d); err != nil {
		return nil, err
	}
	return &d, nil
}

func (d *Document) JSON() string {
	b, _ := json.Marshal(d)
	return string(b)
}

// Match

type ProcessedData struct {
	Module string
	json.RawMessage
}

// Errors

var ErrAlreadyExists = errors.New("document already exists")

// Interfaces

type QueueModule interface {
	Configure(moduleConfig *modules.ModuleConfig) error
	AddDocument(queue string, d *Document, cacheTTL time.Duration) error
	GetDocument(queue string, ttl time.Duration) (string, *Document, func (time.Duration) error, error)
	ReleaseDocument(queue string, d *Document) error
	DeleteDocument(queue string, d *Document) error
	Bookkeep(queues []string)
}

type InputModule interface {
	Configure(*modules.ModuleConfig) error
	List(*stopper.Stopper, chan *Document) error
	Scrape(*Document) error
}

type ProcessorModule interface {
	Configure(*modules.ModuleConfig) error
	Process(*Document) error
}

type OutputModule interface {
	Configure(*modules.ModuleConfig) error
	Send(*Document) error
}

// Copyright 2019 clair authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package stopper

import (
	"sync"
	"time"
)

// Stopper eases the graceful termination of a group of goroutines
type Stopper struct {
	running bool
	wg   sync.WaitGroup
	stop chan struct{}
}

// NewStopper initializes a new Stopper instance
func NewStopper() *Stopper {
	return &Stopper{stop: make(chan struct{}, 0), running: true}
}

// Begin indicates that a new goroutine has started.
func (s *Stopper) Begin() {
	s.wg.Add(1)
}

// End indicates that a goroutine has stopped.
func (s *Stopper) End() {
	s.wg.Done()
}

// Chan returns the channel on which goroutines could listen to determine if
// they should stop. The channel is closed when Stop() is called.
func (s *Stopper) Chan() chan struct{} {
	return s.stop
}

// Sleep puts the current goroutine on sleep during a duration d
// Sleep could be interrupted in the case the goroutine should stop itself,
// in which case Sleep returns false.
func (s *Stopper) Sleep(d time.Duration) bool {
	if d < 0 {
		return true
	}
	select {
	case <-time.After(d):
		return true
	case <-s.stop:
		return false
	}
}

// IsRunning returns whether the application is still expected to run.
func (s *Stopper) IsRunning() bool {
	return s.running
}

// Stop asks every goroutine to end.
func (s *Stopper) Stop() {
	s.running = false
	close(s.stop)
	s.wg.Wait()
}
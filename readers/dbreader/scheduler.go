// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package dbreader

import (
	"time"
)

// Scheduler - used for scheduled reading from databases
type scheduler struct {
	ticker   *time.Ticker
	interval float64
	done     chan bool
}

// Start starts the ticker
func (s *scheduler) start() {
	s.stop()
	s.ticker = time.NewTicker(time.Second * time.Duration(s.interval))
}

// Stop stops the ticker
func (s *scheduler) stop() {
	if s.ticker != nil {
		s.ticker.Stop()
		s.done <- true
	}
}

// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows

package collector

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"

	"github.com/prometheus-community/windows_exporter/internal/mi"
)

var (
	errCollectorBusy    = errors.New("collector is still running")
	errCollectionClosed = errors.New("collection is closed")
)

// collectionState is the state of a Collection that its filtered views share.
type collectionState struct {
	mu         sync.Mutex
	collectors map[string]*collectorState
	miApp      *mi.Application
	miSession  *mi.Session

	// closeMu serializes Close calls.
	closeMu sync.Mutex
}

func newCollectionState() *collectionState {
	return &collectionState{collectors: make(map[string]*collectorState)}
}

// collector returns the state of the named collector, creating it on first use.
func (s *collectionState) collector(name string) *collectorState {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.collectors[name]
	if !ok {
		state = &collectorState{slot: make(chan struct{}, 1)}
		s.collectors[name] = state
	}

	return state
}

func (s *collectionState) session() *mi.Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.miSession
}

func (s *collectionState) setMI(app *mi.Application, session *mi.Session) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.miApp = app
	s.miSession = session
}

// closeMI closes the MI session and application. Later calls do nothing.
func (s *collectionState) closeMI() []error {
	s.mu.Lock()
	app, session := s.miApp, s.miSession
	s.miApp, s.miSession = nil, nil
	s.mu.Unlock()

	var errs []error

	if session != nil {
		if err := session.Close(); err != nil && !errors.Is(err, mi.ErrNotInitialized) {
			errs = append(errs, fmt.Errorf("error from close MI session: %w", err))
		}
	}

	if app != nil {
		if err := app.Close(); err != nil && !errors.Is(err, mi.ErrNotInitialized) {
			errs = append(errs, fmt.Errorf("error from close MI application: %w", err))
		}
	}

	return errs
}

type startResult int

const (
	started startResult = iota
	busy
	notAvailable
)

// collectorState tracks the lifecycle of one collector instance.
type collectorState struct {
	// slot is held while Build, Collect or Close of the collector runs.
	// A Collect call that outlives its scrape keeps the slot until it returns,
	// so no scrape starts a second call on the same instance.
	slot chan struct{}

	mu sync.Mutex
	// buildErr is the error of the last Build. A collector that failed to build is not collected.
	buildErr error
	// built is true once Build was called, so Close only closes collectors that may hold resources.
	built bool
	// closing is true once Close was called. No Build or Collect starts afterwards.
	closing bool
	// closed is true once Collector.Close was called.
	closed bool
	// busyLogged is true once a scrape logged that it skipped the current run.
	busyLogged bool
}

// tryCollect acquires the slot for a Collect call without waiting.
func (s *collectorState) tryCollect() startResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closing || s.buildErr != nil {
		return notAvailable
	}

	select {
	case s.slot <- struct{}{}:
		return started
	default:
		return busy
	}
}

// release releases the slot acquired by tryCollect.
func (s *collectorState) release() {
	s.mu.Lock()
	s.busyLogged = false
	s.mu.Unlock()

	<-s.slot
}

// logBusy reports whether a skipped scrape is the first one of the current run.
func (s *collectorState) logBusy() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	first := !s.busyLogged
	s.busyLogged = true

	return first
}

// build runs Collector.Build while it holds the slot and records the result.
// A panic in Collector.Build is returned as an error.
func (s *collectorState) build(ctx context.Context, logger *slog.Logger, miSession *mi.Session, collector Collector) (err error) {
	select {
	case s.slot <- struct{}{}:
	case <-ctx.Done():
		err = context.Cause(ctx)

		s.mu.Lock()
		s.buildErr = err
		s.mu.Unlock()

		return err
	}

	defer func() {
		s.mu.Lock()
		s.buildErr = err
		s.mu.Unlock()

		<-s.slot
	}()

	s.mu.Lock()
	closing := s.closing
	s.built = !closing
	s.mu.Unlock()

	if closing {
		return errCollectionClosed
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v. stack: %s", r, string(debug.Stack()))
		}
	}()

	return collector.Build(logger, miSession)
}

// close calls Collector.Close once the collector is idle. It waits until ctx is done for a running
// Build or Collect call to return. If the call does not return in time, the collector is left open,
// because closing it would release resources that are still in use.
// After close, the collector is never built or collected again.
func (s *collectorState) close(ctx context.Context, collector Collector) (err error) {
	s.mu.Lock()
	s.closing = true
	done := s.closed || !s.built
	s.mu.Unlock()

	if done {
		return nil
	}

	select {
	case s.slot <- struct{}{}:
		// The slot is never released, so no call starts after Close.
	case <-ctx.Done():
		return errCollectorBusy
	}

	s.mu.Lock()
	s.closed = true
	s.mu.Unlock()

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v. stack: %s", r, string(debug.Stack()))
		}
	}()

	return collector.Close()
}

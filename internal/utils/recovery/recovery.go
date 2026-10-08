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

// Package recovery runs functions in a way that a panic is returned as an error
// instead of terminating the process. A panic in a goroutine cannot be recovered
// by the caller that started it, so every goroutine a collector starts must
// recover on its own.
package recovery

import (
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
)

// ErrPanic is wrapped by every error that Run and Group return for a recovered panic.
var ErrPanic = errors.New("panic")

// Run calls fn and returns its error. If fn panics, Run recovers and returns
// an error wrapping [ErrPanic] that contains the panic value and the stack trace.
func Run(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%w: %v\n%s", ErrPanic, r, debug.Stack())
		}
	}()

	return fn()
}

// Group runs functions in goroutines and collects all of their errors.
// A panic in one function is returned as an error by Wait and does not affect
// the other functions. The zero value is ready to use. Go may be called from
// within a function that runs in the group.
type Group struct {
	wg   sync.WaitGroup
	mu   sync.Mutex
	errs []error
}

// Go calls fn through [Run] in a new goroutine.
func (g *Group) Go(fn func() error) {
	g.wg.Go(func() {
		if err := Run(fn); err != nil {
			g.mu.Lock()
			g.errs = append(g.errs, err)
			g.mu.Unlock()
		}
	})
}

// Wait blocks until all functions have returned and returns their errors joined with [errors.Join].
func (g *Group) Wait() error {
	g.wg.Wait()

	g.mu.Lock()
	defer g.mu.Unlock()

	return errors.Join(g.errs...)
}

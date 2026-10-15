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

package ole

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/utils/recovery"
)

var (
	// ErrWorkerStopped is returned for work submitted after Worker.Close.
	ErrWorkerStopped = errors.New("COM worker stopped")
	// ErrWorkerTimeout is returned when a worker does not answer within the caller's budget.
	ErrWorkerTimeout = errors.New("COM worker did not answer in time")
)

// Worker runs functions on one goroutine that is locked to its OS thread and
// has joined the multithreaded apartment for its whole lifetime.
//
// Initializing and uninitializing COM for each call unloads in-process COM
// servers whenever the apartment's last reference goes away, so every call
// would load them again. A Worker initializes COM once in NewWorker and
// uninitializes it after it stops.
//
// Functions run one at a time in submission order. COM objects created by a
// function belong to the worker's thread: release them before the function
// returns and never hand them to the caller.
type Worker struct {
	jobs chan func()
	stop chan struct{}
	// done is closed after the worker's goroutine returned and uninitialized COM.
	done      chan struct{}
	closeOnce sync.Once
}

// NewWorker starts a worker and returns the error of COM initialization.
// A worker returned without an error must be closed.
func NewWorker() (*Worker, error) {
	w := &Worker{
		jobs: make(chan func()),
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}

	initErrCh := make(chan error, 1)

	go w.run(initErrCh)

	if err := <-initErrCh; err != nil {
		<-w.done

		return nil, err
	}

	return w, nil
}

func (w *Worker) run(initErrCh chan<- error) {
	defer close(w.done)

	// COM initialization and every interface call stay on the same OS thread.
	// The thread is not unlocked: if Uninitialize fails to balance an
	// initialization, the thread exits with the goroutine instead of carrying
	// apartment state to other goroutines.
	runtime.LockOSThread()

	if err := Initialize(); err != nil {
		initErrCh <- err

		return
	}

	defer Uninitialize()

	initErrCh <- nil

	for {
		select {
		case job := <-w.jobs:
			job()
		case <-w.stop:
			return
		}
	}
}

// Close stops the worker. A function that is still running finishes first,
// then the worker uninitializes COM. Close waits up to timeout for that and
// returns ErrWorkerTimeout if the worker is still busy. The worker then stops
// and uninitializes COM on its own once the function returns. Close is safe to
// call more than once.
func (w *Worker) Close(timeout time.Duration) error {
	w.closeOnce.Do(func() { close(w.stop) })

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-w.done:
		return nil
	case <-timer.C:
		return fmt.Errorf("stop: %w", ErrWorkerTimeout)
	}
}

// Do runs fn on the worker and returns its result. A panic in fn is returned
// as an error wrapping [recovery.ErrPanic], and the worker keeps running.
// COM objects that fn created before the panic leak unless fn releases them
// in deferred calls. If fn calls runtime.Goexit, the worker uninitializes COM
// and exits, and this and every later call return ErrWorkerStopped.
//
// If ctx is done before the worker starts or finishes fn, Do returns an error
// wrapping ErrWorkerTimeout and the cause of ctx. A function that already
// started keeps running on the worker; its result is discarded, so fn must not
// write to memory that the caller reads after Do returns. Functions submitted
// later wait until it returns.
//
//nolint:ireturn // T is the concrete result type of fn.
func Do[T any](ctx context.Context, w *Worker, fn func() (T, error)) (T, error) {
	type result struct {
		value T
		err   error
	}

	// Buffered, so that the worker does not block on an abandoned call.
	resultCh := make(chan result, 1)

	job := func() {
		var value T

		err := recovery.Run(func() error {
			var err error

			value, err = fn()

			return err
		})

		resultCh <- result{value: value, err: err}
	}

	var zero T

	if ctx.Err() != nil {
		return zero, fmt.Errorf("%w: %w", ErrWorkerTimeout, context.Cause(ctx))
	}

	select {
	case w.jobs <- job:
	case <-w.stop:
		return zero, ErrWorkerStopped
	case <-w.done:
		return zero, ErrWorkerStopped
	case <-ctx.Done():
		return zero, fmt.Errorf("%w: worker busy: %w", ErrWorkerTimeout, context.Cause(ctx))
	}

	select {
	case r := <-resultCh:
		return r.value, r.err
	case <-w.done:
		// The worker exited, normally after fn sent its result, but also if fn
		// called runtime.Goexit, which recovery.Run does not stop.
		select {
		case r := <-resultCh:
			return r.value, r.err
		default:
			return zero, ErrWorkerStopped
		}
	case <-ctx.Done():
		return zero, fmt.Errorf("%w: %w", ErrWorkerTimeout, context.Cause(ctx))
	}
}

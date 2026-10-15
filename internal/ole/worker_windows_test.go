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
	"testing"
	"time"
	"unsafe"

	"github.com/prometheus-community/windows_exporter/internal/utils/recovery"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

type apartment struct {
	threadID  uint32
	kind      int32
	qualifier int32
}

func currentApartment() (apartment, error) {
	a := apartment{threadID: windows.GetCurrentThreadId()}

	hr, _, _ := ole32.NewProc("CoGetApartmentType").Call(
		uintptr(unsafe.Pointer(&a.kind)),
		uintptr(unsafe.Pointer(&a.qualifier)),
	)

	return a, ResultError(hr)
}

func TestWorkerKeepsApartment(t *testing.T) {
	w, err := NewWorker()
	require.NoError(t, err)

	first, err := Do(t.Context(), w, currentApartment)
	require.NoError(t, err)
	require.Equal(t, int32(1), first.kind, "APTTYPE_MTA")
	require.Equal(t, int32(0), first.qualifier, "explicitly initialized, not the implicit MTA")

	for range 10 {
		next, err := Do(t.Context(), w, currentApartment)
		require.NoError(t, err)
		require.Equal(t, first, next, "every call runs on the same initialized thread")
	}

	require.NoError(t, w.Close(time.Second))
	require.NoError(t, w.Close(time.Second), "Close is idempotent")

	_, err = Do(t.Context(), w, currentApartment)
	require.ErrorIs(t, err, ErrWorkerStopped)
}

func TestWorkerRecoversPanics(t *testing.T) {
	w, err := NewWorker()
	require.NoError(t, err)

	defer func() { require.NoError(t, w.Close(time.Second)) }()

	_, err = Do(t.Context(), w, func() (int, error) { panic("boom") })
	require.ErrorIs(t, err, recovery.ErrPanic)

	value, err := Do(t.Context(), w, func() (int, error) { return 42, nil })
	require.NoError(t, err)
	require.Equal(t, 42, value, "the worker keeps running after a panic")

	errTest := errors.New("test")
	value, err = Do(t.Context(), w, func() (int, error) { return 7, errTest })
	require.ErrorIs(t, err, errTest)
	require.Equal(t, 7, value, "partial results are returned with the error")
}

func TestWorkerBudget(t *testing.T) {
	w, err := NewWorker()
	require.NoError(t, err)

	release := make(chan struct{})
	running := make(chan struct{})

	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()

	_, err = Do(ctx, w, func() (struct{}, error) {
		close(running)
		<-release

		return struct{}{}, nil
	})
	require.ErrorIs(t, err, ErrWorkerTimeout)
	require.ErrorIs(t, err, context.DeadlineExceeded)

	<-running

	// The abandoned call still occupies the worker.
	ctx, cancel = context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()

	_, err = Do(ctx, w, func() (struct{}, error) { return struct{}{}, nil })
	require.ErrorIs(t, err, ErrWorkerTimeout)

	// A done context fails before the call is submitted.
	canceled, cancelNow := context.WithCancel(t.Context())
	cancelNow()

	_, err = Do(canceled, w, func() (struct{}, error) {
		t.Error("a call with a done context must not run")

		return struct{}{}, nil
	})
	require.ErrorIs(t, err, ErrWorkerTimeout)

	// Close waits for the running call within its bound.
	require.ErrorIs(t, w.Close(50*time.Millisecond), ErrWorkerTimeout)

	close(release)

	require.NoError(t, w.Close(time.Second), "the worker stops once the call returns")
}

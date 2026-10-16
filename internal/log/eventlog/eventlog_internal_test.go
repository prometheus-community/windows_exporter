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

package eventlog

import (
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// fakeEventLog records the events it receives and fails every use after Close,
// like a closed event log handle.
type fakeEventLog struct {
	mu     sync.Mutex
	events []string
	closes int
}

func (f *fakeEventLog) record(level, msg string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closes > 0 {
		return os.ErrInvalid
	}

	f.events = append(f.events, level+": "+msg)

	return nil
}

func (f *fakeEventLog) Info(_ uint32, msg string) error    { return f.record("info", msg) }
func (f *fakeEventLog) Warning(_ uint32, msg string) error { return f.record("warning", msg) }
func (f *fakeEventLog) Error(_ uint32, msg string) error   { return f.record("error", msg) }

func (f *fakeEventLog) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.closes++

	return nil
}

func TestWriterLevels(t *testing.T) {
	t.Parallel()

	handle := &fakeEventLog{}
	writer := &Writer{handle: handle}

	for _, line := range []string{
		"time=2026-10-15T14:03:05Z level=INFO msg=started\n",
		"time=2026-10-15T14:03:05Z level=WARN msg=slow\n",
		"time=2026-10-15T14:03:05Z level=ERROR msg=failed\n",
	} {
		n, err := writer.Write([]byte(line))
		require.NoError(t, err)
		require.Equal(t, len(line), n)
	}

	require.Equal(t, []string{"info: msg=started", "warning: msg=slow", "error: msg=failed"}, handle.events)
}

func TestWriterWriteAfterClose(t *testing.T) {
	t.Parallel()

	handle := &fakeEventLog{}
	writer := &Writer{handle: handle}

	require.NoError(t, writer.Close())
	require.NoError(t, writer.Close(), "a second Close is a no-op")
	require.Equal(t, 1, handle.closes, "the handle must be closed once")

	n, err := writer.Write([]byte("level=INFO msg=late\n"))
	require.ErrorIs(t, err, os.ErrClosed)
	require.Zero(t, n)
	require.Empty(t, handle.events)
}

func TestWriterConcurrentClose(t *testing.T) {
	t.Parallel()

	handle := &fakeEventLog{}
	writer := &Writer{handle: handle}

	var wg sync.WaitGroup

	for range 8 {
		wg.Go(func() {
			for range 100 {
				// A write either reaches the open handle or fails with os.ErrClosed.
				if _, err := writer.Write([]byte("level=INFO msg=x\n")); err != nil {
					require.ErrorIs(t, err, os.ErrClosed)
				}
			}
		})
	}

	wg.Go(func() {
		require.NoError(t, writer.Close())
	})

	wg.Wait()

	require.Equal(t, 1, handle.closes)
}

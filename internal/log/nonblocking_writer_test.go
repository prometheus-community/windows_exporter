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

package log

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type blockingWriter struct {
	started chan struct{}
	release chan struct{}

	mu     sync.Mutex
	writes [][]byte
}

func (w *blockingWriter) Write(p []byte) (int, error) {
	select {
	case <-w.started:
	default:
		close(w.started)
	}

	<-w.release

	w.mu.Lock()
	defer w.mu.Unlock()

	w.writes = append(w.writes, append([]byte(nil), p...))

	return len(p), nil
}

func (w *blockingWriter) Writes() [][]byte {
	w.mu.Lock()
	defer w.mu.Unlock()

	result := make([][]byte, len(w.writes))
	copy(result, w.writes)

	return result
}

func TestNonBlockingWriterDropsInsteadOfBlocking(t *testing.T) {
	t.Parallel()

	writer := &blockingWriter{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	var cleanupOnce sync.Once

	nonBlockingWriter := newNonBlockingWriter(writer, 1)
	t.Cleanup(func() {
		cleanupOnce.Do(func() {
			close(writer.release)
			require.NoError(t, nonBlockingWriter.Close())
		})
	})

	_, err := nonBlockingWriter.Write([]byte("first"))
	require.NoError(t, err)

	select {
	case <-writer.started:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for background write to start")
	}

	_, err = nonBlockingWriter.Write([]byte("second"))
	require.NoError(t, err)

	done := make(chan struct{})

	go func() {
		defer close(done)

		_, err := nonBlockingWriter.Write([]byte("third"))
		require.NoError(t, err)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("write blocked while output writer was stalled")
	}

	cleanupOnce.Do(func() {
		close(writer.release)
		require.NoError(t, nonBlockingWriter.Close())
	})

	writes := writer.Writes()
	require.Len(t, writes, 2)
	require.Equal(t, []byte("first"), writes[0])

	for _, write := range writes {
		require.NotEqual(t, []byte("third"), write)
	}
}

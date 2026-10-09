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

package process

import (
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCloseBeforeWorkerStarts(t *testing.T) {
	t.Parallel()

	c := New(nil)
	requests := make(chan processWorkerRequest)
	c.workerCh = requests
	c.workerWG.Add(1)

	closed := make(chan error, 1)
	go func() { closed <- c.Close() }()

	select {
	case _, ok := <-requests:
		require.False(t, ok)
	case <-time.After(time.Second):
		t.Fatal("Close did not close the worker channel")
	}

	// The worker starts after Close has set c.workerCh to nil.
	go c.collectWorker(requests)

	select {
	case err := <-closed:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("Close did not wait for the delayed worker to exit")
	}

	require.NoError(t, c.Close())
}

func TestWorkerContinuesAfterRequestPanic(t *testing.T) {
	t.Parallel()

	c := New(nil)
	c.logger = slog.New(slog.DiscardHandler)
	c.workerCh = make(chan processWorkerRequest, 2)

	c.workerWG.Add(1)
	go c.collectWorker(c.workerCh)

	var requests sync.WaitGroup
	requests.Add(2)

	for range 2 {
		// Missing descriptors make metric construction panic in both requests.
		c.workerCh <- processWorkerRequest{waitGroup: &requests}
	}

	processed := make(chan struct{})

	go func() { requests.Wait(); close(processed) }()

	select {
	case <-processed:
	case <-time.After(time.Second):
		t.Fatal("worker did not recover and process the next request")
	}

	require.NoError(t, c.Close())
}

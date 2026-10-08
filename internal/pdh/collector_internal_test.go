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

package pdh

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

type processThreads struct {
	Name        string
	ThreadCount float64 `perfdata:"Thread Count"`
}

// setFieldIndex points the Thread Count counter at field index.
func setFieldIndex(c *Collector[processThreads], index int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	counter := c.counters["Thread Count"]
	counter.FieldIndexValue = index
	c.counters["Thread Count"] = counter
}

// TestCollectRecoversPanic checks that a panic in the worker goroutine is
// returned as an error instead of terminating the process.
func TestCollectRecoversPanic(t *testing.T) {
	t.Parallel()

	c, err := NewCollector[processThreads](slog.New(slog.DiscardHandler), CounterTypeRaw, "Process", InstancesAll)
	require.NoError(t, err)

	t.Cleanup(c.Close)

	var dst []processThreads

	require.NoError(t, c.Collect(&dst))

	// Point the counter at a field that does not exist, so that reflect panics.
	setFieldIndex(c, 42)

	err = c.Collect(&dst)
	require.ErrorContains(t, err, "panic while collecting performance counters of Process")

	// The worker survives and serves the next request.
	setFieldIndex(c, 1)

	require.NoError(t, c.Collect(&dst))
	require.NotEmpty(t, dst)
}

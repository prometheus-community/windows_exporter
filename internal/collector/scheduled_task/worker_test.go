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

package scheduled_task

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/ole"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

// TestCollectHonorsBudget checks that Collect returns within its budget while
// the COM worker is busy, and that the worker answers again afterwards.
func TestCollectHonorsBudget(t *testing.T) {
	c := New(nil)
	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), nil))

	release := make(chan struct{})
	running := make(chan struct{})
	busyDone := make(chan error, 1)

	go func() {
		_, err := ole.Do(context.Background(), c.worker, func() (struct{}, error) {
			close(running)
			<-release

			return struct{}{}, nil
		})
		busyDone <- err
	}()

	<-running

	ch := make(chan prometheus.Metric, 1024)
	start := time.Now()

	err := c.Collect(ch, 100*time.Millisecond)
	require.ErrorIs(t, err, ole.ErrWorkerTimeout)
	require.Less(t, time.Since(start), time.Second)
	require.Empty(t, ch)

	close(release)
	require.NoError(t, <-busyDone)

	// The worker answers again once the earlier read returned.
	require.NoError(t, c.Collect(ch, time.Minute))
	require.NoError(t, c.Close())
	require.NoError(t, c.Close(), "Close is idempotent")
	require.ErrorIs(t, c.Collect(ch, time.Minute), ole.ErrWorkerStopped)
}

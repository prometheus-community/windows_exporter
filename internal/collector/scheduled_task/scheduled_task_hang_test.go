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

package scheduled_task_test

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/collector/scheduled_task"
	"github.com/prometheus-community/windows_exporter/pkg/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/require"
)

// TestHungWorkerSkipsLaterScrapes checks that a Task Scheduler call that does
// not return times out one scrape, and that later scrapes skip the collector
// at once instead of waiting for the hung worker for their whole budget.
func TestHungWorkerSkipsLaterScrapes(t *testing.T) {
	c := scheduled_task.New(nil)
	collection := collector.New(collector.Map{scheduled_task.Name: c})

	require.NoError(t, collection.Build(t.Context(), slog.New(slog.DiscardHandler)))

	release := scheduled_task.HangWorker(t, c)

	handler, err := collection.NewHandler(300*time.Millisecond, slog.New(slog.DiscardHandler), nil)
	require.NoError(t, err)

	// The first scrape starts Collect, which waits for the hung worker until the budget ends.
	first := gather(t, handler)
	require.Contains(t, first, `windows_exporter_collector_timeout{collector="scheduled_task"} 1`)

	// The second scrape finds the first Collect still running and skips the collector.
	start := time.Now()
	second := gather(t, handler)
	require.Less(t, time.Since(start), 150*time.Millisecond, "a later scrape does not wait for the hung worker")
	require.Contains(t, second, `windows_exporter_collector_timeout{collector="scheduled_task"} 1`)

	release()

	// Once the worker answers, the pending Collect returns and the collector succeeds again.
	require.Eventually(t, func() bool {
		ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
		defer cancel()

		handler, err := collection.NewHandlerWithContext(ctx, 30*time.Second, slog.New(slog.DiscardHandler), nil)
		require.NoError(t, err)

		return strings.Contains(gather(t, handler), `windows_exporter_collector_success{collector="scheduled_task"} 1`)
	}, time.Minute, 100*time.Millisecond)

	require.NoError(t, collection.Close())
}

func gather(t *testing.T, handler prometheus.Collector) string {
	t.Helper()

	registry := prometheus.NewRegistry()
	require.NoError(t, registry.Register(handler))

	families, err := registry.Gather()
	require.NoError(t, err)

	var out strings.Builder

	for _, family := range families {
		_, err := expfmt.MetricFamilyToText(&out, family)
		require.NoError(t, err)
	}

	return out.String()
}

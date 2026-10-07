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

package performancecounter_test

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/performancecounter"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/require"
)

type collectorAdapter struct {
	performancecounter.Collector
}

// Describe implements the prometheus.Collector interface.
func (a collectorAdapter) Describe(_ chan<- *prometheus.Desc) {}

// Collect implements the prometheus.Collector interface.
func (a collectorAdapter) Collect(ch chan<- prometheus.Metric) {
	if err := a.Collector.Collect(ch, 0); err != nil {
		panic(fmt.Sprintf("failed to update collector: %v", err))
	}
}

func TestCollector(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name            string
		object          string
		counterType     pdh.CounterType
		instances       []string
		instanceLabel   string
		buildErr        string
		counters        []performancecounter.Counter
		expectedMetrics *regexp.Regexp
	}{
		{
			name:        "memory",
			object:      "Memory",
			counterType: pdh.CounterTypeRaw,
			instances:   nil,
			buildErr:    "",
			counters:    []performancecounter.Counter{{Name: "Available Bytes", Type: "gauge"}},
			expectedMetrics: regexp.MustCompile(`^# HELP windows_performancecounter_collector_duration_seconds windows_exporter: Duration of an performancecounter child collection.
# TYPE windows_performancecounter_collector_duration_seconds gauge
windows_performancecounter_collector_duration_seconds\{collector="memory"} [0-9.e+-]+
# HELP windows_performancecounter_collector_success windows_exporter: Whether a performancecounter child collector was successful.
# TYPE windows_performancecounter_collector_success gauge
windows_performancecounter_collector_success\{collector="memory"} 1
# HELP windows_performancecounter_memory_available_bytes windows_exporter: custom Performance Counter metric
# TYPE windows_performancecounter_memory_available_bytes gauge
windows_performancecounter_memory_available_bytes [0-9.e+-]+`),
		},
		{
			name:        "process",
			object:      "Process",
			counterType: "",
			instances:   []string{"*"},
			buildErr:    "",
			counters:    []performancecounter.Counter{{Name: "Thread Count", Type: "counter"}},
			expectedMetrics: regexp.MustCompile(`^# HELP windows_performancecounter_collector_duration_seconds windows_exporter: Duration of an performancecounter child collection.
# TYPE windows_performancecounter_collector_duration_seconds gauge
windows_performancecounter_collector_duration_seconds\{collector="process"} [0-9.e+-]+
# HELP windows_performancecounter_collector_success windows_exporter: Whether a performancecounter child collector was successful.
# TYPE windows_performancecounter_collector_success gauge
windows_performancecounter_collector_success\{collector="process"} 1
# HELP windows_performancecounter_process_thread_count windows_exporter: custom Performance Counter metric
# TYPE windows_performancecounter_process_thread_count counter
windows_performancecounter_process_thread_count\{instance=".+"} [0-9.e+-]+
.*`),
		},
		{
			name:          "processor_information",
			object:        "Processor Information",
			counterType:   pdh.CounterTypeRaw,
			instances:     []string{"*"},
			instanceLabel: "core",
			buildErr:      "",
			counters:      []performancecounter.Counter{{Name: "% Processor Time", Metric: "windows_performancecounter_processor_information_processor_time", Labels: map[string]string{"state": "active"}}, {Name: "% Idle Time", Metric: "windows_performancecounter_processor_information_processor_time", Labels: map[string]string{"state": "idle"}}},
			expectedMetrics: regexp.MustCompile(`^# HELP windows_performancecounter_collector_duration_seconds windows_exporter: Duration of an performancecounter child collection.
# TYPE windows_performancecounter_collector_duration_seconds gauge
windows_performancecounter_collector_duration_seconds\{collector="processor_information"} [0-9.e+-]+
# HELP windows_performancecounter_collector_success windows_exporter: Whether a performancecounter child collector was successful.
# TYPE windows_performancecounter_collector_success gauge
windows_performancecounter_collector_success\{collector="processor_information"} 1
# HELP windows_performancecounter_processor_information_processor_time windows_exporter: custom Performance Counter metric
# TYPE windows_performancecounter_processor_information_processor_time counter
windows_performancecounter_processor_information_processor_time\{core="0,0",state="active"} [0-9.e+-]+
windows_performancecounter_processor_information_processor_time\{core="0,0",state="idle"} [0-9.e+-]+
.*`),
		},
		{
			name:          "processor_information_formatted",
			object:        "Processor Information",
			counterType:   pdh.CounterTypeFormatted,
			instances:     []string{"*"},
			instanceLabel: "core",
			buildErr:      "",
			counters:      []performancecounter.Counter{{Name: "% Processor Time", Metric: "windows_performancecounter_processor_information_processor_time", Labels: map[string]string{"state": "active"}}, {Name: "% Idle Time", Metric: "windows_performancecounter_processor_information_processor_time", Labels: map[string]string{"state": "idle"}}},
			expectedMetrics: regexp.MustCompile(`^# HELP windows_performancecounter_collector_duration_seconds windows_exporter: Duration of an performancecounter child collection.
# TYPE windows_performancecounter_collector_duration_seconds gauge
windows_performancecounter_collector_duration_seconds\{collector="processor_information_formatted"} [0-9.e+-]+
# HELP windows_performancecounter_collector_success windows_exporter: Whether a performancecounter child collector was successful.
# TYPE windows_performancecounter_collector_success gauge
windows_performancecounter_collector_success\{collector="processor_information_formatted"} 1
# HELP windows_performancecounter_processor_information_processor_time windows_exporter: custom Performance Counter metric
# TYPE windows_performancecounter_processor_information_processor_time gauge
windows_performancecounter_processor_information_processor_time\{core="0,0",state="active"} [0-9]+
windows_performancecounter_processor_information_processor_time\{core="0,0",state="idle"} [0-9]+
.*`),
		},
		{
			name:            "",
			object:          "Processor Information",
			counterType:     pdh.CounterTypeRaw,
			instances:       nil,
			instanceLabel:   "",
			buildErr:        "object name is required",
			counters:        nil,
			expectedMetrics: nil,
		},
		{
			name:            "double_counter",
			object:          "Memory",
			counterType:     pdh.CounterTypeRaw,
			instances:       nil,
			buildErr:        "counter name Available Bytes is duplicated",
			counters:        []performancecounter.Counter{{Name: "Available Bytes", Type: "gauge"}, {Name: "Available Bytes", Type: "gauge"}},
			expectedMetrics: nil,
		},
		{
			// PDH counter names are case-insensitive, but both names sanitize to the same Go identifier.
			name:        "counter names with the same sanitized name",
			object:      "Memory",
			counterType: pdh.CounterTypeRaw,
			instances:   nil,
			buildErr:    "",
			counters: []performancecounter.Counter{
				{Name: "Available Bytes", Type: "gauge", Metric: "windows_performancecounter_memory_available_bytes"},
				{Name: "available bytes", Type: "gauge", Metric: "windows_performancecounter_memory_available_bytes_lowercase"},
			},
			expectedMetrics: regexp.MustCompile(`^# HELP windows_performancecounter_collector_duration_seconds windows_exporter: Duration of an performancecounter child collection.
# TYPE windows_performancecounter_collector_duration_seconds gauge
windows_performancecounter_collector_duration_seconds\{collector="counter names with the same sanitized name"} [0-9.e+-]+
# HELP windows_performancecounter_collector_success windows_exporter: Whether a performancecounter child collector was successful.
# TYPE windows_performancecounter_collector_success gauge
windows_performancecounter_collector_success\{collector="counter names with the same sanitized name"} 1
# HELP windows_performancecounter_memory_available_bytes windows_exporter: custom Performance Counter metric
# TYPE windows_performancecounter_memory_available_bytes gauge
windows_performancecounter_memory_available_bytes [0-9.e+-]+
# HELP windows_performancecounter_memory_available_bytes_lowercase windows_exporter: custom Performance Counter metric
# TYPE windows_performancecounter_memory_available_bytes_lowercase gauge
windows_performancecounter_memory_available_bytes_lowercase [0-9.e+-]+`),
		},
		{
			name:            "counter with spaces and brackets",
			object:          "invalid",
			counterType:     pdh.CounterTypeRaw,
			instances:       nil,
			buildErr:        pdh.NewPdhError(pdh.CstatusNoObject).Error(),
			counters:        []performancecounter.Counter{{Name: "Total Memory Usage --- Non-Paged Pool", Type: "counter"}, {Name: "Max Session Input Delay (ms)", Type: "counter"}},
			expectedMetrics: nil,
		},
		{
			name:            "invalid counter type",
			object:          "invalid",
			counterType:     "invalid",
			instances:       nil,
			buildErr:        "invalid result type: ",
			counters:        []performancecounter.Counter{{Name: "Total Memory Usage --- Non-Paged Pool", Type: "counter"}, {Name: "Max Session Input Delay (ms)", Type: "counter"}},
			expectedMetrics: nil,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			perfDataCollector := performancecounter.New(&performancecounter.Config{
				Objects: []performancecounter.Object{
					{
						Name:          tc.name,
						Object:        tc.object,
						Type:          tc.counterType,
						Instances:     tc.instances,
						InstanceLabel: tc.instanceLabel,
						Counters:      tc.counters,
					},
				},
			})

			t.Cleanup(func() { require.NoError(t, perfDataCollector.Close()) })

			logger := slog.New(slog.DiscardHandler)
			err := perfDataCollector.Build(logger, nil)

			if tc.buildErr != "" {
				require.ErrorContains(t, err, tc.buildErr)

				return
			}

			require.NoError(t, err)

			registry := prometheus.NewRegistry()
			registry.MustRegister(collectorAdapter{*perfDataCollector})

			rw := httptest.NewRecorder()
			promhttp.HandlerFor(registry, promhttp.HandlerOpts{}).ServeHTTP(rw, &http.Request{})
			require.Equal(t, http.StatusOK, rw.Code)
			got := rw.Body.String()

			require.NotEmpty(t, got)
			require.NotEmpty(t, tc.expectedMetrics)
			require.Regexp(t, tc.expectedMetrics, got)
		})
	}
}

// TestBuildInvalidConfig verifies that invalid configurations are reported as
// errors. Build runs in a goroutine without recover during exporter startup,
// so a panic here would crash the exporter.
func TestBuildInvalidConfig(t *testing.T) {
	t.Parallel()

	// Build writes default metric names into the config, so every case needs its own objects.
	availableBytes := func() performancecounter.Object {
		return performancecounter.Object{
			Name:     "available_bytes",
			Object:   "Memory",
			Counters: []performancecounter.Counter{{Name: "Available Bytes"}},
		}
	}
	committedBytes := func() performancecounter.Object {
		return performancecounter.Object{
			Name:     "committed_bytes",
			Object:   "Memory",
			Counters: []performancecounter.Counter{{Name: "Committed Bytes"}},
		}
	}

	for _, tc := range []struct {
		name     string
		objects  []performancecounter.Object
		buildErr string
	}{
		{
			name: "empty counter name in first object",
			objects: []performancecounter.Object{
				{Name: "memory", Object: "Memory", Counters: []performancecounter.Counter{{Name: ""}, {Name: "Available Bytes"}}},
				committedBytes(),
			},
			buildErr: "object memory: counter name is required",
		},
		{
			name: "empty counter name in third object",
			objects: []performancecounter.Object{
				availableBytes(),
				committedBytes(),
				{Name: "memory", Object: "Memory", Counters: []performancecounter.Counter{{Name: ""}}},
			},
			buildErr: "object memory: counter name is required",
		},
		{
			name: "counter name starting with a digit",
			objects: []performancecounter.Object{
				{Name: "memory", Object: "Memory", Counters: []performancecounter.Counter{{Name: "1 Available Bytes"}}},
			},
			buildErr: pdh.NewPdhError(pdh.CstatusNoCounter).Error(),
		},
		{
			name: "counter names with the same default metric name",
			objects: []performancecounter.Object{
				{Name: "memory", Object: "Memory", Counters: []performancecounter.Counter{{Name: "Available Bytes"}, {Name: "available bytes"}}},
			},
			buildErr: "object memory: counters Available Bytes and available bytes produce identical series windows_performancecounter_memory_available_bytes",
		},
		{
			name: "counters with the same metric name",
			objects: []performancecounter.Object{
				{Name: "memory", Object: "Memory", Counters: []performancecounter.Counter{
					{Name: "Available Bytes", Metric: "windows_memory_bytes"},
					{Name: "Committed Bytes", Metric: "windows_memory_bytes"},
				}},
			},
			buildErr: "object memory: counters Available Bytes and Committed Bytes produce identical series windows_memory_bytes",
		},
		{
			name: "counters with the same metric name and labels",
			objects: []performancecounter.Object{
				{Name: "processor", Object: "Processor Information", Instances: pdh.InstancesAll, Counters: []performancecounter.Counter{
					{Name: "% Processor Time", Metric: "windows_processor_time", Labels: map[string]string{"state": "active"}},
					{Name: "% Idle Time", Metric: "windows_processor_time", Labels: map[string]string{"state": "idle"}},
					{Name: "% User Time", Metric: "windows_processor_time", Labels: map[string]string{"state": "active"}},
				}},
			},
			buildErr: "object processor: counters % Processor Time and % User Time produce identical series windows_processor_time",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := performancecounter.New(&performancecounter.Config{Objects: tc.objects})

			t.Cleanup(func() { require.NoError(t, c.Close()) })

			var err error

			require.NotPanics(t, func() { err = c.Build(slog.New(slog.DiscardHandler), nil) })
			require.ErrorContains(t, err, tc.buildErr)
		})
	}
}

func TestCollectorClose(t *testing.T) {
	t.Parallel()

	c := performancecounter.New(&performancecounter.Config{
		Objects: []performancecounter.Object{{
			Name: "memory", Object: "Memory",
			Counters: []performancecounter.Counter{{Name: "Available Bytes", Type: "gauge"}},
		}},
	})
	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), nil))
	t.Cleanup(func() { require.NoError(t, c.Close()) })

	metrics := make(chan prometheus.Metric, 10)
	require.NoError(t, c.Collect(metrics, 0))
	require.NoError(t, c.Close())

	// A closed collector must no longer hold usable native counter queries.
	require.ErrorIs(t, c.Collect(metrics, 0), pdh.ErrPerformanceCounterNotInitialized)
}

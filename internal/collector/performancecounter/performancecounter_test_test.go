// Copyright 2024 The Prometheus Authors
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
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/performancecounter"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type collectorAdapter struct {
	performancecounter.Collector
}

// Describe implements the prometheus.Collector interface.
func (a collectorAdapter) Describe(_ chan<- *prometheus.Desc) {}

// Collect implements the prometheus.Collector interface.
func (a collectorAdapter) Collect(ch chan<- prometheus.Metric) {
	if err := a.Collector.Collect(ch); err != nil {
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

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			err := perfDataCollector.Build(logger, nil)

			if tc.buildErr != "" {
				require.ErrorContains(t, err, tc.buildErr)

				return
			}

			require.NoError(t, err)

			registry := prometheus.NewRegistry()
			registry.MustRegister(collectorAdapter{*perfDataCollector})

			rw := httptest.NewRecorder()
			promhttp.HandlerFor(registry, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError}).ServeHTTP(rw, &http.Request{})
			got := rw.Body.String()

			assert.NotEmpty(t, got)
			require.NotEmpty(t, tc.expectedMetrics)
			assert.Regexp(t, tc.expectedMetrics, got)
		})
	}
}

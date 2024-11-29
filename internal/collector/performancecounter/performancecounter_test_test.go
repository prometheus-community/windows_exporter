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
		object          string
		instances       []string
		instanceLabel   string
		counters        []performancecounter.Counter
		expectedMetrics *regexp.Regexp
	}{
		{
			object:          "Memory",
			instances:       nil,
			counters:        []performancecounter.Counter{{Name: "Available Bytes", Type: "gauge"}},
			expectedMetrics: regexp.MustCompile(`^# HELP windows_perfdata_memory_available_bytes \S*\s*# TYPE windows_perfdata_memory_available_bytes gauge\s*windows_perfdata_memory_available_bytes \d`),
		},
		{
			object:          "Process",
			instances:       []string{"*"},
			counters:        []performancecounter.Counter{{Name: "Thread Count", Type: "counter"}},
			expectedMetrics: regexp.MustCompile(`^# HELP windows_perfdata_process_thread_count \S*\s*# TYPE windows_perfdata_process_thread_count counter\s*windows_perfdata_process_thread_count\{instance=".+"} \d`),
		},
		{
			object:          "Processor Information",
			instances:       []string{"*"},
			instanceLabel:   "core",
			counters:        []performancecounter.Counter{{Name: "% Processor Time", Metric: "windows_perfdata_processor_information_processor_time", Labels: map[string]string{"state": "active"}}, {Name: "% Idle Time", Metric: "windows_perfdata_processor_information_processor_time", Labels: map[string]string{"state": "idle"}}},
			expectedMetrics: regexp.MustCompile(`^# HELP windows_perfdata_processor_information_processor_time\s+# TYPE windows_perfdata_processor_information_processor_time counter\s+windows_perfdata_processor_information_processor_time\{core="0,0",state="active"} [0-9.e+]+\s+windows_perfdata_processor_information_processor_time\{core="0,0",state="idle"} [0-9.e+]+`),
		},
	} {
		t.Run(tc.object, func(t *testing.T) {
			t.Parallel()

			perfDataCollector := performancecounter.New(&performancecounter.Config{
				Objects: []performancecounter.Object{
					{
						Object:        tc.object,
						Instances:     tc.instances,
						InstanceLabel: tc.instanceLabel,
						Counters:      tc.counters,
					},
				},
			})

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			err := perfDataCollector.Build(logger, nil)
			require.NoError(t, err)

			registry := prometheus.NewRegistry()
			registry.MustRegister(collectorAdapter{*perfDataCollector})

			rw := httptest.NewRecorder()
			promhttp.HandlerFor(registry, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError}).ServeHTTP(rw, &http.Request{})
			got := rw.Body.String()

			assert.NotEmpty(t, got)
			assert.Regexp(t, tc.expectedMetrics, got)
		})
	}
}

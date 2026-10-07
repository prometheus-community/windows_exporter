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

package wmi_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/collector/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/require"
)

// collectorAdapter bridges wmi.Collector's Collect(ch, time.Duration) error
// signature to the prometheus.Collector interface. The Collect error is
// discarded: this test deliberately configures failing queries, which are
// reported via query_success=0 while every other metric is still emitted.
type collectorAdapter struct {
	*wmi.Collector
}

// Describe implements the prometheus.Collector interface.
func (collectorAdapter) Describe(chan<- *prometheus.Desc) {}

// Collect implements the prometheus.Collector interface.
func (a collectorAdapter) Collect(ch chan<- prometheus.Metric) {
	_ = a.Collector.Collect(ch, 5*time.Second)
}

// TestCollectorMetrics asserts the property→Desc→metric wiring end to end
// against WMI instances whose values are identical on every Windows host: the
// System Idle Process (Handle 0, PID 0, parent PID 0) and the primary
// operating system. It covers auto-generated and explicit metric names, gauge
// vs counter, label properties with default and explicit label names,
// constant labels, boolean conversion, and the query_success semantics for an
// unknown class and a non-numeric property.
func TestCollectorMetrics(t *testing.T) {
	t.Parallel()

	c := wmi.New(&wmi.Config{
		Queries: []wmi.Query{
			{
				Name:  "idle",
				Class: "Win32_Process",
				Where: "Handle = 0",
				LabelProperties: []wmi.LabelProperty{
					{Name: "Name", Label: "process"},
					// Default label name: the sanitized, lowercased property name.
					{Name: "Handle"},
				},
				Properties: []wmi.Property{
					// Auto-named gauge → windows_wmi_idle_processid.
					{Name: "ProcessId"},
					// Explicit metric + counter type + help + constant label.
					{
						Name:   "ParentProcessId",
						Metric: "windows_wmi_test_parent_pid",
						Help:   "custom wmi help text",
						Type:   "counter",
						Labels: map[string]string{"foo": "bar"},
					},
				},
			},
			{
				Name:       "os",
				Class:      "Win32_OperatingSystem",
				Properties: []wmi.Property{{Name: "Primary"}},
			},
			{
				Name:       "missing_class",
				Class:      "Win32_WindowsExporterDoesNotExist",
				Properties: []wmi.Property{{Name: "Value"}},
			},
			{
				// Caption is a string, which cannot be exported as a sample.
				Name:       "string_property",
				Class:      "Win32_OperatingSystem",
				Properties: []wmi.Property{{Name: "Caption"}},
			},
		},
	})

	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), newSession(t)))

	reg := prometheus.NewRegistry()
	reg.MustRegister(collectorAdapter{c})

	rw := httptest.NewRecorder()
	promhttp.HandlerFor(reg, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError}).ServeHTTP(rw, &http.Request{})

	// Durations vary between runs; their presence is asserted separately.
	var lines []string

	for line := range strings.SplitSeq(rw.Body.String(), "\n") {
		if !strings.Contains(line, "windows_wmi_query_duration_seconds") {
			lines = append(lines, line)
		}
	}

	got := strings.Join(lines, "\n")

	expected := `# HELP windows_wmi_idle_processid windows_exporter: custom WMI metric
# TYPE windows_wmi_idle_processid gauge
windows_wmi_idle_processid{handle="0",process="System Idle Process"} 0
# HELP windows_wmi_os_primary windows_exporter: custom WMI metric
# TYPE windows_wmi_os_primary gauge
windows_wmi_os_primary 1
# HELP windows_wmi_query_success Whether the WMI query and all of its configured properties could be read successfully.
# TYPE windows_wmi_query_success gauge
windows_wmi_query_success{name="idle"} 1
windows_wmi_query_success{name="missing_class"} 0
windows_wmi_query_success{name="os"} 1
windows_wmi_query_success{name="string_property"} 0
# HELP windows_wmi_test_parent_pid custom wmi help text
# TYPE windows_wmi_test_parent_pid counter
windows_wmi_test_parent_pid{foo="bar",handle="0",process="System Idle Process"} 0
`

	require.Equal(t, expected, got)
	require.Regexp(t, `(?m)^windows_wmi_query_duration_seconds\{name="idle"\} \d`, rw.Body.String())
}

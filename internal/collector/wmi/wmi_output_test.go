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
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/collector/wmi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/require"
)

// collectorAdapter bridges wmi.Collector's Collect(ch, time.Duration) error
// signature to the prometheus.Collector interface. The Collect error is
// discarded: these tests deliberately configure failing queries, which are
// reported via query_success=0 while every other metric is still emitted.
type collectorAdapter struct {
	*wmi.Collector

	maxScrapeDuration time.Duration
}

// Describe implements the prometheus.Collector interface.
func (collectorAdapter) Describe(chan<- *prometheus.Desc) {}

// Collect implements the prometheus.Collector interface.
func (a collectorAdapter) Collect(ch chan<- prometheus.Metric) {
	_ = a.Collector.Collect(ch, a.maxScrapeDuration)
}

// scrape builds c against a live MI session and returns its exposition.
func scrape(t *testing.T, c *wmi.Collector) string {
	t.Helper()

	return scrapeWithTimeout(t, c, 5*time.Second)
}

// scrapeWithTimeout is scrape with a custom scrape timeout.
func scrapeWithTimeout(t *testing.T, c *wmi.Collector, maxScrapeDuration time.Duration) string {
	t.Helper()

	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), newSession(t)))

	reg := prometheus.NewRegistry()
	reg.MustRegister(collectorAdapter{c, maxScrapeDuration})

	rw := httptest.NewRecorder()
	promhttp.HandlerFor(reg, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError}).ServeHTTP(rw, &http.Request{})

	return rw.Body.String()
}

// collect builds c against a live MI session and returns the number of emitted
// metrics and the Collect error.
func collect(t *testing.T, c *wmi.Collector) (int, error) {
	t.Helper()

	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), newSession(t)))

	ch := make(chan prometheus.Metric, 1000)
	err := c.Collect(ch, 5*time.Second)
	close(ch)

	return len(ch), err
}

// TestCollectorMetrics asserts the property→Desc→metric wiring end to end
// against WMI instances whose values are identical on every Windows host: the
// System Idle Process (Handle 0) and the System process (Handle 4), both with
// parent PID 0 and without an executable path or install date, and the
// primary operating system.
//
// It covers multiple instances per query, auto-generated and explicit metric
// names, gauge vs counter, label properties with default and explicit label
// names, a property used as both label and value, null labels and values,
// constant labels, boolean conversion, a metric name shared by two queries,
// and the query_success semantics for an unknown class, a non-numeric
// property next to a numeric one, and a label property of unsupported type.
func TestCollectorMetrics(t *testing.T) {
	t.Parallel()

	got := scrape(t, wmi.New(&wmi.Config{
		Queries: []wmi.Query{
			{
				Name:  "kernel",
				Class: "Win32_Process",
				Where: "Handle = 0 OR Handle = 4",
				LabelProperties: []wmi.LabelProperty{
					{Name: "Name", Label: "process"},
					// Default label name: the sanitized, lowercased property name.
					{Name: "Handle"},
					// Null for both processes, so the label value is empty.
					{Name: "ExecutablePath", Label: "path"},
					// Also exported as a value below; selected only once.
					{Name: "ProcessId", Label: "pid"},
				},
				Properties: []wmi.Property{
					// Auto-named gauge → windows_wmi_kernel_processid.
					{Name: "ProcessId"},
					// Explicit metric + counter type + help + constant label.
					{
						Name:   "ParentProcessId",
						Metric: "windows_wmi_test_parent_pid",
						Help:   "custom wmi help text",
						Type:   "counter",
						Labels: map[string]string{"foo": "bar"},
					},
					// Null for every process, so no sample is exported.
					{Name: "InstallDate"},
				},
			},
			{
				Name:  "os",
				Class: "Win32_OperatingSystem",
				// Whitespace only, so no WHERE clause is generated.
				Where:      " \t ",
				Properties: []wmi.Property{{Name: "Primary"}},
			},
			{
				Name:       "shared_a",
				Class:      "Win32_OperatingSystem",
				Properties: []wmi.Property{{Name: "Primary", Metric: "windows_wmi_test_shared", Labels: map[string]string{"query": "a"}}},
			},
			{
				Name:       "shared_b",
				Class:      "Win32_OperatingSystem",
				Properties: []wmi.Property{{Name: "Primary", Metric: "windows_wmi_test_shared", Labels: map[string]string{"query": "b"}}},
			},
			{
				Name:       "missing_class",
				Class:      "Win32_WindowsExporterDoesNotExist",
				Properties: []wmi.Property{{Name: "Value"}},
			},
			{
				// Caption is a string, which cannot be exported as a sample.
				// Primary is still exported.
				Name:       "partial",
				Class:      "Win32_OperatingSystem",
				Properties: []wmi.Property{{Name: "Caption"}, {Name: "Primary"}},
			},
			{
				// A datetime cannot be used as a label value.
				Name:            "unsupported_label",
				Class:           "Win32_OperatingSystem",
				LabelProperties: []wmi.LabelProperty{{Name: "LastBootUpTime"}},
				Properties:      []wmi.Property{{Name: "Primary"}},
			},
		},
	}))

	// Durations vary between runs; their presence is asserted separately.
	var lines []string

	for line := range strings.SplitSeq(got, "\n") {
		if !strings.Contains(line, "windows_wmi_query_duration_seconds") {
			lines = append(lines, line)
		}
	}

	expected := `# HELP windows_wmi_kernel_processid windows_exporter: custom WMI metric
# TYPE windows_wmi_kernel_processid gauge
windows_wmi_kernel_processid{handle="0",path="",pid="0",process="System Idle Process"} 0
windows_wmi_kernel_processid{handle="4",path="",pid="4",process="System"} 4
# HELP windows_wmi_os_primary windows_exporter: custom WMI metric
# TYPE windows_wmi_os_primary gauge
windows_wmi_os_primary 1
# HELP windows_wmi_partial_primary windows_exporter: custom WMI metric
# TYPE windows_wmi_partial_primary gauge
windows_wmi_partial_primary 1
# HELP windows_wmi_query_success Whether the WMI query and all of its configured properties could be read successfully.
# TYPE windows_wmi_query_success gauge
windows_wmi_query_success{name="kernel"} 1
windows_wmi_query_success{name="missing_class"} 0
windows_wmi_query_success{name="os"} 1
windows_wmi_query_success{name="partial"} 0
windows_wmi_query_success{name="shared_a"} 1
windows_wmi_query_success{name="shared_b"} 1
windows_wmi_query_success{name="unsupported_label"} 0
# HELP windows_wmi_test_parent_pid custom wmi help text
# TYPE windows_wmi_test_parent_pid counter
windows_wmi_test_parent_pid{foo="bar",handle="0",path="",pid="0",process="System Idle Process"} 0
windows_wmi_test_parent_pid{foo="bar",handle="4",path="",pid="4",process="System"} 0
# HELP windows_wmi_test_shared windows_exporter: custom WMI metric
# TYPE windows_wmi_test_shared gauge
windows_wmi_test_shared{query="a"} 1
windows_wmi_test_shared{query="b"} 1
`

	require.Equal(t, expected, strings.Join(lines, "\n"))

	for _, name := range []string{"kernel", "missing_class", "os", "partial", "shared_a", "shared_b", "unsupported_label"} {
		require.Regexp(t, `(?m)^windows_wmi_query_duration_seconds\{name="`+name+`"\} \d`, got)
	}
}

// TestCollectorDatetime asserts that a DATETIME timestamp is exported as
// seconds since the Unix epoch, including the UTC offset WMI reports.
func TestCollectorDatetime(t *testing.T) {
	t.Parallel()

	got := scrape(t, wmi.New(&wmi.Config{
		Queries: []wmi.Query{
			{
				Name:       "os",
				Class:      "Win32_OperatingSystem",
				Properties: []wmi.Property{{Name: "LocalDateTime", Metric: "windows_wmi_test_now_timestamp_seconds"}},
			},
		},
	}))

	match := regexp.MustCompile(`(?m)^windows_wmi_test_now_timestamp_seconds (\S+)$`).FindStringSubmatch(got)
	require.Len(t, match, 2, got)

	value, err := strconv.ParseFloat(match[1], 64)
	require.NoError(t, err)
	require.InDelta(t, float64(time.Now().Unix()), value, 60)
}

// TestCollectPropertyErrorReportedOnce verifies that a property failing for
// every instance is reported once per scrape, not once per instance, and that
// the other properties of those instances are still exported.
func TestCollectPropertyErrorReportedOnce(t *testing.T) {
	t.Parallel()

	metrics, err := collect(t, wmi.New(&wmi.Config{
		Queries: []wmi.Query{
			{
				Name:            "kernel",
				Class:           "Win32_Process",
				Where:           "Handle = 0 OR Handle = 4",
				LabelProperties: []wmi.LabelProperty{{Name: "Handle"}},
				Properties:      []wmi.Property{{Name: "Name"}, {Name: "ProcessId"}},
			},
		},
	}))

	// Like the registry collector, a failing query is reported via
	// query_success and does not fail the whole collector.
	require.ErrorIs(t, err, types.ErrNoData)
	require.Equal(t, 1, strings.Count(err.Error(), "failed to read property Name"), err.Error())

	// Two ProcessId samples plus query_success and query_duration_seconds.
	require.Equal(t, 4, metrics)
}

// TestCollectScrapeTimeout verifies that queries are not started once the
// scrape timeout is used up, and are reported as failed.
func TestCollectScrapeTimeout(t *testing.T) {
	t.Parallel()

	c := wmi.New(&wmi.Config{
		Queries: []wmi.Query{
			{Name: "a", Class: "Win32_OperatingSystem", Properties: []wmi.Property{{Name: "Primary"}}},
			{Name: "b", Class: "Win32_OperatingSystem", Properties: []wmi.Property{{Name: "Primary"}}, LabelProperties: []wmi.LabelProperty{{Name: "Caption"}}},
		},
	})

	got := scrapeWithTimeout(t, c, time.Nanosecond)

	require.Contains(t, got, `windows_wmi_query_success{name="a"} 0`)
	require.Contains(t, got, `windows_wmi_query_success{name="b"} 0`)
	require.NotContains(t, got, "windows_wmi_a_primary")
	require.NotContains(t, got, "windows_wmi_b_primary")

	ch := make(chan prometheus.Metric, 10)
	err := c.Collect(ch, time.Nanosecond)
	close(ch)

	require.ErrorIs(t, err, types.ErrNoData)
	require.Equal(t, 2, strings.Count(err.Error(), "scrape timeout exceeded"), err.Error())
}

// TestNewWithFlags verifies that the queries flag is parsed as YAML and JSON,
// and that invalid input is rejected at parse time.
func TestNewWithFlags(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		queries string
	}{
		{
			name: "yaml",
			queries: `- name: os
  class: Win32_OperatingSystem
  properties:
    - name: Primary`,
		},
		{
			name:    "json",
			queries: `[{"name":"os","class":"Win32_OperatingSystem","properties":[{"name":"Primary"}]}]`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := kingpin.New("test", "")
			c := wmi.NewWithFlags(app)

			// The = form is required, because kingpin reads a separate value
			// starting with "-", like a YAML list, as another flag.
			_, err := app.Parse([]string{"--collector.wmi.queries=" + tc.queries})
			require.NoError(t, err)

			metrics, err := collect(t, c)
			require.NoError(t, err)

			// Primary plus query_success and query_duration_seconds.
			require.Equal(t, 3, metrics)
		})
	}

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()

		app := kingpin.New("test", "")
		_ = wmi.NewWithFlags(app)

		_, err := app.Parse([]string{"--collector.wmi.queries", `[{"name": `})
		require.Error(t, err)
	})

	t.Run("unknown field", func(t *testing.T) {
		t.Parallel()

		app := kingpin.New("test", "")
		_ = wmi.NewWithFlags(app)

		_, err := app.Parse([]string{"--collector.wmi.queries", `[{"name":"os","class":"Win32_OperatingSystem","propertys":[{"name":"Primary"}]}]`})
		require.ErrorContains(t, err, "propertys")
	})
}

// TestCollectConcurrent runs overlapping scrapes, as concurrent HTTP requests
// do, so the race detector in CI can catch state shared between them.
func TestCollectConcurrent(t *testing.T) {
	t.Parallel()

	c := wmi.New(&wmi.Config{
		Queries: []wmi.Query{
			{
				Name:            "kernel",
				Class:           "Win32_Process",
				Where:           "Handle = 0 OR Handle = 4",
				LabelProperties: []wmi.LabelProperty{{Name: "Handle"}},
				Properties:      []wmi.Property{{Name: "ProcessId"}, {Name: "Name"}},
			},
		},
	})

	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), newSession(t)))

	var wg sync.WaitGroup

	for range 8 {
		wg.Go(func() {
			ch := make(chan prometheus.Metric, 100)
			err := c.Collect(ch, 5*time.Second)
			close(ch)

			// Name is a string, so every scrape reports exactly that error.
			if err == nil || len(ch) != 4 {
				t.Errorf("unexpected scrape result: %d metrics, err=%v", len(ch), err)
			}
		})
	}

	wg.Wait()
}

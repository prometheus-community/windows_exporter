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

package httphandler_test

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/synctest"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/httphandler"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/pkg/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestMetricsHTTPHandler(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		query    string
		options  *httphandler.Options
		status   int
		included []string
		excluded []string
	}{
		{name: "default", status: http.StatusOK, included: []string{"windows_test_first 42", "windows_test_second 42", "go_goroutines", "windows_exporter_build_info", "go_sched_goroutines_running_goroutines", "go_sched_goroutines_waiting_goroutines", "go_sched_goroutines_created_goroutines_total", "go_sched_threads_total_threads", "go_sched_latencies_seconds_bucket"}},
		{name: "disable exporter metrics", options: &httphandler.Options{DisableExporterMetrics: true, TimeoutMargin: 0.5}, status: http.StatusOK, included: []string{"windows_test_first 42", "windows_exporter_build_info"}, excluded: []string{"go_goroutines", "process_cpu_seconds_total", "go_sched_goroutines_running_goroutines"}},
		{name: "filter", query: "?collect[]=first", status: http.StatusOK, included: []string{"windows_test_first 42"}, excluded: []string{"windows_test_second"}},
		{name: "multiple collectors", query: "?collect[]=first&collect[]=second", status: http.StatusOK, included: []string{"windows_test_first 42", "windows_test_second 42"}},
		{name: "unknown collector", query: "?collect[]=missing", status: http.StatusBadRequest, included: []string{"unknown collector missing"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			collection := collector.New(collector.Map{
				"first":  &testCollector{name: "first"},
				"second": &testCollector{name: "second"},
			})
			handler := httphandler.New(slog.New(slog.DiscardHandler), collection, tc.options)
			request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/metrics"+tc.query, nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			require.Equal(t, tc.status, response.Code)

			for _, value := range tc.included {
				require.Contains(t, response.Body.String(), value)
			}

			for _, value := range tc.excluded {
				require.NotContains(t, response.Body.String(), value)
			}

			// Filtering one request must not remove collectors from later requests.
			if tc.query == "?collect[]=first" {
				response = httptest.NewRecorder()
				handler.ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/metrics", nil))
				require.Equal(t, http.StatusOK, response.Code)
				require.Contains(t, response.Body.String(), "windows_test_second 42")
			}
		})
	}
}

func TestMetricsHTTPHandlerCollectorStatus(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		err        error
		panicValue bool
		success    string
	}{
		{name: "success", success: "1"},
		{name: "failure", err: errors.New("collector failed"), success: "0"},
		{name: "no data", err: types.ErrNoData, success: "1"},
		{name: "panic", panicValue: true, success: "0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			collection := collector.New(collector.Map{"test": &testCollector{name: "test", err: tc.err, panicValue: tc.panicValue}})
			handler := httphandler.New(slog.New(slog.DiscardHandler), collection, &httphandler.Options{DisableExporterMetrics: true})
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/metrics", nil))

			require.Equal(t, http.StatusOK, response.Code)
			require.Contains(t, response.Body.String(), `windows_exporter_collector_success{collector="test"} `+tc.success)
			require.Contains(t, response.Body.String(), `windows_exporter_collector_timeout{collector="test"} 0`)
		})
	}
}

// TestMetricsHTTPHandlerNameCollision checks a collector metric that has the name of an exporter metric,
// like a textfile metric named go_goroutines.
func TestMetricsHTTPHandlerNameCollision(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		options  *httphandler.Options
		included []string
		excluded []string
	}{
		{
			name: "exporter metrics",
			// The collector metric has another help text, so it is dropped and the other metrics are kept.
			included: []string{"# HELP go_goroutines Number of goroutines that currently exist.", "windows_test_first 42", "windows_exporter_build_info"},
			excluded: []string{"go_goroutines 7", "Collector goroutines"},
		},
		{
			name:     "disable exporter metrics",
			options:  &httphandler.Options{DisableExporterMetrics: true, TimeoutMargin: 0.5},
			included: []string{"# HELP go_goroutines Collector goroutines", "go_goroutines 7", "windows_test_first 42", "windows_exporter_build_info"},
			excluded: []string{"Number of goroutines that currently exist."},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			collection := collector.New(collector.Map{
				"first":     &testCollector{name: "first"},
				"collision": &collisionCollector{},
			})
			handler := httphandler.New(slog.New(slog.DiscardHandler), collection, tc.options)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/metrics", nil))

			require.Equal(t, http.StatusOK, response.Code)

			for _, value := range tc.included {
				require.Contains(t, response.Body.String(), value)
			}

			for _, value := range tc.excluded {
				require.NotContains(t, response.Body.String(), value)
			}

			if tc.options == nil {
				// The gathering error of the first scrape is counted like before.
				response = httptest.NewRecorder()
				handler.ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/metrics", nil))
				require.Contains(t, response.Body.String(), `promhttp_metric_handler_errors_total{cause="gathering"} 1`)
			}
		})
	}
}

func TestMetricsHTTPHandlerScrapeTimeout(t *testing.T) {
	for _, tc := range []struct {
		name    string
		header  string
		timeout bool
	}{
		{name: "missing header"},
		{name: "invalid header", header: "invalid"},
		{name: "zero header", header: "0"},
		{name: "sufficient timeout", header: "2"},
		{name: "timeout", header: "0.1", timeout: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// The collector's delay and deadline use synthetic time.
			synctest.Test(t, func(t *testing.T) {
				collection := collector.New(collector.Map{"test": &testCollector{name: "test", delay: time.Second}})
				handler := httphandler.New(slog.New(slog.DiscardHandler), collection, &httphandler.Options{DisableExporterMetrics: true})
				request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/metrics", nil)
				request.Header.Set("X-Prometheus-Scrape-Timeout-Seconds", tc.header)

				response := httptest.NewRecorder()
				handler.ServeHTTP(response, request)

				require.Equal(t, http.StatusOK, response.Code)

				if tc.timeout {
					require.Contains(t, response.Body.String(), `windows_exporter_collector_timeout{collector="test"} 1`)
					require.NotContains(t, response.Body.String(), "windows_test_test 42")

					// Finish the delayed collector and its drain goroutine before leaving the bubble.
					time.Sleep(time.Second)
					synctest.Wait()
				} else {
					require.Contains(t, response.Body.String(), `windows_exporter_collector_timeout{collector="test"} 0`)
					require.Contains(t, response.Body.String(), "windows_test_test 42")
				}
			})
		})
	}
}

func TestMetricsHTTPHandlerScrapeBudget(t *testing.T) {
	const maxScrapeTimeout = 5 * time.Minute

	for _, tc := range []struct {
		name   string
		header string
		margin float64
		want   time.Duration
	}{
		{name: "missing header", margin: 0.5, want: 9500 * time.Millisecond},
		{name: "invalid header", header: "invalid", margin: 0.5, want: 9500 * time.Millisecond},
		{name: "zero", header: "0", margin: 0.5, want: 9500 * time.Millisecond},
		{name: "negative", header: "-1", margin: 0.5, want: 9500 * time.Millisecond},
		{name: "negative infinity", header: "-Inf", margin: 0.5, want: 9500 * time.Millisecond},
		{name: "NaN", header: "NaN", margin: 0.5, want: 9500 * time.Millisecond},
		{name: "infinity", header: "+Inf", margin: 0.5, want: maxScrapeTimeout - 500*time.Millisecond},
		{name: "out of float range", header: "1e400", margin: 0.5, want: maxScrapeTimeout - 500*time.Millisecond},
		{name: "huge", header: "1e12", margin: 0.5, want: maxScrapeTimeout - 500*time.Millisecond},
		{name: "long", header: "120", margin: 0.5, want: 119500 * time.Millisecond},
		{name: "regular", header: "2", margin: 0.5, want: 1500 * time.Millisecond},
		{name: "margin larger than timeout", header: "0.3", margin: 0.5, want: 150 * time.Millisecond},
		{name: "margin takes at most half", header: "0.8", margin: 0.5, want: 400 * time.Millisecond},
		{name: "tiny", header: "0.000001", margin: 0.5, want: 10 * time.Millisecond},
		{name: "no margin", header: "2", want: 2 * time.Second},
		{name: "negative margin", header: "2", margin: -1, want: 2 * time.Second},
		{name: "NaN margin", header: "2", margin: math.NaN(), want: 2 * time.Second},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Synthetic time doesn't pass between the request and the collector's call.
			synctest.Test(t, func(t *testing.T) {
				test := &budgetCollector{}
				collection := collector.New(collector.Map{"test": test})
				handler := httphandler.New(slog.New(slog.DiscardHandler), collection, &httphandler.Options{DisableExporterMetrics: true, TimeoutMargin: tc.margin})
				request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/metrics", nil)

				if tc.header != "" {
					request.Header.Set("X-Prometheus-Scrape-Timeout-Seconds", tc.header)
				}

				response := httptest.NewRecorder()
				handler.ServeHTTP(response, request)

				require.Equal(t, http.StatusOK, response.Code)
				require.Contains(t, response.Body.String(), `windows_exporter_collector_success{collector="test"} 1`)
				require.Equal(t, tc.want, test.budget)
			})
		})
	}
}

func TestMetricsHTTPHandlerCanceledRequest(t *testing.T) {
	test := &budgetCollector{}
	collection := collector.New(collector.Map{"test": test})
	handler := httphandler.New(slog.New(slog.DiscardHandler), collection, &httphandler.Options{DisableExporterMetrics: true})

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequestWithContext(ctx, http.MethodGet, "/metrics", nil))

	// The client is gone, so the collectors don't run.
	require.False(t, test.called)
}

// budgetCollector records the time it is given.
type budgetCollector struct {
	called bool
	budget time.Duration
}

func (c *budgetCollector) GetName() string                           { return "test" }
func (c *budgetCollector) Build(_ *slog.Logger, _ *mi.Session) error { return nil }
func (c *budgetCollector) Close() error                              { return nil }
func (c *budgetCollector) Collect(_ chan<- prometheus.Metric, maxScrapeDuration time.Duration) error {
	c.called = true
	c.budget = maxScrapeDuration

	return nil
}

type testCollector struct {
	name       string
	err        error
	panicValue bool
	delay      time.Duration
}

func (c *testCollector) GetName() string                           { return c.name }
func (c *testCollector) Build(_ *slog.Logger, _ *mi.Session) error { return nil }
func (c *testCollector) Close() error                              { return nil }
func (c *testCollector) Collect(ch chan<- prometheus.Metric, _ time.Duration) error {
	if c.panicValue {
		panic("test collector panic")
	}

	time.Sleep(c.delay)

	if c.err != nil {
		return c.err
	}

	ch <- prometheus.MustNewConstMetric(prometheus.NewDesc("windows_test_"+c.name, "Test metric", nil, nil), prometheus.GaugeValue, 42)

	return nil
}

// collisionCollector emits a metric with the name of a Go runtime metric.
type collisionCollector struct{}

func (c *collisionCollector) GetName() string                           { return "collision" }
func (c *collisionCollector) Build(_ *slog.Logger, _ *mi.Session) error { return nil }
func (c *collisionCollector) Close() error                              { return nil }
func (c *collisionCollector) Collect(ch chan<- prometheus.Metric, _ time.Duration) error {
	ch <- prometheus.MustNewConstMetric(prometheus.NewDesc("go_goroutines", "Collector goroutines", nil, nil), prometheus.GaugeValue, 7)

	return nil
}

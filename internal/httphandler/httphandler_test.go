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
	"errors"
	"log/slog"
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
		{name: "default", status: http.StatusOK, included: []string{"windows_test_first 42", "windows_test_second 42", "go_goroutines", "windows_exporter_build_info"}},
		{name: "disable exporter metrics", options: &httphandler.Options{DisableExporterMetrics: true, TimeoutMargin: 0.5}, status: http.StatusOK, included: []string{"windows_test_first 42", "windows_exporter_build_info"}, excluded: []string{"go_goroutines", "process_cpu_seconds_total"}},
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

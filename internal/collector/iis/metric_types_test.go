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

package iis

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

type gatheredMetric struct {
	kind   dto.MetricType
	value  float64
	labels map[string]string
}

// gatherCollect registers collect with a registry and returns the gathered
// metrics by name, so the test checks names, types and labels as published.
func gatherCollect(t *testing.T, collect func(ch chan<- prometheus.Metric) error) map[string][]gatheredMetric {
	t.Helper()

	ch := make(chan prometheus.Metric, 1000)
	require.NoError(t, collect(ch))
	close(ch)

	collected := make([]prometheus.Metric, 0, len(ch))
	for metric := range ch {
		collected = append(collected, metric)
	}

	registry := prometheus.NewPedanticRegistry()
	require.NoError(t, registry.Register(collectFunc(func(ch chan<- prometheus.Metric) {
		for _, metric := range collected {
			ch <- metric
		}
	})))

	families, err := registry.Gather()
	require.NoError(t, err)

	result := make(map[string][]gatheredMetric)

	for _, family := range families {
		for _, metric := range family.GetMetric() {
			labels := make(map[string]string)
			for _, label := range metric.GetLabel() {
				labels[label.GetName()] = label.GetValue()
			}

			var value float64

			switch family.GetType() {
			case dto.MetricType_COUNTER:
				value = metric.GetCounter().GetValue()
			case dto.MetricType_GAUGE:
				value = metric.GetGauge().GetValue()
			default:
				t.Fatalf("unexpected type %s for %s", family.GetType(), family.GetName())
			}

			result[family.GetName()] = append(result[family.GetName()], gatheredMetric{kind: family.GetType(), value: value, labels: labels})
		}
	}

	return result
}

// collectFunc is an unchecked collector: it describes no metrics, so the
// registry validates the metrics it gathers instead of the descriptions.
type collectFunc func(ch chan<- prometheus.Metric)

func (collectFunc) Describe(chan<- *prometheus.Desc) {}

func (f collectFunc) Collect(ch chan<- prometheus.Metric) { f(ch) }

func requireMetric(t *testing.T, metrics map[string][]gatheredMetric, name string, kind dto.MetricType, value float64, labels map[string]string) {
	t.Helper()

	require.Len(t, metrics[name], 1, name)
	require.Equal(t, kind, metrics[name][0].kind, name)
	require.InDelta(t, value, metrics[name][0].value, 1e-9, name)

	for key, expected := range labels {
		require.Equal(t, expected, metrics[name][0].labels[key], "%s label %s", name, key)
	}
}

func TestAppPoolWASUptimeMetrics(t *testing.T) {
	t.Parallel()

	c := New(nil)
	c.buildAppPoolWASDescriptors()
	c.perfDataCollectorAppPoolWAS = cacheFixture[perfDataCounterValuesAppPoolWAS]{sources: map[string]float64{
		// internal/pdh turns the PERF_ELAPSED_TIME start times into elapsed seconds.
		"Current Application Pool Uptime":        298,
		"Total Application Pool Uptime":          600,
		"Time Since Last Worker Process Failure": 298,
		"Total Application Pool Recycles":        2,
	}}

	metrics := gatherCollect(t, c.collectAppPoolWAS)
	app := map[string]string{"app": "123_Example"}

	requireMetric(t, metrics, "windows_iis_current_application_pool_uptime_seconds", dto.MetricType_GAUGE, 298, app)
	requireMetric(t, metrics, "windows_iis_total_application_pool_uptime_seconds", dto.MetricType_GAUGE, 600, app)
	requireMetric(t, metrics, "windows_iis_time_since_last_worker_process_failure_seconds", dto.MetricType_GAUGE, 298, app)
	requireMetric(t, metrics, "windows_iis_total_application_pool_recycles", dto.MetricType_COUNTER, 2, app)
	require.NotContains(t, metrics, "windows_iis_current_application_pool_start_time")
	require.NotContains(t, metrics, "windows_iis_total_application_pool_start_time")
}

func TestHttpServiceRequestQueuesMetrics(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		maxAge     float64
		wantMaxAge bool
	}{
		// An empty queue reports math.MaxInt64 as start time, which internal/pdh
		// turns into a large negative number of seconds.
		{name: "empty queue", maxAge: -922337203685.4775, wantMaxAge: false},
		{name: "waiting request", maxAge: 1.5, wantMaxAge: true},
		{name: "request just queued", maxAge: 0, wantMaxAge: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := New(nil)
			c.buildHttpServiceRequestQueuesDescriptors()
			c.perfDataCollectorHttpServiceRequestQueues = cacheFixture[perfDataCounterValuesHttpServiceRequestQueues]{sources: map[string]float64{
				"CurrentQueueSize": 1,
				"RejectedRequests": 3,
				"MaxQueueItemAge":  tc.maxAge,
				"ArrivalRate":      67,
			}}

			metrics := gatherCollect(t, c.collectHttpServiceRequestQueues)
			app := map[string]string{"app": "123_Example"}

			requireMetric(t, metrics, "windows_iis_http_requests_current_queue_size", dto.MetricType_GAUGE, 1, app)
			requireMetric(t, metrics, "windows_iis_http_requests_rejected_total", dto.MetricType_COUNTER, 3, app)
			requireMetric(t, metrics, "windows_iis_http_requests_arrived_total", dto.MetricType_COUNTER, 67, app)

			if tc.wantMaxAge {
				requireMetric(t, metrics, "windows_iis_http_requests_max_queue_item_age_seconds", dto.MetricType_GAUGE, tc.maxAge, app)
			} else {
				require.NotContains(t, metrics, "windows_iis_http_requests_max_queue_item_age_seconds")
			}
		})
	}
}

func TestW3SVCW3WPCurrentValuesAreGauges(t *testing.T) {
	t.Parallel()

	c := New(nil)
	c.buildW3SVCW3WPDescriptors()

	sources := map[string]float64{
		"Maximum Threads Count":      256,
		"Active Requests":            4,
		"WebSocket Active Requests":  2,
		"Total HTTP Requests Served": 10,
	}
	c.w3SVCW3WPPerfDataCollector = cacheFixture[perfDataCounterValuesW3SVCW3WP]{sources: sources}
	c.w3SVCW3WPPerfDataCollectorV8 = cacheFixture[perfDataCounterValuesW3SVCW3WPV8]{sources: sources}

	metrics := gatherCollect(t, func(ch chan<- prometheus.Metric) error {
		if err := c.collectW3SVCW3WPv7(ch); err != nil {
			return err
		}

		return c.collectW3SVCW3WPv8(ch)
	})
	labels := map[string]string{"app": "Example", "pid": "123"}

	requireMetric(t, metrics, "windows_iis_worker_max_threads", dto.MetricType_GAUGE, 256, labels)
	requireMetric(t, metrics, "windows_iis_worker_current_requests", dto.MetricType_GAUGE, 4, labels)
	requireMetric(t, metrics, "windows_iis_worker_current_websocket_requests", dto.MetricType_GAUGE, 2, labels)
	requireMetric(t, metrics, "windows_iis_worker_requests_total", dto.MetricType_COUNTER, 10, labels)
}

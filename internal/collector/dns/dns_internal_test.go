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

package dns

import (
	"reflect"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func TestCollectPerfData(t *testing.T) {
	t.Parallel()

	data := perfDataCounterValues{
		AxfrRequestSent:            1,
		IxfrRequestSent:            2,
		ZoneTransferSOARequestSent: 3,
		AxfrSuccessReceived:        4,
		IxfrTCPSuccessReceived:     5,
		IxfrUDPSuccessReceived:     6,
		DynamicUpdateQueued:        93,
		RecursiveSendTimeOuts:      7,
	}

	c := New(nil)
	c.buildDescriptors()

	registry := prometheus.NewPedanticRegistry()
	require.NoError(t, registry.Register(metricCollector{collector: c, data: data}))

	families, err := registry.Gather()
	require.NoError(t, err)

	type sample struct {
		metricType dto.MetricType
		value      float64
	}

	got := make(map[string]sample)

	for _, family := range families {
		for _, metric := range family.GetMetric() {
			labels := make([]string, 0, len(metric.GetLabel()))
			for _, label := range metric.GetLabel() {
				labels = append(labels, label.GetName()+"="+label.GetValue())
			}

			value := metric.GetGauge().GetValue()
			if family.GetType() == dto.MetricType_COUNTER {
				value = metric.GetCounter().GetValue()
			}

			got[family.GetName()+"{"+strings.Join(labels, ",")+"}"] = sample{metricType: family.GetType(), value: value}
		}
	}

	for name, want := range map[string]sample{
		"windows_dns_zone_transfer_requests_sent_total{qtype=full}":                        {dto.MetricType_COUNTER, 1},
		"windows_dns_zone_transfer_requests_sent_total{qtype=incremental}":                 {dto.MetricType_COUNTER, 2},
		"windows_dns_zone_transfer_requests_sent_total{qtype=soa}":                         {dto.MetricType_COUNTER, 3},
		"windows_dns_zone_transfer_success_received_total{protocol=tcp,qtype=full}":        {dto.MetricType_COUNTER, 4},
		"windows_dns_zone_transfer_success_received_total{protocol=tcp,qtype=incremental}": {dto.MetricType_COUNTER, 5},
		"windows_dns_zone_transfer_success_received_total{protocol=udp,qtype=incremental}": {dto.MetricType_COUNTER, 6},
		"windows_dns_dynamic_updates_queued_total{}":                                       {dto.MetricType_COUNTER, 93},
		"windows_dns_recursive_query_send_timeouts_total{}":                                {dto.MetricType_COUNTER, 7},
		"windows_dns_memory_used_bytes{area=caching}":                                      {dto.MetricType_GAUGE, 0},
	} {
		require.Contains(t, got, name)
		require.Equal(t, want, got[name], name)
	}

	require.NotContains(t, got, "windows_dns_dynamic_updates_queued{}")
}

// TestPerfDataCounterNames guards against a field reading a counter of a different name.
func TestPerfDataCounterNames(t *testing.T) {
	t.Parallel()

	for field, counter := range map[string]string{
		"ZoneTransferSOARequestSent": "Zone Transfer SOA Request Sent",
		"IxfrUDPSuccessReceived":     "IXFR UDP Success Received",
		"IxfrTCPSuccessReceived":     "IXFR TCP Success Received",
		"DynamicUpdateQueued":        "Dynamic Update Queued",
		"RecursiveSendTimeOuts":      "Recursive Send TimeOuts",
	} {
		f, ok := reflect.TypeFor[perfDataCounterValues]().FieldByName(field)
		require.True(t, ok, field)
		require.Equal(t, counter, f.Tag.Get("perfdata"), field)
	}
}

type metricCollector struct {
	collector *Collector
	data      perfDataCounterValues
}

func (c metricCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c metricCollector) Collect(ch chan<- prometheus.Metric) {
	c.collector.collectPerfData(ch, c.data)
}

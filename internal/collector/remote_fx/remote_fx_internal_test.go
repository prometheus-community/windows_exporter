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

package remote_fx

import (
	"reflect"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

type expectedSample struct {
	metricType dto.MetricType
	value      float64
}

func TestCollectNetworkMetrics(t *testing.T) {
	t.Parallel()

	rows := []perfDataCounterValuesNetwork{
		{Name: "Services", LossRate: 99},
		{
			Name:                     "RDP-tcp 0",
			BaseTCPRTT:               10,
			BaseUDPRTT:               20,
			CurrentTCPBandwidth:      151159,
			CurrentTCPRTT:            30,
			CurrentUDPBandwidth:      8,
			CurrentUDPRTT:            40,
			TotalReceivedBytes:       3920173,
			TotalSentBytes:           29262741,
			UDPPacketsReceivedPersec: 5,
			UDPPacketsSentPersec:     6,
			FECRate:                  7,
			FECRateBase:              70,
			LossRate:                 2,
			LossRateBase:             200,
			RetransmissionRate:       3,
			RetransmissionRateBase:   300,
		},
	}

	expected := map[string]expectedSample{
		"windows_remote_fx_net_base_tcp_rtt_seconds":                   {dto.MetricType_GAUGE, 0.01},
		"windows_remote_fx_net_base_udp_rtt_seconds":                   {dto.MetricType_GAUGE, 0.02},
		"windows_remote_fx_net_current_tcp_bandwidth_bytes_per_second": {dto.MetricType_GAUGE, 18894875},
		"windows_remote_fx_net_current_tcp_rtt_seconds":                {dto.MetricType_GAUGE, 0.03},
		"windows_remote_fx_net_current_udp_bandwidth_bytes_per_second": {dto.MetricType_GAUGE, 1000},
		"windows_remote_fx_net_current_udp_rtt_seconds":                {dto.MetricType_GAUGE, 0.04},
		"windows_remote_fx_net_received_bytes_total":                   {dto.MetricType_COUNTER, 3920173},
		"windows_remote_fx_net_sent_bytes_total":                       {dto.MetricType_COUNTER, 29262741},
		"windows_remote_fx_net_udp_packets_received_total":             {dto.MetricType_COUNTER, 5},
		"windows_remote_fx_net_udp_packets_sent_total":                 {dto.MetricType_COUNTER, 6},
		"windows_remote_fx_net_fec_total":                              {dto.MetricType_COUNTER, 7},
		"windows_remote_fx_net_fec_base_total":                         {dto.MetricType_COUNTER, 70},
		"windows_remote_fx_net_loss_total":                             {dto.MetricType_COUNTER, 2},
		"windows_remote_fx_net_loss_base_total":                        {dto.MetricType_COUNTER, 200},
		"windows_remote_fx_net_retransmission_total":                   {dto.MetricType_COUNTER, 3},
		"windows_remote_fx_net_retransmission_base_total":              {dto.MetricType_COUNTER, 300},
	}

	c := New(nil)
	c.buildDescriptors()

	families := gather(t, func(ch chan<- prometheus.Metric) { c.collectNetworkMetrics(ch, rows) })
	require.Len(t, families, len(expected))

	for _, family := range families {
		want, ok := expected[family.GetName()]
		require.True(t, ok, "unexpected metric %s", family.GetName())
		require.Equal(t, want.metricType, family.GetType(), family.GetName())
		require.Len(t, family.GetMetric(), 1, family.GetName())

		metric := family.GetMetric()[0]
		require.Equal(t, map[string]string{"session_name": "RDP-Tcp 0"}, labels(metric))
		require.InDelta(t, want.value, value(family.GetType(), metric), 1e-9, family.GetName())
	}
}

func TestCollectGraphicsMetrics(t *testing.T) {
	t.Parallel()

	rows := []perfDataCounterValuesGraphics{
		{Name: "Console", FrameQuality: 100},
		{
			Name:                "RDP-Tcp 0",
			AverageEncodingTime: 4,
			FrameQuality:        85,
			FramesSkippedPerSecondInsufficientClientResources:  11,
			FramesSkippedPerSecondInsufficientNetworkResources: 12,
			FramesSkippedPerSecondInsufficientServerResources:  13,
			GraphicsCompressionratio:                           2,
			InputFramesPerSecond:                               27086,
			OutputFramesPerSecond:                              26142,
			SourceFramesPerSecond:                              9,
		},
	}

	expected := map[string]expectedSample{
		"windows_remote_fx_gfx_average_encoding_time_seconds": {dto.MetricType_GAUGE, 0.004},
		"windows_remote_fx_gfx_frame_quality_ratio":           {dto.MetricType_GAUGE, 0.85},
		"windows_remote_fx_gfx_graphics_compression_ratio":    {dto.MetricType_GAUGE, 2},
		"windows_remote_fx_gfx_input_frames_total":            {dto.MetricType_COUNTER, 27086},
		"windows_remote_fx_gfx_output_frames_total":           {dto.MetricType_COUNTER, 26142},
		"windows_remote_fx_gfx_source_frames_total":           {dto.MetricType_COUNTER, 9},
	}
	skipped := map[string]float64{"client": 11, "network": 12, "server": 13}

	c := New(nil)
	c.buildDescriptors()

	families := gather(t, func(ch chan<- prometheus.Metric) { c.collectGraphicsMetrics(ch, rows) })
	require.Len(t, families, len(expected)+1)

	for _, family := range families {
		if family.GetName() == "windows_remote_fx_gfx_frames_skipped_insufficient_resource_total" {
			require.Equal(t, dto.MetricType_COUNTER, family.GetType())
			require.Len(t, family.GetMetric(), len(skipped))

			for _, metric := range family.GetMetric() {
				l := labels(metric)
				require.Equal(t, "RDP-Tcp 0", l["session_name"])
				require.InDelta(t, skipped[l["resource"]], metric.GetCounter().GetValue(), 0, l["resource"])
			}

			continue
		}

		want, ok := expected[family.GetName()]
		require.True(t, ok, "unexpected metric %s", family.GetName())
		require.Equal(t, want.metricType, family.GetType(), family.GetName())
		require.Len(t, family.GetMetric(), 1, family.GetName())

		metric := family.GetMetric()[0]
		require.Equal(t, map[string]string{"session_name": "RDP-Tcp 0"}, labels(metric))
		require.InDelta(t, want.value, value(family.GetType(), metric), 1e-9, family.GetName())
	}
}

// TestFramesSkippedCounterNames guards the mapping from struct field to
// Windows counter; the client and server counters were swapped before.
func TestFramesSkippedCounterNames(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeFor[perfDataCounterValuesGraphics]()

	for _, resource := range []string{"Client", "Network", "Server"} {
		field, ok := typ.FieldByName("FramesSkippedPerSecondInsufficient" + resource + "Resources")
		require.True(t, ok, resource)
		require.Equal(t, "Frames Skipped/Second - Insufficient "+resource+" Resources", field.Tag.Get("perfdata"))
	}

	for _, counter := range []string{"FEC rate", "Loss rate", "Retransmission rate"} {
		found := 0

		for field := range reflect.TypeFor[perfDataCounterValuesNetwork]().Fields() {
			if strings.HasPrefix(field.Tag.Get("perfdata"), counter) {
				found++
			}
		}

		require.Equal(t, 2, found, "%s needs numerator and base", counter)
	}
}

func gather(t *testing.T, collect func(ch chan<- prometheus.Metric)) []*dto.MetricFamily {
	t.Helper()

	registry := prometheus.NewPedanticRegistry()
	require.NoError(t, registry.Register(metricCollector(collect)))

	families, err := registry.Gather()
	require.NoError(t, err)

	return families
}

type metricCollector func(ch chan<- prometheus.Metric)

func (c metricCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c metricCollector) Collect(ch chan<- prometheus.Metric) {
	c(ch)
}

func labels(metric *dto.Metric) map[string]string {
	result := make(map[string]string, len(metric.GetLabel()))
	for _, label := range metric.GetLabel() {
		result[label.GetName()] = label.GetValue()
	}

	return result
}

func value(metricType dto.MetricType, metric *dto.Metric) float64 {
	if metricType == dto.MetricType_COUNTER {
		return metric.GetCounter().GetValue()
	}

	return metric.GetGauge().GetValue()
}

// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package udp_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/udp"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, udp.Name, udp.NewWithFlags)
}

func TestCollector(t *testing.T) {
	metrics := testutils.TestCollector(t, udp.New, nil)

	// All four datagram totals are cumulative raw PDH values and documented as
	// counters in docs/collector.udp.md; datagram_received_total used to be
	// published as a gauge.
	for _, name := range []string{
		"windows_udp_datagram_no_port_total",
		"windows_udp_datagram_received_errors_total",
		"windows_udp_datagram_received_total",
		"windows_udp_datagram_sent_total",
	} {
		require.Contains(t, metrics, name)
		require.Equal(t, dto.MetricType_COUNTER, metrics[name].GetType(), name)
	}
}

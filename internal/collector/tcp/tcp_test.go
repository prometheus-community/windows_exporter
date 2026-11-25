// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package tcp_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/tcp"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, tcp.Name, tcp.NewWithFlags)
}

func TestCollector(t *testing.T) {
	metrics := testutils.TestCollector(t, tcp.New, nil)

	// The shared dynamic port range has no address family label.
	for _, name := range []string{
		"windows_tcp_dynamic_port_range_start",
		"windows_tcp_dynamic_port_range_size",
	} {
		require.Contains(t, metrics, name)
		require.Equal(t, dto.MetricType_GAUGE, metrics[name].GetType(), name)
		require.Len(t, metrics[name].GetMetric(), 1, name)
		require.Empty(t, metrics[name].GetMetric()[0].GetLabel(), name)
	}
}

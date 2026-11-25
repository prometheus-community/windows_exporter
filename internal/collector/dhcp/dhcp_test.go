// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package dhcp_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/dhcp"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, dhcp.Name, dhcp.NewWithFlags)
}

func TestCollector(t *testing.T) {
	metrics := testutils.TestCollector(t, dhcp.New, nil)
	testutils.RequireFixtureMetric(t, metrics, dhcp.Name, "windows_dhcp_scope_info", prometheus.Labels{"scope_name": "GitHubActions", "scope": "192.0.2.0/24"})

	// CI adds 20 reservations to the scope.
	if reserved := testutils.RequireFixtureMetric(t, metrics, dhcp.Name, "windows_dhcp_scope_reserved_address", prometheus.Labels{"scope": "192.0.2.0/24"}); reserved != nil {
		require.InDelta(t, 20, reserved.GetGauge().GetValue(), 0)
	}
}

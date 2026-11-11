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
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, dhcp.Name, dhcp.NewWithFlags)
}

func TestCollector(t *testing.T) {
	metrics := testutils.TestCollector(t, dhcp.New, nil)
	testutils.RequireFixtureMetric(t, metrics, dhcp.Name, "windows_dhcp_scope_info", prometheus.Labels{"scope_name": "GitHubActions", "scope": "192.0.2.0/24"})
}

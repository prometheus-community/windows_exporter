// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package terminal_services_test

import (
	"os"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/terminal_services"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, terminal_services.Name, terminal_services.NewWithFlags)
}

func TestCollector(t *testing.T) {
	metrics := testutils.TestCollector(t, terminal_services.New, nil)

	// Pending Connections is a current count, published as the pending
	// connections gauge, not as a series of the connection counter.
	for _, metric := range metrics["windows_terminal_services_connection_broker_performance_total"].GetMetric() {
		for _, label := range metric.GetLabel() {
			require.False(t, label.GetName() == "connection" && label.GetValue() == "Pending", "connection_broker_performance_total must not have a Pending series")
		}
	}

	userName := os.Getenv("WINDOWS_EXPORTER_TEST_RDP_USER")
	if userName == "" {
		return
	}

	sessionName := os.Getenv("WINDOWS_EXPORTER_TEST_RDP_SESSION")
	require.NotEmpty(t, sessionName, "RDP fixture did not reach an authenticated active session")
	info := testutils.RequireFixtureMetric(t, metrics, terminal_services.Name, "windows_terminal_services_session_info", prometheus.Labels{
		"user":         os.Getenv("COMPUTERNAME") + `\` + userName,
		"session_name": sessionName,
		"state":        "active",
	})
	require.NotNil(t, info)
	require.InDelta(t, 1, info.GetGauge().GetValue(), 0, "RDP fixture is no longer active")
	handles := testutils.RequireFixtureMetric(t, metrics, terminal_services.Name, "windows_terminal_services_handles", prometheus.Labels{"session_name": sessionName})
	require.NotNil(t, handles)
	require.Positive(t, handles.GetGauge().GetValue(), "RDP fixture has no session processes")

	cpuTime := testutils.RequireFixtureMetric(t, metrics, terminal_services.Name, "windows_terminal_services_cpu_time_seconds_total", prometheus.Labels{
		"mode":         "user",
		"session_name": sessionName,
	})
	require.NotNil(t, cpuTime)
}

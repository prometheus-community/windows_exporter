// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package mscluster_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/mscluster"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/prometheus/client_golang/prometheus"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, mscluster.Name, mscluster.NewWithFlags)
}

func TestCollector(t *testing.T) {
	metrics := testutils.TestCollector(t, mscluster.New, nil)

	// Resource fixtures from the CI cluster setup cover each resource state.
	for name, state := range map[string]float64{
		"CI IP Address":         2, // Online
		"CI Generic Service":    2, // Online
		"CI Offline IP Address": 3, // Offline
		"CI Failed Service":     4, // Failed
	} {
		metric := testutils.RequireFixtureMetric(t, metrics, mscluster.Name, "windows_mscluster_resource_state", prometheus.Labels{"resource": name})
		if metric != nil && metric.GetGauge().GetValue() != state {
			t.Errorf("resource %q state = %v, want %v", name, metric.GetGauge().GetValue(), state)
		}
	}
}

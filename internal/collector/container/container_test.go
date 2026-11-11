// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package container_test

import (
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/collector/container"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, container.Name, container.NewWithFlags)
}

func TestCollector(t *testing.T) {
	metrics := testutils.TestCollector(t, container.New, nil)
	testutils.RequireFixtureMetric(t, metrics, container.Name, "windows_container_available", prometheus.Labels{"container": "hostprocess", "namespace": "default", "hostprocess": "true"})
	testutils.RequireFixtureMetric(t, metrics, container.Name, "windows_container_available", prometheus.Labels{"container": "nanoserver", "namespace": "default", "hostprocess": "false"})
	testutils.RequireFixtureMetric(t, metrics, container.Name, "windows_container_storage_writable_layer_usage_bytes", prometheus.Labels{"container": "nanoserver", "namespace": "default"})

	for _, name := range []string{"hostprocess", "nanoserver"} {
		labels := prometheus.Labels{"container": name, "namespace": "default"}

		if metric := testutils.RequireFixtureMetric(t, metrics, container.Name, "windows_container_processes", labels); metric != nil {
			require.Positive(t, metric.GetGauge().GetValue(), "container %s has no processes", name)
		}

		if metric := testutils.RequireFixtureMetric(t, metrics, container.Name, "windows_container_start_time_seconds", labels); metric != nil {
			require.InDelta(t, float64(time.Now().Unix()), metric.GetGauge().GetValue(), float64(24*time.Hour/time.Second),
				"container %s start time is not within the last day", name)
		}
	}

	testutils.RequireFixtureMetric(t, metrics, container.Name, "windows_container_memory_page_faults_total", prometheus.Labels{"container": "hostprocess", "namespace": "default"})
}

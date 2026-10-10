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

package hyperv_test

import (
	"log/slog"
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/collector/hyperv"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, hyperv.Name, hyperv.NewWithFlags)
}

func TestCollector(t *testing.T) {
	metrics := testutils.TestCollector(t, hyperv.New, nil)
	testutils.RequireFixtureMetric(t, metrics, hyperv.Name, "windows_hyperv_dynamic_memory_vm_physical_bytes", prometheus.Labels{"vm": "GitHubActions"})
	testutils.RequireFixtureMetric(t, metrics, hyperv.Name, "windows_hyperv_vm_processor_count", prometheus.Labels{"vm": "GitHubActions"})

	// The fixture VM is not replicated, but every VM has a primary replication relationship.
	if metric := testutils.RequireFixtureMetric(t, metrics, hyperv.Name, "windows_hyperv_replica_vm_state", prometheus.Labels{"vm": "GitHubActions", "relationship": "primary"}); metric != nil {
		require.InDelta(t, 0, metric.GetGauge().GetValue(), 0, "expected replication to be disabled")
	}

	// The health query reports failures as 0 instead of an error, so check the value.
	if metric := testutils.RequireFixtureMetric(t, metrics, hyperv.Name, "windows_hyperv_wmi_health", nil); metric != nil {
		require.InDelta(t, 1, metric.GetGauge().GetValue(), 0, "Hyper-V WMI health query failed")
	}
}

// TestCollectorVirtualStorageDeviceLatency checks that the deprecated latency gauges
// expose the raw cumulative value in 100ns ticks of the *_io_latency_seconds_total counters.
func TestCollectorVirtualStorageDeviceLatency(t *testing.T) {
	metrics := testutils.TestCollector(t, hyperv.New, &hyperv.Config{CollectorsEnabled: []string{"virtual_storage_device"}})

	for deprecated, replacement := range map[string]string{
		"windows_hyperv_virtual_storage_device_latency_seconds":       "windows_hyperv_virtual_storage_device_io_latency_seconds_total",
		"windows_hyperv_virtual_storage_device_lower_latency_seconds": "windows_hyperv_virtual_storage_device_lower_io_latency_seconds_total",
	} {
		want := testutils.MetricValuesByLabel(metrics, deprecated, "device")
		got := testutils.MetricValuesByLabel(metrics, replacement, "device")

		require.Len(t, got, len(want))

		if len(got) > 0 {
			require.Equal(t, dto.MetricType_COUNTER, metrics[replacement].GetType())
		}

		for device, ticks := range want {
			require.Contains(t, got, device)
			require.InDelta(t, ticks*pdh.TicksToSecondScaleFactor, got[device], 1e-6, "%s{device=%q}", replacement, device)
		}
	}
}

// TestCollectorLegacyNetworkAdapter checks that the legacy network adapter
// metrics are published with the counter type documented in
// docs/collector.hyperv.md; bytes_dropped_total used to be published as a
// gauge. The CI fixture VM carries a legacy network adapter.
func TestCollectorLegacyNetworkAdapter(t *testing.T) {
	metrics := testutils.TestCollector(t, hyperv.New, &hyperv.Config{CollectorsEnabled: []string{"legacy_network_adapter"}})

	for _, name := range []string{
		"windows_hyperv_legacy_network_adapter_bytes_dropped_total",
		"windows_hyperv_legacy_network_adapter_bytes_received_total",
		"windows_hyperv_legacy_network_adapter_bytes_sent_total",
		"windows_hyperv_legacy_network_adapter_frames_dropped_total",
		"windows_hyperv_legacy_network_adapter_frames_received_total",
		"windows_hyperv_legacy_network_adapter_frames_sent_total",
	} {
		testutils.RequireFixtureMetric(t, metrics, hyperv.Name, name, nil)

		if family, ok := metrics[name]; ok {
			require.Equal(t, dto.MetricType_COUNTER, family.GetType(), name)
		}
	}
}

// TestCloseReleasesQuery ensures Close releases the PDH queries of the sub-collectors.
func TestCloseReleasesQuery(t *testing.T) {
	t.Parallel()

	c := hyperv.New(&hyperv.Config{CollectorsEnabled: []string{"dynamic_memory_balancer"}})

	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), nil))
	require.NoError(t, c.Close())

	ch := make(chan prometheus.Metric, 100)

	require.ErrorIs(t, c.Collect(ch, time.Second), pdh.ErrPerformanceCounterNotInitialized)
}

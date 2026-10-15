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

package hyperv

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

// publishCollector adapts a publish function to prometheus.Collector, so the test
// checks the metrics as a registry gathers them.
type publishCollector func(ch chan<- prometheus.Metric) error

func (publishCollector) Describe(chan<- *prometheus.Desc) {}

func (p publishCollector) Collect(ch chan<- prometheus.Metric) {
	if err := p(ch); err != nil {
		panic(err)
	}
}

type gatheredSample struct {
	metricType dto.MetricType
	value      float64
}

func gatherSamples(t *testing.T, publish publishCollector) map[string]gatheredSample {
	t.Helper()

	registry := prometheus.NewPedanticRegistry()
	require.NoError(t, registry.Register(publish))

	families, err := registry.Gather()
	require.NoError(t, err)

	samples := make(map[string]gatheredSample, len(families))

	for _, family := range families {
		metric := family.GetMetric()[0]

		sample := gatheredSample{metricType: family.GetType()}

		switch family.GetType() {
		case dto.MetricType_COUNTER:
			sample.value = metric.GetCounter().GetValue()
		case dto.MetricType_GAUGE:
			sample.value = metric.GetGauge().GetValue()
		default:
			t.Fatalf("unexpected metric type %s for %s", family.GetType(), family.GetName())
		}

		samples[family.GetName()] = sample
	}

	return samples
}

func requireSample(t *testing.T, samples map[string]gatheredSample, name string, metricType dto.MetricType, value float64) {
	t.Helper()

	sample, ok := samples[name]
	require.Truef(t, ok, "metric %s not published", name)
	require.Equalf(t, metricType, sample.metricType, "type of %s", name)
	require.InDeltaf(t, value, sample.value, 1e-9, "value of %s", name)
}

type rootPartitionFixture struct{}

func (rootPartitionFixture) Collect(dst *[]perfDataCounterValuesHypervisorRootPartition) error {
	*dst = []perfDataCounterValuesHypervisorRootPartition{{
		HypervisorRootPartitionGPAPages:                      11,
		HypervisorRootPartitionSkippedTimerTicks:             22,
		HypervisorRootPartitionDeviceInterruptThrottleEvents: 33,
		HypervisorRootPartitionIOTLBFlushCost:                0.5,
		HypervisorRootPartitionIOTLBFlushes:                  44,
		HypervisorRootPartitionGPASpaceModifications:         55,
		HypervisorRootPartitionVirtualTLBFlushEntries:        66,
	}}

	return nil
}

func (rootPartitionFixture) Close() {}

func TestRootPartitionMetricNamesAndTypes(t *testing.T) {
	t.Parallel()

	c := New(nil)
	c.buildHypervisorRootPartitionDescriptors()
	c.perfDataCollectorHypervisorRootPartition = rootPartitionFixture{}

	samples := gatherSamples(t, func(ch chan<- prometheus.Metric) error {
		return c.collectHypervisorRootPartition(ch, nil)
	})

	requireSample(t, samples, "windows_hyperv_root_partition_gpa_pages", dto.MetricType_GAUGE, 11)
	requireSample(t, samples, "windows_hyperv_root_partition_skipped_timer_ticks_total", dto.MetricType_COUNTER, 22)
	requireSample(t, samples, "windows_hyperv_root_partition_device_interrupt_throttle_events_total", dto.MetricType_COUNTER, 33)
	requireSample(t, samples, "windows_hyperv_root_partition_io_tlb_flush_seconds_total", dto.MetricType_COUNTER, 0.5)
	requireSample(t, samples, "windows_hyperv_root_partition_io_tlb_flushes_total", dto.MetricType_COUNTER, 44)
	requireSample(t, samples, "windows_hyperv_root_partition_gpa_space_modifications_total", dto.MetricType_COUNTER, 55)
	requireSample(t, samples, "windows_hyperv_root_partition_virtual_tlb_flush_entries_total", dto.MetricType_COUNTER, 66)

	for _, removed := range []string{
		"windows_hyperv_root_partition_preferred_numa_node_index",
		"windows_hyperv_root_partition_physical_pages_allocated",
		"windows_hyperv_root_partition_io_tlb_flush_cost",
	} {
		require.NotContains(t, samples, removed)
	}
}

func TestVirtualProcessorCPUWaitMetrics(t *testing.T) {
	t.Parallel()

	c := New(nil)
	c.buildHypervisorVirtualProcessorDescriptors()
	c.buildHypervisorRootVirtualProcessorDescriptors()

	// pdh has already scaled the 100ns timers to seconds.
	c.perfDataObjectHypervisorVirtualProcessor = []perfDataCounterValuesHypervisorVirtualProcessor{{
		Name: "vm1:Hv VP 0",
		HypervisorVirtualProcessorCPUWaitTimePerDispatch: 39.13,
		HypervisorVirtualProcessorDispatches:             1086776,
	}}
	c.perfDataObjectHypervisorRootVirtualProcessor = []perfDataCounterValuesHypervisorRootVirtualProcessor{{
		Name: "Root VP 0",
		HypervisorRootVirtualProcessorCPUWaitTimePerDispatch: 9.98,
		HypervisorRootVirtualProcessorDispatches:             1082841,
	}}

	samples := gatherSamples(t, func(ch chan<- prometheus.Metric) error {
		if err := c.publishHypervisorVirtualProcessor(ch, nil); err != nil {
			return err
		}

		return c.publishHypervisorRootVirtualProcessor(ch, nil)
	})

	requireSample(t, samples, "windows_hyperv_hypervisor_virtual_processor_cpu_wait_seconds_total", dto.MetricType_COUNTER, 39.13)
	requireSample(t, samples, "windows_hyperv_hypervisor_virtual_processor_dispatches_total", dto.MetricType_COUNTER, 1086776)
	requireSample(t, samples, "windows_hyperv_hypervisor_root_virtual_processor_cpu_wait_seconds_total", dto.MetricType_COUNTER, 9.98)
	requireSample(t, samples, "windows_hyperv_hypervisor_root_virtual_processor_dispatches_total", dto.MetricType_COUNTER, 1082841)
	require.NotContains(t, samples, "windows_hyperv_hypervisor_virtual_processor_cpu_wait_time_per_dispatch_total")
	require.NotContains(t, samples, "windows_hyperv_hypervisor_root_virtual_cpu_wait_time_per_dispatch_total")
}

func TestVirtualStorageDeviceCounterTypes(t *testing.T) {
	t.Parallel()

	c := New(nil)
	c.buildVirtualStorageDeviceDescriptors()
	c.perfDataObjectVirtualStorageDevice = []perfDataCounterValuesVirtualStorageDevice{{
		Name:                                     "ubuntu.vhdx",
		VirtualStorageDeviceThroughput:           22336,
		VirtualStorageDeviceNormalizedThroughput: 36413,
		VirtualStorageDeviceIOQuotaReplenishmentRate: 7,
	}}

	samples := gatherSamples(t, func(ch chan<- prometheus.Metric) error {
		c.publishVirtualStorageDevice(ch, nil)

		return nil
	})

	requireSample(t, samples, "windows_hyperv_virtual_storage_device_throughput_total", dto.MetricType_COUNTER, 22336)
	requireSample(t, samples, "windows_hyperv_virtual_storage_device_normalized_throughput_total", dto.MetricType_COUNTER, 36413)
	requireSample(t, samples, "windows_hyperv_virtual_storage_device_io_quota_replenishment_total", dto.MetricType_COUNTER, 7)
	require.NotContains(t, samples, "windows_hyperv_virtual_storage_device_normalized_throughput")
	require.NotContains(t, samples, "windows_hyperv_io_quota_replenishment_rate")
}

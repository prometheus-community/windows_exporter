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
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

// Hyper-V Virtual Storage Device metrics
type collectorVirtualStorageDevice struct {
	perfDataCollectorVirtualStorageDevice *pdh.Collector[perfDataCounterValuesVirtualStorageDevice]
	perfDataObjectVirtualStorageDevice    []perfDataCounterValuesVirtualStorageDevice

	virtualStorageDeviceErrorCount               *prometheus.Desc // \Hyper-V Virtual Storage Device(*)\Error Count
	virtualStorageDeviceQueueLength              *prometheus.Desc // \Hyper-V Virtual Storage Device(*)\Queue Length
	virtualStorageDeviceReadBytes                *prometheus.Desc // \Hyper-V Virtual Storage Device(*)\Read Bytes/sec
	virtualStorageDeviceReadOperations           *prometheus.Desc // \Hyper-V Virtual Storage Device(*)\Read Operations/Sec
	virtualStorageDeviceWriteBytes               *prometheus.Desc // \Hyper-V Virtual Storage Device(*)\Write Bytes/sec
	virtualStorageDeviceWriteOperations          *prometheus.Desc // \Hyper-V Virtual Storage Device(*)\Write Operations/Sec
	virtualStorageDeviceLatency                  *prometheus.Desc // \Hyper-V Virtual Storage Device(*)\Latency
	virtualStorageDeviceThroughput               *prometheus.Desc // \Hyper-V Virtual Storage Device(*)\Throughput
	virtualStorageDeviceNormalizedThroughput     *prometheus.Desc // \Hyper-V Virtual Storage Device(*)\Normalized Throughput
	virtualStorageDeviceLowerQueueLength         *prometheus.Desc // \Hyper-V Virtual Storage Device(*)\Lower Queue Length
	virtualStorageDeviceLowerLatency             *prometheus.Desc // \Hyper-V Virtual Storage Device(*)\Lower Latency
	virtualStorageDeviceIOQuotaReplenishmentRate *prometheus.Desc // \Hyper-V Virtual Storage Device(*)\IO Quota Replenishment Rate
	virtualStorageDeviceIOLatency                *prometheus.Desc // \Hyper-V Virtual Storage Device(*)\Latency
	virtualStorageDeviceLowerIOLatency           *prometheus.Desc // \Hyper-V Virtual Storage Device(*)\Lower Latency
}

type perfDataCounterValuesVirtualStorageDevice struct {
	Name string

	VirtualStorageDeviceErrorCount               float64 `perfdata:"Error Count"`
	VirtualStorageDeviceQueueLength              float64 `perfdata:"Queue Length"`
	VirtualStorageDeviceReadBytes                float64 `perfdata:"Read Bytes/sec"`
	VirtualStorageDeviceReadOperations           float64 `perfdata:"Read Count"`
	VirtualStorageDeviceWriteBytes               float64 `perfdata:"Write Bytes/sec"`
	VirtualStorageDeviceWriteOperations          float64 `perfdata:"Write Count"`
	VirtualStorageDeviceLatency                  float64 `perfdata:"Latency"`
	VirtualStorageDeviceThroughput               float64 `perfdata:"Throughput"`
	VirtualStorageDeviceNormalizedThroughput     float64 `perfdata:"Normalized Throughput"`
	VirtualStorageDeviceLowerQueueLength         float64 `perfdata:"Lower Queue Length"`
	VirtualStorageDeviceLowerLatency             float64 `perfdata:"Lower Latency"`
	VirtualStorageDeviceIOQuotaReplenishmentRate float64 `perfdata:"IO Quota Replenishment Rate"`
}

func (c *Collector) buildVirtualStorageDevice() error {
	var err error

	c.perfDataCollectorVirtualStorageDevice, err = pdh.NewCollector[perfDataCounterValuesVirtualStorageDevice](c.logger, pdh.CounterTypeRaw, "Hyper-V Virtual Storage Device", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create Hyper-V Virtual Storage Device collector: %w", err)
	}

	c.virtualStorageDeviceErrorCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_storage_device_error_count_total"),
		"Represents the total number of errors that have occurred on this virtual device.",
		[]string{"device"},
		nil,
	)
	c.virtualStorageDeviceQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_storage_device_queue_length"),
		"DEPRECATED: use rate(windows_hyperv_virtual_storage_device_io_latency_seconds_total). "+
			"Cumulative queue length in 100ns ticks, identical to the raw latency.",
		[]string{"device"},
		nil,
	)
	c.virtualStorageDeviceReadBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_storage_device_bytes_read"),
		"Represents the total number of bytes that have been read on this virtual device.",
		[]string{"device"},
		nil,
	)
	c.virtualStorageDeviceReadOperations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_storage_device_operations_read_total"),
		"Represents the total number of read operations that have occurred on this virtual device.",
		[]string{"device"},
		nil,
	)
	c.virtualStorageDeviceWriteBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_storage_device_bytes_written"),
		"Represents the total number of bytes that have been written on this virtual device.",
		[]string{"device"},
		nil,
	)
	c.virtualStorageDeviceWriteOperations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_storage_device_operations_written_total"),
		"Represents the total number of write operations that have occurred on this virtual device.",
		[]string{"device"},
		nil,
	)
	c.virtualStorageDeviceLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_storage_device_latency_seconds"),
		"DEPRECATED: use windows_hyperv_virtual_storage_device_io_latency_seconds_total. "+
			"Cumulative IO transfer latency in 100ns ticks for this virtual device.",
		[]string{"device"},
		nil,
	)
	c.virtualStorageDeviceThroughput = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_storage_device_throughput_total"),
		"Represents the total number of 8KB IO transfers completed by this virtual device.",
		[]string{"device"},
		nil,
	)
	c.virtualStorageDeviceNormalizedThroughput = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_storage_device_normalized_throughput"),
		"Represents the average number of IO transfers completed by this virtual device.",
		[]string{"device"},
		nil,
	)
	c.virtualStorageDeviceLowerQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_storage_device_lower_queue_length"),
		"DEPRECATED: use rate(windows_hyperv_virtual_storage_device_lower_io_latency_seconds_total). "+
			"Cumulative queue length in 100ns ticks on the underlying storage subsystem, identical to the raw lower latency.",
		[]string{"device"},
		nil,
	)
	c.virtualStorageDeviceLowerLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_storage_device_lower_latency_seconds"),
		"DEPRECATED: use windows_hyperv_virtual_storage_device_lower_io_latency_seconds_total. "+
			"Cumulative IO transfer latency in 100ns ticks on the underlying storage subsystem for this virtual device.",
		[]string{"device"},
		nil,
	)
	c.virtualStorageDeviceIOLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_storage_device_io_latency_seconds_total"),
		"Represents the total IO transfer latency for this virtual device. "+
			"Divide its rate by the rate of windows_hyperv_virtual_storage_device_throughput_total for the average latency.",
		[]string{"device"},
		nil,
	)
	c.virtualStorageDeviceLowerIOLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_storage_device_lower_io_latency_seconds_total"),
		"Represents the total IO transfer latency on the underlying storage subsystem for this virtual device. "+
			"Divide its rate by the rate of windows_hyperv_virtual_storage_device_throughput_total for the average latency.",
		[]string{"device"},
		nil,
	)
	c.virtualStorageDeviceIOQuotaReplenishmentRate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "io_quota_replenishment_rate"),
		"Represents the IO quota replenishment rate for this virtual device.",
		[]string{"device"},
		nil,
	)

	return nil
}

func (c *Collector) collectVirtualStorageDevice(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollectorVirtualStorageDevice.Collect(&c.perfDataObjectVirtualStorageDevice)
	if err != nil {
		return fmt.Errorf("failed to collect Hyper-V Virtual Storage Device metrics: %w", err)
	}

	for _, data := range c.perfDataObjectVirtualStorageDevice {
		ch <- prometheus.MustNewConstMetric(
			c.virtualStorageDeviceErrorCount,
			prometheus.CounterValue,
			data.VirtualStorageDeviceErrorCount,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualStorageDeviceQueueLength,
			prometheus.GaugeValue,
			data.VirtualStorageDeviceQueueLength,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualStorageDeviceReadBytes,
			prometheus.CounterValue,
			data.VirtualStorageDeviceReadBytes,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualStorageDeviceReadOperations,
			prometheus.CounterValue,
			data.VirtualStorageDeviceReadOperations,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualStorageDeviceWriteBytes,
			prometheus.CounterValue,
			data.VirtualStorageDeviceWriteBytes,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualStorageDeviceWriteOperations,
			prometheus.CounterValue,
			data.VirtualStorageDeviceWriteOperations,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualStorageDeviceLatency,
			prometheus.GaugeValue,
			data.VirtualStorageDeviceLatency,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualStorageDeviceThroughput,
			prometheus.CounterValue,
			data.VirtualStorageDeviceThroughput,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualStorageDeviceNormalizedThroughput,
			prometheus.GaugeValue,
			data.VirtualStorageDeviceNormalizedThroughput,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualStorageDeviceLowerQueueLength,
			prometheus.GaugeValue,
			data.VirtualStorageDeviceLowerQueueLength,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualStorageDeviceLowerLatency,
			prometheus.GaugeValue,
			data.VirtualStorageDeviceLowerLatency,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualStorageDeviceIOQuotaReplenishmentRate,
			prometheus.GaugeValue,
			data.VirtualStorageDeviceIOQuotaReplenishmentRate,
			data.Name,
		)

		// The raw value of a PERF_AVERAGE_TIMER counter is the sum of all IO latencies in
		// 100ns ticks; the operation count it is averaged over is Throughput.
		ch <- prometheus.MustNewConstMetric(
			c.virtualStorageDeviceIOLatency,
			prometheus.CounterValue,
			data.VirtualStorageDeviceLatency*pdh.TicksToSecondScaleFactor,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualStorageDeviceLowerIOLatency,
			prometheus.CounterValue,
			data.VirtualStorageDeviceLowerLatency*pdh.TicksToSecondScaleFactor,
			data.Name,
		)
	}

	return nil
}

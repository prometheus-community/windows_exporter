package hyperv

import (
	"errors"
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

// Hyper-V Virtual Storage Device metrics
type collectorVirtualStorageDevice struct {
	perfDataCollectorVirtualStorageDevice *perfdata.Collector

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
}

const (
	virtualStorageDeviceErrorCount               = "Error Count"
	virtualStorageDeviceQueueLength              = "Queue Length"
	virtualStorageDeviceReadBytes                = "Read Bytes/sec"
	virtualStorageDeviceReadOperations           = "Read Count"
	virtualStorageDeviceWriteBytes               = "Write Bytes/sec"
	virtualStorageDeviceWriteOperations          = "Write Count"
	virtualStorageDeviceLatency                  = "Latency"
	virtualStorageDeviceThroughput               = "Throughput"
	virtualStorageDeviceNormalizedThroughput     = "Normalized Throughput"
	virtualStorageDeviceLowerQueueLength         = "Lower Queue Length"
	virtualStorageDeviceLowerLatency             = "Lower Latency"
	virtualStorageDeviceIOQuotaReplenishmentRate = "IO Quota Replenishment Rate"
)

func (c *Collector) buildVirtualStorageDevice() error {
	var err error

	c.perfDataCollectorVirtualStorageDevice, err = perfdata.NewCollector("Hyper-V Virtual Storage Device", perfdata.InstanceAll, []string{
		virtualStorageDeviceErrorCount,
		virtualStorageDeviceQueueLength,
		virtualStorageDeviceReadBytes,
		virtualStorageDeviceReadOperations,
		virtualStorageDeviceWriteBytes,
		virtualStorageDeviceWriteOperations,
		virtualStorageDeviceLatency,
		virtualStorageDeviceThroughput,
		virtualStorageDeviceNormalizedThroughput,
		virtualStorageDeviceLowerQueueLength,
		virtualStorageDeviceLowerLatency,
		virtualStorageDeviceIOQuotaReplenishmentRate,
	})
	if err != nil && !errors.Is(err, perfdata.ErrNoData) {
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
		"Represents the average queue length on this virtual device.",
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
		"Represents the average IO transfer latency for this virtual device.",
		[]string{"device"},
		nil,
	)
	c.virtualStorageDeviceThroughput = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_storage_device_throughput"),
		"Represents the average number of 8KB IO transfers completed by this virtual device.",
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
		"Represents the average queue length on the underlying storage subsystem for this device.",
		[]string{"device"},
		nil,
	)
	c.virtualStorageDeviceLowerLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_storage_device_lower_latency_seconds"),
		"Represents the average IO transfer latency on the underlying storage subsystem for this virtual device.",
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
	data, err := c.perfDataCollectorVirtualStorageDevice.Collect()
	if err != nil && !errors.Is(err, perfdata.ErrNoData) {
		return fmt.Errorf("failed to collect Hyper-V Virtual Storage Device metrics: %w", err)
	}

	for name, device := range data {
		ch <- prometheus.MustNewConstMetric(
			c.virtualStorageDeviceErrorCount,
			prometheus.CounterValue,
			device[virtualStorageDeviceErrorCount].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualStorageDeviceQueueLength,
			prometheus.GaugeValue,
			device[virtualStorageDeviceQueueLength].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualStorageDeviceReadBytes,
			prometheus.CounterValue,
			device[virtualStorageDeviceReadBytes].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualStorageDeviceReadOperations,
			prometheus.CounterValue,
			device[virtualStorageDeviceReadOperations].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualStorageDeviceWriteBytes,
			prometheus.CounterValue,
			device[virtualStorageDeviceWriteBytes].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualStorageDeviceWriteOperations,
			prometheus.CounterValue,
			device[virtualStorageDeviceWriteOperations].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualStorageDeviceLatency,
			prometheus.GaugeValue,
			device[virtualStorageDeviceLatency].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualStorageDeviceThroughput,
			prometheus.GaugeValue,
			device[virtualStorageDeviceThroughput].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualStorageDeviceNormalizedThroughput,
			prometheus.GaugeValue,
			device[virtualStorageDeviceNormalizedThroughput].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualStorageDeviceLowerQueueLength,
			prometheus.GaugeValue,
			device[virtualStorageDeviceLowerQueueLength].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualStorageDeviceLowerLatency,
			prometheus.GaugeValue,
			device[virtualStorageDeviceLowerLatency].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualStorageDeviceIOQuotaReplenishmentRate,
			prometheus.GaugeValue,
			device[virtualStorageDeviceIOQuotaReplenishmentRate].FirstValue,
			name,
		)
	}

	return nil
}

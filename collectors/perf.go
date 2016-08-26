// https://msdn.microsoft.com/en-us/windows/hardware/aa394307(v=vs.71) - Win32_PerfRawData_PerfDisk_LogicalDisk class

// TODO export all disks ... currently only first disk is exported

package collectors

import (
	"log"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

// A PerfCollector is a Prometheus collector for WMI Win32_PerfRawData_PerfDisk_LogicalDisk metrics
type PerfCollector struct {
	AvgDiskBytesPerRead          *prometheus.Desc
	AvgDiskBytesPerRead_Base     *prometheus.Desc
	AvgDiskBytesPerTransfer      *prometheus.Desc
	AvgDiskBytesPerTransfer_Base *prometheus.Desc
	AvgDiskBytesPerWrite         *prometheus.Desc
	AvgDiskBytesPerWrite_Base    *prometheus.Desc
	AvgDiskQueueLength           *prometheus.Desc
	AvgDiskReadQueueLength       *prometheus.Desc
	AvgDiskSecPerRead            *prometheus.Desc
	AvgDiskSecPerRead_Base       *prometheus.Desc
	AvgDiskSecPerTransfer        *prometheus.Desc
	AvgDiskSecPerTransfer_Base   *prometheus.Desc
	AvgDiskSecPerWrite           *prometheus.Desc
	AvgDiskSecPerWrite_Base      *prometheus.Desc
	AvgDiskWriteQueueLength      *prometheus.Desc
	CurrentDiskQueueLength       *prometheus.Desc
	DiskBytesPerSec              *prometheus.Desc
	DiskReadBytesPerSec          *prometheus.Desc
	DiskReadsPerSec              *prometheus.Desc
	DiskTransfersPerSec          *prometheus.Desc
	DiskWriteBytesPerSec         *prometheus.Desc
	DiskWritesPerSec             *prometheus.Desc
	FreeMegabytes                *prometheus.Desc
	PercentDiskReadTime          *prometheus.Desc
	PercentDiskReadTime_Base     *prometheus.Desc
	PercentDiskTime              *prometheus.Desc
	PercentDiskTime_Base         *prometheus.Desc
	PercentDiskWriteTime         *prometheus.Desc
	PercentDiskWriteTime_Base    *prometheus.Desc
	PercentFreeSpace             *prometheus.Desc
	PercentFreeSpace_Base        *prometheus.Desc
	PercentIdleTime              *prometheus.Desc
	PercentIdleTime_Base         *prometheus.Desc
	SplitIOPerSec                *prometheus.Desc
}

// NewPerfCollector ...
func NewPerfCollector() *PerfCollector {

	return &PerfCollector{
		AvgDiskBytesPerRead: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "avg_disk_bytes_per_read"),
			"Average number of bytes transferred from the disk during read operations.",
			nil,
			nil,
		),

		AvgDiskBytesPerRead_Base: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "avg_disk_bytes_per_read_base"),
			"Base value for AvgDiskBytesPerRead. This value represents the accumulated number of operations that have taken place.",
			nil,
			nil,
		),

		AvgDiskBytesPerTransfer: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "avg_disk_bytes_per_transfer"),
			"Average number of bytes transferred to or from the disk during write or read operations.",
			nil,
			nil,
		),

		AvgDiskBytesPerTransfer_Base: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "avg_disk_bytes_per_transfer_base"),
			"Base value for AvgDiskBytesPerTransfer. This value represents the accumulated number of operations that have taken place.",
			nil,
			nil,
		),

		AvgDiskBytesPerWrite: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "avg_disk_bytes_per_write"),
			"Average number of bytes transferred to the disk during write operations.",
			nil,
			nil,
		),

		AvgDiskBytesPerWrite_Base: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "avg_disk_bytes_per_write_base"),
			"Base value for AvgDiskBytesPerWrite. This value represents the accumulated number of operations that have taken place.",
			nil,
			nil,
		),

		AvgDiskQueueLength: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "avg_disk_queue_length"),
			"Average number of both read and write requests that were queued for the selected disk during the sample interval.",
			nil,
			nil,
		),

		AvgDiskReadQueueLength: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "avg_disk_read_queue_length"),
			"Average number of read requests that were queued for the selected disk during the sample interval.",
			nil,
			nil,
		),

		AvgDiskSecPerRead: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "avg_disk_sec_per_read"),
			"Average time, in seconds, of a read operation of data from the disk.",
			nil,
			nil,
		),

		AvgDiskSecPerRead_Base: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "avg_disk_sec_per_read_base"),
			"Base value for AvgDiskSecPerRead. This value represents the accumulated number of operations that have taken place.",
			nil,
			nil,
		),

		AvgDiskSecPerTransfer: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "avg_disk_sec_per_transfer"),
			"Time, in seconds, of the average disk transfer.",
			nil,
			nil,
		),

		AvgDiskSecPerTransfer_Base: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "avg_disk_sec_per_transfer_base"),
			"Base value for AvgDiskSecPerTransfer. This value represents the accumulated number of operations that have taken place.",
			nil,
			nil,
		),

		AvgDiskSecPerWrite: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "avg_disk_sec_per_write"),
			"Average time, in seconds, of a write operation of data to the disk.",
			nil,
			nil,
		),

		AvgDiskSecPerWrite_Base: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "avg_disk_sec_per_write_base"),
			"Base value for AvgDiskSecPerWrite. This value represents the accumulated number of operations that have taken place.",
			nil,
			nil,
		),

		AvgDiskWriteQueueLength: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "avg_disk_write_queue_length"),
			"Average number of write requests that were queued for the selected disk during the sample interval. The time base is 100 nanoseconds.",
			nil,
			nil,
		),

		CurrentDiskQueueLength: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "current_disk_queue_length"),
			"Number of requests outstanding on the disk at the time the performance data is collected.",
			nil,
			nil,
		),

		DiskBytesPerSec: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "disk_bytes_per_sec"),
			"Rate at which bytes are transferred to or from the disk during write or read operations.",
			nil,
			nil,
		),

		DiskReadBytesPerSec: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "disk_read_bytes_per_sec"),
			"Rate at which bytes are transferred from the disk during read operations.",
			nil,
			nil,
		),

		DiskReadsPerSec: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "disk_reads_per_sec"),
			"Rate of read operations on the disk.",
			nil,
			nil,
		),

		DiskTransfersPerSec: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "disk_transfers_per_sec"),
			"Rate of read and write operations on the disk.",
			nil,
			nil,
		),

		DiskWriteBytesPerSec: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "disk_write_bytes_per_sec"),
			"Rate at which bytes are transferred to the disk during write operations.",
			nil,
			nil,
		),

		DiskWritesPerSec: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "disk_writes_per_sec"),
			"Rate of write operations on the disk.",
			nil,
			nil,
		),

		FreeMegabytes: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "free_megabytes"),
			"Unallocated space on the disk drive in megabytes.",
			nil,
			nil,
		),

		PercentDiskReadTime: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "percent_disk_read_time"),
			"Percentage of elapsed time that the selected disk drive is busy servicing read requests.",
			nil,
			nil,
		),

		PercentDiskReadTime_Base: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "percent_disk_read_time_base"),
			"Base value for PercentDiskReadTime.",
			nil,
			nil,
		),

		PercentDiskTime: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "percent_disk_time"),
			"Percentage of elapsed time that the selected disk drive is busy servicing read or write requests.",
			nil,
			nil,
		),

		PercentDiskTime_Base: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "percent_disk_time_base"),
			"Base value for PercentDiskTime.",
			nil,
			nil,
		),

		PercentDiskWriteTime: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "percent_disk_write_time"),
			"Percentage of elapsed time that the selected disk drive is busy servicing write requests.",
			nil,
			nil,
		),

		PercentDiskWriteTime_Base: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "percent_disk_write_time_base"),
			"Base value for PercentDiskWriteTime.",
			nil,
			nil,
		),

		PercentFreeSpace: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "percent_free_space"),
			"Ratio of the free space available on the logical disk unit to the total usable space provided by the selected logical disk drive.",
			nil,
			nil,
		),

		PercentFreeSpace_Base: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "percent_free_space_base"),
			"Base value for PercentFreeSpace.",
			nil,
			nil,
		),

		PercentIdleTime: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "percent_idle_time"),
			"Percentage of time during the sample interval that the disk was idle.",
			nil,
			nil,
		),

		PercentIdleTime_Base: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "percent_idle_time_base"),
			"Base value for PercentIdleTime.",
			nil,
			nil,
		),

		SplitIOPerSec: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "perf", "split_io_per_sec"),
			"Rate at which I/Os to the disk were split into multiple I/Os.",
			nil,
			nil,
		),
	}
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *PerfCollector) Collect(ch chan<- prometheus.Metric) {
	if desc, err := c.collect(ch); err != nil {
		log.Println("[ERROR] failed collecting perf metrics:", desc, err)
		return
	}
}

// Describe sends the descriptors of each metric over to the provided channel.
// The corresponding metric values are sent separately.
func (c *PerfCollector) Describe(ch chan<- *prometheus.Desc) {

	ch <- c.AvgDiskBytesPerRead
	ch <- c.AvgDiskBytesPerRead_Base
	ch <- c.AvgDiskBytesPerTransfer
	ch <- c.AvgDiskBytesPerTransfer_Base
	ch <- c.AvgDiskBytesPerWrite
	ch <- c.AvgDiskBytesPerWrite_Base
	ch <- c.AvgDiskQueueLength
	ch <- c.AvgDiskReadQueueLength
	ch <- c.AvgDiskSecPerRead
	ch <- c.AvgDiskSecPerRead_Base
	ch <- c.AvgDiskSecPerTransfer
	ch <- c.AvgDiskSecPerTransfer_Base
	ch <- c.AvgDiskSecPerWrite
	ch <- c.AvgDiskSecPerWrite_Base
	ch <- c.AvgDiskWriteQueueLength
	ch <- c.CurrentDiskQueueLength
	ch <- c.DiskBytesPerSec
	ch <- c.DiskReadBytesPerSec
	ch <- c.DiskReadsPerSec
	ch <- c.DiskTransfersPerSec
	ch <- c.DiskWriteBytesPerSec
	ch <- c.DiskWritesPerSec
	ch <- c.FreeMegabytes
	ch <- c.PercentDiskReadTime
	ch <- c.PercentDiskReadTime_Base
	ch <- c.PercentDiskTime
	ch <- c.PercentDiskTime_Base
	ch <- c.PercentDiskWriteTime
	ch <- c.PercentDiskWriteTime_Base
	ch <- c.PercentFreeSpace
	ch <- c.PercentFreeSpace_Base
	ch <- c.PercentIdleTime
	ch <- c.PercentIdleTime_Base
	ch <- c.SplitIOPerSec
}

type Win32_PerfRawData_PerfDisk_LogicalDisk struct {
	AvgDiskBytesPerRead          uint64
	AvgDiskBytesPerRead_Base     uint32
	AvgDiskBytesPerTransfer      uint64
	AvgDiskBytesPerTransfer_Base uint32
	AvgDiskBytesPerWrite         uint64
	AvgDiskBytesPerWrite_Base    uint32
	AvgDiskQueueLength           uint64
	AvgDiskReadQueueLength       uint64
	AvgDiskSecPerRead            uint32
	AvgDiskSecPerRead_Base       uint32
	AvgDiskSecPerTransfer        uint32
	AvgDiskSecPerTransfer_Base   uint32
	AvgDiskSecPerWrite           uint32
	AvgDiskSecPerWrite_Base      uint32
	AvgDiskWriteQueueLength      uint64
	CurrentDiskQueueLength       uint32
	DiskBytesPerSec              uint64
	DiskReadBytesPerSec          uint64
	DiskReadsPerSec              uint32
	DiskTransfersPerSec          uint32
	DiskWriteBytesPerSec         uint64
	DiskWritesPerSec             uint32
	FreeMegabytes                uint32
	PercentDiskReadTime          uint64
	PercentDiskReadTime_Base     uint64
	PercentDiskTime              uint64
	PercentDiskTime_Base         uint64
	PercentDiskWriteTime         uint64
	PercentDiskWriteTime_Base    uint64
	PercentFreeSpace             uint32
	PercentFreeSpace_Base        uint32
	PercentIdleTime              uint64
	PercentIdleTime_Base         uint64
	SplitIOPerSec                uint32
}

func (c *PerfCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_PerfDisk_LogicalDisk
	q := wmi.CreateQuery(&dst, "")
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	ch <- prometheus.MustNewConstMetric(
		c.AvgDiskBytesPerRead,
		prometheus.GaugeValue,
		float64(dst[0].AvgDiskBytesPerRead),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AvgDiskBytesPerRead_Base,
		prometheus.GaugeValue,
		float64(dst[0].AvgDiskBytesPerRead_Base),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AvgDiskBytesPerTransfer,
		prometheus.GaugeValue,
		float64(dst[0].AvgDiskBytesPerTransfer),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AvgDiskBytesPerTransfer_Base,
		prometheus.GaugeValue,
		float64(dst[0].AvgDiskBytesPerTransfer_Base),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AvgDiskBytesPerWrite,
		prometheus.GaugeValue,
		float64(dst[0].AvgDiskBytesPerWrite),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AvgDiskBytesPerWrite_Base,
		prometheus.GaugeValue,
		float64(dst[0].AvgDiskBytesPerWrite_Base),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AvgDiskQueueLength,
		prometheus.GaugeValue,
		float64(dst[0].AvgDiskQueueLength),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AvgDiskReadQueueLength,
		prometheus.GaugeValue,
		float64(dst[0].AvgDiskReadQueueLength),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AvgDiskSecPerRead,
		prometheus.GaugeValue,
		float64(dst[0].AvgDiskSecPerRead),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AvgDiskSecPerRead_Base,
		prometheus.GaugeValue,
		float64(dst[0].AvgDiskSecPerRead_Base),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AvgDiskSecPerTransfer,
		prometheus.GaugeValue,
		float64(dst[0].AvgDiskSecPerTransfer),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AvgDiskSecPerTransfer_Base,
		prometheus.GaugeValue,
		float64(dst[0].AvgDiskSecPerTransfer_Base),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AvgDiskSecPerWrite,
		prometheus.GaugeValue,
		float64(dst[0].AvgDiskSecPerWrite),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AvgDiskSecPerWrite_Base,
		prometheus.GaugeValue,
		float64(dst[0].AvgDiskSecPerWrite_Base),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AvgDiskWriteQueueLength,
		prometheus.GaugeValue,
		float64(dst[0].AvgDiskWriteQueueLength),
	)

	ch <- prometheus.MustNewConstMetric(
		c.CurrentDiskQueueLength,
		prometheus.GaugeValue,
		float64(dst[0].CurrentDiskQueueLength),
	)

	ch <- prometheus.MustNewConstMetric(
		c.DiskBytesPerSec,
		prometheus.GaugeValue,
		float64(dst[0].DiskBytesPerSec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.DiskReadBytesPerSec,
		prometheus.GaugeValue,
		float64(dst[0].DiskReadBytesPerSec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.DiskReadsPerSec,
		prometheus.GaugeValue,
		float64(dst[0].DiskReadsPerSec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.DiskTransfersPerSec,
		prometheus.GaugeValue,
		float64(dst[0].DiskTransfersPerSec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.DiskWriteBytesPerSec,
		prometheus.GaugeValue,
		float64(dst[0].DiskWriteBytesPerSec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.DiskWritesPerSec,
		prometheus.GaugeValue,
		float64(dst[0].DiskWritesPerSec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.FreeMegabytes,
		prometheus.GaugeValue,
		float64(dst[0].FreeMegabytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PercentDiskReadTime,
		prometheus.GaugeValue,
		float64(dst[0].PercentDiskReadTime),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PercentDiskReadTime_Base,
		prometheus.GaugeValue,
		float64(dst[0].PercentDiskReadTime_Base),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PercentDiskTime,
		prometheus.GaugeValue,
		float64(dst[0].PercentDiskTime),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PercentDiskTime_Base,
		prometheus.GaugeValue,
		float64(dst[0].PercentDiskTime_Base),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PercentDiskWriteTime,
		prometheus.GaugeValue,
		float64(dst[0].PercentDiskWriteTime),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PercentDiskWriteTime_Base,
		prometheus.GaugeValue,
		float64(dst[0].PercentDiskWriteTime_Base),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PercentFreeSpace,
		prometheus.GaugeValue,
		float64(dst[0].PercentFreeSpace),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PercentFreeSpace_Base,
		prometheus.GaugeValue,
		float64(dst[0].PercentFreeSpace_Base),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PercentIdleTime,
		prometheus.GaugeValue,
		float64(dst[0].PercentIdleTime),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PercentIdleTime_Base,
		prometheus.GaugeValue,
		float64(dst[0].PercentIdleTime_Base),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SplitIOPerSec,
		prometheus.GaugeValue,
		float64(dst[0].SplitIOPerSec),
	)

	return nil, nil
}

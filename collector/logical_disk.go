//go:build windows
// +build windows

package collector

import (
	"fmt"
	"regexp"

	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/alecthomas/kingpin.v2"
)

func init() {
	registerCollector("logical_disk", NewLogicalDiskCollector, "LogicalDisk")
}

var (
	volumeWhitelist = kingpin.Flag(
		"collector.logical_disk.volume-whitelist",
		"Regexp of volumes to whitelist. Volume name must both match whitelist and not match blacklist to be included.",
	).Default(".+").String()
	volumeBlacklist = kingpin.Flag(
		"collector.logical_disk.volume-blacklist",
		"Regexp of volumes to blacklist. Volume name must both match whitelist and not match blacklist to be included.",
	).Default("").String()
)

// A LogicalDiskCollector is a Prometheus collector for perflib logicalDisk metrics
type LogicalDiskCollector struct {
	RequestsQueued   *prometheus.Desc
	ReadBytesTotal   *prometheus.Desc
	ReadsTotal       *prometheus.Desc
	WriteBytesTotal  *prometheus.Desc
	WritesTotal      *prometheus.Desc
	ReadTime         *prometheus.Desc
	WriteTime        *prometheus.Desc
	TotalSpace       *prometheus.Desc
	FreeSpace        *prometheus.Desc
	IdleTime         *prometheus.Desc
	SplitIOs         *prometheus.Desc
	ReadLatency      *prometheus.Desc
	WriteLatency     *prometheus.Desc
	ReadWriteLatency *prometheus.Desc

	volumeWhitelistPattern *regexp.Regexp
	volumeBlacklistPattern *regexp.Regexp
}

// NewLogicalDiskCollector ...
func NewLogicalDiskCollector() (Collector, error) {
	const subsystem = "logical_disk"

	return &LogicalDiskCollector{
		RequestsQueued: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "requests_queued"),
			"The number of requests queued to the disk (LogicalDisk.CurrentDiskQueueLength)",
			[]string{"volume"},
			nil,
		),

		ReadBytesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "read_bytes_total"),
			"The number of bytes transferred from the disk during read operations (LogicalDisk.DiskReadBytesPerSec)",
			[]string{"volume"},
			nil,
		),

		ReadsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "reads_total"),
			"The number of read operations on the disk (LogicalDisk.DiskReadsPerSec)",
			[]string{"volume"},
			nil,
		),

		WriteBytesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "write_bytes_total"),
			"The number of bytes transferred to the disk during write operations (LogicalDisk.DiskWriteBytesPerSec)",
			[]string{"volume"},
			nil,
		),

		WritesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "writes_total"),
			"The number of write operations on the disk (LogicalDisk.DiskWritesPerSec)",
			[]string{"volume"},
			nil,
		),

		ReadTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "read_seconds_total"),
			"Seconds that the disk was busy servicing read requests (LogicalDisk.PercentDiskReadTime)",
			[]string{"volume"},
			nil,
		),

		WriteTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "write_seconds_total"),
			"Seconds that the disk was busy servicing write requests (LogicalDisk.PercentDiskWriteTime)",
			[]string{"volume"},
			nil,
		),

		FreeSpace: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "free_bytes"),
			"Free space in bytes, updates every 10-15 min (LogicalDisk.PercentFreeSpace)",
			[]string{"volume"},
			nil,
		),

		TotalSpace: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "size_bytes"),
			"Total space in bytes, updates every 10-15 min (LogicalDisk.PercentFreeSpace_Base)",
			[]string{"volume"},
			nil,
		),

		IdleTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "idle_seconds_total"),
			"Seconds that the disk was idle (LogicalDisk.PercentIdleTime)",
			[]string{"volume"},
			nil,
		),

		SplitIOs: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "split_ios_total"),
			"The number of I/Os to the disk were split into multiple I/Os (LogicalDisk.SplitIOPerSec)",
			[]string{"volume"},
			nil,
		),

		ReadLatency: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "read_latency_seconds_total"),
			"Shows the average time, in seconds, of a read operation from the disk (LogicalDisk.AvgDiskSecPerRead)",
			[]string{"volume"},
			nil,
		),

		WriteLatency: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "write_latency_seconds_total"),
			"Shows the average time, in seconds, of a write operation to the disk (LogicalDisk.AvgDiskSecPerWrite)",
			[]string{"volume"},
			nil,
		),

		ReadWriteLatency: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "read_write_latency_seconds_total"),
			"Shows the time, in seconds, of the average disk transfer (LogicalDisk.AvgDiskSecPerTransfer)",
			[]string{"volume"},
			nil,
		),

		volumeWhitelistPattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *volumeWhitelist)),
		volumeBlacklistPattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *volumeBlacklist)),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *LogicalDiskCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ctx, ch); err != nil {
		log.Error("failed collecting logical_disk metrics:", desc, err)
		return err
	}
	return nil
}

// Win32_PerfRawData_PerfDisk_LogicalDisk docs:
// - https://msdn.microsoft.com/en-us/windows/hardware/aa394307(v=vs.71) - Win32_PerfRawData_PerfDisk_LogicalDisk class
// - https://msdn.microsoft.com/en-us/library/ms803973.aspx - LogicalDisk object reference
type logicalDisk struct {
	Name                   string
	CurrentDiskQueueLength float64 `perflib:"Current Disk Queue Length"`
	DiskReadBytesPerSec    float64 `perflib:"Disk Read Bytes/sec"`
	DiskReadsPerSec        float64 `perflib:"Disk Reads/sec"`
	DiskWriteBytesPerSec   float64 `perflib:"Disk Write Bytes/sec"`
	DiskWritesPerSec       float64 `perflib:"Disk Writes/sec"`
	PercentDiskReadTime    float64 `perflib:"% Disk Read Time"`
	PercentDiskWriteTime   float64 `perflib:"% Disk Write Time"`
	PercentFreeSpace       float64 `perflib:"% Free Space_Base"`
	PercentFreeSpace_Base  float64 `perflib:"Free Megabytes"`
	PercentIdleTime        float64 `perflib:"% Idle Time"`
	SplitIOPerSec          float64 `perflib:"Split IO/Sec"`
	AvgDiskSecPerRead      float64 `perflib:"Avg. Disk sec/Read"`
	AvgDiskSecPerWrite     float64 `perflib:"Avg. Disk sec/Write"`
	AvgDiskSecPerTransfer  float64 `perflib:"Avg. Disk sec/Transfer"`
}

func (c *LogicalDiskCollector) collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []logicalDisk
	if err := unmarshalObject(ctx.perfObjects["LogicalDisk"], &dst); err != nil {
		return nil, err
	}

	for _, volume := range dst {
		if volume.Name == "_Total" ||
			c.volumeBlacklistPattern.MatchString(volume.Name) ||
			!c.volumeWhitelistPattern.MatchString(volume.Name) {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.RequestsQueued,
			prometheus.GaugeValue,
			volume.CurrentDiskQueueLength,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReadBytesTotal,
			prometheus.CounterValue,
			volume.DiskReadBytesPerSec,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReadsTotal,
			prometheus.CounterValue,
			volume.DiskReadsPerSec,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WriteBytesTotal,
			prometheus.CounterValue,
			volume.DiskWriteBytesPerSec,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WritesTotal,
			prometheus.CounterValue,
			volume.DiskWritesPerSec,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReadTime,
			prometheus.CounterValue,
			volume.PercentDiskReadTime,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WriteTime,
			prometheus.CounterValue,
			volume.PercentDiskWriteTime,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FreeSpace,
			prometheus.GaugeValue,
			volume.PercentFreeSpace_Base*1024*1024,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalSpace,
			prometheus.GaugeValue,
			volume.PercentFreeSpace*1024*1024,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.IdleTime,
			prometheus.CounterValue,
			volume.PercentIdleTime,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SplitIOs,
			prometheus.CounterValue,
			volume.SplitIOPerSec,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReadLatency,
			prometheus.CounterValue,
			volume.AvgDiskSecPerRead*ticksToSecondsScaleFactor,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WriteLatency,
			prometheus.CounterValue,
			volume.AvgDiskSecPerWrite*ticksToSecondsScaleFactor,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReadWriteLatency,
			prometheus.CounterValue,
			volume.AvgDiskSecPerTransfer*ticksToSecondsScaleFactor,
			volume.Name,
		)
	}

	return nil, nil
}

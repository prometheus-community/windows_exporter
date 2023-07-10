//go:build windows
// +build windows

package collector

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	FlagLogicalDiskVolumeOldExclude = "collector.logical_disk.volume-blacklist"
	FlagLogicalDiskVolumeOldInclude = "collector.logical_disk.volume-whitelist"

	FlagLogicalDiskVolumeExclude = "collector.logical_disk.volume-exclude"
	FlagLogicalDiskVolumeInclude = "collector.logical_disk.volume-include"
)

var (
	volumeOldInclude *string
	volumeOldExclude *string

	volumeInclude *string
	volumeExclude *string

	volumeIncludeSet bool
	volumeExcludeSet bool
)

// A LogicalDiskCollector is a Prometheus collector for perflib logicalDisk metrics
type LogicalDiskCollector struct {
	logger log.Logger

	RequestsQueued   *prometheus.Desc
	AvgReadQueue     *prometheus.Desc
	AvgWriteQueue    *prometheus.Desc
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

	volumeIncludePattern *regexp.Regexp
	volumeExcludePattern *regexp.Regexp
}

// newLogicalDiskCollectorFlags ...
func newLogicalDiskCollectorFlags(app *kingpin.Application) {
	volumeInclude = app.Flag(
		FlagLogicalDiskVolumeInclude,
		"Regexp of volumes to include. Volume name must both match include and not match exclude to be included.",
	).Default(".+").PreAction(func(c *kingpin.ParseContext) error {
		volumeIncludeSet = true
		return nil
	}).String()

	volumeExclude = app.Flag(
		FlagLogicalDiskVolumeExclude,
		"Regexp of volumes to exclude. Volume name must both match include and not match exclude to be included.",
	).Default("").PreAction(func(c *kingpin.ParseContext) error {
		volumeExcludeSet = true
		return nil
	}).String()

	volumeOldInclude = app.Flag(
		FlagLogicalDiskVolumeOldInclude,
		"DEPRECATED: Use --collector.logical_disk.volume-include",
	).Hidden().String()
	volumeOldExclude = app.Flag(
		FlagLogicalDiskVolumeOldExclude,
		"DEPRECATED: Use --collector.logical_disk.volume-exclude",
	).Hidden().String()
}

// newLogicalDiskCollector ...
func newLogicalDiskCollector(logger log.Logger) (Collector, error) {
	const subsystem = "logical_disk"
	logger = log.With(logger, "collector", subsystem)

	if *volumeOldExclude != "" {
		if !volumeExcludeSet {
			_ = level.Warn(logger).Log("msg", "--collector.logical_disk.volume-blacklist is DEPRECATED and will be removed in a future release, use --collector.logical_disk.volume-exclude")
			*volumeExclude = *volumeOldExclude
		} else {
			return nil, errors.New("--collector.logical_disk.volume-blacklist and --collector.logical_disk.volume-exclude are mutually exclusive")
		}
	}
	if *volumeOldInclude != "" {
		if !volumeIncludeSet {
			_ = level.Warn(logger).Log("msg", "--collector.logical_disk.volume-whitelist is DEPRECATED and will be removed in a future release, use --collector.logical_disk.volume-include")
			*volumeInclude = *volumeOldInclude
		} else {
			return nil, errors.New("--collector.logical_disk.volume-whitelist and --collector.logical_disk.volume-include are mutually exclusive")
		}
	}

	return &LogicalDiskCollector{
		logger: logger,

		RequestsQueued: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "requests_queued"),
			"The number of requests queued to the disk (LogicalDisk.CurrentDiskQueueLength)",
			[]string{"volume"},
			nil,
		),

		AvgReadQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "avg_read_requests_queued"),
			"Average number of read requests that were queued for the selected disk during the sample interval (LogicalDisk.AvgDiskReadQueueLength)",
			[]string{"volume"},
			nil,
		),

		AvgWriteQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "avg_write_requests_queued"),
			"Average number of write requests that were queued for the selected disk during the sample interval (LogicalDisk.AvgDiskWriteQueueLength)",
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

		volumeIncludePattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *volumeInclude)),
		volumeExcludePattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *volumeExclude)),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *LogicalDiskCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ctx, ch); err != nil {
		_ = level.Error(c.logger).Log("failed collecting logical_disk metrics", "desc", desc, "err", err)
		return err
	}
	return nil
}

// Win32_PerfRawData_PerfDisk_LogicalDisk docs:
// - https://msdn.microsoft.com/en-us/windows/hardware/aa394307(v=vs.71) - Win32_PerfRawData_PerfDisk_LogicalDisk class
// - https://msdn.microsoft.com/en-us/library/ms803973.aspx - LogicalDisk object reference
type logicalDisk struct {
	Name                    string
	CurrentDiskQueueLength  float64 `perflib:"Current Disk Queue Length"`
	AvgDiskReadQueueLength  float64 `perflib:"Avg. Disk Read Queue Length"`
	AvgDiskWriteQueueLength float64 `perflib:"Avg. Disk Write Queue Length"`
	DiskReadBytesPerSec     float64 `perflib:"Disk Read Bytes/sec"`
	DiskReadsPerSec         float64 `perflib:"Disk Reads/sec"`
	DiskWriteBytesPerSec    float64 `perflib:"Disk Write Bytes/sec"`
	DiskWritesPerSec        float64 `perflib:"Disk Writes/sec"`
	PercentDiskReadTime     float64 `perflib:"% Disk Read Time"`
	PercentDiskWriteTime    float64 `perflib:"% Disk Write Time"`
	PercentFreeSpace        float64 `perflib:"% Free Space_Base"`
	PercentFreeSpace_Base   float64 `perflib:"Free Megabytes"`
	PercentIdleTime         float64 `perflib:"% Idle Time"`
	SplitIOPerSec           float64 `perflib:"Split IO/Sec"`
	AvgDiskSecPerRead       float64 `perflib:"Avg. Disk sec/Read"`
	AvgDiskSecPerWrite      float64 `perflib:"Avg. Disk sec/Write"`
	AvgDiskSecPerTransfer   float64 `perflib:"Avg. Disk sec/Transfer"`
}

func (c *LogicalDiskCollector) collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []logicalDisk
	if err := unmarshalObject(ctx.perfObjects["LogicalDisk"], &dst, c.logger); err != nil {
		return nil, err
	}

	for _, volume := range dst {
		if volume.Name == "_Total" ||
			c.volumeExcludePattern.MatchString(volume.Name) ||
			!c.volumeIncludePattern.MatchString(volume.Name) {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.RequestsQueued,
			prometheus.GaugeValue,
			volume.CurrentDiskQueueLength,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AvgReadQueue,
			prometheus.GaugeValue,
			volume.AvgDiskReadQueueLength*ticksToSecondsScaleFactor,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AvgWriteQueue,
			prometheus.GaugeValue,
			volume.AvgDiskWriteQueueLength*ticksToSecondsScaleFactor,
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

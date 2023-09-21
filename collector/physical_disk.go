//go:build windows
// +build windows

package collector

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	FlagPhysicalDiskExclude = "collector.physical_disk.disk-exclude"
	FlagPhysicalDiskInclude = "collector.physical_disk.disk-include"
)

var (
	diskInclude *string
	diskExclude *string

	diskIncludeSet bool
	diskExcludeSet bool
)

// A PhysicalDiskCollector is a Prometheus collector for perflib PhysicalDisk metrics
type PhysicalDiskCollector struct {
	logger log.Logger

	RequestsQueued   *prometheus.Desc
	ReadBytesTotal   *prometheus.Desc
	ReadsTotal       *prometheus.Desc
	WriteBytesTotal  *prometheus.Desc
	WritesTotal      *prometheus.Desc
	ReadTime         *prometheus.Desc
	WriteTime        *prometheus.Desc
	IdleTime         *prometheus.Desc
	SplitIOs         *prometheus.Desc
	ReadLatency      *prometheus.Desc
	WriteLatency     *prometheus.Desc
	ReadWriteLatency *prometheus.Desc

	diskIncludePattern *regexp.Regexp
	diskExcludePattern *regexp.Regexp
}

// newPhysicalDiskCollectorFlags ...
func newPhysicalDiskCollectorFlags(app *kingpin.Application) {
	diskInclude = app.Flag(
		FlagPhysicalDiskInclude,
		"Regexp of disks to include. Disk number must both match include and not match exclude to be included.",
	).Default(".+").PreAction(func(c *kingpin.ParseContext) error {
		diskIncludeSet = true
		return nil
	}).String()

	diskExclude = app.Flag(
		FlagPhysicalDiskExclude,
		"Regexp of disks to exclude. Disk number must both match include and not match exclude to be included.",
	).Default("").PreAction(func(c *kingpin.ParseContext) error {
		diskExcludeSet = true
		return nil
	}).String()
}

// NewPhysicalDiskCollector ...
func NewPhysicalDiskCollector(logger log.Logger) (Collector, error) {
	const subsystem = "physical_disk"
	logger = log.With(logger, "collector", subsystem)

	return &PhysicalDiskCollector{
		logger: logger,

		RequestsQueued: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "requests_queued"),
			"The number of requests queued to the disk (PhysicalDisk.CurrentDiskQueueLength)",
			[]string{"disk"},
			nil,
		),

		ReadBytesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "read_bytes_total"),
			"The number of bytes transferred from the disk during read operations (PhysicalDisk.DiskReadBytesPerSec)",
			[]string{"disk"},
			nil,
		),

		ReadsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "reads_total"),
			"The number of read operations on the disk (PhysicalDisk.DiskReadsPerSec)",
			[]string{"disk"},
			nil,
		),

		WriteBytesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "write_bytes_total"),
			"The number of bytes transferred to the disk during write operations (PhysicalDisk.DiskWriteBytesPerSec)",
			[]string{"disk"},
			nil,
		),

		WritesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "writes_total"),
			"The number of write operations on the disk (PhysicalDisk.DiskWritesPerSec)",
			[]string{"disk"},
			nil,
		),

		ReadTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "read_seconds_total"),
			"Seconds that the disk was busy servicing read requests (PhysicalDisk.PercentDiskReadTime)",
			[]string{"disk"},
			nil,
		),

		WriteTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "write_seconds_total"),
			"Seconds that the disk was busy servicing write requests (PhysicalDisk.PercentDiskWriteTime)",
			[]string{"disk"},
			nil,
		),

		IdleTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "idle_seconds_total"),
			"Seconds that the disk was idle (PhysicalDisk.PercentIdleTime)",
			[]string{"disk"},
			nil,
		),

		SplitIOs: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "split_ios_total"),
			"The number of I/Os to the disk were split into multiple I/Os (PhysicalDisk.SplitIOPerSec)",
			[]string{"disk"},
			nil,
		),

		ReadLatency: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "read_latency_seconds_total"),
			"Shows the average time, in seconds, of a read operation from the disk (PhysicalDisk.AvgDiskSecPerRead)",
			[]string{"disk"},
			nil,
		),

		WriteLatency: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "write_latency_seconds_total"),
			"Shows the average time, in seconds, of a write operation to the disk (PhysicalDisk.AvgDiskSecPerWrite)",
			[]string{"disk"},
			nil,
		),

		ReadWriteLatency: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "read_write_latency_seconds_total"),
			"Shows the time, in seconds, of the average disk transfer (PhysicalDisk.AvgDiskSecPerTransfer)",
			[]string{"disk"},
			nil,
		),

		diskIncludePattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *diskInclude)),
		diskExcludePattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *diskExclude)),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *PhysicalDiskCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ctx, ch); err != nil {
		_ = level.Error(c.logger).Log("failed collecting physical_disk metrics", "desc", desc, "err", err)
		return err
	}
	return nil
}

// Win32_PerfRawData_PerfDisk_PhysicalDisk docs:
// - https://docs.microsoft.com/en-us/previous-versions/aa394308(v=vs.85) - Win32_PerfRawData_PerfDisk_PhysicalDisk class
type PhysicalDisk struct {
	Name                   string
	CurrentDiskQueueLength float64 `perflib:"Current Disk Queue Length"`
	DiskReadBytesPerSec    float64 `perflib:"Disk Read Bytes/sec"`
	DiskReadsPerSec        float64 `perflib:"Disk Reads/sec"`
	DiskWriteBytesPerSec   float64 `perflib:"Disk Write Bytes/sec"`
	DiskWritesPerSec       float64 `perflib:"Disk Writes/sec"`
	PercentDiskReadTime    float64 `perflib:"% Disk Read Time"`
	PercentDiskWriteTime   float64 `perflib:"% Disk Write Time"`
	PercentIdleTime        float64 `perflib:"% Idle Time"`
	SplitIOPerSec          float64 `perflib:"Split IO/Sec"`
	AvgDiskSecPerRead      float64 `perflib:"Avg. Disk sec/Read"`
	AvgDiskSecPerWrite     float64 `perflib:"Avg. Disk sec/Write"`
	AvgDiskSecPerTransfer  float64 `perflib:"Avg. Disk sec/Transfer"`
}

func (c *PhysicalDiskCollector) collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []PhysicalDisk
	if err := unmarshalObject(ctx.perfObjects["PhysicalDisk"], &dst, c.logger); err != nil {
		return nil, err
	}

	for _, disk := range dst {
		if disk.Name == "_Total" ||
			c.diskExcludePattern.MatchString(disk.Name) ||
			!c.diskIncludePattern.MatchString(disk.Name) {
			continue
		}

		// Parse physical disk number from disk.Name. Mountpoint information is
		// sometimes included, e.g. "1 C:".
		disk_number, _, _ := strings.Cut(disk.Name, " ")

		ch <- prometheus.MustNewConstMetric(
			c.RequestsQueued,
			prometheus.GaugeValue,
			disk.CurrentDiskQueueLength,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReadBytesTotal,
			prometheus.CounterValue,
			disk.DiskReadBytesPerSec,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReadsTotal,
			prometheus.CounterValue,
			disk.DiskReadsPerSec,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WriteBytesTotal,
			prometheus.CounterValue,
			disk.DiskWriteBytesPerSec,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WritesTotal,
			prometheus.CounterValue,
			disk.DiskWritesPerSec,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReadTime,
			prometheus.CounterValue,
			disk.PercentDiskReadTime,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WriteTime,
			prometheus.CounterValue,
			disk.PercentDiskWriteTime,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.IdleTime,
			prometheus.CounterValue,
			disk.PercentIdleTime,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SplitIOs,
			prometheus.CounterValue,
			disk.SplitIOPerSec,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReadLatency,
			prometheus.CounterValue,
			disk.AvgDiskSecPerRead*ticksToSecondsScaleFactor,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WriteLatency,
			prometheus.CounterValue,
			disk.AvgDiskSecPerWrite*ticksToSecondsScaleFactor,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReadWriteLatency,
			prometheus.CounterValue,
			disk.AvgDiskSecPerTransfer*ticksToSecondsScaleFactor,
			disk_number,
		)
	}

	return nil, nil
}

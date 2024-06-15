//go:build windows

package physical_disk

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	Name                    = "physical_disk"
	FlagPhysicalDiskExclude = "collector.physical_disk.disk-exclude"
	FlagPhysicalDiskInclude = "collector.physical_disk.disk-include"
)

type Config struct {
	DiskInclude string `yaml:"disk_include"`
	DiskExclude string `yaml:"disk_exclude"`
}

var ConfigDefaults = Config{
	DiskInclude: ".+",
	DiskExclude: "",
}

// A collector is a Prometheus collector for perflib PhysicalDisk metrics
type collector struct {
	logger log.Logger

	diskInclude *string
	diskExclude *string

	diskIncludeSet bool
	diskExcludeSet bool

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

func New(logger log.Logger, config *Config) types.Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &collector{
		diskExclude: &config.DiskExclude,
		diskInclude: &config.DiskInclude,
	}
	c.SetLogger(logger)
	return c
}

func NewWithFlags(app *kingpin.Application) types.Collector {
	c := &collector{}

	c.diskInclude = app.Flag(
		FlagPhysicalDiskInclude,
		"Regexp of disks to include. Disk number must both match include and not match exclude to be included.",
	).Default(ConfigDefaults.DiskInclude).PreAction(func(_ *kingpin.ParseContext) error {
		c.diskIncludeSet = true
		return nil
	}).String()

	c.diskExclude = app.Flag(
		FlagPhysicalDiskExclude,
		"Regexp of disks to exclude. Disk number must both match include and not match exclude to be included.",
	).Default(ConfigDefaults.DiskExclude).PreAction(func(_ *kingpin.ParseContext) error {
		c.diskExcludeSet = true
		return nil
	}).String()
	return c
}

func (c *collector) GetName() string {
	return Name
}

func (c *collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *collector) GetPerfCounter() ([]string, error) {
	return []string{"PhysicalDisk"}, nil
}

func (c *collector) Build() error {
	c.RequestsQueued = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "requests_queued"),
		"The number of requests queued to the disk (PhysicalDisk.CurrentDiskQueueLength)",
		[]string{"disk"},
		nil,
	)

	c.ReadBytesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "read_bytes_total"),
		"The number of bytes transferred from the disk during read operations (PhysicalDisk.DiskReadBytesPerSec)",
		[]string{"disk"},
		nil,
	)

	c.ReadsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "reads_total"),
		"The number of read operations on the disk (PhysicalDisk.DiskReadsPerSec)",
		[]string{"disk"},
		nil,
	)

	c.WriteBytesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "write_bytes_total"),
		"The number of bytes transferred to the disk during write operations (PhysicalDisk.DiskWriteBytesPerSec)",
		[]string{"disk"},
		nil,
	)

	c.WritesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "writes_total"),
		"The number of write operations on the disk (PhysicalDisk.DiskWritesPerSec)",
		[]string{"disk"},
		nil,
	)

	c.ReadTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "read_seconds_total"),
		"Seconds that the disk was busy servicing read requests (PhysicalDisk.PercentDiskReadTime)",
		[]string{"disk"},
		nil,
	)

	c.WriteTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "write_seconds_total"),
		"Seconds that the disk was busy servicing write requests (PhysicalDisk.PercentDiskWriteTime)",
		[]string{"disk"},
		nil,
	)

	c.IdleTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "idle_seconds_total"),
		"Seconds that the disk was idle (PhysicalDisk.PercentIdleTime)",
		[]string{"disk"},
		nil,
	)

	c.SplitIOs = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "split_ios_total"),
		"The number of I/Os to the disk were split into multiple I/Os (PhysicalDisk.SplitIOPerSec)",
		[]string{"disk"},
		nil,
	)

	c.ReadLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "read_latency_seconds_total"),
		"Shows the average time, in seconds, of a read operation from the disk (PhysicalDisk.AvgDiskSecPerRead)",
		[]string{"disk"},
		nil,
	)

	c.WriteLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "write_latency_seconds_total"),
		"Shows the average time, in seconds, of a write operation to the disk (PhysicalDisk.AvgDiskSecPerWrite)",
		[]string{"disk"},
		nil,
	)

	c.ReadWriteLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "read_write_latency_seconds_total"),
		"Shows the time, in seconds, of the average disk transfer (PhysicalDisk.AvgDiskSecPerTransfer)",
		[]string{"disk"},
		nil,
	)

	var err error
	c.diskIncludePattern, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", *c.diskInclude))
	if err != nil {
		return err
	}

	c.diskExcludePattern, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", *c.diskExclude))
	if err != nil {
		return err
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *collector) Collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.collect(ctx, ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting physical_disk metrics", "err", err)
		return err
	}
	return nil
}

// PhysicalDisk
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

func (c *collector) collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var dst []PhysicalDisk
	if err := perflib.UnmarshalObject(ctx.PerfObjects["PhysicalDisk"], &dst, c.logger); err != nil {
		return err
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
			disk.AvgDiskSecPerRead*perflib.TicksToSecondScaleFactor,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WriteLatency,
			prometheus.CounterValue,
			disk.AvgDiskSecPerWrite*perflib.TicksToSecondScaleFactor,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReadWriteLatency,
			prometheus.CounterValue,
			disk.AvgDiskSecPerTransfer*perflib.TicksToSecondScaleFactor,
			disk_number,
		)
	}

	return nil
}

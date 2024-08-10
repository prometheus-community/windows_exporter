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

const Name = "physical_disk"

type Config struct {
	DiskInclude string `yaml:"disk_include"`
	DiskExclude string `yaml:"disk_exclude"`
}

var ConfigDefaults = Config{
	DiskInclude: ".+",
	DiskExclude: "",
}

// A Collector is a Prometheus Collector for perflib PhysicalDisk metrics.
type Collector struct {
	logger log.Logger

	diskInclude *string
	diskExclude *string

	diskIncludeSet bool
	diskExcludeSet bool

	diskIncludePattern *regexp.Regexp
	diskExcludePattern *regexp.Regexp

	idleTime         *prometheus.Desc
	readBytesTotal   *prometheus.Desc
	readLatency      *prometheus.Desc
	readTime         *prometheus.Desc
	readWriteLatency *prometheus.Desc
	readsTotal       *prometheus.Desc
	requestsQueued   *prometheus.Desc
	splitIOs         *prometheus.Desc
	writeBytesTotal  *prometheus.Desc
	writeLatency     *prometheus.Desc
	writeTime        *prometheus.Desc
	writesTotal      *prometheus.Desc
}

func New(logger log.Logger, config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		diskExclude: &config.DiskExclude,
		diskInclude: &config.DiskInclude,
	}
	c.SetLogger(logger)

	return c
}

func NewWithFlags(app *kingpin.Application) *Collector {
	c := &Collector{}

	c.diskInclude = app.Flag(
		"collector.physical_disk.disk-include",
		"Regexp of disks to include. Disk number must both match include and not match exclude to be included.",
	).Default(ConfigDefaults.DiskInclude).PreAction(func(_ *kingpin.ParseContext) error {
		c.diskIncludeSet = true
		return nil
	}).String()

	c.diskExclude = app.Flag(
		"collector.physical_disk.disk-exclude",
		"Regexp of disks to exclude. Disk number must both match include and not match exclude to be included.",
	).Default(ConfigDefaults.DiskExclude).PreAction(func(_ *kingpin.ParseContext) error {
		c.diskExcludeSet = true
		return nil
	}).String()

	return c
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *Collector) GetPerfCounter() ([]string, error) {
	return []string{"PhysicalDisk"}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build() error {
	c.requestsQueued = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "requests_queued"),
		"The number of requests queued to the disk (PhysicalDisk.CurrentDiskQueueLength)",
		[]string{"disk"},
		nil,
	)

	c.readBytesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "read_bytes_total"),
		"The number of bytes transferred from the disk during read operations (PhysicalDisk.DiskReadBytesPerSec)",
		[]string{"disk"},
		nil,
	)

	c.readsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "reads_total"),
		"The number of read operations on the disk (PhysicalDisk.DiskReadsPerSec)",
		[]string{"disk"},
		nil,
	)

	c.writeBytesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "write_bytes_total"),
		"The number of bytes transferred to the disk during write operations (PhysicalDisk.DiskWriteBytesPerSec)",
		[]string{"disk"},
		nil,
	)

	c.writesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "writes_total"),
		"The number of write operations on the disk (PhysicalDisk.DiskWritesPerSec)",
		[]string{"disk"},
		nil,
	)

	c.readTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "read_seconds_total"),
		"Seconds that the disk was busy servicing read requests (PhysicalDisk.PercentDiskReadTime)",
		[]string{"disk"},
		nil,
	)

	c.writeTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "write_seconds_total"),
		"Seconds that the disk was busy servicing write requests (PhysicalDisk.PercentDiskWriteTime)",
		[]string{"disk"},
		nil,
	)

	c.idleTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "idle_seconds_total"),
		"Seconds that the disk was idle (PhysicalDisk.PercentIdleTime)",
		[]string{"disk"},
		nil,
	)

	c.splitIOs = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "split_ios_total"),
		"The number of I/Os to the disk were split into multiple I/Os (PhysicalDisk.SplitIOPerSec)",
		[]string{"disk"},
		nil,
	)

	c.readLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "read_latency_seconds_total"),
		"Shows the average time, in seconds, of a read operation from the disk (PhysicalDisk.AvgDiskSecPerRead)",
		[]string{"disk"},
		nil,
	)

	c.writeLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "write_latency_seconds_total"),
		"Shows the average time, in seconds, of a write operation to the disk (PhysicalDisk.AvgDiskSecPerWrite)",
		[]string{"disk"},
		nil,
	)

	c.readWriteLatency = prometheus.NewDesc(
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
func (c *Collector) Collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.collect(ctx, ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting physical_disk metrics", "err", err)
		return err
	}
	return nil
}

// PhysicalDisk
// Win32_PerfRawData_PerfDisk_PhysicalDisk docs:
// - https://docs.microsoft.com/en-us/previous-versions/aa394308(v=vs.85) - Win32_PerfRawData_PerfDisk_PhysicalDisk class.
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

func (c *Collector) collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
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
			c.requestsQueued,
			prometheus.GaugeValue,
			disk.CurrentDiskQueueLength,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readBytesTotal,
			prometheus.CounterValue,
			disk.DiskReadBytesPerSec,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readsTotal,
			prometheus.CounterValue,
			disk.DiskReadsPerSec,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writeBytesTotal,
			prometheus.CounterValue,
			disk.DiskWriteBytesPerSec,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writesTotal,
			prometheus.CounterValue,
			disk.DiskWritesPerSec,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readTime,
			prometheus.CounterValue,
			disk.PercentDiskReadTime,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writeTime,
			prometheus.CounterValue,
			disk.PercentDiskWriteTime,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.idleTime,
			prometheus.CounterValue,
			disk.PercentIdleTime,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.splitIOs,
			prometheus.CounterValue,
			disk.SplitIOPerSec,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readLatency,
			prometheus.CounterValue,
			disk.AvgDiskSecPerRead*perflib.TicksToSecondScaleFactor,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writeLatency,
			prometheus.CounterValue,
			disk.AvgDiskSecPerWrite*perflib.TicksToSecondScaleFactor,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readWriteLatency,
			prometheus.CounterValue,
			disk.AvgDiskSecPerTransfer*perflib.TicksToSecondScaleFactor,
			disk_number,
		)
	}

	return nil
}

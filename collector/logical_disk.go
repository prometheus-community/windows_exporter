// returns data points from Win32_PerfRawData_PerfDisk_LogicalDisk
// https://msdn.microsoft.com/en-us/windows/hardware/aa394307(v=vs.71) - Win32_PerfRawData_PerfDisk_LogicalDisk class
// https://msdn.microsoft.com/en-us/library/ms803973.aspx - LogicalDisk object reference

package collector

import (
	"fmt"
	"regexp"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"gopkg.in/alecthomas/kingpin.v2"
)

func init() {
	Factories["logical_disk"] = NewLogicalDiskCollector
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

// A LogicalDiskCollector is a Prometheus collector for WMI Win32_PerfRawData_PerfDisk_LogicalDisk metrics
type LogicalDiskCollector struct {
	RequestsQueued  *prometheus.Desc
	ReadBytesTotal  *prometheus.Desc
	ReadsTotal      *prometheus.Desc
	WriteBytesTotal *prometheus.Desc
	WritesTotal     *prometheus.Desc
	ReadTime        *prometheus.Desc
	WriteTime       *prometheus.Desc
	TotalSpace      *prometheus.Desc
	FreeSpace       *prometheus.Desc
	IdleTime        *prometheus.Desc
	SplitIOs        *prometheus.Desc

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
			"Free space in bytes (LogicalDisk.PercentFreeSpace)",
			[]string{"volume"},
			nil,
		),

		TotalSpace: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "size_bytes"),
			"Total space in bytes (LogicalDisk.PercentFreeSpace_Base)",
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

		volumeWhitelistPattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *volumeWhitelist)),
		volumeBlacklistPattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *volumeBlacklist)),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *LogicalDiskCollector) Collect(ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting logical_disk metrics:", desc, err)
		return err
	}
	return nil
}

type Win32_PerfRawData_PerfDisk_LogicalDisk struct {
	Name                   string
	CurrentDiskQueueLength uint32
	DiskReadBytesPerSec    uint64
	DiskReadsPerSec        uint32
	DiskWriteBytesPerSec   uint64
	DiskWritesPerSec       uint32
	PercentDiskReadTime    uint64
	PercentDiskWriteTime   uint64
	PercentFreeSpace       uint32
	PercentFreeSpace_Base  uint32
	PercentIdleTime        uint64
	SplitIOPerSec          uint32
}

func (c *LogicalDiskCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_PerfDisk_LogicalDisk
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
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
			float64(volume.CurrentDiskQueueLength),
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReadBytesTotal,
			prometheus.CounterValue,
			float64(volume.DiskReadBytesPerSec),
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReadsTotal,
			prometheus.CounterValue,
			float64(volume.DiskReadsPerSec),
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WriteBytesTotal,
			prometheus.CounterValue,
			float64(volume.DiskWriteBytesPerSec),
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WritesTotal,
			prometheus.CounterValue,
			float64(volume.DiskWritesPerSec),
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReadTime,
			prometheus.CounterValue,
			float64(volume.PercentDiskReadTime)*ticksToSecondsScaleFactor,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WriteTime,
			prometheus.CounterValue,
			float64(volume.PercentDiskWriteTime)*ticksToSecondsScaleFactor,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FreeSpace,
			prometheus.GaugeValue,
			float64(volume.PercentFreeSpace)*1024*1024,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalSpace,
			prometheus.GaugeValue,
			float64(volume.PercentFreeSpace_Base)*1024*1024,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.IdleTime,
			prometheus.CounterValue,
			float64(volume.PercentIdleTime)*ticksToSecondsScaleFactor,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SplitIOs,
			prometheus.CounterValue,
			float64(volume.SplitIOPerSec),
			volume.Name,
		)
	}

	return nil, nil
}

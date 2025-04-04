// Copyright 2024 The Prometheus Authors
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

package physical_disk

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "physical_disk"

type Config struct {
	DiskInclude *regexp.Regexp `yaml:"disk_include"`
	DiskExclude *regexp.Regexp `yaml:"disk_exclude"`
}

//nolint:gochecknoglobals
var ConfigDefaults = Config{
	DiskInclude: types.RegExpAny,
	DiskExclude: types.RegExpEmpty,
}

// A Collector is a Prometheus Collector for perflib PhysicalDisk metrics.
type Collector struct {
	config Config

	perfDataCollector *pdh.Collector
	perfDataObject    []perfDataCounterValues

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

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	if config.DiskExclude == nil {
		config.DiskExclude = ConfigDefaults.DiskExclude
	}

	if config.DiskInclude == nil {
		config.DiskInclude = ConfigDefaults.DiskInclude
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(app *kingpin.Application) *Collector {
	c := &Collector{
		config: ConfigDefaults,
	}

	var diskExclude, diskInclude string

	app.Flag(
		"collector.physical_disk.disk-exclude",
		"Regexp of disks to exclude. Disk number must both match include and not match exclude to be included.",
	).Default("").StringVar(&diskExclude)

	app.Flag(
		"collector.physical_disk.disk-include",
		"Regexp of disks to include. Disk number must both match include and not match exclude to be included.",
	).Default(".+").StringVar(&diskInclude)

	app.Action(func(*kingpin.ParseContext) error {
		var err error

		c.config.DiskExclude, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", diskExclude))
		if err != nil {
			return fmt.Errorf("collector.physical_disk.disk-exclude: %w", err)
		}

		c.config.DiskInclude, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", diskInclude))
		if err != nil {
			return fmt.Errorf("collector.physical_disk.disk-include: %w", err)
		}

		return nil
	})

	return c
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *mi.Session) error {
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

	c.perfDataCollector, err = pdh.NewCollector[perfDataCounterValues](pdh.CounterTypeRaw, "PhysicalDisk", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create PhysicalDisk collector: %w", err)
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollector.Collect(&c.perfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect PhysicalDisk metrics: %w", err)
	}

	for _, data := range c.perfDataObject {
		if c.config.DiskExclude.MatchString(data.Name) ||
			!c.config.DiskInclude.MatchString(data.Name) {
			continue
		}

		// Parse physical disk number from disk.Name. Mountpoint information is
		// sometimes included, e.g. "1 C:".
		disk_number, _, _ := strings.Cut(data.Name, " ")

		ch <- prometheus.MustNewConstMetric(
			c.requestsQueued,
			prometheus.GaugeValue,
			data.CurrentDiskQueueLength,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readBytesTotal,
			prometheus.CounterValue,
			data.DiskReadBytesPerSec,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readsTotal,
			prometheus.CounterValue,
			data.DiskReadsPerSec,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writeBytesTotal,
			prometheus.CounterValue,
			data.DiskWriteBytesPerSec,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writesTotal,
			prometheus.CounterValue,
			data.DiskWritesPerSec,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readTime,
			prometheus.CounterValue,
			data.PercentDiskReadTime,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writeTime,
			prometheus.CounterValue,
			data.PercentDiskWriteTime,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.idleTime,
			prometheus.CounterValue,
			data.PercentIdleTime,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.splitIOs,
			prometheus.CounterValue,
			data.SplitIOPerSec,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readLatency,
			prometheus.CounterValue,
			data.AvgDiskSecPerRead*pdh.TicksToSecondScaleFactor,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writeLatency,
			prometheus.CounterValue,
			data.AvgDiskSecPerWrite*pdh.TicksToSecondScaleFactor,
			disk_number,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readWriteLatency,
			prometheus.CounterValue,
			data.AvgDiskSecPerTransfer*pdh.TicksToSecondScaleFactor,
			disk_number,
		)
	}

	return nil
}

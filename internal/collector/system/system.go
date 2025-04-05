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

package system

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/headers/kernel32"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "system"

type Config struct{}

//nolint:gochecknoglobals
var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI metrics.
type Collector struct {
	config Config

	bootTimeTimestamp float64

	perfDataCollector *pdh.Collector
	perfDataObject    []perfDataCounterValues

	contextSwitchesTotal     *prometheus.Desc
	exceptionDispatchesTotal *prometheus.Desc
	processorQueueLength     *prometheus.Desc
	processes                *prometheus.Desc
	processesLimit           *prometheus.Desc
	systemCallsTotal         *prometheus.Desc
	// Deprecated: Use windows_system_boot_time_timestamp instead
	bootTimeSeconds *prometheus.Desc
	bootTime        *prometheus.Desc
	threads         *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	c.perfDataCollector.Close()

	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *mi.Session) error {
	c.bootTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "boot_time_timestamp"),
		"Unix timestamp of system boot time",
		nil,
		nil,
	)
	c.bootTimeSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "boot_time_timestamp_seconds"),
		"Deprecated: Use windows_system_boot_time_timestamp instead",
		nil,
		nil,
	)
	c.contextSwitchesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "context_switches_total"),
		"Total number of context switches (WMI source: PerfOS_System.ContextSwitchesPersec)",
		nil,
		nil,
	)
	c.exceptionDispatchesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "exception_dispatches_total"),
		"Total number of exceptions dispatched (WMI source: PerfOS_System.ExceptionDispatchesPersec)",
		nil,
		nil,
	)
	c.processes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "processes"),
		"Current number of processes (WMI source: PerfOS_System.Processes)",
		nil,
		nil,
	)
	c.processesLimit = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "processes_limit"),
		"Maximum number of processes.",
		nil,
		nil,
	)

	c.processorQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "processor_queue_length"),
		"Length of processor queue (WMI source: PerfOS_System.ProcessorQueueLength)",
		nil,
		nil,
	)
	c.systemCallsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "system_calls_total"),
		"Total number of system calls (WMI source: PerfOS_System.SystemCallsPersec)",
		nil,
		nil,
	)
	c.threads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "threads"),
		"Current number of threads (WMI source: PerfOS_System.Threads)",
		nil,
		nil,
	)

	c.bootTimeTimestamp = float64(time.Now().Unix() - int64(kernel32.GetTickCount64()/1000))

	var err error

	c.perfDataCollector, err = pdh.NewCollector[perfDataCounterValues](pdh.CounterTypeRaw, "System", nil)
	if err != nil {
		return fmt.Errorf("failed to create System collector: %w", err)
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollector.Collect(&c.perfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect System metrics: %w", err)
	}

	ch <- prometheus.MustNewConstMetric(
		c.contextSwitchesTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].ContextSwitchesPerSec,
	)
	ch <- prometheus.MustNewConstMetric(
		c.exceptionDispatchesTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].ExceptionDispatchesPerSec,
	)
	ch <- prometheus.MustNewConstMetric(
		c.processorQueueLength,
		prometheus.GaugeValue,
		c.perfDataObject[0].ProcessorQueueLength,
	)
	ch <- prometheus.MustNewConstMetric(
		c.processes,
		prometheus.GaugeValue,
		c.perfDataObject[0].Processes,
	)
	ch <- prometheus.MustNewConstMetric(
		c.systemCallsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].SystemCallsPerSec,
	)
	ch <- prometheus.MustNewConstMetric(
		c.threads,
		prometheus.GaugeValue,
		c.perfDataObject[0].Threads,
	)

	ch <- prometheus.MustNewConstMetric(
		c.bootTimeSeconds,
		prometheus.GaugeValue,
		c.bootTimeTimestamp,
	)

	ch <- prometheus.MustNewConstMetric(
		c.bootTime,
		prometheus.GaugeValue,
		c.bootTimeTimestamp,
	)

	// Windows has no defined limit, and is based off available resources. This currently isn't calculated by WMI and is set to default value.
	// https://techcommunity.microsoft.com/t5/windows-blog-archive/pushing-the-limits-of-windows-processes-and-threads/ba-p/723824
	// https://docs.microsoft.com/en-us/windows/win32/cimwin32prov/win32-operatingsystem
	ch <- prometheus.MustNewConstMetric(
		c.processesLimit,
		prometheus.GaugeValue,
		float64(4294967295),
	)

	return nil
}

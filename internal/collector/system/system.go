//go:build windows

package system

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "system"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI metrics.
type Collector struct {
	config Config

	perfDataCollector *perfdata.Collector

	contextSwitchesTotal     *prometheus.Desc
	exceptionDispatchesTotal *prometheus.Desc
	processorQueueLength     *prometheus.Desc
	processes                *prometheus.Desc
	processesLimit           *prometheus.Desc
	systemCallsTotal         *prometheus.Desc
	systemUpTime             *prometheus.Desc
	threads                  *prometheus.Desc
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
	var err error

	c.perfDataCollector, err = perfdata.NewCollector("System", nil, []string{
		contextSwitchesPersec,
		exceptionDispatchesPersec,
		processorQueueLength,
		systemCallsPersec,
		systemUpTime,
		processes,
		threads,
	})
	if err != nil {
		return fmt.Errorf("failed to create System collector: %w", err)
	}

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
	c.systemUpTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "system_up_time"),
		"System boot time (WMI source: PerfOS_System.SystemUpTime)",
		nil,
		nil,
	)
	c.threads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "threads"),
		"Current number of threads (WMI source: PerfOS_System.Threads)",
		nil,
		nil,
	)

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect System metrics: %w", err)
	}

	data, ok := perfData[perfdata.EmptyInstance]
	if !ok {
		return errors.New("query for System returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.contextSwitchesTotal,
		prometheus.CounterValue,
		data[contextSwitchesPersec].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.exceptionDispatchesTotal,
		prometheus.CounterValue,
		data[exceptionDispatchesPersec].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.processorQueueLength,
		prometheus.GaugeValue,
		data[processorQueueLength].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.processes,
		prometheus.GaugeValue,
		data[processes].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.systemCallsTotal,
		prometheus.CounterValue,
		data[systemCallsPersec].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.systemUpTime,
		prometheus.GaugeValue,
		data[systemUpTime].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.threads,
		prometheus.GaugeValue,
		data[threads].FirstValue,
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

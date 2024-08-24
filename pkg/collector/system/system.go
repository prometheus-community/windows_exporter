//go:build windows

package system

import (
	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "system"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI metrics.
type Collector struct {
	config Config

	contextSwitchesTotal     *prometheus.Desc
	exceptionDispatchesTotal *prometheus.Desc
	processorQueueLength     *prometheus.Desc
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

func (c *Collector) GetPerfCounter(_ log.Logger) ([]string, error) {
	return []string{"System"}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(_ log.Logger) error {
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
func (c *Collector) Collect(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	if err := c.collect(ctx, logger, ch); err != nil {
		_ = level.Error(logger).Log("msg", "failed collecting system metrics", "err", err)
		return err
	}
	return nil
}

// Win32_PerfRawData_PerfOS_System docs:
// - https://web.archive.org/web/20050830140516/http://msdn.microsoft.com/library/en-us/wmisdk/wmi/win32_perfrawdata_perfos_system.asp
type system struct {
	ContextSwitchesPersec     float64 `perflib:"Context Switches/sec"`
	ExceptionDispatchesPersec float64 `perflib:"Exception Dispatches/sec"`
	ProcessorQueueLength      float64 `perflib:"Processor Queue Length"`
	SystemCallsPersec         float64 `perflib:"System Calls/sec"`
	SystemUpTime              float64 `perflib:"System Up Time"`
	Threads                   float64 `perflib:"Threads"`
}

func (c *Collector) collect(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	var dst []system
	if err := perflib.UnmarshalObject(ctx.PerfObjects["System"], &dst, logger); err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		c.contextSwitchesTotal,
		prometheus.CounterValue,
		dst[0].ContextSwitchesPersec,
	)
	ch <- prometheus.MustNewConstMetric(
		c.exceptionDispatchesTotal,
		prometheus.CounterValue,
		dst[0].ExceptionDispatchesPersec,
	)
	ch <- prometheus.MustNewConstMetric(
		c.processorQueueLength,
		prometheus.GaugeValue,
		dst[0].ProcessorQueueLength,
	)
	ch <- prometheus.MustNewConstMetric(
		c.systemCallsTotal,
		prometheus.CounterValue,
		dst[0].SystemCallsPersec,
	)
	ch <- prometheus.MustNewConstMetric(
		c.systemUpTime,
		prometheus.GaugeValue,
		dst[0].SystemUpTime,
	)
	ch <- prometheus.MustNewConstMetric(
		c.threads,
		prometheus.GaugeValue,
		dst[0].Threads,
	)
	return nil
}

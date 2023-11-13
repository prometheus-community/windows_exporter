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

// A collector is a Prometheus collector for WMI metrics
type collector struct {
	logger log.Logger

	ContextSwitchesTotal     *prometheus.Desc
	ExceptionDispatchesTotal *prometheus.Desc
	ProcessorQueueLength     *prometheus.Desc
	SystemCallsTotal         *prometheus.Desc
	SystemUpTime             *prometheus.Desc
	Threads                  *prometheus.Desc
}

func New(logger log.Logger, _ *Config) types.Collector {
	c := &collector{}
	c.SetLogger(logger)
	return c
}

func NewWithFlags(_ *kingpin.Application) types.Collector {
	return &collector{}
}

func (c *collector) GetName() string {
	return Name
}

func (c *collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *collector) GetPerfCounter() ([]string, error) {
	return []string{"System"}, nil
}

func (c *collector) Build() error {
	c.ContextSwitchesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "context_switches_total"),
		"Total number of context switches (WMI source: PerfOS_System.ContextSwitchesPersec)",
		nil,
		nil,
	)
	c.ExceptionDispatchesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "exception_dispatches_total"),
		"Total number of exceptions dispatched (WMI source: PerfOS_System.ExceptionDispatchesPersec)",
		nil,
		nil,
	)
	c.ProcessorQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "processor_queue_length"),
		"Length of processor queue (WMI source: PerfOS_System.ProcessorQueueLength)",
		nil,
		nil,
	)
	c.SystemCallsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "system_calls_total"),
		"Total number of system calls (WMI source: PerfOS_System.SystemCallsPersec)",
		nil,
		nil,
	)
	c.SystemUpTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "system_up_time"),
		"System boot time (WMI source: PerfOS_System.SystemUpTime)",
		nil,
		nil,
	)
	c.Threads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "threads"),
		"Current number of threads (WMI source: PerfOS_System.Threads)",
		nil,
		nil,
	)
	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *collector) Collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ctx, ch); err != nil {
		_ = level.Error(c.logger).Log("failed collecting system metrics", "desc", desc, "err", err)
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

func (c *collector) collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []system
	if err := perflib.UnmarshalObject(ctx.PerfObjects["System"], &dst, c.logger); err != nil {
		return nil, err
	}

	ch <- prometheus.MustNewConstMetric(
		c.ContextSwitchesTotal,
		prometheus.CounterValue,
		dst[0].ContextSwitchesPersec,
	)
	ch <- prometheus.MustNewConstMetric(
		c.ExceptionDispatchesTotal,
		prometheus.CounterValue,
		dst[0].ExceptionDispatchesPersec,
	)
	ch <- prometheus.MustNewConstMetric(
		c.ProcessorQueueLength,
		prometheus.GaugeValue,
		dst[0].ProcessorQueueLength,
	)
	ch <- prometheus.MustNewConstMetric(
		c.SystemCallsTotal,
		prometheus.CounterValue,
		dst[0].SystemCallsPersec,
	)
	ch <- prometheus.MustNewConstMetric(
		c.SystemUpTime,
		prometheus.GaugeValue,
		dst[0].SystemUpTime,
	)
	ch <- prometheus.MustNewConstMetric(
		c.Threads,
		prometheus.GaugeValue,
		dst[0].Threads,
	)
	return nil, nil
}

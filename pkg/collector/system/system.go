//go:build windows

package system

import (
	"errors"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
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

func (c *Collector) GetPerfCounter(_ *slog.Logger) ([]string, error) {
	return []string{"System"}, nil
}

func (c *Collector) Close(_ *slog.Logger) error {
	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *wmi.Client) error {
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
func (c *Collector) Collect(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	logger = logger.With(slog.String("collector", Name))
	if err := c.collect(ctx, logger, ch); err != nil {
		logger.Error("failed collecting system metrics",
			slog.Any("err", err),
		)

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
	Processes                 float64 `perflib:"Processes"`
	Threads                   float64 `perflib:"Threads"`
}

func (c *Collector) collect(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	logger = logger.With(slog.String("collector", Name))

	var dst []system

	if err := perflib.UnmarshalObject(ctx.PerfObjects["System"], &dst, logger); err != nil {
		return err
	}

	if len(dst) == 0 {
		return errors.New("no data returned from Performance Counter")
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
		c.processes,
		prometheus.GaugeValue,
		dst[0].Processes,
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

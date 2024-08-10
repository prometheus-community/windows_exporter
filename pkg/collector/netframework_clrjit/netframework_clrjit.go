//go:build windows

package netframework_clrjit

import (
	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "netframework_clrjit"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_NETFramework_NETCLRJit metrics
type Collector struct {
	logger log.Logger

	numberOfMethodsJitted      *prometheus.Desc
	timeInJit                  *prometheus.Desc
	standardJitFailures        *prometheus.Desc
	totalNumberOfILBytesJitted *prometheus.Desc
}

func New(logger log.Logger, _ *Config) *Collector {
	c := &Collector{}
	c.SetLogger(logger)

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *Collector) GetPerfCounter() ([]string, error) {
	return []string{}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build() error {
	c.numberOfMethodsJitted = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "jit_methods_total"),
		"Displays the total number of methods JIT-compiled since the application started. This counter does not include pre-JIT-compiled methods.",
		[]string{"process"},
		nil,
	)
	c.timeInJit = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "jit_time_percent"),
		"Displays the percentage of time spent in JIT compilation. This counter is updated at the end of every JIT compilation phase. A JIT compilation phase occurs when a method and its dependencies are compiled.",
		[]string{"process"},
		nil,
	)
	c.standardJitFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "jit_standard_failures_total"),
		"Displays the peak number of methods the JIT compiler has failed to compile since the application started. This failure can occur if the MSIL cannot be verified or if there is an internal error in the JIT compiler.",
		[]string{"process"},
		nil,
	)
	c.totalNumberOfILBytesJitted = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "jit_il_bytes_total"),
		"Displays the total number of Microsoft intermediate language (MSIL) bytes compiled by the just-in-time (JIT) compiler since the application started",
		[]string{"process"},
		nil,
	)
	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.collect(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting win32_perfrawdata_netframework_netclrjit metrics", "err", err)
		return err
	}
	return nil
}

type Win32_PerfRawData_NETFramework_NETCLRJit struct {
	Name string

	Frequency_PerfTime         uint32
	ILBytesJittedPersec        uint32
	NumberofILBytesJitted      uint32
	NumberofMethodsJitted      uint32
	PercentTimeinJit           uint32
	StandardJitFailures        uint32
	TotalNumberofILBytesJitted uint32
}

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_NETFramework_NETCLRJit
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}

	for _, process := range dst {
		if process.Name == "_Global_" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.numberOfMethodsJitted,
			prometheus.CounterValue,
			float64(process.NumberofMethodsJitted),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.timeInJit,
			prometheus.GaugeValue,
			float64(process.PercentTimeinJit)/float64(process.Frequency_PerfTime),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.standardJitFailures,
			prometheus.GaugeValue,
			float64(process.StandardJitFailures),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.totalNumberOfILBytesJitted,
			prometheus.CounterValue,
			float64(process.TotalNumberofILBytesJitted),
			process.Name,
		)
	}

	return nil
}

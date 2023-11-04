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

// A collector is a Prometheus collector for WMI Win32_PerfRawData_NETFramework_NETCLRJit metrics
type collector struct {
	logger log.Logger

	NumberofMethodsJitted      *prometheus.Desc
	TimeinJit                  *prometheus.Desc
	StandardJitFailures        *prometheus.Desc
	TotalNumberofILBytesJitted *prometheus.Desc
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
	return []string{}, nil
}

func (c *collector) Build() error {
	c.NumberofMethodsJitted = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "jit_methods_total"),
		"Displays the total number of methods JIT-compiled since the application started. This counter does not include pre-JIT-compiled methods.",
		[]string{"process"},
		nil,
	)
	c.TimeinJit = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "jit_time_percent"),
		"Displays the percentage of time spent in JIT compilation. This counter is updated at the end of every JIT compilation phase. A JIT compilation phase occurs when a method and its dependencies are compiled.",
		[]string{"process"},
		nil,
	)
	c.StandardJitFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "jit_standard_failures_total"),
		"Displays the peak number of methods the JIT compiler has failed to compile since the application started. This failure can occur if the MSIL cannot be verified or if there is an internal error in the JIT compiler.",
		[]string{"process"},
		nil,
	)
	c.TotalNumberofILBytesJitted = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "jit_il_bytes_total"),
		"Displays the total number of Microsoft intermediate language (MSIL) bytes compiled by the just-in-time (JIT) compiler since the application started",
		[]string{"process"},
		nil,
	)
	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		_ = level.Error(c.logger).Log("failed collecting win32_perfrawdata_netframework_netclrjit metrics", "desc", desc, "err", err)
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

func (c *collector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_NETFramework_NETCLRJit
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	for _, process := range dst {

		if process.Name == "_Global_" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.NumberofMethodsJitted,
			prometheus.CounterValue,
			float64(process.NumberofMethodsJitted),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TimeinJit,
			prometheus.GaugeValue,
			float64(process.PercentTimeinJit)/float64(process.Frequency_PerfTime),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.StandardJitFailures,
			prometheus.GaugeValue,
			float64(process.StandardJitFailures),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalNumberofILBytesJitted,
			prometheus.CounterValue,
			float64(process.TotalNumberofILBytesJitted),
			process.Name,
		)
	}

	return nil, nil
}

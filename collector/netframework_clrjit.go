//go:build windows
// +build windows

package collector

import (
	"github.com/StackExchange/wmi"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("netframework_clrjit", NewNETFramework_NETCLRJitCollector)
}

// A NETFramework_NETCLRJitCollector is a Prometheus collector for WMI Win32_PerfRawData_NETFramework_NETCLRJit metrics
type NETFramework_NETCLRJitCollector struct {
	NumberofMethodsJitted      *prometheus.Desc
	TimeinJit                  *prometheus.Desc
	StandardJitFailures        *prometheus.Desc
	TotalNumberofILBytesJitted *prometheus.Desc
}

// NewNETFramework_NETCLRJitCollector ...
func NewNETFramework_NETCLRJitCollector() (Collector, error) {
	const subsystem = "netframework_clrjit"
	return &NETFramework_NETCLRJitCollector{
		NumberofMethodsJitted: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "jit_methods_total"),
			"Displays the total number of methods JIT-compiled since the application started. This counter does not include pre-JIT-compiled methods.",
			[]string{"process"},
			nil,
		),
		TimeinJit: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "jit_time_percent"),
			"Displays the percentage of time spent in JIT compilation. This counter is updated at the end of every JIT compilation phase. A JIT compilation phase occurs when a method and its dependencies are compiled.",
			[]string{"process"},
			nil,
		),
		StandardJitFailures: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "jit_standard_failures_total"),
			"Displays the peak number of methods the JIT compiler has failed to compile since the application started. This failure can occur if the MSIL cannot be verified or if there is an internal error in the JIT compiler.",
			[]string{"process"},
			nil,
		),
		TotalNumberofILBytesJitted: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "jit_il_bytes_total"),
			"Displays the total number of Microsoft intermediate language (MSIL) bytes compiled by the just-in-time (JIT) compiler since the application started",
			[]string{"process"},
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *NETFramework_NETCLRJitCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting win32_perfrawdata_netframework_netclrjit metrics:", desc, err)
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

func (c *NETFramework_NETCLRJitCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_NETFramework_NETCLRJit
	q := queryAll(&dst)
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

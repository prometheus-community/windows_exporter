//go:build windows

package netframework

import (
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

func (c *Collector) buildClrJIT() {
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

func (c *Collector) collectClrJIT(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_NETFramework_NETCLRJit
	if err := c.wmiClient.Query("SELECT * FROM Win32_PerfRawData_NETFramework_NETCLRJit", &dst); err != nil {
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

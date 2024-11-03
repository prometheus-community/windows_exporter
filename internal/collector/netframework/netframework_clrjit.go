//go:build windows

package netframework

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
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
	Name string `mi:"Name"`

	Frequency_PerfTime         uint32 `mi:"Frequency_PerfTime"`
	ILBytesJittedPersec        uint32 `mi:"ILBytesJittedPersec"`
	NumberofILBytesJitted      uint32 `mi:"NumberofILBytesJitted"`
	NumberofMethodsJitted      uint32 `mi:"NumberofMethodsJitted"`
	PercentTimeinJit           uint32 `mi:"PercentTimeinJit"`
	StandardJitFailures        uint32 `mi:"StandardJitFailures"`
	TotalNumberofILBytesJitted uint32 `mi:"TotalNumberofILBytesJitted"`
}

func (c *Collector) collectClrJIT(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_NETFramework_NETCLRJit
	if err := c.miSession.Query(&dst, mi.NamespaceRootCIMv2, utils.Must(mi.NewQuery("SELECT * Win32_PerfRawData_NETFramework_NETCLRJit"))); err != nil {
		return fmt.Errorf("WMI query failed: %w", err)
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

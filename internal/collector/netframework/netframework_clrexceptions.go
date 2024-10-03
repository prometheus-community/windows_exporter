//go:build windows

package netframework

import (
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

func (c *Collector) buildClrExceptions() {
	c.numberOfExceptionsThrown = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "exceptions_thrown_total"),
		"Displays the total number of exceptions thrown since the application started. This includes both .NET exceptions and unmanaged exceptions that are converted into .NET exceptions.",
		[]string{"process"},
		nil,
	)
	c.numberOfFilters = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "exceptions_filters_total"),
		"Displays the total number of .NET exception filters executed. An exception filter evaluates regardless of whether an exception is handled.",
		[]string{"process"},
		nil,
	)
	c.numberOfFinally = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "exceptions_finallys_total"),
		"Displays the total number of finally blocks executed. Only the finally blocks executed for an exception are counted; finally blocks on normal code paths are not counted by this counter.",
		[]string{"process"},
		nil,
	)
	c.throwToCatchDepth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "throw_to_catch_depth_total"),
		"Displays the total number of stack frames traversed, from the frame that threw the exception to the frame that handled the exception.",
		[]string{"process"},
		nil,
	)
}

type Win32_PerfRawData_NETFramework_NETCLRExceptions struct {
	Name string

	NumberofExcepsThrown       uint32
	NumberofExcepsThrownPersec uint32
	NumberofFiltersPersec      uint32
	NumberofFinallysPersec     uint32
	ThrowToCatchDepthPersec    uint32
}

func (c *Collector) collectClrExceptions(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_NETFramework_NETCLRExceptions
	if err := c.wmiClient.Query("SELECT * FROM Win32_PerfRawData_NETFramework_NETCLRExceptions", &dst); err != nil {
		return err
	}

	for _, process := range dst {
		if process.Name == "_Global_" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.numberOfExceptionsThrown,
			prometheus.CounterValue,
			float64(process.NumberofExcepsThrown),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberOfFilters,
			prometheus.CounterValue,
			float64(process.NumberofFiltersPersec),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberOfFinally,
			prometheus.CounterValue,
			float64(process.NumberofFinallysPersec),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.throwToCatchDepth,
			prometheus.CounterValue,
			float64(process.ThrowToCatchDepthPersec),
			process.Name,
		)
	}

	return nil
}

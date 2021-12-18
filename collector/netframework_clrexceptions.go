//go:build windows
// +build windows

package collector

import (
	"github.com/StackExchange/wmi"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("netframework_clrexceptions", NewNETFramework_NETCLRExceptionsCollector)
}

// A NETFramework_NETCLRExceptionsCollector is a Prometheus collector for WMI Win32_PerfRawData_NETFramework_NETCLRExceptions metrics
type NETFramework_NETCLRExceptionsCollector struct {
	NumberofExcepsThrown *prometheus.Desc
	NumberofFilters      *prometheus.Desc
	NumberofFinallys     *prometheus.Desc
	ThrowToCatchDepth    *prometheus.Desc
}

// NewNETFramework_NETCLRExceptionsCollector ...
func NewNETFramework_NETCLRExceptionsCollector() (Collector, error) {
	const subsystem = "netframework_clrexceptions"
	return &NETFramework_NETCLRExceptionsCollector{
		NumberofExcepsThrown: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "exceptions_thrown_total"),
			"Displays the total number of exceptions thrown since the application started. This includes both .NET exceptions and unmanaged exceptions that are converted into .NET exceptions.",
			[]string{"process"},
			nil,
		),
		NumberofFilters: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "exceptions_filters_total"),
			"Displays the total number of .NET exception filters executed. An exception filter evaluates regardless of whether an exception is handled.",
			[]string{"process"},
			nil,
		),
		NumberofFinallys: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "exceptions_finallys_total"),
			"Displays the total number of finally blocks executed. Only the finally blocks executed for an exception are counted; finally blocks on normal code paths are not counted by this counter.",
			[]string{"process"},
			nil,
		),
		ThrowToCatchDepth: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "throw_to_catch_depth_total"),
			"Displays the total number of stack frames traversed, from the frame that threw the exception to the frame that handled the exception.",
			[]string{"process"},
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *NETFramework_NETCLRExceptionsCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting win32_perfrawdata_netframework_netclrexceptions metrics:", desc, err)
		return err
	}
	return nil
}

type Win32_PerfRawData_NETFramework_NETCLRExceptions struct {
	Name string

	NumberofExcepsThrown       uint32
	NumberofExcepsThrownPersec uint32
	NumberofFiltersPersec      uint32
	NumberofFinallysPersec     uint32
	ThrowToCatchDepthPersec    uint32
}

func (c *NETFramework_NETCLRExceptionsCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_NETFramework_NETCLRExceptions
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	for _, process := range dst {

		if process.Name == "_Global_" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.NumberofExcepsThrown,
			prometheus.CounterValue,
			float64(process.NumberofExcepsThrown),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberofFilters,
			prometheus.CounterValue,
			float64(process.NumberofFiltersPersec),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberofFinallys,
			prometheus.CounterValue,
			float64(process.NumberofFinallysPersec),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ThrowToCatchDepth,
			prometheus.CounterValue,
			float64(process.ThrowToCatchDepthPersec),
			process.Name,
		)
	}

	return nil, nil
}

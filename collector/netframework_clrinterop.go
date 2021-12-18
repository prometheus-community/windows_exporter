//go:build windows
// +build windows

package collector

import (
	"github.com/StackExchange/wmi"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("netframework_clrinterop", NewNETFramework_NETCLRInteropCollector)
}

// A NETFramework_NETCLRInteropCollector is a Prometheus collector for WMI Win32_PerfRawData_NETFramework_NETCLRInterop metrics
type NETFramework_NETCLRInteropCollector struct {
	NumberofCCWs        *prometheus.Desc
	Numberofmarshalling *prometheus.Desc
	NumberofStubs       *prometheus.Desc
}

// NewNETFramework_NETCLRInteropCollector ...
func NewNETFramework_NETCLRInteropCollector() (Collector, error) {
	const subsystem = "netframework_clrinterop"
	return &NETFramework_NETCLRInteropCollector{
		NumberofCCWs: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "com_callable_wrappers_total"),
			"Displays the current number of COM callable wrappers (CCWs). A CCW is a proxy for a managed object being referenced from an unmanaged COM client.",
			[]string{"process"},
			nil,
		),
		Numberofmarshalling: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "interop_marshalling_total"),
			"Displays the total number of times arguments and return values have been marshaled from managed to unmanaged code, and vice versa, since the application started.",
			[]string{"process"},
			nil,
		),
		NumberofStubs: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "interop_stubs_created_total"),
			"Displays the current number of stubs created by the common language runtime. Stubs are responsible for marshaling arguments and return values from managed to unmanaged code, and vice versa, during a COM interop call or a platform invoke call.",
			[]string{"process"},
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *NETFramework_NETCLRInteropCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting win32_perfrawdata_netframework_netclrinterop metrics:", desc, err)
		return err
	}
	return nil
}

type Win32_PerfRawData_NETFramework_NETCLRInterop struct {
	Name string

	NumberofCCWs             uint32
	Numberofmarshalling      uint32
	NumberofStubs            uint32
	NumberofTLBexportsPersec uint32
	NumberofTLBimportsPersec uint32
}

func (c *NETFramework_NETCLRInteropCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_NETFramework_NETCLRInterop
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	for _, process := range dst {

		if process.Name == "_Global_" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.NumberofCCWs,
			prometheus.CounterValue,
			float64(process.NumberofCCWs),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Numberofmarshalling,
			prometheus.CounterValue,
			float64(process.Numberofmarshalling),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberofStubs,
			prometheus.CounterValue,
			float64(process.NumberofStubs),
			process.Name,
		)
	}

	return nil, nil
}

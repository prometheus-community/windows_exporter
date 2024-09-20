//go:build windows

package netframework

import (
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

func (c *Collector) buildClrInterop() {
	c.numberOfCCWs = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "com_callable_wrappers_total"),
		"Displays the current number of COM callable wrappers (CCWs). A CCW is a proxy for a managed object being referenced from an unmanaged COM client.",
		[]string{"process"},
		nil,
	)
	c.numberOfMarshalling = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "interop_marshalling_total"),
		"Displays the total number of times arguments and return values have been marshaled from managed to unmanaged code, and vice versa, since the application started.",
		[]string{"process"},
		nil,
	)
	c.numberOfStubs = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "interop_stubs_created_total"),
		"Displays the current number of stubs created by the common language runtime. Stubs are responsible for marshaling arguments and return values from managed to unmanaged code, and vice versa, during a COM interop call or a platform invoke call.",
		[]string{"process"},
		nil,
	)
}

type Win32_PerfRawData_NETFramework_NETCLRInterop struct {
	Name string

	NumberofCCWs             uint32
	Numberofmarshalling      uint32
	NumberofStubs            uint32
	NumberofTLBexportsPersec uint32
	NumberofTLBimportsPersec uint32
}

func (c *Collector) collectClrInterop(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_NETFramework_NETCLRInterop
	if err := c.wmiClient.Query("SELECT * FROM Win32_PerfRawData_NETFramework_NETCLRInterop", &dst); err != nil {
		return err
	}

	for _, process := range dst {
		if process.Name == "_Global_" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.numberOfCCWs,
			prometheus.CounterValue,
			float64(process.NumberofCCWs),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberOfMarshalling,
			prometheus.CounterValue,
			float64(process.Numberofmarshalling),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberOfStubs,
			prometheus.CounterValue,
			float64(process.NumberofStubs),
			process.Name,
		)
	}

	return nil
}

//go:build windows

package netframework

import (
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

func (c *Collector) buildClrLoading() {
	c.bytesInLoaderHeap = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "loader_heap_size_bytes"),
		"Displays the current size, in bytes, of the memory committed by the class loader across all application domains. Committed memory is the physical space reserved in the disk paging file.",
		[]string{"process"},
		nil,
	)
	c.currentAppDomains = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "appdomains_loaded_current"),
		"Displays the current number of application domains loaded in this application.",
		[]string{"process"},
		nil,
	)
	c.currentAssemblies = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "assemblies_loaded_current"),
		"Displays the current number of assemblies loaded across all application domains in the currently running application. If the assembly is loaded as domain-neutral from multiple application domains, this counter is incremented only once.",
		[]string{"process"},
		nil,
	)
	c.currentClassesLoaded = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "classes_loaded_current"),
		"Displays the current number of classes loaded in all assemblies.",
		[]string{"process"},
		nil,
	)
	c.totalAppDomains = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "appdomains_loaded_total"),
		"Displays the peak number of application domains loaded since the application started.",
		[]string{"process"},
		nil,
	)
	c.totalAppDomainsUnloaded = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "appdomains_unloaded_total"),
		"Displays the total number of application domains unloaded since the application started. If an application domain is loaded and unloaded multiple times, this counter increments each time the application domain is unloaded.",
		[]string{"process"},
		nil,
	)
	c.totalAssemblies = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "assemblies_loaded_total"),
		"Displays the total number of assemblies loaded since the application started. If the assembly is loaded as domain-neutral from multiple application domains, this counter is incremented only once.",
		[]string{"process"},
		nil,
	)
	c.totalClassesLoaded = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "classes_loaded_total"),
		"Displays the cumulative number of classes loaded in all assemblies since the application started.",
		[]string{"process"},
		nil,
	)
	c.totalNumberOfLoadFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "class_load_failures_total"),
		"Displays the peak number of classes that have failed to load since the application started.",
		[]string{"process"},
		nil,
	)
}

type Win32_PerfRawData_NETFramework_NETCLRLoading struct {
	Name string

	AssemblySearchLength      uint32
	BytesinLoaderHeap         uint64
	Currentappdomains         uint32
	CurrentAssemblies         uint32
	CurrentClassesLoaded      uint32
	PercentTimeLoading        uint64
	Rateofappdomains          uint32
	Rateofappdomainsunloaded  uint32
	RateofAssemblies          uint32
	RateofClassesLoaded       uint32
	RateofLoadFailures        uint32
	TotalAppdomains           uint32
	Totalappdomainsunloaded   uint32
	TotalAssemblies           uint32
	TotalClassesLoaded        uint32
	TotalNumberofLoadFailures uint32
}

func (c *Collector) collectClrLoading(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_NETFramework_NETCLRLoading
	if err := c.wmiClient.Query("SELECT * FROM Win32_PerfRawData_NETFramework_NETCLRLoading", &dst); err != nil {
		return err
	}

	for _, process := range dst {
		if process.Name == "_Global_" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.bytesInLoaderHeap,
			prometheus.GaugeValue,
			float64(process.BytesinLoaderHeap),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.currentAppDomains,
			prometheus.GaugeValue,
			float64(process.Currentappdomains),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.currentAssemblies,
			prometheus.GaugeValue,
			float64(process.CurrentAssemblies),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.currentClassesLoaded,
			prometheus.GaugeValue,
			float64(process.CurrentClassesLoaded),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.totalAppDomains,
			prometheus.CounterValue,
			float64(process.TotalAppdomains),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.totalAppDomainsUnloaded,
			prometheus.CounterValue,
			float64(process.Totalappdomainsunloaded),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.totalAssemblies,
			prometheus.CounterValue,
			float64(process.TotalAssemblies),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.totalClassesLoaded,
			prometheus.CounterValue,
			float64(process.TotalClassesLoaded),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.totalNumberOfLoadFailures,
			prometheus.CounterValue,
			float64(process.TotalNumberofLoadFailures),
			process.Name,
		)
	}

	return nil
}

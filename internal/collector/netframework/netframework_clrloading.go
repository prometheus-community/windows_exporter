//go:build windows

package netframework

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
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
	Name string `mi:"Name"`

	AssemblySearchLength      uint32 `mi:"AssemblySearchLength"`
	BytesinLoaderHeap         uint64 `mi:"BytesinLoaderHeap"`
	Currentappdomains         uint32 `mi:"Currentappdomains"`
	CurrentAssemblies         uint32 `mi:"CurrentAssemblies"`
	CurrentClassesLoaded      uint32 `mi:"CurrentClassesLoaded"`
	PercentTimeLoading        uint64 `mi:"PercentTimeLoading"`
	Rateofappdomains          uint32 `mi:"Rateofappdomains"`
	Rateofappdomainsunloaded  uint32 `mi:"Rateofappdomainsunloaded"`
	RateofAssemblies          uint32 `mi:"RateofAssemblies"`
	RateofClassesLoaded       uint32 `mi:"RateofClassesLoaded"`
	RateofLoadFailures        uint32 `mi:"RateofLoadFailures"`
	TotalAppdomains           uint32 `mi:"TotalAppdomains"`
	Totalappdomainsunloaded   uint32 `mi:"Totalappdomainsunloaded"`
	TotalAssemblies           uint32 `mi:"TotalAssemblies"`
	TotalClassesLoaded        uint32 `mi:"TotalClassesLoaded"`
	TotalNumberofLoadFailures uint32 `mi:"TotalNumberofLoadFailures"`
}

func (c *Collector) collectClrLoading(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_NETFramework_NETCLRLoading
	if err := c.miSession.Query(&dst, mi.NamespaceRootCIMv2, utils.Must(mi.NewQuery("SELECT * Win32_PerfRawData_NETFramework_NETCLRLoading"))); err != nil {
		return fmt.Errorf("WMI query failed: %w", err)
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

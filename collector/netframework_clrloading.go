//go:build windows
// +build windows

package collector

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
)

// A NETFramework_NETCLRLoadingCollector is a Prometheus collector for WMI Win32_PerfRawData_NETFramework_NETCLRLoading metrics
type NETFramework_NETCLRLoadingCollector struct {
	logger log.Logger

	BytesinLoaderHeap         *prometheus.Desc
	Currentappdomains         *prometheus.Desc
	CurrentAssemblies         *prometheus.Desc
	CurrentClassesLoaded      *prometheus.Desc
	TotalAppdomains           *prometheus.Desc
	Totalappdomainsunloaded   *prometheus.Desc
	TotalAssemblies           *prometheus.Desc
	TotalClassesLoaded        *prometheus.Desc
	TotalNumberofLoadFailures *prometheus.Desc
}

// newNETFramework_NETCLRLoadingCollector ...
func newNETFramework_NETCLRLoadingCollector(logger log.Logger) (Collector, error) {
	const subsystem = "netframework_clrloading"
	return &NETFramework_NETCLRLoadingCollector{
		logger: log.With(logger, "collector", subsystem),
		BytesinLoaderHeap: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "loader_heap_size_bytes"),
			"Displays the current size, in bytes, of the memory committed by the class loader across all application domains. Committed memory is the physical space reserved in the disk paging file.",
			[]string{"process"},
			nil,
		),
		Currentappdomains: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "appdomains_loaded_current"),
			"Displays the current number of application domains loaded in this application.",
			[]string{"process"},
			nil,
		),
		CurrentAssemblies: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "assemblies_loaded_current"),
			"Displays the current number of assemblies loaded across all application domains in the currently running application. If the assembly is loaded as domain-neutral from multiple application domains, this counter is incremented only once.",
			[]string{"process"},
			nil,
		),
		CurrentClassesLoaded: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "classes_loaded_current"),
			"Displays the current number of classes loaded in all assemblies.",
			[]string{"process"},
			nil,
		),
		TotalAppdomains: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "appdomains_loaded_total"),
			"Displays the peak number of application domains loaded since the application started.",
			[]string{"process"},
			nil,
		),
		Totalappdomainsunloaded: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "appdomains_unloaded_total"),
			"Displays the total number of application domains unloaded since the application started. If an application domain is loaded and unloaded multiple times, this counter increments each time the application domain is unloaded.",
			[]string{"process"},
			nil,
		),
		TotalAssemblies: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "assemblies_loaded_total"),
			"Displays the total number of assemblies loaded since the application started. If the assembly is loaded as domain-neutral from multiple application domains, this counter is incremented only once.",
			[]string{"process"},
			nil,
		),
		TotalClassesLoaded: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "classes_loaded_total"),
			"Displays the cumulative number of classes loaded in all assemblies since the application started.",
			[]string{"process"},
			nil,
		),
		TotalNumberofLoadFailures: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "class_load_failures_total"),
			"Displays the peak number of classes that have failed to load since the application started.",
			[]string{"process"},
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *NETFramework_NETCLRLoadingCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		_ = level.Error(c.logger).Log("failed collecting win32_perfrawdata_netframework_netclrloading metrics", "desc", desc, "err", err)
		return err
	}
	return nil
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

func (c *NETFramework_NETCLRLoadingCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_NETFramework_NETCLRLoading
	q := queryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	for _, process := range dst {

		if process.Name == "_Global_" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.BytesinLoaderHeap,
			prometheus.GaugeValue,
			float64(process.BytesinLoaderHeap),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Currentappdomains,
			prometheus.GaugeValue,
			float64(process.Currentappdomains),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.CurrentAssemblies,
			prometheus.GaugeValue,
			float64(process.CurrentAssemblies),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.CurrentClassesLoaded,
			prometheus.GaugeValue,
			float64(process.CurrentClassesLoaded),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalAppdomains,
			prometheus.CounterValue,
			float64(process.TotalAppdomains),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Totalappdomainsunloaded,
			prometheus.CounterValue,
			float64(process.Totalappdomainsunloaded),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalAssemblies,
			prometheus.CounterValue,
			float64(process.TotalAssemblies),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalClassesLoaded,
			prometheus.CounterValue,
			float64(process.TotalClassesLoaded),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalNumberofLoadFailures,
			prometheus.CounterValue,
			float64(process.TotalNumberofLoadFailures),
			process.Name,
		)
	}

	return nil, nil
}

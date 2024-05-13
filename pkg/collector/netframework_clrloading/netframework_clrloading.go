//go:build windows

package netframework_clrloading

import (
	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "netframework_clrloading"

type Config struct{}

var ConfigDefaults = Config{}

// A collector is a Prometheus collector for WMI Win32_PerfRawData_NETFramework_NETCLRLoading metrics
type collector struct {
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

func New(logger log.Logger, _ *Config) types.Collector {
	c := &collector{}
	c.SetLogger(logger)
	return c
}

func NewWithFlags(_ *kingpin.Application) types.Collector {
	return &collector{}
}

func (c *collector) GetName() string {
	return Name
}

func (c *collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *collector) GetPerfCounter() ([]string, error) {
	return []string{}, nil
}

func (c *collector) Build() error {
	c.BytesinLoaderHeap = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "loader_heap_size_bytes"),
		"Displays the current size, in bytes, of the memory committed by the class loader across all application domains. Committed memory is the physical space reserved in the disk paging file.",
		[]string{"process"},
		nil,
	)
	c.Currentappdomains = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "appdomains_loaded_current"),
		"Displays the current number of application domains loaded in this application.",
		[]string{"process"},
		nil,
	)
	c.CurrentAssemblies = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "assemblies_loaded_current"),
		"Displays the current number of assemblies loaded across all application domains in the currently running application. If the assembly is loaded as domain-neutral from multiple application domains, this counter is incremented only once.",
		[]string{"process"},
		nil,
	)
	c.CurrentClassesLoaded = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "classes_loaded_current"),
		"Displays the current number of classes loaded in all assemblies.",
		[]string{"process"},
		nil,
	)
	c.TotalAppdomains = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "appdomains_loaded_total"),
		"Displays the peak number of application domains loaded since the application started.",
		[]string{"process"},
		nil,
	)
	c.Totalappdomainsunloaded = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "appdomains_unloaded_total"),
		"Displays the total number of application domains unloaded since the application started. If an application domain is loaded and unloaded multiple times, this counter increments each time the application domain is unloaded.",
		[]string{"process"},
		nil,
	)
	c.TotalAssemblies = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "assemblies_loaded_total"),
		"Displays the total number of assemblies loaded since the application started. If the assembly is loaded as domain-neutral from multiple application domains, this counter is incremented only once.",
		[]string{"process"},
		nil,
	)
	c.TotalClassesLoaded = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "classes_loaded_total"),
		"Displays the cumulative number of classes loaded in all assemblies since the application started.",
		[]string{"process"},
		nil,
	)
	c.TotalNumberofLoadFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "class_load_failures_total"),
		"Displays the peak number of classes that have failed to load since the application started.",
		[]string{"process"},
		nil,
	)
	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.collect(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting win32_perfrawdata_netframework_netclrloading metrics", "err", err)
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

func (c *collector) collect(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_NETFramework_NETCLRLoading
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
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

	return nil
}

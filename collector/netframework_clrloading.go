// +build windows

package collector

import (
	"fmt"
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	registerCollector("netframework_clrloading", NewNETFrameworkCLRLoadingCollector, ".NET CLR Loading")
}

// A NETFrameworkCLRLoadingCollector is a Prometheus collector for Perflib .NET CLR Loading metrics
type NETFrameworkCLRLoadingCollector struct {
	BytesinLoaderHeap         *prometheus.Desc
	Currentappdomains         *prometheus.Desc
	CurrentAssemblies         *prometheus.Desc
	CurrentClassesLoaded      *prometheus.Desc
	TotalAppdomains           *prometheus.Desc
	Totalappdomainsunloaded   *prometheus.Desc
	TotalAssemblies           *prometheus.Desc
	TotalClassesLoaded        *prometheus.Desc
	TotalNumberofLoadFailures *prometheus.Desc

	processWhitelistPattern *regexp.Regexp
	processBlacklistPattern *regexp.Regexp
}

// NewNETFrameworkCLRLoadingCollector ...
func NewNETFrameworkCLRLoadingCollector() (Collector, error) {
	const subsystem = "netframework_clrloading"
	commonFlags := GetNETFrameworkFlags()
	return &NETFrameworkCLRLoadingCollector{
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
		processWhitelistPattern: commonFlags.whitelistRegexp,
		processBlacklistPattern: commonFlags.blacklistRegexp,
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *NETFrameworkCLRLoadingCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ctx, ch); err != nil {
		log.Error("failed collecting netframework_clrloading metrics:", desc, err)
		return err
	}
	return nil
}

type netframeworkCLRLoading struct {
	Name string

	AssemblySearchLength      float64 `perflib:"Assembly Search Length"`
	BytesinLoaderHeap         float64 `perflib:"Bytes in Loader Heap"`
	Currentappdomains         float64 `perflib:"Current appdomains"`
	CurrentAssemblies         float64 `perflib:"Current Assemblies"`
	CurrentClassesLoaded      float64 `perflib:"Current Classes Loaded"`
	PercentTimeLoading        float64 `perflib:"% Time Loading"`
	Rateofappdomains          float64 `perflib:"Rate of appdomains"`
	Rateofappdomainsunloaded  float64 `perflib:"Rate of appdomains unloaded"`
	RateofAssemblies          float64 `perflib:"Rate of Assemblies"`
	RateofClassesLoaded       float64 `perflib:"Rate of Classes Loaded"`
	RateofLoadFailures        float64 `perflib:"Rate of Load Failures"`
	TotalAppdomains           float64 `perflib:"Total Appdomains"`
	Totalappdomainsunloaded   float64 `perflib:"Total appdomains unloaded"`
	TotalAssemblies           float64 `perflib:"Total Assemblies"`
	TotalClassesLoaded        float64 `perflib:"Total Classes Loaded"`
	TotalNumberofLoadFailures float64 `perflib:"Total # of Load Failures"`
}

func (c *NETFrameworkCLRLoadingCollector) collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []netframeworkCLRLoading

	if err := unmarshalObject(ctx.perfObjects[".NET CLR Loading"], &dst); err != nil {
		return nil, err
	}

	var names = make(map[string]int, len(dst))
	for _, process := range dst {

		if process.Name == "_Global_" {
			continue
		}

		// Append "#1", "#2", etc., to process names to disambiguate duplicates.
		name := process.Name
		procnum, exists := names[name]
		if exists {
			name = fmt.Sprintf("%s#%d", name, procnum)
			names[name]++
		} else {
			names[name] = 1
		}

		// The pattern matching against the whitelist and blacklist has to occur
		// after appending #N above to be consistent with other collectors.
		if c.processBlacklistPattern.MatchString(name) ||
			!c.processWhitelistPattern.MatchString(name) {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.BytesinLoaderHeap,
			prometheus.GaugeValue,
			process.BytesinLoaderHeap,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Currentappdomains,
			prometheus.GaugeValue,
			process.Currentappdomains,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.CurrentAssemblies,
			prometheus.GaugeValue,
			process.CurrentAssemblies,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.CurrentClassesLoaded,
			prometheus.GaugeValue,
			process.CurrentClassesLoaded,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalAppdomains,
			prometheus.CounterValue,
			process.TotalAppdomains,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Totalappdomainsunloaded,
			prometheus.CounterValue,
			process.Totalappdomainsunloaded,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalAssemblies,
			prometheus.CounterValue,
			process.TotalAssemblies,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalClassesLoaded,
			prometheus.CounterValue,
			process.TotalClassesLoaded,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalNumberofLoadFailures,
			prometheus.CounterValue,
			process.TotalNumberofLoadFailures,
			name,
		)
	}

	return nil, nil
}

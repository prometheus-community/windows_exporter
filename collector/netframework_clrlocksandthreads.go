// +build windows

package collector

import (
	"fmt"
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	registerCollector("netframework_clrlocksandthreads", NewNETFramework_NETCLRLocksAndThreadsCollector, ".NET CLR LocksAndThreads")
}

// A NETFrameworkCLRLocksAndThreadsCollector is a Prometheus collector for Perflib .NET CLR LocksAndThreads metrics
type NETFrameworkCLRLocksAndThreadsCollector struct {
	CurrentQueueLength               *prometheus.Desc
	NumberofcurrentlogicalThreads    *prometheus.Desc
	NumberofcurrentphysicalThreads   *prometheus.Desc
	Numberofcurrentrecognizedthreads *prometheus.Desc
	Numberoftotalrecognizedthreads   *prometheus.Desc
	QueueLengthPeak                  *prometheus.Desc
	TotalNumberofContentions         *prometheus.Desc

	processWhitelistPattern *regexp.Regexp
	processBlacklistPattern *regexp.Regexp
}

// NewNETFramework_NETCLRLocksAndThreadsCollector ...
func NewNETFramework_NETCLRLocksAndThreadsCollector() (Collector, error) {
	const subsystem = "netframework_clrlocksandthreads"
	commonFlags := GetNETFrameworkFlags()
	return &NETFrameworkCLRLocksAndThreadsCollector{
		CurrentQueueLength: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "current_queue_length"),
			"Displays the total number of threads that are currently waiting to acquire a managed lock in the application.",
			[]string{"process"},
			nil,
		),
		NumberofcurrentlogicalThreads: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "current_logical_threads"),
			"Displays the number of current managed thread objects in the application. This counter maintains the count of both running and stopped threads. ",
			[]string{"process"},
			nil,
		),
		NumberofcurrentphysicalThreads: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "physical_threads_current"),
			"Displays the number of native operating system threads created and owned by the common language runtime to act as underlying threads for managed thread objects. This counter's value does not include the threads used by the runtime in its internal operations; it is a subset of the threads in the operating system process.",
			[]string{"process"},
			nil,
		),
		Numberofcurrentrecognizedthreads: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "recognized_threads_current"),
			"Displays the number of threads that are currently recognized by the runtime. These threads are associated with a corresponding managed thread object. The runtime does not create these threads, but they have run inside the runtime at least once.",
			[]string{"process"},
			nil,
		),
		Numberoftotalrecognizedthreads: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "recognized_threads_total"),
			"Displays the total number of threads that have been recognized by the runtime since the application started. These threads are associated with a corresponding managed thread object. The runtime does not create these threads, but they have run inside the runtime at least once.",
			[]string{"process"},
			nil,
		),
		QueueLengthPeak: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "queue_length_total"),
			"Displays the total number of threads that waited to acquire a managed lock since the application started.",
			[]string{"process"},
			nil,
		),
		TotalNumberofContentions: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "contentions_total"),
			"Displays the total number of times that threads in the runtime have attempted to acquire a managed lock unsuccessfully.",
			[]string{"process"},
			nil,
		),
		processWhitelistPattern: commonFlags.whitelistRegexp,
		processBlacklistPattern: commonFlags.blacklistRegexp,
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *NETFrameworkCLRLocksAndThreadsCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ctx, ch); err != nil {
		log.Error("failed collecting netframework_clrlocksandthreads metrics:", desc, err)
		return err
	}
	return nil
}

type netframeworkCLRLocksAndThreads struct {
	Name string

	ContentionRatePersec             float64 `perflib:"Contention Rate / sec"`
	CurrentQueueLength               float64 `perflib:"Current Queue Length"`
	NumberofcurrentlogicalThreads    float64 `perflib:"# of current logical Threads"`
	NumberofcurrentphysicalThreads   float64 `perflib:"# of current physical Threads"`
	Numberofcurrentrecognizedthreads float64 `perflib:"# of current recognized threads"`
	Numberoftotalrecognizedthreads   float64 `perflib:"# of total recognized threads"`
	QueueLengthPeak                  float64 `perflib:"Queue Length Peak"`
	QueueLengthPersec                float64 `perflib:"Queue Length / sec"`
	RateOfRecognizedThreadsPersec    float64 `perflib:"rate of recognized threads / sec"`
	TotalNumberofContentions         float64 `perflib:"Total # of Contentions"`
}

func (c *NETFrameworkCLRLocksAndThreadsCollector) collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []netframeworkCLRLocksAndThreads

	if err := unmarshalObject(ctx.perfObjects[".NET CLR LocksAndThreads"], &dst); err != nil {
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
			c.CurrentQueueLength,
			prometheus.GaugeValue,
			process.CurrentQueueLength,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberofcurrentlogicalThreads,
			prometheus.GaugeValue,
			process.NumberofcurrentlogicalThreads,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberofcurrentphysicalThreads,
			prometheus.GaugeValue,
			process.NumberofcurrentphysicalThreads,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Numberofcurrentrecognizedthreads,
			prometheus.GaugeValue,
			process.Numberofcurrentrecognizedthreads,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Numberoftotalrecognizedthreads,
			prometheus.CounterValue,
			process.Numberoftotalrecognizedthreads,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.QueueLengthPeak,
			prometheus.CounterValue,
			process.QueueLengthPeak,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalNumberofContentions,
			prometheus.CounterValue,
			process.TotalNumberofContentions,
			name,
		)
	}

	return nil, nil
}

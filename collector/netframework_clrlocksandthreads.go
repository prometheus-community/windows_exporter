//go:build windows
// +build windows

package collector

import (
	"github.com/StackExchange/wmi"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("netframework_clrlocksandthreads", NewNETFramework_NETCLRLocksAndThreadsCollector)
}

// A NETFramework_NETCLRLocksAndThreadsCollector is a Prometheus collector for WMI Win32_PerfRawData_NETFramework_NETCLRLocksAndThreads metrics
type NETFramework_NETCLRLocksAndThreadsCollector struct {
	CurrentQueueLength               *prometheus.Desc
	NumberofcurrentlogicalThreads    *prometheus.Desc
	NumberofcurrentphysicalThreads   *prometheus.Desc
	Numberofcurrentrecognizedthreads *prometheus.Desc
	Numberoftotalrecognizedthreads   *prometheus.Desc
	QueueLengthPeak                  *prometheus.Desc
	TotalNumberofContentions         *prometheus.Desc
}

// NewNETFramework_NETCLRLocksAndThreadsCollector ...
func NewNETFramework_NETCLRLocksAndThreadsCollector() (Collector, error) {
	const subsystem = "netframework_clrlocksandthreads"
	return &NETFramework_NETCLRLocksAndThreadsCollector{
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
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *NETFramework_NETCLRLocksAndThreadsCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting win32_perfrawdata_netframework_netclrlocksandthreads metrics:", desc, err)
		return err
	}
	return nil
}

type Win32_PerfRawData_NETFramework_NETCLRLocksAndThreads struct {
	Name string

	ContentionRatePersec             uint32
	CurrentQueueLength               uint32
	NumberofcurrentlogicalThreads    uint32
	NumberofcurrentphysicalThreads   uint32
	Numberofcurrentrecognizedthreads uint32
	Numberoftotalrecognizedthreads   uint32
	QueueLengthPeak                  uint32
	QueueLengthPersec                uint32
	RateOfRecognizedThreadsPersec    uint32
	TotalNumberofContentions         uint32
}

func (c *NETFramework_NETCLRLocksAndThreadsCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_NETFramework_NETCLRLocksAndThreads
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	for _, process := range dst {

		if process.Name == "_Global_" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.CurrentQueueLength,
			prometheus.GaugeValue,
			float64(process.CurrentQueueLength),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberofcurrentlogicalThreads,
			prometheus.GaugeValue,
			float64(process.NumberofcurrentlogicalThreads),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberofcurrentphysicalThreads,
			prometheus.GaugeValue,
			float64(process.NumberofcurrentphysicalThreads),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Numberofcurrentrecognizedthreads,
			prometheus.GaugeValue,
			float64(process.Numberofcurrentrecognizedthreads),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Numberoftotalrecognizedthreads,
			prometheus.CounterValue,
			float64(process.Numberoftotalrecognizedthreads),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.QueueLengthPeak,
			prometheus.CounterValue,
			float64(process.QueueLengthPeak),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalNumberofContentions,
			prometheus.CounterValue,
			float64(process.TotalNumberofContentions),
			process.Name,
		)
	}

	return nil, nil
}

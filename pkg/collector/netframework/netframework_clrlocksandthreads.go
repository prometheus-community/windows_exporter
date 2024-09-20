//go:build windows

package netframework

import (
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

func (c *Collector) buildClrLocksAndThreads() {
	c.currentQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_queue_length"),
		"Displays the total number of threads that are currently waiting to acquire a managed lock in the application.",
		[]string{"process"},
		nil,
	)
	c.numberOfCurrentLogicalThreads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_logical_threads"),
		"Displays the number of current managed thread objects in the application. This counter maintains the count of both running and stopped threads. ",
		[]string{"process"},
		nil,
	)
	c.numberOfCurrentPhysicalThreads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "physical_threads_current"),
		"Displays the number of native operating system threads created and owned by the common language runtime to act as underlying threads for managed thread objects. This counter's value does not include the threads used by the runtime in its internal operations; it is a subset of the threads in the operating system process.",
		[]string{"process"},
		nil,
	)
	c.numberOfCurrentRecognizedThreads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "recognized_threads_current"),
		"Displays the number of threads that are currently recognized by the runtime. These threads are associated with a corresponding managed thread object. The runtime does not create these threads, but they have run inside the runtime at least once.",
		[]string{"process"},
		nil,
	)
	c.numberOfTotalRecognizedThreads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "recognized_threads_total"),
		"Displays the total number of threads that have been recognized by the runtime since the application started. These threads are associated with a corresponding managed thread object. The runtime does not create these threads, but they have run inside the runtime at least once.",
		[]string{"process"},
		nil,
	)
	c.queueLengthPeak = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "queue_length_total"),
		"Displays the total number of threads that waited to acquire a managed lock since the application started.",
		[]string{"process"},
		nil,
	)
	c.totalNumberOfContentions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "contentions_total"),
		"Displays the total number of times that threads in the runtime have attempted to acquire a managed lock unsuccessfully.",
		[]string{"process"},
		nil,
	)
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

func (c *Collector) collectClrLocksAndThreads(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_NETFramework_NETCLRLocksAndThreads
	if err := c.wmiClient.Query("SELECT * FROM Win32_PerfRawData_NETFramework_NETCLRLocksAndThreads", &dst); err != nil {
		return err
	}

	for _, process := range dst {
		if process.Name == "_Global_" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.currentQueueLength,
			prometheus.GaugeValue,
			float64(process.CurrentQueueLength),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberOfCurrentLogicalThreads,
			prometheus.GaugeValue,
			float64(process.NumberofcurrentlogicalThreads),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberOfCurrentPhysicalThreads,
			prometheus.GaugeValue,
			float64(process.NumberofcurrentphysicalThreads),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberOfCurrentRecognizedThreads,
			prometheus.GaugeValue,
			float64(process.Numberofcurrentrecognizedthreads),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberOfTotalRecognizedThreads,
			prometheus.CounterValue,
			float64(process.Numberoftotalrecognizedthreads),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.queueLengthPeak,
			prometheus.CounterValue,
			float64(process.QueueLengthPeak),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.totalNumberOfContentions,
			prometheus.CounterValue,
			float64(process.TotalNumberofContentions),
			process.Name,
		)
	}

	return nil
}

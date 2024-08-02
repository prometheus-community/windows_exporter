//go:build windows

package netframework_clrlocksandthreads

import (
	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "netframework_clrlocksandthreads"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_NETFramework_NETCLRLocksAndThreads metrics
type Collector struct {
	logger log.Logger

	CurrentQueueLength               *prometheus.Desc
	NumberofcurrentlogicalThreads    *prometheus.Desc
	NumberofcurrentphysicalThreads   *prometheus.Desc
	Numberofcurrentrecognizedthreads *prometheus.Desc
	Numberoftotalrecognizedthreads   *prometheus.Desc
	QueueLengthPeak                  *prometheus.Desc
	TotalNumberofContentions         *prometheus.Desc
}

func New(logger log.Logger, _ *Config) *Collector {
	c := &Collector{}
	c.SetLogger(logger)

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *Collector) GetPerfCounter() ([]string, error) {
	return []string{}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build() error {
	c.CurrentQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_queue_length"),
		"Displays the total number of threads that are currently waiting to acquire a managed lock in the application.",
		[]string{"process"},
		nil,
	)
	c.NumberofcurrentlogicalThreads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_logical_threads"),
		"Displays the number of current managed thread objects in the application. This counter maintains the count of both running and stopped threads. ",
		[]string{"process"},
		nil,
	)
	c.NumberofcurrentphysicalThreads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "physical_threads_current"),
		"Displays the number of native operating system threads created and owned by the common language runtime to act as underlying threads for managed thread objects. This counter's value does not include the threads used by the runtime in its internal operations; it is a subset of the threads in the operating system process.",
		[]string{"process"},
		nil,
	)
	c.Numberofcurrentrecognizedthreads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "recognized_threads_current"),
		"Displays the number of threads that are currently recognized by the runtime. These threads are associated with a corresponding managed thread object. The runtime does not create these threads, but they have run inside the runtime at least once.",
		[]string{"process"},
		nil,
	)
	c.Numberoftotalrecognizedthreads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "recognized_threads_total"),
		"Displays the total number of threads that have been recognized by the runtime since the application started. These threads are associated with a corresponding managed thread object. The runtime does not create these threads, but they have run inside the runtime at least once.",
		[]string{"process"},
		nil,
	)
	c.QueueLengthPeak = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "queue_length_total"),
		"Displays the total number of threads that waited to acquire a managed lock since the application started.",
		[]string{"process"},
		nil,
	)
	c.TotalNumberofContentions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "contentions_total"),
		"Displays the total number of times that threads in the runtime have attempted to acquire a managed lock unsuccessfully.",
		[]string{"process"},
		nil,
	)
	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.collect(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting win32_perfrawdata_netframework_netclrlocksandthreads metrics", "err", err)
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

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_NETFramework_NETCLRLocksAndThreads
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
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

	return nil
}

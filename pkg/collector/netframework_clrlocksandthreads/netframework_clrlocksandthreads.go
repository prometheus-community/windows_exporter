//go:build windows

package netframework_clrlocksandthreads

import (
	"errors"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
)

const Name = "netframework_clrlocksandthreads"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_NETFramework_NETCLRLocksAndThreads metrics.
type Collector struct {
	config    Config
	wmiClient *wmi.Client

	currentQueueLength               *prometheus.Desc
	numberOfCurrentLogicalThreads    *prometheus.Desc
	numberOfCurrentPhysicalThreads   *prometheus.Desc
	numberOfCurrentRecognizedThreads *prometheus.Desc
	numberOfTotalRecognizedThreads   *prometheus.Desc
	queueLengthPeak                  *prometheus.Desc
	totalNumberOfContentions         *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) GetPerfCounter(_ *slog.Logger) ([]string, error) {
	return []string{}, nil
}

func (c *Collector) Close(_ *slog.Logger) error {
	return nil
}

func (c *Collector) Build(_ *slog.Logger, wmiClient *wmi.Client) error {
	if wmiClient == nil || wmiClient.SWbemServicesClient == nil {
		return errors.New("wmiClient or SWbemServicesClient is nil")
	}

	c.wmiClient = wmiClient

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

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	logger = logger.With(slog.String("collector", Name))
	if err := c.collect(ch); err != nil {
		logger.Error("failed collecting win32_perfrawdata_netframework_netclrlocksandthreads metrics",
			slog.Any("err", err),
		)

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

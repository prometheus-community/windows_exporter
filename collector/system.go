// returns data points from Win32_PerfRawData_PerfOS_System class
// https://web.archive.org/web/20050830140516/http://msdn.microsoft.com/library/en-us/wmisdk/wmi/win32_perfrawdata_perfos_system.asp

package collector

import (
	"errors"
	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	Factories["system"] = NewSystemCollector
}

// A SystemCollector is a Prometheus collector for WMI metrics
type SystemCollector struct {
	ContextSwitchesTotal     *prometheus.Desc
	ExceptionDispatchesTotal *prometheus.Desc
	ProcessorQueueLength     *prometheus.Desc
	SystemCallsTotal         *prometheus.Desc
	SystemUpTime             *prometheus.Desc
	Threads                  *prometheus.Desc
}

// NewSystemCollector ...
func NewSystemCollector() (Collector, error) {
	const subsystem = "system"

	return &SystemCollector{
		ContextSwitchesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "context_switches_total"),
			"Total number of context switches (WMI source: PerfOS_System.ContextSwitchesPersec)",
			nil,
			nil,
		),
		ExceptionDispatchesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "exception_dispatches_total"),
			"Total number of exceptions dispatched (WMI source: PerfOS_System.ExceptionDispatchesPersec)",
			nil,
			nil,
		),
		ProcessorQueueLength: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "processor_queue_length"),
			"Length of processor queue (WMI source: PerfOS_System.ProcessorQueueLength)",
			nil,
			nil,
		),
		SystemCallsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "system_calls_total"),
			"Total number of system calls (WMI source: PerfOS_System.SystemCallsPersec)",
			nil,
			nil,
		),
		SystemUpTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "system_up_time"),
			"System boot time (WMI source: PerfOS_System.SystemUpTime)",
			nil,
			nil,
		),
		Threads: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "threads"),
			"Current number of threads (WMI source: PerfOS_System.Threads)",
			nil,
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *SystemCollector) Collect(ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting system metrics:", desc, err)
		return err
	}
	return nil
}

type Win32_PerfRawData_PerfOS_System struct {
	ContextSwitchesPersec     uint32
	ExceptionDispatchesPersec uint32
	Frequency_Object          uint64
	ProcessorQueueLength      uint32
	SystemCallsPersec         uint32
	SystemUpTime              uint64
	Threads                   uint32
	Timestamp_Object          uint64
}

func (c *SystemCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_PerfOS_System
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}
	if len(dst) == 0 {
		return nil, errors.New("WMI query returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.ContextSwitchesTotal,
		prometheus.CounterValue,
		float64(dst[0].ContextSwitchesPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.ExceptionDispatchesTotal,
		prometheus.CounterValue,
		float64(dst[0].ExceptionDispatchesPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.ProcessorQueueLength,
		prometheus.GaugeValue,
		float64(dst[0].ProcessorQueueLength),
	)
	ch <- prometheus.MustNewConstMetric(
		c.SystemCallsTotal,
		prometheus.CounterValue,
		float64(dst[0].SystemCallsPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.SystemUpTime,
		prometheus.GaugeValue,
		// convert from Windows timestamp (1 jan 1601) to unix timestamp (1 jan 1970)
		float64(dst[0].SystemUpTime-116444736000000000)/float64(dst[0].Frequency_Object),
	)
	ch <- prometheus.MustNewConstMetric(
		c.Threads,
		prometheus.GaugeValue,
		float64(dst[0].Threads),
	)
	return nil, nil
}

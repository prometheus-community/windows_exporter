// returns data points from Win32_PerfRawData_Tcpip_TCPv4

// https://msdn.microsoft.com/en-us/library/aa394341(v=vs.85).aspx (Win32_PerfRawData_Tcpip_TCPv4 class)

package collector

import (
	"log"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	Factories["tcp"] = NewTCPCollector
}

// A TCPCollector is a Prometheus collector for WMI Win32_PerfRawData_Tcpip_TCPv4 metrics
type TCPCollector struct {
	ConnectionFailures          *prometheus.Desc
	ConnectionsActive           *prometheus.Desc
	ConnectionsEstablished      *prometheus.Desc
	ConnectionsPassive          *prometheus.Desc
	ConnectionsReset            *prometheus.Desc
	FrequencyObject             *prometheus.Desc
	FrequencyPerfTime           *prometheus.Desc
	FrequencySys100NS           *prometheus.Desc
	SegmentsPerSec              *prometheus.Desc
	SegmentsReceivedPerSec      *prometheus.Desc
	SegmentsRetransmittedPerSec *prometheus.Desc
	SegmentsSentPerSec          *prometheus.Desc
	TimestampObject             *prometheus.Desc
	TimestampPerfTime           *prometheus.Desc
	TimestampSys100NS           *prometheus.Desc
}

// NewTCPCollector ...
func NewTCPCollector() (Collector, error) {
	const subsystem = "tcp"

	return &TCPCollector{
		ConnectionFailures: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "connection_failures"),
			"(TCP.ConnectionFailures)",
			nil,
			nil,
		),
		ConnectionsActive: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "connections_active"),
			"(TCP.ConnectionsActive)",
			nil,
			nil,
		),
		ConnectionsEstablished: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "connections_established"),
			"(TCP.ConnectionsEstablished)",
			nil,
			nil,
		),
		ConnectionsPassive: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "connections_passive"),
			"(TCP.ConnectionsPassive)",
			nil,
			nil,
		),
		ConnectionsReset: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "connections_reset"),
			"(TCP.ConnectionsReset)",
			nil,
			nil,
		),
		FrequencyObject: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "frequency_object"),
			"(TCP.FrequencyObject)",
			nil,
			nil,
		),
		FrequencyPerfTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "frequency_perftime"),
			"(TCP.FrequencyPerfTime)",
			nil,
			nil,
		),
		FrequencySys100NS: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "frequency_sys100ns"),
			"(TCP.FrequencySys100NS)",
			nil,
			nil,
		),
		SegmentsPerSec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "segments_per_sec"),
			"(TCP.SegmentsPerSec)",
			nil,
			nil,
		),
		SegmentsReceivedPerSec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "segments_received_per_sec"),
			"(TCP.SegmentsReceivedPerSec)",
			nil,
			nil,
		),
		SegmentsRetransmittedPerSec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "segments_retransmitted_per_sec"),
			"(TCP.SegmentsRetransmittedPerSec)",
			nil,
			nil,
		),
		SegmentsSentPerSec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "segments_sent_per_sec"),
			"(TCP.SegmentsSentPerSec)",
			nil,
			nil,
		),
		TimestampObject: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "timestamp_object"),
			"(TCP.TimestampObject)",
			nil,
			nil,
		),
		TimestampPerfTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "timestamp_perftime"),
			"(TCP.TimestampPerfTime)",
			nil,
			nil,
		),
		TimestampSys100NS: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "frequency_sys100ns"),
			"(TCP.TimestampSys100NS)",
			nil,
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *TCPCollector) Collect(ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Println("[ERROR] failed collecting tcp metrics:", desc, err)
		return err
	}
	return nil
}

type Win32_PerfRawData_Tcpip_TCPv4 struct {
	Caption                     string
	ConnectionFailures          uint64
	ConnectionsActive           uint64
	ConnectionsEstablished      uint64
	ConnectionsPassive          uint64
	ConnectionsReset            uint64
	Description                 string
	Frequency_Object            uint64
	Frequency_PerfTime          uint64
	Frequency_Sys100NS          uint64
	Name                        string
	SegmentsPersec              uint64
	SegmentsReceivedPersec      uint64
	SegmentsRetransmittedPersec uint64
	SegmentsSentPersec          uint64
	Timestamp_Object            uint64
	Timestamp_PerfTime          uint64
	Timestamp_Sys100NS          uint64
}

func (c *TCPCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_Tcpip_TCPv4

	q := wmi.CreateQuery(&dst, "")
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	// Counters
	ch <- prometheus.MustNewConstMetric(
		c.ConnectionFailures,
		prometheus.CounterValue,
		float64(dst[0].ConnectionFailures),
	)
	ch <- prometheus.MustNewConstMetric(
		c.ConnectionsActive,
		prometheus.CounterValue,
		float64(dst[0].ConnectionsActive),
	)
	ch <- prometheus.MustNewConstMetric(
		c.ConnectionsEstablished,
		prometheus.CounterValue,
		float64(dst[0].ConnectionsEstablished),
	)
	ch <- prometheus.MustNewConstMetric(
		c.ConnectionsPassive,
		prometheus.CounterValue,
		float64(dst[0].ConnectionsPassive),
	)
	ch <- prometheus.MustNewConstMetric(
		c.ConnectionsReset,
		prometheus.CounterValue,
		float64(dst[0].ConnectionsReset),
	)
	ch <- prometheus.MustNewConstMetric(
		c.FrequencyObject,
		prometheus.CounterValue,
		float64(dst[0].Frequency_Object),
	)
	ch <- prometheus.MustNewConstMetric(
		c.FrequencyPerfTime,
		prometheus.CounterValue,
		float64(dst[0].Frequency_PerfTime),
	)
	ch <- prometheus.MustNewConstMetric(
		c.FrequencySys100NS,
		prometheus.CounterValue,
		float64(dst[0].Frequency_Sys100NS),
	)
	ch <- prometheus.MustNewConstMetric(
		c.SegmentsPerSec,
		prometheus.CounterValue,
		float64(dst[0].SegmentsPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.SegmentsReceivedPerSec,
		prometheus.CounterValue,
		float64(dst[0].SegmentsReceivedPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.SegmentsRetransmittedPerSec,
		prometheus.CounterValue,
		float64(dst[0].SegmentsRetransmittedPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.SegmentsSentPerSec,
		prometheus.CounterValue,
		float64(dst[0].SegmentsSentPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.TimestampObject,
		prometheus.CounterValue,
		float64(dst[0].Timestamp_Object),
	)
	ch <- prometheus.MustNewConstMetric(
		c.TimestampPerfTime,
		prometheus.CounterValue,
		float64(dst[0].Timestamp_PerfTime),
	)
	ch <- prometheus.MustNewConstMetric(
		c.TimestampSys100NS,
		prometheus.CounterValue,
		float64(dst[0].Timestamp_Sys100NS),
	)

	return nil, nil
}

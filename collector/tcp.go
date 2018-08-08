// returns data points from Win32_PerfRawData_Tcpip_TCPv4

// https://msdn.microsoft.com/en-us/library/aa394341(v=vs.85).aspx (Win32_PerfRawData_Tcpip_TCPv4 class)

package collector

import (
	"errors"
	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	Factories["tcp"] = NewTCPCollector
}

// A TCPCollector is a Prometheus collector for WMI Win32_PerfRawData_Tcpip_TCPv4 metrics
type TCPCollector struct {
	ConnectionFailures         *prometheus.Desc
	ConnectionsActive          *prometheus.Desc
	ConnectionsEstablished     *prometheus.Desc
	ConnectionsPassive         *prometheus.Desc
	ConnectionsReset           *prometheus.Desc
	SegmentsTotal              *prometheus.Desc
	SegmentsReceivedTotal      *prometheus.Desc
	SegmentsRetransmittedTotal *prometheus.Desc
	SegmentsSentTotal          *prometheus.Desc
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
		SegmentsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "segments_total"),
			"(TCP.SegmentsTotal)",
			nil,
			nil,
		),
		SegmentsReceivedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "segments_received_total"),
			"(TCP.SegmentsReceivedTotal)",
			nil,
			nil,
		),
		SegmentsRetransmittedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "segments_retransmitted_total"),
			"(TCP.SegmentsRetransmittedTotal)",
			nil,
			nil,
		),
		SegmentsSentTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "segments_sent_total"),
			"(TCP.SegmentsSentTotal)",
			nil,
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *TCPCollector) Collect(ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting tcp metrics:", desc, err)
		return err
	}
	return nil
}

type Win32_PerfRawData_Tcpip_TCPv4 struct {
	ConnectionFailures          uint64
	ConnectionsActive           uint64
	ConnectionsEstablished      uint64
	ConnectionsPassive          uint64
	ConnectionsReset            uint64
	SegmentsPersec              uint64
	SegmentsReceivedPersec      uint64
	SegmentsRetransmittedPersec uint64
	SegmentsSentPersec          uint64
}

func (c *TCPCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_Tcpip_TCPv4

	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}
	if len(dst) == 0 {
		return nil, errors.New("WMI query returned empty result set")
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
		c.SegmentsTotal,
		prometheus.CounterValue,
		float64(dst[0].SegmentsPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.SegmentsReceivedTotal,
		prometheus.CounterValue,
		float64(dst[0].SegmentsReceivedPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.SegmentsRetransmittedTotal,
		prometheus.CounterValue,
		float64(dst[0].SegmentsRetransmittedPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.SegmentsSentTotal,
		prometheus.CounterValue,
		float64(dst[0].SegmentsSentPersec),
	)

	return nil, nil
}

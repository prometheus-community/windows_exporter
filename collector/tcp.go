//go:build windows
// +build windows

package collector

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// A TCPCollector is a Prometheus collector for WMI Win32_PerfRawData_Tcpip_TCPv{4,6} metrics
type TCPCollector struct {
	logger log.Logger

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

// newTCPCollector ...
func newTCPCollector(logger log.Logger) (Collector, error) {
	const subsystem = "tcp"

	return &TCPCollector{
		logger: log.With(logger, "collector", subsystem),
		ConnectionFailures: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "connection_failures_total"),
			"(TCP.ConnectionFailures)",
			[]string{"af"},
			nil,
		),
		ConnectionsActive: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "connections_active_total"),
			"(TCP.ConnectionsActive)",
			[]string{"af"},
			nil,
		),
		ConnectionsEstablished: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "connections_established"),
			"(TCP.ConnectionsEstablished)",
			[]string{"af"},
			nil,
		),
		ConnectionsPassive: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "connections_passive_total"),
			"(TCP.ConnectionsPassive)",
			[]string{"af"},
			nil,
		),
		ConnectionsReset: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "connections_reset_total"),
			"(TCP.ConnectionsReset)",
			[]string{"af"},
			nil,
		),
		SegmentsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "segments_total"),
			"(TCP.SegmentsTotal)",
			[]string{"af"},
			nil,
		),
		SegmentsReceivedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "segments_received_total"),
			"(TCP.SegmentsReceivedTotal)",
			[]string{"af"},
			nil,
		),
		SegmentsRetransmittedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "segments_retransmitted_total"),
			"(TCP.SegmentsRetransmittedTotal)",
			[]string{"af"},
			nil,
		),
		SegmentsSentTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "segments_sent_total"),
			"(TCP.SegmentsSentTotal)",
			[]string{"af"},
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *TCPCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ctx, ch); err != nil {
		_ = level.Error(c.logger).Log("failed collecting tcp metrics", "desc", desc, "err", err)
		return err
	}
	return nil
}

// Win32_PerfRawData_Tcpip_TCPv4 docs
// - https://msdn.microsoft.com/en-us/library/aa394341(v=vs.85).aspx
// The TCPv6 performance object uses the same fields.
type tcp struct {
	ConnectionFailures          float64 `perflib:"Connection Failures"`
	ConnectionsActive           float64 `perflib:"Connections Active"`
	ConnectionsEstablished      float64 `perflib:"Connections Established"`
	ConnectionsPassive          float64 `perflib:"Connections Passive"`
	ConnectionsReset            float64 `perflib:"Connections Reset"`
	SegmentsPersec              float64 `perflib:"Segments/sec"`
	SegmentsReceivedPersec      float64 `perflib:"Segments Received/sec"`
	SegmentsRetransmittedPersec float64 `perflib:"Segments Retransmitted/sec"`
	SegmentsSentPersec          float64 `perflib:"Segments Sent/sec"`
}

func writeTCPCounters(metrics tcp, labels []string, c *TCPCollector, ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(
		c.ConnectionFailures,
		prometheus.CounterValue,
		metrics.ConnectionFailures,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.ConnectionsActive,
		prometheus.CounterValue,
		metrics.ConnectionsActive,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.ConnectionsEstablished,
		prometheus.GaugeValue,
		metrics.ConnectionsEstablished,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.ConnectionsPassive,
		prometheus.CounterValue,
		metrics.ConnectionsPassive,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.ConnectionsReset,
		prometheus.CounterValue,
		metrics.ConnectionsReset,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.SegmentsTotal,
		prometheus.CounterValue,
		metrics.SegmentsPersec,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.SegmentsReceivedTotal,
		prometheus.CounterValue,
		metrics.SegmentsReceivedPersec,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.SegmentsRetransmittedTotal,
		prometheus.CounterValue,
		metrics.SegmentsRetransmittedPersec,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.SegmentsSentTotal,
		prometheus.CounterValue,
		metrics.SegmentsSentPersec,
		labels...,
	)
}

func (c *TCPCollector) collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []tcp

	// TCPv4 counters
	if err := unmarshalObject(ctx.perfObjects["TCPv4"], &dst, c.logger); err != nil {
		return nil, err
	}
	if len(dst) != 0 {
		writeTCPCounters(dst[0], []string{"ipv4"}, c, ch)
	}

	// TCPv6 counters
	if err := unmarshalObject(ctx.perfObjects["TCPv6"], &dst, c.logger); err != nil {
		return nil, err
	}
	if len(dst) != 0 {
		writeTCPCounters(dst[0], []string{"ipv6"}, c, ch)
	}

	return nil, nil
}

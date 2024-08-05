//go:build windows

package tcp

import (
	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "tcp"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_Tcpip_TCPv{4,6} metrics
type Collector struct {
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
	return []string{"TCPv4"}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build() error {
	c.ConnectionFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connection_failures_total"),
		"(TCP.ConnectionFailures)",
		[]string{"af"},
		nil,
	)
	c.ConnectionsActive = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connections_active_total"),
		"(TCP.ConnectionsActive)",
		[]string{"af"},
		nil,
	)
	c.ConnectionsEstablished = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connections_established"),
		"(TCP.ConnectionsEstablished)",
		[]string{"af"},
		nil,
	)
	c.ConnectionsPassive = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connections_passive_total"),
		"(TCP.ConnectionsPassive)",
		[]string{"af"},
		nil,
	)
	c.ConnectionsReset = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connections_reset_total"),
		"(TCP.ConnectionsReset)",
		[]string{"af"},
		nil,
	)
	c.SegmentsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "segments_total"),
		"(TCP.SegmentsTotal)",
		[]string{"af"},
		nil,
	)
	c.SegmentsReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "segments_received_total"),
		"(TCP.SegmentsReceivedTotal)",
		[]string{"af"},
		nil,
	)
	c.SegmentsRetransmittedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "segments_retransmitted_total"),
		"(TCP.SegmentsRetransmittedTotal)",
		[]string{"af"},
		nil,
	)
	c.SegmentsSentTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "segments_sent_total"),
		"(TCP.SegmentsSentTotal)",
		[]string{"af"},
		nil,
	)
	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.collect(ctx, ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting tcp metrics", "err", err)
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

func writeTCPCounters(metrics tcp, labels []string, c *Collector, ch chan<- prometheus.Metric) {
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

func (c *Collector) collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var dst []tcp

	// TCPv4 counters
	if err := perflib.UnmarshalObject(ctx.PerfObjects["TCPv4"], &dst, c.logger); err != nil {
		return err
	}
	if len(dst) != 0 {
		writeTCPCounters(dst[0], []string{"ipv4"}, c, ch)
	}

	// TCPv6 counters
	if err := perflib.UnmarshalObject(ctx.PerfObjects["TCPv6"], &dst, c.logger); err != nil {
		return err
	}
	if len(dst) != 0 {
		writeTCPCounters(dst[0], []string{"ipv6"}, c, ch)
	}

	return nil
}

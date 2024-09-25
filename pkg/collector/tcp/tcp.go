//go:build windows

package tcp

import (
	"fmt"
	"log/slog"

    "github.com/prometheus-community/windows_exporter/pkg/perfdata"
    "github.com/yusufpapurcu/wmi"
	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/pkg/headers/iphlpapi"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_Tcpip_TCPv{4,6} metrics.
type Collector struct {
	config Config

	perfDataCollector4 *perfdata.Collector
	perfDataCollector6 *perfdata.Collector

	connectionFailures         *prometheus.Desc
	connectionsActive          *prometheus.Desc
	connectionsEstablished     *prometheus.Desc
	connectionsPassive         *prometheus.Desc
	connectionsReset           *prometheus.Desc
	segmentsTotal              *prometheus.Desc
	segmentsReceivedTotal      *prometheus.Desc
	segmentsRetransmittedTotal *prometheus.Desc
	segmentsSentTotal          *prometheus.Desc
	connectionsStateCount      *prometheus.Desc
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

func (c *Collector) Build(_ *slog.Logger, _ *wmi.Client) error {
	counters := []string{
		ConnectionFailures,
		ConnectionsActive,
		ConnectionsEstablished,
		ConnectionsPassive,
		ConnectionsReset,
		SegmentsPersec,
		SegmentsReceivedPersec,
		SegmentsRetransmittedPersec,
		SegmentsSentPersec,
	}

	var err error

	c.perfDataCollector4, err = perfdata.NewCollector("TCPv4", nil, counters)
	if err != nil {
		return fmt.Errorf("failed to create TCPv4 collector: %w", err)
	}

	c.perfDataCollector6, err = perfdata.NewCollector("TCPv6", nil, counters)
	if err != nil {
		return fmt.Errorf("failed to create TCPv6 collector: %w", err)
	}

	c.connectionFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connection_failures_total"),
		"(TCP.ConnectionFailures)",
		[]string{"af"},
		nil,
	)
	c.connectionsActive = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connections_active_total"),
		"(TCP.ConnectionsActive)",
		[]string{"af"},
		nil,
	)
	c.connectionsEstablished = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connections_established"),
		"(TCP.ConnectionsEstablished)",
		[]string{"af"},
		nil,
	)
	c.connectionsPassive = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connections_passive_total"),
		"(TCP.ConnectionsPassive)",
		[]string{"af"},
		nil,
	)
	c.connectionsReset = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connections_reset_total"),
		"(TCP.ConnectionsReset)",
		[]string{"af"},
		nil,
	)
	c.segmentsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "segments_total"),
		"(TCP.SegmentsTotal)",
		[]string{"af"},
		nil,
	)
	c.segmentsReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "segments_received_total"),
		"(TCP.SegmentsReceivedTotal)",
		[]string{"af"},
		nil,
	)
	c.segmentsRetransmittedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "segments_retransmitted_total"),
		"(TCP.SegmentsRetransmittedTotal)",
		[]string{"af"},
		nil,
	)
	c.segmentsSentTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "segments_sent_total"),
		"(TCP.SegmentsSentTotal)",
		[]string{"af"},
		nil,
	)
	c.connectionsStateCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connections_state_count"),
		"Number of TCP connections by state and address family",
		[]string{"af", "state"}, nil,
	)

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	logger = logger.With(slog.String("collector", Name))
	if err := c.collect(ch); err != nil {
		logger.Error("failed collecting tcp metrics",
			slog.Any("err", err),
		)
		return err
	}

	return nil
}

func writeTCPCounters(metrics map[string]perfdata.CounterValues, labels []string, c *Collector, ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(
		c.connectionFailures,
		prometheus.CounterValue,
		metrics[ConnectionFailures].FirstValue,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.connectionsActive,
		prometheus.CounterValue,
		metrics[ConnectionsActive].FirstValue,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.connectionsEstablished,
		prometheus.GaugeValue,
		metrics[ConnectionsEstablished].FirstValue,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.connectionsPassive,
		prometheus.CounterValue,
		metrics[ConnectionsPassive].FirstValue,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.connectionsReset,
		prometheus.CounterValue,
		metrics[ConnectionsReset].FirstValue,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.segmentsTotal,
		prometheus.CounterValue,
		metrics[SegmentsPersec].FirstValue,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.segmentsReceivedTotal,
		prometheus.CounterValue,
		metrics[SegmentsReceivedPersec].FirstValue,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.segmentsRetransmittedTotal,
		prometheus.CounterValue,
		metrics[SegmentsRetransmittedPersec].FirstValue,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.segmentsSentTotal,
		prometheus.CounterValue,
		metrics[SegmentsSentPersec].FirstValue,
		labels...,
	)
}

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	stateCounts, err := iphlpapi.GetTCPConnectionStates()
	if err != nil {
		return fmt.Errorf("failed to collect TCP connection states for %s: %w", "ipv4", err)
	}
	c.sendTCPStateMetrics(ch, stateCounts, "ipv4")

	stateCounts, err = iphlpapi.GetTCP6ConnectionStates()

	if err != nil {
		return fmt.Errorf("failed to collect TCP6 connection states for %s: %w", "ipv6", err)
	}
	c.sendTCPStateMetrics(ch, stateCounts, "ipv6")

	return nil
}

func (c *Collector) sendTCPStateMetrics(ch chan<- prometheus.Metric, stateCounts map[uint32]uint32, af string) {
	for state, count := range stateCounts {
		stateName := getTCPStateName(state)
		ch <- prometheus.MustNewConstMetric(
			c.connectionsStateCount,
			prometheus.GaugeValue,
			float64(count),
			af,
			stateName,
		)
	}
}

func getTCPStateName(state uint32) string {
	switch state {
	case TCPStateClosed:
		return "CLOSED"
	case TCPStateListening:
		return "LISTENING"
	case TCPStateSynSent:
		return "SYN_SENT"
	case TCPStateSynRcvd:
		return "SYN_RECEIVED"
	case TCPStateEstablished:
		return "ESTABLISHED"
	case TCPStateFinWait1:
		return "FIN_WAIT1"
	case TCPStateFinWait2:
		return "FIN_WAIT2"
	case TCPStateCloseWait:
		return "CLOSE_WAIT"
	case TCPStateClosing:
		return "CLOSING"
	case TCPStateLastAck:
		return "LAST_ACK"
	case TCPStateTimeWait:
		return "TIME_WAIT"
	case TCPStateDeleteTcb:
		return "DELETE_TCB"
	default:
		return fmt.Sprintf("UNKNOWN_%d", state)
	}
}

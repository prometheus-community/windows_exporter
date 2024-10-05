//go:build windows

package tcp

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/headers/iphlpapi"
	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/perfdata/perftypes"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
	"golang.org/x/sys/windows"
)

const Name = "tcp"

type Config struct {
	CollectorsEnabled []string `yaml:"collectors_enabled"`
}

var ConfigDefaults = Config{
	CollectorsEnabled: []string{
		"metrics",
		"connections_state",
	},
}

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_Tcpip_TCPv{4,6} metrics.
type Collector struct {
	config Config

	perfDataCollector4 perfdata.Collector
	perfDataCollector6 perfdata.Collector

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

	if config.CollectorsEnabled == nil {
		config.CollectorsEnabled = ConfigDefaults.CollectorsEnabled
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(app *kingpin.Application) *Collector {
	c := &Collector{
		config: ConfigDefaults,
	}
	c.config.CollectorsEnabled = make([]string, 0)

	var collectorsEnabled string

	app.Flag(
		"collector.tcp.enabled",
		"Comma-separated list of collectors to use. Defaults to all, if not specified.",
	).Default(strings.Join(ConfigDefaults.CollectorsEnabled, ",")).StringVar(&collectorsEnabled)

	app.Action(func(*kingpin.ParseContext) error {
		c.config.CollectorsEnabled = strings.Split(collectorsEnabled, ",")

		return nil
	})

	return c
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
		connectionFailures,
		connectionsActive,
		connectionsEstablished,
		connectionsPassive,
		connectionsReset,
		segmentsPerSec,
		segmentsReceivedPerSec,
		segmentsRetransmittedPerSec,
		segmentsSentPerSec,
	}

	var err error

	c.perfDataCollector4, err = perfdata.NewCollector(perfdata.V1, "TCPv4", nil, counters)
	if err != nil {
		return fmt.Errorf("failed to create TCPv4 collector: %w", err)
	}

	c.perfDataCollector6, err = perfdata.NewCollector(perfdata.V1, "TCPv6", nil, counters)
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

	if slices.Contains(c.config.CollectorsEnabled, "metrics") {
		if err := c.collect(ch); err != nil {
			logger.Error("failed collecting tcp metrics",
				slog.Any("err", err),
			)

			return err
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, "connections_state") {
		if err := c.collectConnectionsState(ch); err != nil {
			logger.Error("failed collecting tcp connection state metrics",
				slog.Any("err", err),
			)

			return err
		}
	}

	return nil
}

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	data, err := c.perfDataCollector4.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect TCPv4 metrics: %w", err)
	}

	c.writeTCPCounters(ch, data[perftypes.EmptyInstance], []string{"ipv4"})

	data, err = c.perfDataCollector6.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect TCPv6 metrics: %w", err)
	}

	c.writeTCPCounters(ch, data[perftypes.EmptyInstance], []string{"ipv6"})

	return nil
}

func (c *Collector) writeTCPCounters(ch chan<- prometheus.Metric, metrics map[string]perftypes.CounterValues, labels []string) {
	ch <- prometheus.MustNewConstMetric(
		c.connectionFailures,
		prometheus.CounterValue,
		metrics[connectionFailures].FirstValue,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.connectionsActive,
		prometheus.CounterValue,
		metrics[connectionsActive].FirstValue,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.connectionsEstablished,
		prometheus.GaugeValue,
		metrics[connectionsEstablished].FirstValue,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.connectionsPassive,
		prometheus.CounterValue,
		metrics[connectionsPassive].FirstValue,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.connectionsReset,
		prometheus.CounterValue,
		metrics[connectionsReset].FirstValue,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.segmentsTotal,
		prometheus.CounterValue,
		metrics[segmentsPerSec].FirstValue,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.segmentsReceivedTotal,
		prometheus.CounterValue,
		metrics[segmentsReceivedPerSec].FirstValue,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.segmentsRetransmittedTotal,
		prometheus.CounterValue,
		metrics[segmentsRetransmittedPerSec].FirstValue,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.segmentsSentTotal,
		prometheus.CounterValue,
		metrics[segmentsSentPerSec].FirstValue,
		labels...,
	)
}

func (c *Collector) collectConnectionsState(ch chan<- prometheus.Metric) error {
	stateCounts, err := iphlpapi.GetTCPConnectionStates(windows.AF_INET)
	if err != nil {
		return fmt.Errorf("failed to collect TCP connection states for %s: %w", "ipv4", err)
	}

	c.sendTCPStateMetrics(ch, stateCounts, "ipv4")

	stateCounts, err = iphlpapi.GetTCPConnectionStates(windows.AF_INET6)
	if err != nil {
		return fmt.Errorf("failed to collect TCP6 connection states for %s: %w", "ipv6", err)
	}

	c.sendTCPStateMetrics(ch, stateCounts, "ipv6")

	return nil
}

func (c *Collector) sendTCPStateMetrics(ch chan<- prometheus.Metric, stateCounts map[iphlpapi.MIB_TCP_STATE]uint32, af string) {
	for state, count := range stateCounts {
		ch <- prometheus.MustNewConstMetric(
			c.connectionsStateCount,
			prometheus.GaugeValue,
			float64(count),
			af,
			state.String(),
		)
	}
}

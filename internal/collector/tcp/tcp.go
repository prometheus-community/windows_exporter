// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package tcp

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/headers/iphlpapi"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows"
)

const (
	Name = "tcp"

	ipAddressFamilyIPv4 = "ipv4"
	ipAddressFamilyIPv6 = "ipv6"

	subCollectorMetrics          = "metrics"
	subCollectorConnectionsState = "connections_state"
	subCollectorDynamicPorts     = "dynamic_ports"
)

type Config struct {
	CollectorsEnabled []string `yaml:"enabled"`
}

//nolint:gochecknoglobals
var ConfigDefaults = Config{
	CollectorsEnabled: []string{
		subCollectorMetrics,
		subCollectorConnectionsState,
		subCollectorDynamicPorts,
	},
}

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_Tcpip_TCPv{4,6} metrics.
type Collector struct {
	config Config

	// perfDataQuery samples the IPv4 and IPv6 objects together, so the
	// collector waits only once for the other performance counter queries.
	perfDataQuery      *pdh.Query
	perfDataCollector4 *pdh.Collector[perfDataCounterValues]
	perfDataCollector6 *pdh.Collector[perfDataCounterValues]
	perfDataObject4    []perfDataCounterValues
	perfDataObject6    []perfDataCounterValues

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

	// dynamicPortRange is the range Windows assigns local ports of outgoing
	// connections from, read once in Build. It is empty when the
	// dynamic_ports sub-collector is disabled.
	dynamicPortRange     iphlpapi.PortRange
	dynamicPortsInUse    *prometheus.Desc
	dynamicPortRangeSize *prometheus.Desc
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

func (c *Collector) Close() error {
	if slices.Contains(c.config.CollectorsEnabled, subCollectorMetrics) {
		c.perfDataCollector4.Close()
		c.perfDataCollector6.Close()
		c.perfDataQuery.Close()
	}

	return nil
}

func (c *Collector) Build(logger *slog.Logger, miSession *mi.Session) error {
	subCollectors := []string{subCollectorMetrics, subCollectorConnectionsState, subCollectorDynamicPorts}

	for _, collector := range c.config.CollectorsEnabled {
		if !slices.Contains(subCollectors, collector) {
			return fmt.Errorf("unknown sub collector: %s. Possible values: %s", collector,
				strings.Join(subCollectors, ", "),
			)
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, subCollectorConnectionsState) ||
		slices.Contains(c.config.CollectorsEnabled, subCollectorDynamicPorts) {
		if err := iphlpapi.Load(); err != nil {
			return err
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, subCollectorDynamicPorts) {
		portRange, err := queryDynamicPortRange(miSession)
		if err != nil {
			// The other sub-collectors do not need MI, so a broken provider
			// disables only this one instead of the whole collector.
			logger.Warn("failed to read the TCP dynamic port range, disabling the dynamic_ports sub-collector",
				slog.String("collector", Name),
				slog.Any("err", err),
			)

			c.config.CollectorsEnabled = slices.DeleteFunc(slices.Clone(c.config.CollectorsEnabled), func(name string) bool {
				return name == subCollectorDynamicPorts
			})
		} else {
			c.dynamicPortRange = portRange
		}
	}

	labels := []string{"af"}

	c.connectionFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connection_failures_total"),
		"(TCP.ConnectionFailures)",
		labels,
		nil,
	)
	c.connectionsActive = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connections_active_total"),
		"(TCP.ConnectionsActive)",
		labels,
		nil,
	)
	c.connectionsEstablished = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connections_established"),
		"(TCP.ConnectionsEstablished)",
		labels,
		nil,
	)
	c.connectionsPassive = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connections_passive_total"),
		"(TCP.ConnectionsPassive)",
		labels,
		nil,
	)
	c.connectionsReset = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connections_reset_total"),
		"(TCP.ConnectionsReset)",
		labels,
		nil,
	)
	c.segmentsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "segments_total"),
		"(TCP.SegmentsTotal)",
		labels,
		nil,
	)
	c.segmentsReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "segments_received_total"),
		"(TCP.SegmentsReceivedTotal)",
		labels,
		nil,
	)
	c.segmentsRetransmittedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "segments_retransmitted_total"),
		"(TCP.SegmentsRetransmittedTotal)",
		labels,
		nil,
	)
	c.segmentsSentTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "segments_sent_total"),
		"(TCP.SegmentsSentTotal)",
		labels,
		nil,
	)
	c.connectionsStateCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connections"),
		"Number of TCP connections by state and address family",
		[]string{"af", "state"},
		nil,
	)
	c.dynamicPortsInUse = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_ports_in_use"),
		"Number of distinct local ports inside the TCP dynamic port range that a connection in any state, including TIME_WAIT, holds",
		labels,
		nil,
	)
	c.dynamicPortRangeSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_port_range_size"),
		"Number of ports in the TCP dynamic port range, read at exporter start",
		labels,
		nil,
	)

	errs := make([]error, 0)

	if slices.Contains(c.config.CollectorsEnabled, subCollectorMetrics) {
		var err error

		c.perfDataQuery, err = pdh.NewQuery()
		if err != nil {
			return fmt.Errorf("failed to open TCP performance counter query: %w", err)
		}

		c.perfDataCollector4, err = pdh.NewQueryCollector[perfDataCounterValues](logger.With(slog.String("collector", Name)), c.perfDataQuery, pdh.CounterTypeRaw, "TCPv4", nil)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to create TCPv4 collector: %w", err))
		}

		c.perfDataCollector6, err = pdh.NewQueryCollector[perfDataCounterValues](logger.With(slog.String("collector", Name)), c.perfDataQuery, pdh.CounterTypeRaw, "TCPv6", nil)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to create TCPv6 collector: %w", err))
		}
	}

	return errors.Join(errs...)
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric, _ time.Duration) error {
	errs := make([]error, 0)

	if slices.Contains(c.config.CollectorsEnabled, subCollectorMetrics) {
		if err := c.collect(ch); err != nil {
			errs = append(errs, fmt.Errorf("failed collecting tcp metrics: %w", err))
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, subCollectorConnectionsState) ||
		slices.Contains(c.config.CollectorsEnabled, subCollectorDynamicPorts) {
		if err := c.collectConnectionTable(ch); err != nil {
			errs = append(errs, fmt.Errorf("failed collecting tcp connection table metrics: %w", err))
		}
	}

	return errors.Join(errs...)
}

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	errs := make([]error, 0)

	if err := c.perfDataQuery.Collect(); err != nil {
		return fmt.Errorf("failed to collect TCP metrics: %w", err)
	}

	if err := c.perfDataCollector4.Collect(&c.perfDataObject4); err != nil {
		errs = append(errs, fmt.Errorf("failed to collect TCPv4 metrics. %w", err))
	} else if len(c.perfDataObject4) == 0 {
		errs = append(errs, fmt.Errorf("failed to collect TCPv4 metrics: %w", types.ErrNoDataUnexpected))
	} else {
		c.writeTCPCounters(ch, c.perfDataObject4, ipAddressFamilyIPv4)
	}

	if err := c.perfDataCollector6.Collect(&c.perfDataObject6); err != nil {
		errs = append(errs, fmt.Errorf("failed to collect TCPv6 metrics. %w", err))
	} else if len(c.perfDataObject6) == 0 {
		errs = append(errs, fmt.Errorf("failed to collect TCPv6 metrics: %w", types.ErrNoDataUnexpected))
	} else {
		c.writeTCPCounters(ch, c.perfDataObject6, ipAddressFamilyIPv6)
	}

	return errors.Join(errs...)
}

func (c *Collector) writeTCPCounters(ch chan<- prometheus.Metric, metrics []perfDataCounterValues, af string) {
	ch <- prometheus.MustNewConstMetric(
		c.connectionFailures,
		prometheus.CounterValue,
		metrics[0].ConnectionFailures,
		af,
	)

	ch <- prometheus.MustNewConstMetric(
		c.connectionsActive,
		prometheus.CounterValue,
		metrics[0].ConnectionsActive,
		af,
	)

	ch <- prometheus.MustNewConstMetric(
		c.connectionsEstablished,
		prometheus.GaugeValue,
		metrics[0].ConnectionsEstablished,
		af,
	)

	ch <- prometheus.MustNewConstMetric(
		c.connectionsPassive,
		prometheus.CounterValue,
		metrics[0].ConnectionsPassive,
		af,
	)

	ch <- prometheus.MustNewConstMetric(
		c.connectionsReset,
		prometheus.CounterValue,
		metrics[0].ConnectionsReset,
		af,
	)

	ch <- prometheus.MustNewConstMetric(
		c.segmentsTotal,
		prometheus.CounterValue,
		metrics[0].SegmentsPerSec,
		af,
	)

	ch <- prometheus.MustNewConstMetric(
		c.segmentsReceivedTotal,
		prometheus.CounterValue,
		metrics[0].SegmentsReceivedPerSec,
		af,
	)

	ch <- prometheus.MustNewConstMetric(
		c.segmentsRetransmittedTotal,
		prometheus.CounterValue,
		metrics[0].SegmentsRetransmittedPerSec,
		af,
	)

	ch <- prometheus.MustNewConstMetric(
		c.segmentsSentTotal,
		prometheus.CounterValue,
		metrics[0].SegmentsSentPerSec,
		af,
	)
}

// collectConnectionTable reads the connection table of each address family
// once and publishes the connection states and the dynamic port usage from it.
func (c *Collector) collectConnectionTable(ch chan<- prometheus.Metric) error {
	errs := make([]error, 0)

	for _, family := range []struct {
		id   uint32
		name string
	}{
		{windows.AF_INET, ipAddressFamilyIPv4},
		{windows.AF_INET6, ipAddressFamilyIPv6},
	} {
		stats, err := iphlpapi.GetTCPConnectionStats(family.id, c.dynamicPortRange)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to read the TCP connection table for %s: %w", family.name, err))

			continue
		}

		c.writeConnectionTableMetrics(ch, stats, family.name)
	}

	return errors.Join(errs...)
}

func (c *Collector) writeConnectionTableMetrics(ch chan<- prometheus.Metric, stats iphlpapi.TCPConnectionStats, af string) {
	if slices.Contains(c.config.CollectorsEnabled, subCollectorConnectionsState) {
		for state, count := range stats.States {
			ch <- prometheus.MustNewConstMetric(
				c.connectionsStateCount,
				prometheus.GaugeValue,
				float64(count),
				af,
				state.String(),
			)
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, subCollectorDynamicPorts) {
		ch <- prometheus.MustNewConstMetric(
			c.dynamicPortsInUse,
			prometheus.GaugeValue,
			float64(stats.PortsInRange),
			af,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dynamicPortRangeSize,
			prometheus.GaugeValue,
			float64(c.dynamicPortRange.NumberOfPorts),
			af,
		)
	}
}

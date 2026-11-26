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
	"github.com/prometheus-community/windows_exporter/internal/headers/nsi"
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
	subCollectorDynamicPortRange = "dynamic_port_range"
)

type Config struct {
	CollectorsEnabled []string `yaml:"enabled"`
}

//nolint:gochecknoglobals
var ConfigDefaults = Config{
	CollectorsEnabled: []string{
		subCollectorMetrics,
		subCollectorConnectionsState,
		subCollectorDynamicPortRange,
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
	dynamicPortRangeSize       *prometheus.Desc
	dynamicPortRangeExcluded   *prometheus.Desc
	dynamicPortRangeUsed       *prometheus.Desc
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

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	subCollectors := []string{subCollectorMetrics, subCollectorConnectionsState, subCollectorDynamicPortRange}

	for _, collector := range c.config.CollectorsEnabled {
		if !slices.Contains(subCollectors, collector) {
			return fmt.Errorf("unknown sub collector: %s. Possible values: %s", collector,
				strings.Join(subCollectors, ", "),
			)
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, subCollectorConnectionsState) ||
		slices.Contains(c.config.CollectorsEnabled, subCollectorDynamicPortRange) {
		if err := iphlpapi.Load(); err != nil {
			return err
		}
	}

	// NSI is undocumented. Without it, run without the range instead of
	// failing the whole collector.
	if slices.Contains(c.config.CollectorsEnabled, subCollectorDynamicPortRange) {
		if err := nsi.Load(); err != nil {
			logger.Warn("failed to load the network store interface, disabling the dynamic_port_range sub-collector",
				slog.String("collector", Name),
				slog.Any("err", err),
			)

			c.config.CollectorsEnabled = slices.DeleteFunc(slices.Clone(c.config.CollectorsEnabled), func(name string) bool {
				return name == subCollectorDynamicPortRange
			})
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
	c.dynamicPortRangeSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_port_range_size"),
		"Number of ports in the TCP dynamic port range, which IPv4 and IPv6 share",
		nil,
		nil,
	)
	c.dynamicPortRangeExcluded = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_port_range_excluded_ports"),
		"Number of ports in the TCP dynamic port range that an excluded port range reserves",
		nil,
		nil,
	)
	c.dynamicPortRangeUsed = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_port_range_observed_used_ports"),
		"Number of ports in the TCP dynamic port range that a socket holds, as seen in the TCP tables; excluded ports are not counted",
		nil,
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

	connectionsState := slices.Contains(c.config.CollectorsEnabled, subCollectorConnectionsState)
	dynamicPortRange := slices.Contains(c.config.CollectorsEnabled, subCollectorDynamicPortRange)

	// Both sub-collectors read the same connection tables, so read them once.
	var ports *iphlpapi.PortSet
	if dynamicPortRange {
		ports = new(iphlpapi.PortSet)
	}

	portsComplete := true

	if connectionsState || dynamicPortRange {
		if err := c.collectConnectionsState(ch, connectionsState, ports); err != nil {
			errs = append(errs, fmt.Errorf("failed to read the tcp connection tables: %w", err))
			portsComplete = false
		}
	}

	if dynamicPortRange {
		if !portsComplete {
			ports = nil
		}

		if err := c.collectDynamicPortRange(ch, ports); err != nil {
			errs = append(errs, err)
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

// collectConnectionsState reads the connection table of both address
// families. It sends the state metrics if send is set, and adds the local
// ports that connections hold to ports if ports is not nil.
func (c *Collector) collectConnectionsState(ch chan<- prometheus.Metric, send bool, ports *iphlpapi.PortSet) error {
	errs := make([]error, 0)

	for _, family := range []struct {
		af   uint32
		name string
	}{
		{af: windows.AF_INET, name: ipAddressFamilyIPv4},
		{af: windows.AF_INET6, name: ipAddressFamilyIPv6},
	} {
		stateCounts, err := iphlpapi.GetTCPConnectionStates(family.af, ports)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to collect TCP connection states for %s: %w", family.name, err))

			continue
		}

		if send {
			c.sendTCPStateMetrics(ch, stateCounts, family.name)
		}
	}

	return errors.Join(errs...)
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

// collectDynamicPortRange reads the active range and exclusions on every
// scrape, so a change with netsh shows up without a restart. ports holds the
// local ports of the connection tables, or is nil if a table read failed. A
// failed read drops the metrics that depend on it rather than reporting zero
// or a default.
func (c *Collector) collectDynamicPortRange(ch chan<- prometheus.Metric, ports *iphlpapi.PortSet) error {
	portRange, err := nsi.DynamicPortRange(nsi.ProtocolTCP)
	if err != nil {
		return fmt.Errorf("failed to read the TCP dynamic port range: %w", err)
	}

	ch <- prometheus.MustNewConstMetric(
		c.dynamicPortRangeSize,
		prometheus.GaugeValue,
		float64(portRange.NumberOfPorts),
	)

	exclusions, err := nsi.ExcludedPortRanges(nsi.ProtocolTCP)
	if err != nil {
		return fmt.Errorf("failed to read the TCP excluded port ranges: %w", err)
	}

	var errs []error

	if ports != nil {
		// Sockets that are bound but neither listen nor connect are missing
		// from the connection tables.
		bound, err := nsi.BoundTCPPorts()
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to read the bound TCP endpoints: %w", err))
			ports = nil
		}

		for _, port := range bound {
			ports.Add(port)
		}
	}

	excluded, used := portRange.Usage(exclusions, func(port uint16) bool {
		return ports != nil && ports.Contains(port)
	})

	ch <- prometheus.MustNewConstMetric(
		c.dynamicPortRangeExcluded,
		prometheus.GaugeValue,
		float64(excluded),
	)

	if ports != nil {
		ch <- prometheus.MustNewConstMetric(
			c.dynamicPortRangeUsed,
			prometheus.GaugeValue,
			float64(used),
		)
	}

	return errors.Join(errs...)
}

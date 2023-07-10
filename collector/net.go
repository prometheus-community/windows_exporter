//go:build windows
// +build windows

package collector

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	FlagNicOldExclude = "collector.net.nic-blacklist"
	FlagNicOldInclude = "collector.net.nic-whitelist"

	FlagNicExclude = "collector.net.nic-exclude"
	FlagNicInclude = "collector.net.nic-include"
)

var (
	nicOldInclude *string
	nicOldExclude *string

	nicInclude *string
	nicExclude *string

	nicIncludeSet bool
	nicExcludeSet bool

	nicNameToUnderscore = regexp.MustCompile("[^a-zA-Z0-9]")
)

// A NetworkCollector is a Prometheus collector for Perflib Network Interface metrics
type NetworkCollector struct {
	logger log.Logger

	BytesReceivedTotal       *prometheus.Desc
	BytesSentTotal           *prometheus.Desc
	BytesTotal               *prometheus.Desc
	OutputQueueLength        *prometheus.Desc
	PacketsOutboundDiscarded *prometheus.Desc
	PacketsOutboundErrors    *prometheus.Desc
	PacketsTotal             *prometheus.Desc
	PacketsReceivedDiscarded *prometheus.Desc
	PacketsReceivedErrors    *prometheus.Desc
	PacketsReceivedTotal     *prometheus.Desc
	PacketsReceivedUnknown   *prometheus.Desc
	PacketsSentTotal         *prometheus.Desc
	CurrentBandwidth         *prometheus.Desc

	nicIncludePattern *regexp.Regexp
	nicExcludePattern *regexp.Regexp
}

// newNetworkCollectorFlags ...
func newNetworkCollectorFlags(app *kingpin.Application) {
	nicInclude = app.Flag(
		FlagNicInclude,
		"Regexp of NIC:s to include. NIC name must both match include and not match exclude to be included.",
	).Default(".+").PreAction(func(c *kingpin.ParseContext) error {
		nicIncludeSet = true
		return nil
	}).String()

	nicExclude = app.Flag(
		FlagNicExclude,
		"Regexp of NIC:s to exclude. NIC name must both match include and not match exclude to be included.",
	).Default("").PreAction(func(c *kingpin.ParseContext) error {
		nicExcludeSet = true
		return nil
	}).String()

	nicOldInclude = app.Flag(
		FlagNicOldInclude,
		"DEPRECATED: Use --collector.net.nic-include",
	).Hidden().String()
	nicOldExclude = app.Flag(
		FlagNicOldExclude,
		"DEPRECATED: Use --collector.net.nic-exclude",
	).Hidden().String()

}

// newNetworkCollector ...
func newNetworkCollector(logger log.Logger) (Collector, error) {
	const subsystem = "net"
	logger = log.With(logger, "collector", subsystem)

	if *nicOldExclude != "" {
		if !nicExcludeSet {
			_ = level.Warn(logger).Log("msg", "--collector.net.nic-blacklist is DEPRECATED and will be removed in a future release, use --collector.net.nic-exclude")
			*nicExclude = *nicOldExclude
		} else {
			return nil, errors.New("--collector.net.nic-blacklist and --collector.net.nic-exclude are mutually exclusive")
		}
	}
	if *nicOldInclude != "" {
		if !nicIncludeSet {
			_ = level.Warn(logger).Log("msg", "--collector.net.nic-whitelist is DEPRECATED and will be removed in a future release, use --collector.net.nic-include")
			*nicInclude = *nicOldInclude
		} else {
			return nil, errors.New("--collector.net.nic-whitelist and --collector.net.nic-include are mutually exclusive")
		}
	}

	return &NetworkCollector{
		logger: logger,
		BytesReceivedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bytes_received_total"),
			"(Network.BytesReceivedPerSec)",
			[]string{"nic"},
			nil,
		),
		BytesSentTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bytes_sent_total"),
			"(Network.BytesSentPerSec)",
			[]string{"nic"},
			nil,
		),
		BytesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bytes_total"),
			"(Network.BytesTotalPerSec)",
			[]string{"nic"},
			nil,
		),
		OutputQueueLength: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "output_queue_length_packets"),
			"(Network.OutputQueueLength)",
			[]string{"nic"},
			nil,
		),
		PacketsOutboundDiscarded: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "packets_outbound_discarded_total"),
			"(Network.PacketsOutboundDiscarded)",
			[]string{"nic"},
			nil,
		),
		PacketsOutboundErrors: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "packets_outbound_errors_total"),
			"(Network.PacketsOutboundErrors)",
			[]string{"nic"},
			nil,
		),
		PacketsReceivedDiscarded: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "packets_received_discarded_total"),
			"(Network.PacketsReceivedDiscarded)",
			[]string{"nic"},
			nil,
		),
		PacketsReceivedErrors: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "packets_received_errors_total"),
			"(Network.PacketsReceivedErrors)",
			[]string{"nic"},
			nil,
		),
		PacketsReceivedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "packets_received_total"),
			"(Network.PacketsReceivedPerSec)",
			[]string{"nic"},
			nil,
		),
		PacketsReceivedUnknown: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "packets_received_unknown_total"),
			"(Network.PacketsReceivedUnknown)",
			[]string{"nic"},
			nil,
		),
		PacketsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "packets_total"),
			"(Network.PacketsPerSec)",
			[]string{"nic"},
			nil,
		),
		PacketsSentTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "packets_sent_total"),
			"(Network.PacketsSentPerSec)",
			[]string{"nic"},
			nil,
		),
		CurrentBandwidth: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "current_bandwidth_bytes"),
			"(Network.CurrentBandwidth)",
			[]string{"nic"},
			nil,
		),

		nicIncludePattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *nicInclude)),
		nicExcludePattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *nicExclude)),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *NetworkCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ctx, ch); err != nil {
		_ = level.Error(c.logger).Log("failed collecting net metrics", "desc", desc, "err", err)
		return err
	}
	return nil
}

// mangleNetworkName mangles Network Adapter name (non-alphanumeric to _)
// that is used in networkInterface.
func mangleNetworkName(name string) string {
	return nicNameToUnderscore.ReplaceAllString(name, "_")
}

// Win32_PerfRawData_Tcpip_NetworkInterface docs:
// - https://technet.microsoft.com/en-us/security/aa394340(v=vs.80)
type networkInterface struct {
	BytesReceivedPerSec      float64 `perflib:"Bytes Received/sec"`
	BytesSentPerSec          float64 `perflib:"Bytes Sent/sec"`
	BytesTotalPerSec         float64 `perflib:"Bytes Total/sec"`
	Name                     string
	OutputQueueLength        float64 `perflib:"Output Queue Length"`
	PacketsOutboundDiscarded float64 `perflib:"Packets Outbound Discarded"`
	PacketsOutboundErrors    float64 `perflib:"Packets Outbound Errors"`
	PacketsPerSec            float64 `perflib:"Packets/sec"`
	PacketsReceivedDiscarded float64 `perflib:"Packets Received Discarded"`
	PacketsReceivedErrors    float64 `perflib:"Packets Received Errors"`
	PacketsReceivedPerSec    float64 `perflib:"Packets Received/sec"`
	PacketsReceivedUnknown   float64 `perflib:"Packets Received Unknown"`
	PacketsSentPerSec        float64 `perflib:"Packets Sent/sec"`
	CurrentBandwidth         float64 `perflib:"Current Bandwidth"`
}

func (c *NetworkCollector) collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []networkInterface

	if err := unmarshalObject(ctx.perfObjects["Network Interface"], &dst, c.logger); err != nil {
		return nil, err
	}

	for _, nic := range dst {
		if c.nicExcludePattern.MatchString(nic.Name) ||
			!c.nicIncludePattern.MatchString(nic.Name) {
			continue
		}

		name := mangleNetworkName(nic.Name)
		if name == "" {
			continue
		}

		// Counters
		ch <- prometheus.MustNewConstMetric(
			c.BytesReceivedTotal,
			prometheus.CounterValue,
			nic.BytesReceivedPerSec,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.BytesSentTotal,
			prometheus.CounterValue,
			nic.BytesSentPerSec,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.BytesTotal,
			prometheus.CounterValue,
			nic.BytesTotalPerSec,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.OutputQueueLength,
			prometheus.GaugeValue,
			nic.OutputQueueLength,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PacketsOutboundDiscarded,
			prometheus.CounterValue,
			nic.PacketsOutboundDiscarded,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PacketsOutboundErrors,
			prometheus.CounterValue,
			nic.PacketsOutboundErrors,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PacketsTotal,
			prometheus.CounterValue,
			nic.PacketsPerSec,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PacketsReceivedDiscarded,
			prometheus.CounterValue,
			nic.PacketsReceivedDiscarded,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PacketsReceivedErrors,
			prometheus.CounterValue,
			nic.PacketsReceivedErrors,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PacketsReceivedTotal,
			prometheus.CounterValue,
			nic.PacketsReceivedPerSec,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PacketsReceivedUnknown,
			prometheus.CounterValue,
			nic.PacketsReceivedUnknown,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PacketsSentTotal,
			prometheus.CounterValue,
			nic.PacketsSentPerSec,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.CurrentBandwidth,
			prometheus.GaugeValue,
			nic.CurrentBandwidth/8,
			name,
		)
	}
	return nil, nil
}

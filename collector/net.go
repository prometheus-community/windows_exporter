// +build windows

package collector

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/StackExchange/wmi"
	"gopkg.in/alecthomas/kingpin.v2"
)

func init() {
	registerCollector("net", NewNetworkCollector, "Network Interface")
}

const (
	win32NetworkQuery = "SELECT Name, NetConnectionID FROM Win32_NetworkAdapter"
)

var (
	nicWhitelist = kingpin.Flag(
		"collector.net.nic-whitelist",
		"Regexp of NIC:s to whitelist. NIC name must both match whitelist and not match blacklist to be included.",
	).Default(".+").String()
	nicBlacklist = kingpin.Flag(
		"collector.net.nic-blacklist",
		"Regexp of NIC:s to blacklist. NIC name must both match whitelist and not match blacklist to be included.",
	).Default("").String()
	nicNameToUnderscore = regexp.MustCompile("[^a-zA-Z0-9]")
)

// A NetworkCollector is a Prometheus collector for Perflib Network Interface metrics
type NetworkCollector struct {
	BytesReceivedTotal       *prometheus.Desc
	BytesSentTotal           *prometheus.Desc
	BytesTotal               *prometheus.Desc
	PacketsOutboundDiscarded *prometheus.Desc
	PacketsOutboundErrors    *prometheus.Desc
	PacketsTotal             *prometheus.Desc
	PacketsReceivedDiscarded *prometheus.Desc
	PacketsReceivedErrors    *prometheus.Desc
	PacketsReceivedTotal     *prometheus.Desc
	PacketsReceivedUnknown   *prometheus.Desc
	PacketsSentTotal         *prometheus.Desc
	CurrentBandwidth         *prometheus.Desc

	nicWhitelistPattern *regexp.Regexp
	nicBlacklistPattern *regexp.Regexp
}

// NewNetworkCollector ...
func NewNetworkCollector() (Collector, error) {
	const subsystem = "net"

	return &NetworkCollector{
		BytesReceivedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bytes_received_total"),
			"(Network.BytesReceivedPerSec)",
			[]string{"nic", "netConnectionId"},
			nil,
		),
		BytesSentTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bytes_sent_total"),
			"(Network.BytesSentPerSec)",
			[]string{"nic", "netConnectionId"},
			nil,
		),
		BytesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bytes_total"),
			"(Network.BytesTotalPerSec)",
			[]string{"nic", "netConnectionId"},
			nil,
		),
		PacketsOutboundDiscarded: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "packets_outbound_discarded_total"),
			"(Network.PacketsOutboundDiscarded)",
			[]string{"nic", "netConnectionId"},
			nil,
		),
		PacketsOutboundErrors: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "packets_outbound_errors_total"),
			"(Network.PacketsOutboundErrors)",
			[]string{"nic", "netConnectionId"},
			nil,
		),
		PacketsReceivedDiscarded: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "packets_received_discarded_total"),
			"(Network.PacketsReceivedDiscarded)",
			[]string{"nic", "netConnectionId"},
			nil,
		),
		PacketsReceivedErrors: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "packets_received_errors_total"),
			"(Network.PacketsReceivedErrors)",
			[]string{"nic", "netConnectionId"},
			nil,
		),
		PacketsReceivedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "packets_received_total"),
			"(Network.PacketsReceivedPerSec)",
			[]string{"nic", "netConnectionId"},
			nil,
		),
		PacketsReceivedUnknown: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "packets_received_unknown_total"),
			"(Network.PacketsReceivedUnknown)",
			[]string{"nic", "netConnectionId"},
			nil,
		),
		PacketsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "packets_total"),
			"(Network.PacketsPerSec)",
			[]string{"nic", "netConnectionId"},
			nil,
		),
		PacketsSentTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "packets_sent_total"),
			"(Network.PacketsSentPerSec)",
			[]string{"nic", "netConnectionId"},
			nil,
		),
		CurrentBandwidth: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "current_bandwidth"),
			"(Network.CurrentBandwidth)",
			[]string{"nic", "netConnectionId"},
			nil,
		),

		nicWhitelistPattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *nicWhitelist)),
		nicBlacklistPattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *nicBlacklist)),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *NetworkCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ctx, ch); err != nil {
		log.Error("failed collecting net metrics:", desc, err)
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

// Win32_NetworkAdapter docs:
// - https://docs.microsoft.com/en-us/windows/win32/cimwin32prov/win32-networkadapter
type win32_NetworkAdapter struct {
	Name              string
	NetConnectionID   string
}

func (c *NetworkCollector) collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []networkInterface
	var networkDst []win32_NetworkAdapter

	// Make the netConnectionID map
	var netConnMap map[string]string
	netConnMap = make(map[string]string)

	if err := wmi.Query(win32NetworkQuery, &networkDst); err != nil {
		log.Error("Error running WMI query", err)
		return nil, err
	}

	if len(networkDst) == 0 {
		return nil, errors.New("WMI query for NetConnectionID returned empty result set")
	}

	// Map the NIC name to the NetConnectionID, mangled so it works fine
	for _, nic := range networkDst {
		netConnMap[mangleNetworkName(nic.Name)] = nic.NetConnectionID
	}

	if err := unmarshalObject(ctx.perfObjects["Network Interface"], &dst); err != nil {
		return nil, err
	}

	for _, nic := range dst {
		if c.nicBlacklistPattern.MatchString(nic.Name) ||
			!c.nicWhitelistPattern.MatchString(nic.Name) {
			continue
		}

		name := mangleNetworkName(nic.Name)
		if name == "" {
			continue
		}

		netConnectionId := mangleNetworkName(netConnMap[name])

		// Counters
		ch <- prometheus.MustNewConstMetric(
			c.BytesReceivedTotal,
			prometheus.CounterValue,
			nic.BytesReceivedPerSec,
			name,
			netConnectionId,
		)
		ch <- prometheus.MustNewConstMetric(
			c.BytesSentTotal,
			prometheus.CounterValue,
			nic.BytesSentPerSec,
			name,
			netConnectionId,
		)
		ch <- prometheus.MustNewConstMetric(
			c.BytesTotal,
			prometheus.CounterValue,
			nic.BytesTotalPerSec,
			name,
			netConnectionId,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PacketsOutboundDiscarded,
			prometheus.CounterValue,
			nic.PacketsOutboundDiscarded,
			name,
			netConnectionId,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PacketsOutboundErrors,
			prometheus.CounterValue,
			nic.PacketsOutboundErrors,
			name,
			netConnectionId,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PacketsTotal,
			prometheus.CounterValue,
			nic.PacketsPerSec,
			name,
			netConnectionId,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PacketsReceivedDiscarded,
			prometheus.CounterValue,
			nic.PacketsReceivedDiscarded,
			name,
			netConnectionId,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PacketsReceivedErrors,
			prometheus.CounterValue,
			nic.PacketsReceivedErrors,
			name,
			netConnectionId,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PacketsReceivedTotal,
			prometheus.CounterValue,
			nic.PacketsReceivedPerSec,
			name,
			netConnectionId,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PacketsReceivedUnknown,
			prometheus.CounterValue,
			nic.PacketsReceivedUnknown,
			name,
			netConnectionId,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PacketsSentTotal,
			prometheus.CounterValue,
			nic.PacketsSentPerSec,
			name,
			netConnectionId,
		)
		ch <- prometheus.MustNewConstMetric(
			c.CurrentBandwidth,
			prometheus.GaugeValue,
			nic.CurrentBandwidth,
			name,
			netConnectionId,
		)
	}
	return nil, nil
}

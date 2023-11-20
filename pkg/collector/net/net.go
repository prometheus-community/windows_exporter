//go:build windows

package net

import (
	"fmt"
	"regexp"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	Name = "net"

	FlagNicExclude = "collector.net.nic-exclude"
	FlagNicInclude = "collector.net.nic-include"
)

type Config struct {
	NicInclude string `yaml:"nic_include"`
	NicExclude string `yaml:"nic_exclude"`
}

var ConfigDefaults = Config{
	NicInclude: ".+",
	NicExclude: "",
}

var nicNameToUnderscore = regexp.MustCompile("[^a-zA-Z0-9]")

// A collector is a Prometheus collector for Perflib Network Interface metrics
type collector struct {
	logger log.Logger

	nicInclude *string
	nicExclude *string

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

func New(logger log.Logger, config *Config) types.Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &collector{
		nicExclude: &config.NicExclude,
		nicInclude: &config.NicInclude,
	}
	c.SetLogger(logger)
	return c
}

func NewWithFlags(app *kingpin.Application) types.Collector {
	c := &collector{
		nicInclude: app.Flag(
			FlagNicInclude,
			"Regexp of NIC:s to include. NIC name must both match include and not match exclude to be included.",
		).Default(ConfigDefaults.NicInclude).String(),

		nicExclude: app.Flag(
			FlagNicExclude,
			"Regexp of NIC:s to exclude. NIC name must both match include and not match exclude to be included.",
		).Default(ConfigDefaults.NicExclude).String(),
	}

	return c
}

func (c *collector) GetName() string {
	return Name
}

func (c *collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *collector) GetPerfCounter() ([]string, error) {
	return []string{"Network Interface"}, nil
}

func (c *collector) Build() error {
	c.BytesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bytes_received_total"),
		"(Network.BytesReceivedPerSec)",
		[]string{"nic"},
		nil,
	)
	c.BytesSentTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bytes_sent_total"),
		"(Network.BytesSentPerSec)",
		[]string{"nic"},
		nil,
	)
	c.BytesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bytes_total"),
		"(Network.BytesTotalPerSec)",
		[]string{"nic"},
		nil,
	)
	c.OutputQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "output_queue_length_packets"),
		"(Network.OutputQueueLength)",
		[]string{"nic"},
		nil,
	)
	c.PacketsOutboundDiscarded = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "packets_outbound_discarded_total"),
		"(Network.PacketsOutboundDiscarded)",
		[]string{"nic"},
		nil,
	)
	c.PacketsOutboundErrors = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "packets_outbound_errors_total"),
		"(Network.PacketsOutboundErrors)",
		[]string{"nic"},
		nil,
	)
	c.PacketsReceivedDiscarded = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "packets_received_discarded_total"),
		"(Network.PacketsReceivedDiscarded)",
		[]string{"nic"},
		nil,
	)
	c.PacketsReceivedErrors = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "packets_received_errors_total"),
		"(Network.PacketsReceivedErrors)",
		[]string{"nic"},
		nil,
	)
	c.PacketsReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "packets_received_total"),
		"(Network.PacketsReceivedPerSec)",
		[]string{"nic"},
		nil,
	)
	c.PacketsReceivedUnknown = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "packets_received_unknown_total"),
		"(Network.PacketsReceivedUnknown)",
		[]string{"nic"},
		nil,
	)
	c.PacketsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "packets_total"),
		"(Network.PacketsPerSec)",
		[]string{"nic"},
		nil,
	)
	c.PacketsSentTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "packets_sent_total"),
		"(Network.PacketsSentPerSec)",
		[]string{"nic"},
		nil,
	)
	c.CurrentBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_bandwidth_bytes"),
		"(Network.CurrentBandwidth)",
		[]string{"nic"},
		nil,
	)

	var err error
	c.nicIncludePattern, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", *c.nicInclude))
	if err != nil {
		return err
	}

	c.nicExcludePattern, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", *c.nicExclude))
	if err != nil {
		return err
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *collector) Collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
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

func (c *collector) collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []networkInterface

	if err := perflib.UnmarshalObject(ctx.PerfObjects["Network Interface"], &dst, c.logger); err != nil {
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

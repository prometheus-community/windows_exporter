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

const Name = "net"

type Config struct {
	NicExclude *regexp.Regexp `yaml:"nic_exclude"`
	NicInclude *regexp.Regexp `yaml:"nic_include"`
}

var ConfigDefaults = Config{
	NicExclude: types.RegExpEmpty,
	NicInclude: types.RegExpAny,
}

var nicNameToUnderscore = regexp.MustCompile("[^a-zA-Z0-9]")

// A Collector is a Prometheus Collector for Perflib Network Interface metrics.
type Collector struct {
	config Config

	bytesReceivedTotal       *prometheus.Desc
	bytesSentTotal           *prometheus.Desc
	bytesTotal               *prometheus.Desc
	outputQueueLength        *prometheus.Desc
	packetsOutboundDiscarded *prometheus.Desc
	packetsOutboundErrors    *prometheus.Desc
	packetsTotal             *prometheus.Desc
	packetsReceivedDiscarded *prometheus.Desc
	packetsReceivedErrors    *prometheus.Desc
	packetsReceivedTotal     *prometheus.Desc
	packetsReceivedUnknown   *prometheus.Desc
	packetsSentTotal         *prometheus.Desc
	currentBandwidth         *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	if config.NicExclude == nil {
		config.NicExclude = ConfigDefaults.NicExclude
	}

	if config.NicInclude == nil {
		config.NicInclude = ConfigDefaults.NicInclude
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

	var nicExclude, nicInclude string

	app.Flag(
		"collector.net.nic-exclude",
		"Regexp of NIC:s to exclude. NIC name must both match include and not match exclude to be included.",
	).Default(c.config.NicExclude.String()).StringVar(&nicExclude)

	app.Flag(
		"collector.net.nic-include",
		"Regexp of NIC:s to include. NIC name must both match include and not match exclude to be included.",
	).Default(c.config.NicInclude.String()).StringVar(&nicInclude)

	app.Action(func(*kingpin.ParseContext) error {
		var err error

		c.config.NicExclude, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", nicExclude))
		if err != nil {
			return fmt.Errorf("collector.net.nic-exclude: %w", err)
		}

		c.config.NicInclude, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", nicInclude))
		if err != nil {
			return fmt.Errorf("collector.net.nic-include: %w", err)
		}

		return nil
	})

	return c
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) GetPerfCounter(_ log.Logger) ([]string, error) {
	return []string{"Network Interface"}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(_ log.Logger) error {
	c.bytesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bytes_received_total"),
		"(Network.BytesReceivedPerSec)",
		[]string{"nic"},
		nil,
	)
	c.bytesSentTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bytes_sent_total"),
		"(Network.BytesSentPerSec)",
		[]string{"nic"},
		nil,
	)
	c.bytesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bytes_total"),
		"(Network.BytesTotalPerSec)",
		[]string{"nic"},
		nil,
	)
	c.outputQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "output_queue_length_packets"),
		"(Network.OutputQueueLength)",
		[]string{"nic"},
		nil,
	)
	c.packetsOutboundDiscarded = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "packets_outbound_discarded_total"),
		"(Network.PacketsOutboundDiscarded)",
		[]string{"nic"},
		nil,
	)
	c.packetsOutboundErrors = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "packets_outbound_errors_total"),
		"(Network.PacketsOutboundErrors)",
		[]string{"nic"},
		nil,
	)
	c.packetsReceivedDiscarded = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "packets_received_discarded_total"),
		"(Network.PacketsReceivedDiscarded)",
		[]string{"nic"},
		nil,
	)
	c.packetsReceivedErrors = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "packets_received_errors_total"),
		"(Network.PacketsReceivedErrors)",
		[]string{"nic"},
		nil,
	)
	c.packetsReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "packets_received_total"),
		"(Network.PacketsReceivedPerSec)",
		[]string{"nic"},
		nil,
	)
	c.packetsReceivedUnknown = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "packets_received_unknown_total"),
		"(Network.PacketsReceivedUnknown)",
		[]string{"nic"},
		nil,
	)
	c.packetsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "packets_total"),
		"(Network.PacketsPerSec)",
		[]string{"nic"},
		nil,
	)
	c.packetsSentTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "packets_sent_total"),
		"(Network.PacketsSentPerSec)",
		[]string{"nic"},
		nil,
	)
	c.currentBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_bandwidth_bytes"),
		"(Network.CurrentBandwidth)",
		[]string{"nic"},
		nil,
	)

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	if err := c.collect(ctx, logger, ch); err != nil {
		_ = level.Error(logger).Log("msg", "failed collecting net metrics", "err", err)
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

func (c *Collector) collect(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	var dst []networkInterface

	if err := perflib.UnmarshalObject(ctx.PerfObjects["Network Interface"], &dst, logger); err != nil {
		return err
	}

	for _, nic := range dst {
		if c.config.NicExclude.MatchString(nic.Name) ||
			!c.config.NicInclude.MatchString(nic.Name) {
			continue
		}

		name := mangleNetworkName(nic.Name)
		if name == "" {
			continue
		}

		// Counters
		ch <- prometheus.MustNewConstMetric(
			c.bytesReceivedTotal,
			prometheus.CounterValue,
			nic.BytesReceivedPerSec,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.bytesSentTotal,
			prometheus.CounterValue,
			nic.BytesSentPerSec,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.bytesTotal,
			prometheus.CounterValue,
			nic.BytesTotalPerSec,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.outputQueueLength,
			prometheus.GaugeValue,
			nic.OutputQueueLength,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.packetsOutboundDiscarded,
			prometheus.CounterValue,
			nic.PacketsOutboundDiscarded,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.packetsOutboundErrors,
			prometheus.CounterValue,
			nic.PacketsOutboundErrors,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.packetsTotal,
			prometheus.CounterValue,
			nic.PacketsPerSec,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.packetsReceivedDiscarded,
			prometheus.CounterValue,
			nic.PacketsReceivedDiscarded,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.packetsReceivedErrors,
			prometheus.CounterValue,
			nic.PacketsReceivedErrors,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.packetsReceivedTotal,
			prometheus.CounterValue,
			nic.PacketsReceivedPerSec,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.packetsReceivedUnknown,
			prometheus.CounterValue,
			nic.PacketsReceivedUnknown,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.packetsSentTotal,
			prometheus.CounterValue,
			nic.PacketsSentPerSec,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.currentBandwidth,
			prometheus.GaugeValue,
			nic.CurrentBandwidth/8,
			name,
		)
	}

	return nil
}

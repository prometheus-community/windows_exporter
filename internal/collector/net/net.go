//go:build windows

package net

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"slices"
	"strings"
	"unsafe"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/perfdata/registry"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
	"golang.org/x/sys/windows"
)

const Name = "net"

type Config struct {
	NicExclude        *regexp.Regexp `yaml:"nic_exclude"`
	NicInclude        *regexp.Regexp `yaml:"nic_include"`
	CollectorsEnabled []string       `yaml:"collectors_enabled"`
}

var ConfigDefaults = Config{
	NicExclude: types.RegExpEmpty,
	NicInclude: types.RegExpAny,
	CollectorsEnabled: []string{
		"metrics",
		"nic_addresses",
	},
}

// A Collector is a Prometheus Collector for Perflib Network Interface metrics.
type Collector struct {
	config Config

	perfDataCollector perfdata.Collector

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

	nicAddressInfo *prometheus.Desc
	routeInfo      *prometheus.Desc
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

	var nicExclude, nicInclude string

	var collectorsEnabled string

	app.Flag(
		"collector.net.nic-exclude",
		"Regexp of NIC:s to exclude. NIC name must both match include and not match exclude to be included.",
	).Default(c.config.NicExclude.String()).StringVar(&nicExclude)

	app.Flag(
		"collector.net.nic-include",
		"Regexp of NIC:s to include. NIC name must both match include and not match exclude to be included.",
	).Default(c.config.NicInclude.String()).StringVar(&nicInclude)

	app.Flag(
		"collector.net.enabled",
		"Comma-separated list of collectors to use. Defaults to all, if not specified.",
	).Default(strings.Join(ConfigDefaults.CollectorsEnabled, ",")).StringVar(&collectorsEnabled)

	app.Action(func(*kingpin.ParseContext) error {
		c.config.CollectorsEnabled = strings.Split(collectorsEnabled, ",")

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

func (c *Collector) GetPerfCounter(_ *slog.Logger) ([]string, error) {
	if utils.PDHEnabled() {
		return []string{}, nil
	}

	return []string{"Network Interface"}, nil
}

func (c *Collector) Close(_ *slog.Logger) error {
	return nil
}

func (c *Collector) Build(logger *slog.Logger, _ *wmi.Client) error {
	if utils.PDHEnabled() {
		counters := []string{
			BytesReceivedPerSec,
			BytesSentPerSec,
			BytesTotalPerSec,
			OutputQueueLength,
			PacketsOutboundDiscarded,
			PacketsOutboundErrors,
			PacketsPerSec,
			PacketsReceivedDiscarded,
			PacketsReceivedErrors,
			PacketsReceivedPerSec,
			PacketsReceivedUnknown,
			PacketsSentPerSec,
			CurrentBandwidth,
		}

		var err error

		c.perfDataCollector, err = perfdata.NewCollector(perfdata.Registry, "Network Interface", perfdata.AllInstances, counters)
		if err != nil {
			return fmt.Errorf("failed to create Processor Information collector: %w", err)
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, "addresses") {
		logger.Info("nic/addresses collector is in an experimental state! The configuration and metrics may change in future. Please report any issues.",
			slog.String("collector", Name),
		)
	}

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
	c.nicAddressInfo = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "nic_address_info"),
		"A metric with a constant '1' value labeled with the network interface's address information.",
		[]string{"nic", "friendly_name", "address", "family"},
		nil,
	)
	c.routeInfo = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "route_info"),
		"A metric with a constant '1' value labeled with the network interface's route information.",
		[]string{"nic", "src", "dest", "metric"},
		nil,
	)

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	logger = logger.With(slog.String("collector", Name))

	if slices.Contains(c.config.CollectorsEnabled, "metrics") {
		var err error

		if utils.PDHEnabled() {
			err = c.collectPDH(ch)
		} else {
			err = c.collect(ctx, logger, ch)
		}

		if err != nil {
			return fmt.Errorf("failed collecting net metrics: %w", err)
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, "nic_addresses") {
		if err := c.collectNICAddresses(ch); err != nil {
			return fmt.Errorf("failed collecting net addresses: %w", err)
		}
	}

	return nil
}

func (c *Collector) collect(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	var dst []perflibNetworkInterface

	if err := registry.UnmarshalObject(ctx.PerfObjects["Network Interface"], &dst, logger); err != nil {
		return err
	}

	for _, nic := range dst {
		if c.config.NicExclude.MatchString(nic.Name) ||
			!c.config.NicInclude.MatchString(nic.Name) {
			continue
		}

		// Counters
		ch <- prometheus.MustNewConstMetric(
			c.bytesReceivedTotal,
			prometheus.CounterValue,
			nic.BytesReceivedPerSec,
			nic.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.bytesSentTotal,
			prometheus.CounterValue,
			nic.BytesSentPerSec,
			nic.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.bytesTotal,
			prometheus.CounterValue,
			nic.BytesTotalPerSec,
			nic.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.outputQueueLength,
			prometheus.GaugeValue,
			nic.OutputQueueLength,
			nic.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.packetsOutboundDiscarded,
			prometheus.CounterValue,
			nic.PacketsOutboundDiscarded,
			nic.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.packetsOutboundErrors,
			prometheus.CounterValue,
			nic.PacketsOutboundErrors,
			nic.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.packetsTotal,
			prometheus.CounterValue,
			nic.PacketsPerSec,
			nic.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.packetsReceivedDiscarded,
			prometheus.CounterValue,
			nic.PacketsReceivedDiscarded,
			nic.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.packetsReceivedErrors,
			prometheus.CounterValue,
			nic.PacketsReceivedErrors,
			nic.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.packetsReceivedTotal,
			prometheus.CounterValue,
			nic.PacketsReceivedPerSec,
			nic.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.packetsReceivedUnknown,
			prometheus.CounterValue,
			nic.PacketsReceivedUnknown,
			nic.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.packetsSentTotal,
			prometheus.CounterValue,
			nic.PacketsSentPerSec,
			nic.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.currentBandwidth,
			prometheus.GaugeValue,
			nic.CurrentBandwidth/8,
			nic.Name,
		)
	}

	return nil
}

func (c *Collector) collectPDH(ch chan<- prometheus.Metric) error {
	data, err := c.perfDataCollector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect Network Information metrics: %w", err)
	}

	for nicName, nicData := range data {
		if c.config.NicExclude.MatchString(nicName) ||
			!c.config.NicInclude.MatchString(nicName) {
			continue
		}

		// Counters
		ch <- prometheus.MustNewConstMetric(
			c.bytesReceivedTotal,
			prometheus.CounterValue,
			nicData[BytesReceivedPerSec].FirstValue,
			nicName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.bytesSentTotal,
			prometheus.CounterValue,
			nicData[BytesSentPerSec].FirstValue,
			nicName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.bytesTotal,
			prometheus.CounterValue,
			nicData[BytesTotalPerSec].FirstValue,
			nicName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.outputQueueLength,
			prometheus.GaugeValue,
			nicData[OutputQueueLength].FirstValue,
			nicName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.packetsOutboundDiscarded,
			prometheus.CounterValue,
			nicData[PacketsOutboundDiscarded].FirstValue,
			nicName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.packetsOutboundErrors,
			prometheus.CounterValue,
			nicData[PacketsOutboundErrors].FirstValue,
			nicName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.packetsTotal,
			prometheus.CounterValue,
			nicData[PacketsPerSec].FirstValue,
			nicName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.packetsReceivedDiscarded,
			prometheus.CounterValue,
			nicData[PacketsReceivedDiscarded].FirstValue,
			nicName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.packetsReceivedErrors,
			prometheus.CounterValue,
			nicData[PacketsReceivedErrors].FirstValue,
			nicName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.packetsReceivedTotal,
			prometheus.CounterValue,
			nicData[PacketsReceivedPerSec].FirstValue,
			nicName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.packetsReceivedUnknown,
			prometheus.CounterValue,
			nicData[PacketsReceivedUnknown].FirstValue,
			nicName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.packetsSentTotal,
			prometheus.CounterValue,
			nicData[PacketsSentPerSec].FirstValue,
			nicName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.currentBandwidth,
			prometheus.GaugeValue,
			nicData[CurrentBandwidth].FirstValue/8,
			nicName,
		)
	}

	return nil
}

var addressFamily = map[uint16]string{
	windows.AF_INET:  "ipv4",
	windows.AF_INET6: "ipv6",
}

func (c *Collector) collectNICAddresses(ch chan<- prometheus.Metric) error {
	nicAdapterAddresses, err := adapterAddresses()
	if err != nil {
		return err
	}

	convertNicName := strings.NewReplacer("(", "[", ")", "]")

	for _, nicAdapterAddress := range nicAdapterAddresses {
		friendlyName := windows.UTF16PtrToString(nicAdapterAddress.FriendlyName)
		nicName := windows.UTF16PtrToString(nicAdapterAddress.Description)

		if c.config.NicExclude.MatchString(nicName) ||
			!c.config.NicInclude.MatchString(nicName) {
			continue
		}

		for address := nicAdapterAddress.FirstUnicastAddress; address != nil; address = address.Next {
			ipAddr := address.Address.IP()

			if ipAddr == nil || !ipAddr.IsGlobalUnicast() {
				continue
			}

			ch <- prometheus.MustNewConstMetric(
				c.nicAddressInfo,
				prometheus.GaugeValue,
				1,
				convertNicName.Replace(nicName),
				friendlyName,
				ipAddr.String(),
				addressFamily[address.Address.Sockaddr.Addr.Family],
			)
		}

		for address := nicAdapterAddress.FirstAnycastAddress; address != nil; address = address.Next {
			ipAddr := address.Address.IP()

			if ipAddr == nil || !ipAddr.IsGlobalUnicast() {
				continue
			}

			ch <- prometheus.MustNewConstMetric(
				c.nicAddressInfo,
				prometheus.GaugeValue,
				1,
				convertNicName.Replace(nicName),
				friendlyName,
				ipAddr.String(),
				addressFamily[address.Address.Sockaddr.Addr.Family],
			)
		}
	}

	return nil
}

// adapterAddresses returns a list of IP adapter and address
// structures. The structure contains an IP adapter and flattened
// multiple IP addresses including unicast, anycast and multicast
// addresses.
func adapterAddresses() ([]*windows.IpAdapterAddresses, error) {
	var b []byte

	l := uint32(15000) // recommended initial size

	for {
		b = make([]byte, l)

		const flags = windows.GAA_FLAG_SKIP_MULTICAST | windows.GAA_FLAG_SKIP_DNS_SERVER

		err := windows.GetAdaptersAddresses(windows.AF_UNSPEC, flags, 0, (*windows.IpAdapterAddresses)(unsafe.Pointer(&b[0])), &l)
		if err == nil {
			if l == 0 {
				return nil, nil
			}

			break
		}

		if !errors.Is(err, windows.ERROR_BUFFER_OVERFLOW) {
			return nil, os.NewSyscallError("getadaptersaddresses", err)
		}

		if l <= uint32(len(b)) {
			return nil, os.NewSyscallError("getadaptersaddresses", err)
		}
	}

	var addresses []*windows.IpAdapterAddresses
	for address := (*windows.IpAdapterAddresses)(unsafe.Pointer(&b[0])); address != nil; address = address.Next {
		addresses = append(addresses, address)
	}

	return addresses, nil
}

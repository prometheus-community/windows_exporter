// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	"time"
	"unsafe"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/labelcache"
	"github.com/prometheus-community/windows_exporter/internal/metriccache"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows"
)

const (
	Name = "net"

	subCollectorMetrics = "metrics"
	subCollectorNicInfo = "nic_info"
)

type Config struct {
	NicExclude        *regexp.Regexp `yaml:"nic-exclude"`
	NicInclude        *regexp.Regexp `yaml:"nic-include"`
	CollectorsEnabled []string       `yaml:"enabled"`
}

//nolint:gochecknoglobals
var ConfigDefaults = Config{
	NicExclude: types.RegExpEmpty,
	NicInclude: types.RegExpAny,
	CollectorsEnabled: []string{
		subCollectorMetrics,
		subCollectorNicInfo,
	},
}

// nicNameReplacer converts an adapter description into the instance name of
// the adapter in the Network Interface performance counter object.
//
//nolint:gochecknoglobals
var nicNameReplacer = strings.NewReplacer("(", "[", ")", "]", "#", "_", "/", "_", "\\", "_")

// A Collector is a Prometheus Collector for Perflib Network Interface metrics.
type Collector struct {
	config Config

	perfDataCollector *pdh.Collector[perfDataCounterValues]
	perfDataObject    []perfDataCounterValues

	// labels keeps the label pairs of the Network Interface series of the last
	// scrape. Their values change on every scrape, but their labels stay the
	// same while the adapters do.
	labels labelcache.Cache

	// adapterAddressesBuf holds the result of GetAdaptersAddresses. It is
	// reused between scrapes, so its size also seeds the next call.
	adapterAddressesBuf []byte

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

	nicIPAddressInfo *prometheus.Desc
	nicOperStatus    *prometheus.Desc
	nicInfo          *prometheus.Desc

	// nicMetricCache reuses the nic_info, nic_operation_status and
	// nic_address_info metrics of adapters whose information did not change
	// since the last scrape. Each adapter publishes at least 8 of these series.
	// The adapter addresses are still queried on every scrape, so it saves
	// allocations, not system calls.
	nicMetricCache metriccache.Cache[string, nicValues]
}

// nicValues holds the inputs of the nic_info metrics of an adapter besides its
// name.
type nicValues struct {
	friendlyName string
	macAddress   string
	operStatus   uint32
	// addresses encodes the global unicast and anycast addresses of an adapter
	// that is up, in emission order. Each address is its length, its bytes and
	// its address family, so the encoding is unambiguous.
	addresses string
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
	).Default("").StringVar(&nicExclude)

	app.Flag(
		"collector.net.nic-include",
		"Regexp of NIC:s to include. NIC name must both match include and not match exclude to be included.",
	).Default(".+").StringVar(&nicInclude)

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

func (c *Collector) Close() error {
	c.perfDataCollector.Close()
	c.nicMetricCache.Reset()
	c.labels.Reset()
	c.adapterAddressesBuf = nil

	return nil
}

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	for _, collector := range c.config.CollectorsEnabled {
		if !slices.Contains([]string{subCollectorMetrics, subCollectorNicInfo}, collector) {
			return fmt.Errorf("unknown sub collector: %s. Possible values: %s", collector,
				strings.Join([]string{subCollectorMetrics, subCollectorNicInfo}, ", "),
			)
		}
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
	c.nicIPAddressInfo = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "nic_address_info"),
		"A metric with a constant '1' value labeled with the network interface's address information.",
		[]string{"nic", "address", "family"},
		nil,
	)
	c.nicOperStatus = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "nic_operation_status"),
		"The operational status for the interface as defined in RFC 2863 as IfOperStatus.",
		[]string{"nic", "status"},
		nil,
	)
	c.nicInfo = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "nic_info"),
		"A metric with a constant '1' value labeled with the network interface's general information.",
		[]string{"nic", "friendly_name", "mac"},
		nil,
	)

	// Cached metrics refer to the descriptors of an earlier Build.
	c.nicMetricCache.Reset()
	c.labels.Reset()

	var err error

	c.perfDataCollector, err = pdh.NewCollector[perfDataCounterValues](logger.With(slog.String("collector", Name)), pdh.CounterTypeRaw, "Network Interface", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create Network Interface collector: %w", err)
	}

	if slices.Contains(c.config.CollectorsEnabled, subCollectorNicInfo) {
		logger.Info("nic/addresses collector is in an experimental state! The configuration and metrics may change in future. Please report any issues.",
			slog.String("collector", Name),
		)
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric, _ time.Duration) error {
	errs := make([]error, 0)

	if slices.Contains(c.config.CollectorsEnabled, subCollectorMetrics) {
		if err := c.collect(ch); err != nil {
			errs = append(errs, fmt.Errorf("failed collecting metrics: %w", err))
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, subCollectorNicInfo) {
		if err := c.collectNICInfo(ch); err != nil {
			errs = append(errs, fmt.Errorf("failed collecting net addresses: %w", err))
		}
	}

	return errors.Join(errs...)
}

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollector.Collect(&c.perfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect Network Information metrics: %w", err)
	}

	scrape := c.labels.Begin()
	defer scrape.Commit()

	for _, data := range c.perfDataObject {
		if c.config.NicExclude.MatchString(data.Name) || !c.config.NicInclude.MatchString(data.Name) {
			continue
		}

		// Counters
		ch <- scrape.MustNewConstMetric(
			c.bytesReceivedTotal,
			prometheus.CounterValue,
			data.BytesReceivedPerSec,
			data.Name,
		)

		ch <- scrape.MustNewConstMetric(
			c.bytesSentTotal,
			prometheus.CounterValue,
			data.BytesSentPerSec,
			data.Name,
		)

		ch <- scrape.MustNewConstMetric(
			c.bytesTotal,
			prometheus.CounterValue,
			data.BytesTotalPerSec,
			data.Name,
		)

		ch <- scrape.MustNewConstMetric(
			c.outputQueueLength,
			prometheus.GaugeValue,
			data.OutputQueueLength,
			data.Name,
		)

		ch <- scrape.MustNewConstMetric(
			c.packetsOutboundDiscarded,
			prometheus.CounterValue,
			data.PacketsOutboundDiscarded,
			data.Name,
		)

		ch <- scrape.MustNewConstMetric(
			c.packetsOutboundErrors,
			prometheus.CounterValue,
			data.PacketsOutboundErrors,
			data.Name,
		)

		ch <- scrape.MustNewConstMetric(
			c.packetsTotal,
			prometheus.CounterValue,
			data.PacketsPerSec,
			data.Name,
		)

		ch <- scrape.MustNewConstMetric(
			c.packetsReceivedDiscarded,
			prometheus.CounterValue,
			data.PacketsReceivedDiscarded,
			data.Name,
		)

		ch <- scrape.MustNewConstMetric(
			c.packetsReceivedErrors,
			prometheus.CounterValue,
			data.PacketsReceivedErrors,
			data.Name,
		)

		ch <- scrape.MustNewConstMetric(
			c.packetsReceivedTotal,
			prometheus.CounterValue,
			data.PacketsReceivedPerSec,
			data.Name,
		)

		ch <- scrape.MustNewConstMetric(
			c.packetsReceivedUnknown,
			prometheus.CounterValue,
			data.PacketsReceivedUnknown,
			data.Name,
		)

		ch <- scrape.MustNewConstMetric(
			c.packetsSentTotal,
			prometheus.CounterValue,
			data.PacketsSentPerSec,
			data.Name,
		)

		ch <- scrape.MustNewConstMetric(
			c.currentBandwidth,
			prometheus.GaugeValue,
			data.CurrentBandwidth/8,
			data.Name,
		)
	}

	return nil
}

func (c *Collector) collectNICInfo(ch chan<- prometheus.Metric) error {
	var (
		nicAdapters *windows.IpAdapterAddresses
		err         error
	)

	nicAdapters, c.adapterAddressesBuf, err = adapterAddresses(c.adapterAddressesBuf)
	if err != nil {
		return err
	}

	scrape := c.nicMetricCache.Begin()

	var addresses []byte

	for nicAdapter := nicAdapters; nicAdapter != nil; nicAdapter = nicAdapter.Next {
		friendlyName := windows.UTF16PtrToString(nicAdapter.FriendlyName)
		nicName := nicNameReplacer.Replace(windows.UTF16PtrToString(nicAdapter.Description))

		if c.config.NicExclude.MatchString(nicName) ||
			!c.config.NicInclude.MatchString(nicName) {
			continue
		}

		macAddress := formatPhysicalAddress(nicAdapter)

		addresses = addresses[:0]

		if nicAdapter.OperStatus == windows.IfOperStatusUp {
			addresses = appendNICAddresses(addresses, nicAdapter)
		}

		values := nicValues{
			friendlyName: friendlyName,
			macAddress:   macAddress,
			operStatus:   nicAdapter.OperStatus,
			addresses:    string(addresses),
		}

		metrics := scrape.Get(nicName, values, func() []prometheus.Metric {
			return c.nicMetrics(nicName, nicAdapter, values)
		})

		for _, metric := range metrics {
			ch <- metric
		}
	}

	scrape.Commit()

	return nil
}

// appendNICAddresses appends the encoding of the global unicast and anycast
// addresses of nicAdapter to b, in the order nicMetrics emits them.
func appendNICAddresses(b []byte, nicAdapter *windows.IpAdapterAddresses) []byte {
	appendAddress := func(address *windows.SocketAddress) {
		ipAddr := address.IP()

		if ipAddr == nil || !ipAddr.IsGlobalUnicast() {
			return
		}

		family := address.Sockaddr.Addr.Family

		b = append(b, byte(len(ipAddr)))
		b = append(b, ipAddr...)
		b = append(b, byte(family>>8), byte(family))
	}

	for address := nicAdapter.FirstUnicastAddress; address != nil; address = address.Next {
		appendAddress(&address.Address)
	}

	for address := nicAdapter.FirstAnycastAddress; address != nil; address = address.Next {
		appendAddress(&address.Address)
	}

	return b
}

// nicMetrics builds the nic_info, nic_operation_status and nic_address_info
// metrics of an adapter.
func (c *Collector) nicMetrics(nicName string, nicAdapter *windows.IpAdapterAddresses, values nicValues) []prometheus.Metric {
	metrics := make([]prometheus.Metric, 0, 1+len(operStatus))

	metrics = append(metrics, prometheus.MustNewConstMetric(
		c.nicInfo,
		prometheus.GaugeValue,
		1,
		nicName,
		values.friendlyName,
		values.macAddress,
	))

	for operState, labelValue := range operStatus {
		var metricStatus float64
		if operState == values.operStatus {
			metricStatus = 1
		}

		metrics = append(metrics, prometheus.MustNewConstMetric(
			c.nicOperStatus,
			prometheus.GaugeValue,
			metricStatus,
			nicName,
			labelValue,
		))
	}

	if values.operStatus != windows.IfOperStatusUp {
		return metrics
	}

	for address := nicAdapter.FirstUnicastAddress; address != nil; address = address.Next {
		ipAddr := address.Address.IP()

		if ipAddr == nil || !ipAddr.IsGlobalUnicast() {
			continue
		}

		metrics = append(metrics, prometheus.MustNewConstMetric(
			c.nicIPAddressInfo,
			prometheus.GaugeValue,
			1,
			nicName,
			ipAddr.String(),
			addressFamily[address.Address.Sockaddr.Addr.Family],
		))
	}

	for address := nicAdapter.FirstAnycastAddress; address != nil; address = address.Next {
		ipAddr := address.Address.IP()

		if ipAddr == nil || !ipAddr.IsGlobalUnicast() {
			continue
		}

		metrics = append(metrics, prometheus.MustNewConstMetric(
			c.nicIPAddressInfo,
			prometheus.GaugeValue,
			1,
			nicName,
			ipAddr.String(),
			addressFamily[address.Address.Sockaddr.Addr.Family],
		))
	}

	return metrics
}

// formatPhysicalAddress returns the physical address of nicAdapter as
// colon-separated upper case hex bytes. It is empty for adapters without one,
// like the loopback adapter.
func formatPhysicalAddress(nicAdapter *windows.IpAdapterAddresses) string {
	const hexDigits = "0123456789ABCDEF"

	n := min(int(nicAdapter.PhysicalAddressLength), len(nicAdapter.PhysicalAddress))
	if n == 0 {
		return ""
	}

	b := make([]byte, 0, 3*n-1)

	for i, octet := range nicAdapter.PhysicalAddress[:n] {
		if i > 0 {
			b = append(b, ':')
		}

		b = append(b, hexDigits[octet>>4], hexDigits[octet&0x0F])
	}

	return string(b)
}

// adapterAddresses returns the linked list of IP adapter addresses of the
// system, without multicast and DNS server addresses. The list lives in the
// returned buffer, which the caller passes to the next call for reuse. The
// list is nil if the system has no adapters.
func adapterAddresses(b []byte) (*windows.IpAdapterAddresses, []byte, error) {
	const flags = windows.GAA_FLAG_SKIP_MULTICAST | windows.GAA_FLAG_SKIP_DNS_SERVER

	l := uint32(max(len(b), 15000)) // 15000 is the recommended initial size.

	for {
		if len(b) < int(l) {
			b = make([]byte, l)
		}

		l = uint32(len(b))

		err := windows.GetAdaptersAddresses(windows.AF_UNSPEC, flags, 0, (*windows.IpAdapterAddresses)(unsafe.Pointer(&b[0])), &l)
		if err == nil {
			return (*windows.IpAdapterAddresses)(unsafe.Pointer(&b[0])), b, nil
		}

		if errors.Is(err, windows.ERROR_NO_DATA) {
			return nil, b, nil
		}

		if !errors.Is(err, windows.ERROR_BUFFER_OVERFLOW) || l <= uint32(len(b)) {
			return nil, b, os.NewSyscallError("getadaptersaddresses", err)
		}
	}
}

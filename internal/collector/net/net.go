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
	"unsafe"

	"github.com/alecthomas/kingpin/v2"
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

// A Collector is a Prometheus Collector for Perflib Network Interface metrics.
type Collector struct {
	config Config

	perfDataCollector *pdh.Collector
	perfDataObject    []perfDataCounterValues

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
	routeInfo        *prometheus.Desc
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
	c.routeInfo = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "route_info"),
		"A metric with a constant '1' value labeled with the network interface's route information.",
		[]string{"nic", "src", "dest", "metric"},
		nil,
	)

	var err error

	c.perfDataCollector, err = pdh.NewCollector[perfDataCounterValues](pdh.CounterTypeRaw, "Network Interface", pdh.InstancesAll)
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
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
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

	for _, data := range c.perfDataObject {
		if c.config.NicExclude.MatchString(data.Name) || !c.config.NicInclude.MatchString(data.Name) {
			continue
		}

		// Counters
		ch <- prometheus.MustNewConstMetric(
			c.bytesReceivedTotal,
			prometheus.CounterValue,
			data.BytesReceivedPerSec,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bytesSentTotal,
			prometheus.CounterValue,
			data.BytesSentPerSec,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bytesTotal,
			prometheus.CounterValue,
			data.BytesTotalPerSec,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.outputQueueLength,
			prometheus.GaugeValue,
			data.OutputQueueLength,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.packetsOutboundDiscarded,
			prometheus.CounterValue,
			data.PacketsOutboundDiscarded,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.packetsOutboundErrors,
			prometheus.CounterValue,
			data.PacketsOutboundErrors,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.packetsTotal,
			prometheus.CounterValue,
			data.PacketsPerSec,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.packetsReceivedDiscarded,
			prometheus.CounterValue,
			data.PacketsReceivedDiscarded,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.packetsReceivedErrors,
			prometheus.CounterValue,
			data.PacketsReceivedErrors,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.packetsReceivedTotal,
			prometheus.CounterValue,
			data.PacketsReceivedPerSec,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.packetsReceivedUnknown,
			prometheus.CounterValue,
			data.PacketsReceivedUnknown,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.packetsSentTotal,
			prometheus.CounterValue,
			data.PacketsSentPerSec,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.currentBandwidth,
			prometheus.GaugeValue,
			data.CurrentBandwidth/8,
			data.Name,
		)
	}

	return nil
}

func (c *Collector) collectNICInfo(ch chan<- prometheus.Metric) error {
	nicAdapterAddresses, err := adapterAddresses()
	if err != nil {
		return err
	}

	convertNicName := strings.NewReplacer("(", "[", ")", "]", "#", "_")

	for _, nicAdapter := range nicAdapterAddresses {
		friendlyName := windows.UTF16PtrToString(nicAdapter.FriendlyName)
		nicName := convertNicName.Replace(windows.UTF16PtrToString(nicAdapter.Description))

		if c.config.NicExclude.MatchString(nicName) ||
			!c.config.NicInclude.MatchString(nicName) {
			continue
		}

		macAddress := fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X",
			nicAdapter.PhysicalAddress[0],
			nicAdapter.PhysicalAddress[1],
			nicAdapter.PhysicalAddress[2],
			nicAdapter.PhysicalAddress[3],
			nicAdapter.PhysicalAddress[4],
			nicAdapter.PhysicalAddress[5],
		)

		ch <- prometheus.MustNewConstMetric(
			c.nicInfo,
			prometheus.GaugeValue,
			1,
			nicName,
			friendlyName,
			macAddress,
		)

		for operState, labelValue := range operStatus {
			var metricStatus float64
			if operState == nicAdapter.OperStatus {
				metricStatus = 1
			}

			ch <- prometheus.MustNewConstMetric(
				c.nicOperStatus,
				prometheus.GaugeValue,
				metricStatus,
				nicName,
				labelValue,
			)
		}

		if nicAdapter.OperStatus != windows.IfOperStatusUp {
			continue
		}

		for address := nicAdapter.FirstUnicastAddress; address != nil; address = address.Next {
			ipAddr := address.Address.IP()

			if ipAddr == nil || !ipAddr.IsGlobalUnicast() {
				continue
			}

			ch <- prometheus.MustNewConstMetric(
				c.nicIPAddressInfo,
				prometheus.GaugeValue,
				1,
				nicName,
				ipAddr.String(),
				addressFamily[address.Address.Sockaddr.Addr.Family],
			)
		}

		for address := nicAdapter.FirstAnycastAddress; address != nil; address = address.Next {
			ipAddr := address.Address.IP()

			if ipAddr == nil || !ipAddr.IsGlobalUnicast() {
				continue
			}

			ch <- prometheus.MustNewConstMetric(
				c.nicIPAddressInfo,
				prometheus.GaugeValue,
				1,
				nicName,
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

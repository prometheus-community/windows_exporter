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

package tcp

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

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
)

type Config struct {
	CollectorsEnabled []string `yaml:"enabled"`
}

//nolint:gochecknoglobals
var ConfigDefaults = Config{
	CollectorsEnabled: []string{
		subCollectorMetrics,
		subCollectorConnectionsState,
	},
}

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_Tcpip_TCPv{4,6} metrics.
type Collector struct {
	config Config

	perfDataCollector4 *pdh.Collector
	perfDataCollector6 *pdh.Collector
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
	}

	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *mi.Session) error {
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
		prometheus.BuildFQName(types.Namespace, Name, "connections_state_count"),
		"Number of TCP connections by state and address family",
		[]string{"af", "state"},
		nil,
	)

	errs := make([]error, 0)

	if slices.Contains(c.config.CollectorsEnabled, subCollectorMetrics) {
		var err error

		c.perfDataCollector4, err = pdh.NewCollector[perfDataCounterValues](pdh.CounterTypeRaw, "TCPv4", nil)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to create TCPv4 collector: %w", err))
		}

		c.perfDataCollector6, err = pdh.NewCollector[perfDataCounterValues](pdh.CounterTypeRaw, "TCPv6", nil)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to create TCPv6 collector: %w", err))
		}
	}

	return errors.Join(errs...)
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	errs := make([]error, 0)

	if slices.Contains(c.config.CollectorsEnabled, subCollectorMetrics) {
		if err := c.collect(ch); err != nil {
			errs = append(errs, fmt.Errorf("failed collecting tcp metrics: %w", err))
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, subCollectorConnectionsState) {
		if err := c.collectConnectionsState(ch); err != nil {
			errs = append(errs, fmt.Errorf("failed collecting tcp connection state metrics: %w", err))
		}
	}

	return errors.Join(errs...)
}

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	errs := make([]error, 0)

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

func (c *Collector) collectConnectionsState(ch chan<- prometheus.Metric) error {
	errs := make([]error, 0)

	if stateCounts, err := iphlpapi.GetTCPConnectionStates(windows.AF_INET); err != nil {
		errs = append(errs, fmt.Errorf("failed to collect TCP connection states for %s: %w", ipAddressFamilyIPv4, err))
	} else {
		c.sendTCPStateMetrics(ch, stateCounts, ipAddressFamilyIPv4)
	}

	if stateCounts, err := iphlpapi.GetTCPConnectionStates(windows.AF_INET6); err != nil {
		errs = append(errs, fmt.Errorf("failed to collect TCP6 connection states for %s: %w", ipAddressFamilyIPv6, err))
	} else {
		c.sendTCPStateMetrics(ch, stateCounts, ipAddressFamilyIPv6)
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

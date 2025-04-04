// Copyright 2024 The Prometheus Authors
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

package udp

import (
	"fmt"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "udp"

type Config struct{}

//nolint:gochecknoglobals
var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_Tcpip_TCPv{4,6} metrics.
type Collector struct {
	config Config

	perfDataCollector4 *pdh.Collector
	perfDataCollector6 *pdh.Collector
	perfDataObject4    []perfDataCounterValues
	perfDataObject6    []perfDataCounterValues

	datagramsNoPortTotal         *prometheus.Desc
	datagramsReceivedTotal       *prometheus.Desc
	datagramsReceivedErrorsTotal *prometheus.Desc
	datagramsSentTotal           *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	c := &Collector{
		config: ConfigDefaults,
	}

	return c
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	c.perfDataCollector4.Close()
	c.perfDataCollector6.Close()

	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *mi.Session) error {
	c.datagramsNoPortTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datagram_no_port_total"),
		"Number of received UDP datagrams for which there was no application at the destination port",
		[]string{"af"},
		nil,
	)
	c.datagramsReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datagram_received_total"),
		"UDP datagrams are delivered to UDP users",
		[]string{"af"},
		nil,
	)
	c.datagramsReceivedErrorsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datagram_received_errors_total"),
		"Number of received UDP datagrams that could not be delivered for reasons other than the lack of an application at the destination port",
		[]string{"af"},
		nil,
	)
	c.datagramsSentTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datagram_sent_total"),
		"UDP datagrams are sent from the entity",
		[]string{"af"},
		nil,
	)

	var err error

	c.perfDataCollector4, err = pdh.NewCollector[perfDataCounterValues](pdh.CounterTypeRaw, "UDPv4", nil)
	if err != nil {
		return fmt.Errorf("failed to create UDPv4 collector: %w", err)
	}

	c.perfDataCollector6, err = pdh.NewCollector[perfDataCounterValues](pdh.CounterTypeRaw, "UDPv6", nil)
	if err != nil {
		return fmt.Errorf("failed to create UDPv6 collector: %w", err)
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	return c.collect(ch)
}

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollector4.Collect(&c.perfDataObject4)
	if err != nil {
		return fmt.Errorf("failed to collect UDPv4 metrics: %w", err)
	}

	c.writeUDPCounters(ch, c.perfDataObject4, []string{"ipv4"})

	err = c.perfDataCollector6.Collect(&c.perfDataObject6)
	if err != nil {
		return fmt.Errorf("failed to collect UDPv6 metrics: %w", err)
	}

	c.writeUDPCounters(ch, c.perfDataObject6, []string{"ipv6"})

	return nil
}

func (c *Collector) writeUDPCounters(ch chan<- prometheus.Metric, metrics []perfDataCounterValues, labels []string) {
	ch <- prometheus.MustNewConstMetric(
		c.datagramsNoPortTotal,
		prometheus.CounterValue,
		metrics[0].DatagramsNoPortPerSec,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.datagramsReceivedErrorsTotal,
		prometheus.CounterValue,
		metrics[0].DatagramsReceivedErrors,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.datagramsReceivedTotal,
		prometheus.GaugeValue,
		metrics[0].DatagramsReceivedPerSec,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.datagramsSentTotal,
		prometheus.CounterValue,
		metrics[0].DatagramsSentPerSec,
		labels...,
	)
}

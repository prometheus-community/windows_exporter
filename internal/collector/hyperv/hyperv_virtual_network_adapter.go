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

package hyperv

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

// collectorVirtualNetworkAdapter Hyper-V Virtual Network Adapter metrics
type collectorVirtualNetworkAdapter struct {
	perfDataCollectorVirtualNetworkAdapter *pdh.Collector
	perfDataObjectVirtualNetworkAdapter    []perfDataCounterValuesVirtualNetworkAdapter

	virtualNetworkAdapterBytesReceived          *prometheus.Desc // \Hyper-V Virtual Network Adapter(*)\Bytes Received/sec
	virtualNetworkAdapterBytesSent              *prometheus.Desc // \Hyper-V Virtual Network Adapter(*)\Bytes Sent/sec
	virtualNetworkAdapterDroppedPacketsIncoming *prometheus.Desc // \Hyper-V Virtual Network Adapter(*)\Dropped Packets Incoming/sec
	virtualNetworkAdapterDroppedPacketsOutgoing *prometheus.Desc // \Hyper-V Virtual Network Adapter(*)\Dropped Packets Outgoing/sec
	virtualNetworkAdapterPacketsReceived        *prometheus.Desc // \Hyper-V Virtual Network Adapter(*)\Packets Received/sec
	virtualNetworkAdapterPacketsSent            *prometheus.Desc // \Hyper-V Virtual Network Adapter(*)\Packets Sent/sec
}

type perfDataCounterValuesVirtualNetworkAdapter struct {
	Name string

	VirtualNetworkAdapterBytesReceived          float64 `perfdata:"Bytes Received/sec"`
	VirtualNetworkAdapterBytesSent              float64 `perfdata:"Bytes Sent/sec"`
	VirtualNetworkAdapterDroppedPacketsIncoming float64 `perfdata:"Dropped Packets Incoming/sec"`
	VirtualNetworkAdapterDroppedPacketsOutgoing float64 `perfdata:"Dropped Packets Outgoing/sec"`
	VirtualNetworkAdapterPacketsReceived        float64 `perfdata:"Packets Received/sec"`
	VirtualNetworkAdapterPacketsSent            float64 `perfdata:"Packets Sent/sec"`
}

func (c *Collector) buildVirtualNetworkAdapter() error {
	var err error

	c.perfDataCollectorVirtualNetworkAdapter, err = pdh.NewCollector[perfDataCounterValuesVirtualNetworkAdapter](pdh.CounterTypeRaw, "Hyper-V Virtual Network Adapter", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create Hyper-V Virtual Network Adapter collector: %w", err)
	}

	c.virtualNetworkAdapterBytesReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_network_adapter_received_bytes_total"),
		"Represents the total number of bytes received per second by the network adapter",
		[]string{"adapter"},
		nil,
	)
	c.virtualNetworkAdapterBytesSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_network_adapter_sent_bytes_total"),
		"Represents the total number of bytes sent per second by the network adapter",
		[]string{"adapter"},
		nil,
	)
	c.virtualNetworkAdapterDroppedPacketsIncoming = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_network_adapter_incoming_dropped_packets_total"),
		"Represents the total number of dropped packets per second in the incoming direction of the network adapter",
		[]string{"adapter"},
		nil,
	)
	c.virtualNetworkAdapterDroppedPacketsOutgoing = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_network_adapter_outgoing_dropped_packets_total"),
		"Represents the total number of dropped packets per second in the outgoing direction of the network adapter",
		[]string{"adapter"},
		nil,
	)
	c.virtualNetworkAdapterPacketsReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_network_adapter_received_packets_total"),
		"Represents the total number of packets received per second by the network adapter",
		[]string{"adapter"},
		nil,
	)
	c.virtualNetworkAdapterPacketsSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_network_adapter_sent_packets_total"),
		"Represents the total number of packets sent per second by the network adapter",
		[]string{"adapter"},
		nil,
	)

	return nil
}

func (c *Collector) collectVirtualNetworkAdapter(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollectorVirtualNetworkAdapter.Collect(&c.perfDataObjectVirtualNetworkAdapter)
	if err != nil {
		return fmt.Errorf("failed to collect Hyper-V Virtual Network Adapter metrics: %w", err)
	}

	for _, data := range c.perfDataObjectVirtualNetworkAdapter {
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterBytesReceived,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterBytesReceived,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterBytesSent,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterBytesSent,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDroppedPacketsIncoming,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDroppedPacketsIncoming,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDroppedPacketsOutgoing,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDroppedPacketsOutgoing,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterPacketsReceived,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterPacketsReceived,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterPacketsSent,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterPacketsSent,
			data.Name,
		)
	}

	return nil
}

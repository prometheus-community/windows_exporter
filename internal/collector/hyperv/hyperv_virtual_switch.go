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

// collectorVirtualMachineHealthSummary Hyper-V Virtual Switch Summary metrics
type collectorVirtualSwitch struct {
	perfDataCollectorVirtualSwitch *pdh.Collector
	perfDataObjectVirtualSwitch    []perfDataCounterValuesVirtualSwitch

	virtualSwitchBroadcastPacketsReceived         *prometheus.Desc // \Hyper-V Virtual Switch(*)\Broadcast Packets Received/sec
	virtualSwitchBroadcastPacketsSent             *prometheus.Desc // \Hyper-V Virtual Switch(*)\Broadcast Packets Sent/sec
	virtualSwitchBytes                            *prometheus.Desc // \Hyper-V Virtual Switch(*)\Bytes/sec
	virtualSwitchBytesReceived                    *prometheus.Desc // \Hyper-V Virtual Switch(*)\Bytes Received/sec
	virtualSwitchBytesSent                        *prometheus.Desc // \Hyper-V Virtual Switch(*)\Bytes Sent/sec
	virtualSwitchDirectedPacketsReceived          *prometheus.Desc // \Hyper-V Virtual Switch(*)\Directed Packets Received/sec
	virtualSwitchDirectedPacketsSent              *prometheus.Desc // \Hyper-V Virtual Switch(*)\Directed Packets Sent/sec
	virtualSwitchDroppedPacketsIncoming           *prometheus.Desc // \Hyper-V Virtual Switch(*)\Dropped Packets Incoming/sec
	virtualSwitchDroppedPacketsOutgoing           *prometheus.Desc // \Hyper-V Virtual Switch(*)\Dropped Packets Outgoing/sec
	virtualSwitchExtensionsDroppedPacketsIncoming *prometheus.Desc // \Hyper-V Virtual Switch(*)\Extensions Dropped Packets Incoming/sec
	virtualSwitchExtensionsDroppedPacketsOutgoing *prometheus.Desc // \Hyper-V Virtual Switch(*)\Extensions Dropped Packets Outgoing/sec
	virtualSwitchLearnedMacAddresses              *prometheus.Desc // \Hyper-V Virtual Switch(*)\Learned Mac Addresses
	virtualSwitchMulticastPacketsReceived         *prometheus.Desc // \Hyper-V Virtual Switch(*)\Multicast Packets Received/sec
	virtualSwitchMulticastPacketsSent             *prometheus.Desc // \Hyper-V Virtual Switch(*)\Multicast Packets Sent/sec
	virtualSwitchNumberOfSendChannelMoves         *prometheus.Desc // \Hyper-V Virtual Switch(*)\Number of Send Channel Moves/sec
	virtualSwitchNumberOfVMQMoves                 *prometheus.Desc // \Hyper-V Virtual Switch(*)\Number of VMQ Moves/sec
	virtualSwitchPacketsFlooded                   *prometheus.Desc // \Hyper-V Virtual Switch(*)\Packets Flooded
	virtualSwitchPackets                          *prometheus.Desc // \Hyper-V Virtual Switch(*)\Packets/sec
	virtualSwitchPacketsReceived                  *prometheus.Desc // \Hyper-V Virtual Switch(*)\Packets Received/sec
	virtualSwitchPacketsSent                      *prometheus.Desc // \Hyper-V Virtual Switch(*)\Packets Sent/sec
	virtualSwitchPurgedMacAddresses               *prometheus.Desc // \Hyper-V Virtual Switch(*)\Purged Mac Addresses
}

type perfDataCounterValuesVirtualSwitch struct {
	Name string

	VirtualSwitchBroadcastPacketsReceived         float64 `perfdata:"Broadcast Packets Received/sec"`
	VirtualSwitchBroadcastPacketsSent             float64 `perfdata:"Broadcast Packets Sent/sec"`
	VirtualSwitchBytes                            float64 `perfdata:"Bytes/sec"`
	VirtualSwitchBytesReceived                    float64 `perfdata:"Bytes Received/sec"`
	VirtualSwitchBytesSent                        float64 `perfdata:"Bytes Sent/sec"`
	VirtualSwitchDirectedPacketsReceived          float64 `perfdata:"Directed Packets Received/sec"`
	VirtualSwitchDirectedPacketsSent              float64 `perfdata:"Directed Packets Sent/sec"`
	VirtualSwitchDroppedPacketsIncoming           float64 `perfdata:"Dropped Packets Incoming/sec"`
	VirtualSwitchDroppedPacketsOutgoing           float64 `perfdata:"Dropped Packets Outgoing/sec"`
	VirtualSwitchExtensionsDroppedPacketsIncoming float64 `perfdata:"Extensions Dropped Packets Incoming/sec"`
	VirtualSwitchExtensionsDroppedPacketsOutgoing float64 `perfdata:"Extensions Dropped Packets Outgoing/sec"`
	VirtualSwitchLearnedMacAddresses              float64 `perfdata:"Learned Mac Addresses"`
	VirtualSwitchMulticastPacketsReceived         float64 `perfdata:"Multicast Packets Received/sec"`
	VirtualSwitchMulticastPacketsSent             float64 `perfdata:"Multicast Packets Sent/sec"`
	VirtualSwitchNumberOfSendChannelMoves         float64 `perfdata:"Number of Send Channel Moves/sec"`
	VirtualSwitchNumberOfVMQMoves                 float64 `perfdata:"Number of VMQ Moves/sec"`
	VirtualSwitchPacketsFlooded                   float64 `perfdata:"Packets Flooded"`
	VirtualSwitchPackets                          float64 `perfdata:"Packets/sec"`
	VirtualSwitchPacketsReceived                  float64 `perfdata:"Packets Received/sec"`
	VirtualSwitchPacketsSent                      float64 `perfdata:"Packets Sent/sec"`
	VirtualSwitchPurgedMacAddresses               float64 `perfdata:"Purged Mac Addresses"`
}

func (c *Collector) buildVirtualSwitch() error {
	var err error

	c.perfDataCollectorVirtualSwitch, err = pdh.NewCollector[perfDataCounterValuesVirtualSwitch](pdh.CounterTypeRaw, "Hyper-V Virtual Switch", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create Hyper-V Virtual Switch collector: %w", err)
	}

	c.virtualSwitchBroadcastPacketsReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vswitch_broadcast_packets_received_total"),
		"Represents the total number of broadcast packets received per second by the virtual switch",
		[]string{"vswitch"},
		nil,
	)
	c.virtualSwitchBroadcastPacketsSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vswitch_broadcast_packets_sent_total"),
		"Represents the total number of broadcast packets sent per second by the virtual switch",
		[]string{"vswitch"},
		nil,
	)
	c.virtualSwitchBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vswitch_bytes_total"),
		"Represents the total number of bytes per second traversing the virtual switch",
		[]string{"vswitch"},
		nil,
	)
	c.virtualSwitchBytesReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vswitch_bytes_received_total"),
		"Represents the total number of bytes received per second by the virtual switch",
		[]string{"vswitch"},
		nil,
	)
	c.virtualSwitchBytesSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vswitch_bytes_sent_total"),
		"Represents the total number of bytes sent per second by the virtual switch",
		[]string{"vswitch"},
		nil,
	)
	c.virtualSwitchDirectedPacketsReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vswitch_directed_packets_received_total"),
		"Represents the total number of directed packets received per second by the virtual switch",
		[]string{"vswitch"},
		nil,
	)
	c.virtualSwitchDirectedPacketsSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vswitch_directed_packets_send_total"),
		"Represents the total number of directed packets sent per second by the virtual switch",
		[]string{"vswitch"},
		nil,
	)
	c.virtualSwitchDroppedPacketsIncoming = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vswitch_dropped_packets_incoming_total"),
		"Represents the total number of packet dropped per second by the virtual switch in the incoming direction",
		[]string{"vswitch"},
		nil,
	)
	c.virtualSwitchDroppedPacketsOutgoing = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vswitch_dropped_packets_outcoming_total"),
		"Represents the total number of packet dropped per second by the virtual switch in the outgoing direction",
		[]string{"vswitch"},
		nil,
	)
	c.virtualSwitchExtensionsDroppedPacketsIncoming = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vswitch_extensions_dropped_packets_incoming_total"),
		"Represents the total number of packet dropped per second by the virtual switch extensions in the incoming direction",
		[]string{"vswitch"},
		nil,
	)
	c.virtualSwitchExtensionsDroppedPacketsOutgoing = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vswitch_extensions_dropped_packets_outcoming_total"),
		"Represents the total number of packet dropped per second by the virtual switch extensions in the outgoing direction",
		[]string{"vswitch"},
		nil,
	)
	c.virtualSwitchLearnedMacAddresses = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vswitch_learned_mac_addresses_total"),
		"Represents the total number of learned MAC addresses of the virtual switch",
		[]string{"vswitch"},
		nil,
	)
	c.virtualSwitchMulticastPacketsReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vswitch_multicast_packets_received_total"),
		"Represents the total number of multicast packets received per second by the virtual switch",
		[]string{"vswitch"},
		nil,
	)
	c.virtualSwitchMulticastPacketsSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vswitch_multicast_packets_sent_total"),
		"Represents the total number of multicast packets sent per second by the virtual switch",
		[]string{"vswitch"},
		nil,
	)
	c.virtualSwitchNumberOfSendChannelMoves = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vswitch_number_of_send_channel_moves_total"),
		"Represents the total number of send channel moves per second on this virtual switch",
		[]string{"vswitch"},
		nil,
	)
	c.virtualSwitchNumberOfVMQMoves = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vswitch_number_of_vmq_moves_total"),
		"Represents the total number of VMQ moves per second on this virtual switch",
		[]string{"vswitch"},
		nil,
	)
	c.virtualSwitchPacketsFlooded = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vswitch_packets_flooded_total"),
		"Represents the total number of packets flooded by the virtual switch",
		[]string{"vswitch"},
		nil,
	)
	c.virtualSwitchPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vswitch_packets_total"),
		"Represents the total number of packets per second traversing the virtual switch",
		[]string{"vswitch"},
		nil,
	)
	c.virtualSwitchPacketsReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vswitch_packets_received_total"),
		"Represents the total number of packets received per second by the virtual switch",
		[]string{"vswitch"},
		nil,
	)
	c.virtualSwitchPacketsSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vswitch_packets_sent_total"),
		"Represents the total number of packets send per second by the virtual switch",
		[]string{"vswitch"},
		nil,
	)
	c.virtualSwitchPurgedMacAddresses = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vswitch_purged_mac_addresses_total"),
		"Represents the total number of purged MAC addresses of the virtual switch",
		[]string{"vswitch"},
		nil,
	)

	return nil
}

func (c *Collector) collectVirtualSwitch(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollectorVirtualSwitch.Collect(&c.perfDataObjectVirtualSwitch)
	if err != nil {
		return fmt.Errorf("failed to collect Hyper-V Virtual Switch metrics: %w", err)
	}

	for _, data := range c.perfDataObjectVirtualSwitch {
		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchBroadcastPacketsReceived,
			prometheus.CounterValue,
			data.VirtualSwitchBroadcastPacketsReceived,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchBroadcastPacketsSent,
			prometheus.CounterValue,
			data.VirtualSwitchBroadcastPacketsSent,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchBytes,
			prometheus.CounterValue,
			data.VirtualSwitchBytes,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchBytesReceived,
			prometheus.CounterValue,
			data.VirtualSwitchBytesReceived,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchBytesSent,
			prometheus.CounterValue,
			data.VirtualSwitchBytesSent,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchDirectedPacketsReceived,
			prometheus.CounterValue,
			data.VirtualSwitchDirectedPacketsReceived,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchDirectedPacketsSent,
			prometheus.CounterValue,
			data.VirtualSwitchDirectedPacketsSent,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchDroppedPacketsIncoming,
			prometheus.CounterValue,
			data.VirtualSwitchDroppedPacketsIncoming,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchDroppedPacketsOutgoing,
			prometheus.CounterValue,
			data.VirtualSwitchDroppedPacketsOutgoing,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchExtensionsDroppedPacketsIncoming,
			prometheus.CounterValue,
			data.VirtualSwitchExtensionsDroppedPacketsIncoming,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchExtensionsDroppedPacketsOutgoing,
			prometheus.CounterValue,
			data.VirtualSwitchExtensionsDroppedPacketsOutgoing,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchLearnedMacAddresses,
			prometheus.CounterValue,
			data.VirtualSwitchLearnedMacAddresses,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchMulticastPacketsReceived,
			prometheus.CounterValue,
			data.VirtualSwitchMulticastPacketsReceived,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchMulticastPacketsSent,
			prometheus.CounterValue,
			data.VirtualSwitchMulticastPacketsSent,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchNumberOfSendChannelMoves,
			prometheus.CounterValue,
			data.VirtualSwitchNumberOfSendChannelMoves,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchNumberOfVMQMoves,
			prometheus.CounterValue,
			data.VirtualSwitchNumberOfVMQMoves,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchPacketsFlooded,
			prometheus.CounterValue,
			data.VirtualSwitchPacketsFlooded,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchPackets,
			prometheus.CounterValue,
			data.VirtualSwitchPackets,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchPacketsReceived,
			prometheus.CounterValue,
			data.VirtualSwitchPacketsReceived,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchPacketsSent,
			prometheus.CounterValue,
			data.VirtualSwitchPacketsSent,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchPurgedMacAddresses,
			prometheus.CounterValue,
			data.VirtualSwitchPurgedMacAddresses,
			data.Name,
		)
	}

	return nil
}

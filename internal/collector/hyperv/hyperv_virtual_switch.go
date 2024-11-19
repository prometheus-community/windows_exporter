package hyperv

import (
	"errors"
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

// collectorVirtualMachineHealthSummary Hyper-V Virtual Switch Summary metrics
type collectorVirtualSwitch struct {
	perfDataCollectorVirtualSwitch                *perfdata.Collector
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

const (
	virtualSwitchBroadcastPacketsReceived         = "Broadcast Packets Received/sec"
	virtualSwitchBroadcastPacketsSent             = "Broadcast Packets Sent/sec"
	virtualSwitchBytes                            = "Bytes/sec"
	virtualSwitchBytesReceived                    = "Bytes Received/sec"
	virtualSwitchBytesSent                        = "Bytes Sent/sec"
	virtualSwitchDirectedPacketsReceived          = "Directed Packets Received/sec"
	virtualSwitchDirectedPacketsSent              = "Directed Packets Sent/sec"
	virtualSwitchDroppedPacketsIncoming           = "Dropped Packets Incoming/sec"
	virtualSwitchDroppedPacketsOutgoing           = "Dropped Packets Outgoing/sec"
	virtualSwitchExtensionsDroppedPacketsIncoming = "Extensions Dropped Packets Incoming/sec"
	virtualSwitchExtensionsDroppedPacketsOutgoing = "Extensions Dropped Packets Outgoing/sec"
	virtualSwitchLearnedMacAddresses              = "Learned Mac Addresses"
	virtualSwitchMulticastPacketsReceived         = "Multicast Packets Received/sec"
	virtualSwitchMulticastPacketsSent             = "Multicast Packets Sent/sec"
	virtualSwitchNumberOfSendChannelMoves         = "Number of Send Channel Moves/sec"
	virtualSwitchNumberOfVMQMoves                 = "Number of VMQ Moves/sec"
	virtualSwitchPacketsFlooded                   = "Packets Flooded"
	virtualSwitchPackets                          = "Packets/sec"
	virtualSwitchPacketsReceived                  = "Packets Received/sec"
	virtualSwitchPacketsSent                      = "Packets Sent/sec"
	virtualSwitchPurgedMacAddresses               = "Purged Mac Addresses"
)

func (c *Collector) buildVirtualSwitch() error {
	var err error

	c.perfDataCollectorVirtualSwitch, err = perfdata.NewCollector("Hyper-V Virtual Switch", perfdata.InstanceAll, []string{
		virtualSwitchBroadcastPacketsReceived,
		virtualSwitchBroadcastPacketsSent,
		virtualSwitchBytes,
		virtualSwitchBytesReceived,
		virtualSwitchBytesSent,
		virtualSwitchDirectedPacketsReceived,
		virtualSwitchDirectedPacketsSent,
		virtualSwitchDroppedPacketsIncoming,
		virtualSwitchDroppedPacketsOutgoing,
		virtualSwitchExtensionsDroppedPacketsIncoming,
		virtualSwitchExtensionsDroppedPacketsOutgoing,
		virtualSwitchLearnedMacAddresses,
		virtualSwitchMulticastPacketsReceived,
		virtualSwitchMulticastPacketsSent,
		virtualSwitchNumberOfSendChannelMoves,
		virtualSwitchNumberOfVMQMoves,
		virtualSwitchPacketsFlooded,
		virtualSwitchPackets,
		virtualSwitchPacketsReceived,
		virtualSwitchPacketsSent,
		virtualSwitchPurgedMacAddresses,
	})
	if err != nil && !errors.Is(err, perfdata.ErrNoData) {
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
	data, err := c.perfDataCollectorVirtualSwitch.Collect()
	if err != nil && !errors.Is(err, perfdata.ErrNoData) {
		return fmt.Errorf("failed to collect Hyper-V Virtual Switch metrics: %w", err)
	}

	for name, switchData := range data {
		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchBroadcastPacketsReceived,
			prometheus.CounterValue,
			switchData[virtualSwitchBroadcastPacketsReceived].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchBroadcastPacketsSent,
			prometheus.CounterValue,
			switchData[virtualSwitchBroadcastPacketsSent].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchBytes,
			prometheus.CounterValue,
			switchData[virtualSwitchBytes].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchBytesReceived,
			prometheus.CounterValue,
			switchData[virtualSwitchBytesReceived].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchBytesSent,
			prometheus.CounterValue,
			switchData[virtualSwitchBytesSent].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchDirectedPacketsReceived,
			prometheus.CounterValue,
			switchData[virtualSwitchDirectedPacketsReceived].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchDirectedPacketsSent,
			prometheus.CounterValue,
			switchData[virtualSwitchDirectedPacketsSent].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchDroppedPacketsIncoming,
			prometheus.CounterValue,
			switchData[virtualSwitchDroppedPacketsIncoming].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchDroppedPacketsOutgoing,
			prometheus.CounterValue,
			switchData[virtualSwitchDroppedPacketsOutgoing].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchExtensionsDroppedPacketsIncoming,
			prometheus.CounterValue,
			switchData[virtualSwitchExtensionsDroppedPacketsIncoming].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchExtensionsDroppedPacketsOutgoing,
			prometheus.CounterValue,
			switchData[virtualSwitchExtensionsDroppedPacketsOutgoing].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchLearnedMacAddresses,
			prometheus.CounterValue,
			switchData[virtualSwitchLearnedMacAddresses].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchMulticastPacketsReceived,
			prometheus.CounterValue,
			switchData[virtualSwitchMulticastPacketsReceived].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchMulticastPacketsSent,
			prometheus.CounterValue,
			switchData[virtualSwitchMulticastPacketsSent].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchNumberOfSendChannelMoves,
			prometheus.CounterValue,
			switchData[virtualSwitchNumberOfSendChannelMoves].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchNumberOfVMQMoves,
			prometheus.CounterValue,
			switchData[virtualSwitchNumberOfVMQMoves].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchPacketsFlooded,
			prometheus.CounterValue,
			switchData[virtualSwitchPacketsFlooded].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchPackets,
			prometheus.CounterValue,
			switchData[virtualSwitchPackets].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchPacketsReceived,
			prometheus.CounterValue,
			switchData[virtualSwitchPacketsReceived].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchPacketsSent,
			prometheus.CounterValue,
			switchData[virtualSwitchPacketsSent].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualSwitchPurgedMacAddresses,
			prometheus.CounterValue,
			switchData[virtualSwitchPurgedMacAddresses].FirstValue,
			name,
		)
	}

	return nil
}

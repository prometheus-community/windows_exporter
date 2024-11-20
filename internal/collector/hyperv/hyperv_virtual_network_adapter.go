package hyperv

import (
	"errors"
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

// collectorVirtualNetworkAdapter Hyper-V Virtual Network Adapter metrics
type collectorVirtualNetworkAdapter struct {
	perfDataCollectorVirtualNetworkAdapter *perfdata.Collector

	virtualNetworkAdapterBytesReceived          *prometheus.Desc // \Hyper-V Virtual Network Adapter(*)\Bytes Received/sec
	virtualNetworkAdapterBytesSent              *prometheus.Desc // \Hyper-V Virtual Network Adapter(*)\Bytes Sent/sec
	virtualNetworkAdapterDroppedPacketsIncoming *prometheus.Desc // \Hyper-V Virtual Network Adapter(*)\Dropped Packets Incoming/sec
	virtualNetworkAdapterDroppedPacketsOutgoing *prometheus.Desc // \Hyper-V Virtual Network Adapter(*)\Dropped Packets Outgoing/sec
	virtualNetworkAdapterPacketsReceived        *prometheus.Desc // \Hyper-V Virtual Network Adapter(*)\Packets Received/sec
	virtualNetworkAdapterPacketsSent            *prometheus.Desc // \Hyper-V Virtual Network Adapter(*)\Packets Sent/sec
}

const (
	virtualNetworkAdapterBytesReceived          = "Bytes Received/sec"
	virtualNetworkAdapterBytesSent              = "Bytes Sent/sec"
	virtualNetworkAdapterDroppedPacketsIncoming = "Dropped Packets Incoming/sec"
	virtualNetworkAdapterDroppedPacketsOutgoing = "Dropped Packets Outgoing/sec"
	virtualNetworkAdapterPacketsReceived        = "Packets Received/sec"
	virtualNetworkAdapterPacketsSent            = "Packets Sent/sec"
)

func (c *Collector) buildVirtualNetworkAdapter() error {
	var err error

	c.perfDataCollectorVirtualNetworkAdapter, err = perfdata.NewCollector("Hyper-V Virtual Network Adapter", perfdata.InstanceAll, []string{
		virtualNetworkAdapterBytesReceived,
		virtualNetworkAdapterBytesSent,
		virtualNetworkAdapterDroppedPacketsIncoming,
		virtualNetworkAdapterDroppedPacketsOutgoing,
		virtualNetworkAdapterPacketsReceived,
		virtualNetworkAdapterPacketsSent,
	})
	if err != nil && !errors.Is(err, perfdata.ErrNoData) {
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
	data, err := c.perfDataCollectorVirtualNetworkAdapter.Collect()
	if err != nil && !errors.Is(err, perfdata.ErrNoData) {
		return fmt.Errorf("failed to collect Hyper-V Virtual Network Adapter metrics: %w", err)
	}

	for name, adapterData := range data {
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterBytesReceived,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterBytesReceived].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterBytesSent,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterBytesSent].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDroppedPacketsIncoming,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDroppedPacketsIncoming].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDroppedPacketsOutgoing,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDroppedPacketsOutgoing].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterPacketsReceived,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterPacketsReceived].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterPacketsSent,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterPacketsSent].FirstValue,
			name,
		)
	}

	return nil
}

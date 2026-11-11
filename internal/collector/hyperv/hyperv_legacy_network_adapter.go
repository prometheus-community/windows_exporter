// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package hyperv

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/labelcache"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

// collectorLegacyNetworkAdapter Hyper-V Legacy Network Adapter metrics
type collectorLegacyNetworkAdapter struct {
	perfDataCollectorLegacyNetworkAdapter *pdh.Collector[perfDataCounterValuesLegacyNetworkAdapter]
	perfDataObjectLegacyNetworkAdapter    []perfDataCounterValuesLegacyNetworkAdapter

	legacyNetworkAdapterBytesDropped   *prometheus.Desc // \Hyper-V Legacy Network Adapter(*)\Bytes Dropped
	legacyNetworkAdapterBytesReceived  *prometheus.Desc // \Hyper-V Legacy Network Adapter(*)\Bytes Received/sec
	legacyNetworkAdapterBytesSent      *prometheus.Desc // \Hyper-V Legacy Network Adapter(*)\Bytes Sent/sec
	legacyNetworkAdapterFramesDropped  *prometheus.Desc // \Hyper-V Legacy Network Adapter(*)\Frames Dropped
	legacyNetworkAdapterFramesReceived *prometheus.Desc // \Hyper-V Legacy Network Adapter(*)\Frames Received/sec
	legacyNetworkAdapterFramesSent     *prometheus.Desc // \Hyper-V Legacy Network Adapter(*)\Frames Sent/sec
}

type perfDataCounterValuesLegacyNetworkAdapter struct {
	Name string

	LegacyNetworkAdapterBytesDropped   float64 `perfdata:"Bytes Dropped"`
	LegacyNetworkAdapterBytesReceived  float64 `perfdata:"Bytes Received/sec"`
	LegacyNetworkAdapterBytesSent      float64 `perfdata:"Bytes Sent/sec"`
	LegacyNetworkAdapterFramesDropped  float64 `perfdata:"Frames Dropped"`
	LegacyNetworkAdapterFramesReceived float64 `perfdata:"Frames Received/sec"`
	LegacyNetworkAdapterFramesSent     float64 `perfdata:"Frames Sent/sec"`
}

func (c *Collector) buildLegacyNetworkAdapter() error {
	var err error

	c.perfDataCollectorLegacyNetworkAdapter, err = pdh.NewCollector[perfDataCounterValuesLegacyNetworkAdapter](c.logger, pdh.CounterTypeRaw, "Hyper-V Legacy Network Adapter", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create Hyper-V Legacy Network Adapter collector: %w", err)
	}

	c.legacyNetworkAdapterBytesDropped = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "legacy_network_adapter_bytes_dropped_total"),
		"Bytes Dropped is the number of bytes dropped on the network adapter",
		[]string{"adapter"},
		nil,
	)
	c.legacyNetworkAdapterBytesReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "legacy_network_adapter_bytes_received_total"),
		"Bytes received is the number of bytes received on the network adapter",
		[]string{"adapter"},
		nil,
	)
	c.legacyNetworkAdapterBytesSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "legacy_network_adapter_bytes_sent_total"),
		"Bytes sent is the number of bytes sent over the network adapter",
		[]string{"adapter"},
		nil,
	)
	c.legacyNetworkAdapterFramesDropped = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "legacy_network_adapter_frames_dropped_total"),
		"Frames Dropped is the number of frames dropped on the network adapter",
		[]string{"adapter"},
		nil,
	)
	c.legacyNetworkAdapterFramesReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "legacy_network_adapter_frames_received_total"),
		"Frames received is the number of frames received on the network adapter",
		[]string{"adapter"},
		nil,
	)
	c.legacyNetworkAdapterFramesSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "legacy_network_adapter_frames_sent_total"),
		"Frames sent is the number of frames sent over the network adapter",
		[]string{"adapter"},
		nil,
	)

	return nil
}

func (c *Collector) collectLegacyNetworkAdapter(ch chan<- prometheus.Metric, scrape *labelcache.Scrape) error {
	err := c.perfDataCollectorLegacyNetworkAdapter.Collect(&c.perfDataObjectLegacyNetworkAdapter)
	if err != nil {
		return fmt.Errorf("failed to collect Hyper-V Legacy Network Adapter metrics: %w", err)
	}

	for _, data := range c.perfDataObjectLegacyNetworkAdapter {
		ch <- scrape.MustNewConstMetric(
			c.legacyNetworkAdapterBytesDropped,
			prometheus.CounterValue,
			data.LegacyNetworkAdapterBytesDropped,
			data.Name,
		)

		ch <- scrape.MustNewConstMetric(
			c.legacyNetworkAdapterBytesReceived,
			prometheus.CounterValue,
			data.LegacyNetworkAdapterBytesReceived,
			data.Name,
		)

		ch <- scrape.MustNewConstMetric(
			c.legacyNetworkAdapterBytesSent,
			prometheus.CounterValue,
			data.LegacyNetworkAdapterBytesSent,
			data.Name,
		)

		ch <- scrape.MustNewConstMetric(
			c.legacyNetworkAdapterFramesReceived,
			prometheus.CounterValue,
			data.LegacyNetworkAdapterFramesReceived,
			data.Name,
		)

		ch <- scrape.MustNewConstMetric(
			c.legacyNetworkAdapterFramesDropped,
			prometheus.CounterValue,
			data.LegacyNetworkAdapterFramesDropped,
			data.Name,
		)

		ch <- scrape.MustNewConstMetric(
			c.legacyNetworkAdapterFramesSent,
			prometheus.CounterValue,
			data.LegacyNetworkAdapterFramesSent,
			data.Name,
		)
	}

	return nil
}

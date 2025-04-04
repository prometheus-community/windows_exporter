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

package remote_fx

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "remote_fx"

type Config struct{}

//nolint:gochecknoglobals
var ConfigDefaults = Config{}

// Collector
// A RemoteFxNetworkCollector is a Prometheus Collector for
// WMI Win32_PerfRawData_Counters_RemoteFXNetwork & Win32_PerfRawData_Counters_RemoteFXGraphics metrics
// https://wutils.com/wmi/root/cimv2/win32_perfrawdata_counters_remotefxnetwork/
// https://wutils.com/wmi/root/cimv2/win32_perfrawdata_counters_remotefxgraphics/
type Collector struct {
	config Config

	perfDataCollectorNetwork  *pdh.Collector
	perfDataObjectNetwork     []perfDataCounterValuesNetwork
	perfDataCollectorGraphics *pdh.Collector
	perfDataObjectGraphics    []perfDataCounterValuesGraphics

	// net
	baseTCPRTT               *prometheus.Desc
	baseUDPRTT               *prometheus.Desc
	currentTCPBandwidth      *prometheus.Desc
	currentTCPRTT            *prometheus.Desc
	currentUDPBandwidth      *prometheus.Desc
	currentUDPRTT            *prometheus.Desc
	fecRate                  *prometheus.Desc
	lossRate                 *prometheus.Desc
	retransmissionRate       *prometheus.Desc
	totalReceivedBytes       *prometheus.Desc
	totalSentBytes           *prometheus.Desc
	udpPacketsReceivedPerSec *prometheus.Desc
	udpPacketsSentPerSec     *prometheus.Desc

	// gfx
	averageEncodingTime                         *prometheus.Desc
	frameQuality                                *prometheus.Desc
	framesSkippedPerSecondInsufficientResources *prometheus.Desc
	graphicsCompressionRatio                    *prometheus.Desc
	inputFramesPerSecond                        *prometheus.Desc
	outputFramesPerSecond                       *prometheus.Desc
	sourceFramesPerSecond                       *prometheus.Desc
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
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	c.perfDataCollectorNetwork.Close()
	c.perfDataCollectorGraphics.Close()

	return nil
}

func (c *Collector) Build(*slog.Logger, *mi.Session) error {
	// net
	c.baseTCPRTT = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_base_tcp_rtt_seconds"),
		"Base TCP round-trip time (RTT) detected in seconds",
		[]string{"session_name"},
		nil,
	)
	c.baseUDPRTT = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_base_udp_rtt_seconds"),
		"Base UDP round-trip time (RTT) detected in seconds.",
		[]string{"session_name"},
		nil,
	)
	c.currentTCPBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_current_tcp_bandwidth"),
		"TCP Bandwidth detected in bytes per second.",
		[]string{"session_name"},
		nil,
	)
	c.currentTCPRTT = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_current_tcp_rtt_seconds"),
		"Average TCP round-trip time (RTT) detected in seconds.",
		[]string{"session_name"},
		nil,
	)
	c.currentUDPBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_current_udp_bandwidth"),
		"UDP Bandwidth detected in bytes per second.",
		[]string{"session_name"},
		nil,
	)
	c.currentUDPRTT = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_current_udp_rtt_seconds"),
		"Average UDP round-trip time (RTT) detected in seconds.",
		[]string{"session_name"},
		nil,
	)
	c.totalReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_received_bytes_total"),
		"(TotalReceivedBytes)",
		[]string{"session_name"},
		nil,
	)
	c.totalSentBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_sent_bytes_total"),
		"(TotalSentBytes)",
		[]string{"session_name"},
		nil,
	)
	c.udpPacketsReceivedPerSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_udp_packets_received_total"),
		"Rate in packets per second at which packets are received over UDP.",
		[]string{"session_name"},
		nil,
	)
	c.udpPacketsSentPerSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_udp_packets_sent_total"),
		"Rate in packets per second at which packets are sent over UDP.",
		[]string{"session_name"},
		nil,
	)
	c.fecRate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_fec_rate"),
		"Forward Error Correction (FEC) percentage",
		[]string{"session_name"},
		nil,
	)
	c.lossRate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_loss_rate"),
		"Loss percentage",
		[]string{"session_name"},
		nil,
	)
	c.retransmissionRate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_retransmission_rate"),
		"Percentage of packets that have been retransmitted",
		[]string{"session_name"},
		nil,
	)

	// gfx
	c.averageEncodingTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "gfx_average_encoding_time_seconds"),
		"Average frame encoding time in seconds",
		[]string{"session_name"},
		nil,
	)
	c.frameQuality = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "gfx_frame_quality"),
		"Quality of the output frame expressed as a percentage of the quality of the source frame.",
		[]string{"session_name"},
		nil,
	)
	c.framesSkippedPerSecondInsufficientResources = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "gfx_frames_skipped_insufficient_resource_total"),
		"Number of frames skipped per second due to insufficient client resources.",
		[]string{"session_name", "resource"},
		nil,
	)
	c.graphicsCompressionRatio = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "gfx_graphics_compression_ratio"),
		"Ratio of the number of bytes encoded to the number of bytes input.",
		[]string{"session_name"},
		nil,
	)
	c.inputFramesPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "gfx_input_frames_total"),
		"Number of sources frames provided as input to RemoteFX graphics per second.",
		[]string{"session_name"},
		nil,
	)
	c.outputFramesPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "gfx_output_frames_total"),
		"Number of frames sent to the client per second.",
		[]string{"session_name"},
		nil,
	)
	c.sourceFramesPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "gfx_source_frames_total"),
		"Number of frames composed by the source (DWM) per second.",
		[]string{"session_name"},
		nil,
	)

	var err error

	errs := make([]error, 0)

	c.perfDataCollectorNetwork, err = pdh.NewCollector[perfDataCounterValuesNetwork](pdh.CounterTypeRaw, "RemoteFX Network", pdh.InstancesAll)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create RemoteFX Network collector: %w", err))
	}

	c.perfDataCollectorGraphics, err = pdh.NewCollector[perfDataCounterValuesGraphics](pdh.CounterTypeRaw, "RemoteFX Graphics", pdh.InstancesAll)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create RemoteFX Graphics collector: %w", err))
	}

	return errors.Join(errs...)
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	errs := make([]error, 0)

	if err := c.collectRemoteFXNetworkCount(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting RemoteFX Network metrics: %w", err))
	}

	if err := c.collectRemoteFXGraphicsCounters(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting RemoteFX Graphics metrics: %w", err))
	}

	return errors.Join(errs...)
}

func (c *Collector) collectRemoteFXNetworkCount(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollectorNetwork.Collect(&c.perfDataObjectNetwork)
	if err != nil {
		return fmt.Errorf("failed to collect RemoteFX Network metrics: %w", err)
	}

	for _, data := range c.perfDataObjectNetwork {
		// only connect metrics for remote named sessions
		sessionName := normalizeSessionName(data.Name)
		if n := strings.ToLower(sessionName); n == "" || n == "services" || n == "console" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.baseTCPRTT,
			prometheus.GaugeValue,
			utils.MilliSecToSec(data.BaseTCPRTT),
			sessionName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.baseUDPRTT,
			prometheus.GaugeValue,
			utils.MilliSecToSec(data.BaseUDPRTT),
			sessionName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.currentTCPBandwidth,
			prometheus.GaugeValue,
			(data.CurrentTCPBandwidth*1000)/8,
			sessionName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.currentTCPRTT,
			prometheus.GaugeValue,
			utils.MilliSecToSec(data.CurrentTCPRTT),
			sessionName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.currentUDPBandwidth,
			prometheus.GaugeValue,
			(data.CurrentUDPBandwidth*1000)/8,
			sessionName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.currentUDPRTT,
			prometheus.GaugeValue,
			utils.MilliSecToSec(data.CurrentUDPRTT),
			sessionName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalReceivedBytes,
			prometheus.CounterValue,
			data.TotalReceivedBytes,
			sessionName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalSentBytes,
			prometheus.CounterValue,
			data.TotalSentBytes,
			sessionName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.udpPacketsReceivedPerSec,
			prometheus.CounterValue,
			data.UDPPacketsReceivedPersec,
			sessionName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.udpPacketsSentPerSec,
			prometheus.CounterValue,
			data.UDPPacketsSentPersec,
			sessionName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.fecRate,
			prometheus.GaugeValue,
			data.FECRate,
			sessionName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.lossRate,
			prometheus.GaugeValue,
			data.LossRate,
			sessionName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.retransmissionRate,
			prometheus.GaugeValue,
			data.RetransmissionRate,
			sessionName,
		)
	}

	return nil
}

func (c *Collector) collectRemoteFXGraphicsCounters(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollectorGraphics.Collect(&c.perfDataObjectGraphics)
	if err != nil {
		return fmt.Errorf("failed to collect RemoteFX Graphics metrics: %w", err)
	}

	for _, data := range c.perfDataObjectGraphics {
		// only connect metrics for remote named sessions
		sessionName := normalizeSessionName(data.Name)
		if n := strings.ToLower(sessionName); n == "" || n == "services" || n == "console" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.averageEncodingTime,
			prometheus.GaugeValue,
			utils.MilliSecToSec(data.AverageEncodingTime),
			sessionName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.frameQuality,
			prometheus.GaugeValue,
			data.FrameQuality,
			sessionName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.framesSkippedPerSecondInsufficientResources,
			prometheus.CounterValue,
			data.FramesSkippedPerSecondInsufficientClientResources,
			sessionName,
			"client",
		)
		ch <- prometheus.MustNewConstMetric(
			c.framesSkippedPerSecondInsufficientResources,
			prometheus.CounterValue,
			data.FramesSkippedPerSecondInsufficientNetworkResources,
			sessionName,
			"network",
		)
		ch <- prometheus.MustNewConstMetric(
			c.framesSkippedPerSecondInsufficientResources,
			prometheus.CounterValue,
			data.FramesSkippedPerSecondInsufficientServerResources,
			sessionName,
			"server",
		)
		ch <- prometheus.MustNewConstMetric(
			c.graphicsCompressionRatio,
			prometheus.GaugeValue,
			data.GraphicsCompressionratio,
			sessionName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.inputFramesPerSecond,
			prometheus.CounterValue,
			data.InputFramesPerSecond,
			sessionName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.outputFramesPerSecond,
			prometheus.CounterValue,
			data.OutputFramesPerSecond,
			sessionName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.sourceFramesPerSecond,
			prometheus.CounterValue,
			data.SourceFramesPerSecond,
			sessionName,
		)
	}

	return nil
}

// normalizeSessionName ensure that the session is the same between WTS API and performance counters.
func normalizeSessionName(sessionName string) string {
	return strings.Replace(sessionName, "RDP-tcp", "RDP-Tcp", 1)
}

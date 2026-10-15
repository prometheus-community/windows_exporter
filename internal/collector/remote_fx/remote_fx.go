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

package remote_fx

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

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

	perfDataCollectorNetwork  *pdh.Collector[perfDataCounterValuesNetwork]
	perfDataObjectNetwork     []perfDataCounterValuesNetwork
	perfDataCollectorGraphics *pdh.Collector[perfDataCounterValuesGraphics]
	perfDataObjectGraphics    []perfDataCounterValuesGraphics

	// net
	baseTCPRTT               *prometheus.Desc
	baseUDPRTT               *prometheus.Desc
	currentTCPBandwidth      *prometheus.Desc
	currentTCPRTT            *prometheus.Desc
	currentUDPBandwidth      *prometheus.Desc
	currentUDPRTT            *prometheus.Desc
	fec                      *prometheus.Desc
	fecBase                  *prometheus.Desc
	loss                     *prometheus.Desc
	lossBase                 *prometheus.Desc
	retransmission           *prometheus.Desc
	retransmissionBase       *prometheus.Desc
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

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	c.buildDescriptors()

	var err error

	errs := make([]error, 0)

	c.perfDataCollectorNetwork, err = pdh.NewCollector[perfDataCounterValuesNetwork](logger.With(slog.String("collector", Name)), pdh.CounterTypeRaw, "RemoteFX Network", pdh.InstancesAll)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create RemoteFX Network collector: %w", err))
	}

	c.perfDataCollectorGraphics, err = pdh.NewCollector[perfDataCounterValuesGraphics](logger.With(slog.String("collector", Name)), pdh.CounterTypeRaw, "RemoteFX Graphics", pdh.InstancesAll)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create RemoteFX Graphics collector: %w", err))
	}

	return errors.Join(errs...)
}

func (c *Collector) buildDescriptors() {
	// net
	c.baseTCPRTT = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_base_tcp_rtt_seconds"),
		"Base TCP round-trip time (RTT) detected in seconds.",
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
		prometheus.BuildFQName(types.Namespace, Name, "net_current_tcp_bandwidth_bytes_per_second"),
		"TCP bandwidth detected in bytes per second.",
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
		prometheus.BuildFQName(types.Namespace, Name, "net_current_udp_bandwidth_bytes_per_second"),
		"UDP bandwidth detected in bytes per second.",
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
		"Total number of bytes received since the connection started.",
		[]string{"session_name"},
		nil,
	)
	c.totalSentBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_sent_bytes_total"),
		"Total number of bytes sent since the connection started.",
		[]string{"session_name"},
		nil,
	)
	c.udpPacketsReceivedPerSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_udp_packets_received_total"),
		"Total number of packets received over UDP.",
		[]string{"session_name"},
		nil,
	)
	c.udpPacketsSentPerSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_udp_packets_sent_total"),
		"Total number of packets sent over UDP.",
		[]string{"session_name"},
		nil,
	)
	c.fec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_fec_total"),
		"Numerator of the Forward Error Correction (FEC) rate. Divide its rate by the rate of windows_remote_fx_net_fec_base_total to get the FEC ratio.",
		[]string{"session_name"},
		nil,
	)
	c.fecBase = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_fec_base_total"),
		"Base (denominator) of the Forward Error Correction (FEC) rate.",
		[]string{"session_name"},
		nil,
	)
	c.loss = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_loss_total"),
		"Numerator of the loss rate. Divide its rate by the rate of windows_remote_fx_net_loss_base_total to get the loss ratio.",
		[]string{"session_name"},
		nil,
	)
	c.lossBase = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_loss_base_total"),
		"Base (denominator) of the loss rate.",
		[]string{"session_name"},
		nil,
	)
	c.retransmission = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_retransmission_total"),
		"Numerator of the packet retransmission rate. Divide its rate by the rate of windows_remote_fx_net_retransmission_base_total to get the retransmission ratio.",
		[]string{"session_name"},
		nil,
	)
	c.retransmissionBase = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_retransmission_base_total"),
		"Base (denominator) of the packet retransmission rate.",
		[]string{"session_name"},
		nil,
	)

	// gfx
	c.averageEncodingTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "gfx_average_encoding_time_seconds"),
		"Average frame encoding time in seconds.",
		[]string{"session_name"},
		nil,
	)
	c.frameQuality = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "gfx_frame_quality_ratio"),
		"Quality of the output frame relative to the quality of the source frame (0-1).",
		[]string{"session_name"},
		nil,
	)
	c.framesSkippedPerSecondInsufficientResources = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "gfx_frames_skipped_insufficient_resource_total"),
		"Total number of frames skipped due to insufficient client, network or server resources.",
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
		"Total number of source frames provided as input to RemoteFX graphics.",
		[]string{"session_name"},
		nil,
	)
	c.outputFramesPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "gfx_output_frames_total"),
		"Total number of frames sent to the client.",
		[]string{"session_name"},
		nil,
	)
	c.sourceFramesPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "gfx_source_frames_total"),
		"Total number of frames composed by the source (DWM).",
		[]string{"session_name"},
		nil,
	)
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric, _ time.Duration) error {
	errs := make([]error, 0)

	if err := c.perfDataCollectorNetwork.Collect(&c.perfDataObjectNetwork); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting RemoteFX Network metrics: %w", err))
	} else {
		c.collectNetworkMetrics(ch, c.perfDataObjectNetwork)
	}

	if err := c.perfDataCollectorGraphics.Collect(&c.perfDataObjectGraphics); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting RemoteFX Graphics metrics: %w", err))
	} else {
		c.collectGraphicsMetrics(ch, c.perfDataObjectGraphics)
	}

	return errors.Join(errs...)
}

func (c *Collector) collectNetworkMetrics(ch chan<- prometheus.Metric, rows []perfDataCounterValuesNetwork) {
	for _, data := range rows {
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

		// Current TCP/UDP Bandwidth is in kbit/s (1000 bit/s).
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
			c.fec,
			prometheus.CounterValue,
			data.FECRate,
			sessionName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.fecBase,
			prometheus.CounterValue,
			data.FECRateBase,
			sessionName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.loss,
			prometheus.CounterValue,
			data.LossRate,
			sessionName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.lossBase,
			prometheus.CounterValue,
			data.LossRateBase,
			sessionName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.retransmission,
			prometheus.CounterValue,
			data.RetransmissionRate,
			sessionName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.retransmissionBase,
			prometheus.CounterValue,
			data.RetransmissionRateBase,
			sessionName,
		)
	}
}

func (c *Collector) collectGraphicsMetrics(ch chan<- prometheus.Metric, rows []perfDataCounterValuesGraphics) {
	for _, data := range rows {
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

		// Frame Quality is a percentage.
		ch <- prometheus.MustNewConstMetric(
			c.frameQuality,
			prometheus.GaugeValue,
			data.FrameQuality/100,
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
}

// normalizeSessionName ensure that the session is the same between WTS API and performance counters.
func normalizeSessionName(sessionName string) string {
	return strings.Replace(sessionName, "RDP-tcp", "RDP-Tcp", 1)
}

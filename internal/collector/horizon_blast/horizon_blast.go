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

package horizon_blast

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "horizon_blast"

type Config struct{}

//nolint:gochecknoglobals
var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for Horizon Blast performance counters.
type Collector struct {
	config Config

	// PDH Collectors
	perfDataCollectorSession           *pdh.Collector
	perfDataCollectorImaging           *pdh.Collector
	perfDataCollectorAudio             *pdh.Collector
	perfDataCollectorCDR               *pdh.Collector
	perfDataCollectorClipboard         *pdh.Collector
	perfDataCollectorHTML5MMR          *pdh.Collector
	perfDataCollectorOtherFeature      *pdh.Collector
	perfDataCollectorPrinting          *pdh.Collector
	perfDataCollectorRdeServer         *pdh.Collector
	perfDataCollectorRTAV              *pdh.Collector
	perfDataCollectorSDR               *pdh.Collector
	perfDataCollectorSerialPortScanner *pdh.Collector
	perfDataCollectorSmartCard         *pdh.Collector
	perfDataCollectorUSB               *pdh.Collector
	perfDataCollectorViewScanner       *pdh.Collector
	perfDataCollectorWindowsMediaMMR   *pdh.Collector

	// PDH Data Objects
	perfDataObjectSession           []perfDataCounterValuesSession
	perfDataObjectImaging           []perfDataCounterValuesImaging
	perfDataObjectAudio             []perfDataCounterValuesAudio
	perfDataObjectCDR               []perfDataCounterValuesCDR
	perfDataObjectClipboard         []perfDataCounterValuesClipboard
	perfDataObjectHTML5MMR          []perfDataCounterValuesHTML5MMR
	perfDataObjectOtherFeature      []perfDataCounterValuesOtherFeature
	perfDataObjectPrinting          []perfDataCounterValuesPrinting
	perfDataObjectRdeServer         []perfDataCounterValuesRdeServer
	perfDataObjectRTAV              []perfDataCounterValuesRTAV
	perfDataObjectSDR               []perfDataCounterValuesSDR
	perfDataObjectSerialPortScanner []perfDataCounterValuesSerialPortScanner
	perfDataObjectSmartCard         []perfDataCounterValuesSmartCard
	perfDataObjectUSB               []perfDataCounterValuesUSB
	perfDataObjectViewScanner       []perfDataCounterValuesViewScanner
	perfDataObjectWindowsMediaMMR   []perfDataCounterValuesWindowsMediaMMR

	// Session Counters metrics
	sessionAutoReconnectCount               *prometheus.Desc
	sessionCumulativeReceivedBytesUDP       *prometheus.Desc
	sessionCumulativeTransmittedBytesUDP    *prometheus.Desc
	sessionCumulativeReceivedBytesTCP       *prometheus.Desc
	sessionCumulativeTransmittedBytesTCP    *prometheus.Desc
	sessionInstantaneousReceivedBytesUDP    *prometheus.Desc
	sessionInstantaneousTransmittedBytesUDP *prometheus.Desc
	sessionInstantaneousReceivedBytesTCP    *prometheus.Desc
	sessionInstantaneousTransmittedBytesTCP *prometheus.Desc
	sessionReceivedPackets                  *prometheus.Desc
	sessionTransmittedPackets               *prometheus.Desc
	sessionReceivedBytes                    *prometheus.Desc
	sessionTransmittedBytes                 *prometheus.Desc
	sessionJitterUplink                     *prometheus.Desc
	sessionRTT                              *prometheus.Desc
	sessionPacketLossUplink                 *prometheus.Desc
	sessionEstimatedBandwidthUplink         *prometheus.Desc

	// Imaging Counters metrics
	imagingEncoderType          *prometheus.Desc
	imagingTotalDirtyFrames     *prometheus.Desc
	imagingTotalPoll            *prometheus.Desc
	imagingTotalFBC             *prometheus.Desc
	imagingTotalFrames          *prometheus.Desc
	imagingDirtyFramesPerSecond *prometheus.Desc
	imagingPollRate             *prometheus.Desc
	imagingFBCRate              *prometheus.Desc
	imagingFramesPerSecond      *prometheus.Desc
	imagingOutQueueingTime      *prometheus.Desc
	imagingInboundBandwidth     *prometheus.Desc
	imagingOutboundBandwidth    *prometheus.Desc
	imagingReceivedPackets      *prometheus.Desc
	imagingTransmittedPackets   *prometheus.Desc
	imagingReceivedBytes        *prometheus.Desc
	imagingTransmittedBytes     *prometheus.Desc

	// Audio Counters metrics
	audioOutQueueingTime    *prometheus.Desc
	audioInboundBandwidth   *prometheus.Desc
	audioOutboundBandwidth  *prometheus.Desc
	audioReceivedPackets    *prometheus.Desc
	audioTransmittedPackets *prometheus.Desc
	audioReceivedBytes      *prometheus.Desc
	audioTransmittedBytes   *prometheus.Desc

	// CDR Counters metrics
	cdrOutQueueingTime    *prometheus.Desc
	cdrInboundBandwidth   *prometheus.Desc
	cdrOutboundBandwidth  *prometheus.Desc
	cdrReceivedPackets    *prometheus.Desc
	cdrTransmittedPackets *prometheus.Desc
	cdrReceivedBytes      *prometheus.Desc
	cdrTransmittedBytes   *prometheus.Desc

	// Clipboard Counters metrics
	clipboardOutQueueingTime    *prometheus.Desc
	clipboardInboundBandwidth   *prometheus.Desc
	clipboardOutboundBandwidth  *prometheus.Desc
	clipboardReceivedPackets    *prometheus.Desc
	clipboardTransmittedPackets *prometheus.Desc
	clipboardReceivedBytes      *prometheus.Desc
	clipboardTransmittedBytes   *prometheus.Desc

	// HTML5 MMR Counters metrics
	html5mmrOutQueueingTime    *prometheus.Desc
	html5mmrInboundBandwidth   *prometheus.Desc
	html5mmrOutboundBandwidth  *prometheus.Desc
	html5mmrReceivedPackets    *prometheus.Desc
	html5mmrTransmittedPackets *prometheus.Desc
	html5mmrReceivedBytes      *prometheus.Desc
	html5mmrTransmittedBytes   *prometheus.Desc

	// Other Feature Counters metrics
	otherFeatureOutQueueingTime    *prometheus.Desc
	otherFeatureInboundBandwidth   *prometheus.Desc
	otherFeatureOutboundBandwidth  *prometheus.Desc
	otherFeatureReceivedPackets    *prometheus.Desc
	otherFeatureTransmittedPackets *prometheus.Desc
	otherFeatureReceivedBytes      *prometheus.Desc
	otherFeatureTransmittedBytes   *prometheus.Desc

	// Printing Counters metrics
	printingOutQueueingTime    *prometheus.Desc
	printingInboundBandwidth   *prometheus.Desc
	printingOutboundBandwidth  *prometheus.Desc
	printingReceivedPackets    *prometheus.Desc
	printingTransmittedPackets *prometheus.Desc
	printingReceivedBytes      *prometheus.Desc
	printingTransmittedBytes   *prometheus.Desc

	// RdeServer Counters metrics
	rdeServerOutQueueingTime    *prometheus.Desc
	rdeServerInboundBandwidth   *prometheus.Desc
	rdeServerOutboundBandwidth  *prometheus.Desc
	rdeServerReceivedPackets    *prometheus.Desc
	rdeServerTransmittedPackets *prometheus.Desc
	rdeServerReceivedBytes      *prometheus.Desc
	rdeServerTransmittedBytes   *prometheus.Desc

	// RTAV Counters metrics
	rtavOutQueueingTime    *prometheus.Desc
	rtavInboundBandwidth   *prometheus.Desc
	rtavOutboundBandwidth  *prometheus.Desc
	rtavReceivedPackets    *prometheus.Desc
	rtavTransmittedPackets *prometheus.Desc
	rtavReceivedBytes      *prometheus.Desc
	rtavTransmittedBytes   *prometheus.Desc

	// SDR Counters metrics
	sdrOutQueueingTime    *prometheus.Desc
	sdrInboundBandwidth   *prometheus.Desc
	sdrOutboundBandwidth  *prometheus.Desc
	sdrReceivedPackets    *prometheus.Desc
	sdrTransmittedPackets *prometheus.Desc
	sdrReceivedBytes      *prometheus.Desc
	sdrTransmittedBytes   *prometheus.Desc

	// Serial Port and Scanner Counters metrics
	serialPortScannerOutQueueingTime    *prometheus.Desc
	serialPortScannerInboundBandwidth   *prometheus.Desc
	serialPortScannerOutboundBandwidth  *prometheus.Desc
	serialPortScannerReceivedPackets    *prometheus.Desc
	serialPortScannerTransmittedPackets *prometheus.Desc
	serialPortScannerReceivedBytes      *prometheus.Desc
	serialPortScannerTransmittedBytes   *prometheus.Desc

	// Smart Card Counters metrics
	smartCardOutQueueingTime    *prometheus.Desc
	smartCardInboundBandwidth   *prometheus.Desc
	smartCardOutboundBandwidth  *prometheus.Desc
	smartCardReceivedPackets    *prometheus.Desc
	smartCardTransmittedPackets *prometheus.Desc
	smartCardReceivedBytes      *prometheus.Desc
	smartCardTransmittedBytes   *prometheus.Desc

	// USB Counters metrics
	usbOutQueueingTime    *prometheus.Desc
	usbInboundBandwidth   *prometheus.Desc
	usbOutboundBandwidth  *prometheus.Desc
	usbReceivedPackets    *prometheus.Desc
	usbTransmittedPackets *prometheus.Desc
	usbReceivedBytes      *prometheus.Desc
	usbTransmittedBytes   *prometheus.Desc

	// View Scanner Counters metrics
	viewScannerOutQueueingTime    *prometheus.Desc
	viewScannerInboundBandwidth   *prometheus.Desc
	viewScannerOutboundBandwidth  *prometheus.Desc
	viewScannerReceivedPackets    *prometheus.Desc
	viewScannerTransmittedPackets *prometheus.Desc
	viewScannerReceivedBytes      *prometheus.Desc
	viewScannerTransmittedBytes   *prometheus.Desc

	// Windows Media MMR Counters metrics
	windowsMediaMMROutQueueingTime    *prometheus.Desc
	windowsMediaMMRInboundBandwidth   *prometheus.Desc
	windowsMediaMMROutboundBandwidth  *prometheus.Desc
	windowsMediaMMRReceivedPackets    *prometheus.Desc
	windowsMediaMMRTransmittedPackets *prometheus.Desc
	windowsMediaMMRReceivedBytes      *prometheus.Desc
	windowsMediaMMRTransmittedBytes   *prometheus.Desc
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
	if c.perfDataCollectorSession != nil {
		c.perfDataCollectorSession.Close()
	}

	if c.perfDataCollectorImaging != nil {
		c.perfDataCollectorImaging.Close()
	}

	if c.perfDataCollectorAudio != nil {
		c.perfDataCollectorAudio.Close()
	}

	if c.perfDataCollectorCDR != nil {
		c.perfDataCollectorCDR.Close()
	}

	if c.perfDataCollectorClipboard != nil {
		c.perfDataCollectorClipboard.Close()
	}

	if c.perfDataCollectorHTML5MMR != nil {
		c.perfDataCollectorHTML5MMR.Close()
	}

	if c.perfDataCollectorOtherFeature != nil {
		c.perfDataCollectorOtherFeature.Close()
	}

	if c.perfDataCollectorPrinting != nil {
		c.perfDataCollectorPrinting.Close()
	}

	if c.perfDataCollectorRdeServer != nil {
		c.perfDataCollectorRdeServer.Close()
	}

	if c.perfDataCollectorRTAV != nil {
		c.perfDataCollectorRTAV.Close()
	}

	if c.perfDataCollectorSDR != nil {
		c.perfDataCollectorSDR.Close()
	}

	if c.perfDataCollectorSerialPortScanner != nil {
		c.perfDataCollectorSerialPortScanner.Close()
	}

	if c.perfDataCollectorSmartCard != nil {
		c.perfDataCollectorSmartCard.Close()
	}

	if c.perfDataCollectorUSB != nil {
		c.perfDataCollectorUSB.Close()
	}

	if c.perfDataCollectorViewScanner != nil {
		c.perfDataCollectorViewScanner.Close()
	}

	if c.perfDataCollectorWindowsMediaMMR != nil {
		c.perfDataCollectorWindowsMediaMMR.Close()
	}

	return nil
}

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	var err error

	errs := make([]error, 0)

	// Initialize PDH collectors
	// All Horizon Blast counters use instance notation (*) so we use pdh.InstancesAll
	c.perfDataCollectorSession, err = pdh.NewCollector[perfDataCounterValuesSession](
		logger.With(slog.String("collector", Name)), pdh.CounterTypeRaw, "Horizon Blast Session Counters", pdh.InstancesAll)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create Horizon Blast Session Counters collector: %w", err))
	}

	c.perfDataCollectorImaging, err = pdh.NewCollector[perfDataCounterValuesImaging](
		logger.With(slog.String("collector", Name)), pdh.CounterTypeRaw, "Horizon Blast Imaging Counters", pdh.InstancesAll)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create Horizon Blast Imaging Counters collector: %w", err))
	}

	c.perfDataCollectorAudio, err = pdh.NewCollector[perfDataCounterValuesAudio](
		logger.With(slog.String("collector", Name)), pdh.CounterTypeRaw, "Horizon Blast Audio Counters", pdh.InstancesAll)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create Horizon Blast Audio Counters collector: %w", err))
	}

	c.perfDataCollectorCDR, err = pdh.NewCollector[perfDataCounterValuesCDR](
		logger.With(slog.String("collector", Name)), pdh.CounterTypeRaw, "Horizon Blast CDR Counters", pdh.InstancesAll)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create Horizon Blast CDR Counters collector: %w", err))
	}

	c.perfDataCollectorClipboard, err = pdh.NewCollector[perfDataCounterValuesClipboard](
		logger.With(slog.String("collector", Name)), pdh.CounterTypeRaw, "Horizon Blast Clipboard Counters", pdh.InstancesAll)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create Horizon Blast Clipboard Counters collector: %w", err))
	}

	c.perfDataCollectorHTML5MMR, err = pdh.NewCollector[perfDataCounterValuesHTML5MMR](
		logger.With(slog.String("collector", Name)), pdh.CounterTypeRaw, "Horizon Blast HTML5 MMR Counters", pdh.InstancesAll)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create Horizon Blast HTML5 MMR Counters collector: %w", err))
	}

	c.perfDataCollectorOtherFeature, err = pdh.NewCollector[perfDataCounterValuesOtherFeature](
		logger.With(slog.String("collector", Name)), pdh.CounterTypeRaw, "Horizon Blast Other Feature Counters", pdh.InstancesAll)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create Horizon Blast Other Feature Counters collector: %w", err))
	}

	c.perfDataCollectorPrinting, err = pdh.NewCollector[perfDataCounterValuesPrinting](
		logger.With(slog.String("collector", Name)), pdh.CounterTypeRaw, "Horizon Blast Printing Counters", pdh.InstancesAll)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create Horizon Blast Printing Counters collector: %w", err))
	}

	c.perfDataCollectorRdeServer, err = pdh.NewCollector[perfDataCounterValuesRdeServer](
		logger.With(slog.String("collector", Name)), pdh.CounterTypeRaw, "Horizon Blast RdeServer Counters", pdh.InstancesAll)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create Horizon Blast RdeServer Counters collector: %w", err))
	}

	c.perfDataCollectorRTAV, err = pdh.NewCollector[perfDataCounterValuesRTAV](
		logger.With(slog.String("collector", Name)), pdh.CounterTypeRaw, "Horizon Blast RTAV Counters", pdh.InstancesAll)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create Horizon Blast RTAV Counters collector: %w", err))
	}

	c.perfDataCollectorSDR, err = pdh.NewCollector[perfDataCounterValuesSDR](
		logger.With(slog.String("collector", Name)), pdh.CounterTypeRaw, "Horizon Blast SDR Counters", pdh.InstancesAll)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create Horizon Blast SDR Counters collector: %w", err))
	}

	c.perfDataCollectorSerialPortScanner, err = pdh.NewCollector[perfDataCounterValuesSerialPortScanner](
		logger.With(slog.String("collector", Name)), pdh.CounterTypeRaw, "Horizon Blast Serial Port and Scanner Counters", pdh.InstancesAll)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create Horizon Blast Serial Port and Scanner Counters collector: %w", err))
	}

	c.perfDataCollectorSmartCard, err = pdh.NewCollector[perfDataCounterValuesSmartCard](
		logger.With(slog.String("collector", Name)), pdh.CounterTypeRaw, "Horizon Blast Smart Card Counters", pdh.InstancesAll)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create Horizon Blast Smart Card Counters collector: %w", err))
	}

	c.perfDataCollectorUSB, err = pdh.NewCollector[perfDataCounterValuesUSB](
		logger.With(slog.String("collector", Name)), pdh.CounterTypeRaw, "Horizon Blast USB Counters", pdh.InstancesAll)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create Horizon Blast USB Counters collector: %w", err))
	}

	c.perfDataCollectorViewScanner, err = pdh.NewCollector[perfDataCounterValuesViewScanner](
		logger.With(slog.String("collector", Name)), pdh.CounterTypeRaw, "Horizon Blast View Scanner Counters", pdh.InstancesAll)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create Horizon Blast View Scanner Counters collector: %w", err))
	}

	c.perfDataCollectorWindowsMediaMMR, err = pdh.NewCollector[perfDataCounterValuesWindowsMediaMMR](
		logger.With(slog.String("collector", Name)), pdh.CounterTypeRaw, "Horizon Blast Windows Media MMR Counters", pdh.InstancesAll)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create Horizon Blast Windows Media MMR Counters collector: %w", err))
	}

	// Initialize Session Counters metrics
	c.sessionAutoReconnectCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_automatic_reconnect_count_total"),
		"The number of automatic reconnects for the session.",
		nil, nil,
	)
	c.sessionCumulativeReceivedBytesUDP = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_cumulative_received_bytes_udp_total"),
		"Cumulative bytes received over UDP.",
		nil, nil,
	)
	c.sessionCumulativeTransmittedBytesUDP = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_cumulative_transmitted_bytes_udp_total"),
		"Cumulative bytes transmitted over UDP.",
		nil, nil,
	)
	c.sessionCumulativeReceivedBytesTCP = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_cumulative_received_bytes_tcp_total"),
		"Cumulative bytes received over TCP.",
		nil, nil,
	)
	c.sessionCumulativeTransmittedBytesTCP = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_cumulative_transmitted_bytes_tcp_total"),
		"Cumulative bytes transmitted over TCP.",
		nil, nil,
	)
	c.sessionInstantaneousReceivedBytesUDP = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_instantaneous_received_bytes_udp"),
		"Instantaneous bytes received over UDP.",
		nil, nil,
	)
	c.sessionInstantaneousTransmittedBytesUDP = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_instantaneous_transmitted_bytes_udp"),
		"Instantaneous bytes transmitted over UDP.",
		nil, nil,
	)
	c.sessionInstantaneousReceivedBytesTCP = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_instantaneous_received_bytes_tcp"),
		"Instantaneous bytes received over TCP.",
		nil, nil,
	)
	c.sessionInstantaneousTransmittedBytesTCP = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_instantaneous_transmitted_bytes_tcp"),
		"Instantaneous bytes transmitted over TCP.",
		nil, nil,
	)
	c.sessionReceivedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_received_packets_total"),
		"Total packets received.",
		nil, nil,
	)
	c.sessionTransmittedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_transmitted_packets_total"),
		"Total packets transmitted.",
		nil, nil,
	)
	c.sessionReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_received_bytes_total"),
		"Total bytes received.",
		nil, nil,
	)
	c.sessionTransmittedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_transmitted_bytes_total"),
		"Total bytes transmitted.",
		nil, nil,
	)
	c.sessionJitterUplink = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_jitter_uplink_milliseconds"),
		"Uplink jitter in milliseconds.",
		nil, nil,
	)
	c.sessionRTT = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_rtt_milliseconds"),
		"Round-trip time in milliseconds.",
		nil, nil,
	)
	c.sessionPacketLossUplink = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_packet_loss_uplink_percent"),
		"Uplink packet loss percentage.",
		nil, nil,
	)
	c.sessionEstimatedBandwidthUplink = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_estimated_bandwidth_uplink_kbps"),
		"Estimated uplink bandwidth in Kbps.",
		nil, nil,
	)

	// Initialize Imaging Counters metrics
	c.imagingEncoderType = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_encoder_type"),
		"Encoder type used for imaging.",
		nil, nil,
	)
	c.imagingTotalDirtyFrames = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_total_dirty_frames_total"),
		"Total number of dirty frames.",
		nil, nil,
	)
	c.imagingTotalPoll = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_total_poll_total"),
		"Total number of polls.",
		nil, nil,
	)
	c.imagingTotalFBC = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_total_fbc_total"),
		"Total number of FBC operations.",
		nil, nil,
	)
	c.imagingTotalFrames = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_total_frames_total"),
		"Total number of frames.",
		nil, nil,
	)
	c.imagingDirtyFramesPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_dirty_frames_per_second"),
		"Dirty frames per second.",
		nil, nil,
	)
	c.imagingPollRate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_poll_rate"),
		"Poll rate.",
		nil, nil,
	)
	c.imagingFBCRate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_fbc_rate"),
		"FBC rate.",
		nil, nil,
	)
	c.imagingFramesPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_frames_per_second"),
		"Frames per second.",
		nil, nil,
	)
	c.imagingOutQueueingTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_out_queueing_time_seconds"),
		"Out queueing time in seconds.",
		nil, nil,
	)
	c.imagingInboundBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_inbound_bandwidth_kbps"),
		"Inbound bandwidth in Kbps.",
		nil, nil,
	)
	c.imagingOutboundBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_outbound_bandwidth_kbps"),
		"Outbound bandwidth in Kbps.",
		nil, nil,
	)
	c.imagingReceivedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_received_packets_total"),
		"Total packets received.",
		nil, nil,
	)
	c.imagingTransmittedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_transmitted_packets_total"),
		"Total packets transmitted.",
		nil, nil,
	)
	c.imagingReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_received_bytes_total"),
		"Total bytes received.",
		nil, nil,
	)
	c.imagingTransmittedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_transmitted_bytes_total"),
		"Total bytes transmitted.",
		nil, nil,
	)

	// Initialize Audio Counters metrics
	c.audioOutQueueingTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "audio_out_queueing_time_seconds"),
		"Out queueing time in seconds.",
		nil, nil,
	)
	c.audioInboundBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "audio_inbound_bandwidth_kbps"),
		"Inbound bandwidth in Kbps.",
		nil, nil,
	)
	c.audioOutboundBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "audio_outbound_bandwidth_kbps"),
		"Outbound bandwidth in Kbps.",
		nil, nil,
	)
	c.audioReceivedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "audio_received_packets_total"),
		"Total packets received.",
		nil, nil,
	)
	c.audioTransmittedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "audio_transmitted_packets_total"),
		"Total packets transmitted.",
		nil, nil,
	)
	c.audioReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "audio_received_bytes_total"),
		"Total bytes received.",
		nil, nil,
	)
	c.audioTransmittedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "audio_transmitted_bytes_total"),
		"Total bytes transmitted.",
		nil, nil,
	)

	// Initialize CDR Counters metrics
	c.cdrOutQueueingTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cdr_out_queueing_time_seconds"),
		"Out queueing time in seconds.",
		nil, nil,
	)
	c.cdrInboundBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cdr_inbound_bandwidth_kbps"),
		"Inbound bandwidth in Kbps.",
		nil, nil,
	)
	c.cdrOutboundBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cdr_outbound_bandwidth_kbps"),
		"Outbound bandwidth in Kbps.",
		nil, nil,
	)
	c.cdrReceivedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cdr_received_packets_total"),
		"Total packets received.",
		nil, nil,
	)
	c.cdrTransmittedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cdr_transmitted_packets_total"),
		"Total packets transmitted.",
		nil, nil,
	)
	c.cdrReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cdr_received_bytes_total"),
		"Total bytes received.",
		nil, nil,
	)
	c.cdrTransmittedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cdr_transmitted_bytes_total"),
		"Total bytes transmitted.",
		nil, nil,
	)

	// Initialize Clipboard Counters metrics
	c.clipboardOutQueueingTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "clipboard_out_queueing_time_seconds"),
		"Out queueing time in seconds.",
		nil, nil,
	)
	c.clipboardInboundBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "clipboard_inbound_bandwidth_kbps"),
		"Inbound bandwidth in Kbps.",
		nil, nil,
	)
	c.clipboardOutboundBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "clipboard_outbound_bandwidth_kbps"),
		"Outbound bandwidth in Kbps.",
		nil, nil,
	)
	c.clipboardReceivedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "clipboard_received_packets_total"),
		"Total packets received.",
		nil, nil,
	)
	c.clipboardTransmittedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "clipboard_transmitted_packets_total"),
		"Total packets transmitted.",
		nil, nil,
	)
	c.clipboardReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "clipboard_received_bytes_total"),
		"Total bytes received.",
		nil, nil,
	)
	c.clipboardTransmittedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "clipboard_transmitted_bytes_total"),
		"Total bytes transmitted.",
		nil, nil,
	)

	// Initialize HTML5 MMR Counters metrics
	c.html5mmrOutQueueingTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "html5_mmr_out_queueing_time_seconds"),
		"Out queueing time in seconds.",
		nil, nil,
	)
	c.html5mmrInboundBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "html5_mmr_inbound_bandwidth_kbps"),
		"Inbound bandwidth in Kbps.",
		nil, nil,
	)
	c.html5mmrOutboundBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "html5_mmr_outbound_bandwidth_kbps"),
		"Outbound bandwidth in Kbps.",
		nil, nil,
	)
	c.html5mmrReceivedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "html5_mmr_received_packets_total"),
		"Total packets received.",
		nil, nil,
	)
	c.html5mmrTransmittedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "html5_mmr_transmitted_packets_total"),
		"Total packets transmitted.",
		nil, nil,
	)
	c.html5mmrReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "html5_mmr_received_bytes_total"),
		"Total bytes received.",
		nil, nil,
	)
	c.html5mmrTransmittedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "html5_mmr_transmitted_bytes_total"),
		"Total bytes transmitted.",
		nil, nil,
	)

	// Initialize Other Feature Counters metrics
	c.otherFeatureOutQueueingTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "other_feature_out_queueing_time_seconds"),
		"Out queueing time in seconds.",
		[]string{"feature"}, nil,
	)
	c.otherFeatureInboundBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "other_feature_inbound_bandwidth_kbps"),
		"Inbound bandwidth in Kbps.",
		[]string{"feature"}, nil,
	)
	c.otherFeatureOutboundBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "other_feature_outbound_bandwidth_kbps"),
		"Outbound bandwidth in Kbps.",
		[]string{"feature"}, nil,
	)
	c.otherFeatureReceivedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "other_feature_received_packets_total"),
		"Total packets received.",
		[]string{"feature"}, nil,
	)
	c.otherFeatureTransmittedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "other_feature_transmitted_packets_total"),
		"Total packets transmitted.",
		[]string{"feature"}, nil,
	)
	c.otherFeatureReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "other_feature_received_bytes_total"),
		"Total bytes received.",
		[]string{"feature"}, nil,
	)
	c.otherFeatureTransmittedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "other_feature_transmitted_bytes_total"),
		"Total bytes transmitted.",
		[]string{"feature"}, nil,
	)

	// Initialize Printing Counters metrics
	c.printingOutQueueingTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "printing_out_queueing_time_seconds"),
		"Out queueing time in seconds.",
		nil, nil,
	)
	c.printingInboundBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "printing_inbound_bandwidth_kbps"),
		"Inbound bandwidth in Kbps.",
		nil, nil,
	)
	c.printingOutboundBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "printing_outbound_bandwidth_kbps"),
		"Outbound bandwidth in Kbps.",
		nil, nil,
	)
	c.printingReceivedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "printing_received_packets_total"),
		"Total packets received.",
		nil, nil,
	)
	c.printingTransmittedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "printing_transmitted_packets_total"),
		"Total packets transmitted.",
		nil, nil,
	)
	c.printingReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "printing_received_bytes_total"),
		"Total bytes received.",
		nil, nil,
	)
	c.printingTransmittedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "printing_transmitted_bytes_total"),
		"Total bytes transmitted.",
		nil, nil,
	)

	// Initialize RdeServer Counters metrics
	c.rdeServerOutQueueingTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rde_server_out_queueing_time_seconds"),
		"Out queueing time in seconds.",
		nil, nil,
	)
	c.rdeServerInboundBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rde_server_inbound_bandwidth_kbps"),
		"Inbound bandwidth in Kbps.",
		nil, nil,
	)
	c.rdeServerOutboundBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rde_server_outbound_bandwidth_kbps"),
		"Outbound bandwidth in Kbps.",
		nil, nil,
	)
	c.rdeServerReceivedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rde_server_received_packets_total"),
		"Total packets received.",
		nil, nil,
	)
	c.rdeServerTransmittedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rde_server_transmitted_packets_total"),
		"Total packets transmitted.",
		nil, nil,
	)
	c.rdeServerReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rde_server_received_bytes_total"),
		"Total bytes received.",
		nil, nil,
	)
	c.rdeServerTransmittedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rde_server_transmitted_bytes_total"),
		"Total bytes transmitted.",
		nil, nil,
	)

	// Initialize RTAV Counters metrics
	c.rtavOutQueueingTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rtav_out_queueing_time_seconds"),
		"Out queueing time in seconds.",
		nil, nil,
	)
	c.rtavInboundBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rtav_inbound_bandwidth_kbps"),
		"Inbound bandwidth in Kbps.",
		nil, nil,
	)
	c.rtavOutboundBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rtav_outbound_bandwidth_kbps"),
		"Outbound bandwidth in Kbps.",
		nil, nil,
	)
	c.rtavReceivedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rtav_received_packets_total"),
		"Total packets received.",
		nil, nil,
	)
	c.rtavTransmittedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rtav_transmitted_packets_total"),
		"Total packets transmitted.",
		nil, nil,
	)
	c.rtavReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rtav_received_bytes_total"),
		"Total bytes received.",
		nil, nil,
	)
	c.rtavTransmittedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rtav_transmitted_bytes_total"),
		"Total bytes transmitted.",
		nil, nil,
	)

	// Initialize SDR Counters metrics
	c.sdrOutQueueingTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sdr_out_queueing_time_seconds"),
		"Out queueing time in seconds.",
		nil, nil,
	)
	c.sdrInboundBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sdr_inbound_bandwidth_kbps"),
		"Inbound bandwidth in Kbps.",
		nil, nil,
	)
	c.sdrOutboundBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sdr_outbound_bandwidth_kbps"),
		"Outbound bandwidth in Kbps.",
		nil, nil,
	)
	c.sdrReceivedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sdr_received_packets_total"),
		"Total packets received.",
		nil, nil,
	)
	c.sdrTransmittedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sdr_transmitted_packets_total"),
		"Total packets transmitted.",
		nil, nil,
	)
	c.sdrReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sdr_received_bytes_total"),
		"Total bytes received.",
		nil, nil,
	)
	c.sdrTransmittedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sdr_transmitted_bytes_total"),
		"Total bytes transmitted.",
		nil, nil,
	)

	// Initialize Serial Port and Scanner Counters metrics
	c.serialPortScannerOutQueueingTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "serial_port_scanner_out_queueing_time_seconds"),
		"Out queueing time in seconds.",
		nil, nil,
	)
	c.serialPortScannerInboundBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "serial_port_scanner_inbound_bandwidth_kbps"),
		"Inbound bandwidth in Kbps.",
		nil, nil,
	)
	c.serialPortScannerOutboundBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "serial_port_scanner_outbound_bandwidth_kbps"),
		"Outbound bandwidth in Kbps.",
		nil, nil,
	)
	c.serialPortScannerReceivedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "serial_port_scanner_received_packets_total"),
		"Total packets received.",
		nil, nil,
	)
	c.serialPortScannerTransmittedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "serial_port_scanner_transmitted_packets_total"),
		"Total packets transmitted.",
		nil, nil,
	)
	c.serialPortScannerReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "serial_port_scanner_received_bytes_total"),
		"Total bytes received.",
		nil, nil,
	)
	c.serialPortScannerTransmittedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "serial_port_scanner_transmitted_bytes_total"),
		"Total bytes transmitted.",
		nil, nil,
	)

	// Initialize Smart Card Counters metrics
	c.smartCardOutQueueingTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "smart_card_out_queueing_time_seconds"),
		"Out queueing time in seconds.",
		nil, nil,
	)
	c.smartCardInboundBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "smart_card_inbound_bandwidth_kbps"),
		"Inbound bandwidth in Kbps.",
		nil, nil,
	)
	c.smartCardOutboundBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "smart_card_outbound_bandwidth_kbps"),
		"Outbound bandwidth in Kbps.",
		nil, nil,
	)
	c.smartCardReceivedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "smart_card_received_packets_total"),
		"Total packets received.",
		nil, nil,
	)
	c.smartCardTransmittedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "smart_card_transmitted_packets_total"),
		"Total packets transmitted.",
		nil, nil,
	)
	c.smartCardReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "smart_card_received_bytes_total"),
		"Total bytes received.",
		nil, nil,
	)
	c.smartCardTransmittedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "smart_card_transmitted_bytes_total"),
		"Total bytes transmitted.",
		nil, nil,
	)

	// Initialize USB Counters metrics
	c.usbOutQueueingTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "usb_out_queueing_time_seconds"),
		"Out queueing time in seconds.",
		nil, nil,
	)
	c.usbInboundBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "usb_inbound_bandwidth_kbps"),
		"Inbound bandwidth in Kbps.",
		nil, nil,
	)
	c.usbOutboundBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "usb_outbound_bandwidth_kbps"),
		"Outbound bandwidth in Kbps.",
		nil, nil,
	)
	c.usbReceivedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "usb_received_packets_total"),
		"Total packets received.",
		nil, nil,
	)
	c.usbTransmittedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "usb_transmitted_packets_total"),
		"Total packets transmitted.",
		nil, nil,
	)
	c.usbReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "usb_received_bytes_total"),
		"Total bytes received.",
		nil, nil,
	)
	c.usbTransmittedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "usb_transmitted_bytes_total"),
		"Total bytes transmitted.",
		nil, nil,
	)

	// Initialize View Scanner Counters metrics
	c.viewScannerOutQueueingTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "view_scanner_out_queueing_time_seconds"),
		"Out queueing time in seconds.",
		nil, nil,
	)
	c.viewScannerInboundBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "view_scanner_inbound_bandwidth_kbps"),
		"Inbound bandwidth in Kbps.",
		nil, nil,
	)
	c.viewScannerOutboundBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "view_scanner_outbound_bandwidth_kbps"),
		"Outbound bandwidth in Kbps.",
		nil, nil,
	)
	c.viewScannerReceivedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "view_scanner_received_packets_total"),
		"Total packets received.",
		nil, nil,
	)
	c.viewScannerTransmittedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "view_scanner_transmitted_packets_total"),
		"Total packets transmitted.",
		nil, nil,
	)
	c.viewScannerReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "view_scanner_received_bytes_total"),
		"Total bytes received.",
		nil, nil,
	)
	c.viewScannerTransmittedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "view_scanner_transmitted_bytes_total"),
		"Total bytes transmitted.",
		nil, nil,
	)

	// Initialize Windows Media MMR Counters metrics
	c.windowsMediaMMROutQueueingTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "windows_media_mmr_out_queueing_time_seconds"),
		"Out queueing time in seconds.",
		nil, nil,
	)
	c.windowsMediaMMRInboundBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "windows_media_mmr_inbound_bandwidth_kbps"),
		"Inbound bandwidth in Kbps.",
		nil, nil,
	)
	c.windowsMediaMMROutboundBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "windows_media_mmr_outbound_bandwidth_kbps"),
		"Outbound bandwidth in Kbps.",
		nil, nil,
	)
	c.windowsMediaMMRReceivedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "windows_media_mmr_received_packets_total"),
		"Total packets received.",
		nil, nil,
	)
	c.windowsMediaMMRTransmittedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "windows_media_mmr_transmitted_packets_total"),
		"Total packets transmitted.",
		nil, nil,
	)
	c.windowsMediaMMRReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "windows_media_mmr_received_bytes_total"),
		"Total bytes received.",
		nil, nil,
	)
	c.windowsMediaMMRTransmittedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "windows_media_mmr_transmitted_bytes_total"),
		"Total bytes transmitted.",
		nil, nil,
	)

	return errors.Join(errs...)
}

// Collect sends the metric values for each metric to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	errs := make([]error, 0)

	if err := c.collectSession(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting Horizon Blast Session metrics: %w", err))
	}

	if err := c.collectImaging(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting Horizon Blast Imaging metrics: %w", err))
	}

	if err := c.collectAudio(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting Horizon Blast Audio metrics: %w", err))
	}

	if err := c.collectCDR(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting Horizon Blast CDR metrics: %w", err))
	}

	if err := c.collectClipboard(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting Horizon Blast Clipboard metrics: %w", err))
	}

	if err := c.collectHTML5MMR(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting Horizon Blast HTML5 MMR metrics: %w", err))
	}

	if err := c.collectOtherFeature(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting Horizon Blast Other Feature metrics: %w", err))
	}

	if err := c.collectPrinting(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting Horizon Blast Printing metrics: %w", err))
	}

	if err := c.collectRdeServer(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting Horizon Blast RdeServer metrics: %w", err))
	}

	if err := c.collectRTAV(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting Horizon Blast RTAV metrics: %w", err))
	}

	if err := c.collectSDR(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting Horizon Blast SDR metrics: %w", err))
	}

	if err := c.collectSerialPortScanner(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting Horizon Blast Serial Port Scanner metrics: %w", err))
	}

	if err := c.collectSmartCard(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting Horizon Blast Smart Card metrics: %w", err))
	}

	if err := c.collectUSB(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting Horizon Blast USB metrics: %w", err))
	}

	if err := c.collectViewScanner(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting Horizon Blast View Scanner metrics: %w", err))
	}

	if err := c.collectWindowsMediaMMR(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting Horizon Blast Windows Media MMR metrics: %w", err))
	}

	return errors.Join(errs...)
}

// microsecondsToSeconds converts microseconds to seconds.
func microsecondsToSeconds(us float64) float64 {
	return us / 1_000_000
}

func (c *Collector) collectSession(ch chan<- prometheus.Metric) error {
	if c.perfDataCollectorSession == nil {
		return nil
	}

	err := c.perfDataCollectorSession.Collect(&c.perfDataObjectSession)
	if err != nil {
		return fmt.Errorf("failed to collect Horizon Blast Session metrics: %w", err)
	}

	for _, data := range c.perfDataObjectSession {
		ch <- prometheus.MustNewConstMetric(c.sessionAutoReconnectCount, prometheus.CounterValue, data.AutomaticReconnectCount)

		ch <- prometheus.MustNewConstMetric(c.sessionCumulativeReceivedBytesUDP, prometheus.CounterValue, data.CumulativeReceivedBytesOverUDP)

		ch <- prometheus.MustNewConstMetric(c.sessionCumulativeTransmittedBytesUDP, prometheus.CounterValue, data.CumulativeTransmittedBytesOverUDP)

		ch <- prometheus.MustNewConstMetric(c.sessionCumulativeReceivedBytesTCP, prometheus.CounterValue, data.CumulativeReceivedBytesOverTCP)

		ch <- prometheus.MustNewConstMetric(c.sessionCumulativeTransmittedBytesTCP, prometheus.CounterValue, data.CumulativeTransmittedBytesOverTCP)

		ch <- prometheus.MustNewConstMetric(c.sessionInstantaneousReceivedBytesUDP, prometheus.GaugeValue, data.InstantaneousReceivedBytesOverUDP)

		ch <- prometheus.MustNewConstMetric(c.sessionInstantaneousTransmittedBytesUDP, prometheus.GaugeValue, data.InstantaneousTransmittedBytesOverUDP)

		ch <- prometheus.MustNewConstMetric(c.sessionInstantaneousReceivedBytesTCP, prometheus.GaugeValue, data.InstantaneousReceivedBytesOverTCP)

		ch <- prometheus.MustNewConstMetric(c.sessionInstantaneousTransmittedBytesTCP, prometheus.GaugeValue, data.InstantaneousTransmittedBytesOverTCP)

		ch <- prometheus.MustNewConstMetric(c.sessionReceivedPackets, prometheus.CounterValue, data.ReceivedPackets)

		ch <- prometheus.MustNewConstMetric(c.sessionTransmittedPackets, prometheus.CounterValue, data.TransmittedPackets)

		ch <- prometheus.MustNewConstMetric(c.sessionReceivedBytes, prometheus.CounterValue, data.ReceivedBytes)

		ch <- prometheus.MustNewConstMetric(c.sessionTransmittedBytes, prometheus.CounterValue, data.TransmittedBytes)

		ch <- prometheus.MustNewConstMetric(c.sessionJitterUplink, prometheus.GaugeValue, data.JitterUplink)

		ch <- prometheus.MustNewConstMetric(c.sessionRTT, prometheus.GaugeValue, data.RTT)

		ch <- prometheus.MustNewConstMetric(c.sessionPacketLossUplink, prometheus.GaugeValue, data.PacketLossUplink)

		ch <- prometheus.MustNewConstMetric(c.sessionEstimatedBandwidthUplink, prometheus.GaugeValue, data.EstimatedBandwidthUplink)
	}

	return nil
}

func (c *Collector) collectImaging(ch chan<- prometheus.Metric) error {
	if c.perfDataCollectorImaging == nil {
		return nil
	}

	err := c.perfDataCollectorImaging.Collect(&c.perfDataObjectImaging)
	if err != nil {
		return fmt.Errorf("failed to collect Horizon Blast Imaging metrics: %w", err)
	}

	for _, data := range c.perfDataObjectImaging {
		ch <- prometheus.MustNewConstMetric(c.imagingEncoderType, prometheus.GaugeValue, data.EncoderType)

		ch <- prometheus.MustNewConstMetric(c.imagingTotalDirtyFrames, prometheus.CounterValue, data.TotalDirtyFrames)

		ch <- prometheus.MustNewConstMetric(c.imagingTotalPoll, prometheus.CounterValue, data.TotalPoll)

		ch <- prometheus.MustNewConstMetric(c.imagingTotalFBC, prometheus.CounterValue, data.TotalFBC)

		ch <- prometheus.MustNewConstMetric(c.imagingTotalFrames, prometheus.CounterValue, data.TotalFrames)

		ch <- prometheus.MustNewConstMetric(c.imagingDirtyFramesPerSecond, prometheus.GaugeValue, data.DirtyFramesPerSecond)

		ch <- prometheus.MustNewConstMetric(c.imagingPollRate, prometheus.GaugeValue, data.PollRate)

		ch <- prometheus.MustNewConstMetric(c.imagingFBCRate, prometheus.GaugeValue, data.FBCRate)

		ch <- prometheus.MustNewConstMetric(c.imagingFramesPerSecond, prometheus.GaugeValue, data.FramesPerSecond)

		ch <- prometheus.MustNewConstMetric(c.imagingOutQueueingTime, prometheus.GaugeValue, microsecondsToSeconds(data.OutQueueingTimeUs))

		ch <- prometheus.MustNewConstMetric(c.imagingInboundBandwidth, prometheus.GaugeValue, data.InboundBandwidthKbps)

		ch <- prometheus.MustNewConstMetric(c.imagingOutboundBandwidth, prometheus.GaugeValue, data.OutboundBandwidthKbps)

		ch <- prometheus.MustNewConstMetric(c.imagingReceivedPackets, prometheus.CounterValue, data.ReceivedPackets)

		ch <- prometheus.MustNewConstMetric(c.imagingTransmittedPackets, prometheus.CounterValue, data.TransmittedPackets)

		ch <- prometheus.MustNewConstMetric(c.imagingReceivedBytes, prometheus.CounterValue, data.ReceivedBytes)

		ch <- prometheus.MustNewConstMetric(c.imagingTransmittedBytes, prometheus.CounterValue, data.TransmittedBytes)
	}

	return nil
}

func (c *Collector) collectAudio(ch chan<- prometheus.Metric) error {
	if c.perfDataCollectorAudio == nil {
		return nil
	}

	err := c.perfDataCollectorAudio.Collect(&c.perfDataObjectAudio)
	if err != nil {
		return fmt.Errorf("failed to collect Horizon Blast Audio metrics: %w", err)
	}

	for _, data := range c.perfDataObjectAudio {
		ch <- prometheus.MustNewConstMetric(c.audioOutQueueingTime, prometheus.GaugeValue, microsecondsToSeconds(data.OutQueueingTimeUs))

		ch <- prometheus.MustNewConstMetric(c.audioInboundBandwidth, prometheus.GaugeValue, data.InboundBandwidthKbps)

		ch <- prometheus.MustNewConstMetric(c.audioOutboundBandwidth, prometheus.GaugeValue, data.OutboundBandwidthKbps)

		ch <- prometheus.MustNewConstMetric(c.audioReceivedPackets, prometheus.CounterValue, data.ReceivedPackets)

		ch <- prometheus.MustNewConstMetric(c.audioTransmittedPackets, prometheus.CounterValue, data.TransmittedPackets)

		ch <- prometheus.MustNewConstMetric(c.audioReceivedBytes, prometheus.CounterValue, data.ReceivedBytes)

		ch <- prometheus.MustNewConstMetric(c.audioTransmittedBytes, prometheus.CounterValue, data.TransmittedBytes)
	}

	return nil
}

func (c *Collector) collectCDR(ch chan<- prometheus.Metric) error {
	if c.perfDataCollectorCDR == nil {
		return nil
	}

	err := c.perfDataCollectorCDR.Collect(&c.perfDataObjectCDR)
	if err != nil {
		return fmt.Errorf("failed to collect Horizon Blast CDR metrics: %w", err)
	}

	for _, data := range c.perfDataObjectCDR {
		ch <- prometheus.MustNewConstMetric(c.cdrOutQueueingTime, prometheus.GaugeValue, microsecondsToSeconds(data.OutQueueingTimeUs))

		ch <- prometheus.MustNewConstMetric(c.cdrInboundBandwidth, prometheus.GaugeValue, data.InboundBandwidthKbps)

		ch <- prometheus.MustNewConstMetric(c.cdrOutboundBandwidth, prometheus.GaugeValue, data.OutboundBandwidthKbps)

		ch <- prometheus.MustNewConstMetric(c.cdrReceivedPackets, prometheus.CounterValue, data.ReceivedPackets)

		ch <- prometheus.MustNewConstMetric(c.cdrTransmittedPackets, prometheus.CounterValue, data.TransmittedPackets)

		ch <- prometheus.MustNewConstMetric(c.cdrReceivedBytes, prometheus.CounterValue, data.ReceivedBytes)

		ch <- prometheus.MustNewConstMetric(c.cdrTransmittedBytes, prometheus.CounterValue, data.TransmittedBytes)
	}

	return nil
}

func (c *Collector) collectClipboard(ch chan<- prometheus.Metric) error {
	if c.perfDataCollectorClipboard == nil {
		return nil
	}

	err := c.perfDataCollectorClipboard.Collect(&c.perfDataObjectClipboard)
	if err != nil {
		return fmt.Errorf("failed to collect Horizon Blast Clipboard metrics: %w", err)
	}

	for _, data := range c.perfDataObjectClipboard {
		ch <- prometheus.MustNewConstMetric(c.clipboardOutQueueingTime, prometheus.GaugeValue, microsecondsToSeconds(data.OutQueueingTimeUs))

		ch <- prometheus.MustNewConstMetric(c.clipboardInboundBandwidth, prometheus.GaugeValue, data.InboundBandwidthKbps)

		ch <- prometheus.MustNewConstMetric(c.clipboardOutboundBandwidth, prometheus.GaugeValue, data.OutboundBandwidthKbps)

		ch <- prometheus.MustNewConstMetric(c.clipboardReceivedPackets, prometheus.CounterValue, data.ReceivedPackets)

		ch <- prometheus.MustNewConstMetric(c.clipboardTransmittedPackets, prometheus.CounterValue, data.TransmittedPackets)

		ch <- prometheus.MustNewConstMetric(c.clipboardReceivedBytes, prometheus.CounterValue, data.ReceivedBytes)

		ch <- prometheus.MustNewConstMetric(c.clipboardTransmittedBytes, prometheus.CounterValue, data.TransmittedBytes)
	}

	return nil
}

func (c *Collector) collectHTML5MMR(ch chan<- prometheus.Metric) error {
	if c.perfDataCollectorHTML5MMR == nil {
		return nil
	}

	err := c.perfDataCollectorHTML5MMR.Collect(&c.perfDataObjectHTML5MMR)
	if err != nil {
		return fmt.Errorf("failed to collect Horizon Blast HTML5 MMR metrics: %w", err)
	}

	for _, data := range c.perfDataObjectHTML5MMR {
		ch <- prometheus.MustNewConstMetric(c.html5mmrOutQueueingTime, prometheus.GaugeValue, microsecondsToSeconds(data.OutQueueingTimeUs))

		ch <- prometheus.MustNewConstMetric(c.html5mmrInboundBandwidth, prometheus.GaugeValue, data.InboundBandwidthKbps)

		ch <- prometheus.MustNewConstMetric(c.html5mmrOutboundBandwidth, prometheus.GaugeValue, data.OutboundBandwidthKbps)

		ch <- prometheus.MustNewConstMetric(c.html5mmrReceivedPackets, prometheus.CounterValue, data.ReceivedPackets)

		ch <- prometheus.MustNewConstMetric(c.html5mmrTransmittedPackets, prometheus.CounterValue, data.TransmittedPackets)

		ch <- prometheus.MustNewConstMetric(c.html5mmrReceivedBytes, prometheus.CounterValue, data.ReceivedBytes)

		ch <- prometheus.MustNewConstMetric(c.html5mmrTransmittedBytes, prometheus.CounterValue, data.TransmittedBytes)
	}

	return nil
}

func (c *Collector) collectOtherFeature(ch chan<- prometheus.Metric) error {
	if c.perfDataCollectorOtherFeature == nil {
		return nil
	}

	err := c.perfDataCollectorOtherFeature.Collect(&c.perfDataObjectOtherFeature)
	if err != nil {
		return fmt.Errorf("failed to collect Horizon Blast Other Feature metrics: %w", err)
	}

	for _, data := range c.perfDataObjectOtherFeature {
		featureName := data.Name

		ch <- prometheus.MustNewConstMetric(c.otherFeatureOutQueueingTime, prometheus.GaugeValue, microsecondsToSeconds(data.OutQueueingTimeUs), featureName)

		ch <- prometheus.MustNewConstMetric(c.otherFeatureInboundBandwidth, prometheus.GaugeValue, data.InboundBandwidthKbps, featureName)

		ch <- prometheus.MustNewConstMetric(c.otherFeatureOutboundBandwidth, prometheus.GaugeValue, data.OutboundBandwidthKbps, featureName)

		ch <- prometheus.MustNewConstMetric(c.otherFeatureReceivedPackets, prometheus.CounterValue, data.ReceivedPackets, featureName)

		ch <- prometheus.MustNewConstMetric(c.otherFeatureTransmittedPackets, prometheus.CounterValue, data.TransmittedPackets, featureName)

		ch <- prometheus.MustNewConstMetric(c.otherFeatureReceivedBytes, prometheus.CounterValue, data.ReceivedBytes, featureName)

		ch <- prometheus.MustNewConstMetric(c.otherFeatureTransmittedBytes, prometheus.CounterValue, data.TransmittedBytes, featureName)
	}

	return nil
}

func (c *Collector) collectPrinting(ch chan<- prometheus.Metric) error {
	if c.perfDataCollectorPrinting == nil {
		return nil
	}

	err := c.perfDataCollectorPrinting.Collect(&c.perfDataObjectPrinting)
	if err != nil {
		return fmt.Errorf("failed to collect Horizon Blast Printing metrics: %w", err)
	}

	for _, data := range c.perfDataObjectPrinting {
		ch <- prometheus.MustNewConstMetric(c.printingOutQueueingTime, prometheus.GaugeValue, microsecondsToSeconds(data.OutQueueingTimeUs))

		ch <- prometheus.MustNewConstMetric(c.printingInboundBandwidth, prometheus.GaugeValue, data.InboundBandwidthKbps)

		ch <- prometheus.MustNewConstMetric(c.printingOutboundBandwidth, prometheus.GaugeValue, data.OutboundBandwidthKbps)

		ch <- prometheus.MustNewConstMetric(c.printingReceivedPackets, prometheus.CounterValue, data.ReceivedPackets)

		ch <- prometheus.MustNewConstMetric(c.printingTransmittedPackets, prometheus.CounterValue, data.TransmittedPackets)

		ch <- prometheus.MustNewConstMetric(c.printingReceivedBytes, prometheus.CounterValue, data.ReceivedBytes)

		ch <- prometheus.MustNewConstMetric(c.printingTransmittedBytes, prometheus.CounterValue, data.TransmittedBytes)
	}

	return nil
}

func (c *Collector) collectRdeServer(ch chan<- prometheus.Metric) error {
	if c.perfDataCollectorRdeServer == nil {
		return nil
	}

	err := c.perfDataCollectorRdeServer.Collect(&c.perfDataObjectRdeServer)
	if err != nil {
		return fmt.Errorf("failed to collect Horizon Blast RdeServer metrics: %w", err)
	}

	for _, data := range c.perfDataObjectRdeServer {
		ch <- prometheus.MustNewConstMetric(c.rdeServerOutQueueingTime, prometheus.GaugeValue, microsecondsToSeconds(data.OutQueueingTimeUs))

		ch <- prometheus.MustNewConstMetric(c.rdeServerInboundBandwidth, prometheus.GaugeValue, data.InboundBandwidthKbps)

		ch <- prometheus.MustNewConstMetric(c.rdeServerOutboundBandwidth, prometheus.GaugeValue, data.OutboundBandwidthKbps)

		ch <- prometheus.MustNewConstMetric(c.rdeServerReceivedPackets, prometheus.CounterValue, data.ReceivedPackets)

		ch <- prometheus.MustNewConstMetric(c.rdeServerTransmittedPackets, prometheus.CounterValue, data.TransmittedPackets)

		ch <- prometheus.MustNewConstMetric(c.rdeServerReceivedBytes, prometheus.CounterValue, data.ReceivedBytes)

		ch <- prometheus.MustNewConstMetric(c.rdeServerTransmittedBytes, prometheus.CounterValue, data.TransmittedBytes)
	}

	return nil
}

func (c *Collector) collectRTAV(ch chan<- prometheus.Metric) error {
	if c.perfDataCollectorRTAV == nil {
		return nil
	}

	err := c.perfDataCollectorRTAV.Collect(&c.perfDataObjectRTAV)
	if err != nil {
		return fmt.Errorf("failed to collect Horizon Blast RTAV metrics: %w", err)
	}

	for _, data := range c.perfDataObjectRTAV {
		ch <- prometheus.MustNewConstMetric(c.rtavOutQueueingTime, prometheus.GaugeValue, microsecondsToSeconds(data.OutQueueingTimeUs))

		ch <- prometheus.MustNewConstMetric(c.rtavInboundBandwidth, prometheus.GaugeValue, data.InboundBandwidthKbps)

		ch <- prometheus.MustNewConstMetric(c.rtavOutboundBandwidth, prometheus.GaugeValue, data.OutboundBandwidthKbps)

		ch <- prometheus.MustNewConstMetric(c.rtavReceivedPackets, prometheus.CounterValue, data.ReceivedPackets)

		ch <- prometheus.MustNewConstMetric(c.rtavTransmittedPackets, prometheus.CounterValue, data.TransmittedPackets)

		ch <- prometheus.MustNewConstMetric(c.rtavReceivedBytes, prometheus.CounterValue, data.ReceivedBytes)

		ch <- prometheus.MustNewConstMetric(c.rtavTransmittedBytes, prometheus.CounterValue, data.TransmittedBytes)
	}

	return nil
}

func (c *Collector) collectSDR(ch chan<- prometheus.Metric) error {
	if c.perfDataCollectorSDR == nil {
		return nil
	}

	err := c.perfDataCollectorSDR.Collect(&c.perfDataObjectSDR)
	if err != nil {
		return fmt.Errorf("failed to collect Horizon Blast SDR metrics: %w", err)
	}

	for _, data := range c.perfDataObjectSDR {
		ch <- prometheus.MustNewConstMetric(c.sdrOutQueueingTime, prometheus.GaugeValue, microsecondsToSeconds(data.OutQueueingTimeUs))

		ch <- prometheus.MustNewConstMetric(c.sdrInboundBandwidth, prometheus.GaugeValue, data.InboundBandwidthKbps)

		ch <- prometheus.MustNewConstMetric(c.sdrOutboundBandwidth, prometheus.GaugeValue, data.OutboundBandwidthKbps)

		ch <- prometheus.MustNewConstMetric(c.sdrReceivedPackets, prometheus.CounterValue, data.ReceivedPackets)

		ch <- prometheus.MustNewConstMetric(c.sdrTransmittedPackets, prometheus.CounterValue, data.TransmittedPackets)

		ch <- prometheus.MustNewConstMetric(c.sdrReceivedBytes, prometheus.CounterValue, data.ReceivedBytes)

		ch <- prometheus.MustNewConstMetric(c.sdrTransmittedBytes, prometheus.CounterValue, data.TransmittedBytes)
	}

	return nil
}

func (c *Collector) collectSerialPortScanner(ch chan<- prometheus.Metric) error {
	if c.perfDataCollectorSerialPortScanner == nil {
		return nil
	}

	err := c.perfDataCollectorSerialPortScanner.Collect(&c.perfDataObjectSerialPortScanner)
	if err != nil {
		return fmt.Errorf("failed to collect Horizon Blast Serial Port Scanner metrics: %w", err)
	}

	for _, data := range c.perfDataObjectSerialPortScanner {
		ch <- prometheus.MustNewConstMetric(c.serialPortScannerOutQueueingTime, prometheus.GaugeValue, microsecondsToSeconds(data.OutQueueingTimeUs))

		ch <- prometheus.MustNewConstMetric(c.serialPortScannerInboundBandwidth, prometheus.GaugeValue, data.InboundBandwidthKbps)

		ch <- prometheus.MustNewConstMetric(c.serialPortScannerOutboundBandwidth, prometheus.GaugeValue, data.OutboundBandwidthKbps)

		ch <- prometheus.MustNewConstMetric(c.serialPortScannerReceivedPackets, prometheus.CounterValue, data.ReceivedPackets)

		ch <- prometheus.MustNewConstMetric(c.serialPortScannerTransmittedPackets, prometheus.CounterValue, data.TransmittedPackets)

		ch <- prometheus.MustNewConstMetric(c.serialPortScannerReceivedBytes, prometheus.CounterValue, data.ReceivedBytes)

		ch <- prometheus.MustNewConstMetric(c.serialPortScannerTransmittedBytes, prometheus.CounterValue, data.TransmittedBytes)
	}

	return nil
}

func (c *Collector) collectSmartCard(ch chan<- prometheus.Metric) error {
	if c.perfDataCollectorSmartCard == nil {
		return nil
	}

	err := c.perfDataCollectorSmartCard.Collect(&c.perfDataObjectSmartCard)
	if err != nil {
		return fmt.Errorf("failed to collect Horizon Blast Smart Card metrics: %w", err)
	}

	for _, data := range c.perfDataObjectSmartCard {
		ch <- prometheus.MustNewConstMetric(c.smartCardOutQueueingTime, prometheus.GaugeValue, microsecondsToSeconds(data.OutQueueingTimeUs))

		ch <- prometheus.MustNewConstMetric(c.smartCardInboundBandwidth, prometheus.GaugeValue, data.InboundBandwidthKbps)

		ch <- prometheus.MustNewConstMetric(c.smartCardOutboundBandwidth, prometheus.GaugeValue, data.OutboundBandwidthKbps)

		ch <- prometheus.MustNewConstMetric(c.smartCardReceivedPackets, prometheus.CounterValue, data.ReceivedPackets)

		ch <- prometheus.MustNewConstMetric(c.smartCardTransmittedPackets, prometheus.CounterValue, data.TransmittedPackets)

		ch <- prometheus.MustNewConstMetric(c.smartCardReceivedBytes, prometheus.CounterValue, data.ReceivedBytes)

		ch <- prometheus.MustNewConstMetric(c.smartCardTransmittedBytes, prometheus.CounterValue, data.TransmittedBytes)
	}

	return nil
}

func (c *Collector) collectUSB(ch chan<- prometheus.Metric) error {
	if c.perfDataCollectorUSB == nil {
		return nil
	}

	err := c.perfDataCollectorUSB.Collect(&c.perfDataObjectUSB)
	if err != nil {
		return fmt.Errorf("failed to collect Horizon Blast USB metrics: %w", err)
	}

	for _, data := range c.perfDataObjectUSB {
		ch <- prometheus.MustNewConstMetric(c.usbOutQueueingTime, prometheus.GaugeValue, microsecondsToSeconds(data.OutQueueingTimeUs))

		ch <- prometheus.MustNewConstMetric(c.usbInboundBandwidth, prometheus.GaugeValue, data.InboundBandwidthKbps)

		ch <- prometheus.MustNewConstMetric(c.usbOutboundBandwidth, prometheus.GaugeValue, data.OutboundBandwidthKbps)

		ch <- prometheus.MustNewConstMetric(c.usbReceivedPackets, prometheus.CounterValue, data.ReceivedPackets)

		ch <- prometheus.MustNewConstMetric(c.usbTransmittedPackets, prometheus.CounterValue, data.TransmittedPackets)

		ch <- prometheus.MustNewConstMetric(c.usbReceivedBytes, prometheus.CounterValue, data.ReceivedBytes)

		ch <- prometheus.MustNewConstMetric(c.usbTransmittedBytes, prometheus.CounterValue, data.TransmittedBytes)
	}

	return nil
}

func (c *Collector) collectViewScanner(ch chan<- prometheus.Metric) error {
	if c.perfDataCollectorViewScanner == nil {
		return nil
	}

	err := c.perfDataCollectorViewScanner.Collect(&c.perfDataObjectViewScanner)
	if err != nil {
		return fmt.Errorf("failed to collect Horizon Blast View Scanner metrics: %w", err)
	}

	for _, data := range c.perfDataObjectViewScanner {
		ch <- prometheus.MustNewConstMetric(c.viewScannerOutQueueingTime, prometheus.GaugeValue, microsecondsToSeconds(data.OutQueueingTimeUs))

		ch <- prometheus.MustNewConstMetric(c.viewScannerInboundBandwidth, prometheus.GaugeValue, data.InboundBandwidthKbps)

		ch <- prometheus.MustNewConstMetric(c.viewScannerOutboundBandwidth, prometheus.GaugeValue, data.OutboundBandwidthKbps)

		ch <- prometheus.MustNewConstMetric(c.viewScannerReceivedPackets, prometheus.CounterValue, data.ReceivedPackets)

		ch <- prometheus.MustNewConstMetric(c.viewScannerTransmittedPackets, prometheus.CounterValue, data.TransmittedPackets)

		ch <- prometheus.MustNewConstMetric(c.viewScannerReceivedBytes, prometheus.CounterValue, data.ReceivedBytes)

		ch <- prometheus.MustNewConstMetric(c.viewScannerTransmittedBytes, prometheus.CounterValue, data.TransmittedBytes)
	}

	return nil
}

func (c *Collector) collectWindowsMediaMMR(ch chan<- prometheus.Metric) error {
	if c.perfDataCollectorWindowsMediaMMR == nil {
		return nil
	}

	err := c.perfDataCollectorWindowsMediaMMR.Collect(&c.perfDataObjectWindowsMediaMMR)
	if err != nil {
		return fmt.Errorf("failed to collect Horizon Blast Windows Media MMR metrics: %w", err)
	}

	for _, data := range c.perfDataObjectWindowsMediaMMR {
		ch <- prometheus.MustNewConstMetric(c.windowsMediaMMROutQueueingTime, prometheus.GaugeValue, microsecondsToSeconds(data.OutQueueingTimeUs))

		ch <- prometheus.MustNewConstMetric(c.windowsMediaMMRInboundBandwidth, prometheus.GaugeValue, data.InboundBandwidthKbps)

		ch <- prometheus.MustNewConstMetric(c.windowsMediaMMROutboundBandwidth, prometheus.GaugeValue, data.OutboundBandwidthKbps)

		ch <- prometheus.MustNewConstMetric(c.windowsMediaMMRReceivedPackets, prometheus.CounterValue, data.ReceivedPackets)

		ch <- prometheus.MustNewConstMetric(c.windowsMediaMMRTransmittedPackets, prometheus.CounterValue, data.TransmittedPackets)

		ch <- prometheus.MustNewConstMetric(c.windowsMediaMMRReceivedBytes, prometheus.CounterValue, data.ReceivedBytes)

		ch <- prometheus.MustNewConstMetric(c.windowsMediaMMRTransmittedBytes, prometheus.CounterValue, data.TransmittedBytes)
	}

	return nil
}

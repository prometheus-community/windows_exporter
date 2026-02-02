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

// perfDataCounterValuesSession represents Horizon Blast Session Counters.
type perfDataCounterValuesSession struct {
	Name string

	AutomaticReconnectCount              float64 `perfdata:"Automatic Reconnect Count"`
	CumulativeReceivedBytesOverUDP       float64 `perfdata:"Cumulative Received Bytes over UDP"`
	CumulativeTransmittedBytesOverUDP    float64 `perfdata:"Cumulative Transmitted Bytes over UDP"`
	CumulativeReceivedBytesOverTCP       float64 `perfdata:"Cumulative Received Bytes over TCP"`
	CumulativeTransmittedBytesOverTCP    float64 `perfdata:"Cumulative Transmitted Bytes over TCP"`
	InstantaneousReceivedBytesOverUDP    float64 `perfdata:"Instantaneous Received Bytes over UDP"`
	InstantaneousTransmittedBytesOverUDP float64 `perfdata:"Instantaneous Transmitted Bytes over UDP"`
	InstantaneousReceivedBytesOverTCP    float64 `perfdata:"Instantaneous Received Bytes over TCP"`
	InstantaneousTransmittedBytesOverTCP float64 `perfdata:"Instantaneous Transmitted Bytes over TCP"`
	ReceivedPackets                      float64 `perfdata:"Received Packets"`
	TransmittedPackets                   float64 `perfdata:"Transmitted Packets"`
	ReceivedBytes                        float64 `perfdata:"Received Bytes"`
	TransmittedBytes                     float64 `perfdata:"Transmitted Bytes"`
	JitterUplink                         float64 `perfdata:"Jitter (Uplink)"`
	RTT                                  float64 `perfdata:"RTT"`
	PacketLossUplink                     float64 `perfdata:"Packet Loss (Uplink)"`
	EstimatedBandwidthUplink             float64 `perfdata:"Estimated Bandwidth (Uplink)"`
}

// perfDataCounterValuesImaging represents Horizon Blast Imaging Counters.
type perfDataCounterValuesImaging struct {
	Name string

	EncoderType           float64 `perfdata:"Encoder Type"`
	TotalDirtyFrames      float64 `perfdata:"Total dirty frames"`
	TotalPoll             float64 `perfdata:"Total poll"`
	TotalFBC              float64 `perfdata:"Total FBC"`
	TotalFrames           float64 `perfdata:"Total frames"`
	DirtyFramesPerSecond  float64 `perfdata:"Dirty frames per second"`
	PollRate              float64 `perfdata:"Poll Rate"`
	FBCRate               float64 `perfdata:"FBC Rate"`
	FramesPerSecond       float64 `perfdata:"Frames per second"`
	OutQueueingTimeUs     float64 `perfdata:"Out Queueing time (us)"`
	InboundBandwidthKbps  float64 `perfdata:"Inbound Bandwidth (Kbps)"`
	OutboundBandwidthKbps float64 `perfdata:"Outbound Bandwidth (Kbps)"`
	ReceivedPackets       float64 `perfdata:"Received Packets"`
	TransmittedPackets    float64 `perfdata:"Transmitted Packets"`
	ReceivedBytes         float64 `perfdata:"Received Bytes"`
	TransmittedBytes      float64 `perfdata:"Transmitted Bytes"`
}

// perfDataCounterValuesAudio represents Horizon Blast Audio Counters.
type perfDataCounterValuesAudio struct {
	Name string

	OutQueueingTimeUs     float64 `perfdata:"Out Queueing time (us)"`
	InboundBandwidthKbps  float64 `perfdata:"Inbound Bandwidth (Kbps)"`
	OutboundBandwidthKbps float64 `perfdata:"Outbound Bandwidth (Kbps)"`
	ReceivedPackets       float64 `perfdata:"Received Packets"`
	TransmittedPackets    float64 `perfdata:"Transmitted Packets"`
	ReceivedBytes         float64 `perfdata:"Received Bytes"`
	TransmittedBytes      float64 `perfdata:"Transmitted Bytes"`
}

// perfDataCounterValuesCDR represents Horizon Blast CDR Counters.
type perfDataCounterValuesCDR struct {
	Name string

	OutQueueingTimeUs     float64 `perfdata:"Out Queueing time (us)"`
	InboundBandwidthKbps  float64 `perfdata:"Inbound Bandwidth (Kbps)"`
	OutboundBandwidthKbps float64 `perfdata:"Outbound Bandwidth (Kbps)"`
	ReceivedPackets       float64 `perfdata:"Received Packets"`
	TransmittedPackets    float64 `perfdata:"Transmitted Packets"`
	ReceivedBytes         float64 `perfdata:"Received Bytes"`
	TransmittedBytes      float64 `perfdata:"Transmitted Bytes"`
}

// perfDataCounterValuesClipboard represents Horizon Blast Clipboard Counters.
type perfDataCounterValuesClipboard struct {
	Name string

	OutQueueingTimeUs     float64 `perfdata:"Out Queueing time (us)"`
	InboundBandwidthKbps  float64 `perfdata:"Inbound Bandwidth (Kbps)"`
	OutboundBandwidthKbps float64 `perfdata:"Outbound Bandwidth (Kbps)"`
	ReceivedPackets       float64 `perfdata:"Received Packets"`
	TransmittedPackets    float64 `perfdata:"Transmitted Packets"`
	ReceivedBytes         float64 `perfdata:"Received Bytes"`
	TransmittedBytes      float64 `perfdata:"Transmitted Bytes"`
}

// perfDataCounterValuesHTML5MMR represents Horizon Blast HTML5 MMR Counters.
type perfDataCounterValuesHTML5MMR struct {
	Name string

	OutQueueingTimeUs     float64 `perfdata:"Out Queueing time (us)"`
	InboundBandwidthKbps  float64 `perfdata:"Inbound Bandwidth (Kbps)"`
	OutboundBandwidthKbps float64 `perfdata:"Outbound Bandwidth (Kbps)"`
	ReceivedPackets       float64 `perfdata:"Received Packets"`
	TransmittedPackets    float64 `perfdata:"Transmitted Packets"`
	ReceivedBytes         float64 `perfdata:"Received Bytes"`
	TransmittedBytes      float64 `perfdata:"Transmitted Bytes"`
}

// perfDataCounterValuesOtherFeature represents Horizon Blast Other Feature Counters.
type perfDataCounterValuesOtherFeature struct {
	Name string

	OutQueueingTimeUs     float64 `perfdata:"Out Queueing time (us)"`
	InboundBandwidthKbps  float64 `perfdata:"Inbound Bandwidth (Kbps)"`
	OutboundBandwidthKbps float64 `perfdata:"Outbound Bandwidth (Kbps)"`
	ReceivedPackets       float64 `perfdata:"Received Packets"`
	TransmittedPackets    float64 `perfdata:"Transmitted Packets"`
	ReceivedBytes         float64 `perfdata:"Received Bytes"`
	TransmittedBytes      float64 `perfdata:"Transmitted Bytes"`
}

// perfDataCounterValuesPrinting represents Horizon Blast Printing Counters.
type perfDataCounterValuesPrinting struct {
	Name string

	OutQueueingTimeUs     float64 `perfdata:"Out Queueing time (us)"`
	InboundBandwidthKbps  float64 `perfdata:"Inbound Bandwidth (Kbps)"`
	OutboundBandwidthKbps float64 `perfdata:"Outbound Bandwidth (Kbps)"`
	ReceivedPackets       float64 `perfdata:"Received Packets"`
	TransmittedPackets    float64 `perfdata:"Transmitted Packets"`
	ReceivedBytes         float64 `perfdata:"Received Bytes"`
	TransmittedBytes      float64 `perfdata:"Transmitted Bytes"`
}

// perfDataCounterValuesRdeServer represents Horizon Blast RdeServer Counters.
type perfDataCounterValuesRdeServer struct {
	Name string

	OutQueueingTimeUs     float64 `perfdata:"Out Queueing time (us)"`
	InboundBandwidthKbps  float64 `perfdata:"Inbound Bandwidth (Kbps)"`
	OutboundBandwidthKbps float64 `perfdata:"Outbound Bandwidth (Kbps)"`
	ReceivedPackets       float64 `perfdata:"Received Packets"`
	TransmittedPackets    float64 `perfdata:"Transmitted Packets"`
	ReceivedBytes         float64 `perfdata:"Received Bytes"`
	TransmittedBytes      float64 `perfdata:"Transmitted Bytes"`
}

// perfDataCounterValuesRTAV represents Horizon Blast RTAV Counters.
type perfDataCounterValuesRTAV struct {
	Name string

	OutQueueingTimeUs     float64 `perfdata:"Out Queueing time (us)"`
	InboundBandwidthKbps  float64 `perfdata:"Inbound Bandwidth (Kbps)"`
	OutboundBandwidthKbps float64 `perfdata:"Outbound Bandwidth (Kbps)"`
	ReceivedPackets       float64 `perfdata:"Received Packets"`
	TransmittedPackets    float64 `perfdata:"Transmitted Packets"`
	ReceivedBytes         float64 `perfdata:"Received Bytes"`
	TransmittedBytes      float64 `perfdata:"Transmitted Bytes"`
}

// perfDataCounterValuesSDR represents Horizon Blast SDR Counters.
type perfDataCounterValuesSDR struct {
	Name string

	OutQueueingTimeUs     float64 `perfdata:"Out Queueing time (us)"`
	InboundBandwidthKbps  float64 `perfdata:"Inbound Bandwidth (Kbps)"`
	OutboundBandwidthKbps float64 `perfdata:"Outbound Bandwidth (Kbps)"`
	ReceivedPackets       float64 `perfdata:"Received Packets"`
	TransmittedPackets    float64 `perfdata:"Transmitted Packets"`
	ReceivedBytes         float64 `perfdata:"Received Bytes"`
	TransmittedBytes      float64 `perfdata:"Transmitted Bytes"`
}

// perfDataCounterValuesSerialPortScanner represents Horizon Blast Serial Port and Scanner Counters.
type perfDataCounterValuesSerialPortScanner struct {
	Name string

	OutQueueingTimeUs     float64 `perfdata:"Out Queueing time (us)"`
	InboundBandwidthKbps  float64 `perfdata:"Inbound Bandwidth (Kbps)"`
	OutboundBandwidthKbps float64 `perfdata:"Outbound Bandwidth (Kbps)"`
	ReceivedPackets       float64 `perfdata:"Received Packets"`
	TransmittedPackets    float64 `perfdata:"Transmitted Packets"`
	ReceivedBytes         float64 `perfdata:"Received Bytes"`
	TransmittedBytes      float64 `perfdata:"Transmitted Bytes"`
}

// perfDataCounterValuesSmartCard represents Horizon Blast Smart Card Counters.
type perfDataCounterValuesSmartCard struct {
	Name string

	OutQueueingTimeUs     float64 `perfdata:"Out Queueing time (us)"`
	InboundBandwidthKbps  float64 `perfdata:"Inbound Bandwidth (Kbps)"`
	OutboundBandwidthKbps float64 `perfdata:"Outbound Bandwidth (Kbps)"`
	ReceivedPackets       float64 `perfdata:"Received Packets"`
	TransmittedPackets    float64 `perfdata:"Transmitted Packets"`
	ReceivedBytes         float64 `perfdata:"Received Bytes"`
	TransmittedBytes      float64 `perfdata:"Transmitted Bytes"`
}

// perfDataCounterValuesUSB represents Horizon Blast USB Counters.
type perfDataCounterValuesUSB struct {
	Name string

	OutQueueingTimeUs     float64 `perfdata:"Out Queueing time (us)"`
	InboundBandwidthKbps  float64 `perfdata:"Inbound Bandwidth (Kbps)"`
	OutboundBandwidthKbps float64 `perfdata:"Outbound Bandwidth (Kbps)"`
	ReceivedPackets       float64 `perfdata:"Received Packets"`
	TransmittedPackets    float64 `perfdata:"Transmitted Packets"`
	ReceivedBytes         float64 `perfdata:"Received Bytes"`
	TransmittedBytes      float64 `perfdata:"Transmitted Bytes"`
}

// perfDataCounterValuesViewScanner represents Horizon Blast View Scanner Counters.
type perfDataCounterValuesViewScanner struct {
	Name string

	OutQueueingTimeUs     float64 `perfdata:"Out Queueing time (us)"`
	InboundBandwidthKbps  float64 `perfdata:"Inbound Bandwidth (Kbps)"`
	OutboundBandwidthKbps float64 `perfdata:"Outbound Bandwidth (Kbps)"`
	ReceivedPackets       float64 `perfdata:"Received Packets"`
	TransmittedPackets    float64 `perfdata:"Transmitted Packets"`
	ReceivedBytes         float64 `perfdata:"Received Bytes"`
	TransmittedBytes      float64 `perfdata:"Transmitted Bytes"`
}

// perfDataCounterValuesWindowsMediaMMR represents Horizon Blast Windows Media MMR Counters.
type perfDataCounterValuesWindowsMediaMMR struct {
	Name string

	OutQueueingTimeUs     float64 `perfdata:"Out Queueing time (us)"`
	InboundBandwidthKbps  float64 `perfdata:"Inbound Bandwidth (Kbps)"`
	OutboundBandwidthKbps float64 `perfdata:"Outbound Bandwidth (Kbps)"`
	ReceivedPackets       float64 `perfdata:"Received Packets"`
	TransmittedPackets    float64 `perfdata:"Transmitted Packets"`
	ReceivedBytes         float64 `perfdata:"Received Bytes"`
	TransmittedBytes      float64 `perfdata:"Transmitted Bytes"`
}

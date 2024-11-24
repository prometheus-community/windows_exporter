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

const (
	BaseTCPRTT               = "Base TCP RTT"
	BaseUDPRTT               = "Base UDP RTT"
	CurrentTCPBandwidth      = "Current TCP Bandwidth"
	CurrentTCPRTT            = "Current TCP RTT"
	CurrentUDPBandwidth      = "Current UDP Bandwidth"
	CurrentUDPRTT            = "Current UDP RTT"
	TotalReceivedBytes       = "Total Received Bytes"
	TotalSentBytes           = "Total Sent Bytes"
	UDPPacketsReceivedPersec = "UDP Packets Received/sec"
	UDPPacketsSentPersec     = "UDP Packets Sent/sec"
	FECRate                  = "FEC rate"
	LossRate                 = "Loss rate"
	RetransmissionRate       = "Retransmission rate"

	AverageEncodingTime                                = "Average Encoding Time"
	FrameQuality                                       = "Frame Quality"
	FramesSkippedPerSecondInsufficientClientResources  = "Frames Skipped/Second - Insufficient Server Resources"
	FramesSkippedPerSecondInsufficientNetworkResources = "Frames Skipped/Second - Insufficient Network Resources"
	FramesSkippedPerSecondInsufficientServerResources  = "Frames Skipped/Second - Insufficient Client Resources"
	GraphicsCompressionratio                           = "Graphics Compression ratio"
	InputFramesPerSecond                               = "Input Frames/Second"
	OutputFramesPerSecond                              = "Output Frames/Second"
	SourceFramesPerSecond                              = "Source Frames/Second"
)

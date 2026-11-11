// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package remote_fx

type perfDataCounterValuesNetwork struct {
	Name string

	BaseTCPRTT               float64 `perfdata:"Base TCP RTT"`
	BaseUDPRTT               float64 `perfdata:"Base UDP RTT"`
	CurrentTCPBandwidth      float64 `perfdata:"Current TCP Bandwidth"`
	CurrentTCPRTT            float64 `perfdata:"Current TCP RTT"`
	CurrentUDPBandwidth      float64 `perfdata:"Current UDP Bandwidth"`
	CurrentUDPRTT            float64 `perfdata:"Current UDP RTT"`
	TotalReceivedBytes       float64 `perfdata:"Total Received Bytes"`
	TotalSentBytes           float64 `perfdata:"Total Sent Bytes"`
	UDPPacketsReceivedPersec float64 `perfdata:"UDP Packets Received/sec"`
	UDPPacketsSentPersec     float64 `perfdata:"UDP Packets Sent/sec"`
	// FEC Rate, Loss Rate and Retransmission Rate are PERF_SAMPLE_FRACTION
	// counters: the first value is a cumulative numerator, the second value
	// the cumulative base. Neither is a percentage on its own.
	FECRate                float64 `perfdata:"FEC rate"`
	FECRateBase            float64 `perfdata:"FEC rate,secondvalue"`
	LossRate               float64 `perfdata:"Loss rate"`
	LossRateBase           float64 `perfdata:"Loss rate,secondvalue"`
	RetransmissionRate     float64 `perfdata:"Retransmission rate"`
	RetransmissionRateBase float64 `perfdata:"Retransmission rate,secondvalue"`
}

type perfDataCounterValuesGraphics struct {
	Name string

	AverageEncodingTime                                float64 `perfdata:"Average Encoding Time"`
	FrameQuality                                       float64 `perfdata:"Frame Quality"`
	FramesSkippedPerSecondInsufficientClientResources  float64 `perfdata:"Frames Skipped/Second - Insufficient Client Resources"`
	FramesSkippedPerSecondInsufficientNetworkResources float64 `perfdata:"Frames Skipped/Second - Insufficient Network Resources"`
	FramesSkippedPerSecondInsufficientServerResources  float64 `perfdata:"Frames Skipped/Second - Insufficient Server Resources"`
	GraphicsCompressionratio                           float64 `perfdata:"Graphics Compression ratio"`
	InputFramesPerSecond                               float64 `perfdata:"Input Frames/Second"`
	OutputFramesPerSecond                              float64 `perfdata:"Output Frames/Second"`
	SourceFramesPerSecond                              float64 `perfdata:"Source Frames/Second"`
}

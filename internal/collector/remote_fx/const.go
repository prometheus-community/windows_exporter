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
	FECRate                  = "Forward Error Correction (FEC) percentage"
	LossRate                 = "Loss percentage"
	RetransmissionRate       = "Percentage of packets that have been retransmitted"

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

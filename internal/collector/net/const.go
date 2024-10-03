package net

const (
	BytesReceivedPerSec      = "Bytes Received/sec"
	BytesSentPerSec          = "Bytes Sent/sec"
	BytesTotalPerSec         = "Bytes Total/sec"
	OutputQueueLength        = "Output Queue Length"
	PacketsOutboundDiscarded = "Packets Outbound Discarded"
	PacketsOutboundErrors    = "Packets Outbound Errors"
	PacketsPerSec            = "Packets/sec"
	PacketsReceivedDiscarded = "Packets Received Discarded"
	PacketsReceivedErrors    = "Packets Received Errors"
	PacketsReceivedPerSec    = "Packets Received/sec"
	PacketsReceivedUnknown   = "Packets Received Unknown"
	PacketsSentPerSec        = "Packets Sent/sec"
	CurrentBandwidth         = "Current Bandwidth"
)

// Win32_PerfRawData_Tcpip_NetworkInterface docs:
// - https://technet.microsoft.com/en-us/security/aa394340(v=vs.80)
type perflibNetworkInterface struct {
	BytesReceivedPerSec      float64 `perflib:"Bytes Received/sec"`
	BytesSentPerSec          float64 `perflib:"Bytes Sent/sec"`
	BytesTotalPerSec         float64 `perflib:"Bytes Total/sec"`
	Name                     string
	OutputQueueLength        float64 `perflib:"Output Queue Length"`
	PacketsOutboundDiscarded float64 `perflib:"Packets Outbound Discarded"`
	PacketsOutboundErrors    float64 `perflib:"Packets Outbound Errors"`
	PacketsPerSec            float64 `perflib:"Packets/sec"`
	PacketsReceivedDiscarded float64 `perflib:"Packets Received Discarded"`
	PacketsReceivedErrors    float64 `perflib:"Packets Received Errors"`
	PacketsReceivedPerSec    float64 `perflib:"Packets Received/sec"`
	PacketsReceivedUnknown   float64 `perflib:"Packets Received Unknown"`
	PacketsSentPerSec        float64 `perflib:"Packets Sent/sec"`
	CurrentBandwidth         float64 `perflib:"Current Bandwidth"`
}

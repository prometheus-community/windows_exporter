//go:build windows

package tcp

// Win32_PerfRawData_Tcpip_TCPv4 docs
// - https://msdn.microsoft.com/en-us/library/aa394341(v=vs.85).aspx
// The TCPv6 performance object uses the same fields.
// https://learn.microsoft.com/en-us/dotnet/api/system.net.networkinformation.tcpstate?view=net-8.0.
const (
	connectionFailures          = "Connection Failures"
	connectionsActive           = "Connections Active"
	connectionsEstablished      = "Connections Established"
	connectionsPassive          = "Connections Passive"
	connectionsReset            = "Connections Reset"
	segmentsPerSec              = "Segments/sec"
	segmentsReceivedPerSec      = "Segments Received/sec"
	segmentsRetransmittedPerSec = "Segments Retransmitted/sec"
	segmentsSentPerSec          = "Segments Sent/sec"
)

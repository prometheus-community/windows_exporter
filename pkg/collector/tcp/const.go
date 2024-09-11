package tcp

// Win32_PerfRawData_Tcpip_TCPv4 docs
// - https://msdn.microsoft.com/en-us/library/aa394341(v=vs.85).aspx
// The TCPv6 performance object uses the same fields.
const (
	ConnectionFailures          = "Connection Failures"
	ConnectionsActive           = "Connections Active"
	ConnectionsEstablished      = "Connections Established"
	ConnectionsPassive          = "Connections Passive"
	ConnectionsReset            = "Connections Reset"
	SegmentsPersec              = "Segments/sec"
	SegmentsReceivedPersec      = "Segments Received/sec"
	SegmentsRetransmittedPersec = "Segments Retransmitted/sec"
	SegmentsSentPersec          = "Segments Sent/sec"
)

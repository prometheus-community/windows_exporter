package tcp

// Win32_PerfRawData_Tcpip_TCPv4 docs
// - https://msdn.microsoft.com/en-us/library/aa394341(v=vs.85).aspx
// The TCPv6 performance object uses the same fields.
// https://learn.microsoft.com/en-us/dotnet/api/system.net.networkinformation.tcpstate?view=net-8.0
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

    TCPTableClass               = 5

	TCPStateClosed              = 1
    TCPStateListening           = 2
    TCPStateSynSent             = 3
    TCPStateSynRcvd             = 4
    TCPStateEstablished         = 5
    TCPStateFinWait1            = 6
    TCPStateFinWait2            = 7
    TCPStateCloseWait           = 8
    TCPStateClosing             = 9
    TCPStateLastAck             = 10
    TCPStateTimeWait            = 11
    TCPStateDeleteTcb           = 12
)

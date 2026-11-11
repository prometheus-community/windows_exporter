// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package tcp

// Win32_PerfRawData_Tcpip_TCPv4 docs
// - https://msdn.microsoft.com/en-us/library/aa394341(v=vs.85).aspx
// The TCPv6 performance object uses the same fields.
// https://learn.microsoft.com/en-us/dotnet/api/system.net.networkinformation.tcpstate?view=net-8.0.
type perfDataCounterValues struct {
	ConnectionFailures          float64 `perfdata:"Connection Failures"`
	ConnectionsActive           float64 `perfdata:"Connections Active"`
	ConnectionsEstablished      float64 `perfdata:"Connections Established"`
	ConnectionsPassive          float64 `perfdata:"Connections Passive"`
	ConnectionsReset            float64 `perfdata:"Connections Reset"`
	SegmentsPerSec              float64 `perfdata:"Segments/sec"`
	SegmentsReceivedPerSec      float64 `perfdata:"Segments Received/sec"`
	SegmentsRetransmittedPerSec float64 `perfdata:"Segments Retransmitted/sec"`
	SegmentsSentPerSec          float64 `perfdata:"Segments Sent/sec"`
}

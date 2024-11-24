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

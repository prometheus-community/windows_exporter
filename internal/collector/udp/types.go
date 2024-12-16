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

package udp

// The TCPv6 performance object uses the same fields.
// https://learn.microsoft.com/en-us/dotnet/api/system.net.networkinformation.tcpstate?view=net-8.0.
type perfDataCounterValues struct {
	DatagramsNoPortPerSec   float64 `perfdata:"Datagrams No Port/sec"`
	DatagramsReceivedPerSec float64 `perfdata:"Datagrams Received/sec"`
	DatagramsReceivedErrors float64 `perfdata:"Datagrams Received Errors"`
	DatagramsSentPerSec     float64 `perfdata:"Datagrams Sent/sec"`
}

// Datagrams No Port/sec is the rate of received UDP datagrams for which there was no application at the destination port.
// Datagrams Received Errors is the number of received UDP datagrams that could not be delivered for reasons other than the lack of an application at the destination port.
// Datagrams Received/sec is the rate at which UDP datagrams are delivered to UDP users.
// Datagrams Sent/sec is the rate at which UDP datagrams are sent from the entity.

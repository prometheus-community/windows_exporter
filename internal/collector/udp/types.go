// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

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

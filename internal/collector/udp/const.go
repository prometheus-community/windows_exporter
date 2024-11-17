//go:build windows

package udp

// The TCPv6 performance object uses the same fields.
// https://learn.microsoft.com/en-us/dotnet/api/system.net.networkinformation.tcpstate?view=net-8.0.
const (
	datagramsNoPortPerSec   = "Datagrams No Port/sec"
	datagramsReceivedPerSec = "Datagrams Received/sec"
	datagramsReceivedErrors = "Datagrams Received Errors"
	datagramsSentPerSec     = "Datagrams Sent/sec"
)

// Datagrams No Port/sec is the rate of received UDP datagrams for which there was no application at the destination port.
// Datagrams Received Errors is the number of received UDP datagrams that could not be delivered for reasons other than the lack of an application at the destination port.
// Datagrams Received/sec is the rate at which UDP datagrams are delivered to UDP users.
// Datagrams Sent/sec is the rate at which UDP datagrams are sent from the entity.

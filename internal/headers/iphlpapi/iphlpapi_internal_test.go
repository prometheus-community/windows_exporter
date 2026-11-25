// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package iphlpapi

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"
)

// networkPort returns port as GetExtendedTcpTable stores it, in network byte
// order in the low 16 bits.
func networkPort(port uint16) BigEndianUint32 {
	var b [2]byte

	binary.BigEndian.PutUint16(b[:], port)

	return BigEndianUint32(binary.LittleEndian.Uint16(b[:]))
}

func TestSummarizeTCPTable(t *testing.T) {
	t.Parallel()

	table := []MIB_TCPROW_OWNER_PID{
		{dwState: TCPStateListening, dwLocalPort: networkPort(445)},
		// A listener and an accepted connection share port 49152.
		{dwState: TCPStateListening, dwLocalPort: networkPort(49152)},
		{dwState: TCPStateEstablished, dwLocalPort: networkPort(49152)},
		{dwState: TCPStateTimeWait, dwLocalPort: networkPort(50000)},
		{dwState: TCPStateEstablished, dwLocalPort: networkPort(65535)},
		// The port just below the range start is outside.
		{dwState: TCPStateEstablished, dwLocalPort: networkPort(49151)},
	}

	var ports PortSet

	stats := summarizeTCPTable(table, PortRange{Start: 49152, NumberOfPorts: 16384}, &ports)
	require.Equal(t, map[MIB_TCP_STATE]uint32{
		TCPStateListening:   2,
		TCPStateEstablished: 3,
		TCPStateTimeWait:    1,
	}, stats.States)
	require.EqualValues(t, 3, ports.Len())

	// A second table, as for the other address family, adds only new ports.
	summarizeTCPTable([]MIB_TCP6ROW_OWNER_PID{
		{dwState: TCPStateEstablished, dwLocalPort: networkPort(50000)},
		{dwState: TCPStateEstablished, dwLocalPort: networkPort(50001)},
	}, PortRange{Start: 49152, NumberOfPorts: 16384}, &ports)
	require.EqualValues(t, 4, ports.Len())

	var empty PortSet

	summarizeTCPTable(table, PortRange{}, &empty)
	require.Zero(t, empty.Len())

	// A nil set only counts the states.
	require.Len(t, summarizeTCPTable(table, PortRange{}, nil).States, 3)
}

func TestPortRangeContains(t *testing.T) {
	t.Parallel()

	r := PortRange{Start: 64000, NumberOfPorts: 1536}
	require.False(t, r.contains(63999))
	require.True(t, r.contains(64000))
	require.True(t, r.contains(65535))
}

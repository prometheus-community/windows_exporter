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
		// The last port inside a range that ends at 49151 is outside.
		{dwState: TCPStateEstablished, dwLocalPort: networkPort(49151)},
	}

	stats := summarizeTCPTable(table, PortRange{Start: 49152, NumberOfPorts: 16384})
	require.Equal(t, map[MIB_TCP_STATE]uint32{
		TCPStateListening:   2,
		TCPStateEstablished: 3,
		TCPStateTimeWait:    1,
	}, stats.States)
	require.EqualValues(t, 3, stats.PortsInRange)

	require.Zero(t, summarizeTCPTable(table, PortRange{}).PortsInRange)
}

func TestPortRangeContains(t *testing.T) {
	t.Parallel()

	r := PortRange{Start: 64000, NumberOfPorts: 1536}
	require.False(t, r.contains(63999))
	require.True(t, r.contains(64000))
	require.True(t, r.contains(65535))
}

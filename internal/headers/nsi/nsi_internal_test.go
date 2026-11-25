// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package nsi

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
)

func TestNPIModuleIDLayout(t *testing.T) {
	t.Parallel()

	require.EqualValues(t, 24, unsafe.Sizeof(npiModuleID{}))
	require.EqualValues(t, 4, unsafe.Offsetof(npiModuleID{}.idType))
	require.EqualValues(t, 8, unsafe.Offsetof(npiModuleID{}.moduleID))
}

func TestParsePortRange(t *testing.T) {
	t.Parallel()

	// Bytes read on Windows Server 2025 for the default range and after
	// netsh int ipv4 set dynamicport tcp start=49153 num=16383.
	for _, tc := range []struct {
		buf          [4]byte
		start, count uint16
	}{
		{buf: [4]byte{0xc0, 0x00, 0x00, 0x40}, start: 49152, count: 16384},
		{buf: [4]byte{0xc0, 0x01, 0xff, 0x3f}, start: 49153, count: 16383},
	} {
		start, count := parsePortRange(tc.buf)
		require.Equal(t, tc.start, start)
		require.Equal(t, tc.count, count)
	}
}

func TestTCPDynamicPortRange(t *testing.T) {
	t.Parallel()

	require.NoError(t, Load())

	start, count, err := TCPDynamicPortRange()
	require.NoError(t, err)
	// The bounds tcpip.sys enforces when it loads or changes the range.
	require.GreaterOrEqual(t, start, uint16(1024))
	require.GreaterOrEqual(t, count, uint16(255))
	require.LessOrEqual(t, uint32(start)+uint32(count), uint32(65536))
}

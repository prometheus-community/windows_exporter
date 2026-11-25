// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package nsi

import (
	"sync"
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
		buf  [4]byte
		want PortRange
	}{
		{buf: [4]byte{0xc0, 0x00, 0x00, 0x40}, want: PortRange{Start: 49152, NumberOfPorts: 16384}},
		{buf: [4]byte{0xc0, 0x01, 0xff, 0x3f}, want: PortRange{Start: 49153, NumberOfPorts: 16383}},
	} {
		require.Equal(t, tc.want, parsePortRange(tc.buf))
	}
}

func TestDynamicPortRangeUnsupportedProtocol(t *testing.T) {
	t.Parallel()

	_, err := DynamicPortRange(Protocol(-1))
	require.Error(t, err)

	_, err = DynamicPortRange(Protocol(len(dynamicPortRangeObjects)))
	require.Error(t, err)
}

func TestDynamicPortRange(t *testing.T) {
	t.Parallel()

	require.NoError(t, Load())

	for _, protocol := range []Protocol{ProtocolTCP, ProtocolUDP} {
		r, err := DynamicPortRange(protocol)
		require.NoError(t, err)

		// The bounds tcpip.sys enforces when it loads or changes the range.
		require.GreaterOrEqual(t, r.Start, uint16(1024))
		require.GreaterOrEqual(t, r.NumberOfPorts, uint16(255))
		require.LessOrEqual(t, uint32(r.Start)+uint32(r.NumberOfPorts), uint32(65536))
	}
}

func TestDynamicPortRangeConcurrent(t *testing.T) {
	t.Parallel()

	require.NoError(t, Load())

	want, err := DynamicPortRange(ProtocolTCP)
	require.NoError(t, err)

	var wg sync.WaitGroup

	for range 8 {
		wg.Go(func() {
			for range 100 {
				got, err := DynamicPortRange(ProtocolTCP)
				if err != nil || got != want {
					t.Errorf("DynamicPortRange = %v, %v; want %v", got, err, want)

					return
				}
			}
		})
	}

	wg.Wait()
}

func BenchmarkDynamicPortRange(b *testing.B) {
	require.NoError(b, Load())

	b.ReportAllocs()

	for b.Loop() {
		_, _ = DynamicPortRange(ProtocolTCP)
	}
}

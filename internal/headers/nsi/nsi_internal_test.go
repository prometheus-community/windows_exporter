// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package nsi

import (
	"encoding/binary"
	"slices"
	"sync"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
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

	_, err = DynamicPortRange(Protocol(len(objects)))
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

	_, err := DynamicPortRange(ProtocolTCP)
	require.NoError(b, err)

	b.ReportAllocs()

	for b.Loop() {
		_, _ = DynamicPortRange(ProtocolTCP)
	}
}

func TestParseExcludedPortRanges(t *testing.T) {
	t.Parallel()

	// Keys read on Windows Server 2025: 50000-50009 and 1080 (one port). A
	// trailing partial key is ignored.
	raw := []byte{0x50, 0xc3, 0x0a, 0x00, 0x38, 0x04, 0x01, 0x00, 0xff}

	require.Equal(t, []PortRange{
		{Start: 50000, NumberOfPorts: 10},
		{Start: 1080, NumberOfPorts: 1},
	}, parseExcludedPortRanges(raw))
	require.Empty(t, parseExcludedPortRanges(nil))
}

func TestParseBoundTCPPorts(t *testing.T) {
	t.Parallel()

	key := func(family uint16, port uint16) []byte {
		k := make([]byte, tcpBoundEndpointKeySize)
		binary.LittleEndian.PutUint16(k[0:2], family)
		binary.BigEndian.PutUint16(k[2:4], port)

		return k
	}

	raw := slices.Concat(
		key(windows.AF_INET, 49152),
		key(windows.AF_INET6, 65535),
		key(0, 1234), // Unknown family, skipped.
		[]byte{0x02, 0x00},
	)

	require.Equal(t, []uint16{49152, 65535}, parseBoundTCPPorts(raw))
}

func TestExcludedPortRanges(t *testing.T) {
	t.Parallel()

	require.NoError(t, Load())

	for _, protocol := range []Protocol{ProtocolTCP, ProtocolUDP} {
		ranges, err := ExcludedPortRanges(protocol)
		require.NoError(t, err)

		for _, r := range ranges {
			require.NotZero(t, r.NumberOfPorts)
			require.LessOrEqual(t, uint32(r.Start)+uint32(r.NumberOfPorts), uint32(65536))
		}
	}
}

func TestBoundTCPPorts(t *testing.T) {
	t.Parallel()

	require.NoError(t, Load())

	// A socket that is bound to port 0 but neither listens nor connects is
	// only visible in the bound endpoint table.
	fd, err := windows.Socket(windows.AF_INET, windows.SOCK_STREAM, windows.IPPROTO_TCP)
	require.NoError(t, err)

	t.Cleanup(func() { _ = windows.Closesocket(fd) })

	require.NoError(t, windows.Bind(fd, &windows.SockaddrInet4{Addr: [4]byte{127, 0, 0, 1}}))

	sa, err := windows.Getsockname(fd)
	require.NoError(t, err)

	port := uint16(sa.(*windows.SockaddrInet4).Port) //nolint:forcetypeassert

	ports, err := BoundTCPPorts()
	require.NoError(t, err)
	require.Contains(t, ports, port)
}

func TestPortRangeUsage(t *testing.T) {
	t.Parallel()

	r := PortRange{Start: 65530, NumberOfPorts: 6}
	exclusions := []PortRange{
		{Start: 65000, NumberOfPorts: 531}, // Overlaps 65530.
		{Start: 65535, NumberOfPorts: 1},
		{Start: 1080, NumberOfPorts: 1}, // Outside the range.
	}
	inUse := func(port uint16) bool { return port == 65531 || port == 65535 || port == 1 }

	excluded, used := r.Usage(exclusions, inUse)
	require.EqualValues(t, 2, excluded)
	// 65535 is excluded and counts only once.
	require.EqualValues(t, 1, used)
}

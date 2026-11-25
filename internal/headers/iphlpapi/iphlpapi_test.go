// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package iphlpapi_test

import (
	"net"
	"os"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/headers/iphlpapi"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func TestGetTCPConnectionStats(t *testing.T) {
	t.Parallel()

	var listenConf net.ListenConfig

	listener, err := listenConf.Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, listener.Close())
	})

	tcpAddr, ok := listener.Addr().(*net.TCPAddr)
	require.True(t, ok)

	port := uint16(tcpAddr.Port)

	var ports iphlpapi.PortSet

	stats, err := iphlpapi.GetTCPConnectionStats(windows.AF_INET, iphlpapi.PortRange{Start: port, NumberOfPorts: 1}, &ports)
	require.NoError(t, err)
	require.NotZero(t, stats.States[iphlpapi.TCPStateListening])
	require.EqualValues(t, 1, ports.Len())
}

func TestGetOwnerPIDOfTCPPort(t *testing.T) {
	t.Parallel()

	var listenConf net.ListenConfig

	lister, err := listenConf.Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, lister.Close())
	})

	tcpAddr, ok := lister.Addr().(*net.TCPAddr)
	require.True(t, ok)

	pid, err := iphlpapi.GetOwnerPIDOfTCPPort(windows.AF_INET, uint16(tcpAddr.Port))
	require.NoError(t, err)
	require.EqualValues(t, os.Getpid(), pid)
}

func BenchmarkGetTCPConnectionStats(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		var ports iphlpapi.PortSet

		_, _ = iphlpapi.GetTCPConnectionStats(windows.AF_INET, iphlpapi.PortRange{Start: 49152, NumberOfPorts: 16384}, &ports)
	}
}

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

func TestGetTCPConnectionStates(t *testing.T) {
	t.Parallel()

	pid, err := iphlpapi.GetTCPConnectionStates(windows.AF_INET)
	require.NoError(t, err)
	require.NotEmpty(t, pid)
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

func BenchmarkGetTCPConnectionStates(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		_, _ = iphlpapi.GetTCPConnectionStates(windows.AF_INET)
	}
}

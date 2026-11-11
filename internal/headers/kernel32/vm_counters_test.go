// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package kernel32_test

import (
	"testing"
	"unsafe"

	"github.com/prometheus-community/windows_exporter/internal/headers/kernel32"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func TestVMCountersEX2Layout(t *testing.T) {
	t.Parallel()

	var counters kernel32.VM_COUNTERS_EX2
	require.Equal(t, uintptr(112), unsafe.Sizeof(counters))
	require.Equal(t, uintptr(88), unsafe.Offsetof(counters.PrivateUsage))
	require.Equal(t, uintptr(96), unsafe.Offsetof(counters.PrivateWorkingSetSize))
	require.Equal(t, uintptr(104), unsafe.Offsetof(counters.SharedCommitUsage))
}

func TestVMCountersPrivateWorkingSet(t *testing.T) {
	t.Parallel()

	const allocation = 64 << 20

	address, err := windows.VirtualAlloc(0, allocation, windows.MEM_RESERVE|windows.MEM_COMMIT, windows.PAGE_READWRITE)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, windows.VirtualFree(address, 0, windows.MEM_RELEASE)) })

	// Leave the committed pages untouched so they need not be resident.
	var (
		counters kernel32.VM_COUNTERS_EX2
		returned uint32
	)

	err = windows.NtQueryInformationProcess(windows.CurrentProcess(), windows.ProcessVmCounters,
		unsafe.Pointer(&counters), uint32(unsafe.Sizeof(counters)), &returned)
	require.NoError(t, err)
	require.Equal(t, uint32(unsafe.Sizeof(counters)), returned)
	require.Greater(t, counters.PrivateWorkingSetSize, uintptr(0))
	require.Greater(t, counters.PrivateUsage, counters.PrivateWorkingSetSize+allocation/2)
}

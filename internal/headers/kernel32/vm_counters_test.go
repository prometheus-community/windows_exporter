// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

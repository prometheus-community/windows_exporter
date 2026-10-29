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

package w32time

import (
	"encoding/binary"
	"encoding/hex"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

// statusSample is the W32TIME_STATUS_INFO that W32TimeQueryStatus returned
// on a Windows Server 2025 Hyper-V guest, up to the source string.
const statusSample = "" +
	"78000000 00000000 01000000 06000000" +
	"564d5450 00000000 f0b7a4eb a867dd01" +
	"e5060000 00000000 a2860100 00000000" +
	"e9ffffff 00000000 e8e5a530 70010000" +
	"c0fdffff ffffffff 02000000 03000000" +
	"59620200 40000000 00000000 00000000" +
	"f6ccb901 00000000 00000000 00000000" +
	"00000000 00000000"

func TestParseStatus(t *testing.T) {
	t.Parallel()

	raw, err := hex.DecodeString(strings.ReplaceAll(statusSample, " ", ""))
	require.NoError(t, err)
	require.Len(t, raw, int(unsafe.Sizeof(statusInfo{})))

	source, err := windows.UTF16PtrFromString("VM IC Time Synchronization Provider")
	require.NoError(t, err)

	// The sample points into the native buffer; point it at a Go string instead.
	binary.LittleEndian.PutUint64(raw[unsafe.Offsetof(statusInfo{}.wszSource):], uint64(uintptr(unsafe.Pointer(source))))

	info := new(statusInfo)
	copy(unsafe.Slice((*byte)(unsafe.Pointer(info)), len(raw)), raw)

	status, err := parseStatus(info)
	require.NoError(t, err)
	require.True(t, status.LastSuccessfulSync.Equal(time.Date(2026, 10, 29, 13, 25, 9, 659032000, time.UTC)), status.LastSuccessfulSync)

	status.LastSuccessfulSync = time.Time{}
	require.Equal(t, Status{
		LeapIndicator:  0,
		Stratum:        1,
		PollInterval:   64 * time.Second,
		RootDelay:      176500 * time.Nanosecond,
		RootDispersion: 10000200 * time.Nanosecond,
		Precision:      119 * time.Nanosecond, // 2^-23 s.
		Source:         "VM IC Time Synchronization Provider",
		PhaseOffset:    -57600 * time.Nanosecond,
		State:          StateSync,
		LastSyncResult: SyncResultSuccess,
	}, status)
}

func TestParseStatusRejectsShortStructure(t *testing.T) {
	t.Parallel()

	_, err := parseStatus(&statusInfo{ulSize: 64})
	require.ErrorContains(t, err, "unexpected W32TIME_STATUS_INFO size 64")
}

func TestLog2Duration(t *testing.T) {
	t.Parallel()

	require.Equal(t, 1024*time.Second, log2Duration(10))
	require.Equal(t, time.Second, log2Duration(0))
	require.Equal(t, 500*time.Millisecond, log2Duration(-1))
	require.Zero(t, log2Duration(64))
}

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

package fveapi

import (
	"os"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
)

func TestStatusV1Layout(t *testing.T) {
	t.Parallel()

	// FVE_STATUS version 1 is 32 bytes with Flags at offset 8 and ConvertedPercent at offset 16.
	require.Equal(t, uintptr(32), unsafe.Sizeof(statusV1{}))
	require.Equal(t, uintptr(8), unsafe.Offsetof(statusV1{}.Flags))
	require.Equal(t, uintptr(16), unsafe.Offsetof(statusV1{}.ConvertedPercent))
	require.Equal(t, uintptr(24), unsafe.Offsetof(statusV1{}.LastConvertStatus))
}

func TestStatusState(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		flags uint32
		want  State
	}{
		// Flags observed on Windows 11 and compared with Win32_EncryptableVolume.
		{name: "on, observed", flags: 0x00045309, want: StateOn},
		{name: "off, observed", flags: 0x00000004, want: StateOff},

		// Flags derived from the flag names, not observed yet.
		{name: "suspended", flags: 0x00044709, want: StateSuspended},
		{name: "waiting for activation", flags: 0x00000409, want: StateWaitingForActivation},
		{name: "encrypting", flags: 0x00040321, want: StateEncrypting},
		{name: "encryption paused", flags: 0x00040361, want: StateEncrypting},
		{name: "decrypting", flags: 0x00040311, want: StateDecrypting},
		{name: "locked", flags: 0x00040909, want: StateLocked},

		// Unexpected combinations must not be reported as off.
		{name: "no flags", flags: 0, want: StateUnknown},
		{name: "initialized only", flags: StatusFlagInitialized, want: StateUnknown},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, Status{Flags: tc.flags}.State())
		})
	}
}

func TestGetStatus(t *testing.T) {
	t.Parallel()

	systemDrive := os.Getenv("SystemDrive")
	require.NotEmpty(t, systemDrive)

	status, err := GetStatus(`\\.\` + systemDrive)
	require.NoError(t, err)
	require.NotEqual(t, StateUnknown, status.State(), "unexpected flags 0x%08X", status.Flags)

	_, err = GetStatus(`\\.\Volume{00000000-0000-0000-0000-000000000000}`)
	require.Error(t, err)
	require.NotErrorIs(t, err, ErrUnavailable)
}

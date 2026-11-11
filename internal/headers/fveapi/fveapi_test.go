// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

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
		// Flags observed on Windows 11 and compared with Win32_EncryptableVolume and manage-bde.
		{name: "fully decrypted", flags: 0x00000004, want: StateOff},
		{name: "encryption paused", flags: 0x000401A1, want: StateEncrypting},
		{name: "encryption running", flags: 0x00042121, want: StateEncrypting},
		{name: "protection on", flags: 0x00041109, want: StateOn},
		{name: "protection on, OS volume with TPM", flags: 0x00045309, want: StateOn},
		{name: "protection suspended", flags: 0x00040509, want: StateSuspended},
		{name: "locked", flags: 0x00041909, want: StateLocked},
		{name: "decryption paused", flags: 0x00040191, want: StateDecrypting},
		{name: "decryption running", flags: 0x00042111, want: StateDecrypting},

		// Derived from the flag names, not observed yet.
		{name: "waiting for activation", flags: 0x00000409, want: StateWaitingForActivation},

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

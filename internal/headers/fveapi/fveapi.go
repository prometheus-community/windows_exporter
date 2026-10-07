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

// Package fveapi reads the BitLocker status of a volume through fveapi.dll.
//
// fveapi.dll is the private user-mode BitLocker (Full Volume Encryption) API.
// It is undocumented and Microsoft gives no ABI guarantee, so this package
// only uses the oldest and smallest surface: FveGetStatusW with the
// version 1 FVE_STATUS structure. The DLL is loaded lazily, and a missing
// DLL or export results in an error instead of a crash.
//
// Why not the Shell property System.Volume.BitLockerProtection: the Shell
// returns VT_EMPTY for built-in service accounts (LocalSystem, virtual
// service accounts), so it does not work for windows_exporter running as a
// service (https://github.com/prometheus-community/windows_exporter/issues/2290).
// FveGetStatusW returns the same result for LocalSystem and for a
// non-elevated user.
//
// Volume names must use the Win32 device namespace, for example
// `\\.\C:`, `\\.\Volume{GUID}` or `\\.\GLOBALROOT\Device\HarddiskVolume1`.
// A bare `C:` fails with E_ACCESSDENIED.
//
// The flag values come from the reverse-engineered header of KNSoft.NDK (MIT):
// https://github.com/KNSoft/KNSoft.NDK/blob/main/Source/Include/KNSoft/NDK/Win32/FVE/FveApi.h
//
// Verified on Windows 11 (build 26300) against Win32_EncryptableVolume and
// Get-BitLockerVolume, as LocalSystem and as a non-elevated user:
//
//	Protection On, FullyEncrypted  -> 0x00045309 (INITIALIZED|FULLY_ENCRYPTED|PROTECTION_ACTIVE|...)
//	Protection Off, FullyDecrypted -> 0x00000004 (FULLY_DECRYPTED)
//	EFI and recovery partitions    -> 0x00000004 (FULLY_DECRYPTED)
//	Non-existent volume            -> HRESULT 0x80070002
//
// The remaining states (suspended, encryption/decryption in progress or paused,
// locked, waiting for activation) are mapped from the flag names, but not
// verified yet.
package fveapi

import (
	"errors"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var (
	modFveapi         = windows.NewLazySystemDLL("fveapi.dll")
	procFveGetStatusW = modFveapi.NewProc("FveGetStatusW")
)

// FVE_STATUS_FLAG_* values of the Flags field.
const (
	StatusFlagInitialized          uint32 = 0x00000001
	StatusFlagFullyDecrypted       uint32 = 0x00000004
	StatusFlagFullyEncrypted       uint32 = 0x00000008
	StatusFlagDecryptionInProgress uint32 = 0x00000010
	StatusFlagEncryptionInProgress uint32 = 0x00000020
	StatusFlagConversionPausedMask uint32 = 0x000000C0
	StatusFlagNonTPMProtector      uint32 = 0x00000100
	StatusFlagTPMProtector         uint32 = 0x00000200
	StatusFlagClearKey             uint32 = 0x00000400
	StatusFlagLocked               uint32 = 0x00000800
	StatusFlagProtectionActive     uint32 = 0x00001000
	StatusFlagOSVolume             uint32 = 0x00004000
)

// statusV1 is the version 1 FVE_STATUS structure (32 bytes).
// The padding fields keep the layout independent of the platform alignment of float64.
type statusV1 struct {
	StructureSize     uint32
	StructureVersion  uint32
	Flags             uint32
	_                 uint32
	ConvertedPercent  float64
	LastConvertStatus int32
	_                 uint32
}

// Status is the BitLocker status of a volume as reported by FveGetStatusW.
type Status struct {
	Flags            uint32
	ConvertedPercent float64
}

// State is the BitLocker state derived from Status.Flags.
type State int

const (
	StateUnknown State = iota
	StateOff
	StateOn
	StateEncrypting
	StateDecrypting
	StateSuspended
	StateLocked
	StateWaitingForActivation
)

var ErrUnavailable = errors.New("fveapi.dll or FveGetStatusW is not available")

// GetStatus returns the BitLocker status of volumeName, e.g. `\\.\C:`.
func GetStatus(volumeName string) (Status, error) {
	if err := procFveGetStatusW.Find(); err != nil {
		return Status{}, fmt.Errorf("%w: %w", ErrUnavailable, err)
	}

	volumeNamePtr, err := windows.UTF16PtrFromString(volumeName)
	if err != nil {
		return Status{}, fmt.Errorf("invalid volume name %q: %w", volumeName, err)
	}

	status := statusV1{
		StructureSize:    uint32(unsafe.Sizeof(statusV1{})),
		StructureVersion: 1,
	}

	hr, _, _ := procFveGetStatusW.Call(
		uintptr(unsafe.Pointer(volumeNamePtr)),
		uintptr(unsafe.Pointer(&status)),
	)

	// FveGetStatusW returns an HRESULT. Only the sign bit indicates failure.
	if int32(hr) < 0 {
		// HRESULT_FROM_WIN32 (facility 7) wraps a Win32 error code, e.g. 0x80070002.
		if uint32(hr)&0xFFFF0000 == 0x80070000 {
			return Status{}, fmt.Errorf("FveGetStatusW(%s) failed: %w", volumeName, windows.Errno(uint32(hr)&0xFFFF))
		}

		return Status{}, fmt.Errorf("FveGetStatusW(%s) failed: HRESULT 0x%08X", volumeName, uint32(hr))
	}

	return Status{
		Flags:            status.Flags,
		ConvertedPercent: status.ConvertedPercent,
	}, nil
}

// State maps the status flags to a BitLocker state.
// A fully encrypted volume is only reported as on if protection is active,
// because a suspended volume stays fully encrypted with protection off.
// Flag combinations that do not match a known state return StateUnknown
// instead of being reported as off.
func (s Status) State() State {
	flags := s.Flags

	switch {
	case flags&StatusFlagLocked != 0:
		return StateLocked
	case flags&StatusFlagFullyDecrypted != 0:
		return StateOff
	case flags&StatusFlagInitialized == 0:
		return StateUnknown
	case flags&StatusFlagEncryptionInProgress != 0:
		return StateEncrypting
	case flags&StatusFlagDecryptionInProgress != 0:
		return StateDecrypting
	case flags&StatusFlagFullyEncrypted == 0:
		return StateUnknown
	case flags&StatusFlagProtectionActive != 0:
		return StateOn
	case flags&StatusFlagClearKey != 0 && flags&(StatusFlagTPMProtector|StatusFlagNonTPMProtector) == 0:
		// Encrypted with a clear key, before any key protector was added.
		return StateWaitingForActivation
	default:
		return StateSuspended
	}
}

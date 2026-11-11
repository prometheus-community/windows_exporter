// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package sysinfoapi

import (
	"fmt"
	"unsafe"
)

//nolint:gochecknoglobals
var procGetSystemFirmwareTable = kernel32.NewProc("GetSystemFirmwareTable")

// SMBIOS table provider signature ('RSMB').
const firmwareTableProviderSigRSMB uint32 = 0x52534D42

// GetSMBIOSInfo retrieves the SMBIOS BIOS, System, Baseboard and Chassis
// information using the GetSystemFirmwareTable Win32 API.  No WMI dependency.
func GetSMBIOSInfo() (*SMBIOSInfo, error) {
	data, err := getSystemFirmwareTable(firmwareTableProviderSigRSMB, 0)
	if err != nil {
		return nil, fmt.Errorf("GetSystemFirmwareTable: %w", err)
	}

	return parseSMBIOSInfo(data)
}

// getSystemFirmwareTable wraps the GetSystemFirmwareTable Win32 API.
// https://learn.microsoft.com/en-us/windows/win32/api/sysinfoapi/nf-sysinfoapi-getsystemfirmwaretable
func getSystemFirmwareTable(providerSig, tableID uint32) ([]byte, error) {
	// First call: determine buffer size.
	r1, _, err := procGetSystemFirmwareTable.Call(
		uintptr(providerSig),
		uintptr(tableID),
		0,
		0,
	)

	size := uint32(r1)
	if size == 0 {
		return nil, fmt.Errorf("failed to get firmware table size: %w", err)
	}

	buf := make([]byte, size)

	r1, _, err = procGetSystemFirmwareTable.Call(
		uintptr(providerSig),
		uintptr(tableID),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(size),
	)

	written := uint32(r1)
	if written == 0 {
		return nil, fmt.Errorf("failed to get firmware table: %w", err)
	}

	if written > size {
		return nil, fmt.Errorf("firmware table size changed between calls (need %d, have %d)", written, size)
	}

	return buf[:written], nil
}

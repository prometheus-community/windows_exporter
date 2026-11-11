// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package kernel32

import (
	"errors"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var procGetDiskSpaceInformationW = modkernel32.NewProc("GetDiskSpaceInformationW")

// ErrDiskSpaceInformationUnavailable is returned by GetDiskSpaceInformation on Windows versions
// before Windows 10 1809 and Windows Server 2019, which lack GetDiskSpaceInformationW.
var ErrDiskSpaceInformationUnavailable = errors.New("GetDiskSpaceInformationW is not available")

// DiskSpaceInformation contains disk space information for a volume.
// 📑 https://learn.microsoft.com/en-us/windows/win32/api/fileapi/ns-fileapi-disk_space_information
type DiskSpaceInformation struct {
	ActualTotalAllocationUnits           uint64
	ActualAvailableAllocationUnits       uint64
	ActualPoolUnavailableAllocationUnits uint64
	CallerTotalAllocationUnits           uint64
	CallerAvailableAllocationUnits       uint64
	CallerPoolUnavailableAllocationUnits uint64
	UsedAllocationUnits                  uint64
	TotalReservedAllocationUnits         uint64
	VolumeStorageReserveAllocationUnits  uint64
	AvailableCommittedAllocationUnits    uint64
	PoolAvailableAllocationUnits         uint64
	SectorsPerAllocationUnit             uint32
	BytesPerSector                       uint32
}

// AllocationUnitSize returns the size of an allocation unit in bytes.
func (d DiskSpaceInformation) AllocationUnitSize() uint64 {
	return uint64(d.SectorsPerAllocationUnit) * uint64(d.BytesPerSector)
}

// GetDiskSpaceInformation returns the disk space information of the volume with the root path rootPath.
// Unlike GetDiskFreeSpaceEx, it reports the size of the volume regardless of disk quotas.
// 📑 https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-getdiskspaceinformationw
func GetDiskSpaceInformation(rootPath *uint16) (DiskSpaceInformation, error) {
	var info DiskSpaceInformation

	if err := procGetDiskSpaceInformationW.Find(); err != nil {
		return info, fmt.Errorf("%w: %w", ErrDiskSpaceInformationUnavailable, err)
	}

	hr, _, _ := procGetDiskSpaceInformationW.Call(
		uintptr(unsafe.Pointer(rootPath)),
		uintptr(unsafe.Pointer(&info)),
	)
	if int32(hr) < 0 {
		if uint32(hr)&0xFFFF0000 == 0x80070000 {
			return info, fmt.Errorf("GetDiskSpaceInformationW failed: %w", windows.Errno(uint32(hr)&0xFFFF))
		}

		return info, fmt.Errorf("GetDiskSpaceInformationW failed: HRESULT 0x%08X", uint32(hr))
	}

	return info, nil
}

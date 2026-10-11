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

// Package virtdisk reads the size and type of virtual hard disks (VHD, VHDX,
// VHD Set) through virtdisk.dll, the API behind Get-VHD.
// https://learn.microsoft.com/en-us/windows/win32/api/virtdisk/
package virtdisk

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var (
	modVirtdisk                   = windows.NewLazySystemDLL("virtdisk.dll")
	procOpenVirtualDisk           = modVirtdisk.NewProc("OpenVirtualDisk")
	procGetVirtualDiskInformation = modVirtdisk.NewProc("GetVirtualDiskInformation")
)

// Provider subtypes of GET_VIRTUAL_DISK_INFO_PROVIDER_SUBTYPE.
const (
	ProviderSubtypeFixed        uint32 = 2
	ProviderSubtypeDynamic      uint32 = 3
	ProviderSubtypeDifferencing uint32 = 4
)

const (
	virtualDiskAccessGetInfo     = 0x00080000 // VIRTUAL_DISK_ACCESS_GET_INFO
	openVirtualDiskFlagNoParents = 0x00000001 // OPEN_VIRTUAL_DISK_FLAG_NO_PARENTS

	getVirtualDiskInfoSize            = 1 // GET_VIRTUAL_DISK_INFO_SIZE
	getVirtualDiskInfoProviderSubtype = 7 // GET_VIRTUAL_DISK_INFO_PROVIDER_SUBTYPE
)

//nolint:gochecknoglobals
var virtualStorageTypeVendorMicrosoft = windows.GUID{
	Data1: 0xEC984AEC, Data2: 0xA0F9, Data3: 0x47E9,
	Data4: [8]byte{0x90, 0x1F, 0x71, 0x41, 0x5A, 0x66, 0x34, 0x5B},
}

// virtualStorageType is VIRTUAL_STORAGE_TYPE. The zero value is
// VIRTUAL_STORAGE_TYPE_DEVICE_UNKNOWN with VIRTUAL_STORAGE_TYPE_VENDOR_UNKNOWN,
// which lets OpenVirtualDisk detect VHD, VHDX and VHD Set files.
type virtualStorageType struct {
	DeviceID uint32
	VendorID windows.GUID
}

// getVirtualDiskInfo is GET_VIRTUAL_DISK_INFO. The union is 8-byte aligned
// and its largest member is the 24-byte Size struct.
type getVirtualDiskInfo struct {
	Version uint32
	_       uint32
	Union   [24]byte
}

// Info is the size and type of a virtual hard disk.
type Info struct {
	// VirtualSize is the maximum size the virtual machine sees, in bytes.
	VirtualSize uint64
	// PhysicalSize is the size of the file on storage, in bytes.
	PhysicalSize uint64
	// ProviderSubtype is ProviderSubtypeFixed, ProviderSubtypeDynamic or ProviderSubtypeDifferencing.
	ProviderSubtype uint32
}

// Available reports whether virtdisk.dll and the functions used by GetInfo
// can be loaded. LazyProc.Call panics if they can't, so check this first.
func Available() error {
	if err := procOpenVirtualDisk.Find(); err != nil {
		return err
	}

	return procGetVirtualDiskInformation.Find()
}

// GetInfo opens the virtual hard disk at path read-only, without its parents,
// and returns its size and provider subtype. It works on disks attached to
// running virtual machines.
func GetInfo(path string) (Info, error) {
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return Info{}, fmt.Errorf("virtual disk %s: %w", path, err)
	}

	var (
		storageType virtualStorageType
		handle      windows.Handle
	)

	ret, _, _ := procOpenVirtualDisk.Call(
		uintptr(unsafe.Pointer(&storageType)),
		uintptr(unsafe.Pointer(pathPtr)),
		virtualDiskAccessGetInfo,
		openVirtualDiskFlagNoParents,
		0, // OPEN_VIRTUAL_DISK_PARAMETERS: NULL means version 1 defaults
		uintptr(unsafe.Pointer(&handle)),
	)
	if ret != 0 {
		return Info{}, fmt.Errorf("OpenVirtualDisk %s: %w", path, windows.Errno(ret))
	}

	defer func() { _ = windows.CloseHandle(handle) }()

	size, err := getInformation(handle, getVirtualDiskInfoSize)
	if err != nil {
		return Info{}, fmt.Errorf("GetVirtualDiskInformation size %s: %w", path, err)
	}

	subtype, err := getInformation(handle, getVirtualDiskInfoProviderSubtype)
	if err != nil {
		return Info{}, fmt.Errorf("GetVirtualDiskInformation provider subtype %s: %w", path, err)
	}

	return Info{
		VirtualSize:     binary.LittleEndian.Uint64(size.Union[0:8]),
		PhysicalSize:    binary.LittleEndian.Uint64(size.Union[8:16]),
		ProviderSubtype: binary.LittleEndian.Uint32(subtype.Union[0:4]),
	}, nil
}

func getInformation(handle windows.Handle, version uint32) (getVirtualDiskInfo, error) {
	info := getVirtualDiskInfo{Version: version}
	infoSize := uint32(unsafe.Sizeof(info))

	ret, _, _ := procGetVirtualDiskInformation.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(&infoSize)),
		uintptr(unsafe.Pointer(&info)),
		0, // SizeUsed is optional
	)
	if ret != 0 {
		return getVirtualDiskInfo{}, windows.Errno(ret)
	}

	return info, nil
}

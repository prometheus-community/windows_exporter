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
	virtualDiskAccessNone        = 0x00000000 // VIRTUAL_DISK_ACCESS_NONE, required with version 2 parameters
	openVirtualDiskFlagNoParents = 0x00000001 // OPEN_VIRTUAL_DISK_FLAG_NO_PARENTS
	openVirtualDiskVersion2      = 2          // OPEN_VIRTUAL_DISK_VERSION_2

	getVirtualDiskInfoSize            = 1 // GET_VIRTUAL_DISK_INFO_SIZE
	getVirtualDiskInfoParentLocation  = 3 // GET_VIRTUAL_DISK_INFO_PARENT_LOCATION
	getVirtualDiskInfoProviderSubtype = 7 // GET_VIRTUAL_DISK_INFO_PROVIDER_SUBTYPE
)

// virtualStorageType is VIRTUAL_STORAGE_TYPE. The zero value is
// VIRTUAL_STORAGE_TYPE_DEVICE_UNKNOWN with VIRTUAL_STORAGE_TYPE_VENDOR_UNKNOWN,
// which lets OpenVirtualDisk detect VHD, VHDX and VHD Set files.
type virtualStorageType struct {
	DeviceID uint32
	VendorID windows.GUID
}

// openVirtualDiskParametersV2 is OPEN_VIRTUAL_DISK_PARAMETERS with
// OPEN_VIRTUAL_DISK_VERSION_2. The union is 4-byte aligned.
type openVirtualDiskParametersV2 struct {
	Version        uint32
	GetInfoOnly    int32 // BOOL
	ReadOnly       int32 // BOOL
	ResiliencyGUID windows.GUID
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
	handle, err := open(path)
	if err != nil {
		return Info{}, err
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

// ParentLocations returns the paths at which the parent of the differencing
// disk at path is stored, as recorded in the disk. Disks without a parent
// return none.
func ParentLocations(path string) ([]string, error) {
	handle, err := open(path)
	if err != nil {
		return nil, err
	}

	defer func() { _ = windows.CloseHandle(handle) }()

	subtype, err := getInformation(handle, getVirtualDiskInfoProviderSubtype)
	if err != nil {
		return nil, fmt.Errorf("GetVirtualDiskInformation provider subtype %s: %w", path, err)
	}

	if binary.LittleEndian.Uint32(subtype.Union[0:4]) != ProviderSubtypeDifferencing {
		return nil, nil
	}

	// GET_VIRTUAL_DISK_INFO with ParentLocation: the 8-byte aligned union starts
	// at offset 8 with the BOOL ParentResolved, followed by the MULTI_SZ
	// ParentLocationBuffer. A uint64 buffer keeps the alignment.
	buffer := make([]uint64, 512)

	for range 4 {
		bytes := unsafe.Slice((*byte)(unsafe.Pointer(&buffer[0])), len(buffer)*8)
		binary.LittleEndian.PutUint32(bytes, getVirtualDiskInfoParentLocation)

		infoSize := uint32(len(bytes))

		ret, _, _ := procGetVirtualDiskInformation.Call(
			uintptr(handle),
			uintptr(unsafe.Pointer(&infoSize)),
			uintptr(unsafe.Pointer(&buffer[0])),
			0, // SizeUsed is optional
		)

		switch windows.Errno(ret) {
		case 0:
			if infoSize < 12 || int(infoSize) > len(bytes) {
				return nil, fmt.Errorf("GetVirtualDiskInformation parent location %s: invalid size %d", path, infoSize)
			}

			chars := unsafe.Slice((*uint16)(unsafe.Pointer(&bytes[12])), (int(infoSize)-12)/2)

			return decodeMultiString(chars), nil
		case windows.ERROR_INSUFFICIENT_BUFFER, windows.ERROR_MORE_DATA:
			if int(infoSize) <= len(bytes) {
				return nil, fmt.Errorf("GetVirtualDiskInformation parent location %s: %w", path, windows.Errno(ret))
			}

			buffer = make([]uint64, (int(infoSize)+7)/8)
		default:
			return nil, fmt.Errorf("GetVirtualDiskInformation parent location %s: %w", path, windows.Errno(ret))
		}
	}

	return nil, fmt.Errorf("GetVirtualDiskInformation parent location %s: buffer did not stabilize", path)
}

// decodeMultiString splits a MULTI_SZ into its strings.
func decodeMultiString(chars []uint16) []string {
	var out []string

	for len(chars) > 0 {
		end := 0
		for end < len(chars) && chars[end] != 0 {
			end++
		}

		if end == 0 {
			break
		}

		out = append(out, windows.UTF16ToString(chars[:end]))

		if end == len(chars) {
			break
		}

		chars = chars[end+1:]
	}

	return out
}

// open opens the virtual hard disk at path for reading its information only.
// GetInfoOnly also opens a disk that is in use by a running virtual machine.
func open(path string) (windows.Handle, error) {
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, fmt.Errorf("virtual disk %s: %w", path, err)
	}

	var (
		storageType virtualStorageType
		handle      windows.Handle
	)

	params := openVirtualDiskParametersV2{Version: openVirtualDiskVersion2, GetInfoOnly: 1}

	ret, _, _ := procOpenVirtualDisk.Call(
		uintptr(unsafe.Pointer(&storageType)),
		uintptr(unsafe.Pointer(pathPtr)),
		virtualDiskAccessNone,
		openVirtualDiskFlagNoParents,
		uintptr(unsafe.Pointer(&params)),
		uintptr(unsafe.Pointer(&handle)),
	)
	if ret != 0 {
		return 0, fmt.Errorf("OpenVirtualDisk %s: %w", path, windows.Errno(ret))
	}

	return handle, nil
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

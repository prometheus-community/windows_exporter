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

package virtdisk

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var (
	procCreateVirtualDisk = modVirtdisk.NewProc("CreateVirtualDisk")

	virtualStorageTypeVendorMicrosoft = windows.GUID{
		Data1: 0xEC984AEC, Data2: 0xA0F9, Data3: 0x47E9,
		Data4: [8]byte{0x90, 0x1F, 0x71, 0x41, 0x5A, 0x66, 0x34, 0x5B},
	}
)

// createVirtualDiskParametersV2 is CREATE_VIRTUAL_DISK_PARAMETERS with
// CREATE_VIRTUAL_DISK_VERSION_2, which VHDX requires. The union starts at offset 8.
type createVirtualDiskParametersV2 struct {
	Version                   uint32
	_                         uint32
	UniqueID                  windows.GUID
	MaximumSize               uint64
	BlockSizeInBytes          uint32
	SectorSizeInBytes         uint32
	PhysicalSectorSizeInBytes uint32
	_                         uint32
	ParentPath                *uint16
	SourcePath                *uint16
	OpenFlags                 uint32
	ParentVirtualStorageType  virtualStorageType
	SourceVirtualStorageType  virtualStorageType
	ResiliencyGUID            windows.GUID
}

const (
	createVirtualDiskVersion2              = 2
	createVirtualDiskFlagFullPhysicalAlloc = 0x1
	virtualStorageTypeDeviceVHDX           = 3
)

// createVHDX creates a VHDX for the test and closes it. Creating a virtual
// disk needs administrator rights; the test is skipped without them.
func createVHDX(t *testing.T, path string, size uint64, fixed bool) {
	t.Helper()

	createVirtualDisk(t, path, createVirtualDiskParametersV2{Version: createVirtualDiskVersion2, MaximumSize: size}, fixed)
}

// createDifferencingVHDX creates a differencing VHDX with parent as its parent.
func createDifferencingVHDX(t *testing.T, path, parent string) {
	t.Helper()

	parentPtr, err := windows.UTF16PtrFromString(parent)
	require.NoError(t, err)

	createVirtualDisk(t, path, createVirtualDiskParametersV2{Version: createVirtualDiskVersion2, ParentPath: parentPtr}, false)
}

func createVirtualDisk(t *testing.T, path string, params createVirtualDiskParametersV2, fixed bool) {
	t.Helper()

	pathPtr, err := windows.UTF16PtrFromString(path)
	require.NoError(t, err)

	storageType := virtualStorageType{DeviceID: virtualStorageTypeDeviceVHDX, VendorID: virtualStorageTypeVendorMicrosoft}

	var flags uintptr
	if fixed {
		flags = createVirtualDiskFlagFullPhysicalAlloc
	}

	var handle windows.Handle

	ret, _, _ := procCreateVirtualDisk.Call(
		uintptr(unsafe.Pointer(&storageType)),
		uintptr(unsafe.Pointer(pathPtr)),
		virtualDiskAccessNone,
		0, // security descriptor
		flags,
		0, // provider specific flags
		uintptr(unsafe.Pointer(&params)),
		0, // overlapped
		uintptr(unsafe.Pointer(&handle)),
	)
	if errors.Is(windows.Errno(ret), windows.ERROR_ACCESS_DENIED) || errors.Is(windows.Errno(ret), windows.ERROR_PRIVILEGE_NOT_HELD) {
		t.Skipf("creating a virtual disk needs administrator rights: %v", windows.Errno(ret))
	}

	require.Zero(t, ret, "CreateVirtualDisk: %v", windows.Errno(ret))
	require.NoError(t, windows.CloseHandle(handle))
}

func TestAvailable(t *testing.T) {
	t.Parallel()

	require.NoError(t, Available())
}

func TestGetInfo(t *testing.T) {
	t.Parallel()

	const size = 64 << 20 // 64 MiB

	for name, tc := range map[string]struct {
		fixed   bool
		subtype uint32
	}{
		"fixed":   {true, ProviderSubtypeFixed},
		"dynamic": {false, ProviderSubtypeDynamic},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), name+".vhdx")
			createVHDX(t, path, size, tc.fixed)

			info, err := GetInfo(path)
			require.NoError(t, err)
			require.Equal(t, uint64(size), info.VirtualSize)
			require.Equal(t, tc.subtype, info.ProviderSubtype)

			if tc.fixed {
				require.GreaterOrEqual(t, info.PhysicalSize, uint64(size))
			} else {
				require.Less(t, info.PhysicalSize, uint64(size))
			}
		})
	}
}

func TestGetInfoMissingFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "missing.vhdx")

	_, err := GetInfo(path)
	require.ErrorContains(t, err, path)
}

func TestParentLocations(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	parent := filepath.Join(dir, "parent.vhdx")
	child := filepath.Join(dir, "child.avhdx")

	createVHDX(t, parent, 64<<20, false)
	createDifferencingVHDX(t, child, parent)

	info, err := GetInfo(child)
	require.NoError(t, err)
	require.Equal(t, ProviderSubtypeDifferencing, info.ProviderSubtype)

	locations, err := ParentLocations(child)
	require.NoError(t, err)

	// Compare files, not strings: the temp dir can be a short 8.3 path
	// (RUNNER~1) while the disk records the long one.
	parentInfo, err := os.Stat(parent)
	require.NoError(t, err)
	require.Condition(t, func() bool {
		for _, location := range locations {
			if info, err := os.Stat(location); err == nil && os.SameFile(info, parentInfo) {
				return true
			}
		}

		return false
	}, "parent %s not in %v", parent, locations)

	// A disk without a parent has no parent locations.
	locations, err = ParentLocations(parent)
	require.NoError(t, err)
	require.Empty(t, locations)
}

func TestDecodeMultiString(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		in   []uint16
		want []string
	}{
		"two":          {append(append(windows.StringToUTF16(`..\parent.vhdx`), windows.StringToUTF16(`C:\vm\parent.vhdx`)...), 0), []string{`..\parent.vhdx`, `C:\vm\parent.vhdx`}},
		"one":          {windows.StringToUTF16(`C:\vm\parent.vhd`), []string{`C:\vm\parent.vhd`}},
		"empty":        {[]uint16{0, 0}, nil},
		"unterminated": {[]uint16{'a', 'b'}, []string{"ab"}},
		"nil":          {nil, nil},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, decodeMultiString(tc.in))
		})
	}
}

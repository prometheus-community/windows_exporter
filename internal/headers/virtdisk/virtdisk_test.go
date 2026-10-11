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
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var procCreateVirtualDisk = modVirtdisk.NewProc("CreateVirtualDisk")

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

	pathPtr, err := windows.UTF16PtrFromString(path)
	require.NoError(t, err)

	storageType := virtualStorageType{DeviceID: virtualStorageTypeDeviceVHDX, VendorID: virtualStorageTypeVendorMicrosoft}
	params := createVirtualDiskParametersV2{Version: createVirtualDiskVersion2, MaximumSize: size}

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

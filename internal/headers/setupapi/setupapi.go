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

// Package setupapi contains the SetupAPI device interface functions that
// golang.org/x/sys/windows does not provide.
package setupapi

import (
	"errors"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var (
	modsetupapi = windows.NewLazySystemDLL("setupapi.dll")

	procSetupDiEnumDeviceInterfaces      = modsetupapi.NewProc("SetupDiEnumDeviceInterfaces")
	procSetupDiGetDeviceInterfaceDetailW = modsetupapi.NewProc("SetupDiGetDeviceInterfaceDetailW")
)

// GUID_DEVINTERFACE_DISK is the device interface class of disk devices.
// https://learn.microsoft.com/en-us/windows-hardware/drivers/install/guid-devinterface-disk
//
//nolint:gochecknoglobals
var GUID_DEVINTERFACE_DISK = windows.GUID{
	Data1: 0x53f56307,
	Data2: 0xb6bf,
	Data3: 0x11d0,
	Data4: [8]byte{0x94, 0xf2, 0x00, 0xa0, 0xc9, 0x1e, 0xfb, 0x8b},
}

// maxDetailSize bounds the interface detail buffer. Interface paths are far
// shorter; the bound only protects against a corrupt required-size answer.
const maxDetailSize = 64 * 1024

// SP_DEVICE_INTERFACE_DATA
// https://learn.microsoft.com/en-us/windows/win32/api/setupapi/ns-setupapi-sp_device_interface_data
type SP_DEVICE_INTERFACE_DATA struct {
	CbSize             uint32
	InterfaceClassGuid windows.GUID
	Flags              uint32
	Reserved           uintptr
}

// SP_DEVINFO_DATA
// https://learn.microsoft.com/en-us/windows/win32/api/setupapi/ns-setupapi-sp_devinfo_data
//
// windows.DevInfoData can't be used here because its size field is unexported.
type SP_DEVINFO_DATA struct {
	CbSize    uint32
	ClassGUID windows.GUID
	DevInst   uint32
	Reserved  uintptr
}

// DeviceInterface is a present device interface and the device node that
// exposes it.
type DeviceInterface struct {
	// Path is the device interface path that can be passed to CreateFile.
	Path string
	// DevInst is the configuration manager device instance handle.
	DevInst uint32
}

// GetDeviceInterfaces returns the present device interfaces of the given
// interface class. It uses SetupDiGetClassDevs(DIGCF_PRESENT|DIGCF_DEVICEINTERFACE),
// SetupDiEnumDeviceInterfaces and SetupDiGetDeviceInterfaceDetailW.
func GetDeviceInterfaces(interfaceClass *windows.GUID) ([]DeviceInterface, error) {
	devInfo, err := windows.SetupDiGetClassDevsEx(interfaceClass, "", 0, windows.DIGCF_PRESENT|windows.DIGCF_DEVICEINTERFACE, 0, "")
	if err != nil {
		return nil, fmt.Errorf("SetupDiGetClassDevs: %w", err)
	}

	defer func() {
		_ = devInfo.Close()
	}()

	var interfaces []DeviceInterface

	for index := uint32(0); ; index++ {
		interfaceData := SP_DEVICE_INTERFACE_DATA{
			CbSize: uint32(unsafe.Sizeof(SP_DEVICE_INTERFACE_DATA{})),
		}

		if err := SetupDiEnumDeviceInterfaces(devInfo, interfaceClass, index, &interfaceData); err != nil {
			if errors.Is(err, windows.ERROR_NO_MORE_ITEMS) {
				return interfaces, nil
			}

			return nil, fmt.Errorf("SetupDiEnumDeviceInterfaces: %w", err)
		}

		path, devInst, err := getDeviceInterfaceDetail(devInfo, &interfaceData)
		if err != nil {
			return nil, err
		}

		interfaces = append(interfaces, DeviceInterface{Path: path, DevInst: devInst})
	}
}

func getDeviceInterfaceDetail(devInfo windows.DevInfo, interfaceData *SP_DEVICE_INTERFACE_DATA) (string, uint32, error) {
	var requiredSize uint32

	// The size query is expected to fail with ERROR_INSUFFICIENT_BUFFER.
	err := SetupDiGetDeviceInterfaceDetail(devInfo, interfaceData, nil, 0, &requiredSize, nil)
	if err != nil && !errors.Is(err, windows.ERROR_INSUFFICIENT_BUFFER) {
		return "", 0, fmt.Errorf("SetupDiGetDeviceInterfaceDetailW: %w", err)
	}

	if requiredSize < detailHeaderSize() || requiredSize > maxDetailSize {
		return "", 0, fmt.Errorf("SetupDiGetDeviceInterfaceDetailW: unexpected detail size %d", requiredSize)
	}

	// A []uint32 backing array keeps the DWORD cbSize field aligned.
	buf := make([]uint32, (requiredSize+3)/4)
	buf[0] = detailHeaderSize()

	devInfoData := SP_DEVINFO_DATA{
		CbSize: uint32(unsafe.Sizeof(SP_DEVINFO_DATA{})),
	}

	if err := SetupDiGetDeviceInterfaceDetail(devInfo, interfaceData, unsafe.Pointer(&buf[0]), uint32(len(buf)*4), nil, &devInfoData); err != nil {
		return "", 0, fmt.Errorf("SetupDiGetDeviceInterfaceDetailW: %w", err)
	}

	// DevicePath follows the DWORD cbSize field.
	path := unsafe.Slice((*uint16)(unsafe.Pointer(&buf[1])), (len(buf)-1)*2)

	return windows.UTF16ToString(path), devInfoData.DevInst, nil
}

// detailHeaderSize returns sizeof(SP_DEVICE_INTERFACE_DETAIL_DATA_W): 8 on
// 64-bit Windows (8-byte packing) and 6 on 32-bit Windows (1-byte packing).
func detailHeaderSize() uint32 {
	if unsafe.Sizeof(uintptr(0)) == 8 {
		return 8
	}

	return 6
}

// SetupDiEnumDeviceInterfaces
// https://learn.microsoft.com/en-us/windows/win32/api/setupapi/nf-setupapi-setupdienumdeviceinterfaces
func SetupDiEnumDeviceInterfaces(devInfo windows.DevInfo, interfaceClass *windows.GUID, memberIndex uint32, interfaceData *SP_DEVICE_INTERFACE_DATA) error {
	r1, _, err := procSetupDiEnumDeviceInterfaces.Call(
		uintptr(devInfo),
		0,
		uintptr(unsafe.Pointer(interfaceClass)),
		uintptr(memberIndex),
		uintptr(unsafe.Pointer(interfaceData)),
	)
	if r1 == 0 {
		return err
	}

	return nil
}

// SetupDiGetDeviceInterfaceDetail
// https://learn.microsoft.com/en-us/windows/win32/api/setupapi/nf-setupapi-setupdigetdeviceinterfacedetailw
func SetupDiGetDeviceInterfaceDetail(
	devInfo windows.DevInfo,
	interfaceData *SP_DEVICE_INTERFACE_DATA,
	detail unsafe.Pointer,
	detailSize uint32,
	requiredSize *uint32,
	devInfoData *SP_DEVINFO_DATA,
) error {
	r1, _, err := procSetupDiGetDeviceInterfaceDetailW.Call(
		uintptr(devInfo),
		uintptr(unsafe.Pointer(interfaceData)),
		uintptr(detail),
		uintptr(detailSize),
		uintptr(unsafe.Pointer(requiredSize)),
		uintptr(unsafe.Pointer(devInfoData)),
	)
	if r1 == 0 {
		return err
	}

	return nil
}

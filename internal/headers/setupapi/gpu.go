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

package setupapi

import (
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var GUID_DISPLAY_ADAPTER = sync.OnceValue(func() *windows.GUID {
	return &windows.GUID{
		Data1: 0x4d36e968,
		Data2: 0xe325,
		Data3: 0x11ce,
		Data4: [8]byte{0xbf, 0xc1, 0x08, 0x00, 0x2b, 0xe1, 0x03, 0x18},
	}
})

func GetGPUDevices() ([]GPUDevice, error) {
	hDevInfo, _, err := procSetupDiGetClassDevsW.Call(
		uintptr(unsafe.Pointer(GUID_DISPLAY_ADAPTER())),
		0,
		0,
		DIGCF_PRESENT,
	)

	if windows.Handle(hDevInfo) == windows.InvalidHandle {
		return nil, err
	}

	var (
		devices        []GPUDevice
		deviceData     SP_DEVINFO_DATA
		propertyBuffer [256]uint16
	)

	deviceData.CbSize = uint32(unsafe.Sizeof(deviceData))

	for i := 0; ; i++ {
		ret, _, _ := procSetupDiEnumDeviceInfo.Call(hDevInfo, uintptr(i), uintptr(unsafe.Pointer(&deviceData)))
		if ret == 0 {
			break // No more devices
		}

		ret, _, _ = procSetupDiGetDeviceRegistryPropertyW.Call(
			hDevInfo,
			uintptr(unsafe.Pointer(&deviceData)),
			uintptr(SPDRP_DEVICEDESC),
			0,
			uintptr(unsafe.Pointer(&propertyBuffer[0])),
			uintptr(len(propertyBuffer)*2),
			0,
		)

		gpuDevice := GPUDevice{}

		if ret == 0 {
			gpuDevice.DeviceDesc = ""
		} else {
			gpuDevice.DeviceDesc = windows.UTF16ToString(propertyBuffer[:])
		}

		ret, _, _ = procSetupDiGetDeviceRegistryPropertyW.Call(
			hDevInfo,
			uintptr(unsafe.Pointer(&deviceData)),
			uintptr(SPDRP_FRIENDLYNAME),
			0,
			uintptr(unsafe.Pointer(&propertyBuffer[0])),
			uintptr(len(propertyBuffer)*2),
			0,
		)

		if ret == 0 {
			gpuDevice.FriendlyName = ""
		} else {
			gpuDevice.FriendlyName = windows.UTF16ToString(propertyBuffer[:])
		}

		ret, _, _ = procSetupDiGetDeviceRegistryPropertyW.Call(
			hDevInfo,
			uintptr(unsafe.Pointer(&deviceData)),
			uintptr(SPDRP_HARDWAREID),
			0,
			uintptr(unsafe.Pointer(&propertyBuffer[0])),
			uintptr(len(propertyBuffer)*2),
			0,
		)

		if ret == 0 {
			gpuDevice.HardwareID = "unknown"
		} else {
			gpuDevice.HardwareID = windows.UTF16ToString(propertyBuffer[:])
		}

		ret, _, _ = procSetupDiGetDeviceRegistryPropertyW.Call(
			hDevInfo,
			uintptr(unsafe.Pointer(&deviceData)),
			uintptr(SPDRP_PHYSICAL_DEVICE_OBJECT_NAME),
			0,
			uintptr(unsafe.Pointer(&propertyBuffer[0])),
			uintptr(len(propertyBuffer)*2),
			0,
		)

		if ret == 0 {
			gpuDevice.PhysicalDeviceObjectName = "unknown"
		} else {
			gpuDevice.PhysicalDeviceObjectName = windows.UTF16ToString(propertyBuffer[:])
		}

		devices = append(devices, gpuDevice)
	}

	_, _, _ = procSetupDiDestroyDeviceInfoList.Call(hDevInfo)

	return devices, nil
}

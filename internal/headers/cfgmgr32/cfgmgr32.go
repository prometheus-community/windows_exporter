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

package cfgmgr32

import (
	"fmt"
	"unsafe"

	"github.com/prometheus-community/windows_exporter/internal/headers/win32"
	"golang.org/x/sys/windows"
)

func GetDevicesInstanceIDs(deviceID string) ([]Device, error) {
	var (
		err      error
		listSize uint32
	)

	deviceIDLWStr := win32.NewLPWSTR(deviceID)

	err = CMGetDeviceIDListSize(deviceIDLWStr, &listSize)
	if err != nil {
		return nil, err
	}

	listBuffer := make([]uint16, listSize)

	err = CMGetDeviceIDList(deviceIDLWStr, listBuffer)
	if err != nil {
		return nil, err
	}

	deviceInstanceIDs := win32.ParseMultiSz(listBuffer)
	devices := make([]Device, 0, len(deviceInstanceIDs))

	for _, deviceInstanceID := range deviceInstanceIDs {
		var devNode *windows.Handle

		err = CMLocateDevNode(devNode, deviceInstanceID)
		if err != nil {
			return nil, err
		}

		var (
			busNumber     uint32
			deviceAddress uint32
			propType      uint32
		)

		propLen := uint32(4)

		err = CMGetDevNodeProperty(devNode, DEVPKEYDeviceBusNumber, &propType, unsafe.Pointer(&busNumber), &propLen)
		if err != nil {
			return nil, err
		}

		if propType != DEVPROP_TYPE_UINT32 {
			return nil, fmt.Errorf("unexpected property type: 0x%08X", propType)
		}

		err = CMGetDevNodeProperty(devNode, DEVPKEYDeviceAddress, &propType, unsafe.Pointer(&deviceAddress), &propLen)
		if err != nil {
			return nil, err
		}

		if propType != DEVPROP_TYPE_UINT32 {
			return nil, fmt.Errorf("unexpected property type: 0x%08X", propType)
		}

		devices = append(devices, Device{
			InstanceID:     windows.UTF16ToString(deviceInstanceID),
			BusNumber:      win32.UINT(busNumber),
			DeviceNumber:   win32.UINT(deviceAddress >> 16),
			FunctionNumber: win32.UINT(deviceAddress & 0xFFFF),
		})
	}

	return devices, nil
}

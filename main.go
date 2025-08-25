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

package main

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	// Configuration Manager error codes
	CR_SUCCESS = 0

	// Filter flags for CM_Get_Device_ID_ListW
	CM_GETIDLIST_FILTER_NONE               = 0x00000000
	CM_GETIDLIST_FILTER_ENUMERATOR         = 0x00000001
	CM_GETIDLIST_FILTER_SERVICE            = 0x00000002
	CM_GETIDLIST_FILTER_EJECTRELATIONS     = 0x00000004
	CM_GETIDLIST_FILTER_REMOVALRELATIONS   = 0x00000008
	CM_GETIDLIST_FILTER_POWERRELATIONS     = 0x00000010
	CM_GETIDLIST_FILTER_BUSRELATIONS       = 0x00000020
	CM_GETIDLIST_FILTER_PRESENT            = 0x00000100
	CM_GETIDLIST_FILTER_CLASS              = 0x00000200
	CM_GETIDLIST_FILTER_TRANSPORTRELATIONS = 0x00000400
	CM_GETIDLIST_DONOTGENERATE             = 0x10000040
)

var (
	cfgmgr32                  = windows.NewLazySystemDLL("cfgmgr32.dll")
	procCMGetDeviceIDListW    = cfgmgr32.NewProc("CM_Get_Device_ID_ListW")
	procCMGetDeviceIDListSize = cfgmgr32.NewProc("CM_Get_Device_ID_List_SizeW")
)

// GetDeviceIDListSize gets the required buffer size for the device ID list
func GetDeviceIDListSize(filter string, flags uint32) (uint32, error) {
	var size uint32
	var filterPtr *uint16

	if filter != "" {
		utf16Filter, err := syscall.UTF16PtrFromString(filter)
		if err != nil {
			return 0, err
		}
		filterPtr = utf16Filter
	}

	ret, _, _ := procCMGetDeviceIDListSize.Call(
		uintptr(unsafe.Pointer(filterPtr)),
		uintptr(unsafe.Pointer(&size)),
		uintptr(flags),
	)

	if ret != CR_SUCCESS {
		return 0, fmt.Errorf("CM_Get_Device_ID_List_SizeW failed with code: %d", ret)
	}

	return size, nil
}

// GetDeviceIDList retrieves the list of device IDs
func GetDeviceIDList(filter string, flags uint32) ([]string, error) {
	// First, get the required buffer size
	bufferSize, err := GetDeviceIDListSize(filter, flags)
	if err != nil {
		return nil, err
	}

	// Allocate buffer
	buffer := make([]uint16, bufferSize)
	var filterPtr *uint16

	if filter != "" {
		utf16Filter, err := syscall.UTF16PtrFromString(filter)
		if err != nil {
			return nil, err
		}
		filterPtr = utf16Filter
	}

	ret, _, _ := procCMGetDeviceIDListW.Call(
		uintptr(unsafe.Pointer(filterPtr)),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(bufferSize),
		uintptr(flags),
	)

	if ret != CR_SUCCESS {
		return nil, fmt.Errorf("CM_Get_Device_ID_ListW failed with code: %d", ret)
	}

	// Parse the double-null-terminated string list
	var devices []string
	start := 0
	for i := 0; i < len(buffer); i++ {
		if buffer[i] == 0 {
			if i > start {
				deviceID := syscall.UTF16ToString(buffer[start:i])
				devices = append(devices, deviceID)
			}
			start = i + 1

			// Check for double null termination
			if i+1 < len(buffer) && buffer[i+1] == 0 {
				break
			}
		}
	}

	return devices, nil
}

func main() {
	// Example: Get all present devices
	devices, err := GetDeviceIDList("PCI\\VEN_10DE&DEV_1B81&SUBSYS_61733842&REV_A1", CM_GETIDLIST_FILTER_ENUMERATOR)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Found %d present devices:\n", len(devices))
	for i, device := range devices {
		fmt.Printf("%d: %s\n", i+1, device)
		if i >= 9 { // Limit output for readability
			fmt.Printf("... and %d more devices\n", len(devices)-10)
			break
		}
	}
}

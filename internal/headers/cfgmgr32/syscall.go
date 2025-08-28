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

//nolint:gochecknoglobals
var (
	cfgmgr32 = windows.NewLazySystemDLL("cfgmgr32.dll")

	procCMGetDeviceIDListW    = cfgmgr32.NewProc("CM_Get_Device_ID_ListW")
	procCMGetDeviceIDListSize = cfgmgr32.NewProc("CM_Get_Device_ID_List_SizeW")
	procCMGetDevNodePropertyW = cfgmgr32.NewProc("CM_Get_DevNode_PropertyW")
	procCMLocateDevNodeW      = cfgmgr32.NewProc("CM_Locate_DevNodeW")
)

func CMGetDeviceIDListSize(filter *win32.LPWSTR, size *uint32) error {
	ret, _, _ := procCMGetDeviceIDListSize.Call(
		uintptr(unsafe.Pointer(size)),
		filter.Pointer(),
		uintptr(CM_GETIDLIST_FILTER_PRESENT|CM_GETIDLIST_FILTER_ENUMERATOR),
	)

	if ret != CR_SUCCESS {
		return fmt.Errorf("CMGetDeviceIDListSize failed: 0x%02X", ret)
	}

	return nil
}

func CMGetDeviceIDList(filter *win32.LPWSTR, buf []uint16) error {
	ret, _, _ := procCMGetDeviceIDListW.Call(
		filter.Pointer(),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
		uintptr(CM_GETIDLIST_FILTER_PRESENT|CM_GETIDLIST_FILTER_ENUMERATOR),
	)

	if ret != CR_SUCCESS {
		return fmt.Errorf("CMGetDeviceIDList failed: 0x%02X", ret)
	}

	return nil
}

func CMLocateDevNode(devInst **windows.Handle, deviceID []uint16) error {
	ret, _, _ := procCMLocateDevNodeW.Call(
		uintptr(unsafe.Pointer(devInst)),
		uintptr(unsafe.Pointer(&deviceID[0])),
		0,
	)

	if ret != CR_SUCCESS {
		return fmt.Errorf("CMLocateDevNode failed: 0x%02X", ret)
	}

	return nil
}

func CMGetDevNodeProperty(devInst *windows.Handle, propKey *DEVPROPKEY, propType *uint32, buf unsafe.Pointer, bufLen *uint32) error {
	ret, _, _ := procCMGetDevNodePropertyW.Call(
		uintptr(unsafe.Pointer(devInst)),
		uintptr(unsafe.Pointer(propKey)),
		uintptr(unsafe.Pointer(propType)),
		uintptr(buf),
		uintptr(unsafe.Pointer(bufLen)),
		0,
	)

	if ret != CR_SUCCESS {
		return fmt.Errorf("CMGetDevNodeProperty failed: 0x%02X", ret)
	}

	return nil
}

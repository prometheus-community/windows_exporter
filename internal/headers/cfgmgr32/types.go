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
	"github.com/go-ole/go-ole"
	"github.com/prometheus-community/windows_exporter/internal/headers/win32"
)

const (
	// Configuration Manager return codes
	CR_SUCCESS      = 0x00
	CR_BUFFER_SMALL = 0x1a

	// Filter flags
	CM_GETIDLIST_FILTER_NONE               = 0x00000000
	CM_GETIDLIST_FILTER_ENUMERATOR         = 0x00000001
	CM_GETIDLIST_FILTER_SERVICE            = 0x00000002
	CM_GETIDLIST_FILTER_EJECTRELATIONS     = 0x00000004
	CM_GETIDLIST_FILTER_REMOVALRELATIONS   = 0x00000008
	CM_GETIDLIST_FILTER_POWERRELATIONS     = 0x00000010
	CM_GETIDLIST_FILTER_BUSRELATIONS       = 0x00000020
	CM_GETIDLIST_DONOTGENERATE             = 0x10000040
	CM_GETIDLIST_FILTER_PRESENT            = 0x00000100
	CM_GETIDLIST_FILTER_CLASS              = 0x00000200
	CM_GETIDLIST_FILTER_TRANSPORTRELATIONS = 0x00000080
	CM_GETIDLIST_FILTER_BITS               = 0x100003FF

	DEVPROP_TYPE_UINT32 uint32 = 0x00000007
)

// DEVPROPKEY represents a device property key (GUID + pid)
type DEVPROPKEY struct {
	FmtID ole.GUID
	PID   uint32
}

type Device struct {
	InstanceID     string
	BusNumber      win32.UINT
	DeviceNumber   win32.UINT
	FunctionNumber win32.UINT
}

var (
	// https://github.com/Infinidat/infi.devicemanager/blob/8be9ead6b04ff45c63d9e3bc70d82cceafb75c47/src/infi/devicemanager/setupapi/properties.py#L138C1-L143C34
	DEVPKEYDeviceBusNumber = &DEVPROPKEY{
		FmtID: ole.GUID{
			Data1: 0xa45c254e,
			Data2: 0xdf1c,
			Data3: 0x4efd,
			Data4: [8]byte{0x80, 0x20, 0x67, 0xd1, 0x46, 0xa8, 0x50, 0xe0},
		},
		PID: 23, // DEVPROP_TYPE_UINT32
	}

	// https://github.com/Infinidat/infi.devicemanager/blob/8be9ead6b04ff45c63d9e3bc70d82cceafb75c47/src/infi/devicemanager/setupapi/properties.py#L187-L192
	DEVPKEYDeviceAddress = &DEVPROPKEY{
		FmtID: ole.GUID{
			Data1: 0xa45c254e,
			Data2: 0xdf1c,
			Data3: 0x4efd,
			Data4: [8]byte{0x80, 0x20, 0x67, 0xd1, 0x46, 0xa8, 0x50, 0xe0},
		},
		PID: 30, // DEVPROP_TYPE_UINT32
	}
)

// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

package cfgmgr32

import (
	"github.com/prometheus-community/windows_exporter/internal/headers/win32"
	"golang.org/x/sys/windows"
)

const (
	// Configuration Manager return codes
	CR_SUCCESS = 0x00

	// Filter flags
	CM_GETIDLIST_FILTER_ENUMERATOR = 0x00000001
	CM_GETIDLIST_FILTER_PRESENT    = 0x00000100

	DEVPROP_TYPE_UINT32 uint32 = 0x00000007
)

// DEVPROPKEY represents a device property key (GUID + pid)
type DEVPROPKEY struct {
	FmtID windows.GUID
	PID   uint32
}

type Device struct {
	InstanceID     string
	BusNumber      win32.UINT
	DeviceNumber   win32.UINT
	FunctionNumber win32.UINT
}

//nolint:gochecknoglobals
var (
	// https://github.com/Infinidat/infi.devicemanager/blob/8be9ead6b04ff45c63d9e3bc70d82cceafb75c47/src/infi/devicemanager/setupapi/properties.py#L138C1-L143C34
	DEVPKEYDeviceBusNumber = &DEVPROPKEY{
		FmtID: windows.GUID{
			Data1: 0xa45c254e,
			Data2: 0xdf1c,
			Data3: 0x4efd,
			Data4: [8]byte{0x80, 0x20, 0x67, 0xd1, 0x46, 0xa8, 0x50, 0xe0},
		},
		PID: 23, // DEVPROP_TYPE_UINT32
	}

	// https://github.com/Infinidat/infi.devicemanager/blob/8be9ead6b04ff45c63d9e3bc70d82cceafb75c47/src/infi/devicemanager/setupapi/properties.py#L187-L192
	DEVPKEYDeviceAddress = &DEVPROPKEY{
		FmtID: windows.GUID{
			Data1: 0xa45c254e,
			Data2: 0xdf1c,
			Data3: 0x4efd,
			Data4: [8]byte{0x80, 0x20, 0x67, 0xd1, 0x46, 0xa8, 0x50, 0xe0},
		},
		PID: 30, // DEVPROP_TYPE_UINT32
	}
)

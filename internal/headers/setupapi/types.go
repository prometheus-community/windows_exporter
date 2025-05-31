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

import "golang.org/x/sys/windows"

const (
	DIGCF_PRESENT                     = 0x00000002
	SPDRP_DEVICEDESC                  = 0x00000000
	SPDRP_FRIENDLYNAME                = 0x0000000C
	SPDRP_HARDWAREID                  = 0x00000001
	SPDRP_PHYSICAL_DEVICE_OBJECT_NAME = 0x0000000E
)

type SP_DEVINFO_DATA struct {
	CbSize    uint32
	ClassGuid windows.GUID
	DevInst   uint32
	_         uintptr // Reserved
}

type GPUDevice struct {
	DeviceDesc               string
	FriendlyName             string
	HardwareID               string
	PhysicalDeviceObjectName string
}

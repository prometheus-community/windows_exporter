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

package gdi32

import (
	"unsafe"

	"github.com/prometheus-community/windows_exporter/internal/headers/win32"
	"golang.org/x/sys/windows"
)

type D3DKMT_HANDLE = win32.UINT

type D3DKMT_OPENADAPTERFROMLUID struct {
	AdapterLUID windows.LUID
	HAdapter    D3DKMT_HANDLE
}

type D3DKMT_CLOSEADAPTER struct {
	HAdapter D3DKMT_HANDLE
}

type D3DKMT_QUERYADAPTERINFO struct {
	hAdapter              D3DKMT_HANDLE
	queryType             int32
	pPrivateDriverData    unsafe.Pointer
	privateDriverDataSize uint32
}

type D3DKMT_ENUMADAPTERS2 struct {
	NumAdapters uint32
	PAdapters   *D3DKMT_ADAPTERINFO
}

type D3DKMT_ADAPTERINFO struct {
	HAdapter     D3DKMT_HANDLE
	AdapterLUID  windows.LUID
	NumOfSources win32.ULONG
	Present      win32.BOOL
}

type D3DKMT_ADAPTERREGISTRYINFO struct {
	AdapterString [win32.MAX_PATH]uint16
	BiosString    [win32.MAX_PATH]uint16
	DacType       [win32.MAX_PATH]uint16
	ChipType      [win32.MAX_PATH]uint16
}

type D3DKMT_SEGMENTSIZEINFO struct {
	DedicatedVideoMemorySize  uint64
	DedicatedSystemMemorySize uint64
	SharedSystemMemorySize    uint64
}

type D3DKMT_ADAPTERADDRESS struct {
	BusNumber      win32.UINT
	DeviceNumber   win32.UINT
	FunctionNumber win32.UINT
}

type GPUDevice struct {
	AdapterString             string
	LUID                      windows.LUID
	DedicatedVideoMemorySize  uint64
	DedicatedSystemMemorySize uint64
	SharedSystemMemorySize    uint64
	BusNumber                 win32.UINT
	DeviceNumber              win32.UINT
	FunctionNumber            win32.UINT
}

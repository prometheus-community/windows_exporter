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

type D3DKMT_QUERY_DEVICE_IDS struct {
	PhysicalAdapterIndex win32.UINT
	DeviceIds            struct {
		VendorID    win32.UINT
		DeviceID    win32.UINT
		SubVendorID win32.UINT
		SubSystemID win32.UINT
		RevisionID  win32.UINT
		BusType     win32.UINT
	}
}

type GPUDevice struct {
	AdapterString             string
	LUID                      windows.LUID
	DeviceID                  string
	DedicatedVideoMemorySize  uint64
	DedicatedSystemMemorySize uint64
	SharedSystemMemorySize    uint64
	BusNumber                 win32.UINT
	DeviceNumber              win32.UINT
	FunctionNumber            win32.UINT
	AdapterType               D3DKMT_ADAPTERTYPE
}

// D3DKMT_NODE_PERFDATA is the output of KMTQAITYPE_NODEPERFDATA.
// https://learn.microsoft.com/en-us/windows-hardware/drivers/ddi/d3dkmthk/ns-d3dkmthk-_d3dkmt_node_perfdata
type D3DKMT_NODE_PERFDATA struct {
	NodeOrdinal          uint32
	PhysicalAdapterIndex uint32
	// Frequency is the current clock frequency of the engine in hertz.
	Frequency uint64
	// MaxFrequency is the maximum clock frequency of the engine in hertz, while not overclocked.
	MaxFrequency uint64
	// MaxFrequencyOC is the maximum clock frequency of the engine in hertz, while overclocked.
	MaxFrequencyOC uint64
	// Voltage is the current voltage of the engine in milli volts.
	Voltage      win32.ULONG
	VoltageMax   win32.ULONG
	VoltageMaxOC win32.ULONG
	// MaxTransitionLatency is the maximum transition latency to change the frequency in 100 nanoseconds.
	MaxTransitionLatency uint64
}

// D3DKMT_ADAPTER_PERFDATA is the output of KMTQAITYPE_ADAPTERPERFDATA.
// https://learn.microsoft.com/en-us/windows-hardware/drivers/ddi/d3dkmthk/ns-d3dkmthk-_d3dkmt_adapter_perfdata
type D3DKMT_ADAPTER_PERFDATA struct {
	PhysicalAdapterIndex uint32
	// MemoryFrequency is the clock frequency of the memory in hertz.
	MemoryFrequency uint64
	// MaxMemoryFrequency is the max clock frequency of the memory in hertz, while not overclocked.
	MaxMemoryFrequency uint64
	// MaxMemoryFrequencyOC is the max clock frequency of the memory in hertz, while overclocked.
	MaxMemoryFrequencyOC uint64
	// MemoryBandwidth is the total amount of memory transferred in bytes.
	MemoryBandwidth uint64
	// PCIEBandwidth is the total amount of memory transferred over PCIe in bytes.
	PCIEBandwidth uint64
	// FanRPM is the current rpm of the main fan.
	FanRPM win32.ULONG
	// Power is the current power draw of the adapter in tenths of a percent (1 = 0.1%).
	Power win32.ULONG
	// Temperature is the main temperature sensor reading in tenths of a degree Celsius (1 = 0.1 °C).
	Temperature win32.ULONG
	// PowerStateOverride is 1 if the GPU is powered on, otherwise 0.
	PowerStateOverride uint8
}

// D3DKMT_ADAPTER_PERFDATACAPS is the output of KMTQAITYPE_ADAPTERPERFDATA_CAPS.
// https://learn.microsoft.com/en-us/windows-hardware/drivers/ddi/d3dkmthk/ns-d3dkmthk-_d3dkmt_adapter_perfdatacaps
type D3DKMT_ADAPTER_PERFDATACAPS struct {
	PhysicalAdapterIndex uint32
	// MaxMemoryBandwidth is the max memory bandwidth in bytes for 1 second.
	MaxMemoryBandwidth uint64
	// MaxPCIEBandwidth is the max PCIe bandwidth in bytes for 1 second.
	MaxPCIEBandwidth uint64
	// MaxFanRPM is the max fan rpm.
	MaxFanRPM win32.ULONG
	// TemperatureMax is the max temperature before damage, in tenths of a degree Celsius.
	TemperatureMax win32.ULONG
	// TemperatureWarning is the temperature at which the GPU starts throttling, in tenths of a degree Celsius.
	TemperatureWarning win32.ULONG
}

// D3DKMT_GPUVERSION is the output of KMTQAITYPE_GPUVERSION.
// https://learn.microsoft.com/en-us/windows-hardware/drivers/ddi/d3dkmthk/ns-d3dkmthk-_d3dkmt_gpuversion
type D3DKMT_GPUVERSION struct {
	PhysicalAdapterIndex uint32
	BiosVersion          [32]uint16
	GpuArchitecture      [32]uint16
}

// IsSoftwareDevice reports whether the adapter is a software device,
// e.g. the Microsoft Basic Render Driver.
func (d GPUDevice) IsSoftwareDevice() bool {
	return d.AdapterType&D3DKMT_ADAPTERTYPE_SOFTWARE_DEVICE != 0
}

// D3DKMT_ADAPTERTYPE is the bitfield returned by KMTQAITYPE_ADAPTERTYPE.
// https://learn.microsoft.com/en-us/windows-hardware/drivers/ddi/d3dkmthk/ns-d3dkmthk-_d3dkmt_adaptertype
type D3DKMT_ADAPTERTYPE uint32

const (
	D3DKMT_ADAPTERTYPE_RENDER_SUPPORTED         D3DKMT_ADAPTERTYPE = 1 << 0
	D3DKMT_ADAPTERTYPE_DISPLAY_SUPPORTED        D3DKMT_ADAPTERTYPE = 1 << 1
	D3DKMT_ADAPTERTYPE_SOFTWARE_DEVICE          D3DKMT_ADAPTERTYPE = 1 << 2
	D3DKMT_ADAPTERTYPE_POST_DEVICE              D3DKMT_ADAPTERTYPE = 1 << 3
	D3DKMT_ADAPTERTYPE_HYBRID_DISCRETE          D3DKMT_ADAPTERTYPE = 1 << 4
	D3DKMT_ADAPTERTYPE_HYBRID_INTEGRATED        D3DKMT_ADAPTERTYPE = 1 << 5
	D3DKMT_ADAPTERTYPE_INDIRECT_DISPLAY_DEVICE  D3DKMT_ADAPTERTYPE = 1 << 6
	D3DKMT_ADAPTERTYPE_PARAVIRTUALIZED          D3DKMT_ADAPTERTYPE = 1 << 7
	D3DKMT_ADAPTERTYPE_ACG_SUPPORTED            D3DKMT_ADAPTERTYPE = 1 << 8
	D3DKMT_ADAPTERTYPE_SET_TIMINGS_FROM_VIDPN   D3DKMT_ADAPTERTYPE = 1 << 9
	D3DKMT_ADAPTERTYPE_DETACHABLE               D3DKMT_ADAPTERTYPE = 1 << 10
	D3DKMT_ADAPTERTYPE_COMPUTE_ONLY             D3DKMT_ADAPTERTYPE = 1 << 11
	D3DKMT_ADAPTERTYPE_PROTOTYPE                D3DKMT_ADAPTERTYPE = 1 << 12
	D3DKMT_ADAPTERTYPE_RUNTIME_POWER_MANAGEMENT D3DKMT_ADAPTERTYPE = 1 << 13
)

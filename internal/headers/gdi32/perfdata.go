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
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// OpenAdapterFromLUID opens a handle to the adapter with the given LUID.
// The handle must be released with CloseAdapter.
func OpenAdapterFromLUID(luid windows.LUID) (D3DKMT_HANDLE, error) {
	openAdapter := D3DKMT_OPENADAPTERFROMLUID{
		AdapterLUID: luid,
	}

	if err := D3DKMTOpenAdapterFromLuid(&openAdapter); err != nil {
		return 0, err
	}

	return openAdapter.HAdapter, nil
}

// CloseAdapter closes a handle opened by OpenAdapterFromLUID.
func CloseAdapter(hAdapter D3DKMT_HANDLE) error {
	return D3DKMTCloseAdapter(&D3DKMT_CLOSEADAPTER{
		HAdapter: hAdapter,
	})
}

// queryAdapterInfo calls D3DKMTQueryAdapterInfo with data as private driver data.
// data is used as input and output.
func queryAdapterInfo[T any](hAdapter D3DKMT_HANDLE, queryType int32, data *T) error {
	query := D3DKMT_QUERYADAPTERINFO{
		hAdapter:              hAdapter,
		queryType:             queryType,
		pPrivateDriverData:    unsafe.Pointer(data),
		privateDriverDataSize: uint32(unsafe.Sizeof(*data)),
	}

	return D3DKMTQueryAdapterInfo(&query)
}

// QueryPhysicalAdapterCount returns the number of physical adapters in the LDA chain of the adapter.
func QueryPhysicalAdapterCount(hAdapter D3DKMT_HANDLE) (uint32, error) {
	var count uint32

	if err := queryAdapterInfo(hAdapter, KMTQAITYPE_PHYSICALADAPTERCOUNT, &count); err != nil {
		return 0, fmt.Errorf("physical adapter count: %w", err)
	}

	return count, nil
}

// QueryWDDMVersion returns the WDDM version of the display miniport driver
// as D3DKMT_DRIVERVERSION value, e.g. 3200 for WDDM 3.2.
func QueryWDDMVersion(hAdapter D3DKMT_HANDLE) (uint32, error) {
	var version int32

	if err := queryAdapterInfo(hAdapter, KMTQAITYPE_DRIVERVERSION, &version); err != nil {
		return 0, fmt.Errorf("WDDM version: %w", err)
	}

	return uint32(version), nil
}

// QueryKMDDriverVersion returns the kernel mode driver version, packed as four 16-bit parts.
func QueryKMDDriverVersion(hAdapter D3DKMT_HANDLE) (uint64, error) {
	var version uint64

	if err := queryAdapterInfo(hAdapter, KMTQAITYPE_KMD_DRIVER_VERSION, &version); err != nil {
		return 0, fmt.Errorf("KMD driver version: %w", err)
	}

	return version, nil
}

// QueryAdapterPerfData returns the current performance data of a physical adapter.
func QueryAdapterPerfData(hAdapter D3DKMT_HANDLE, physicalAdapterIndex uint32) (D3DKMT_ADAPTER_PERFDATA, error) {
	data := D3DKMT_ADAPTER_PERFDATA{
		PhysicalAdapterIndex: physicalAdapterIndex,
	}

	if err := queryAdapterInfo(hAdapter, KMTQAITYPE_ADAPTERPERFDATA, &data); err != nil {
		return data, fmt.Errorf("adapter perf data: %w", err)
	}

	return data, nil
}

// QueryAdapterPerfDataCaps returns the performance data capabilities of a physical adapter.
func QueryAdapterPerfDataCaps(hAdapter D3DKMT_HANDLE, physicalAdapterIndex uint32) (D3DKMT_ADAPTER_PERFDATACAPS, error) {
	data := D3DKMT_ADAPTER_PERFDATACAPS{
		PhysicalAdapterIndex: physicalAdapterIndex,
	}

	if err := queryAdapterInfo(hAdapter, KMTQAITYPE_ADAPTERPERFDATA_CAPS, &data); err != nil {
		return data, fmt.Errorf("adapter perf data caps: %w", err)
	}

	return data, nil
}

// QueryNodePerfData returns the current performance data of an engine (node) of a physical adapter.
// The node ordinal matches the eng_N part of the GPU Engine performance counter instances.
// Querying a node ordinal past the last node returns an error.
func QueryNodePerfData(hAdapter D3DKMT_HANDLE, physicalAdapterIndex, nodeOrdinal uint32) (D3DKMT_NODE_PERFDATA, error) {
	data := D3DKMT_NODE_PERFDATA{
		NodeOrdinal:          nodeOrdinal,
		PhysicalAdapterIndex: physicalAdapterIndex,
	}

	if err := queryAdapterInfo(hAdapter, KMTQAITYPE_NODEPERFDATA, &data); err != nil {
		return data, fmt.Errorf("node %d perf data: %w", nodeOrdinal, err)
	}

	return data, nil
}

// QueryGPUVersion returns the BIOS version and architecture of a physical adapter.
func QueryGPUVersion(hAdapter D3DKMT_HANDLE, physicalAdapterIndex uint32) (D3DKMT_GPUVERSION, error) {
	data := D3DKMT_GPUVERSION{
		PhysicalAdapterIndex: physicalAdapterIndex,
	}

	if err := queryAdapterInfo(hAdapter, KMTQUITYPE_GPUVERSION, &data); err != nil {
		return data, fmt.Errorf("GPU version: %w", err)
	}

	return data, nil
}

// FormatKMDDriverVersion formats a version returned by QueryKMDDriverVersion as a.b.c.d.
func FormatKMDDriverVersion(version uint64) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		uint16(version>>48),
		uint16(version>>32),
		uint16(version>>16),
		uint16(version),
	)
}

// FormatWDDMVersion formats a version returned by QueryWDDMVersion as major.minor, e.g. 3200 as 3.2.
func FormatWDDMVersion(version uint32) string {
	return fmt.Sprintf("%d.%d", version/1000, version%1000/100)
}

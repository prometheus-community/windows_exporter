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
	"errors"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// https://learn.microsoft.com/en-us/windows-hardware/drivers/ddi/d3dkmthk/ne-d3dkmthk-_kmtqueryadapterinfotype
// https://github.com/nalilord/AMDPlugin/blob/bb405b6d58ea543ff630f3488384473bee79f447/Common/d3dkmthk.pas#L54
const (
	// KMTQAITYPE_GETSEGMENTSIZE pPrivateDriverData points to a D3DKMT_SEGMENTSIZEINFO structure that contains information about the size of memory and aperture segments.
	KMTQAITYPE_GETSEGMENTSIZE = 3
	// KMTQAITYPE_ADAPTERADDRESS pPrivateDriverData points to a D3DKMT_ADAPTERADDRESS structure that contains information about the physical location on the PCI bus of the adapter.
	KMTQAITYPE_ADAPTERADDRESS = 6
	// KMTQAITYPE_ADAPTERREGISTRYINFO pPrivateDriverData points to a D3DKMT_ADAPTERREGISTRYINFO structure that contains registry information about the graphics adapter.
	KMTQAITYPE_ADAPTERREGISTRYINFO = 8
	// KMTQAITYPE_ADAPTERGUID pPrivateDriverData points to a D3DKMT_QUERY_DEVICE_IDS structure that specifies the device ID(s) of the physical adapters. Supported starting with Windows 10 (WDDM 2.0).
	KMTQAITYPE_PHYSICALADAPTERDEVICEIDS = 31
)

var ErrNoGPUDevices = errors.New("no GPU devices found")

func GetGPUDeviceByLUID(adapterLUID windows.LUID) (GPUDevice, error) {
	open := D3DKMT_OPENADAPTERFROMLUID{
		AdapterLUID: adapterLUID,
	}

	if err := D3DKMTOpenAdapterFromLuid(&open); err != nil {
		return GPUDevice{}, fmt.Errorf("D3DKMTOpenAdapterFromLuid failed: %w", err)
	}

	errs := make([]error, 0)

	gpuDevice, err := GetGPUDevice(open.HAdapter)
	if err != nil {
		errs = append(errs, fmt.Errorf("GetGPUDevice failed: %w", err))
	}

	if err := D3DKMTCloseAdapter(&D3DKMT_CLOSEADAPTER{
		HAdapter: open.HAdapter,
	}); err != nil {
		errs = append(errs, fmt.Errorf("D3DKMTCloseAdapter failed: %w", err))
	}

	if len(errs) > 0 {
		return gpuDevice, fmt.Errorf("errors occurred while getting GPU device: %w", errors.Join(errs...))
	}

	gpuDevice.LUID = adapterLUID

	return gpuDevice, nil
}

func GetGPUDevice(hAdapter D3DKMT_HANDLE) (GPUDevice, error) {
	var gpuDevice GPUDevice

	// Try segment size first
	var size D3DKMT_SEGMENTSIZEINFO

	query := D3DKMT_QUERYADAPTERINFO{
		hAdapter:              hAdapter,
		queryType:             KMTQAITYPE_GETSEGMENTSIZE,
		pPrivateDriverData:    unsafe.Pointer(&size),
		privateDriverDataSize: uint32(unsafe.Sizeof(size)),
	}

	if err := D3DKMTQueryAdapterInfo(&query); err != nil {
		return gpuDevice, fmt.Errorf("D3DKMTQueryAdapterInfo (segment size) failed: %w", err)
	}

	gpuDevice.DedicatedVideoMemorySize = size.DedicatedVideoMemorySize
	gpuDevice.DedicatedSystemMemorySize = size.DedicatedSystemMemorySize
	gpuDevice.SharedSystemMemorySize = size.SharedSystemMemorySize

	// Now try registry info
	var address D3DKMT_ADAPTERADDRESS

	query.queryType = KMTQAITYPE_ADAPTERADDRESS
	query.pPrivateDriverData = unsafe.Pointer(&address)
	query.privateDriverDataSize = uint32(unsafe.Sizeof(address))

	if err := D3DKMTQueryAdapterInfo(&query); err != nil {
		return gpuDevice, fmt.Errorf("D3DKMTQueryAdapterInfo (adapter address) failed: %w", err)
	}

	gpuDevice.BusNumber = address.BusNumber
	gpuDevice.DeviceNumber = address.DeviceNumber
	gpuDevice.FunctionNumber = address.FunctionNumber

	// Now try registry info
	var info D3DKMT_ADAPTERREGISTRYINFO

	query.queryType = KMTQAITYPE_ADAPTERREGISTRYINFO
	query.pPrivateDriverData = unsafe.Pointer(&info)
	query.privateDriverDataSize = uint32(unsafe.Sizeof(info))

	if err := D3DKMTQueryAdapterInfo(&query); err != nil && !errors.Is(err, windows.ERROR_FILE_NOT_FOUND) {
		return gpuDevice, fmt.Errorf("D3DKMTQueryAdapterInfo (info) failed: %w", err)
	}

	gpuDevice.AdapterString = windows.UTF16ToString(info.AdapterString[:])

	var deviceIDs D3DKMT_QUERY_DEVICE_IDS

	query.queryType = KMTQAITYPE_PHYSICALADAPTERDEVICEIDS
	query.pPrivateDriverData = unsafe.Pointer(&deviceIDs)
	query.privateDriverDataSize = uint32(unsafe.Sizeof(deviceIDs))

	if err := D3DKMTQueryAdapterInfo(&query); err != nil && !errors.Is(err, windows.ERROR_FILE_NOT_FOUND) {
		return gpuDevice, fmt.Errorf("D3DKMTQueryAdapterInfo (Device IDs) failed: %w", err)
	}

	gpuDevice.DeviceID = formatPNPDeviceID(deviceIDs, address)

	return gpuDevice, nil
}

func GetGPUDevices() ([]GPUDevice, error) {
	gpuDevices := make([]GPUDevice, 0, 2)

	// First call: Get the number of adapters
	enumAdapters := D3DKMT_ENUMADAPTERS2{
		NumAdapters: 0,
		PAdapters:   nil,
	}

	if err := D3DKMTEnumAdapters2(&enumAdapters); err != nil {
		return gpuDevices, fmt.Errorf("D3DKMTEnumAdapters2 (get count) failed: %w", err)
	}

	if enumAdapters.NumAdapters == 0 {
		return gpuDevices, ErrNoGPUDevices
	}

	// Second call: Get the actual adapter information
	pAdapters := make([]D3DKMT_ADAPTERINFO, enumAdapters.NumAdapters)
	enumAdapters.PAdapters = &pAdapters[0]

	if err := D3DKMTEnumAdapters2(&enumAdapters); err != nil {
		return gpuDevices, fmt.Errorf("D3DKMTEnumAdapters2 (get adapters) failed: %w", err)
	}

	var errs []error

	// Process each adapter
	for i := range enumAdapters.NumAdapters {
		adapter := pAdapters[i]
		// Validate handle before using it
		if adapter.HAdapter == 0 {
			errs = append(errs, fmt.Errorf("adapter %d has null handle", i))

			continue
		}

		func() {
			defer func() {
				if closeErr := D3DKMTCloseAdapter(&D3DKMT_CLOSEADAPTER{
					HAdapter: adapter.HAdapter,
				}); closeErr != nil {
					errs = append(errs, fmt.Errorf("failed to close adapter %v: %w", adapter.AdapterLUID, closeErr))
				}
			}()

			gpuDevice, err := GetGPUDevice(adapter.HAdapter)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to get GPU device for adapter %v: %w", adapter.AdapterLUID, err))

				return
			}

			gpuDevice.LUID = adapter.AdapterLUID

			gpuDevices = append(gpuDevices, gpuDevice)
		}()
	}

	if len(errs) > 0 {
		return gpuDevices, errors.Join(errs...)
	}

	if len(gpuDevices) == 0 {
		return gpuDevices, ErrNoGPUDevices
	}

	return gpuDevices, nil
}

func formatPNPDeviceID(deviceIDs D3DKMT_QUERY_DEVICE_IDS, address D3DKMT_ADAPTERADDRESS) string {
	return fmt.Sprintf("PCI\\VEN_%04X&DEV_%04X&SUBSYS_%04X%04X&REV_%02X",
		uint16(deviceIDs.DeviceIds.VendorID),
		uint16(deviceIDs.DeviceIds.DeviceID),
		uint16(deviceIDs.DeviceIds.SubSystemID),
		uint16(deviceIDs.DeviceIds.SubVendorID),
		uint8(deviceIDs.DeviceIds.RevisionID),
	)
}

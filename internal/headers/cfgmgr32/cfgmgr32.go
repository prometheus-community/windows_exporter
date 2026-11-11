// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

package cfgmgr32

import (
	"errors"
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

	var errs []error

	// Devices whose properties can't be read are skipped. The remaining devices
	// are returned together with an error describing the skipped ones.
	for _, deviceInstanceID := range deviceInstanceIDs {
		device, err := getDevice(deviceInstanceID)
		if err != nil {
			errs = append(errs, fmt.Errorf("device %s: %w", windows.UTF16ToString(deviceInstanceID), err))

			continue
		}

		devices = append(devices, device)
	}

	return devices, errors.Join(errs...)
}

func getDevice(deviceInstanceID []uint16) (Device, error) {
	var devInst uint32

	if err := CMLocateDevNode(&devInst, deviceInstanceID); err != nil {
		return Device{}, err
	}

	busNumber, err := getDevNodePropertyUint32(devInst, DEVPKEYDeviceBusNumber)
	if err != nil {
		return Device{}, fmt.Errorf("bus number: %w", err)
	}

	deviceAddress, err := getDevNodePropertyUint32(devInst, DEVPKEYDeviceAddress)
	if err != nil {
		return Device{}, fmt.Errorf("device address: %w", err)
	}

	return Device{
		InstanceID:     windows.UTF16ToString(deviceInstanceID),
		BusNumber:      win32.UINT(busNumber),
		DeviceNumber:   win32.UINT(deviceAddress >> 16),
		FunctionNumber: win32.UINT(deviceAddress & 0xFFFF),
	}, nil
}

func getDevNodePropertyUint32(devInst uint32, propKey *DEVPROPKEY) (uint32, error) {
	var (
		value    uint32
		propType uint32
	)

	propLen := uint32(unsafe.Sizeof(value))

	if err := CMGetDevNodeProperty(devInst, propKey, &propType, unsafe.Pointer(&value), &propLen); err != nil {
		return 0, err
	}

	if propType != DEVPROP_TYPE_UINT32 {
		return 0, fmt.Errorf("unexpected property type: 0x%08X", propType)
	}

	return value, nil
}

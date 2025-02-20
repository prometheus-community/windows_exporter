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

package hcs

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var (
	modComputeCore = windows.NewLazySystemDLL("computecore.dll")

	procHcsCreateOperation            = modComputeCore.NewProc("HcsCreateOperation")
	procHcsWaitForOperationResult     = modComputeCore.NewProc("HcsWaitForOperationResult")
	procHcsCloseOperation             = modComputeCore.NewProc("HcsCloseOperation")
	procHcsEnumerateComputeSystems    = modComputeCore.NewProc("HcsEnumerateComputeSystems")
	procHcsOpenComputeSystem          = modComputeCore.NewProc("HcsOpenComputeSystem")
	procHcsGetComputeSystemProperties = modComputeCore.NewProc("HcsGetComputeSystemProperties")
	procHcsCloseComputeSystem         = modComputeCore.NewProc("HcsCloseComputeSystem")
)

// CreateOperation creates a new operation.
func CreateOperation() (Operation, error) {
	r1, r2, _ := procHcsCreateOperation.Call(0, 0)
	if r2 != 0 {
		return 0, fmt.Errorf("HcsCreateOperation failed: HRESULT 0x%X: %w", r2, Win32FromHResult(r2))
	}

	return Operation(r1), nil
}

func WaitForOperationResult(operation Operation, timeout uint32) (string, error) {
	var resultDocument *uint16

	r1, _, _ := procHcsWaitForOperationResult.Call(uintptr(operation), uintptr(timeout), uintptr(unsafe.Pointer(&resultDocument)))
	if r1 != 0 {
		return "", fmt.Errorf("HcsWaitForOperationResult failed: HRESULT 0x%X: %w", r1, Win32FromHResult(r1))
	}

	result := windows.UTF16PtrToString(resultDocument)
	windows.CoTaskMemFree(unsafe.Pointer(resultDocument))

	return result, nil
}

// CloseOperation closes an operation.
func CloseOperation(operation Operation) {
	_, _, _ = procHcsCloseOperation.Call(uintptr(operation))
}

// EnumerateComputeSystems enumerates compute systems.
//
// https://learn.microsoft.com/en-us/virtualization/api/hcs/reference/hcsenumeratecomputesystems
func EnumerateComputeSystems(query *uint16, operation Operation) error {
	r1, _, _ := procHcsEnumerateComputeSystems.Call(uintptr(unsafe.Pointer(query)), uintptr(operation))
	if r1 != 0 {
		return fmt.Errorf("HcsEnumerateComputeSystems failed: HRESULT 0x%X: %w", r1, Win32FromHResult(r1))
	}

	return nil
}

// OpenComputeSystem opens a handle to an existing compute system.
//
// https://learn.microsoft.com/en-us/virtualization/api/hcs/reference/hcsopencomputesystem
func OpenComputeSystem(id string) (ComputeSystem, error) {
	idPtr, err := windows.UTF16PtrFromString(id)
	if err != nil {
		return 0, err
	}

	var system ComputeSystem

	r1, _, _ := procHcsOpenComputeSystem.Call(
		uintptr(unsafe.Pointer(idPtr)),
		uintptr(windows.GENERIC_ALL),
		uintptr(unsafe.Pointer(&system)),
	)
	if r1 != 0 {
		return 0, fmt.Errorf("HcsOpenComputeSystem failed: HRESULT 0x%X: %w", r1, Win32FromHResult(r1))
	}

	return system, nil
}

func GetComputeSystemProperties(system ComputeSystem, operation Operation, propertyQuery *uint16) error {
	r1, _, err := procHcsGetComputeSystemProperties.Call(
		uintptr(system),
		uintptr(operation),
		uintptr(unsafe.Pointer(propertyQuery)),
	)
	if r1 != 0 {
		return fmt.Errorf("HcsGetComputeSystemProperties failed: HRESULT 0x%X: %w", r1, err)
	}

	return nil
}

// CloseComputeSystem closes a handle to a compute system.
//
// https://learn.microsoft.com/en-us/virtualization/api/hcs/reference/hcsclosecomputesystem
func CloseComputeSystem(system ComputeSystem) {
	_, _, _ = procHcsCloseComputeSystem.Call(uintptr(system))
}

func Win32FromHResult(hr uintptr) windows.Errno {
	if hr&0x1fff0000 == 0x00070000 {
		return windows.Errno(hr & 0xffff)
	}

	return windows.Errno(hr)
}

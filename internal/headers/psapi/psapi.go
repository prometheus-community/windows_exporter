// Copyright 2024 The Prometheus Authors
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

package psapi

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// PerformanceInformation is a wrapper of the PERFORMANCE_INFORMATION struct.
// https://docs.microsoft.com/en-us/windows/win32/api/psapi/ns-psapi-performance_information
type PerformanceInformation struct {
	cb                uint32
	CommitTotal       uint
	CommitLimit       uint
	CommitPeak        uint
	PhysicalTotal     uint
	PhysicalAvailable uint
	SystemCache       uint
	KernelTotal       uint
	KernelPaged       uint
	KernelNonpaged    uint
	PageSize          uint
	HandleCount       uint32
	ProcessCount      uint32
	ThreadCount       uint32
}

//nolint:gochecknoglobals
var (
	psapi                  = windows.NewLazySystemDLL("psapi.dll")
	procGetPerformanceInfo = psapi.NewProc("GetPerformanceInfo")
)

// GetPerformanceInfo returns the dereferenced version of GetLPPerformanceInfo.
func GetPerformanceInfo() (PerformanceInformation, error) {
	var lppi PerformanceInformation
	size := (uint32)(unsafe.Sizeof(lppi))
	lppi.cb = size
	r1, _, err := procGetPerformanceInfo.Call(uintptr(unsafe.Pointer(&lppi)), uintptr(size))

	if ret := *(*bool)(unsafe.Pointer(&r1)); !ret {
		return PerformanceInformation{}, err
	}

	return lppi, nil
}

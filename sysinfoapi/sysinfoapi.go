// Copyright 2020 Prometheus Team
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

// Package sysinfoapi wraps some WinAPI sysinfoapi functions used by the exporter to produce metrics.
// It is fairly opinionated for the exporter's internal use, and is not intended to be a
// generic wrapper for the WinAPI's sysinfoapi.
package sysinfoapi

import (
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

// WinProcInfo is a wrapper for
type WinProcInfo struct {
	WReserved              uint16
	WProcessorArchitecture uint16
}

// WinSystemInfo is a wrapper for LPSYSTEM_INFO
type WinSystemInfo struct {
	Arch                        WinProcInfo
	DwPageSize                  uint32
	LpMinimumApplicationAddress *byte
	LpMaximumApplicationAddress *byte
	DwActiveProcessorMask       *uint32
	DwNumberOfProcessors        uint32
	DwProcessorType             uint32
	DwAllocationGranularity     uint32
	WProcessorLevel             uint16
	WProcessorRevision          uint16
}

// WinComputerNameFormat is a wrapper for COMPUTER_NAME_FORMAT
type WinComputerNameFormat int

// Definitions for WinComputerNameFormat constants
const (
	ComputerNameNetBIOS WinComputerNameFormat = iota
	ComputerNameDNSHostname
	ComputerNameDNSDomain
	ComputerNameDNSFullyQualified
	ComputerNamePhysicalNetBIOS
	ComputerNamePhysicalDNSHostname
	ComputerNamePhysicalDNSDomain
	ComputerNamePhysicalDNSFullyQualified
	ComputerNameMax
)

// WinMemoryStatus is a wrapper for LPMEMORYSTATUSEX
type WinMemoryStatus struct {
	dwLength                uint32
	dwMemoryLoad            uint32
	ullTotalPhys            uint64
	ullAvailPhys            uint64
	ullTotalPageFile        uint64
	ullAvailPageFile        uint64
	ullTotalVirtual         uint64
	ullAvailVirtual         uint64
	ullAvailExtendedVirtual uint64
}

var (
	kernel32                 = windows.NewLazySystemDLL("kernel32.dll")
	procGetSystemInfo        = kernel32.NewProc("GetSystemInfo")
	procGlobalMemoryStatusEx = kernel32.NewProc("GlobalMemoryStatusEx")
	procGetComputerNameExW   = kernel32.NewProc("GetComputerNameExW")
)

// GetNumLogicalProcessors returns the number of logical processes provided by sysinfoapi's GetSystemInfo function.
func GetNumLogicalProcessors() int {
	var sysInfo WinSystemInfo
	pInfo := uintptr(unsafe.Pointer(&sysInfo))
	procGetSystemInfo.Call(pInfo)
	return int(sysInfo.DwNumberOfProcessors)
}

// GetPhysicalMemory returns the system's installed physical memory provided by sysinfoapi's GlobalMemoryStatusEx function.
func GetPhysicalMemory() (int, error) {
	var wm WinMemoryStatus
	wm.dwLength = (uint32)(unsafe.Sizeof(wm))
	r1, _, err := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&wm)))
	if r1 != 1 {
		return 0, err
	}
	return int(wm.ullTotalPhys), nil
}

// GetComputerName provides the requested computer name provided by sysinfoapi's GetComputerNameEx function.
func GetComputerName(f WinComputerNameFormat) (string, error) {
	size := 4096
	var buffer [4096]uint16
	r1, _, err := procGetComputerNameExW.Call(uintptr(f), uintptr(unsafe.Pointer(&buffer)), uintptr(unsafe.Pointer(&size)))
	if r1 == 0 {
		return "", err
	}
	bytes := buffer[0:size]
	out := utf16.Decode(bytes)
	return string(out), nil
}

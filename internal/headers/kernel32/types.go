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

package kernel32

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

type JobObjectBasicAccountingInformation struct {
	TotalUserTime             uint64
	TotalKernelTime           uint64
	ThisPeriodTotalUserTime   uint64
	ThisPeriodTotalKernelTime uint64
	TotalPageFaultCount       uint32
	TotalProcesses            uint32
	ActiveProcesses           uint32
	TotalTerminatedProcesses  uint32
}

// JobObjectBasicAndIOAccountingInformation is a structure that contains
// both basic accounting information and I/O accounting information
// for a job object. It is used with the QueryInformationJobObject function.
// The structure is defined in the Windows API documentation.
// https://learn.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-jobobject_basic_and_io_accounting_information
type JobObjectBasicAndIOAccountingInformation struct {
	BasicInfo JobObjectBasicAccountingInformation
	IoInfo    windows.IO_COUNTERS
}
type JobObjectMemoryUsageInformation struct {
	JobMemory         uint64
	PeakJobMemoryUsed uint64
}

type JobObjectBasicProcessIDList struct {
	NumberOfAssignedProcesses uint32
	NumberOfProcessIdsInList  uint32
	ProcessIdList             [1]uintptr
}

// PIDs returns all the process Ids in the job object.
func (p *JobObjectBasicProcessIDList) PIDs() []uint32 {
	return unsafe.Slice((*uint32)(unsafe.Pointer(&p.ProcessIdList[0])), int(p.NumberOfProcessIdsInList))
}

type PROCESS_VM_COUNTERS struct {
	PeakVirtualSize            uintptr
	VirtualSize                uintptr
	PageFaultCount             uint32
	PeakWorkingSetSize         uintptr
	WorkingSetSize             uintptr
	QuotaPeakPagedPoolUsage    uintptr
	QuotaPagedPoolUsage        uintptr
	QuotaPeakNonPagedPoolUsage uintptr
	QuotaNonPagedPoolUsage     uintptr
	PagefileUsage              uintptr
	PeakPagefileUsage          uintptr
	PrivateWorkingSetSize      uintptr
}

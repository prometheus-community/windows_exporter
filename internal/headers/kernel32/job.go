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
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	// JobObjectQuery is required to retrieve certain information about a job object,
	// such as attributes and accounting information (see QueryInformationJobObject and IsProcessInJob).
	// https://learn.microsoft.com/en-us/windows/win32/procthread/job-object-security-and-access-rights
	JobObjectQuery = 0x0004
)

func OpenJobObject(name string) (windows.Handle, error) {
	ptr, _ := windows.UTF16PtrFromString(name)
	handle, _, err := procOpenJobObject.Call(
		JobObjectQuery,
		0,
		uintptr(unsafe.Pointer(ptr)),
	)

	if handle == 0 {
		return 0, err
	}

	return windows.Handle(handle), nil
}

func IsProcessInJob(process windows.Handle, job windows.Handle) (bool, error) {
	// IsProcessInJob writes a 4-byte BOOL.
	var result int32

	ret, _, err := procIsProcessInJob.Call(
		uintptr(process),
		uintptr(job),
		uintptr(unsafe.Pointer(&result)),
	)
	if ret == 0 {
		return false, err
	}

	return result != 0, nil
}

// QueryJobObjectProcessIDs returns the IDs of all processes assigned to the job.
func QueryJobObjectProcessIDs(job windows.Handle) ([]uint32, error) {
	const headerSize = unsafe.Offsetof(JobObjectBasicProcessIDList{}.ProcessIdList)

	headerEntries := int((headerSize + unsafe.Sizeof(uintptr(0)) - 1) / unsafe.Sizeof(uintptr(0)))
	capacity := 16

	for {
		// Back the list with []uintptr, so the ULONG_PTR entries are aligned.
		buf := make([]uintptr, headerEntries+capacity)
		list := (*JobObjectBasicProcessIDList)(unsafe.Pointer(&buf[0]))

		err := windows.QueryInformationJobObject(
			job,
			windows.JobObjectBasicProcessIdList,
			uintptr(unsafe.Pointer(&buf[0])),
			uint32(uintptr(len(buf))*unsafe.Sizeof(buf[0])),
			nil,
		)

		if errors.Is(err, windows.ERROR_MORE_DATA) {
			// Processes can join the job between two calls, so leave some headroom.
			capacity = max(capacity*2, int(list.NumberOfAssignedProcesses)+16)

			continue
		}

		if err != nil {
			return nil, err
		}

		ids := unsafe.Slice(&list.ProcessIdList[0], int(list.NumberOfProcessIdsInList))
		pids := make([]uint32, len(ids))

		for i, id := range ids {
			pids[i] = uint32(id)
		}

		return pids, nil
	}
}

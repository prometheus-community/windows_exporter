// Copyright 2025 The Prometheus Authors
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

package kernel32

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	// JOB_OBJECT_QUERY is required to retrieve certain information about a job object,
	// such as attributes and accounting information (see QueryInformationJobObject and IsProcessInJob).
	// https://learn.microsoft.com/en-us/windows/win32/procthread/job-object-security-and-access-rights
	JOB_OBJECT_QUERY = 0x0004

	ClassJobObjectBasicProcessIdList uint16 = 3
	// ClassJobObjectExtendedLimitInformation
	// https://learn.microsoft.com/en-us/windows/win32/api/jobapi2/nf-jobapi2-queryinformationjobobject
	ClassJobObjectExtendedLimitInformation uint16 = 9
)

func OpenJobObject(name string) (windows.Handle, error) {
	handle, _, err := procOpenJobObject.Call(JOB_OBJECT_QUERY, 0, uintptr(unsafe.Pointer(&name)))
	if handle == 0 {
		return 0, err
	}

	return windows.Handle(handle), nil
}

func IsProcessInJob(process windows.Handle, job windows.Handle, result *bool) error {
	ret, _, err := procIsProcessInJob.Call(
		uintptr(process),
		uintptr(job),
		uintptr(unsafe.Pointer(&result)),
	)
	if ret == 0 {
		return err
	}

	return nil
}

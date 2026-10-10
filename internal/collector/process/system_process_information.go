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

package process

import (
	"errors"
	"fmt"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// maxSnapshotAttempts bounds the retries when processes are created
// between the size query and the snapshot.
const maxSnapshotAttempts = 5

// processSnapshot is one process from SystemProcessInformation.
// The values are the same that the perflib Process counter set reads.
type processSnapshot struct {
	name string

	pid       uint32
	parentPID uint32

	// createTime is the FILETIME of the process creation, in 100ns intervals since 1601.
	createTime int64
	// userTime and kernelTime are in 100ns intervals.
	userTime   int64
	kernelTime int64

	basePriority int32
	handleCount  uint32
	threadCount  uint32

	pageFaultCount    uint32
	virtualSize       uint64
	workingSetSize    uint64
	workingSetPeak    uint64
	workingSetPrivate uint64
	pagedPoolUsage    uint64
	nonPagedPoolUsage uint64
	pageFileUsage     uint64
	privatePageCount  uint64

	readOperationCount  uint64
	writeOperationCount uint64
	otherOperationCount uint64
	readTransferCount   uint64
	writeTransferCount  uint64
	otherTransferCount  uint64
}

// snapshotBuffer holds the SystemProcessInformation buffer across scrapes.
// The backing array is []uint64, so the process entries are 8-byte aligned.
type snapshotBuffer struct {
	buf []uint64
}

// query returns all processes of the system with one NtQuerySystemInformation call.
func (s *snapshotBuffer) query(dst []processSnapshot) ([]processSnapshot, error) {
	if len(s.buf) == 0 {
		s.buf = make([]uint64, 256<<10/8)
	}

	var returnLength uint32

	for attempt := 0; ; attempt++ {
		err := windows.NtQuerySystemInformation(
			windows.SystemProcessInformation,
			unsafe.Pointer(&s.buf[0]),
			uint32(len(s.buf)*8),
			&returnLength,
		)
		if err == nil {
			break
		}

		if !errors.Is(err, windows.STATUS_INFO_LENGTH_MISMATCH) || attempt >= maxSnapshotAttempts {
			return dst, fmt.Errorf("NtQuerySystemInformation(SystemProcessInformation): %w", err)
		}

		// Processes can start before the next call. Leave room for them.
		size := max(uint64(returnLength), uint64(len(s.buf))*8)
		s.buf = make([]uint64, (size+size/4)/8+1)
	}

	return parseSystemProcessInformation(dst[:0], unsafe.Slice((*byte)(unsafe.Pointer(&s.buf[0])), len(s.buf)*8), returnLength)
}

// parseSystemProcessInformation decodes the entries of a SystemProcessInformation buffer.
// The buffer must be 8-byte aligned. The image names are copied, so dst does not reference buf.
func parseSystemProcessInformation(dst []processSnapshot, buf []byte, length uint32) ([]processSnapshot, error) {
	end := min(uint64(length), uint64(len(buf)))
	entrySize := uint64(unsafe.Sizeof(windows.SYSTEM_PROCESS_INFORMATION{}))

	for offset := uint64(0); ; {
		if offset+entrySize > end {
			return dst, fmt.Errorf("SystemProcessInformation entry at offset %d exceeds the buffer length %d", offset, end)
		}

		entry := (*windows.SYSTEM_PROCESS_INFORMATION)(unsafe.Pointer(&buf[offset]))

		var imageName string

		if entry.ImageName.Length > 0 && entry.ImageName.Buffer != nil {
			imageName = entry.ImageName.String()
		}

		dst = append(dst, processSnapshot{
			name:                processName(uint32(entry.UniqueProcessID), imageName),
			pid:                 uint32(entry.UniqueProcessID),
			parentPID:           uint32(entry.InheritedFromUniqueProcessID),
			createTime:          entry.CreateTime,
			userTime:            entry.UserTime,
			kernelTime:          entry.KernelTime,
			basePriority:        entry.BasePriority,
			handleCount:         entry.HandleCount,
			threadCount:         entry.NumberOfThreads,
			pageFaultCount:      entry.PageFaultCount,
			virtualSize:         uint64(entry.VirtualSize),
			workingSetSize:      uint64(entry.WorkingSetSize),
			workingSetPeak:      uint64(entry.PeakWorkingSetSize),
			workingSetPrivate:   uint64(entry.WorkingSetPrivateSize),
			pagedPoolUsage:      uint64(entry.QuotaPagedPoolUsage),
			nonPagedPoolUsage:   uint64(entry.QuotaNonPagedPoolUsage),
			pageFileUsage:       uint64(entry.PagefileUsage),
			privatePageCount:    uint64(entry.PrivatePageCount),
			readOperationCount:  uint64(entry.ReadOperationCount),
			writeOperationCount: uint64(entry.WriteOperationCount),
			otherOperationCount: uint64(entry.OtherOperationCount),
			readTransferCount:   uint64(entry.ReadTransferCount),
			writeTransferCount:  uint64(entry.WriteTransferCount),
			otherTransferCount:  uint64(entry.OtherTransferCount),
		})

		if entry.NextEntryOffset == 0 {
			return dst, nil
		}

		offset += uint64(entry.NextEntryOffset)
	}
}

// processName returns the instance name that the perflib Process counter set uses:
// the image name without a case-insensitive ".exe" suffix, and "Idle" for the System Idle Process.
// Other extensions, like ".scr", are kept.
func processName(pid uint32, imageName string) string {
	if pid == 0 && imageName == "" {
		return "Idle"
	}

	if len(imageName) > len(".exe") && strings.EqualFold(imageName[len(imageName)-len(".exe"):], ".exe") {
		return imageName[:len(imageName)-len(".exe")]
	}

	return imageName
}

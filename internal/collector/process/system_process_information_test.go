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
	"log/slog"
	"os"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func TestProcessName(t *testing.T) {
	t.Parallel()

	// The names follow the instances of the perflib Process counter set.
	for _, tc := range []struct {
		pid       uint32
		imageName string
		want      string
	}{
		{0, "", "Idle"},
		{4, "System", "System"},
		{100, "Registry", "Registry"},
		{200, "svchost.exe", "svchost"},
		{300, "EPSDNMON.EXE", "EPSDNMON"},
		{400, "Microsoft.Photos.exe", "Microsoft.Photos"},
		{500, "screensaver.scr", "screensaver.scr"},
		{600, ".exe", ".exe"},
	} {
		require.Equal(t, tc.want, processName(tc.pid, tc.imageName), "pid %d image %q", tc.pid, tc.imageName)
	}
}

func TestSnapshotContainsCurrentProcess(t *testing.T) {
	t.Parallel()

	var s snapshotBuffer

	processes, err := s.query(nil)
	require.NoError(t, err)
	require.NotEmpty(t, processes)

	self := findProcess(t, processes, uint32(os.Getpid()))
	require.Equal(t, uint32(os.Getppid()), self.parentPID)
	require.Equal(t, currentProcessCreateTime(t), self.createTime)
	require.NotZero(t, self.threadCount)
	require.NotZero(t, self.handleCount)
	require.NotZero(t, self.workingSetSize)
	require.NotZero(t, self.privatePageCount)

	idle := findProcess(t, processes, 0)
	require.Equal(t, "Idle", idle.name)
}

func TestSnapshotGrowsBuffer(t *testing.T) {
	t.Parallel()

	s := snapshotBuffer{buf: make([]uint64, 1)}

	processes, err := s.query(nil)
	require.NoError(t, err)
	require.NotEmpty(t, processes)
	require.Greater(t, len(s.buf), 1)
}

func TestParseSystemProcessInformationRejectsTruncatedBuffer(t *testing.T) {
	t.Parallel()

	entrySize := int(unsafe.Sizeof(windows.SYSTEM_PROCESS_INFORMATION{}))
	buf := make([]uint64, entrySize/8+1)
	data := unsafe.Slice((*byte)(unsafe.Pointer(&buf[0])), len(buf)*8)

	// The first entry points to a second entry behind the end of the buffer.
	(*windows.SYSTEM_PROCESS_INFORMATION)(unsafe.Pointer(&data[0])).NextEntryOffset = uint32(entrySize)

	processes, err := parseSystemProcessInformation(nil, data, uint32(len(data)))
	require.Error(t, err)
	require.Len(t, processes, 1)

	_, err = parseSystemProcessInformation(nil, data, uint32(entrySize-1))
	require.Error(t, err)
}

func TestParseSystemProcessInformationRejectsImageNameOutsideBuffer(t *testing.T) {
	t.Parallel()

	entrySize := int(unsafe.Sizeof(windows.SYSTEM_PROCESS_INFORMATION{}))
	buf := make([]uint64, entrySize/8+1)
	data := unsafe.Slice((*byte)(unsafe.Pointer(&buf[0])), len(buf)*8)
	other := []uint16{'a', 'b'}

	entry := (*windows.SYSTEM_PROCESS_INFORMATION)(unsafe.Pointer(&data[0]))
	entry.ImageName.Buffer = &other[0]
	entry.ImageName.Length = 4

	_, err := parseSystemProcessInformation(nil, data, uint32(len(data)))
	require.Error(t, err)

	// A name inside the buffer, but longer than the rest of it.
	entry.ImageName.Buffer = (*uint16)(unsafe.Pointer(&data[len(data)-2]))
	_, err = parseSystemProcessInformation(nil, data, uint32(len(data)))
	require.Error(t, err)
}

func TestResolveProcessInfoCache(t *testing.T) {
	t.Parallel()

	c := New(nil)
	c.logger = slog.New(slog.DiscardHandler)

	var s snapshotBuffer

	processes, err := s.query(nil)
	require.NoError(t, err)

	self := findProcess(t, processes, uint32(os.Getpid()))
	exited := processSnapshot{pid: 1<<32 - 4, createTime: 1}

	infos := c.resolveProcessInfo([]*processSnapshot{&self, &exited})
	require.NotEmpty(t, infos[0].owner)
	require.NotEmpty(t, infos[0].cmdLine)
	require.Contains(t, c.infoCache, self.pid)
	require.NotContains(t, c.infoCache, exited.pid, "failed lookups must be retried")

	// A cached entry with the same creation time is reused without opening the process.
	c.infoCache[self.pid] = processInfo{createTime: self.createTime, owner: "cached"}
	infos = c.resolveProcessInfo([]*processSnapshot{&self})
	require.Equal(t, "cached", infos[0].owner)

	// A reused PID has another creation time and is looked up again.
	c.infoCache[self.pid] = processInfo{createTime: self.createTime - 1, owner: "cached"}
	infos = c.resolveProcessInfo([]*processSnapshot{&self})
	require.NotEqual(t, "cached", infos[0].owner)

	// Processes that are gone are evicted.
	c.resolveProcessInfo(nil)
	require.Empty(t, c.infoCache)
}

func TestGetProcessInformationDetectsReusedPID(t *testing.T) {
	t.Parallel()

	c := New(nil)
	c.logger = slog.New(slog.DiscardHandler)

	_, _, _, err := c.getProcessInformation(uint32(os.Getpid()), currentProcessCreateTime(t)+1)
	require.ErrorIs(t, err, errProcessGone)

	cmdLine, owner, _, err := c.getProcessInformation(uint32(os.Getpid()), currentProcessCreateTime(t))
	require.NoError(t, err)
	require.NotEmpty(t, cmdLine)
	require.NotEmpty(t, owner)
}

func findProcess(t *testing.T, processes []processSnapshot, pid uint32) processSnapshot {
	t.Helper()

	for _, process := range processes {
		if process.pid == pid {
			return process
		}
	}

	require.FailNow(t, "process not found in snapshot", "pid %d", pid)

	return processSnapshot{}
}

func currentProcessCreateTime(t *testing.T) int64 {
	t.Helper()

	var creationTime, exitTime, kernelTime, userTime windows.Filetime

	require.NoError(t, windows.GetProcessTimes(windows.CurrentProcess(), &creationTime, &exitTime, &kernelTime, &userTime))

	return int64(creationTime.HighDateTime)<<32 | int64(creationTime.LowDateTime)
}

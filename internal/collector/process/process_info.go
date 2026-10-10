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
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/windows"
)

// maxInfoWorkers bounds the goroutines that open processes for new cache entries.
const maxInfoWorkers = 8

// errProcessGone reports that the PID no longer belongs to the snapshot process.
var errProcessGone = errors.New("process exited")

// processInfo holds the values of windows_process_info that are fixed for the lifetime of a process.
type processInfo struct {
	// createTime identifies the process together with the PID, because Windows reuses PIDs.
	createTime int64

	cmdLine        string
	owner          string
	processGroupID uint32
}

// resolveProcessInfo returns the info of each process, in the order of processes.
// Cached entries are reused while the PID and creation time match. The others are
// looked up in parallel. The cache keeps only the processes passed in.
func (c *Collector) resolveProcessInfo(processes []*processSnapshot) []processInfo {
	infos := make([]processInfo, len(processes))
	cached := make([]bool, len(processes))
	misses := make([]int, 0)

	for i, process := range processes {
		if info, ok := c.infoCache[process.pid]; ok && info.createTime == process.createTime {
			infos[i] = info
			cached[i] = true

			continue
		}

		misses = append(misses, i)
	}

	if len(misses) > 0 {
		var next atomic.Int64

		wg := sync.WaitGroup{}

		for range min(len(misses), runtime.GOMAXPROCS(0), maxInfoWorkers) {
			wg.Go(func() {
				for {
					n := int(next.Add(1)) - 1
					if n >= len(misses) {
						return
					}

					i := misses[n]
					c.lookupProcessInfo(processes[i], &infos[i], &cached[i])
				}
			})
		}

		wg.Wait()
	}

	cache := make(map[uint32]processInfo, len(processes))

	for i, process := range processes {
		if cached[i] {
			cache[process.pid] = infos[i]
		}
	}

	c.infoCache = cache

	return infos
}

// lookupProcessInfo opens the process and reads its owner, command line and process group ID.
// cacheable is set only if the lookup succeeded. A panic leaves both values unset.
func (c *Collector) lookupProcessInfo(process *processSnapshot, info *processInfo, cacheable *bool) {
	defer func() {
		if r := recover(); r != nil {
			c.logger.LogAttrs(context.Background(), slog.LevelError, "Panic while reading process information",
				slog.Uint64("pid", uint64(process.pid)),
				slog.Any("panic", r),
				slog.String("stack", string(debug.Stack())),
			)
		}
	}()

	cmdLine, owner, processGroupID, err := c.getProcessInformation(process.pid, process.createTime)
	if err != nil {
		if !errors.Is(err, errProcessGone) {
			c.logger.LogAttrs(context.Background(), slog.LevelDebug, "Failed to get process information",
				slog.Uint64("pid", uint64(process.pid)),
				slog.String("process", process.name),
				slog.Any("err", err),
			)
		}
	}

	*info = processInfo{
		createTime:     process.createTime,
		cmdLine:        cmdLine,
		owner:          owner,
		processGroupID: processGroupID,
	}
	*cacheable = err == nil
}

// getProcessInformation returns the command line, owner and process group ID of a process.
// createTime is checked after opening the process, so a reused PID is not attributed to the old process.
//
// ref: https://github.com/microsoft/hcsshim/blob/8beabacfc2d21767a07c20f8dd5f9f3932dbf305/internal/uvm/stats.go#L25
func (c *Collector) getProcessInformation(pid uint32, createTime int64) (string, string, uint32, error) {
	if pid == 0 {
		return "", "", 0, nil
	}

	hProcess, vmReadAccess, err := c.openProcess(pid)
	if err != nil {
		if errors.Is(err, windows.ERROR_ACCESS_DENIED) {
			return "", "", 0, nil
		}

		return "", "", 0, err
	}

	defer func(hProcess windows.Handle) {
		if err := windows.CloseHandle(hProcess); err != nil {
			c.logger.Warn("CloseHandle failed",
				slog.Any("err", err),
			)
		}
	}(hProcess)

	var creationTime, exitTime, kernelTime, userTime windows.Filetime

	if err := windows.GetProcessTimes(hProcess, &creationTime, &exitTime, &kernelTime, &userTime); err == nil &&
		createTime != 0 && int64(creationTime.HighDateTime)<<32|int64(creationTime.LowDateTime) != createTime {
		return "", "", 0, errProcessGone
	}

	owner, err := c.getProcessOwner(c.logger, hProcess)
	if err != nil {
		return "", "", 0, err
	}

	var (
		cmdLine        string
		processGroupID uint32
	)

	if vmReadAccess {
		cmdLine, processGroupID, err = c.getExtendedProcessInformation(hProcess)
		if err != nil {
			return "", owner, processGroupID, err
		}
	}

	return cmdLine, owner, processGroupID, nil
}

func (c *Collector) getExtendedProcessInformation(hProcess windows.Handle) (string, uint32, error) {
	// Get the process environment block (PEB) address
	var pbi windows.PROCESS_BASIC_INFORMATION

	retLen := uint32(unsafe.Sizeof(pbi))
	if err := windows.NtQueryInformationProcess(hProcess, windows.ProcessBasicInformation, unsafe.Pointer(&pbi), retLen, &retLen); err != nil {
		return "", 0, fmt.Errorf("failed to query process basic information: %w", err)
	}

	// Minimal processes like Registry and Memory Compression have no PEB.
	if pbi.PebBaseAddress == nil {
		return "", 0, nil
	}

	peb := windows.PEB{}

	err := windows.ReadProcessMemory(hProcess,
		uintptr(unsafe.Pointer(pbi.PebBaseAddress)),
		(*byte)(unsafe.Pointer(&peb)),
		unsafe.Sizeof(peb),
		nil,
	)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read process memory: %w", err)
	}

	if peb.ProcessParameters == nil {
		return "", 0, nil
	}

	processParameters := windows.RTL_USER_PROCESS_PARAMETERS{}

	err = windows.ReadProcessMemory(hProcess,
		uintptr(unsafe.Pointer(peb.ProcessParameters)),
		(*byte)(unsafe.Pointer(&processParameters)),
		unsafe.Sizeof(processParameters),
		nil,
	)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read process memory: %w", err)
	}

	var cmdLine string

	// CommandLine.Length is in bytes.
	if c.config.EnableCMDLine && processParameters.CommandLine.Length >= 2 && processParameters.CommandLine.Buffer != nil {
		cmdLineUTF16 := make([]uint16, processParameters.CommandLine.Length/2)

		err = windows.ReadProcessMemory(hProcess,
			uintptr(unsafe.Pointer(processParameters.CommandLine.Buffer)),
			(*byte)(unsafe.Pointer(&cmdLineUTF16[0])),
			uintptr(len(cmdLineUTF16)*2),
			nil,
		)
		if err != nil {
			return "", processParameters.ProcessGroupId, fmt.Errorf("failed to read process memory: %w", err)
		}

		cmdLine = strings.TrimSpace(windows.UTF16ToString(cmdLineUTF16))
	}

	return cmdLine, processParameters.ProcessGroupId, nil
}

func (c *Collector) getProcessOwner(logger *slog.Logger, hProcess windows.Handle) (string, error) {
	var tok windows.Token

	if err := windows.OpenProcessToken(hProcess, windows.TOKEN_QUERY, &tok); err != nil {
		if errors.Is(err, windows.ERROR_ACCESS_DENIED) {
			return "", nil
		}

		return "", fmt.Errorf("failed to open process token: %w", err)
	}

	defer func(tok windows.Token) {
		if err := tok.Close(); err != nil {
			logger.Warn("Token close failed",
				slog.Any("err", err),
			)
		}
	}(tok)

	tokenUser, err := tok.GetTokenUser()
	if err != nil {
		return "", fmt.Errorf("failed to get token user: %w", err)
	}

	sid := tokenUser.User.Sid.String()

	var owner string

	ownerVal, ok := c.lookupCache.Load(sid)

	if ok {
		owner, ok = ownerVal.(string)
	}

	if !ok {
		account, domain, _, err := tokenUser.User.Sid.LookupAccount("")
		if err != nil {
			owner = sid
		} else {
			owner = fmt.Sprintf(`%s\%s`, domain, account)
		}

		c.lookupCache.Store(sid, owner)
	}

	return owner, nil
}

func (c *Collector) openProcess(pid uint32) (windows.Handle, bool, error) {
	// Open the process with QUERY_INFORMATION and VM_READ permissions.
	hProcess, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_VM_READ, false, pid)
	if err == nil {
		return hProcess, true, nil
	}

	if errors.Is(err, windows.ERROR_INVALID_PARAMETER) { // The PID does not exist anymore.
		return 0, false, errProcessGone
	}

	if !errors.Is(err, windows.ERROR_ACCESS_DENIED) {
		return 0, false, fmt.Errorf("failed to open process: %w", err)
	}

	hProcess, err = windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		if errors.Is(err, windows.ERROR_INVALID_PARAMETER) {
			return 0, false, errProcessGone
		}

		return 0, false, fmt.Errorf("failed to open process with limited permissions: %w", err)
	}

	return hProcess, false, nil
}

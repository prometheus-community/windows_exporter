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

// Package w32time reads the status of the local Windows Time service.
//
// w32time.dll exports the client side of the W32Time Remote Protocol
// ([MS-W32T]). The functions are not part of the Win32 API documentation; the
// structures follow [MS-W32T] 2.2.17, and their field meanings 3.2.1.1.
//
// [MS-W32T]: https://winprotocoldocs-bhdugrdyduf5h2e4.b02.azurefd.net/MS-W32T/%5bMS-W32T%5d.pdf
package w32time

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"github.com/prometheus-community/windows_exporter/internal/headers/lazyproc"
	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var (
	modw32time = windows.NewLazySystemDLL("w32time.dll")

	procW32TimeQueryStatus = modw32time.NewProc("W32TimeQueryStatus")
	procW32TimeBufferFree  = modw32time.NewProc("W32TimeBufferFree")

	load = lazyproc.Once(modw32time,
		procW32TimeQueryStatus,
		procW32TimeBufferFree,
	)
)

// Load checks that w32time.dll exports every function this package calls.
// Call it in Build: a missing export panics on the first call otherwise.
func Load() error { return load() }

// statusInfo is W32TIME_STATUS_INFO ([MS-W32T] 2.2.17). The layout matches
// the C structure on 64-bit Windows only; QueryStatus checks ulSize.
type statusInfo struct {
	ulSize                uint32
	eLeapIndicator        uint32
	nStratum              uint32
	nPollInterval         int32
	refidSource           uint32
	qwLastSyncTicks       uint64 // FILETIME of the last successful synchronization.
	toRootDelay           int64  // 100 ns units.
	tpRootDispersion      uint64 // 100 ns units.
	nClockPrecision       int32
	wszSource             *uint16
	toSysPhaseOffset      int64 // 100 ns units.
	ulLcState             uint32
	ulTSFlags             uint32
	ulClockRate           uint32
	ulNetlogonServiceBits uint32
	eLastSyncResult       uint32
	tpTimeLastGoodSync    uint64 // Time since the last good synchronization, 100 ns units.
	cEntries              uint32
	pEntries              uintptr
}

// State is the state of the clock discipline algorithm ([MS-W32T] 2.2.7).
type State uint32

const (
	StateUnset State = iota
	StateHold
	StateSync
	StateSpike
)

// SyncResult is a TimeSync_ReturnResult code ([MS-W32T] 3.2.5.1).
type SyncResult uint32

const (
	SyncResultSuccess SyncResult = iota
	SyncResultNoData
	SyncResultStaleData
	SyncResultChangeTooBig
	SyncResultShutdown
)

// Status is the status of the local Windows Time service, as reported by
// w32tm /query /status.
type Status struct {
	LeapIndicator uint32
	Stratum       uint32
	// PollInterval is the time between two synchronizations.
	PollInterval   time.Duration
	RootDelay      time.Duration
	RootDispersion time.Duration
	Precision      time.Duration
	Source         string
	PhaseOffset    time.Duration
	State          State
	LastSyncResult SyncResult
	// LastSuccessfulSync is zero when the service has not synchronized yet.
	LastSuccessfulSync time.Time
}

// QueryStatus returns the status of the local Windows Time service. It needs
// no administrative rights. It fails when the service is not running.
func QueryStatus() (Status, error) {
	var info *statusInfo

	r0, _, _ := syscall.SyscallN(procW32TimeQueryStatus.Addr(), 0, uintptr(unsafe.Pointer(&info)))
	if r0 != 0 {
		return Status{}, fmt.Errorf("W32TimeQueryStatus: %w", windows.Errno(r0))
	}

	if info == nil {
		return Status{}, fmt.Errorf("W32TimeQueryStatus: %w", windows.ERROR_INVALID_DATA)
	}

	// The strings are part of the same buffer, so it is freed once.
	defer syscall.SyscallN(procW32TimeBufferFree.Addr(), uintptr(unsafe.Pointer(info))) //nolint:errcheck

	return parseStatus(info)
}

func parseStatus(info *statusInfo) (Status, error) {
	if uintptr(info.ulSize) < unsafe.Sizeof(*info) {
		return Status{}, fmt.Errorf("W32TimeQueryStatus: unexpected W32TIME_STATUS_INFO size %d, want %d", info.ulSize, unsafe.Sizeof(*info))
	}

	status := Status{
		LeapIndicator:  info.eLeapIndicator,
		Stratum:        info.nStratum,
		PollInterval:   log2Duration(info.nPollInterval),
		RootDelay:      time.Duration(info.toRootDelay) * 100,
		RootDispersion: time.Duration(info.tpRootDispersion) * 100,
		Precision:      log2Duration(info.nClockPrecision),
		PhaseOffset:    time.Duration(info.toSysPhaseOffset) * 100,
		State:          State(info.ulLcState),
		LastSyncResult: SyncResult(info.eLastSyncResult),
	}

	if info.wszSource != nil {
		status.Source = windows.UTF16PtrToString(info.wszSource)
	}

	if info.qwLastSyncTicks != 0 {
		lastSync := windows.Filetime{
			LowDateTime:  uint32(info.qwLastSyncTicks),
			HighDateTime: uint32(info.qwLastSyncTicks >> 32),
		}
		status.LastSuccessfulSync = time.Unix(0, lastSync.Nanoseconds())
	}

	return status, nil
}

// log2Duration converts an NTP interval, a power of two in seconds, such as
// the poll interval or the clock precision.
func log2Duration(exponent int32) time.Duration {
	if exponent < -63 || exponent > 32 {
		return 0
	}

	if exponent >= 0 {
		return time.Duration(1<<exponent) * time.Second
	}

	return time.Duration(float64(time.Second) / float64(uint64(1)<<-exponent))
}

// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package ntdll

import (
	"fmt"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// SystemTimeOfDayInformation is the SYSTEM_TIMEOFDAY_INFORMATION structure.
// The times are in 100-nanosecond intervals since January 1, 1601 (UTC).
type SystemTimeOfDayInformation struct {
	BootTime      int64
	CurrentTime   int64
	TimeZoneBias  int64
	TimeZoneID    uint32
	Reserved      uint32
	BootTimeBias  uint64
	SleepTimeBias uint64
}

// BootTime returns the time the system booted.
//
// The kernel moves the boot time along when the system time is set, so it
// stays correct after a clock correction.
func BootTime() (time.Time, error) {
	var info SystemTimeOfDayInformation

	err := windows.NtQuerySystemInformation(
		windows.SystemTimeOfDayInformation,
		unsafe.Pointer(&info),
		uint32(unsafe.Sizeof(info)),
		nil,
	)
	if err != nil {
		return time.Time{}, fmt.Errorf("NtQuerySystemInformation(SystemTimeOfDayInformation): %w", err)
	}

	ft := windows.Filetime{
		LowDateTime:  uint32(info.BootTime),
		HighDateTime: uint32(info.BootTime >> 32),
	}

	return time.Unix(0, ft.Nanoseconds()), nil
}

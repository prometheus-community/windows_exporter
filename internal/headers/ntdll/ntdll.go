// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package ntdll

import (
	"github.com/prometheus-community/windows_exporter/internal/headers/lazyproc"
	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var (
	modNtdll                  = windows.NewLazySystemDLL("ntdll.dll")
	procRtlNtStatusToDosError = modNtdll.NewProc("RtlNtStatusToDosError")

	load = lazyproc.Once(modNtdll, procRtlNtStatusToDosError)
)

// Load checks that ntdll.dll exports every function of this package. Call it
// in Build: a missing export panics on the first call otherwise.
func Load() error { return load() }

func RtlNtStatusToDosError(status uintptr) error {
	ret, _, _ := procRtlNtStatusToDosError.Call(status)
	if ret == 0 {
		return nil
	}

	return windows.Errno(ret)
}

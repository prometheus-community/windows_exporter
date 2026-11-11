// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package slc

import (
	"errors"
	"unsafe"

	"github.com/prometheus-community/windows_exporter/internal/headers/lazyproc"
	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var (
	slc                         = windows.NewLazySystemDLL("slc.dll")
	procSLIsWindowsGenuineLocal = slc.NewProc("SLIsWindowsGenuineLocal")

	load = lazyproc.Once(slc, procSLIsWindowsGenuineLocal)
)

// Load checks that slc.dll exports every function of this package. Call it in
// Build: a missing export panics on the first call otherwise.
func Load() error { return load() }

// SL_GENUINE_STATE enumeration
//
// https://learn.microsoft.com/en-us/windows/win32/api/slpublic/ne-slpublic-sl_genuine_state
type SL_GENUINE_STATE uint32

const (
	SL_GEN_STATE_IS_GENUINE SL_GENUINE_STATE = iota
	SL_GEN_STATE_INVALID_LICENSE
	SL_GEN_STATE_TAMPERED
	SL_GEN_STATE_OFFLINE
	SL_GEN_STATE_LAST
)

// SLIsWindowsGenuineLocal function wrapper.
func SLIsWindowsGenuineLocal() (SL_GENUINE_STATE, error) {
	var genuineState SL_GENUINE_STATE

	_, _, err := procSLIsWindowsGenuineLocal.Call(
		uintptr(unsafe.Pointer(&genuineState)),
	)
	if !errors.Is(err, windows.NTE_OP_OK) {
		return 0, err
	}

	return genuineState, nil
}

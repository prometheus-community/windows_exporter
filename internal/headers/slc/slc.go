// Copyright 2024 The Prometheus Authors
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

package slc

import (
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var (
	slc                         = windows.NewLazySystemDLL("slc.dll")
	procSLIsWindowsGenuineLocal = slc.NewProc("SLIsWindowsGenuineLocal")
)

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

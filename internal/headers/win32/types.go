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

package win32

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

const MAX_PATH = 260

type (
	BOOL      = int32 // BOOL is a 32-bit signed int in Win32
	DATE_TIME = windows.Filetime
	DWORD     = uint32
	LPWSTR    struct {
		*uint16
	}
	ULONG = uint32 // ULONG is a 32-bit unsigned int in Win32
	UINT  = uint32 // UINT is a 32-bit unsigned int in Win32
)

// NewLPWSTR creates a new LPWSTR from a string.
// If the string is empty, it returns nil.
// This function converts the string to a UTF-16 pointer.
func NewLPWSTR(str string) *LPWSTR {
	if str == "" {
		return nil
	}

	// Convert the string to a UTF-16 pointer
	ptr, _ := windows.UTF16PtrFromString(str)

	return &LPWSTR{ptr}
}

// Pointer returns the uintptr representation of the LPWSTR.
// This is useful for passing the pointer to Windows API functions.
func (s *LPWSTR) Pointer() uintptr {
	return uintptr(unsafe.Pointer(s.uint16))
}

// String converts the LPWSTR back to a string.
func (s *LPWSTR) String() string {
	return windows.UTF16PtrToString(s.uint16)
}

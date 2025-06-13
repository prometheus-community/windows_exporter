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

package shell32

import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/go-ole/go-ole"
	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var (
	modShell32                      = windows.NewLazySystemDLL("shell32.dll")
	procSHCreateItemFromParsingName = modShell32.NewProc("SHCreateItemFromParsingName")
)

func SHCreateItemFromParsingName(path string, iid *ole.GUID) (*ole.IDispatch, error) {
	ptrPath, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, fmt.Errorf("failed to convert path to UTF16: %w", err)
	}

	var result *ole.IDispatch

	hr, _, err := procSHCreateItemFromParsingName.Call(
		uintptr(unsafe.Pointer(ptrPath)),
		0,
		uintptr(unsafe.Pointer(iid)),
		uintptr(unsafe.Pointer(&result)),
	)
	if hr != 0 {
		return nil, fmt.Errorf("syscall failed: %w", err)
	}

	if result == nil {
		return nil, errors.New("SHCreateItemFromParsingName returned nil")
	}

	return result, nil
}

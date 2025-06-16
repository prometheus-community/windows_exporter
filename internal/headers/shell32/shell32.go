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
	"syscall"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/prometheus-community/windows_exporter/internal/headers/propsys"
	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var (
	modShell32                      = windows.NewLazySystemDLL("shell32.dll")
	procSHCreateItemFromParsingName = modShell32.NewProc("SHCreateItemFromParsingName")

	iidIShellItem2 = ole.NewGUID("{7E9FB0D3-919F-4307-AB2E-9B1860310C93}")
)

func SHCreateItemFromParsingName(path string) (*IShellItem2, error) {
	ptrPath, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, fmt.Errorf("failed to convert path to UTF16: %w", err)
	}

	var result *IShellItem2

	hr, _, err := procSHCreateItemFromParsingName.Call(
		uintptr(unsafe.Pointer(ptrPath)),
		0,
		uintptr(unsafe.Pointer(iidIShellItem2)),
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

func (item *IShellItem2) GetProperty(key *propsys.PROPERTYKEY, v *ole.VARIANT) error {
	hr, _, err := syscall.SyscallN(
		item.lpVtbl.GetProperty,
		uintptr(unsafe.Pointer(item)),
		uintptr(unsafe.Pointer(key)),
		uintptr(unsafe.Pointer(v)),
	)

	if hr != 0 {
		return fmt.Errorf("GetProperty failed: %w", err)
	}

	return nil
}

func (item *IShellItem2) Release() {
	_, _, _ = syscall.SyscallN(
		item.lpVtbl.Release,
		uintptr(unsafe.Pointer(item)),
	)
}

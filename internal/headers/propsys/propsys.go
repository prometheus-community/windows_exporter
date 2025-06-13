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

package propsys

import (
	"fmt"
	"unsafe"

	"github.com/go-ole/go-ole"
	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var (
	modPropsys                   = windows.NewLazySystemDLL("propsys.dll")
	procPSGetPropertyKeyFromName = modPropsys.NewProc("PSGetPropertyKeyFromName")
)

type PROPERTYKEY struct {
	Fmtid ole.GUID
	Pid   uint32
}

func PSGetPropertyKeyFromName(name string, key *PROPERTYKEY) error {
	namePtr, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return fmt.Errorf("failed to convert name to UTF16: %w", err)
	}

	hr, _, err := procPSGetPropertyKeyFromName.Call(
		uintptr(unsafe.Pointer(namePtr)),
		uintptr(unsafe.Pointer(key)),
	)

	if hr != 0 {
		return fmt.Errorf("PSGetPropertyKeyFromName failed: %w", err)
	}

	return nil
}

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

package gdi32

import (
	"fmt"
	"unsafe"

	"github.com/prometheus-community/windows_exporter/internal/headers/lazyproc"
	"github.com/prometheus-community/windows_exporter/internal/headers/ntdll"
	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var (
	modGdi32                      = windows.NewLazySystemDLL("gdi32.dll")
	procD3DKMTOpenAdapterFromLuid = modGdi32.NewProc("D3DKMTOpenAdapterFromLuid")
	procD3DKMTQueryAdapterInfo    = modGdi32.NewProc("D3DKMTQueryAdapterInfo")
	procD3DKMTCloseAdapter        = modGdi32.NewProc("D3DKMTCloseAdapter")
	procD3DKMTEnumAdapters2       = modGdi32.NewProc("D3DKMTEnumAdapters2")

	load = lazyproc.Once(modGdi32,
		procD3DKMTOpenAdapterFromLuid,
		procD3DKMTQueryAdapterInfo,
		procD3DKMTCloseAdapter,
		procD3DKMTEnumAdapters2,
	)
)

// Load checks that gdi32.dll, and the ntdll functions this package uses,
// export every function of this package. Call it in Build: a missing export
// panics on the first call otherwise.
func Load() error {
	if err := load(); err != nil {
		return err
	}

	return ntdll.Load()
}

func D3DKMTOpenAdapterFromLuid(ptr *D3DKMT_OPENADAPTERFROMLUID) error {
	ret, _, _ := procD3DKMTOpenAdapterFromLuid.Call(
		uintptr(unsafe.Pointer(ptr)),
	)
	if ret != 0 {
		return fmt.Errorf("D3DKMTOpenAdapterFromLuid failed: 0x%X: %w", ret, ntdll.RtlNtStatusToDosError(ret))
	}

	return nil
}

func D3DKMTEnumAdapters2(ptr *D3DKMT_ENUMADAPTERS2) error {
	ret, _, _ := procD3DKMTEnumAdapters2.Call(
		uintptr(unsafe.Pointer(ptr)),
	)
	if ret != 0 {
		return fmt.Errorf("D3DKMTEnumAdapters2 failed: 0x%X: %w", ret, ntdll.RtlNtStatusToDosError(ret))
	}

	return nil
}

func D3DKMTQueryAdapterInfo(query *D3DKMT_QUERYADAPTERINFO) error {
	ret, _, _ := procD3DKMTQueryAdapterInfo.Call(
		uintptr(unsafe.Pointer(query)),
	)
	if ret != 0 {
		return fmt.Errorf("D3DKMTQueryAdapterInfo failed: 0x%X: %w", ret, ntdll.RtlNtStatusToDosError(ret))
	}

	return nil
}

func D3DKMTCloseAdapter(ptr *D3DKMT_CLOSEADAPTER) error {
	ret, _, _ := procD3DKMTCloseAdapter.Call(
		uintptr(unsafe.Pointer(ptr)),
	)
	if ret != 0 {
		return fmt.Errorf("D3DKMTCloseAdapter failed: 0x%X: %w", ret, ntdll.RtlNtStatusToDosError(ret))
	}

	return nil
}

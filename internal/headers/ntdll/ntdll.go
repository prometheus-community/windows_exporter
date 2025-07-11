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

package ntdll

import (
	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var (
	modNtdll                  = windows.NewLazySystemDLL("ntdll.dll")
	procRtlNtStatusToDosError = modNtdll.NewProc("RtlNtStatusToDosError")
)

func RtlNtStatusToDosError(status uintptr) error {
	ret, _, _ := procRtlNtStatusToDosError.Call(status)
	if ret == 0 {
		return nil
	}

	return windows.Errno(ret)
}

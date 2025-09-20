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

package osversion

import (
	"fmt"
	"sync"

	"github.com/prometheus-community/windows_exporter/internal/headers/sysinfoapi"
	"golang.org/x/sys/windows"
)

// OSVersion is a wrapper for Windows version information
// https://msdn.microsoft.com/en-us/library/windows/desktop/ms724439(v=vs.85).aspx
type OSVersion struct {
	Version            uint32
	MajorVersion       uint8
	MinorVersion       uint8
	Build              uint16
	OperatingSystemSKU sysinfoapi.OperatingSystemSKU
	ProductType        ProductType
}

type ProductType uint8

func (pt ProductType) String() string {
	switch pt {
	case 1:
		return "Workstation"
	case 2:
		return "Domain Controller"
	case 3:
		return "Server"
	default:
		return "Unknown"
	}
}

//nolint:gochecknoglobals
var osv = sync.OnceValue(func() OSVersion {
	v := *windows.RtlGetVersion()

	var operatingSystemSKU sysinfoapi.OperatingSystemSKU

	_ = sysinfoapi.GetProductInfo(
		v.MajorVersion, v.MinorVersion,
		uint32(v.ServicePackMajor), uint32(v.ServicePackMinor),
		&operatingSystemSKU,
	)

	return OSVersion{
		MajorVersion: uint8(v.MajorVersion),
		MinorVersion: uint8(v.MinorVersion),
		Build:        uint16(v.BuildNumber),
		// Fill version value so that existing clients don't break
		Version:            v.BuildNumber<<16 | (v.MinorVersion << 8) | v.MajorVersion,
		OperatingSystemSKU: operatingSystemSKU,
		ProductType:        ProductType(v.ProductType),
	}
})

// Get gets the operating system version on Windows.
// The calling application must be manifested to get the correct version information.
func Get() OSVersion {
	return osv()
}

// Build gets the build-number on Windows
// The calling application must be manifested to get the correct version information.
func Build() uint16 {
	return Get().Build
}

// String returns the OSVersion formatted as a string. It implements the
// [fmt.Stringer] interface.
func (osv OSVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", osv.MajorVersion, osv.MinorVersion, osv.Build)
}

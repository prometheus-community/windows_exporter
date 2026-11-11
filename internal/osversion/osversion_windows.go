// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package osversion

import (
	"fmt"
	"sync"

	"golang.org/x/sys/windows"
)

// OSVersion is a wrapper for Windows version information
// https://msdn.microsoft.com/en-us/library/windows/desktop/ms724439(v=vs.85).aspx
type OSVersion struct {
	Version      uint32
	MajorVersion uint8
	MinorVersion uint8
	Build        uint16
}

//nolint:gochecknoglobals
var osv = sync.OnceValue(func() OSVersion {
	v := *windows.RtlGetVersion()

	return OSVersion{
		MajorVersion: uint8(v.MajorVersion),
		MinorVersion: uint8(v.MinorVersion),
		Build:        uint16(v.BuildNumber),
		// Fill version value so that existing clients don't break
		Version: v.BuildNumber<<16 | (v.MinorVersion << 8) | v.MajorVersion,
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

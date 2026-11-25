// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

// Package nsi reads parameters of the network store interface (nsi.dll), the
// store behind the IP Helper API and netsh. The API is undocumented; the
// object index and the parameter layout here were found by probing Windows
// 11 build 26300 and Windows Server 2025 build 26100, see
// internal/collector/tcp/AGENTS.md.
package nsi

import (
	"encoding/binary"
	"fmt"
	"syscall"
	"unsafe"

	"github.com/prometheus-community/windows_exporter/internal/headers/lazyproc"
	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var (
	modnsi              = windows.NewLazySystemDLL("nsi.dll")
	procNsiGetParameter = modnsi.NewProc("NsiGetParameter")

	load = lazyproc.Once(modnsi, procNsiGetParameter)
)

// Load checks that nsi.dll exports every function this package calls. Call it
// in Build: a missing export panics on the first call otherwise.
func Load() error { return load() }

const (
	// storeActive is NsiActive of NSI_STORE, the values in effect.
	storeActive = 1
	// structTypeRW is NsiStructRw of NSI_STRUCT_TYPE.
	structTypeRW = 0
	// moduleIDTypeGUID is MIT_GUID of NPI_MODULEID_TYPE.
	moduleIDTypeGUID = 1
	// tcpObjectDynamicPortRange is the keyless object of the TCP module that
	// holds the dynamic port range of netsh int ipv4|ipv6 show dynamicport tcp.
	tcpObjectDynamicPortRange = 20
)

// npiModuleID is NPI_MODULEID: Length, the type and a union of a GUID and a
// LUID, 24 bytes.
type npiModuleID struct {
	length   uint16
	_        uint16
	idType   uint32
	moduleID windows.GUID
}

// tcpModuleID is only read; NsiGetParameter does not modify it.
//
//nolint:gochecknoglobals
var tcpModuleID = npiModuleID{
	length: uint16(unsafe.Sizeof(npiModuleID{})),
	idType: moduleIDTypeGUID,
	// NPI_MS_TCP_MODULEID.
	moduleID: windows.GUID{Data1: 0xeb004a03, Data2: 0x9b1a, Data3: 0x11d4, Data4: [8]byte{0x91, 0x23, 0x00, 0x50, 0x04, 0x77, 0x59, 0xbc}},
}

// TCPDynamicPortRange returns the first port and the number of ports of the
// TCP dynamic port range. IPv4 and IPv6 share it.
func TCPDynamicPortRange() (uint16, uint16, error) {
	var buf [4]byte

	r0, _, _ := syscall.SyscallN(procNsiGetParameter.Addr(),
		storeActive,
		uintptr(unsafe.Pointer(&tcpModuleID)),
		tcpObjectDynamicPortRange,
		0, // No key.
		0,
		structTypeRW,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
		0, // Offset.
	)
	// The function returns a 32-bit Win32 error code.
	if uint32(r0) != 0 {
		return 0, 0, fmt.Errorf("NsiGetParameter: %w", windows.Errno(uint32(r0)))
	}

	start, count := parsePortRange(buf)

	return start, count, nil
}

// parsePortRange decodes the 4-byte range: the first port in network byte
// order, then the number of ports in host order.
func parsePortRange(buf [4]byte) (uint16, uint16) {
	return binary.BigEndian.Uint16(buf[0:2]), binary.LittleEndian.Uint16(buf[2:4])
}

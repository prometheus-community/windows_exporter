// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

// Package nsi reads parameters of the network store interface (nsi.dll), the
// store behind the IP Helper API and netsh. The API is undocumented; see
// AGENTS.md in this directory for how the calls and layouts were found.
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
	// storeActive is the NSI store of the values in effect.
	storeActive = 1
	// structTypeRW selects the read-write part of an object.
	structTypeRW = 0
	// moduleIDTypeGUID is MIT_GUID of NPI_MODULEID_TYPE.
	moduleIDTypeGUID = 1
)

// npiModuleID is NPI_MODULEID: Length, the type and a union of a GUID and a
// LUID, 24 bytes.
type npiModuleID struct {
	length   uint16
	_        uint16
	idType   uint32
	moduleID windows.GUID
}

// Protocol selects the transport whose dynamic port range is read.
type Protocol int

const (
	ProtocolTCP Protocol = iota
	ProtocolUDP
)

// dynamicPortRangeObject is the NSI module and the keyless object that hold
// the dynamic port range of a protocol, as shown by
// netsh int ipv4|ipv6 show dynamicport tcp|udp.
type dynamicPortRangeObject struct {
	module      npiModuleID
	objectIndex uint32
}

// The objects are only read; NsiGetParameter does not modify them.
//
//nolint:gochecknoglobals
var dynamicPortRangeObjects = [...]dynamicPortRangeObject{
	ProtocolTCP: {
		// NPI_MS_TCP_MODULEID.
		module:      newModuleID(windows.GUID{Data1: 0xeb004a03, Data2: 0x9b1a, Data3: 0x11d4, Data4: [8]byte{0x91, 0x23, 0x00, 0x50, 0x04, 0x77, 0x59, 0xbc}}),
		objectIndex: 20,
	},
	ProtocolUDP: {
		// NPI_MS_UDP_MODULEID.
		module:      newModuleID(windows.GUID{Data1: 0xeb004a02, Data2: 0x9b1a, Data3: 0x11d4, Data4: [8]byte{0x91, 0x23, 0x00, 0x50, 0x04, 0x77, 0x59, 0xbc}}),
		objectIndex: 4,
	},
}

func newModuleID(guid windows.GUID) npiModuleID {
	return npiModuleID{
		length:   uint16(unsafe.Sizeof(npiModuleID{})),
		idType:   moduleIDTypeGUID,
		moduleID: guid,
	}
}

// PortRange is a range of local ports.
type PortRange struct {
	Start         uint16
	NumberOfPorts uint16
}

// DynamicPortRange returns the active dynamic port range of protocol. IPv4
// and IPv6 share it.
func DynamicPortRange(protocol Protocol) (PortRange, error) {
	if protocol < 0 || int(protocol) >= len(dynamicPortRangeObjects) {
		return PortRange{}, fmt.Errorf("unsupported protocol %d", protocol)
	}

	object := &dynamicPortRangeObjects[protocol]

	var buf [4]byte

	r0, _, _ := syscall.SyscallN(procNsiGetParameter.Addr(),
		storeActive,
		uintptr(unsafe.Pointer(&object.module)),
		uintptr(object.objectIndex),
		0, // No key.
		0,
		structTypeRW,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
		0, // Offset.
	)
	// The function returns a 32-bit Win32 error code.
	if uint32(r0) != 0 {
		return PortRange{}, fmt.Errorf("NsiGetParameter: %w", windows.Errno(uint32(r0)))
	}

	return parsePortRange(buf), nil
}

// parsePortRange decodes the 4-byte range: the first port in network byte
// order, then the number of ports in host order.
func parsePortRange(buf [4]byte) PortRange {
	return PortRange{
		Start:         binary.BigEndian.Uint16(buf[0:2]),
		NumberOfPorts: binary.LittleEndian.Uint16(buf[2:4]),
	}
}

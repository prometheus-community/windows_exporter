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
	"slices"
	"syscall"
	"unsafe"

	"github.com/prometheus-community/windows_exporter/internal/headers/lazyproc"
	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var (
	modnsi                     = windows.NewLazySystemDLL("nsi.dll")
	procNsiGetParameter        = modnsi.NewProc("NsiGetParameter")
	procNsiAllocateAndGetTable = modnsi.NewProc("NsiAllocateAndGetTable")
	procNsiFreeTable           = modnsi.NewProc("NsiFreeTable")

	load = lazyproc.Once(modnsi, procNsiGetParameter, procNsiAllocateAndGetTable, procNsiFreeTable)
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

// protocolObjects are the NSI module and object indices of a protocol.
type protocolObjects struct {
	module npiModuleID
	// dynamicPortRange is the keyless object of
	// netsh int ipv4|ipv6 show dynamicport tcp|udp.
	dynamicPortRange uint32
	// excludedPortRanges is the table of
	// netsh int ipv4|ipv6 show excludedportrange tcp|udp.
	excludedPortRanges uint32
}

// The modules are only read; NSI does not modify them.
//
//nolint:gochecknoglobals
var objects = [...]protocolObjects{
	ProtocolTCP: {
		// NPI_MS_TCP_MODULEID.
		module:             newModuleID(windows.GUID{Data1: 0xeb004a03, Data2: 0x9b1a, Data3: 0x11d4, Data4: [8]byte{0x91, 0x23, 0x00, 0x50, 0x04, 0x77, 0x59, 0xbc}}),
		dynamicPortRange:   20,
		excludedPortRanges: 29,
	},
	ProtocolUDP: {
		// NPI_MS_UDP_MODULEID.
		module:             newModuleID(windows.GUID{Data1: 0xeb004a02, Data2: 0x9b1a, Data3: 0x11d4, Data4: [8]byte{0x91, 0x23, 0x00, 0x50, 0x04, 0x77, 0x59, 0xbc}}),
		dynamicPortRange:   4,
		excludedPortRanges: 5,
	},
}

func objectsOf(protocol Protocol) (*protocolObjects, error) {
	if protocol < 0 || int(protocol) >= len(objects) {
		return nil, fmt.Errorf("unsupported protocol %d", protocol)
	}

	return &objects[protocol], nil
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

// Contains reports whether port lies in the range. The arithmetic uses
// uint32, so a range that ends at port 65535 does not wrap around.
func (r PortRange) Contains(port uint16) bool {
	return port >= r.Start && uint32(port) < uint32(r.Start)+uint32(r.NumberOfPorts)
}

// Usage splits the range into the ports that exclusions reserve and the
// ports that inUse reports as taken and that no exclusion covers. Exclusions
// can extend beyond the range; only the overlap counts.
func (r PortRange) Usage(exclusions []PortRange, inUse func(port uint16) bool) (excluded, used uint32) {
	var reserved [65536 / 64]uint64

	for _, e := range exclusions {
		for p := uint32(e.Start); p < uint32(e.Start)+uint32(e.NumberOfPorts); p++ {
			reserved[p/64] |= 1 << (p % 64)
		}
	}

	for p := uint32(r.Start); p < uint32(r.Start)+uint32(r.NumberOfPorts); p++ {
		switch {
		case reserved[p/64]&(1<<(p%64)) != 0:
			excluded++
		case inUse(uint16(p)):
			used++
		}
	}

	return excluded, used
}

// DynamicPortRange returns the active dynamic port range of protocol. IPv4
// and IPv6 share it.
func DynamicPortRange(protocol Protocol) (PortRange, error) {
	object, err := objectsOf(protocol)
	if err != nil {
		return PortRange{}, err
	}

	var buf [4]byte

	r0, _, _ := syscall.SyscallN(procNsiGetParameter.Addr(),
		storeActive,
		uintptr(unsafe.Pointer(&object.module)),
		uintptr(object.dynamicPortRange),
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

// excludedPortRangeKeySize is the size of a key of the excluded port range
// table: the first port and the number of ports, both in host order.
const excludedPortRangeKeySize = 4

// ExcludedPortRanges returns the active excluded port ranges of protocol, as
// shown by netsh int ipv4 show excludedportrange. IPv4 and IPv6 share them.
func ExcludedPortRanges(protocol Protocol) ([]PortRange, error) {
	object, err := objectsOf(protocol)
	if err != nil {
		return nil, err
	}

	// Only the keys are needed; the 2-byte read-write part holds flags.
	raw, err := getTableKeys(&object.module, object.excludedPortRanges, excludedPortRangeKeySize)
	if err != nil {
		return nil, err
	}

	return parseExcludedPortRanges(raw), nil
}

// parseExcludedPortRanges decodes the keys of the excluded port range table.
func parseExcludedPortRanges(raw []byte) []PortRange {
	ranges := make([]PortRange, 0, len(raw)/excludedPortRangeKeySize)

	for key := range slices.Chunk(raw, excludedPortRangeKeySize) {
		if len(key) < excludedPortRangeKeySize {
			break
		}

		ranges = append(ranges, PortRange{
			Start:         binary.LittleEndian.Uint16(key[0:2]),
			NumberOfPorts: binary.LittleEndian.Uint16(key[2:4]),
		})
	}

	return ranges
}

const (
	// tcpBoundEndpointsObject is the TCP table behind
	// InternalGetBoundTcpEndpointTable and InternalGetBoundTcp6EndpointTable:
	// sockets that are bound to a local port but neither listen nor connect.
	tcpBoundEndpointsObject = 35
	// tcpBoundEndpointKeySize is the size of a key of the bound endpoint
	// table. It starts with the local SOCKADDR_INET.
	tcpBoundEndpointKeySize = 56
)

// BoundTCPPorts returns the local port of every bound TCP endpoint of both
// address families. The list can contain a port more than once.
func BoundTCPPorts() ([]uint16, error) {
	// Only the keys are needed. Requesting the dynamic part as well
	// misaligns the rows on Windows Server 2022.
	raw, err := getTableKeys(&objects[ProtocolTCP].module, tcpBoundEndpointsObject, tcpBoundEndpointKeySize)
	if err != nil {
		return nil, err
	}

	return parseBoundTCPPorts(raw), nil
}

// parseBoundTCPPorts decodes the local ports of the bound endpoint table keys.
// The key starts with SOCKADDR_INET: the address family in host order, then
// the port in network byte order for both AF_INET and AF_INET6.
func parseBoundTCPPorts(raw []byte) []uint16 {
	ports := make([]uint16, 0, len(raw)/tcpBoundEndpointKeySize)

	for key := range slices.Chunk(raw, tcpBoundEndpointKeySize) {
		if len(key) < tcpBoundEndpointKeySize {
			break
		}

		switch binary.LittleEndian.Uint16(key[0:2]) {
		case windows.AF_INET, windows.AF_INET6:
			ports = append(ports, binary.BigEndian.Uint16(key[2:4]))
		}
	}

	return ports
}

// getTableKeys reads only the keys of a table of the active store and returns
// a copy of them.
func getTableKeys(module *npiModuleID, object uint32, keySize uintptr) ([]byte, error) {
	var (
		keys  unsafe.Pointer
		count uint32
	)

	r0, _, _ := syscall.SyscallN(procNsiAllocateAndGetTable.Addr(),
		storeActive,
		uintptr(unsafe.Pointer(module)),
		uintptr(object),
		uintptr(unsafe.Pointer(&keys)),
		keySize,
		0, 0, // No read-write part.
		0, 0, // No read-only dynamic part.
		0, 0, // No read-only static part.
		uintptr(unsafe.Pointer(&count)),
		0,
	)
	// The function returns a 32-bit Win32 error code.
	if uint32(r0) != 0 {
		return nil, fmt.Errorf("NsiAllocateAndGetTable: %w", windows.Errno(uint32(r0)))
	}

	if keys == nil {
		return nil, nil
	}

	defer func() {
		_, _, _ = syscall.SyscallN(procNsiFreeTable.Addr(), uintptr(keys), 0, 0, 0)
	}()

	return slices.Clone(unsafe.Slice((*byte)(keys), uintptr(count)*keySize)), nil
}

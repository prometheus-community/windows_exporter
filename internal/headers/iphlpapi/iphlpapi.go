// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package iphlpapi

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
	modiphlpapi             = windows.NewLazySystemDLL("iphlpapi.dll")
	procGetExtendedTcpTable = modiphlpapi.NewProc("GetExtendedTcpTable")

	load = lazyproc.Once(modiphlpapi, procGetExtendedTcpTable)
)

// Load checks that iphlpapi.dll exports every function of this package. Call
// it in Build: a missing export panics on the first call otherwise.
func Load() error { return load() }

// PortRange is a range of local TCP ports, such as the dynamic port range.
type PortRange struct {
	Start         uint16
	NumberOfPorts uint16
}

// contains reports whether port lies in the range. The arithmetic uses uint32,
// so a range that ends at port 65535 does not wrap around.
func (r PortRange) contains(port uint16) bool {
	return port >= r.Start && uint32(port) < uint32(r.Start)+uint32(r.NumberOfPorts)
}

// TCPConnectionStats summarizes the TCP connection table of one address family.
type TCPConnectionStats struct {
	// States holds the number of connections in each state.
	States map[MIB_TCP_STATE]uint32
	// PortsInRange is the number of distinct local ports inside the port range
	// passed to GetTCPConnectionStats. Every connection counts, whatever its
	// state, because a connection in TIME_WAIT still holds its local port.
	PortsInRange uint32
}

// GetTCPConnectionStats reads the TCP connection table of family once and
// returns the connection count per state and the number of distinct local
// ports inside portRange. An empty portRange counts no ports.
func GetTCPConnectionStats(family uint32, portRange PortRange) (TCPConnectionStats, error) {
	switch family {
	case windows.AF_INET:
		table, err := getExtendedTcpTable[MIB_TCPROW_OWNER_PID](family, TCPTableOwnerPIDAll)
		if err != nil {
			return TCPConnectionStats{}, fmt.Errorf("failed getExtendedTcpTable: %w", err)
		}

		return summarizeTCPTable(table, portRange), nil
	case windows.AF_INET6:
		table, err := getExtendedTcpTable[MIB_TCP6ROW_OWNER_PID](family, TCPTableOwnerPIDAll)
		if err != nil {
			return TCPConnectionStats{}, fmt.Errorf("failed getExtendedTcpTable: %w", err)
		}

		return summarizeTCPTable(table, portRange), nil
	default:
		return TCPConnectionStats{}, fmt.Errorf("unsupported address family %d", family)
	}
}

type tcpRow interface {
	MIB_TCPROW_OWNER_PID | MIB_TCP6ROW_OWNER_PID

	state() MIB_TCP_STATE
	localPort() uint16
}

func summarizeTCPTable[T tcpRow](table []T, portRange PortRange) TCPConnectionStats {
	stats := TCPConnectionStats{
		States: make(map[MIB_TCP_STATE]uint32),
	}

	// One bit per port. Several connections can share a local port, for
	// example a listener and the connections it accepted.
	var seen [65536 / 64]uint64

	for _, row := range table {
		stats.States[row.state()]++

		port := row.localPort()
		if !portRange.contains(port) {
			continue
		}

		word, bit := port/64, uint64(1)<<(port%64)
		if seen[word]&bit == 0 {
			seen[word] |= bit
			stats.PortsInRange++
		}
	}

	return stats
}

func GetOwnerPIDOfTCPPort(family uint32, tcpPort uint16) (uint32, error) {
	switch family {
	case windows.AF_INET:
		table, err := getExtendedTcpTable[MIB_TCPROW_OWNER_PID](family, TCPTableOwnerPIDListener)
		if err != nil {
			return 0, err
		}

		for _, row := range table {
			if row.dwLocalPort.uint16() == tcpPort {
				return row.dwOwningPid, nil
			}
		}

		return 0, fmt.Errorf("no process found for port %d", tcpPort)
	case windows.AF_INET6:
		table, err := getExtendedTcpTable[MIB_TCP6ROW_OWNER_PID](family, TCPTableOwnerPIDListener)
		if err != nil {
			return 0, err
		}

		for _, row := range table {
			if row.dwLocalPort.uint16() == tcpPort {
				return row.dwOwningPid, nil
			}
		}

		return 0, fmt.Errorf("no process found for port %d", tcpPort)
	default:
		return 0, fmt.Errorf("unsupported address family %d", family)
	}
}

func getExtendedTcpTable[T any](ulAf uint32, tableClass uint32) ([]T, error) {
	var size uint32

	ret, _, _ := syscall.SyscallN(
		procGetExtendedTcpTable.Addr(),
		0,
		uintptr(unsafe.Pointer(&size)),
		0,
		uintptr(ulAf),
		uintptr(tableClass),
		0,
	)

	if ret != uintptr(windows.ERROR_INSUFFICIENT_BUFFER) {
		return nil, fmt.Errorf("getExtendedTcpTable (size query) failed with code %d", ret)
	}

	buf := make([]byte, size)

	ret, _, _ = syscall.SyscallN(
		procGetExtendedTcpTable.Addr(),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
		0,
		uintptr(ulAf),
		uintptr(tableClass),
		0,
	)

	if ret != 0 {
		return nil, fmt.Errorf("getExtendedTcpTable (data query) failed with code %d", ret)
	}

	return unsafe.Slice((*T)(unsafe.Pointer(&buf[4])), binary.LittleEndian.Uint32(buf)), nil
}

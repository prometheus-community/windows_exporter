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
	procGetExtendedUdpTable = modiphlpapi.NewProc("GetExtendedUdpTable")

	load = lazyproc.Once(modiphlpapi, procGetExtendedTcpTable, procGetExtendedUdpTable)
)

// Load checks that iphlpapi.dll exports every function of this package. Call
// it in Build: a missing export panics on the first call otherwise.
func Load() error { return load() }

// GetTCPConnectionStates reads the TCP connection table of family once and
// returns the number of connections per state. If ports is not nil, it also
// adds the local port of every connection that holds its port to ports, see
// holdsLocalPort.
func GetTCPConnectionStates(family uint32, ports *PortSet) (map[MIB_TCP_STATE]uint32, error) {
	switch family {
	case windows.AF_INET:
		table, err := getExtendedTcpTable[MIB_TCPROW_OWNER_PID](family, TCPTableOwnerPIDAll)
		if err != nil {
			return nil, fmt.Errorf("failed getExtendedTcpTable: %w", err)
		}

		return summarizeTCPTable(table, ports), nil
	case windows.AF_INET6:
		table, err := getExtendedTcpTable[MIB_TCP6ROW_OWNER_PID](family, TCPTableOwnerPIDAll)
		if err != nil {
			return nil, fmt.Errorf("failed getExtendedTcpTable: %w", err)
		}

		return summarizeTCPTable(table, ports), nil
	default:
		return nil, fmt.Errorf("unsupported address family %d", family)
	}
}

type tcpRow interface {
	MIB_TCPROW_OWNER_PID | MIB_TCP6ROW_OWNER_PID

	state() MIB_TCP_STATE
	localPort() uint16
}

func summarizeTCPTable[T tcpRow](table []T, ports *PortSet) map[MIB_TCP_STATE]uint32 {
	stateCounts := make(map[MIB_TCP_STATE]uint32)

	for _, row := range table {
		state := row.state()
		stateCounts[state]++

		if ports != nil && holdsLocalPort(state) {
			ports.Add(row.localPort())
		}
	}

	return stateCounts
}

// holdsLocalPort reports whether a connection in state blocks its local port
// for new port-0 binds. After closesocket, TIME_WAIT and FIN_WAIT2 keep only a
// weak reference to the port; measured on Windows Server 2022 and 2025. A
// half-closed FIN_WAIT2 socket that is still open does hold its port, but the
// table cannot tell the two apart.
func holdsLocalPort(state MIB_TCP_STATE) bool {
	return state != TCPStateTimeWait && state != TCPStateFinWait2
}

// AddUDPPorts reads the UDP endpoint table of family and adds the local port
// of every endpoint to ports.
func AddUDPPorts(family uint32, ports *PortSet) error {
	switch family {
	case windows.AF_INET:
		table, err := getExtendedTable[MIB_UDPROW_OWNER_PID](procGetExtendedUdpTable, family, UDPTableOwnerPID)
		if err != nil {
			return fmt.Errorf("failed getExtendedUdpTable: %w", err)
		}

		for _, row := range table {
			ports.Add(row.dwLocalPort.uint16())
		}

		return nil
	case windows.AF_INET6:
		table, err := getExtendedTable[MIB_UDP6ROW_OWNER_PID](procGetExtendedUdpTable, family, UDPTableOwnerPID)
		if err != nil {
			return fmt.Errorf("failed getExtendedUdpTable: %w", err)
		}

		for _, row := range table {
			ports.Add(row.dwLocalPort.uint16())
		}

		return nil
	default:
		return fmt.Errorf("unsupported address family %d", family)
	}
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
	return getExtendedTable[T](procGetExtendedTcpTable, ulAf, tableClass)
}

// getExtendedTable calls GetExtendedTcpTable or GetExtendedUdpTable, which
// share their signature. The table can grow between the size query and the
// read, so the read is retried a few times with the new size.
func getExtendedTable[T any](proc *windows.LazyProc, ulAf uint32, tableClass uint32) ([]T, error) {
	var size uint32

	ret, _, _ := syscall.SyscallN(
		proc.Addr(),
		0,
		uintptr(unsafe.Pointer(&size)),
		0,
		uintptr(ulAf),
		uintptr(tableClass),
		0,
	)

	if ret != uintptr(windows.ERROR_INSUFFICIENT_BUFFER) {
		return nil, fmt.Errorf("%s (size query) failed with code %d", proc.Name, ret)
	}

	for range 3 {
		buf := make([]byte, size)

		ret, _, _ = syscall.SyscallN(
			proc.Addr(),
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(unsafe.Pointer(&size)),
			0,
			uintptr(ulAf),
			uintptr(tableClass),
			0,
		)

		switch ret {
		case 0:
			return unsafe.Slice((*T)(unsafe.Pointer(&buf[4])), binary.LittleEndian.Uint32(buf)), nil
		case uintptr(windows.ERROR_INSUFFICIENT_BUFFER):
			continue
		default:
			return nil, fmt.Errorf("%s (data query) failed with code %d", proc.Name, ret)
		}
	}

	return nil, fmt.Errorf("%s (data query) failed: the table kept growing", proc.Name)
}

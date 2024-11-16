//go:build windows

package iphlpapi

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modiphlpapi             = windows.NewLazySystemDLL("iphlpapi.dll")
	procGetExtendedTcpTable = modiphlpapi.NewProc("GetExtendedTcpTable")
)

func GetTCPConnectionStates(family uint32) (map[MIB_TCP_STATE]uint32, error) {
	stateCounts := make(map[MIB_TCP_STATE]uint32)

	switch family {
	case windows.AF_INET:
		table, err := getExtendedTcpTable[MIB_TCPROW_OWNER_PID](family, TCPTableOwnerPIDAll)
		if err != nil {
			return nil, err
		}

		for _, row := range table {
			stateCounts[row.dwState]++
		}

		return stateCounts, nil
	case windows.AF_INET6:
		table, err := getExtendedTcpTable[MIB_TCP6ROW_OWNER_PID](family, TCPTableOwnerPIDAll)
		if err != nil {
			return nil, err
		}

		for _, row := range table {
			stateCounts[row.dwState]++
		}

		return stateCounts, nil
	default:
		return nil, fmt.Errorf("unsupported address family %d", family)
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
	var size uint32

	ret, _, _ := procGetExtendedTcpTable.Call(
		uintptr(0),
		uintptr(unsafe.Pointer(&size)),
		uintptr(0),
		uintptr(ulAf),
		uintptr(tableClass),
		uintptr(0),
	)

	if ret != uintptr(windows.ERROR_INSUFFICIENT_BUFFER) {
		return nil, fmt.Errorf("getExtendedTcpTable (size query) failed with code %d", ret)
	}

	buf := make([]byte, size)

	ret, _, _ = procGetExtendedTcpTable.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
		uintptr(0),
		uintptr(ulAf),
		uintptr(tableClass),
		uintptr(0),
	)

	if ret != 0 {
		return nil, fmt.Errorf("getExtendedTcpTable (data query) failed with code %d", ret)
	}

	return unsafe.Slice((*T)(unsafe.Pointer(&buf[4])), binary.LittleEndian.Uint32(buf)), nil
}

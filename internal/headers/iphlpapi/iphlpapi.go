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

package iphlpapi

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/go-ole/go-ole"
	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var (
	modiphlpapi                    = windows.NewLazySystemDLL("iphlpapi.dll")
	procGetExtendedTcpTable        = modiphlpapi.NewProc("GetExtendedTcpTable")
	procGetIfEntry2Ex              = modiphlpapi.NewProc("GetIfEntry2Ex")
	procConvertInterfaceGuidToLuid = modiphlpapi.NewProc("ConvertInterfaceGuidToLuid")
)

func GetTCPConnectionStates(family uint32) (map[MIB_TCP_STATE]uint32, error) {
	stateCounts := make(map[MIB_TCP_STATE]uint32)

	switch family {
	case windows.AF_INET:
		table, err := getExtendedTcpTable[MIB_TCPROW_OWNER_PID](family, TCPTableOwnerPIDAll)
		if err != nil {
			return nil, fmt.Errorf("failed getExtendedTcpTable: %w", err)
		}

		for _, row := range table {
			stateCounts[row.dwState]++
		}

		return stateCounts, nil
	case windows.AF_INET6:
		table, err := getExtendedTcpTable[MIB_TCP6ROW_OWNER_PID](family, TCPTableOwnerPIDAll)
		if err != nil {
			return nil, fmt.Errorf("failed getExtendedTcpTable: %w", err)
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

	ret, _, _ = procGetExtendedTcpTable.Call(
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

// GetIfEntry2Ex function retrieves the specified level of information for the specified interface on the local computer.
//
// https://learn.microsoft.com/en-us/windows/win32/api/netioapi/nf-netioapi-getifentry2ex
func GetIfEntry2Ex(row *MIB_IF_ROW2) error {
	ret, _, _ := procGetIfEntry2Ex.Call(
		uintptr(0),
		uintptr(unsafe.Pointer(row)),
	)

	if ret != 0 {
		return fmt.Errorf("GetIfEntry2Ex failed with code %d: %w", ret, windows.Errno(ret))
	}

	return nil
}

// ConvertInterfaceGUIDToLUID function converts a globally unique identifier (GUID) for a network interface to the
// locally unique identifier (LUID) for the interface.
//
// https://learn.microsoft.com/en-us/windows/win32/api/netioapi/nf-netioapi-convertinterfaceguidtoluid
func ConvertInterfaceGUIDToLUID(guid ole.GUID) (uint64, error) {
	var luid uint64

	ret, _, _ := procConvertInterfaceGuidToLuid.Call(
		uintptr(unsafe.Pointer(&guid)),
		uintptr(unsafe.Pointer(&luid)),
	)

	if ret != 0 {
		return 0, fmt.Errorf("ConvertInterfaceGUIDToLUID failed with code %d: %w", ret, windows.Errno(ret))
	}

	return luid, nil
}

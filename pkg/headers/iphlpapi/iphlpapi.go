package iphlpapi

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modiphlpapi             = windows.NewLazySystemDLL("iphlpapi.dll")
	procGetExtendedTcpTable = modiphlpapi.NewProc("GetExtendedTcpTable")
)

func GetTCP6ConnectionStates() (map[uint32]uint32, error) {
	return getTCPConnectionStates(AF_INET6, TCP6TableClass)
}

func GetTCPConnectionStates() (map[uint32]uint32, error) {
	return getTCPConnectionStates(AF_INET, TCPTableClass)
}

func getTCPConnectionStates(family, tableClass uint32) (map[uint32]uint32, error) {
	var size uint32
	stateCounts := make(map[uint32]uint32)

	ret := getExtendedTcpTable(0, &size, true, family, tableClass, 0)
	if ret != 0 && ret != uintptr(windows.ERROR_INSUFFICIENT_BUFFER) {
		return nil, fmt.Errorf("getExtendedTcpTable (size query) failed with code %d", ret)
	}

	buf := make([]byte, size)
	ret = getExtendedTcpTable(uintptr(unsafe.Pointer(&buf[0])), &size, true, family, tableClass, 0)
	if ret != 0 {
		return nil, fmt.Errorf("getExtendedTcpTable (data query) failed with code %d", ret)
	}

	numEntries := *(*uint32)(unsafe.Pointer(&buf[0]))
	rowSize := uint32(unsafe.Sizeof(MIB_TCP6ROW_OWNER_PID{}))
	if family == AF_INET {
		rowSize = uint32(unsafe.Sizeof(MIB_TCPROW_OWNER_PID{}))
	}

	for i := uint32(0); i < numEntries; i++ {
		var state uint32
		if family == AF_INET6 {
			row := (*MIB_TCP6ROW_OWNER_PID)(unsafe.Pointer(&buf[4+i*rowSize]))
			state = row.dwState
		} else {
			row := (*MIB_TCPROW_OWNER_PID)(unsafe.Pointer(&buf[4+i*rowSize]))
			state = row.dwState
		}
		stateCounts[state]++
	}

	return stateCounts, nil
}

func getExtendedTcpTable(pTCPTable uintptr, pdwSize *uint32, bOrder bool, ulAf uint32, tableClass uint32, reserved uint32) uintptr {
	ret, _, _ := procGetExtendedTcpTable.Call(
		pTCPTable,
		uintptr(unsafe.Pointer(pdwSize)),
		uintptr(boolToInt(bOrder)),
		uintptr(ulAf),
		uintptr(tableClass),
		uintptr(reserved),
	)

	return ret
}

func boolToInt(b bool) int {
	if b {
		return 1
	}

	return 0
}

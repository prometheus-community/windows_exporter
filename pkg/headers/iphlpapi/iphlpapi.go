package iphlpapi

import (
	"unsafe"

	"github.com/pkg/errors"
	"golang.org/x/sys/windows"
)

var (
	modiphlpapi             = windows.NewLazySystemDLL("iphlpapi.dll")
	procGetExtendedTcpTable = modiphlpapi.NewProc("GetExtendedTcpTable")
)

const TCPTableClass = 5

func GetTCPConnectionStates(family uint32) (map[uint32]uint32, error) {
	var size uint32
	stateCounts := make(map[uint32]uint32)

	ret := getExtendedTcpTable(0, &size, true, family, TCPTableClass, 0)
	if ret != 0 && ret != uintptr(windows.ERROR_INSUFFICIENT_BUFFER) {
		return nil, errors.Errorf("getExtendedTcpTable failed with code %d", ret)
	}

	buf := make([]byte, size)
	ret = getExtendedTcpTable(uintptr(unsafe.Pointer(&buf[0])), &size, true, family, TCPTableClass, 0)
	if ret != 0 {
		return nil, errors.Errorf("getExtendedTcpTable failed with code %d", ret)
	}

	numEntries := *(*uint32)(unsafe.Pointer(&buf[0]))
	for i := uint32(0); i < numEntries; i++ {
		row := (*MIB_TCPROW_OWNER_PID)(unsafe.Pointer(&buf[4+i*uint32(unsafe.Sizeof(MIB_TCPROW_OWNER_PID{}))]))
		stateCounts[row.dwState]++
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

package iphlpapi

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modiphlpapi = windows.NewLazySystemDLL("iphlpapi.dll")

	procGetIPForwardTable2 = modiphlpapi.NewProc("GetIpForwardTable2")
	procFreeMibTable       = modiphlpapi.NewProc("FreeMibTable")
)

// GetIpForwardTable2 function
// (https://docs.microsoft.com/en-us/windows/desktop/api/netioapi/nf-netioapi-getipforwardtable2).
func GetIpForwardTable2(family uint32) ([]*MIB_IPFORWARD_ROW2, error) {
	var pTable *MIB_IPFORWARD_TABLE2

	r1, _, err := procGetIPForwardTable2.Call(
		uintptr(family),
		uintptr(unsafe.Pointer(&pTable)),
	)

	if r1 != 1 {
		return nil, fmt.Errorf("GetIpForwardTable2: %w", err)
	}

	if pTable != nil {
		defer func() {
			_, _, _ = procFreeMibTable.Call(uintptr(unsafe.Pointer(pTable)))
		}()
	}

	rows := make([]*MIB_IPFORWARD_ROW2, pTable.NumEntries)

	pFirstRow := uintptr(unsafe.Pointer(&pTable.Table[0]))
	rowSize := unsafe.Sizeof(pTable.Table[0])

	for i := range pTable.NumEntries {
		row := *(*MIB_IPFORWARD_ROW2)(unsafe.Pointer(pFirstRow + rowSize*uintptr(i))) // Dereferencing and rereferencing in order to force copying.
		rows[i] = &row
	}

	return rows, nil
}

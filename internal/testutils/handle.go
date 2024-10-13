package testutils

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modkernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procGetProcessHandleCount = modkernel32.NewProc("GetProcessHandleCount")
)

func GetProcessHandleCount(handle windows.Handle) (uint32, error) {
	var count uint32

	r1, _, err := procGetProcessHandleCount.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(&count)),
	)

	if r1 != 1 {
		return 0, err
	}

	return count, nil
}

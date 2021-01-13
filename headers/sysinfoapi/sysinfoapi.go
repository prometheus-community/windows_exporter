package sysinfoapi

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// MemoryStatusEx is a wrapper for MEMORYSTATUSEX
// https://docs.microsoft.com/en-us/windows/win32/api/sysinfoapi/ns-sysinfoapi-memorystatusex
type MemoryStatusEx struct {
	dwLength                uint32
	DwMemoryLoad            uint32
	UllTotalPhys            uint64
	UllAvailPhys            uint64
	UllTotalPageFile        uint64
	UllAvailPageFile        uint64
	UllTotalVirtual         uint64
	UllAvailVirtual         uint64
	UllAvailExtendedVirtual uint64
}

var (
	kernel32                 = windows.NewLazySystemDLL("kernel32.dll")
	procGetSystemInfo        = kernel32.NewProc("GetSystemInfo")
	procGlobalMemoryStatusEx = kernel32.NewProc("GlobalMemoryStatusEx")
)

// GlobalMemoryStatusEx retrieves information about the system's current usage of both physical and virtual memory.
// https://docs.microsoft.com/en-us/windows/win32/api/sysinfoapi/nf-sysinfoapi-globalmemorystatusex
func GlobalMemoryStatusEx() (MemoryStatusEx, error) {
	var mse MemoryStatusEx
	mse.dwLength = (uint32)(unsafe.Sizeof(mse))
	pMse := uintptr(unsafe.Pointer(&mse))
	r1, _, err := procGlobalMemoryStatusEx.Call(pMse)

	if ret := *(*bool)(unsafe.Pointer(&r1)); ret == false {
		// returned false
		return MemoryStatusEx{}, err
	}

	return mse, nil
}

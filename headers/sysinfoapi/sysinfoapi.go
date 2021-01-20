package sysinfoapi

import (
	"unicode/utf16"
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

// wProcessorArchitecture is a wrapper for the union found in LP_SYSTEM_INFO
// https://docs.microsoft.com/en-us/windows/win32/api/sysinfoapi/ns-sysinfoapi-system_info
type wProcessorArchitecture struct {
	WReserved              uint16
	WProcessorArchitecture uint16
}

// LpSystemInfo is a wrapper for LPSYSTEM_INFO
// https://docs.microsoft.com/en-us/windows/win32/api/sysinfoapi/ns-sysinfoapi-system_info
type LpSystemInfo struct {
	Arch                        wProcessorArchitecture
	DwPageSize                  uint32
	LpMinimumApplicationAddress *byte
	LpMaximumApplicationAddress *byte
	DwActiveProcessorMask       *uint32
	DwNumberOfProcessors        uint32
	DwProcessorType             uint32
	DwAllocationGranularity     uint32
	WProcessorLevel             uint16
	WProcessorRevision          uint16
}

// WinComputerNameFormat is a wrapper for COMPUTER_NAME_FORMAT
type WinComputerNameFormat int

// Definitions for WinComputerNameFormat constants
const (
	ComputerNameNetBIOS WinComputerNameFormat = iota
	ComputerNameDNSHostname
	ComputerNameDNSDomain
	ComputerNameDNSFullyQualified
	ComputerNamePhysicalNetBIOS
	ComputerNamePhysicalDNSHostname
	ComputerNamePhysicalDNSDomain
	ComputerNamePhysicalDNSFullyQualified
	ComputerNameMax
)

var (
	kernel32                 = windows.NewLazySystemDLL("kernel32.dll")
	procGetSystemInfo        = kernel32.NewProc("GetSystemInfo")
	procGlobalMemoryStatusEx = kernel32.NewProc("GlobalMemoryStatusEx")
	procGetComputerNameExW   = kernel32.NewProc("GetComputerNameExW")
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

// GetSystemInfo wraps the GetSystemInfo function from sysinfoapi
// https://docs.microsoft.com/en-us/windows/win32/api/sysinfoapi/nf-sysinfoapi-getsysteminfo
func GetSystemInfo() LpSystemInfo {
	var info LpSystemInfo
	pInfo := uintptr(unsafe.Pointer(&info))
	procGetSystemInfo.Call(pInfo)
	return info
}

// GetComputerName wraps the GetComputerNameW function in a more Go-like way
// https://docs.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-getcomputernamew
func GetComputerName(f WinComputerNameFormat) (string, error) {
	// 1kb buffer to accept computer name. This should be more than enough as the maximum size
	// returned is the max length of a DNS name, which this author believes is 253 characters.
	size := 1024
	var buffer [4096]uint16
	r1, _, err := procGetComputerNameExW.Call(uintptr(f), uintptr(unsafe.Pointer(&buffer)), uintptr(unsafe.Pointer(&size)))
	if r1 == 0 {
		return "", err
	}
	bytes := buffer[0:size]
	out := utf16.Decode(bytes)
	return string(out), nil
}

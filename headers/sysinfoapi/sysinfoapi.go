package sysinfoapi

import (
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

// MemoryStatusEx is a wrapper for MEMORYSTATUSEX
// https://docs.microsoft.com/en-us/windows/win32/api/sysinfoapi/ns-sysinfoapi-memorystatusex
type memoryStatusEx struct {
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

// MemoryStatus is an idiomatic wrapper for MemoryStatusEx
type MemoryStatus struct {
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

// wProcessorArchitecture is a wrapper for the union found in LP_SYSTEM_INFO
// https://docs.microsoft.com/en-us/windows/win32/api/sysinfoapi/ns-sysinfoapi-system_info
type wProcessorArchitecture struct {
	WProcessorArchitecture uint16
	WReserved              uint16
}

// ProcessorArchitecture is an idiomatic wrapper for wProcessorArchitecture
type ProcessorArchitecture uint16

// Idiomatic values for wProcessorArchitecture
const (
	AMD64   ProcessorArchitecture = 9
	ARM                           = 5
	ARM64                         = 12
	IA64                          = 6
	INTEL                         = 0
	UNKNOWN                       = 0xffff
)

// LpSystemInfo is a wrapper for LPSYSTEM_INFO
// https://docs.microsoft.com/en-us/windows/win32/api/sysinfoapi/ns-sysinfoapi-system_info
type lpSystemInfo struct {
	Arch                        wProcessorArchitecture
	DwPageSize                  uint32
	LpMinimumApplicationAddress uintptr
	LpMaximumApplicationAddress uintptr
	DwActiveProcessorMask       uint
	DwNumberOfProcessors        uint32
	DwProcessorType             uint32
	DwAllocationGranularity     uint32
	WProcessorLevel             uint16
	WProcessorRevision          uint16
}

// SystemInfo is an idiomatic wrapper for LpSystemInfo
type SystemInfo struct {
	Arch                      ProcessorArchitecture
	PageSize                  uint32
	MinimumApplicationAddress uintptr
	MaximumApplicationAddress uintptr
	ActiveProcessorMask       uint
	NumberOfProcessors        uint32
	ProcessorType             uint32
	AllocationGranularity     uint32
	ProcessorLevel            uint16
	ProcessorRevision         uint16
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
func GlobalMemoryStatusEx() (MemoryStatus, error) {
	var mse memoryStatusEx
	mse.dwLength = (uint32)(unsafe.Sizeof(mse))
	r1, _, err := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&mse)))

	if ret := *(*bool)(unsafe.Pointer(&r1)); ret == false {
		return MemoryStatus{}, err
	}

	return MemoryStatus{
		MemoryLoad:           mse.DwMemoryLoad,
		TotalPhys:            mse.UllTotalPhys,
		AvailPhys:            mse.UllAvailPhys,
		TotalPageFile:        mse.UllTotalPageFile,
		AvailPageFile:        mse.UllAvailPageFile,
		TotalVirtual:         mse.UllTotalVirtual,
		AvailVirtual:         mse.UllAvailVirtual,
		AvailExtendedVirtual: mse.UllAvailExtendedVirtual,
	}, nil
}

// GetSystemInfo is an idiomatic wrapper for the GetSystemInfo function from sysinfoapi
// https://docs.microsoft.com/en-us/windows/win32/api/sysinfoapi/nf-sysinfoapi-getsysteminfo
func GetSystemInfo() SystemInfo {
	var info lpSystemInfo
	procGetSystemInfo.Call(uintptr(unsafe.Pointer(&info))) //nolint:errcheck
	return SystemInfo{
		Arch:                      ProcessorArchitecture(info.Arch.WProcessorArchitecture),
		PageSize:                  info.DwPageSize,
		MinimumApplicationAddress: info.LpMinimumApplicationAddress,
		MaximumApplicationAddress: info.LpMinimumApplicationAddress,
		ActiveProcessorMask:       info.DwActiveProcessorMask,
		NumberOfProcessors:        info.DwNumberOfProcessors,
		ProcessorType:             info.DwProcessorType,
		AllocationGranularity:     info.DwAllocationGranularity,
		ProcessorLevel:            info.WProcessorLevel,
		ProcessorRevision:         info.WProcessorRevision,
	}
}

// GetComputerName wraps the GetComputerNameW function in a more Go-like way
// https://docs.microsoft.com/en-us/windows/win32/api/sysinfoapi/nf-sysinfoapi-getcomputernameexw
func GetComputerName(f WinComputerNameFormat) (string, error) {
	// 1kb buffer to accept computer name. This should be more than enough as the maximum size
	// returned is the max length of a DNS name, which this author believes is 253 characters.
	size := 1024
	var buffer [1024]uint16
	r1, _, err := procGetComputerNameExW.Call(uintptr(f), uintptr(unsafe.Pointer(&buffer)), uintptr(unsafe.Pointer(&size)))
	if r1 == 0 {
		return "", err
	}
	bytes := buffer[0:size]
	out := utf16.Decode(bytes)
	return string(out), nil
}

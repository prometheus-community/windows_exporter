package psapi

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// PerformanceInformation is a wrapper of the PERFORMANCE_INFORMATION struct.
// https://docs.microsoft.com/en-us/windows/win32/api/psapi/ns-psapi-performance_information
type PerformanceInformation struct {
	cb                uint32
	CommitTotal       uint
	CommitLimit       uint
	CommitPeak        uint
	PhysicalTotal     uint
	PhysicalAvailable uint
	SystemCache       uint
	KernelTotal       uint
	KernelPaged       uint
	KernelNonpaged    uint
	PageSize          uint
	HandleCount       uint32
	ProcessCount      uint32
	ThreadCount       uint32
}

var (
	psapi                  = windows.NewLazySystemDLL("psapi.dll")
	procGetPerformanceInfo = psapi.NewProc("GetPerformanceInfo")
)

// GetPerformanceInfo returns the dereferenced version of GetLPPerformanceInfo.
func GetPerformanceInfo() (PerformanceInformation, error) {
	var lppi PerformanceInformation
	size := (uint32)(unsafe.Sizeof(lppi))
	lppi.cb = size
	r1, _, err := procGetPerformanceInfo.Call(uintptr(unsafe.Pointer(&lppi)), uintptr(size))

	if ret := *(*bool)(unsafe.Pointer(&r1)); !ret {
		return PerformanceInformation{}, err
	}

	return lppi, nil
}

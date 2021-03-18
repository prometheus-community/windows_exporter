package psapi

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// LPPerformanceInformation is a wrapper of the PERFORMANCE_INFORMATION struct.
// https://docs.microsoft.com/en-us/windows/win32/api/psapi/ns-psapi-performance_information
type LPPerformanceInformation struct {
	cb                uint32
	CommitTotal       *uint32
	CommitLimit       *uint32
	CommitPeak        *uint32
	PhysicalTotal     *uint32
	PhysicalAvailable *uint32
	SystemCache       *uint32
	KernelTotal       *uint32
	KernelPaged       *uint32
	KernelNonpaged    *uint32
	PageSize          *uint32
	HandleCount       uint32
	ProcessCount      uint32
	ThreadCount       uint32
}

// PerformanceInformation is a dereferenced version of LPPerformanceInformation
type PerformanceInformation struct {
	cb                uint32
	CommitTotal       uint32
	CommitLimit       uint32
	CommitPeak        uint32
	PhysicalTotal     uint32
	PhysicalAvailable uint32
	SystemCache       uint32
	KernelTotal       uint32
	KernelPaged       uint32
	KernelNonpaged    uint32
	PageSize          uint32
	HandleCount       uint32
	ProcessCount      uint32
	ThreadCount       uint32
}

var (
	psapi                  = windows.NewLazySystemDLL("psapi.dll")
	procGetPerformanceInfo = psapi.NewProc("GetPerformanceInfo")
)

// GetLPPerformanceInfo retrieves the performance values contained in the LPPerformanceInformation structure.
// https://docs.microsoft.com/en-us/windows/win32/api/psapi/nf-psapi-getperformanceinfo
func GetLPPerformanceInfo() (LPPerformanceInformation, error) {
	var pi LPPerformanceInformation
	size := (uint32)(unsafe.Sizeof(pi))
	pi.cb = size
	r1, _, err := procGetPerformanceInfo.Call(uintptr(unsafe.Pointer(&pi)), uintptr(size))

	if ret := *(*bool)(unsafe.Pointer(&r1)); !ret {
		return LPPerformanceInformation{}, err
	}

	return pi, nil
}

// GetPerformanceInfo returns the dereferenced version of GetLPPerformanceInfo.
func GetPerformanceInfo() (PerformanceInformation, error) {
	var lppi LPPerformanceInformation
	size := (uint32)(unsafe.Sizeof(lppi))
	lppi.cb = size
	r1, _, err := procGetPerformanceInfo.Call(uintptr(unsafe.Pointer(&lppi)), uintptr(size))

	if ret := *(*bool)(unsafe.Pointer(&r1)); !ret {
		return PerformanceInformation{}, err
	}

	var pi PerformanceInformation
	pi.cb = lppi.cb
	pi.CommitTotal = *(lppi.CommitTotal)
	pi.CommitLimit = *(lppi.CommitLimit)
	pi.CommitPeak = *(lppi.CommitPeak)
	pi.PhysicalTotal = *(lppi.PhysicalTotal)
	pi.PhysicalAvailable = *(lppi.PhysicalAvailable)
	pi.SystemCache = *(lppi.SystemCache)
	pi.KernelTotal = *(lppi.KernelTotal)
	pi.KernelPaged = *(lppi.KernelPaged)
	pi.KernelNonpaged = *(lppi.KernelNonpaged)
	pi.PageSize = *(lppi.PageSize)
	pi.HandleCount = lppi.HandleCount
	pi.ProcessCount = lppi.ProcessCount
	pi.ThreadCount = lppi.ThreadCount

	return pi, nil
}

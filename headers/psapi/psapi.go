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
	pPi := uintptr(unsafe.Pointer(&pi))
	r1, _, err := procGetPerformanceInfo.Call(pPi, uintptr(size))

	if ret := *(*bool)(unsafe.Pointer(&r1)); ret == false {
		// returned false
		return LPPerformanceInformation{}, err
	}

	return pi, nil
}

// GetPerformanceInfo returns the dereferenced version of GetLPPerformanceInfo.
func GetPerformanceInfo() (PerformanceInformation, error) {
	var lppi LPPerformanceInformation
	size := (uint32)(unsafe.Sizeof(lppi))
	lppi.cb = size
	pLppi := uintptr(unsafe.Pointer(&lppi))
	r1, _, err := procGetPerformanceInfo.Call(pLppi, uintptr(size))

	if ret := *(*bool)(unsafe.Pointer(&r1)); ret == false {
		// returned false
		return PerformanceInformation{}, err
	}

	var pi PerformanceInformation
	pi.cb = lppi.cb
	pi.CommitTotal = *(*uint32)(unsafe.Pointer(&lppi.CommitTotal))
	pi.CommitLimit = *(*uint32)(unsafe.Pointer(&lppi.CommitLimit))
	pi.CommitPeak = *(*uint32)(unsafe.Pointer(&lppi.CommitPeak))
	pi.PhysicalTotal = *(*uint32)(unsafe.Pointer(&lppi.PhysicalTotal))
	pi.PhysicalAvailable = *(*uint32)(unsafe.Pointer(&lppi.PhysicalAvailable))
	pi.SystemCache = *(*uint32)(unsafe.Pointer(&lppi.SystemCache))
	pi.KernelTotal = *(*uint32)(unsafe.Pointer(&lppi.KernelTotal))
	pi.KernelPaged = *(*uint32)(unsafe.Pointer(&lppi.KernelPaged))
	pi.KernelNonpaged = *(*uint32)(unsafe.Pointer(&lppi.KernelNonpaged))
	pi.PageSize = *(*uint32)(unsafe.Pointer(&lppi.PageSize))
	pi.HandleCount = lppi.HandleCount
	pi.ProcessCount = lppi.ProcessCount
	pi.ThreadCount = lppi.ThreadCount

	return pi, nil
}

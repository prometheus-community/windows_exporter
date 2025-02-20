package kernel32

import "unsafe"

type JobObjectBasicAccountingInformation struct {
	TotalUserTime             uint64
	TotalKernelTime           uint64
	ThisPeriodTotalUserTime   uint64
	ThisPeriodTotalKernelTime uint64
	TotalPageFaultCount       uint32
	TotalProcesses            uint32
	ActiveProcesses           uint32
	TotalTerminatedProcesses  uint32
}

type IOCounters struct {
	ReadOperationCount  uint64
	WriteOperationCount uint64
	OtherOperationCount uint64
	ReadTransferCount   uint64
	WriteTransferCount  uint64
	OtherTransferCount  uint64
}

type JobObjectExtendedLimitInformation struct {
	BasicInfo             JobObjectBasicAccountingInformation
	IoInfo                IOCounters
	ProcessMemoryLimit    uint64
	JobMemoryLimit        uint64
	PeakProcessMemoryUsed uint64
	PeakJobMemoryUsed     uint64
}

type JobObjectBasicProcessIDList struct {
	NumberOfAssignedProcesses uint32
	NumberOfProcessIdsInList  uint32
	ProcessIdList             [1]uintptr
}

// PIDs returns all the process Ids in the job object.
func (p *JobObjectBasicProcessIDList) PIDs() []uint32 {
	return unsafe.Slice((*uint32)(unsafe.Pointer(&p.ProcessIdList[0])), int(p.NumberOfProcessIdsInList))
}

type PROCESS_VM_COUNTERS struct {
	PeakVirtualSize            uintptr
	VirtualSize                uintptr
	PageFaultCount             uint32
	PeakWorkingSetSize         uintptr
	WorkingSetSize             uintptr
	QuotaPeakPagedPoolUsage    uintptr
	QuotaPagedPoolUsage        uintptr
	QuotaPeakNonPagedPoolUsage uintptr
	QuotaNonPagedPoolUsage     uintptr
	PagefileUsage              uintptr
	PeakPagefileUsage          uintptr
	PrivateWorkingSetSize      uintptr
}

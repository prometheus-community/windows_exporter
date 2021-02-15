// +build windows

package collector

/*
#include "nvml.h"
#include <windows.h>

// definition of required import symbols
nvmlReturn_t DECLDIR (*nvmlInit_v2Ptr)() = NULL;
const DECLDIR char* (*nvmlErrorStringPtr)(nvmlReturn_t result) = NULL;
nvmlReturn_t DECLDIR (*nvmlDeviceGetCount_v2Ptr)(unsigned int *deviceCount) = NULL;
nvmlReturn_t DECLDIR (*nvmlDeviceGetHandleByIndex_v2Ptr)(unsigned int index, nvmlDevice_t *device);
nvmlReturn_t DECLDIR (*nvmlDeviceGetNamePtr)(nvmlDevice_t device, char *name, unsigned int length);
nvmlReturn_t DECLDIR (*nvmlDeviceGetGraphicsRunningProcesses_v2Ptr)(nvmlDevice_t device, unsigned int *infoCount, nvmlProcessInfo_t *infos);
nvmlReturn_t DECLDIR (*nvmlSystemGetProcessNamePtr)(unsigned int pid, char *name, unsigned int length);

nvmlReturn_t invoke_nvmlInit_v2()
{
	return (*nvmlInit_v2Ptr)();
}

const char *invoke_nvmlErrorString(nvmlReturn_t error)
{
	return (*nvmlErrorStringPtr)(error);
}

nvmlReturn_t invoke_nvmlDeviceGetCount_v2(unsigned int *deviceCount)
{
	return (*nvmlDeviceGetCount_v2Ptr)(deviceCount);
}

nvmlReturn_t invoke_nvmlDeviceGetHandleByIndex_v2(unsigned int index, nvmlDevice_t *device)
{
	return (*nvmlDeviceGetHandleByIndex_v2Ptr)(index, device);
}

nvmlReturn_t invoke_nvmlDeviceGetName(nvmlDevice_t device, char *name, unsigned int length)
{
	return (*nvmlDeviceGetNamePtr)(device, name, length);
}

nvmlReturn_t invoke_nvmlDeviceGetGraphicsRunningProcesses_v2(nvmlDevice_t device, unsigned int *infoCount, nvmlProcessInfo_t *infos)
{
	return (*nvmlDeviceGetGraphicsRunningProcesses_v2Ptr)(device, infoCount, infos);
}

nvmlReturn_t invoke_nvmlSystemGetProcessName(unsigned int pid, char *name, unsigned int length)
{
	return (*nvmlSystemGetProcessNamePtr)(pid, name, length);
}
*/
import "C"
import (
	"fmt"
	"regexp"
	"strconv"
	"syscall"
	"unsafe"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"gopkg.in/alecthomas/kingpin.v2"
)

func init() {
	registerCollector("nvidia", newNvidiaCollector, "NVIDIA")
}

var (
	gpuProcessWhitelist = kingpin.Flag(
		"collector.nvidia.process_whitelist",
		"Regexp of processes to include. Process name must both match whitelist and not match blacklist to be included.",
	).Default(".*").String()
	gpuProcessBlacklist = kingpin.Flag(
		"collector.nvidia.process_blacklist",
		"Regexp of processes to exclude. Process name must both match whitelist and not match blacklist to be included.",
	).Default("").String()
	nvidiaGpuWhitelist = kingpin.Flag(
		"collector.nvidia.gpu_whitelist",
		"Regexp of NVIDIA GPUs to include in per process metrics.",
	).Default(".*").String()
)

var nvmlValueNotAvailable = C.longlong(C.NVML_VALUE_NOT_AVAILABLE)

type nvidiaCollector struct {
	TotalGpuMemoryUsed *prometheus.Desc

	gpuProcessWhitelistPattern *regexp.Regexp
	gpuProcessBlacklistPattern *regexp.Regexp

	nvmlLibrary C.HINSTANCE

	devices []C.nvmlDevice_t
}

func (error C.nvmlReturn_t) Error() string {
	return C.GoString(C.invoke_nvmlErrorString(error))
}

func clen(bytes []byte) int {
	for i := 0; i < len(bytes); i++ {
		if bytes[i] == 0 {
			return i
		}
	}
	return 0
}

// newNvidiaCollector ...
func newNvidiaCollector() (Collector, error) {
	nvmlLibrary := C.LoadLibrary(C.CString("nvml.dll"))
	if nvmlLibrary == C.HINSTANCE(C.NULL) {
		return nil, syscall.Errno(C.GetLastError())
	}

	C.nvmlInit_v2Ptr = C.GetProcAddress(nvmlLibrary, C.CString("nvmlInit_v2"))
	if C.nvmlInit_v2Ptr == C.FARPROC(C.NULL) {
		return nil, syscall.Errno(C.GetLastError())
	}

	C.nvmlErrorStringPtr = C.GetProcAddress(nvmlLibrary, C.CString("nvmlErrorString"))
	if C.nvmlErrorStringPtr == C.FARPROC(C.NULL) {
		return nil, syscall.Errno(C.GetLastError())
	}

	C.nvmlDeviceGetCount_v2Ptr = C.GetProcAddress(nvmlLibrary, C.CString("nvmlDeviceGetCount_v2"))
	if C.nvmlDeviceGetCount_v2Ptr == C.FARPROC(C.NULL) {
		return nil, syscall.Errno(C.GetLastError())
	}

	C.nvmlDeviceGetHandleByIndex_v2Ptr = C.GetProcAddress(nvmlLibrary, C.CString("nvmlDeviceGetHandleByIndex_v2"))
	if C.nvmlDeviceGetHandleByIndex_v2Ptr == C.FARPROC(C.NULL) {
		return nil, syscall.Errno(C.GetLastError())
	}

	C.nvmlDeviceGetNamePtr = C.GetProcAddress(nvmlLibrary, C.CString("nvmlDeviceGetName"))
	if C.nvmlDeviceGetNamePtr == C.FARPROC(C.NULL) {
		return nil, syscall.Errno(C.GetLastError())
	}

	C.nvmlDeviceGetGraphicsRunningProcesses_v2Ptr = C.GetProcAddress(nvmlLibrary, C.CString("nvmlDeviceGetGraphicsRunningProcesses_v2"))
	if C.nvmlDeviceGetGraphicsRunningProcesses_v2Ptr == C.FARPROC(C.NULL) {
		return nil, syscall.Errno(C.GetLastError())
	}

	C.nvmlSystemGetProcessNamePtr = C.GetProcAddress(nvmlLibrary, C.CString("nvmlSystemGetProcessName"))
	if C.nvmlSystemGetProcessNamePtr == C.FARPROC(C.NULL) {
		return nil, syscall.Errno(C.GetLastError())
	}

	err := C.invoke_nvmlInit_v2()
	if err != C.NVML_SUCCESS {
		return nil, err
	}

	const subsystem = "nvidia"

	if *processWhitelist == ".*" && *processBlacklist == "" {
		log.Warn("No filters specified for nvidia collector. This will generate a very large number of metrics!")
	}

	var devices []C.nvmlDevice_t
	if len(*nvidiaGpuWhitelist) > 0 {
		var count C.uint = 0
		err := C.invoke_nvmlDeviceGetCount_v2(&count)
		if err != C.NVML_SUCCESS {
			return nil, err
		}
		var nvidiaGpuWhitelistPatttern = regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *nvidiaGpuWhitelist))
		for i := C.uint(0); i < count; i++ {
			var device C.nvmlDevice_t
			err := C.invoke_nvmlDeviceGetHandleByIndex_v2(i, &device)
			if err != C.NVML_SUCCESS {
				return nil, err
			}
			nameBuffer := make([]byte, C.NVML_DEVICE_NAME_BUFFER_SIZE)
			err = C.invoke_nvmlDeviceGetName(device, (*C.char)(unsafe.Pointer(&nameBuffer[0])), C.NVML_DEVICE_NAME_BUFFER_SIZE)
			if err != C.NVML_SUCCESS {
				return nil, err
			}
			name := string(nameBuffer[:clen(nameBuffer)])
			if nvidiaGpuWhitelistPatttern.MatchString(name) {
				devices = append(devices, device)
			}
		}
	}

	return &nvidiaCollector{
		TotalGpuMemoryUsed: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "total_gpu_memory_used"),
			"Total gpu memory used by this process.",
			[]string{"process", "process_id"},
			nil,
		),
		gpuProcessWhitelistPattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *gpuProcessWhitelist)),
		gpuProcessBlacklistPattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *gpuProcessBlacklist)),

		devices: devices,
	}, nil
}

func (c *nvidiaCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {

	var pidMems map[uint]uint64
	for _, device := range c.devices {
		var infoCount C.uint = 0
		var infos = make([]C.nvmlProcessInfo_t, infoCount)
		err := C.invoke_nvmlDeviceGetGraphicsRunningProcesses_v2(device, &infoCount, (*C.nvmlProcessInfo_t)(C.NULL))
		if err == C.NVML_ERROR_INSUFFICIENT_SIZE {
			infos = make([]C.nvmlProcessInfo_t, infoCount)
			err = C.invoke_nvmlDeviceGetGraphicsRunningProcesses_v2(device, &infoCount, &infos[0])
		}
		if err != C.NVML_SUCCESS {
			return nil
		}
		for _, info := range infos {
			if info.usedGpuMemory == C.ulonglong(nvmlValueNotAvailable) {
				continue
			}
			var pid uint = uint(info.pid)
			_, present := pidMems[pid]
			if present == false {
				pidMems[pid] = uint64(info.usedGpuMemory)
			} else {
				pidMems[pid] += uint64(info.usedGpuMemory)
			}
		}
	}

	for pid, memory := range pidMems {
		var nameBuffer = make(byte[], C.MAX_PATH)
		err := C.invoke_nvmlSystemGetProcessName(pid, (*C.char)(unsafe.Pointer(&nameBuffer[0])), C.MAX_PATH)
		if err != C.NVML_SUCCESS {
			return nil
		}
		name := string(nameBuffer[:clen(nameBuffer)])
		ch <- prometheus.MustNewConstMetric(
			c.TotalGpuMemoryUsed,
			prometheus.GaugeValue,
			float64(memory),
			name,
			strconv.Itoa(int(pid)),
		)
	}

	return nil
}

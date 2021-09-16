//go:build windows
// +build windows

package collector

/*
#cgo LDFLAGS: -lDXGI
#include <windows.h>
#include <dxgi.h>

#include <initguid.h>
DEFINE_GUID(GUID_IDXGI_FACTORY, 0x7b7166ec, 0x21c7, 0x44ae, 0xb2, 0x1a, 0xc9, 0xae, 0x32, 0x1a, 0xe3, 0x69);

struct DxgiAdapterDescription
{
	wchar_t description[128];
	LUID 	luid;
};

UINT GetDxgiAdapterCount()
{
	UINT dxgi_adapter_count = 0;
	IDXGIFactory *dxgi_factory = NULL;
	if (CreateDXGIFactory(&GUID_IDXGI_FACTORY, (void**)&dxgi_factory) == S_OK && dxgi_factory != NULL)
	{
		IDXGIAdapter *dxgi_adapter = NULL;
		while ((*dxgi_factory->lpVtbl->EnumAdapters)(dxgi_factory, dxgi_adapter_count, &dxgi_adapter) == S_OK)
		{
			dxgi_adapter_count++;
			if (dxgi_adapter != NULL)
			{
				(*dxgi_adapter->lpVtbl->Release)(dxgi_adapter);
			}
		}
		(*dxgi_factory->lpVtbl->Release)(dxgi_factory);
	}
	return dxgi_adapter_count;
}

UINT GetDxgiAdapterDescriptions(struct DxgiAdapterDescription *dxgi_adapter_descriptions, UINT dxgi_adapter_description_count)
{
	IDXGIFactory *dxgi_factory = NULL;
	struct DxgiAdapterDescription *current_dxgi_adapter_description = dxgi_adapter_descriptions;
	if (CreateDXGIFactory(&GUID_IDXGI_FACTORY, (void**)&dxgi_factory) == S_OK && dxgi_factory != NULL)
	{
		UINT dxgi_adapter_index = 0;
		IDXGIAdapter *dxgi_adapter = NULL;
		while (dxgi_adapter_description_count && (*dxgi_factory->lpVtbl->EnumAdapters)(dxgi_factory, dxgi_adapter_index, &dxgi_adapter) == S_OK)
		{
			dxgi_adapter_index++;
			if (dxgi_adapter != NULL)
			{
				DXGI_ADAPTER_DESC dxgi_adapter_description;
				if ((*dxgi_adapter->lpVtbl->GetDesc)(dxgi_adapter, &dxgi_adapter_description) == S_OK)
				{
					memcpy(current_dxgi_adapter_description->description, dxgi_adapter_description.Description, sizeof(current_dxgi_adapter_description->description));
					current_dxgi_adapter_description->luid = dxgi_adapter_description.AdapterLuid;
					++current_dxgi_adapter_description;
					--dxgi_adapter_description_count;
				}
				(*dxgi_adapter->lpVtbl->Release)(dxgi_adapter);
			}
		}
		(*dxgi_factory->lpVtbl->Release)(dxgi_factory);
		return dxgi_adapter_index;
	}
	return current_dxgi_adapter_description - dxgi_adapter_descriptions;
}
*/
import "C"
import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/StackExchange/wmi"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/alecthomas/kingpin.v2"
)

func init() {
	registerCollector("process", newProcessCollector, "Process", "GPU Process Memory", "GPU Engine", "GPU Adapter Memory")
}

var (
	processWhitelist = kingpin.Flag(
		"collector.process.whitelist",
		"Regexp of processes to include. Process name must both match whitelist and not match blacklist to be included.",
	).Default(".*").String()
	processBlacklist = kingpin.Flag(
		"collector.process.blacklist",
		"Regexp of processes to exclude. Process name must both match whitelist and not match blacklist to be included.",
	).Default("").String()
)

type processCollector struct {
	StartTime          *prometheus.Desc
	CPUTimeTotal       *prometheus.Desc
	HandleCount        *prometheus.Desc
	IOBytesTotal       *prometheus.Desc
	IOOperationsTotal  *prometheus.Desc
	PageFaultsTotal    *prometheus.Desc
	PageFileBytes      *prometheus.Desc
	PoolBytes          *prometheus.Desc
	PriorityBase       *prometheus.Desc
	PrivateBytes       *prometheus.Desc
	ThreadCount        *prometheus.Desc
	VirtualBytes      *prometheus.Desc
	WorkingSetPrivate *prometheus.Desc
	WorkingSetPeak    *prometheus.Desc
	WorkingSet        *prometheus.Desc
	GpuSharedMemory    *prometheus.Desc
	GpuDedicatedMemory *prometheus.Desc

	processWhitelistPattern *regexp.Regexp
	processBlacklistPattern *regexp.Regexp

	dxgiAdapterLuidDescriptionMap map[string]string
}

// https://docs.microsoft.com/en-us/windows/win32/api/dxgi/ns-dxgi-dxgi_adapter_desc
type dxgiAdapterDescription struct {
	description  [128]C.wchar_t
	luidLowPart  C.DWORD
	luidHighPart C.LONG
}

// NewProcessCollector ...
func newProcessCollector() (Collector, error) {
	const subsystem = "process"

	dxgiAdapterCount := C.GetDxgiAdapterCount()
	var dxgiAdapterDescriptions []C.struct_DxgiAdapterDescription
	dxgiAdapterLuidDescriptionMap := make(map[string]string)
	if dxgiAdapterCount > 0 {
		dxgiAdapterDescriptions = make([]C.struct_DxgiAdapterDescription, dxgiAdapterCount)
		dxgiAdapterDescriptionCount := C.GetDxgiAdapterDescriptions(&dxgiAdapterDescriptions[0], dxgiAdapterCount)
		for dxgiAdapterDescriptionIndex := C.UINT(0); dxgiAdapterDescriptionIndex < dxgiAdapterDescriptionCount; dxgiAdapterDescriptionIndex++ {
			description := syscall.UTF16ToString((*[128]uint16)(unsafe.Pointer(&dxgiAdapterDescriptions[dxgiAdapterDescriptionIndex].description))[:])
			luid := fmt.Sprintf("0x%08X_0x%08X", dxgiAdapterDescriptions[dxgiAdapterDescriptionIndex].luid.HighPart, dxgiAdapterDescriptions[dxgiAdapterDescriptionIndex].luid.LowPart)
			dxgiAdapterLuidDescriptionMap[luid] = description
		}
	}

	if *processWhitelist == ".*" && *processBlacklist == "" {
		log.Warn("No filters specified for process collector. This will generate a very large number of metrics!")
	}

	return &processCollector{
		StartTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "start_time"),
			"Time of process start.",
			[]string{"process", "process_id", "creating_process_id"},
			nil,
		),
		CPUTimeTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cpu_time_total"),
			"Returns elapsed time that all of the threads of this process used the processor to execute instructions by mode (privileged, user).",
			[]string{"process", "process_id", "creating_process_id", "mode"},
			nil,
		),
		HandleCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "handles"),
			"Total number of handles the process has open. This number is the sum of the handles currently open by each thread in the process.",
			[]string{"process", "process_id", "creating_process_id"},
			nil,
		),
		IOBytesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "io_bytes_total"),
			"Bytes issued to I/O operations in different modes (read, write, other).",
			[]string{"process", "process_id", "creating_process_id", "mode"},
			nil,
		),
		IOOperationsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "io_operations_total"),
			"I/O operations issued in different modes (read, write, other).",
			[]string{"process", "process_id", "creating_process_id", "mode"},
			nil,
		),
		PageFaultsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "page_faults_total"),
			"Page faults by the threads executing in this process.",
			[]string{"process", "process_id", "creating_process_id"},
			nil,
		),
		PageFileBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "page_file_bytes"),
			"Current number of bytes this process has used in the paging file(s).",
			[]string{"process", "process_id", "creating_process_id"},
			nil,
		),
		PoolBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pool_bytes"),
			"Pool Bytes is the last observed number of bytes in the paged or nonpaged pool.",
			[]string{"process", "process_id", "creating_process_id", "pool"},
			nil,
		),
		PriorityBase: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "priority_base"),
			"Current base priority of this process. Threads within a process can raise and lower their own base priority relative to the process base priority of the process.",
			[]string{"process", "process_id", "creating_process_id"},
			nil,
		),
		PrivateBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "private_bytes"),
			"Current number of bytes this process has allocated that cannot be shared with other processes.",
			[]string{"process", "process_id", "creating_process_id"},
			nil,
		),
		ThreadCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "threads"),
			"Number of threads currently active in this process.",
			[]string{"process", "process_id", "creating_process_id"},
			nil,
		),
		VirtualBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "virtual_bytes"),
			"Current size, in bytes, of the virtual address space that the process is using.",
			[]string{"process", "process_id", "creating_process_id"},
			nil,
		),
		WorkingSetPrivate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "working_set_private_bytes"),
			"Size of the working set, in bytes, that is use for this process only and not shared nor shareable by other processes.",
			[]string{"process", "process_id", "creating_process_id"},
			nil,
		),
		WorkingSetPeak: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "working_set_peak_bytes"),
			"Maximum size, in bytes, of the Working Set of this process at any point in time. The Working Set is the set of memory pages touched recently by the threads in the process.",
			[]string{"process", "process_id", "creating_process_id"},
			nil,
		),
		WorkingSet: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "working_set_bytes"),
			"Maximum number of bytes in the working set of this process at any point in time. The working set is the set of memory pages touched recently by the threads in the process.",
			[]string{"process", "process_id", "creating_process_id"},
			nil,
		),
		GpuSharedMemory: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "gpu_shared_memory"),
			"Number of bytes of system memory used as shared memory by the gpu.",
			[]string{"process", "process_id", "creating_process_id", "gpu_title"},
			nil,
		),
		GpuDedicatedMemory: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "gpu_dedicated_memory"),
			"Number of bytes of dedicated gpu memory used on the gpu.",
			[]string{"process", "process_id", "creating_process_id", "gpu_title"},
			nil,
		),
		processWhitelistPattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *processWhitelist)),
		processBlacklistPattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *processBlacklist)),

		dxgiAdapterLuidDescriptionMap: dxgiAdapterLuidDescriptionMap,
	}, nil
}

type perflibProcess struct {
	Name                    string
	PercentProcessorTime    float64 `perflib:"% Processor Time"`
	PercentPrivilegedTime   float64 `perflib:"% Privileged Time"`
	PercentUserTime         float64 `perflib:"% User Time"`
	CreatingProcessID       float64 `perflib:"Creating Process ID"`
	ElapsedTime             float64 `perflib:"Elapsed Time"`
	HandleCount             float64 `perflib:"Handle Count"`
	IDProcess               float64 `perflib:"ID Process"`
	IODataBytesPerSec       float64 `perflib:"IO Data Bytes/sec"`
	IODataOperationsPerSec  float64 `perflib:"IO Data Operations/sec"`
	IOOtherBytesPerSec      float64 `perflib:"IO Other Bytes/sec"`
	IOOtherOperationsPerSec float64 `perflib:"IO Other Operations/sec"`
	IOReadBytesPerSec       float64 `perflib:"IO Read Bytes/sec"`
	IOReadOperationsPerSec  float64 `perflib:"IO Read Operations/sec"`
	IOWriteBytesPerSec      float64 `perflib:"IO Write Bytes/sec"`
	IOWriteOperationsPerSec float64 `perflib:"IO Write Operations/sec"`
	PageFaultsPerSec        float64 `perflib:"Page Faults/sec"`
	PageFileBytesPeak       float64 `perflib:"Page File Bytes Peak"`
	PageFileBytes           float64 `perflib:"Page File Bytes"`
	PoolNonpagedBytes       float64 `perflib:"Pool Nonpaged Bytes"`
	PoolPagedBytes          float64 `perflib:"Pool Paged Bytes"`
	PriorityBase            float64 `perflib:"Priority Base"`
	PrivateBytes            float64 `perflib:"Private Bytes"`
	ThreadCount             float64 `perflib:"Thread Count"`
	VirtualBytesPeak        float64 `perflib:"Virtual Bytes Peak"`
	VirtualBytes            float64 `perflib:"Virtual Bytes"`
	WorkingSetPrivate       float64 `perflib:"Working Set - Private"`
	WorkingSetPeak          float64 `perflib:"Working Set Peak"`
	WorkingSet              float64 `perflib:"Working Set"`
}

type perflibProcessGpuMemory struct {
	Name           string
	DedicatedUsage float64 `perflib:"Dedicated Usage"`
	SharedUsage    float64 `perflib:"Shared Usage"`
	TotalCommitted float64 `perflib:"Total Committed"`
}

type perflibProcessGpuEngine struct {
	Name                  string
	UtilizationPercentage float64 `perflib:"Utilization Percentage"`
}

type processGpuMetrics struct {
	DedicatedUsage        map[string]float64
	SharedUsage           map[string]float64
	UtilizationPercentage map[string]map[string]float64
}

func createProcessGpuMetrics() *processGpuMetrics {
	return &processGpuMetrics{
		DedicatedUsage:        make(map[string]float64),
		SharedUsage:           make(map[string]float64),
		UtilizationPercentage: make(map[string]map[string]float64),
	}
}

func (processGpuMetrics *processGpuMetrics) setProcessGpuMemory(processGpuMemory perflibProcessGpuMemory, gpuTitle string) {
	processGpuMetrics.DedicatedUsage[gpuTitle] = processGpuMemory.DedicatedUsage
	processGpuMetrics.SharedUsage[gpuTitle] = processGpuMemory.SharedUsage
}

func (processGpuMetrics *processGpuMetrics) setProcessGpuEngineUsage(processGpuEngine perflibProcessGpuEngine, gpuTitle string) {
	_, present := processGpuMetrics.UtilizationPercentage[gpuTitle]
	if present == false {
		processGpuMetrics.UtilizationPercentage[gpuTitle] = make(map[string]float64)
	}
	videoEngineType := extractVideoEngineType(processGpuEngine.Name)
	if videoEngineType != nil {
		processGpuMetrics.UtilizationPercentage[gpuTitle][*videoEngineType] = processGpuEngine.UtilizationPercentage
	}
}

func printGpuEngineUsage(processGpuMetrics *processGpuMetrics) {
	for gpuTitle, gpuEngineUsageEntry := range processGpuMetrics.UtilizationPercentage {
		for gpuEngineTitle, gpuEngineUsage := range gpuEngineUsageEntry {
			log.Debugf("Engine %s utilization on %s: %f\n", gpuEngineTitle, gpuTitle, gpuEngineUsage)
		}
	}
}

func (processGpuMetrics *processGpuMetrics) exposeMetrics(ch chan<- prometheus.Metric, c *processCollector, processName string, pid string, cpid string) {
	// this may come in handy when someone wants to deal with the usage and different engine types
	// printGpuEngineUsage(processGpuMetrics)
	for gpuTitle, sharedUsage := range processGpuMetrics.SharedUsage {
		ch <- prometheus.MustNewConstMetric(
			c.GpuSharedMemory,
			prometheus.GaugeValue,
			sharedUsage,
			processName,
			pid,
			cpid,
			gpuTitle,
		)
	}

	for gpuTitle, dedicatedUsage := range processGpuMetrics.DedicatedUsage {
		ch <- prometheus.MustNewConstMetric(
			c.GpuDedicatedMemory,
			prometheus.GaugeValue,
			dedicatedUsage,
			processName,
			pid,
			cpid,
			gpuTitle,
		)
	}
}

type WorkerProcess struct {
	AppPoolName string
	ProcessId   uint64
}

var pidLuidRegexp = regexp.MustCompile("pid_([0-9]+)_luid_(0x[0-9a-zA-Z]{8}_0x[0-9a-zA-Z]{8})")

func extractPidAndLuid(name string) (string, string) {
	match := pidLuidRegexp.FindStringSubmatch(name)
	return match[1], match[2]
}

var videoEngineTypeRegexp = regexp.MustCompile("engtype_(.+)")

func extractVideoEngineType(name string) *string {
	match := videoEngineTypeRegexp.FindStringSubmatch(name)
	if match != nil {
		return &match[1]
	}
	return nil
}

func (c *processCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	data := make([]perflibProcess, 0)
	err := unmarshalObject(ctx.perfObjects["Process"], &data)
	if err != nil {
		return err
	}

	// dont check for error when unmarshalling "GPU Process Memory" and "GPU Engine" counters since they may not be present if there are no WDM compatible GPUs
	processGpuMemory := make([]perflibProcessGpuMemory, 0)
	unmarshalObject(ctx.perfObjects["GPU Process Memory"], &processGpuMemory)

	gpuEngine := make([]perflibProcessGpuEngine, 0)
	unmarshalObject(ctx.perfObjects["GPU Engine"], &gpuEngine)

	for _, processGpuMemoryEntry := range processGpuMemory {
		pid, luid := extractPidAndLuid(processGpuMemoryEntry.Name)
		_, dxgiAdapterPresent := c.dxgiAdapterLuidDescriptionMap[luid]
		if dxgiAdapterPresent {
			_, pidPresent := processGpuMetrics[pid]
			if pidPresent == false {
				processGpuMetrics[pid] = createProcessGpuMetrics()
			}
			processGpuMetrics[pid].setProcessGpuMemory(processGpuMemoryEntry, c.dxgiAdapterLuidDescriptionMap[luid])
		}

	}
	for _, processGpuEngineEntry := range gpuEngine {
		pid, luid := extractPidAndLuid(processGpuEngineEntry.Name)
		_, dxgiAdapterPresent := c.dxgiAdapterLuidDescriptionMap[luid]
		if dxgiAdapterPresent {
			_, pidPresent := processGpuMetrics[pid]
			if pidPresent == false {
				processGpuMetrics[pid] = createProcessGpuMetrics()
			}
			processGpuMetrics[pid].setProcessGpuEngineUsage(processGpuEngineEntry, c.dxgiAdapterLuidDescriptionMap[luid])
		}
	}

	var dst_wp []WorkerProcess
	q_wp := queryAll(&dst_wp)
	if err := wmi.QueryNamespace(q_wp, &dst_wp, "root\\WebAdministration"); err != nil {
		log.Debugf("Could not query WebAdministration namespace for IIS worker processes: %v. Skipping", err)
	}

	for _, process := range data {
		if process.Name == "_Total" ||
			c.processBlacklistPattern.MatchString(process.Name) ||
			!c.processWhitelistPattern.MatchString(process.Name) {
			continue
		}
		// Duplicate processes are suffixed # and an index number. Remove those.
		processName := strings.Split(process.Name, "#")[0]
		pid := strconv.FormatUint(uint64(process.IDProcess), 10)
		cpid := strconv.FormatUint(uint64(process.CreatingProcessID), 10)

		for _, wp := range dst_wp {
			if wp.ProcessId == uint64(process.IDProcess) {
				processName = strings.Join([]string{processName, wp.AppPoolName}, "_")
				break
			}
		}

		ch <- prometheus.MustNewConstMetric(
			c.StartTime,
			prometheus.GaugeValue,
			process.ElapsedTime,
			processName,
			pid,
			cpid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.HandleCount,
			prometheus.GaugeValue,
			process.HandleCount,
			processName,
			pid,
			cpid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.CPUTimeTotal,
			prometheus.CounterValue,
			process.PercentPrivilegedTime,
			processName,
			pid,
			cpid,
			"privileged",
		)

		ch <- prometheus.MustNewConstMetric(
			c.CPUTimeTotal,
			prometheus.CounterValue,
			process.PercentUserTime,
			processName,
			pid,
			cpid,
			"user",
		)

		ch <- prometheus.MustNewConstMetric(
			c.IOBytesTotal,
			prometheus.CounterValue,
			process.IOOtherBytesPerSec,
			processName,
			pid,
			cpid,
			"other",
		)

		ch <- prometheus.MustNewConstMetric(
			c.IOOperationsTotal,
			prometheus.CounterValue,
			process.IOOtherOperationsPerSec,
			processName,
			pid,
			cpid,
			"other",
		)

		ch <- prometheus.MustNewConstMetric(
			c.IOBytesTotal,
			prometheus.CounterValue,
			process.IOReadBytesPerSec,
			processName,
			pid,
			cpid,
			"read",
		)

		ch <- prometheus.MustNewConstMetric(
			c.IOOperationsTotal,
			prometheus.CounterValue,
			process.IOReadOperationsPerSec,
			processName,
			pid,
			cpid,
			"read",
		)

		ch <- prometheus.MustNewConstMetric(
			c.IOBytesTotal,
			prometheus.CounterValue,
			process.IOWriteBytesPerSec,
			processName,
			pid,
			cpid,
			"write",
		)

		ch <- prometheus.MustNewConstMetric(
			c.IOOperationsTotal,
			prometheus.CounterValue,
			process.IOWriteOperationsPerSec,
			processName,
			pid,
			cpid,
			"write",
		)

		ch <- prometheus.MustNewConstMetric(
			c.PageFaultsTotal,
			prometheus.CounterValue,
			process.PageFaultsPerSec,
			processName,
			pid,
			cpid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PageFileBytes,
			prometheus.GaugeValue,
			process.PageFileBytes,
			processName,
			pid,
			cpid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PoolBytes,
			prometheus.GaugeValue,
			process.PoolNonpagedBytes,
			processName,
			pid,
			cpid,
			"nonpaged",
		)

		ch <- prometheus.MustNewConstMetric(
			c.PoolBytes,
			prometheus.GaugeValue,
			process.PoolPagedBytes,
			processName,
			pid,
			cpid,
			"paged",
		)

		ch <- prometheus.MustNewConstMetric(
			c.PriorityBase,
			prometheus.GaugeValue,
			process.PriorityBase,
			processName,
			pid,
			cpid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PrivateBytes,
			prometheus.GaugeValue,
			process.PrivateBytes,
			processName,
			pid,
			cpid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ThreadCount,
			prometheus.GaugeValue,
			process.ThreadCount,
			processName,
			pid,
			cpid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.VirtualBytes,
			prometheus.GaugeValue,
			process.VirtualBytes,
			processName,
			pid,
			cpid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WorkingSetPrivate,
			prometheus.GaugeValue,
			process.WorkingSetPrivate,
			processName,
			pid,
			cpid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WorkingSetPeak,
			prometheus.GaugeValue,
			process.WorkingSetPeak,
			processName,
			pid,
			cpid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WorkingSet,
			prometheus.GaugeValue,
			process.WorkingSet,
			processName,
			pid,
			cpid,
		)

		processGpuMetricsEntry, processGpuMetricsEntryPresent := processGpuMetrics[pid]
		if processGpuMetricsEntryPresent {
			processGpuMetricsEntry.exposeMetrics(ch, c, processName, pid, cpid)
		}
	}

	return nil
}

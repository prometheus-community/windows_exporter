//go:build windows
// +build windows

package collector

import (
	"github.com/StackExchange/wmi"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("netframework_clrmemory", NewNETFramework_NETCLRMemoryCollector)
}

// A NETFramework_NETCLRMemoryCollector is a Prometheus collector for WMI Win32_PerfRawData_NETFramework_NETCLRMemory metrics
type NETFramework_NETCLRMemoryCollector struct {
	AllocatedBytes                     *prometheus.Desc
	FinalizationSurvivors              *prometheus.Desc
	HeapSize                           *prometheus.Desc
	PromotedBytes                      *prometheus.Desc
	NumberGCHandles                    *prometheus.Desc
	NumberCollections                  *prometheus.Desc
	NumberInducedGC                    *prometheus.Desc
	NumberofPinnedObjects              *prometheus.Desc
	NumberofSinkBlocksinuse            *prometheus.Desc
	NumberTotalCommittedBytes          *prometheus.Desc
	NumberTotalreservedBytes           *prometheus.Desc
	TimeinGC                           *prometheus.Desc
	PromotedFinalizationMemoryfromGen0 *prometheus.Desc
	PromotedMemoryfromGen0             *prometheus.Desc
	PromotedMemoryfromGen1             *prometheus.Desc
}

// NewNETFramework_NETCLRMemoryCollector ...
func NewNETFramework_NETCLRMemoryCollector() (Collector, error) {
	const subsystem = "netframework_clrmemory"
	return &NETFramework_NETCLRMemoryCollector{
		AllocatedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "allocated_bytes_total"),
			"Displays the total number of bytes allocated on the garbage collection heap.",
			[]string{"process"},
			nil,
		),
		FinalizationSurvivors: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "finalization_survivors"),
			"Displays the number of garbage-collected objects that survive a collection because they are waiting to be finalized.",
			[]string{"process"},
			nil,
		),
		HeapSize: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "heap_size_bytes"),
			"Displays the maximum bytes that can be allocated; it does not indicate the current number of bytes allocated.",
			[]string{"process", "area"},
			nil,
		),
		PromotedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "promoted_bytes"),
			"Displays the bytes that were promoted from the generation to the next one during the last GC. Memory is promoted when it survives a garbage collection.",
			[]string{"process", "area"},
			nil,
		),
		NumberGCHandles: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "number_gc_handles"),
			"Displays the current number of garbage collection handles in use. Garbage collection handles are handles to resources external to the common language runtime and the managed environment.",
			[]string{"process"},
			nil,
		),
		NumberCollections: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "collections_total"),
			"Displays the number of times the generation objects are garbage collected since the application started.",
			[]string{"process", "area"},
			nil,
		),
		NumberInducedGC: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "induced_gc_total"),
			"Displays the peak number of times garbage collection was performed because of an explicit call to GC.Collect.",
			[]string{"process"},
			nil,
		),
		NumberofPinnedObjects: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "number_pinned_objects"),
			"Displays the number of pinned objects encountered in the last garbage collection.",
			[]string{"process"},
			nil,
		),
		NumberofSinkBlocksinuse: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "number_sink_blocksinuse"),
			"Displays the current number of synchronization blocks in use. Synchronization blocks are per-object data structures allocated for storing synchronization information. They hold weak references to managed objects and must be scanned by the garbage collector.",
			[]string{"process"},
			nil,
		),
		NumberTotalCommittedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "committed_bytes"),
			"Displays the amount of virtual memory, in bytes, currently committed by the garbage collector. Committed memory is the physical memory for which space has been reserved in the disk paging file.",
			[]string{"process"},
			nil,
		),
		NumberTotalreservedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "reserved_bytes"),
			"Displays the amount of virtual memory, in bytes, currently reserved by the garbage collector. Reserved memory is the virtual memory space reserved for the application when no disk or main memory pages have been used.",
			[]string{"process"},
			nil,
		),
		TimeinGC: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "gc_time_percent"),
			"Displays the percentage of time that was spent performing a garbage collection in the last sample.",
			[]string{"process"},
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *NETFramework_NETCLRMemoryCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting win32_perfrawdata_netframework_netclrmemory metrics:", desc, err)
		return err
	}
	return nil
}

type Win32_PerfRawData_NETFramework_NETCLRMemory struct {
	Name string

	AllocatedBytesPersec      uint64
	FinalizationSurvivors     uint64
	Frequency_PerfTime        uint64
	Gen0heapsize              uint64
	Gen0PromotedBytesPerSec   uint64
	Gen1heapsize              uint64
	Gen1PromotedBytesPerSec   uint64
	Gen2heapsize              uint64
	LargeObjectHeapsize       uint64
	NumberBytesinallHeaps     uint64
	NumberGCHandles           uint64
	NumberGen0Collections     uint64
	NumberGen1Collections     uint64
	NumberGen2Collections     uint64
	NumberInducedGC           uint64
	NumberofPinnedObjects     uint64
	NumberofSinkBlocksinuse   uint64
	NumberTotalcommittedBytes uint64
	NumberTotalreservedBytes  uint64
	// PercentTimeinGC has countertype=PERF_RAW_FRACTION.
	// Formula: (100 * CounterValue) / BaseValue
	// By docs https://docs.microsoft.com/en-us/previous-versions/windows/internet-explorer/ie-developer/scripting-articles/ms974615(v=msdn.10)#perf_raw_fraction
	PercentTimeinGC uint32
	// BaseValue is just a "magic" number used to make the calculation come out right.
	PercentTimeinGC_base               uint32
	ProcessID                          uint64
	PromotedFinalizationMemoryfromGen0 uint64
	PromotedMemoryfromGen0             uint64
	PromotedMemoryfromGen1             uint64
}

func (c *NETFramework_NETCLRMemoryCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_NETFramework_NETCLRMemory
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	for _, process := range dst {

		if process.Name == "_Global_" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.AllocatedBytes,
			prometheus.CounterValue,
			float64(process.AllocatedBytesPersec),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FinalizationSurvivors,
			prometheus.GaugeValue,
			float64(process.FinalizationSurvivors),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.HeapSize,
			prometheus.GaugeValue,
			float64(process.Gen0heapsize),
			process.Name,
			"Gen0",
		)

		ch <- prometheus.MustNewConstMetric(
			c.PromotedBytes,
			prometheus.GaugeValue,
			float64(process.Gen0PromotedBytesPerSec),
			process.Name,
			"Gen0",
		)

		ch <- prometheus.MustNewConstMetric(
			c.HeapSize,
			prometheus.GaugeValue,
			float64(process.Gen1heapsize),
			process.Name,
			"Gen1",
		)

		ch <- prometheus.MustNewConstMetric(
			c.PromotedBytes,
			prometheus.GaugeValue,
			float64(process.Gen1PromotedBytesPerSec),
			process.Name,
			"Gen1",
		)

		ch <- prometheus.MustNewConstMetric(
			c.HeapSize,
			prometheus.GaugeValue,
			float64(process.Gen2heapsize),
			process.Name,
			"Gen2",
		)

		ch <- prometheus.MustNewConstMetric(
			c.HeapSize,
			prometheus.GaugeValue,
			float64(process.LargeObjectHeapsize),
			process.Name,
			"LOH",
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberGCHandles,
			prometheus.GaugeValue,
			float64(process.NumberGCHandles),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberCollections,
			prometheus.CounterValue,
			float64(process.NumberGen0Collections),
			process.Name,
			"Gen0",
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberCollections,
			prometheus.CounterValue,
			float64(process.NumberGen1Collections),
			process.Name,
			"Gen1",
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberCollections,
			prometheus.CounterValue,
			float64(process.NumberGen2Collections),
			process.Name,
			"Gen2",
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberInducedGC,
			prometheus.CounterValue,
			float64(process.NumberInducedGC),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberofPinnedObjects,
			prometheus.GaugeValue,
			float64(process.NumberofPinnedObjects),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberofSinkBlocksinuse,
			prometheus.GaugeValue,
			float64(process.NumberofSinkBlocksinuse),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberTotalCommittedBytes,
			prometheus.GaugeValue,
			float64(process.NumberTotalcommittedBytes),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberTotalreservedBytes,
			prometheus.GaugeValue,
			float64(process.NumberTotalreservedBytes),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TimeinGC,
			prometheus.GaugeValue,
			float64(100*process.PercentTimeinGC)/float64(process.PercentTimeinGC_base),
			process.Name,
		)
	}

	return nil, nil
}

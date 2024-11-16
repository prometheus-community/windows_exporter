//go:build windows

package netframework

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
)

func (c *Collector) buildClrMemory() {
	c.allocatedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "allocated_bytes_total"),
		"Displays the total number of bytes allocated on the garbage collection heap.",
		[]string{"process"},
		nil,
	)
	c.finalizationSurvivors = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "finalization_survivors"),
		"Displays the number of garbage-collected objects that survive a collection because they are waiting to be finalized.",
		[]string{"process"},
		nil,
	)
	c.heapSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "heap_size_bytes"),
		"Displays the maximum bytes that can be allocated; it does not indicate the current number of bytes allocated.",
		[]string{"process", "area"},
		nil,
	)
	c.promotedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "promoted_bytes"),
		"Displays the bytes that were promoted from the generation to the next one during the last GC. Memory is promoted when it survives a garbage collection.",
		[]string{"process", "area"},
		nil,
	)
	c.numberGCHandles = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "number_gc_handles"),
		"Displays the current number of garbage collection handles in use. Garbage collection handles are handles to resources external to the common language runtime and the managed environment.",
		[]string{"process"},
		nil,
	)
	c.numberCollections = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "collections_total"),
		"Displays the number of times the generation objects are garbage collected since the application started.",
		[]string{"process", "area"},
		nil,
	)
	c.numberInducedGC = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "induced_gc_total"),
		"Displays the peak number of times garbage collection was performed because of an explicit call to GC.Collect.",
		[]string{"process"},
		nil,
	)
	c.numberOfPinnedObjects = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "number_pinned_objects"),
		"Displays the number of pinned objects encountered in the last garbage collection.",
		[]string{"process"},
		nil,
	)
	c.numberOfSinkBlocksInUse = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "number_sink_blocksinuse"),
		"Displays the current number of synchronization blocks in use. Synchronization blocks are per-object data structures allocated for storing synchronization information. They hold weak references to managed objects and must be scanned by the garbage collector.",
		[]string{"process"},
		nil,
	)
	c.numberTotalCommittedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "committed_bytes"),
		"Displays the amount of virtual memory, in bytes, currently committed by the garbage collector. Committed memory is the physical memory for which space has been reserved in the disk paging file.",
		[]string{"process"},
		nil,
	)
	c.numberTotalReservedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "reserved_bytes"),
		"Displays the amount of virtual memory, in bytes, currently reserved by the garbage collector. Reserved memory is the virtual memory space reserved for the application when no disk or main memory pages have been used.",
		[]string{"process"},
		nil,
	)
	c.timeInGC = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "gc_time_percent"),
		"Displays the percentage of time that was spent performing a garbage collection in the last sample.",
		[]string{"process"},
		nil,
	)
}

type Win32_PerfRawData_NETFramework_NETCLRMemory struct {
	Name string `mi:"Name"`

	AllocatedBytesPersec      uint64 `mi:"AllocatedBytesPersec"`
	FinalizationSurvivors     uint64 `mi:"FinalizationSurvivors"`
	Frequency_PerfTime        uint64 `mi:"Frequency_PerfTime"`
	Gen0heapsize              uint64 `mi:"Gen0heapsize"`
	Gen0PromotedBytesPerSec   uint64 `mi:"Gen0PromotedBytesPersec"`
	Gen1heapsize              uint64 `mi:"Gen1heapsize"`
	Gen1PromotedBytesPerSec   uint64 `mi:"Gen1PromotedBytesPersec"`
	Gen2heapsize              uint64 `mi:"Gen2heapsize"`
	LargeObjectHeapsize       uint64 `mi:"LargeObjectHeapsize"`
	NumberBytesinallHeaps     uint64 `mi:"NumberBytesinallHeaps"`
	NumberGCHandles           uint64 `mi:"NumberGCHandles"`
	NumberGen0Collections     uint64 `mi:"NumberGen0Collections"`
	NumberGen1Collections     uint64 `mi:"NumberGen1Collections"`
	NumberGen2Collections     uint64 `mi:"NumberGen2Collections"`
	NumberInducedGC           uint64 `mi:"NumberInducedGC"`
	NumberofPinnedObjects     uint64 `mi:"NumberofPinnedObjects"`
	NumberofSinkBlocksinuse   uint64 `mi:"NumberofSinkBlocksinuse"`
	NumberTotalcommittedBytes uint64 `mi:"NumberTotalcommittedBytes"`
	NumberTotalreservedBytes  uint64 `mi:"NumberTotalreservedBytes"`
	// PercentTimeinGC has countertype=PERF_RAW_FRACTION.
	// Formula: (100 * CounterValue) / BaseValue
	// By docs https://docs.microsoft.com/en-us/previous-versions/windows/internet-explorer/ie-developer/scripting-articles/ms974615(v=msdn.10)#perf_raw_fraction
	PercentTimeinGC uint32 `mi:"PercentTimeinGC"`
	// BaseValue is just a "magic" number used to make the calculation come out right.
	PercentTimeinGC_base               uint32 `mi:"PercentTimeinGC_base"`
	ProcessID                          uint64 `mi:"ProcessID"`
	PromotedFinalizationMemoryfromGen0 uint64 `mi:"PromotedFinalizationMemoryfromGen0"`
	PromotedMemoryfromGen0             uint64 `mi:"PromotedMemoryfromGen0"`
	PromotedMemoryfromGen1             uint64 `mi:"PromotedMemoryfromGen1"`
}

func (c *Collector) collectClrMemory(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_NETFramework_NETCLRMemory
	if err := c.miSession.Query(&dst, mi.NamespaceRootCIMv2, utils.Must(mi.NewQuery("SELECT * Win32_PerfRawData_NETFramework_NETCLRMemory"))); err != nil {
		return fmt.Errorf("WMI query failed: %w", err)
	}

	for _, process := range dst {
		if process.Name == "_Global_" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.allocatedBytes,
			prometheus.CounterValue,
			float64(process.AllocatedBytesPersec),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.finalizationSurvivors,
			prometheus.GaugeValue,
			float64(process.FinalizationSurvivors),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.heapSize,
			prometheus.GaugeValue,
			float64(process.Gen0heapsize),
			process.Name,
			"Gen0",
		)

		ch <- prometheus.MustNewConstMetric(
			c.promotedBytes,
			prometheus.GaugeValue,
			float64(process.Gen0PromotedBytesPerSec),
			process.Name,
			"Gen0",
		)

		ch <- prometheus.MustNewConstMetric(
			c.heapSize,
			prometheus.GaugeValue,
			float64(process.Gen1heapsize),
			process.Name,
			"Gen1",
		)

		ch <- prometheus.MustNewConstMetric(
			c.promotedBytes,
			prometheus.GaugeValue,
			float64(process.Gen1PromotedBytesPerSec),
			process.Name,
			"Gen1",
		)

		ch <- prometheus.MustNewConstMetric(
			c.heapSize,
			prometheus.GaugeValue,
			float64(process.Gen2heapsize),
			process.Name,
			"Gen2",
		)

		ch <- prometheus.MustNewConstMetric(
			c.heapSize,
			prometheus.GaugeValue,
			float64(process.LargeObjectHeapsize),
			process.Name,
			"LOH",
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberGCHandles,
			prometheus.GaugeValue,
			float64(process.NumberGCHandles),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberCollections,
			prometheus.CounterValue,
			float64(process.NumberGen0Collections),
			process.Name,
			"Gen0",
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberCollections,
			prometheus.CounterValue,
			float64(process.NumberGen1Collections),
			process.Name,
			"Gen1",
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberCollections,
			prometheus.CounterValue,
			float64(process.NumberGen2Collections),
			process.Name,
			"Gen2",
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberInducedGC,
			prometheus.CounterValue,
			float64(process.NumberInducedGC),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberOfPinnedObjects,
			prometheus.GaugeValue,
			float64(process.NumberofPinnedObjects),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberOfSinkBlocksInUse,
			prometheus.GaugeValue,
			float64(process.NumberofSinkBlocksinuse),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberTotalCommittedBytes,
			prometheus.GaugeValue,
			float64(process.NumberTotalcommittedBytes),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberTotalReservedBytes,
			prometheus.GaugeValue,
			float64(process.NumberTotalreservedBytes),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.timeInGC,
			prometheus.GaugeValue,
			float64(100*process.PercentTimeinGC)/float64(process.PercentTimeinGC_base),
			process.Name,
		)
	}

	return nil
}

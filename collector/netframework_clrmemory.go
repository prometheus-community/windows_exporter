// +build windows

package collector

import (
	"fmt"
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	registerCollector("netframework_clrmemory", NewNETFrameworkCLRMemoryCollector, ".NET CLR Memory")
}

// A NETFrameworkCLRMemoryCollector is a Prometheus collector for Perflib .NET CLR Memory metrics
type NETFrameworkCLRMemoryCollector struct {
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

	processWhitelistPattern *regexp.Regexp
	processBlacklistPattern *regexp.Regexp
}

// NewNETFrameworkCLRMemoryCollector ...
func NewNETFrameworkCLRMemoryCollector() (Collector, error) {
	const subsystem = "netframework_clrmemory"
	commonFlags := GetNETFrameworkFlags()
	return &NETFrameworkCLRMemoryCollector{
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
		processWhitelistPattern: commonFlags.whitelistRegexp,
		processBlacklistPattern: commonFlags.blacklistRegexp,
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *NETFrameworkCLRMemoryCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ctx, ch); err != nil {
		log.Error("failed collecting netframework_clrmemory metrics:", desc, err)
		return err
	}
	return nil
}

type netframeworkCLRMemory struct {
	Name string

	AllocatedBytesPersec               float64 `perflib:"Allocated Bytes/sec"`
	FinalizationSurvivors              float64 `perflib:"Finalization Survivors"`
	Frequency_PerfTime                 float64 `perflib:"Not Displayed_Base"`
	Gen0heapsize                       float64 `perflib:"Gen 0 heap size"`
	Gen0PromotedBytesPerSec            float64 `perflib:"Gen 0 Promoted Bytes/Sec"`
	Gen1heapsize                       float64 `perflib:"Gen 1 heap size"`
	Gen1PromotedBytesPerSec            float64 `perflib:"Gen 1 Promoted Bytes/Sec"`
	Gen2heapsize                       float64 `perflib:"Gen 2 heap size"`
	LargeObjectHeapsize                float64 `perflib:"Large Object Heap size"`
	NumberBytesinallHeaps              float64 `perflib:"# Bytes in all Heaps"`
	NumberGCHandles                    float64 `perflib:"# GC Handles"`
	NumberGen0Collections              float64 `perflib:"# Gen 0 Collections"`
	NumberGen1Collections              float64 `perflib:"# Gen 1 Collections"`
	NumberGen2Collections              float64 `perflib:"# Gen 2 Collections"`
	NumberInducedGC                    float64 `perflib:"# Induced GC"`
	NumberofPinnedObjects              float64 `perflib:"# of Pinned Objects"`
	NumberofSinkBlocksinuse            float64 `perflib:"# of Sink Blocks in use"`
	NumberTotalcommittedBytes          float64 `perflib:"# Total committed Bytes"`
	NumberTotalreservedBytes           float64 `perflib:"# Total reserved Bytes"`
	PercentTimeinGC                    float64 `perflib:"% Time in GC"`
	ProcessID                          float64 `perflib:"Process ID"`
	PromotedFinalizationMemoryfromGen0 float64 `perflib:"Promoted Finalization-Memory from Gen 0"`
	PromotedMemoryfromGen0             float64 `perflib:"Promoted Memory from Gen 0"`
	PromotedMemoryfromGen1             float64 `perflib:"Promoted Memory from Gen 1"`
}

func (c *NETFrameworkCLRMemoryCollector) collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []netframeworkCLRMemory

	if err := unmarshalObject(ctx.perfObjects[".NET CLR Memory"], &dst); err != nil {
		return nil, err
	}

	var names = make(map[string]int, len(dst))
	for _, process := range dst {

		if process.Name == "_Global_" {
			continue
		}

		// Append "#1", "#2", etc., to process names to disambiguate duplicates.
		name := process.Name
		procnum, exists := names[name]
		if exists {
			names[name]++
			name = fmt.Sprintf("%s#%d", name, procnum)
		} else {
			names[name] = 1
		}

		// The pattern matching against the whitelist and blacklist has to occur
		// after appending #N above to be consistent with other collectors.
		if c.processBlacklistPattern.MatchString(name) ||
			!c.processWhitelistPattern.MatchString(name) {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.AllocatedBytes,
			prometheus.CounterValue,
			process.AllocatedBytesPersec,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FinalizationSurvivors,
			prometheus.GaugeValue,
			process.FinalizationSurvivors,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.HeapSize,
			prometheus.GaugeValue,
			process.Gen0heapsize,
			name,
			"Gen0",
		)

		ch <- prometheus.MustNewConstMetric(
			c.PromotedBytes,
			prometheus.GaugeValue,
			process.Gen0PromotedBytesPerSec,
			name,
			"Gen0",
		)

		ch <- prometheus.MustNewConstMetric(
			c.HeapSize,
			prometheus.GaugeValue,
			process.Gen1heapsize,
			name,
			"Gen1",
		)

		ch <- prometheus.MustNewConstMetric(
			c.PromotedBytes,
			prometheus.GaugeValue,
			process.Gen1PromotedBytesPerSec,
			name,
			"Gen1",
		)

		ch <- prometheus.MustNewConstMetric(
			c.HeapSize,
			prometheus.GaugeValue,
			process.Gen2heapsize,
			name,
			"Gen2",
		)

		ch <- prometheus.MustNewConstMetric(
			c.HeapSize,
			prometheus.GaugeValue,
			process.LargeObjectHeapsize,
			name,
			"LOH",
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberGCHandles,
			prometheus.GaugeValue,
			process.NumberGCHandles,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberCollections,
			prometheus.CounterValue,
			process.NumberGen0Collections,
			name,
			"Gen0",
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberCollections,
			prometheus.CounterValue,
			process.NumberGen1Collections,
			name,
			"Gen1",
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberCollections,
			prometheus.CounterValue,
			process.NumberGen2Collections,
			name,
			"Gen2",
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberInducedGC,
			prometheus.CounterValue,
			process.NumberInducedGC,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberofPinnedObjects,
			prometheus.GaugeValue,
			process.NumberofPinnedObjects,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberofSinkBlocksinuse,
			prometheus.GaugeValue,
			process.NumberofSinkBlocksinuse,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberTotalCommittedBytes,
			prometheus.GaugeValue,
			process.NumberTotalcommittedBytes,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberTotalreservedBytes,
			prometheus.GaugeValue,
			process.NumberTotalreservedBytes,
			name,
		)

		timeinGC := 0.0
		if process.Frequency_PerfTime != 0 {
			timeinGC = process.PercentTimeinGC / process.Frequency_PerfTime
		}
		ch <- prometheus.MustNewConstMetric(
			c.TimeinGC,
			prometheus.GaugeValue,
			timeinGC,
			name,
		)
	}

	return nil, nil
}

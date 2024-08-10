//go:build windows

package netframework_clrmemory

import (
	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "netframework_clrmemory"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_NETFramework_NETCLRMemory metrics.
type Collector struct {
	logger log.Logger

	allocatedBytes            *prometheus.Desc
	finalizationSurvivors     *prometheus.Desc
	heapSize                  *prometheus.Desc
	promotedBytes             *prometheus.Desc
	numberGCHandles           *prometheus.Desc
	numberCollections         *prometheus.Desc
	numberInducedGC           *prometheus.Desc
	numberOfPinnedObjects     *prometheus.Desc
	numberOfSinkBlocksInUse   *prometheus.Desc
	numberTotalCommittedBytes *prometheus.Desc
	numberTotalReservedBytes  *prometheus.Desc
	timeInGC                  *prometheus.Desc
}

func New(logger log.Logger, _ *Config) *Collector {
	c := &Collector{}
	c.SetLogger(logger)

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *Collector) GetPerfCounter() ([]string, error) {
	return []string{}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build() error {
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
	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.collect(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting win32_perfrawdata_netframework_netclrmemory metrics", "err", err)
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

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_NETFramework_NETCLRMemory
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
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

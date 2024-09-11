// returns data points from Win32_PerfRawData_PerfOS_Memory
// <add link to documentation here> - Win32_PerfRawData_PerfOS_Memory class

//go:build windows

package memory

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/pkg/headers/sysinfoapi"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
)

const Name = "memory"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for perflib Memory metrics.
type Collector struct {
	config Config

	// Performance metrics
	availableBytes                  *prometheus.Desc
	cacheBytes                      *prometheus.Desc
	cacheBytesPeak                  *prometheus.Desc
	cacheFaultsTotal                *prometheus.Desc
	commitLimit                     *prometheus.Desc
	committedBytes                  *prometheus.Desc
	demandZeroFaultsTotal           *prometheus.Desc
	freeAndZeroPageListBytes        *prometheus.Desc
	freeSystemPageTableEntries      *prometheus.Desc
	modifiedPageListBytes           *prometheus.Desc
	pageFaultsTotal                 *prometheus.Desc
	swapPageReadsTotal              *prometheus.Desc
	swapPagesReadTotal              *prometheus.Desc
	swapPagesWrittenTotal           *prometheus.Desc
	swapPageOperationsTotal         *prometheus.Desc
	swapPageWritesTotal             *prometheus.Desc
	poolNonPagedAllocationsTotal    *prometheus.Desc
	poolNonPagedBytes               *prometheus.Desc
	poolPagedAllocationsTotal       *prometheus.Desc
	poolPagedBytes                  *prometheus.Desc
	poolPagedResidentBytes          *prometheus.Desc
	standbyCacheCoreBytes           *prometheus.Desc
	standbyCacheNormalPriorityBytes *prometheus.Desc
	standbyCacheReserveBytes        *prometheus.Desc
	systemCacheResidentBytes        *prometheus.Desc
	systemCodeResidentBytes         *prometheus.Desc
	systemCodeTotalBytes            *prometheus.Desc
	systemDriverResidentBytes       *prometheus.Desc
	systemDriverTotalBytes          *prometheus.Desc
	transitionFaultsTotal           *prometheus.Desc
	transitionPagesRepurposedTotal  *prometheus.Desc
	writeCopiesTotal                *prometheus.Desc

	// Global memory status
	processMemoryLimitBytes  *prometheus.Desc
	physicalMemoryTotalBytes *prometheus.Desc
	physicalMemoryFreeBytes  *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) GetPerfCounter(_ *slog.Logger) ([]string, error) {
	return []string{"Memory"}, nil
}

func (c *Collector) Close(_ *slog.Logger) error {
	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *wmi.Client) error {
	c.availableBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "available_bytes"),
		"The amount of physical memory immediately available for allocation to a process or for system use. It is equal to the sum of memory assigned to"+
			" the standby (cached), free and zero page lists (AvailableBytes)",
		nil,
		nil,
	)
	c.cacheBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cache_bytes"),
		"(CacheBytes)",
		nil,
		nil,
	)
	c.cacheBytesPeak = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cache_bytes_peak"),
		"(CacheBytesPeak)",
		nil,
		nil,
	)
	c.cacheFaultsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cache_faults_total"),
		"Number of faults which occur when a page sought in the file system cache is not found there and must be retrieved from elsewhere in memory (soft fault) "+
			"or from disk (hard fault) (Cache Faults/sec)",
		nil,
		nil,
	)
	c.commitLimit = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "commit_limit"),
		"(CommitLimit)",
		nil,
		nil,
	)
	c.committedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "committed_bytes"),
		"(CommittedBytes)",
		nil,
		nil,
	)
	c.demandZeroFaultsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "demand_zero_faults_total"),
		"The number of zeroed pages required to satisfy faults. Zeroed pages, pages emptied of previously stored data and filled with zeros, are a security"+
			" feature of Windows that prevent processes from seeing data stored by earlier processes that used the memory space (Demand Zero Faults/sec)",
		nil,
		nil,
	)
	c.freeAndZeroPageListBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "free_and_zero_page_list_bytes"),
		"The amount of physical memory, in bytes, that is assigned to the free and zero page lists. This memory does not contain cached data. It is immediately"+
			" available for allocation to a process or for system use (FreeAndZeroPageListBytes)",
		nil,
		nil,
	)
	c.freeSystemPageTableEntries = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "free_system_page_table_entries"),
		"(FreeSystemPageTableEntries)",
		nil,
		nil,
	)
	c.modifiedPageListBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "modified_page_list_bytes"),
		"The amount of physical memory, in bytes, that is assigned to the modified page list. This memory contains cached data and code that is not actively in "+
			"use by processes, the system and the system cache (ModifiedPageListBytes)",
		nil,
		nil,
	)
	c.pageFaultsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "page_faults_total"),
		"Overall rate at which faulted pages are handled by the processor (Page Faults/sec)",
		nil,
		nil,
	)
	c.swapPageReadsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "swap_page_reads_total"),
		"Number of disk page reads (a single read operation reading several pages is still only counted once) (PageReadsPersec)",
		nil,
		nil,
	)
	c.swapPagesReadTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "swap_pages_read_total"),
		"Number of pages read across all page reads (ie counting all pages read even if they are read in a single operation) (PagesInputPersec)",
		nil,
		nil,
	)
	c.swapPagesWrittenTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "swap_pages_written_total"),
		"Number of pages written across all page writes (ie counting all pages written even if they are written in a single operation) (PagesOutputPersec)",
		nil,
		nil,
	)
	c.swapPageOperationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "swap_page_operations_total"),
		"Total number of swap page read and writes (PagesPersec)",
		nil,
		nil,
	)
	c.swapPageWritesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "swap_page_writes_total"),
		"Number of disk page writes (a single write operation writing several pages is still only counted once) (PageWritesPersec)",
		nil,
		nil,
	)
	c.poolNonPagedAllocationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "pool_nonpaged_allocs_total"),
		"The number of calls to allocate space in the nonpaged pool. The nonpaged pool is an area of system memory area for objects that cannot be written"+
			" to disk, and must remain in physical memory as long as they are allocated (PoolNonpagedAllocs)",
		nil,
		nil,
	)
	c.poolNonPagedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "pool_nonpaged_bytes"),
		"Number of bytes in the non-paged pool, an area of the system virtual memory that is used for objects that cannot be written to disk, but must "+
			"remain in physical memory as long as they are allocated (PoolNonpagedBytes)",
		nil,
		nil,
	)
	c.poolPagedAllocationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "pool_paged_allocs_total"),
		"Number of calls to allocate space in the paged pool, regardless of the amount of space allocated in each call (PoolPagedAllocs)",
		nil,
		nil,
	)
	c.poolPagedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "pool_paged_bytes"),
		"(PoolPagedBytes)",
		nil,
		nil,
	)
	c.poolPagedResidentBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "pool_paged_resident_bytes"),
		"The size, in bytes, of the portion of the paged pool that is currently resident and active in physical memory. The paged pool is an area of the "+
			"system virtual memory that is used for objects that can be written to disk when they are not being used (PoolPagedResidentBytes)",
		nil,
		nil,
	)
	c.standbyCacheCoreBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "standby_cache_core_bytes"),
		"The amount of physical memory, in bytes, that is assigned to the core standby cache page lists. This memory contains cached data and code that is "+
			"not actively in use by processes, the system and the system cache (StandbyCacheCoreBytes)",
		nil,
		nil,
	)
	c.standbyCacheNormalPriorityBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "standby_cache_normal_priority_bytes"),
		"The amount of physical memory, in bytes, that is assigned to the normal priority standby cache page lists. This memory contains cached data and "+
			"code that is not actively in use by processes, the system and the system cache (StandbyCacheNormalPriorityBytes)",
		nil,
		nil,
	)
	c.standbyCacheReserveBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "standby_cache_reserve_bytes"),
		"The amount of physical memory, in bytes, that is assigned to the reserve standby cache page lists. This memory contains cached data and code "+
			"that is not actively in use by processes, the system and the system cache (StandbyCacheReserveBytes)",
		nil,
		nil,
	)
	c.systemCacheResidentBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "system_cache_resident_bytes"),
		"The size, in bytes, of the portion of the system file cache which is currently resident and active in physical memory (SystemCacheResidentBytes)",
		nil,
		nil,
	)
	c.systemCodeResidentBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "system_code_resident_bytes"),
		"The size, in bytes, of the pageable operating system code that is currently resident and active in physical memory (SystemCodeResidentBytes)",
		nil,
		nil,
	)
	c.systemCodeTotalBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "system_code_total_bytes"),
		"The size, in bytes, of the pageable operating system code currently mapped into the system virtual address space (SystemCodeTotalBytes)",
		nil,
		nil,
	)
	c.systemDriverResidentBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "system_driver_resident_bytes"),
		"The size, in bytes, of the pageable physical memory being used by device drivers. It is the working set (physical memory area) of the drivers (SystemDriverResidentBytes)",
		nil,
		nil,
	)
	c.systemDriverTotalBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "system_driver_total_bytes"),
		"The size, in bytes, of the pageable virtual memory currently being used by device drivers. Pageable memory can be written to disk when it is not being used (SystemDriverTotalBytes)",
		nil,
		nil,
	)
	c.transitionFaultsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transition_faults_total"),
		"Number of faults rate at which page faults are resolved by recovering pages that were being used by another process sharing the page, or were on the "+
			"modified page list or the standby list, or were being written to disk at the time of the page fault (TransitionFaultsPersec)",
		nil,
		nil,
	)
	c.transitionPagesRepurposedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transition_pages_repurposed_total"),
		"Transition Pages RePurposed is the rate at which the number of transition cache pages were reused for a different purpose (TransitionPagesRePurposedPersec)",
		nil,
		nil,
	)
	c.writeCopiesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "write_copies_total"),
		"The number of page faults caused by attempting to write that were satisfied by copying the page from elsewhere in physical memory (WriteCopiesPersec)",
		nil,
		nil,
	)
	c.processMemoryLimitBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "process_memory_limit_bytes"),
		"The size of the user-mode portion of the virtual address space of the calling process, in bytes. This value depends on the type of process, the type of processor, and the configuration of the operating system.",
		nil,
		nil,
	)
	c.physicalMemoryTotalBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "physical_total_bytes"),
		"The amount of actual physical memory, in bytes.",
		nil,
		nil,
	)
	c.physicalMemoryFreeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "physical_free_bytes"),
		"The amount of physical memory currently available, in bytes. This is the amount of physical memory that can be immediately reused without having to write its contents to disk first. It is the sum of the size of the standby, free, and zero lists.",
		nil,
		nil,
	)

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	logger = logger.With(slog.String("collector", Name))

	errs := make([]error, 0, 2)

	if err := c.collectPerformanceData(ctx, logger, ch); err != nil {
		logger.Error("failed collecting memory metrics",
			slog.Any("err", err),
		)

		errs = append(errs, err)
	}

	if err := c.collectGlobalMemoryStatus(ch); err != nil {
		logger.Error("failed collecting memory metrics",
			slog.Any("err", err),
		)

		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func (c *Collector) collectGlobalMemoryStatus(ch chan<- prometheus.Metric) error {
	memoryStatusEx, err := sysinfoapi.GlobalMemoryStatusEx()
	if err != nil {
		return fmt.Errorf("failed to get memory status: %w", err)
	}

	ch <- prometheus.MustNewConstMetric(
		c.processMemoryLimitBytes,
		prometheus.GaugeValue,
		float64(memoryStatusEx.TotalVirtual),
	)

	ch <- prometheus.MustNewConstMetric(
		c.physicalMemoryTotalBytes,
		prometheus.GaugeValue,
		float64(memoryStatusEx.TotalPhys),
	)

	ch <- prometheus.MustNewConstMetric(
		c.physicalMemoryFreeBytes,
		prometheus.GaugeValue,
		float64(memoryStatusEx.AvailPhys),
	)

	return nil
}

type memory struct {
	AvailableBytes                  float64 `perflib:"Available Bytes"`
	AvailableKBytes                 float64 `perflib:"Available KBytes"`
	AvailableMBytes                 float64 `perflib:"Available MBytes"`
	CacheBytes                      float64 `perflib:"Cache Bytes"`
	CacheBytesPeak                  float64 `perflib:"Cache Bytes Peak"`
	CacheFaultsPersec               float64 `perflib:"Cache Faults/sec"`
	CommitLimit                     float64 `perflib:"Commit Limit"`
	CommittedBytes                  float64 `perflib:"Committed Bytes"`
	DemandZeroFaultsPersec          float64 `perflib:"Demand Zero Faults/sec"`
	FreeAndZeroPageListBytes        float64 `perflib:"Free & Zero Page List Bytes"`
	FreeSystemPageTableEntries      float64 `perflib:"Free System Page Table Entries"`
	ModifiedPageListBytes           float64 `perflib:"Modified Page List Bytes"`
	PageFaultsPersec                float64 `perflib:"Page Faults/sec"`
	PageReadsPersec                 float64 `perflib:"Page Reads/sec"`
	PagesInputPersec                float64 `perflib:"Pages Input/sec"`
	PagesOutputPersec               float64 `perflib:"Pages Output/sec"`
	PagesPersec                     float64 `perflib:"Pages/sec"`
	PageWritesPersec                float64 `perflib:"Page Writes/sec"`
	PoolNonpagedAllocs              float64 `perflib:"Pool Nonpaged Allocs"`
	PoolNonpagedBytes               float64 `perflib:"Pool Nonpaged Bytes"`
	PoolPagedAllocs                 float64 `perflib:"Pool Paged Allocs"`
	PoolPagedBytes                  float64 `perflib:"Pool Paged Bytes"`
	PoolPagedResidentBytes          float64 `perflib:"Pool Paged Resident Bytes"`
	StandbyCacheCoreBytes           float64 `perflib:"Standby Cache Core Bytes"`
	StandbyCacheNormalPriorityBytes float64 `perflib:"Standby Cache Normal Priority Bytes"`
	StandbyCacheReserveBytes        float64 `perflib:"Standby Cache Reserve Bytes"`
	SystemCacheResidentBytes        float64 `perflib:"System Cache Resident Bytes"`
	SystemCodeResidentBytes         float64 `perflib:"System Code Resident Bytes"`
	SystemCodeTotalBytes            float64 `perflib:"System Code Total Bytes"`
	SystemDriverResidentBytes       float64 `perflib:"System Driver Resident Bytes"`
	SystemDriverTotalBytes          float64 `perflib:"System Driver Total Bytes"`
	TransitionFaultsPersec          float64 `perflib:"Transition Faults/sec"`
	TransitionPagesRePurposedPersec float64 `perflib:"Transition Pages RePurposed/sec"`
	WriteCopiesPersec               float64 `perflib:"Write Copies/sec"`
}

func (c *Collector) collectPerformanceData(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	logger = logger.With(slog.String("collector", Name))

	var dst []memory

	if err := perflib.UnmarshalObject(ctx.PerfObjects["Memory"], &dst, logger); err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		c.availableBytes,
		prometheus.GaugeValue,
		dst[0].AvailableBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.cacheBytes,
		prometheus.GaugeValue,
		dst[0].CacheBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.cacheBytesPeak,
		prometheus.GaugeValue,
		dst[0].CacheBytesPeak,
	)

	ch <- prometheus.MustNewConstMetric(
		c.cacheFaultsTotal,
		prometheus.CounterValue,
		dst[0].CacheFaultsPersec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.commitLimit,
		prometheus.GaugeValue,
		dst[0].CommitLimit,
	)

	ch <- prometheus.MustNewConstMetric(
		c.committedBytes,
		prometheus.GaugeValue,
		dst[0].CommittedBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.demandZeroFaultsTotal,
		prometheus.CounterValue,
		dst[0].DemandZeroFaultsPersec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.freeAndZeroPageListBytes,
		prometheus.GaugeValue,
		dst[0].FreeAndZeroPageListBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.freeSystemPageTableEntries,
		prometheus.GaugeValue,
		dst[0].FreeSystemPageTableEntries,
	)

	ch <- prometheus.MustNewConstMetric(
		c.modifiedPageListBytes,
		prometheus.GaugeValue,
		dst[0].ModifiedPageListBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.pageFaultsTotal,
		prometheus.CounterValue,
		dst[0].PageFaultsPersec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.swapPageReadsTotal,
		prometheus.CounterValue,
		dst[0].PageReadsPersec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.swapPagesReadTotal,
		prometheus.CounterValue,
		dst[0].PagesInputPersec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.swapPagesWrittenTotal,
		prometheus.CounterValue,
		dst[0].PagesOutputPersec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.swapPageOperationsTotal,
		prometheus.CounterValue,
		dst[0].PagesPersec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.swapPageWritesTotal,
		prometheus.CounterValue,
		dst[0].PageWritesPersec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.poolNonPagedAllocationsTotal,
		prometheus.GaugeValue,
		dst[0].PoolNonpagedAllocs,
	)

	ch <- prometheus.MustNewConstMetric(
		c.poolNonPagedBytes,
		prometheus.GaugeValue,
		dst[0].PoolNonpagedBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.poolPagedAllocationsTotal,
		prometheus.CounterValue,
		dst[0].PoolPagedAllocs,
	)

	ch <- prometheus.MustNewConstMetric(
		c.poolPagedBytes,
		prometheus.GaugeValue,
		dst[0].PoolPagedBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.poolPagedResidentBytes,
		prometheus.GaugeValue,
		dst[0].PoolPagedResidentBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.standbyCacheCoreBytes,
		prometheus.GaugeValue,
		dst[0].StandbyCacheCoreBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.standbyCacheNormalPriorityBytes,
		prometheus.GaugeValue,
		dst[0].StandbyCacheNormalPriorityBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.standbyCacheReserveBytes,
		prometheus.GaugeValue,
		dst[0].StandbyCacheReserveBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.systemCacheResidentBytes,
		prometheus.GaugeValue,
		dst[0].SystemCacheResidentBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.systemCodeResidentBytes,
		prometheus.GaugeValue,
		dst[0].SystemCodeResidentBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.systemCodeTotalBytes,
		prometheus.GaugeValue,
		dst[0].SystemCodeTotalBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.systemDriverResidentBytes,
		prometheus.GaugeValue,
		dst[0].SystemDriverResidentBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.systemDriverTotalBytes,
		prometheus.GaugeValue,
		dst[0].SystemDriverTotalBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.transitionFaultsTotal,
		prometheus.CounterValue,
		dst[0].TransitionFaultsPersec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.transitionPagesRepurposedTotal,
		prometheus.CounterValue,
		dst[0].TransitionPagesRePurposedPersec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.writeCopiesTotal,
		prometheus.CounterValue,
		dst[0].WriteCopiesPersec,
	)

	return nil
}

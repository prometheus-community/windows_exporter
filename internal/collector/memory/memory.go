// returns data points from Win32_PerfRawData_PerfOS_Memory
// <add link to documentation here> - Win32_PerfRawData_PerfOS_Memory class

//go:build windows

package memory

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/headers/sysinfoapi"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "memory"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for perflib Memory metrics.
type Collector struct {
	config Config

	perfDataCollector *perfdata.Collector

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
	return []string{}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *mi.Session) error {
	counters := []string{
		availableBytes,
		availableKBytes,
		availableMBytes,
		cacheBytes,
		cacheBytesPeak,
		cacheFaultsPerSec,
		commitLimit,
		committedBytes,
		demandZeroFaultsPerSec,
		freeAndZeroPageListBytes,
		freeSystemPageTableEntries,
		modifiedPageListBytes,
		pageFaultsPerSec,
		pageReadsPerSec,
		pagesInputPerSec,
		pagesOutputPerSec,
		pagesPerSec,
		pageWritesPerSec,
		poolNonpagedAllocs,
		poolNonpagedBytes,
		poolPagedAllocs,
		poolPagedBytes,
		poolPagedResidentBytes,
		standbyCacheCoreBytes,
		standbyCacheNormalPriorityBytes,
		standbyCacheReserveBytes,
		systemCacheResidentBytes,
		systemCodeResidentBytes,
		systemCodeTotalBytes,
		systemDriverResidentBytes,
		systemDriverTotalBytes,
		transitionFaultsPerSec,
		transitionPagesRePurposedPerSec,
		writeCopiesPerSec,
	}

	var err error

	c.perfDataCollector, err = perfdata.NewCollector("Memory", perfdata.InstanceAll, counters)
	if err != nil {
		return fmt.Errorf("failed to create Memory collector: %w", err)
	}

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
		"Number of disk page reads (a single read operation reading several pages is still only counted once) (PageReadsPerSec)",
		nil,
		nil,
	)
	c.swapPagesReadTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "swap_pages_read_total"),
		"Number of pages read across all page reads (ie counting all pages read even if they are read in a single operation) (PagesInputPerSec)",
		nil,
		nil,
	)
	c.swapPagesWrittenTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "swap_pages_written_total"),
		"Number of pages written across all page writes (ie counting all pages written even if they are written in a single operation) (PagesOutputPerSec)",
		nil,
		nil,
	)
	c.swapPageOperationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "swap_page_operations_total"),
		"Total number of swap page read and writes (PagesPerSec)",
		nil,
		nil,
	)
	c.swapPageWritesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "swap_page_writes_total"),
		"Number of disk page writes (a single write operation writing several pages is still only counted once) (PageWritesPerSec)",
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
			"modified page list or the standby list, or were being written to disk at the time of the page fault (TransitionFaultsPerSec)",
		nil,
		nil,
	)
	c.transitionPagesRepurposedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transition_pages_repurposed_total"),
		"Transition Pages RePurposed is the rate at which the number of transition cache pages were reused for a different purpose (TransitionPagesRePurposedPerSec)",
		nil,
		nil,
	)
	c.writeCopiesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "write_copies_total"),
		"The number of page faults caused by attempting to write that were satisfied by copying the page from elsewhere in physical memory (WriteCopiesPerSec)",
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
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	errs := make([]error, 0, 2)

	if err := c.collectPDH(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting memory metrics: %w", err))
	}

	if err := c.collectGlobalMemoryStatus(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting global memory metrics: %w", err))
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

func (c *Collector) collectPDH(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect Memory metrics: %w", err)
	}

	data, ok := perfData[perfdata.EmptyInstance]

	if !ok {
		return errors.New("perflib query for Memory returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.availableBytes,
		prometheus.GaugeValue,
		data[availableBytes].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.cacheBytes,
		prometheus.GaugeValue,
		data[cacheBytes].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.cacheBytesPeak,
		prometheus.GaugeValue,
		data[cacheBytesPeak].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.cacheFaultsTotal,
		prometheus.CounterValue,
		data[cacheFaultsPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.commitLimit,
		prometheus.GaugeValue,
		data[commitLimit].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.committedBytes,
		prometheus.GaugeValue,
		data[committedBytes].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.demandZeroFaultsTotal,
		prometheus.CounterValue,
		data[demandZeroFaultsPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.freeAndZeroPageListBytes,
		prometheus.GaugeValue,
		data[freeAndZeroPageListBytes].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.freeSystemPageTableEntries,
		prometheus.GaugeValue,
		data[freeSystemPageTableEntries].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.modifiedPageListBytes,
		prometheus.GaugeValue,
		data[modifiedPageListBytes].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.pageFaultsTotal,
		prometheus.CounterValue,
		data[pageFaultsPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.swapPageReadsTotal,
		prometheus.CounterValue,
		data[pageReadsPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.swapPagesReadTotal,
		prometheus.CounterValue,
		data[pagesInputPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.swapPagesWrittenTotal,
		prometheus.CounterValue,
		data[pagesOutputPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.swapPageOperationsTotal,
		prometheus.CounterValue,
		data[pagesPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.swapPageWritesTotal,
		prometheus.CounterValue,
		data[pageWritesPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.poolNonPagedAllocationsTotal,
		prometheus.GaugeValue,
		data[poolNonpagedAllocs].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.poolNonPagedBytes,
		prometheus.GaugeValue,
		data[poolNonpagedBytes].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.poolPagedAllocationsTotal,
		prometheus.CounterValue,
		data[poolPagedAllocs].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.poolPagedBytes,
		prometheus.GaugeValue,
		data[poolPagedBytes].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.poolPagedResidentBytes,
		prometheus.GaugeValue,
		data[poolPagedResidentBytes].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.standbyCacheCoreBytes,
		prometheus.GaugeValue,
		data[standbyCacheCoreBytes].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.standbyCacheNormalPriorityBytes,
		prometheus.GaugeValue,
		data[standbyCacheNormalPriorityBytes].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.standbyCacheReserveBytes,
		prometheus.GaugeValue,
		data[standbyCacheReserveBytes].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.systemCacheResidentBytes,
		prometheus.GaugeValue,
		data[systemCacheResidentBytes].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.systemCodeResidentBytes,
		prometheus.GaugeValue,
		data[systemCodeResidentBytes].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.systemCodeTotalBytes,
		prometheus.GaugeValue,
		data[systemCodeTotalBytes].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.systemDriverResidentBytes,
		prometheus.GaugeValue,
		data[systemDriverResidentBytes].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.systemDriverTotalBytes,
		prometheus.GaugeValue,
		data[systemDriverTotalBytes].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.transitionFaultsTotal,
		prometheus.CounterValue,
		data[transitionFaultsPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.transitionPagesRepurposedTotal,
		prometheus.CounterValue,
		data[transitionPagesRePurposedPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.writeCopiesTotal,
		prometheus.CounterValue,
		data[writeCopiesPerSec].FirstValue,
	)

	return nil
}

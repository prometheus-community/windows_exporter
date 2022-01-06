// returns data points from Win32_PerfRawData_PerfOS_Memory
// <add link to documentation here> - Win32_PerfRawData_PerfOS_Memory class

//go:build windows
// +build windows

package collector

import (
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("memory", NewMemoryCollector, "Memory")
}

// A MemoryCollector is a Prometheus collector for perflib Memory metrics
type MemoryCollector struct {
	AvailableBytes                  *prometheus.Desc
	CacheBytes                      *prometheus.Desc
	CacheBytesPeak                  *prometheus.Desc
	CacheFaultsTotal                *prometheus.Desc
	CommitLimit                     *prometheus.Desc
	CommittedBytes                  *prometheus.Desc
	DemandZeroFaultsTotal           *prometheus.Desc
	FreeAndZeroPageListBytes        *prometheus.Desc
	FreeSystemPageTableEntries      *prometheus.Desc
	ModifiedPageListBytes           *prometheus.Desc
	PageFaultsTotal                 *prometheus.Desc
	SwapPageReadsTotal              *prometheus.Desc
	SwapPagesReadTotal              *prometheus.Desc
	SwapPagesWrittenTotal           *prometheus.Desc
	SwapPageOperationsTotal         *prometheus.Desc
	SwapPageWritesTotal             *prometheus.Desc
	PoolNonpagedAllocsTotal         *prometheus.Desc
	PoolNonpagedBytes               *prometheus.Desc
	PoolPagedAllocsTotal            *prometheus.Desc
	PoolPagedBytes                  *prometheus.Desc
	PoolPagedResidentBytes          *prometheus.Desc
	StandbyCacheCoreBytes           *prometheus.Desc
	StandbyCacheNormalPriorityBytes *prometheus.Desc
	StandbyCacheReserveBytes        *prometheus.Desc
	SystemCacheResidentBytes        *prometheus.Desc
	SystemCodeResidentBytes         *prometheus.Desc
	SystemCodeTotalBytes            *prometheus.Desc
	SystemDriverResidentBytes       *prometheus.Desc
	SystemDriverTotalBytes          *prometheus.Desc
	TransitionFaultsTotal           *prometheus.Desc
	TransitionPagesRepurposedTotal  *prometheus.Desc
	WriteCopiesTotal                *prometheus.Desc
}

// NewMemoryCollector ...
func NewMemoryCollector() (Collector, error) {
	const subsystem = "memory"

	return &MemoryCollector{
		AvailableBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "available_bytes"),
			"The amount of physical memory immediately available for allocation to a process or for system use. It is equal to the sum of memory assigned to"+
				" the standby (cached), free and zero page lists (AvailableBytes)",
			nil,
			nil,
		),
		CacheBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cache_bytes"),
			"(CacheBytes)",
			nil,
			nil,
		),
		CacheBytesPeak: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cache_bytes_peak"),
			"(CacheBytesPeak)",
			nil,
			nil,
		),
		CacheFaultsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cache_faults_total"),
			"Number of faults which occur when a page sought in the file system cache is not found there and must be retrieved from elsewhere in memory (soft fault) "+
				"or from disk (hard fault) (Cache Faults/sec)",
			nil,
			nil,
		),
		CommitLimit: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "commit_limit"),
			"(CommitLimit)",
			nil,
			nil,
		),
		CommittedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "committed_bytes"),
			"(CommittedBytes)",
			nil,
			nil,
		),
		DemandZeroFaultsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "demand_zero_faults_total"),
			"The number of zeroed pages required to satisfy faults. Zeroed pages, pages emptied of previously stored data and filled with zeros, are a security"+
				" feature of Windows that prevent processes from seeing data stored by earlier processes that used the memory space (Demand Zero Faults/sec)",
			nil,
			nil,
		),
		FreeAndZeroPageListBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "free_and_zero_page_list_bytes"),
			"The amount of physical memory, in bytes, that is assigned to the free and zero page lists. This memory does not contain cached data. It is immediately"+
				" available for allocation to a process or for system use (FreeAndZeroPageListBytes)",
			nil,
			nil,
		),
		FreeSystemPageTableEntries: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "free_system_page_table_entries"),
			"(FreeSystemPageTableEntries)",
			nil,
			nil,
		),
		ModifiedPageListBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "modified_page_list_bytes"),
			"The amount of physical memory, in bytes, that is assigned to the modified page list. This memory contains cached data and code that is not actively in "+
				"use by processes, the system and the system cache (ModifiedPageListBytes)",
			nil,
			nil,
		),
		PageFaultsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "page_faults_total"),
			"Overall rate at which faulted pages are handled by the processor (Page Faults/sec)",
			nil,
			nil,
		),
		SwapPageReadsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "swap_page_reads_total"),
			"Number of disk page reads (a single read operation reading several pages is still only counted once) (PageReadsPersec)",
			nil,
			nil,
		),
		SwapPagesReadTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "swap_pages_read_total"),
			"Number of pages read across all page reads (ie counting all pages read even if they are read in a single operation) (PagesInputPersec)",
			nil,
			nil,
		),
		SwapPagesWrittenTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "swap_pages_written_total"),
			"Number of pages written across all page writes (ie counting all pages written even if they are written in a single operation) (PagesOutputPersec)",
			nil,
			nil,
		),
		SwapPageOperationsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "swap_page_operations_total"),
			"Total number of swap page read and writes (PagesPersec)",
			nil,
			nil,
		),
		SwapPageWritesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "swap_page_writes_total"),
			"Number of disk page writes (a single write operation writing several pages is still only counted once) (PageWritesPersec)",
			nil,
			nil,
		),
		PoolNonpagedAllocsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pool_nonpaged_allocs_total"),
			"The number of calls to allocate space in the nonpaged pool. The nonpaged pool is an area of system memory area for objects that cannot be written"+
				" to disk, and must remain in physical memory as long as they are allocated (PoolNonpagedAllocs)",
			nil,
			nil,
		),
		PoolNonpagedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pool_nonpaged_bytes"),
			"Number of bytes in the non-paged pool, an area of the system virtual memory that is used for objects that cannot be written to disk, but must "+
				"remain in physical memory as long as they are allocated (PoolNonpagedBytes)",
			nil,
			nil,
		),
		PoolPagedAllocsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pool_paged_allocs_total"),
			"Number of calls to allocate space in the paged pool, regardless of the amount of space allocated in each call (PoolPagedAllocs)",
			nil,
			nil,
		),
		PoolPagedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pool_paged_bytes"),
			"(PoolPagedBytes)",
			nil,
			nil,
		),
		PoolPagedResidentBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pool_paged_resident_bytes"),
			"The size, in bytes, of the portion of the paged pool that is currently resident and active in physical memory. The paged pool is an area of the "+
				"system virtual memory that is used for objects that can be written to disk when they are not being used (PoolPagedResidentBytes)",
			nil,
			nil,
		),
		StandbyCacheCoreBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "standby_cache_core_bytes"),
			"The amount of physical memory, in bytes, that is assigned to the core standby cache page lists. This memory contains cached data and code that is "+
				"not actively in use by processes, the system and the system cache (StandbyCacheCoreBytes)",
			nil,
			nil,
		),
		StandbyCacheNormalPriorityBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "standby_cache_normal_priority_bytes"),
			"The amount of physical memory, in bytes, that is assigned to the normal priority standby cache page lists. This memory contains cached data and "+
				"code that is not actively in use by processes, the system and the system cache (StandbyCacheNormalPriorityBytes)",
			nil,
			nil,
		),
		StandbyCacheReserveBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "standby_cache_reserve_bytes"),
			"The amount of physical memory, in bytes, that is assigned to the reserve standby cache page lists. This memory contains cached data and code "+
				"that is not actively in use by processes, the system and the system cache (StandbyCacheReserveBytes)",
			nil,
			nil,
		),
		SystemCacheResidentBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "system_cache_resident_bytes"),
			"The size, in bytes, of the portion of the system file cache which is currently resident and active in physical memory (SystemCacheResidentBytes)",
			nil,
			nil,
		),
		SystemCodeResidentBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "system_code_resident_bytes"),
			"The size, in bytes, of the pageable operating system code that is currently resident and active in physical memory (SystemCodeResidentBytes)",
			nil,
			nil,
		),
		SystemCodeTotalBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "system_code_total_bytes"),
			"The size, in bytes, of the pageable operating system code currently mapped into the system virtual address space (SystemCodeTotalBytes)",
			nil,
			nil,
		),
		SystemDriverResidentBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "system_driver_resident_bytes"),
			"The size, in bytes, of the pageable physical memory being used by device drivers. It is the working set (physical memory area) of the drivers (SystemDriverResidentBytes)",
			nil,
			nil,
		),
		SystemDriverTotalBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "system_driver_total_bytes"),
			"The size, in bytes, of the pageable virtual memory currently being used by device drivers. Pageable memory can be written to disk when it is not being used (SystemDriverTotalBytes)",
			nil,
			nil,
		),
		TransitionFaultsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "transition_faults_total"),
			"Number of faults rate at which page faults are resolved by recovering pages that were being used by another process sharing the page, or were on the "+
				"modified page list or the standby list, or were being written to disk at the time of the page fault (TransitionFaultsPersec)",
			nil,
			nil,
		),
		TransitionPagesRepurposedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "transition_pages_repurposed_total"),
			"Transition Pages RePurposed is the rate at which the number of transition cache pages were reused for a different purpose (TransitionPagesRePurposedPersec)",
			nil,
			nil,
		),
		WriteCopiesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "write_copies_total"),
			"The number of page faults caused by attempting to write that were satisfied by copying the page from elsewhere in physical memory (WriteCopiesPersec)",
			nil,
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *MemoryCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ctx, ch); err != nil {
		log.Error("failed collecting memory metrics:", desc, err)
		return err
	}
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

func (c *MemoryCollector) collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []memory
	if err := unmarshalObject(ctx.perfObjects["Memory"], &dst); err != nil {
		return nil, err
	}

	ch <- prometheus.MustNewConstMetric(
		c.AvailableBytes,
		prometheus.GaugeValue,
		dst[0].AvailableBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.CacheBytes,
		prometheus.GaugeValue,
		dst[0].CacheBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.CacheBytesPeak,
		prometheus.GaugeValue,
		dst[0].CacheBytesPeak,
	)

	ch <- prometheus.MustNewConstMetric(
		c.CacheFaultsTotal,
		prometheus.CounterValue,
		dst[0].CacheFaultsPersec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.CommitLimit,
		prometheus.GaugeValue,
		dst[0].CommitLimit,
	)

	ch <- prometheus.MustNewConstMetric(
		c.CommittedBytes,
		prometheus.GaugeValue,
		dst[0].CommittedBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.DemandZeroFaultsTotal,
		prometheus.CounterValue,
		dst[0].DemandZeroFaultsPersec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.FreeAndZeroPageListBytes,
		prometheus.GaugeValue,
		dst[0].FreeAndZeroPageListBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.FreeSystemPageTableEntries,
		prometheus.GaugeValue,
		dst[0].FreeSystemPageTableEntries,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ModifiedPageListBytes,
		prometheus.GaugeValue,
		dst[0].ModifiedPageListBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.PageFaultsTotal,
		prometheus.CounterValue,
		dst[0].PageFaultsPersec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.SwapPageReadsTotal,
		prometheus.CounterValue,
		dst[0].PageReadsPersec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.SwapPagesReadTotal,
		prometheus.CounterValue,
		dst[0].PagesInputPersec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.SwapPagesWrittenTotal,
		prometheus.CounterValue,
		dst[0].PagesOutputPersec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.SwapPageOperationsTotal,
		prometheus.CounterValue,
		dst[0].PagesPersec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.SwapPageWritesTotal,
		prometheus.CounterValue,
		dst[0].PageWritesPersec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.PoolNonpagedAllocsTotal,
		prometheus.GaugeValue,
		dst[0].PoolNonpagedAllocs,
	)

	ch <- prometheus.MustNewConstMetric(
		c.PoolNonpagedBytes,
		prometheus.GaugeValue,
		dst[0].PoolNonpagedBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.PoolPagedAllocsTotal,
		prometheus.CounterValue,
		dst[0].PoolPagedAllocs,
	)

	ch <- prometheus.MustNewConstMetric(
		c.PoolPagedBytes,
		prometheus.GaugeValue,
		dst[0].PoolPagedBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.PoolPagedResidentBytes,
		prometheus.GaugeValue,
		dst[0].PoolPagedResidentBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.StandbyCacheCoreBytes,
		prometheus.GaugeValue,
		dst[0].StandbyCacheCoreBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.StandbyCacheNormalPriorityBytes,
		prometheus.GaugeValue,
		dst[0].StandbyCacheNormalPriorityBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.StandbyCacheReserveBytes,
		prometheus.GaugeValue,
		dst[0].StandbyCacheReserveBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.SystemCacheResidentBytes,
		prometheus.GaugeValue,
		dst[0].SystemCacheResidentBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.SystemCodeResidentBytes,
		prometheus.GaugeValue,
		dst[0].SystemCodeResidentBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.SystemCodeTotalBytes,
		prometheus.GaugeValue,
		dst[0].SystemCodeTotalBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.SystemDriverResidentBytes,
		prometheus.GaugeValue,
		dst[0].SystemDriverResidentBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.SystemDriverTotalBytes,
		prometheus.GaugeValue,
		dst[0].SystemDriverTotalBytes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.TransitionFaultsTotal,
		prometheus.CounterValue,
		dst[0].TransitionFaultsPersec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.TransitionPagesRepurposedTotal,
		prometheus.CounterValue,
		dst[0].TransitionPagesRePurposedPersec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.WriteCopiesTotal,
		prometheus.CounterValue,
		dst[0].WriteCopiesPersec,
	)

	return nil, nil
}

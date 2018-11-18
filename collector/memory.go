// returns data points from Win32_PerfRawData_PerfOS_Memory
// <add link to documentation here> - Win32_PerfRawData_PerfOS_Memory class
package collector

import (
	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	Factories["memory"] = NewMemoryCollector
}

// A MemoryCollector is a Prometheus collector for WMI Win32_PerfRawData_PerfOS_Memory metrics
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
			"(CacheFaultsPersec)",
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
				" feature of Windows that prevent processes from seeing data stored by earlier processes that used the memory space (DemandZeroFaults)",
			nil,
			nil,
		),
		FreeAndZeroPageListBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "free_and_zero_page_list_bytes"),
			"(FreeAndZeroPageListBytes)",
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
			"(ModifiedPageListBytes)",
			nil,
			nil,
		),
		PageFaultsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "page_faults_total"),
			"(PageFaultsPersec)",
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
			prometheus.BuildFQName(Namespace, subsystem, "pool_nonpaged_bytes_total"),
			"(PoolNonpagedBytes)",
			nil,
			nil,
		),
		PoolPagedAllocsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pool_paged_allocs_total"),
			"(PoolPagedAllocs)",
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
			"(PoolPagedResidentBytes)",
			nil,
			nil,
		),
		StandbyCacheCoreBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "standby_cache_core_bytes"),
			"(StandbyCacheCoreBytes)",
			nil,
			nil,
		),
		StandbyCacheNormalPriorityBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "standby_cache_normal_priority_bytes"),
			"(StandbyCacheNormalPriorityBytes)",
			nil,
			nil,
		),
		StandbyCacheReserveBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "standby_cache_reserve_bytes"),
			"(StandbyCacheReserveBytes)",
			nil,
			nil,
		),
		SystemCacheResidentBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "system_cache_resident_bytes"),
			"(SystemCacheResidentBytes)",
			nil,
			nil,
		),
		SystemCodeResidentBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "system_code_resident_bytes"),
			"(SystemCodeResidentBytes)",
			nil,
			nil,
		),
		SystemCodeTotalBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "system_code_total_bytes"),
			"(SystemCodeTotalBytes)",
			nil,
			nil,
		),
		SystemDriverResidentBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "system_driver_resident_bytes"),
			"(SystemDriverResidentBytes)",
			nil,
			nil,
		),
		SystemDriverTotalBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "system_driver_total_bytes"),
			"(SystemDriverTotalBytes)",
			nil,
			nil,
		),
		TransitionFaultsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "transition_faults_total"),
			"(TransitionFaultsPersec)",
			nil,
			nil,
		),
		TransitionPagesRepurposedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "transition_pages_repurposed_total"),
			"(TransitionPagesRePurposedPersec)",
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
func (c *MemoryCollector) Collect(ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting memory metrics:", desc, err)
		return err
	}
	return nil
}

type Win32_PerfRawData_PerfOS_Memory struct {
	AvailableBytes                  uint64
	AvailableKBytes                 uint64
	AvailableMBytes                 uint64
	CacheBytes                      uint64
	CacheBytesPeak                  uint64
	CacheFaultsPersec               uint32
	CommitLimit                     uint64
	CommittedBytes                  uint64
	DemandZeroFaultsPersec          uint32
	FreeAndZeroPageListBytes        uint64
	FreeSystemPageTableEntries      uint32
	ModifiedPageListBytes           uint64
	PageFaultsPersec                uint32
	PageReadsPersec                 uint32
	PagesInputPersec                uint32
	PagesOutputPersec               uint32
	PagesPersec                     uint32
	PageWritesPersec                uint32
	PoolNonpagedAllocs              uint32
	PoolNonpagedBytes               uint64
	PoolPagedAllocs                 uint32
	PoolPagedBytes                  uint64
	PoolPagedResidentBytes          uint64
	StandbyCacheCoreBytes           uint64
	StandbyCacheNormalPriorityBytes uint64
	StandbyCacheReserveBytes        uint64
	SystemCacheResidentBytes        uint64
	SystemCodeResidentBytes         uint64
	SystemCodeTotalBytes            uint64
	SystemDriverResidentBytes       uint64
	SystemDriverTotalBytes          uint64
	TransitionFaultsPersec          uint32
	TransitionPagesRePurposedPersec uint32
	WriteCopiesPersec               uint32
}

func (c *MemoryCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_PerfOS_Memory
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	ch <- prometheus.MustNewConstMetric(
		c.AvailableBytes,
		prometheus.GaugeValue,
		float64(dst[0].AvailableBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.CacheBytes,
		prometheus.GaugeValue,
		float64(dst[0].CacheBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.CacheBytesPeak,
		prometheus.GaugeValue,
		float64(dst[0].CacheBytesPeak),
	)

	ch <- prometheus.MustNewConstMetric(
		c.CacheFaultsTotal,
		prometheus.GaugeValue,
		float64(dst[0].CacheFaultsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.CommitLimit,
		prometheus.GaugeValue,
		float64(dst[0].CommitLimit),
	)

	ch <- prometheus.MustNewConstMetric(
		c.CommittedBytes,
		prometheus.GaugeValue,
		float64(dst[0].CommittedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.DemandZeroFaultsTotal,
		prometheus.GaugeValue,
		float64(dst[0].DemandZeroFaultsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.FreeAndZeroPageListBytes,
		prometheus.GaugeValue,
		float64(dst[0].FreeAndZeroPageListBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.FreeSystemPageTableEntries,
		prometheus.GaugeValue,
		float64(dst[0].FreeSystemPageTableEntries),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ModifiedPageListBytes,
		prometheus.GaugeValue,
		float64(dst[0].ModifiedPageListBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PageFaultsTotal,
		prometheus.GaugeValue,
		float64(dst[0].PageFaultsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SwapPageReadsTotal,
		prometheus.GaugeValue,
		float64(dst[0].PageReadsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SwapPagesReadTotal,
		prometheus.GaugeValue,
		float64(dst[0].PagesInputPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SwapPagesWrittenTotal,
		prometheus.GaugeValue,
		float64(dst[0].PagesOutputPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SwapPageOperationsTotal,
		prometheus.GaugeValue,
		float64(dst[0].PagesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SwapPageWritesTotal,
		prometheus.GaugeValue,
		float64(dst[0].PageWritesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PoolNonpagedAllocsTotal,
		prometheus.GaugeValue,
		float64(dst[0].PoolNonpagedAllocs),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PoolNonpagedBytes,
		prometheus.GaugeValue,
		float64(dst[0].PoolNonpagedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PoolPagedAllocsTotal,
		prometheus.GaugeValue,
		float64(dst[0].PoolPagedAllocs),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PoolPagedBytes,
		prometheus.GaugeValue,
		float64(dst[0].PoolPagedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PoolPagedResidentBytes,
		prometheus.GaugeValue,
		float64(dst[0].PoolPagedResidentBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.StandbyCacheCoreBytes,
		prometheus.GaugeValue,
		float64(dst[0].StandbyCacheCoreBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.StandbyCacheNormalPriorityBytes,
		prometheus.GaugeValue,
		float64(dst[0].StandbyCacheNormalPriorityBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.StandbyCacheReserveBytes,
		prometheus.GaugeValue,
		float64(dst[0].StandbyCacheReserveBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SystemCacheResidentBytes,
		prometheus.GaugeValue,
		float64(dst[0].SystemCacheResidentBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SystemCodeResidentBytes,
		prometheus.GaugeValue,
		float64(dst[0].SystemCodeResidentBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SystemCodeTotalBytes,
		prometheus.GaugeValue,
		float64(dst[0].SystemCodeTotalBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SystemDriverResidentBytes,
		prometheus.GaugeValue,
		float64(dst[0].SystemDriverResidentBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SystemDriverTotalBytes,
		prometheus.GaugeValue,
		float64(dst[0].SystemDriverTotalBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.TransitionFaultsTotal,
		prometheus.GaugeValue,
		float64(dst[0].TransitionFaultsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.TransitionPagesRepurposedTotal,
		prometheus.GaugeValue,
		float64(dst[0].TransitionPagesRePurposedPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.WriteCopiesTotal,
		prometheus.GaugeValue,
		float64(dst[0].WriteCopiesPersec),
	)

	return nil, nil
}

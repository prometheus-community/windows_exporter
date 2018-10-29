// returns data points from Win32_PerfRawData_PerfOS_Memory
// <add link to documentation here> - Win32_PerfRawData_PerfOS_Memory class
package collector

import (
	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"fmt"
)

func init() {
	Factories["os_memory"] = NewOS_MemoryCollector
}

// A OS_MemoryCollector is a Prometheus collector for WMI Win32_PerfRawData_PerfOS_Memory metrics
type OS_MemoryCollector struct {
	AvailableBytes                  *prometheus.Desc
	AvailableKBytes                 *prometheus.Desc
	AvailableMBytes                 *prometheus.Desc
	CacheBytes                      *prometheus.Desc
	CacheBytesPeak                  *prometheus.Desc
	CacheFaultsPersec               *prometheus.Desc
	CommitLimit                     *prometheus.Desc
	CommittedBytes                  *prometheus.Desc
	DemandZeroFaultsPersec          *prometheus.Desc
	FreeAndZeroPageListBytes        *prometheus.Desc
	FreeSystemPageTableEntries      *prometheus.Desc
	ModifiedPageListBytes           *prometheus.Desc
	PageFaultsPersec                *prometheus.Desc
	PageReadsPersec                 *prometheus.Desc
	PagesInputPersec                *prometheus.Desc
	PagesOutputPersec               *prometheus.Desc
	PagesPersec                     *prometheus.Desc
	PageWritesPersec                *prometheus.Desc
	PercentCommittedBytesInUse      *prometheus.Desc
	PoolNonpagedAllocs              *prometheus.Desc
	PoolNonpagedBytes               *prometheus.Desc
	PoolPagedAllocs                 *prometheus.Desc
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
	TransitionFaultsPersec          *prometheus.Desc
	TransitionPagesRePurposedPersec *prometheus.Desc
	WriteCopiesPersec               *prometheus.Desc
}

// NewOS_MemoryCollector ...
func NewOS_MemoryCollector() (Collector, error) {
	const subsystem = "os_memory"

	// for debugging only!
	fmt.Println(Namespace)
	fmt.Println(subsystem)

	return &OS_MemoryCollector{
		AvailableBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "available_bytes"),
			"(AvailableBytes)",
			nil,
			nil,
		),
		AvailableKBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "available_k_bytes"),
			"(AvailableKBytes)",
			nil,
			nil,
		),
		AvailableMBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "available_m_bytes"),
			"(AvailableMBytes)",
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
		CacheFaultsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cache_faults_persec"),
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
		DemandZeroFaultsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "demand_zero_faults_persec"),
			"(DemandZeroFaultsPersec)",
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
		PageFaultsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "page_faults_persec"),
			"(PageFaultsPersec)",
			nil,
			nil,
		),
		PageReadsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "page_reads_persec"),
			"(PageReadsPersec)",
			nil,
			nil,
		),
		PagesInputPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pages_input_persec"),
			"(PagesInputPersec)",
			nil,
			nil,
		),
		PagesOutputPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pages_output_persec"),
			"(PagesOutputPersec)",
			nil,
			nil,
		),
		PagesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pages_persec"),
			"(PagesPersec)",
			nil,
			nil,
		),
		PageWritesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "page_writes_persec"),
			"(PageWritesPersec)",
			nil,
			nil,
		),
		PercentCommittedBytesInUse: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "percent_committed_bytes_in_use"),
			"(PercentCommittedBytesInUse)",
			nil,
			nil,
		),
		PoolNonpagedAllocs: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pool_nonpaged_allocs"),
			"(PoolNonpagedAllocs)",
			nil,
			nil,
		),
		PoolNonpagedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pool_nonpaged_bytes"),
			"(PoolNonpagedBytes)",
			nil,
			nil,
		),
		PoolPagedAllocs: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pool_paged_allocs"),
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
		TransitionFaultsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "transition_faults_persec"),
			"(TransitionFaultsPersec)",
			nil,
			nil,
		),
		TransitionPagesRePurposedPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "transition_pages_re_purposed_persec"),
			"(TransitionPagesRePurposedPersec)",
			nil,
			nil,
		),
		WriteCopiesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "write_copies_persec"),
			"(WriteCopiesPersec)",
			nil,
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *OS_MemoryCollector) Collect(ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting os_memory metrics:", desc, err)
		return err
	}
	return nil
}

type Win32_PerfRawData_PerfOS_Memory struct {
	Name string

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
	PercentCommittedBytesInUse      uint32
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

func (c *OS_MemoryCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
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
		c.AvailableKBytes,
		prometheus.GaugeValue,
		float64(dst[0].AvailableKBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AvailableMBytes,
		prometheus.GaugeValue,
		float64(dst[0].AvailableMBytes),
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
		c.CacheFaultsPersec,
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
		c.DemandZeroFaultsPersec,
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
		c.PageFaultsPersec,
		prometheus.GaugeValue,
		float64(dst[0].PageFaultsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PageReadsPersec,
		prometheus.GaugeValue,
		float64(dst[0].PageReadsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PagesInputPersec,
		prometheus.GaugeValue,
		float64(dst[0].PagesInputPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PagesOutputPersec,
		prometheus.GaugeValue,
		float64(dst[0].PagesOutputPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PagesPersec,
		prometheus.GaugeValue,
		float64(dst[0].PagesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PageWritesPersec,
		prometheus.GaugeValue,
		float64(dst[0].PageWritesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PercentCommittedBytesInUse,
		prometheus.GaugeValue,
		float64(dst[0].PercentCommittedBytesInUse),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PoolNonpagedAllocs,
		prometheus.GaugeValue,
		float64(dst[0].PoolNonpagedAllocs),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PoolNonpagedBytes,
		prometheus.GaugeValue,
		float64(dst[0].PoolNonpagedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PoolPagedAllocs,
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
		c.TransitionFaultsPersec,
		prometheus.GaugeValue,
		float64(dst[0].TransitionFaultsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.TransitionPagesRePurposedPersec,
		prometheus.GaugeValue,
		float64(dst[0].TransitionPagesRePurposedPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.WriteCopiesPersec,
		prometheus.GaugeValue,
		float64(dst[0].WriteCopiesPersec),
	)

	return nil, nil
}

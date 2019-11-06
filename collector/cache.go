package collector

import (
	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	Factories["cache"] = NewCacheCollector
}

// A CacheCollector is a Prometheus collector for WMI Win32_PerfFormattedData_PerfOS_Cache  metrics
type CacheCollector struct {
	AsyncCopyReadsPersec         *prometheus.Desc
	AsyncDataMapsPersec          *prometheus.Desc
	AsyncFastReadsPersec         *prometheus.Desc
	AsyncMDLReadsPersec          *prometheus.Desc
	AsyncPinReadsPersec          *prometheus.Desc
	CopyReadHitsPercent          *prometheus.Desc
	CopyReadsPersec              *prometheus.Desc
	DataFlushesPersec            *prometheus.Desc
	DataFlushPagesPersec         *prometheus.Desc
	DataMapHitsPercent           *prometheus.Desc
	DataMapPinsPersec            *prometheus.Desc
	DataMapsPersec               *prometheus.Desc
	DirtyPages                   *prometheus.Desc
	DirtyPageThreshold           *prometheus.Desc
	FastReadNotPossiblesPersec   *prometheus.Desc
	FastReadResourceMissesPersec *prometheus.Desc
	FastReadsPersec              *prometheus.Desc
	LazyWriteFlushesPersec       *prometheus.Desc
	LazyWritePagesPersec         *prometheus.Desc
	MDLReadHitsPercent           *prometheus.Desc
	MDLReadsPersec               *prometheus.Desc
	PinReadHitsPercent           *prometheus.Desc
	PinReadsPersec               *prometheus.Desc
	ReadAheadsPersec             *prometheus.Desc
	SyncCopyReadsPersec          *prometheus.Desc
	SyncDataMapsPersec           *prometheus.Desc
	SyncFastReadsPersec          *prometheus.Desc
	SyncMDLReadsPersec           *prometheus.Desc
	SyncPinReadsPersec           *prometheus.Desc
}

// NewCacheCollector ...
func NewCacheCollector() (Collector, error) {
	const subsystem = "cache"
	return &CacheCollector{
		AsyncCopyReadsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "async_copy_reads_persec"),
			"(AsyncCopyReadsPersec)",
			nil,
			nil,
		),
		AsyncDataMapsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "async_data_maps_persec"),
			"(AsyncDataMapsPersec)",
			nil,
			nil,
		),
		AsyncFastReadsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "async_fast_reads_persec"),
			"(AsyncFastReadsPersec)",
			nil,
			nil,
		),
		AsyncMDLReadsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "async_mdl_reads_persec"),
			"(AsyncMDLReadsPersec)",
			nil,
			nil,
		),
		AsyncPinReadsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "async_pin_reads_persec"),
			"(AsyncPinReadsPersec)",
			nil,
			nil,
		),
		CopyReadHitsPercent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "copy_read_hits_percent"),
			"(CopyReadHitsPercent)",
			nil,
			nil,
		),
		CopyReadsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "copy_reads_persec"),
			"(CopyReadsPersec)",
			nil,
			nil,
		),
		DataFlushesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "data_flushes_persec"),
			"(DataFlushesPersec)",
			nil,
			nil,
		),
		DataFlushPagesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "data_flush_pages_persec"),
			"(DataFlushPagesPersec)",
			nil,
			nil,
		),
		DataMapHitsPercent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "data_map_hits_percent"),
			"(DataMapHitsPercent)",
			nil,
			nil,
		),
		DataMapPinsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "data_map_pins_persec"),
			"(DataMapPinsPersec)",
			nil,
			nil,
		),
		DataMapsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "data_maps_persec"),
			"(DataMapsPersec)",
			nil,
			nil,
		),
		DirtyPages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dirty_pages"),
			"(DirtyPages)",
			nil,
			nil,
		),
		DirtyPageThreshold: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dirty_page_threshold"),
			"(DirtyPageThreshold)",
			nil,
			nil,
		),
		FastReadNotPossiblesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "fast_read_not_possibles_persec"),
			"(FastReadNotPossiblesPersec)",
			nil,
			nil,
		),
		FastReadResourceMissesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "fast_read_resource_misses_persec"),
			"(FastReadResourceMissesPersec)",
			nil,
			nil,
		),
		FastReadsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "fast_reads_persec"),
			"(FastReadsPersec)",
			nil,
			nil,
		),
		LazyWriteFlushesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lazy_write_flushes_persec"),
			"(LazyWriteFlushesPersec)",
			nil,
			nil,
		),
		LazyWritePagesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lazy_write_pages_persec"),
			"(LazyWritePagesPersec)",
			nil,
			nil,
		),
		MDLReadHitsPercent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "mdl_read_hits_percent"),
			"(MDLReadHitsPercent)",
			nil,
			nil,
		),
		MDLReadsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "mdl_reads_persec"),
			"(MDLReadsPersec)",
			nil,
			nil,
		),
		PinReadHitsPercent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pin_read_hits_percent"),
			"(PinReadHitsPercent)",
			nil,
			nil,
		),
		PinReadsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pin_reads_persec"),
			"(PinReadsPersec)",
			nil,
			nil,
		),
		ReadAheadsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "read_aheads_persec"),
			"(ReadAheadsPersec)",
			nil,
			nil,
		),
		SyncCopyReadsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sync_copy_reads_persec"),
			"(SyncCopyReadsPersec)",
			nil,
			nil,
		),
		SyncDataMapsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sync_data_maps_persec"),
			"(SyncDataMapsPersec)",
			nil,
			nil,
		),
		SyncFastReadsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sync_fast_reads_persec"),
			"(SyncFastReadsPersec)",
			nil,
			nil,
		),
		SyncMDLReadsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sync_mdl_reads_persec"),
			"(SyncMDLReadsPersec)",
			nil,
			nil,
		),
		SyncPinReadsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sync_pin_reads_persec"),
			"(SyncPinReadsPersec)",
			nil,
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *CacheCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting cache metrics:", desc, err)
		return err
	}
	return nil
}

// Win32_PerfFormattedData_PerfOS_Cache  docs:
// - https://docs.microsoft.com/en-us/previous-versions/aa394267(v=vs.85)
type Win32_PerfFormattedData_PerfOS_Cache struct {
	AsyncCopyReadsPersec         uint32
	AsyncDataMapsPersec          uint32
	AsyncFastReadsPersec         uint32
	AsyncMDLReadsPersec          uint32
	AsyncPinReadsPersec          uint32
	CopyReadHitsPercent          uint32
	CopyReadsPersec              uint32
	DataFlushesPersec            uint32
	DataFlushPagesPersec         uint32
	DataMapHitsPercent           uint32
	DataMapPinsPersec            uint32
	DataMapsPersec               uint32
	DirtyPages                   uint64
	DirtyPageThreshold           uint64
	FastReadNotPossiblesPersec   uint32
	FastReadResourceMissesPersec uint32
	FastReadsPersec              uint32
	LazyWriteFlushesPersec       uint32
	LazyWritePagesPersec         uint32
	MDLReadHitsPercent           uint32
	MDLReadsPersec               uint32
	PinReadHitsPercent           uint32
	PinReadsPersec               uint32
	ReadAheadsPersec             uint32
	SyncCopyReadsPersec          uint32
	SyncDataMapsPersec           uint32
	SyncFastReadsPersec          uint32
	SyncMDLReadsPersec           uint32
	SyncPinReadsPersec           uint32
}

func (c *CacheCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfFormattedData_PerfOS_Cache
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	ch <- prometheus.MustNewConstMetric(
		c.AsyncCopyReadsPersec,
		prometheus.GaugeValue,
		float64(dst[0].AsyncCopyReadsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AsyncDataMapsPersec,
		prometheus.GaugeValue,
		float64(dst[0].AsyncDataMapsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AsyncFastReadsPersec,
		prometheus.GaugeValue,
		float64(dst[0].AsyncFastReadsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AsyncMDLReadsPersec,
		prometheus.GaugeValue,
		float64(dst[0].AsyncMDLReadsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AsyncPinReadsPersec,
		prometheus.GaugeValue,
		float64(dst[0].AsyncPinReadsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.CopyReadHitsPercent,
		prometheus.GaugeValue,
		float64(dst[0].CopyReadHitsPercent),
	)

	ch <- prometheus.MustNewConstMetric(
		c.CopyReadsPersec,
		prometheus.GaugeValue,
		float64(dst[0].CopyReadsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.DataFlushesPersec,
		prometheus.GaugeValue,
		float64(dst[0].DataFlushesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.DataFlushPagesPersec,
		prometheus.GaugeValue,
		float64(dst[0].DataFlushPagesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.DataMapHitsPercent,
		prometheus.GaugeValue,
		float64(dst[0].DataMapHitsPercent),
	)

	ch <- prometheus.MustNewConstMetric(
		c.DataMapPinsPersec,
		prometheus.GaugeValue,
		float64(dst[0].DataMapPinsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.DataMapsPersec,
		prometheus.GaugeValue,
		float64(dst[0].DataMapsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.DirtyPages,
		prometheus.GaugeValue,
		float64(dst[0].DirtyPages),
	)

	ch <- prometheus.MustNewConstMetric(
		c.DirtyPageThreshold,
		prometheus.GaugeValue,
		float64(dst[0].DirtyPageThreshold),
	)

	ch <- prometheus.MustNewConstMetric(
		c.FastReadNotPossiblesPersec,
		prometheus.GaugeValue,
		float64(dst[0].FastReadNotPossiblesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.FastReadResourceMissesPersec,
		prometheus.GaugeValue,
		float64(dst[0].FastReadResourceMissesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.FastReadsPersec,
		prometheus.GaugeValue,
		float64(dst[0].FastReadsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LazyWriteFlushesPersec,
		prometheus.GaugeValue,
		float64(dst[0].LazyWriteFlushesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LazyWritePagesPersec,
		prometheus.GaugeValue,
		float64(dst[0].LazyWritePagesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.MDLReadHitsPercent,
		prometheus.GaugeValue,
		float64(dst[0].MDLReadHitsPercent),
	)

	ch <- prometheus.MustNewConstMetric(
		c.MDLReadsPersec,
		prometheus.GaugeValue,
		float64(dst[0].MDLReadsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PinReadHitsPercent,
		prometheus.GaugeValue,
		float64(dst[0].PinReadHitsPercent),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PinReadsPersec,
		prometheus.GaugeValue,
		float64(dst[0].PinReadsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ReadAheadsPersec,
		prometheus.GaugeValue,
		float64(dst[0].ReadAheadsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SyncCopyReadsPersec,
		prometheus.GaugeValue,
		float64(dst[0].SyncCopyReadsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SyncDataMapsPersec,
		prometheus.GaugeValue,
		float64(dst[0].SyncDataMapsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SyncFastReadsPersec,
		prometheus.GaugeValue,
		float64(dst[0].SyncFastReadsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SyncMDLReadsPersec,
		prometheus.GaugeValue,
		float64(dst[0].SyncMDLReadsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SyncPinReadsPersec,
		prometheus.GaugeValue,
		float64(dst[0].SyncPinReadsPersec),
	)

	return nil, nil
}

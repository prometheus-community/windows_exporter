//go:build windows
// +build windows

package collector

import (
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("cache", newCacheCollector, "Cache")
}

// A CacheCollector is a Prometheus collector for Perflib Cache metrics
type CacheCollector struct {
	AsyncCopyReadsTotal         *prometheus.Desc
	AsyncDataMapsTotal          *prometheus.Desc
	AsyncFastReadsTotal         *prometheus.Desc
	AsyncMDLReadsTotal          *prometheus.Desc
	AsyncPinReadsTotal          *prometheus.Desc
	CopyReadHitsTotal           *prometheus.Desc
	CopyReadsTotal              *prometheus.Desc
	DataFlushesTotal            *prometheus.Desc
	DataFlushPagesTotal         *prometheus.Desc
	DataMapHitsPercent          *prometheus.Desc
	DataMapPinsTotal            *prometheus.Desc
	DataMapsTotal               *prometheus.Desc
	DirtyPages                  *prometheus.Desc
	DirtyPageThreshold          *prometheus.Desc
	FastReadNotPossiblesTotal   *prometheus.Desc
	FastReadResourceMissesTotal *prometheus.Desc
	FastReadsTotal              *prometheus.Desc
	LazyWriteFlushesTotal       *prometheus.Desc
	LazyWritePagesTotal         *prometheus.Desc
	MDLReadHitsTotal            *prometheus.Desc
	MDLReadsTotal               *prometheus.Desc
	PinReadHitsTotal            *prometheus.Desc
	PinReadsTotal               *prometheus.Desc
	ReadAheadsTotal             *prometheus.Desc
	SyncCopyReadsTotal          *prometheus.Desc
	SyncDataMapsTotal           *prometheus.Desc
	SyncFastReadsTotal          *prometheus.Desc
	SyncMDLReadsTotal           *prometheus.Desc
	SyncPinReadsTotal           *prometheus.Desc
}

// NewCacheCollector ...
func newCacheCollector() (Collector, error) {
	const subsystem = "cache"
	return &CacheCollector{
		AsyncCopyReadsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "async_copy_reads_total"),
			"(AsyncCopyReadsTotal)",
			nil,
			nil,
		),
		AsyncDataMapsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "async_data_maps_total"),
			"(AsyncDataMapsTotal)",
			nil,
			nil,
		),
		AsyncFastReadsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "async_fast_reads_total"),
			"(AsyncFastReadsTotal)",
			nil,
			nil,
		),
		AsyncMDLReadsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "async_mdl_reads_total"),
			"(AsyncMDLReadsTotal)",
			nil,
			nil,
		),
		AsyncPinReadsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "async_pin_reads_total"),
			"(AsyncPinReadsTotal)",
			nil,
			nil,
		),
		CopyReadHitsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "copy_read_hits_total"),
			"(CopyReadHitsTotal)",
			nil,
			nil,
		),
		CopyReadsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "copy_reads_total"),
			"(CopyReadsTotal)",
			nil,
			nil,
		),
		DataFlushesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "data_flushes_total"),
			"(DataFlushesTotal)",
			nil,
			nil,
		),
		DataFlushPagesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "data_flush_pages_total"),
			"(DataFlushPagesTotal)",
			nil,
			nil,
		),
		DataMapHitsPercent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "data_map_hits_percent"),
			"(DataMapHitsPercent)",
			nil,
			nil,
		),
		DataMapPinsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "data_map_pins_total"),
			"(DataMapPinsTotal)",
			nil,
			nil,
		),
		DataMapsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "data_maps_total"),
			"(DataMapsTotal)",
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
		FastReadNotPossiblesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "fast_read_not_possibles_total"),
			"(FastReadNotPossiblesTotal)",
			nil,
			nil,
		),
		FastReadResourceMissesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "fast_read_resource_misses_total"),
			"(FastReadResourceMissesTotal)",
			nil,
			nil,
		),
		FastReadsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "fast_reads_total"),
			"(FastReadsTotal)",
			nil,
			nil,
		),
		LazyWriteFlushesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lazy_write_flushes_total"),
			"(LazyWriteFlushesTotal)",
			nil,
			nil,
		),
		LazyWritePagesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lazy_write_pages_total"),
			"(LazyWritePagesTotal)",
			nil,
			nil,
		),
		MDLReadHitsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "mdl_read_hits_total"),
			"(MDLReadHitsTotal)",
			nil,
			nil,
		),
		MDLReadsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "mdl_reads_total"),
			"(MDLReadsTotal)",
			nil,
			nil,
		),
		PinReadHitsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pin_read_hits_total"),
			"(PinReadHitsTotal)",
			nil,
			nil,
		),
		PinReadsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pin_reads_total"),
			"(PinReadsTotal)",
			nil,
			nil,
		),
		ReadAheadsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "read_aheads_total"),
			"(ReadAheadsTotal)",
			nil,
			nil,
		),
		SyncCopyReadsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sync_copy_reads_total"),
			"(SyncCopyReadsTotal)",
			nil,
			nil,
		),
		SyncDataMapsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sync_data_maps_total"),
			"(SyncDataMapsTotal)",
			nil,
			nil,
		),
		SyncFastReadsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sync_fast_reads_total"),
			"(SyncFastReadsTotal)",
			nil,
			nil,
		),
		SyncMDLReadsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sync_mdl_reads_total"),
			"(SyncMDLReadsTotal)",
			nil,
			nil,
		),
		SyncPinReadsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sync_pin_reads_total"),
			"(SyncPinReadsTotal)",
			nil,
			nil,
		),
	}, nil
}

// Collect implements the Collector interface
func (c *CacheCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ctx, ch); err != nil {
		log.Error("failed collecting cache metrics:", desc, err)
		return err
	}
	return nil
}

// Perflib "Cache":
// - https://docs.microsoft.com/en-us/previous-versions/aa394267(v=vs.85)
type perflibCache struct {
	AsyncCopyReadsTotal         float64 `perflib:"Async Copy Reads/sec"`
	AsyncDataMapsTotal          float64 `perflib:"Async Data Maps/sec"`
	AsyncFastReadsTotal         float64 `perflib:"Async Fast Reads/sec"`
	AsyncMDLReadsTotal          float64 `perflib:"Async MDL Reads/sec"`
	AsyncPinReadsTotal          float64 `perflib:"Async Pin Reads/sec"`
	CopyReadHitsTotal           float64 `perflib:"Copy Read Hits %"`
	CopyReadsTotal              float64 `perflib:"Copy Reads/sec"`
	DataFlushesTotal            float64 `perflib:"Data Flushes/sec"`
	DataFlushPagesTotal         float64 `perflib:"Data Flush Pages/sec"`
	DataMapHitsPercent          float64 `perflib:"Data Map Hits %"`
	DataMapPinsTotal            float64 `perflib:"Data Map Pins/sec"`
	DataMapsTotal               float64 `perflib:"Data Maps/sec"`
	DirtyPages                  float64 `perflib:"Dirty Pages"`
	DirtyPageThreshold          float64 `perflib:"Dirty Page Threshold"`
	FastReadNotPossiblesTotal   float64 `perflib:"Fast Read Not Possibles/sec"`
	FastReadResourceMissesTotal float64 `perflib:"Fast Read Resource Misses/sec"`
	FastReadsTotal              float64 `perflib:"Fast Reads/sec"`
	LazyWriteFlushesTotal       float64 `perflib:"Lazy Write Flushes/sec"`
	LazyWritePagesTotal         float64 `perflib:"Lazy Write Pages/sec"`
	MDLReadHitsTotal            float64 `perflib:"MDL Read Hits %"`
	MDLReadsTotal               float64 `perflib:"MDL Reads/sec"`
	PinReadHitsTotal            float64 `perflib:"Pin Read Hits %"`
	PinReadsTotal               float64 `perflib:"Pin Reads/sec"`
	ReadAheadsTotal             float64 `perflib:"Read Aheads/sec"`
	SyncCopyReadsTotal          float64 `perflib:"Sync Copy Reads/sec"`
	SyncDataMapsTotal           float64 `perflib:"Sync Data Maps/sec"`
	SyncFastReadsTotal          float64 `perflib:"Sync Fast Reads/sec"`
	SyncMDLReadsTotal           float64 `perflib:"Sync MDL Reads/sec"`
	SyncPinReadsTotal           float64 `perflib:"Sync Pin Reads/sec"`
}

func (c *CacheCollector) collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []perflibCache // Single-instance class, array is required but will have single entry.
	if err := unmarshalObject(ctx.perfObjects["Cache"], &dst); err != nil {
		return nil, err
	}

	ch <- prometheus.MustNewConstMetric(
		c.AsyncCopyReadsTotal,
		prometheus.CounterValue,
		dst[0].AsyncCopyReadsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AsyncDataMapsTotal,
		prometheus.CounterValue,
		dst[0].AsyncDataMapsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AsyncFastReadsTotal,
		prometheus.CounterValue,
		dst[0].AsyncFastReadsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AsyncMDLReadsTotal,
		prometheus.CounterValue,
		dst[0].AsyncMDLReadsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AsyncPinReadsTotal,
		prometheus.CounterValue,
		dst[0].AsyncPinReadsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.CopyReadHitsTotal,
		prometheus.GaugeValue,
		dst[0].CopyReadHitsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.CopyReadsTotal,
		prometheus.CounterValue,
		dst[0].CopyReadsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.DataFlushesTotal,
		prometheus.CounterValue,
		dst[0].DataFlushesTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.DataFlushPagesTotal,
		prometheus.CounterValue,
		dst[0].DataFlushPagesTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.DataMapHitsPercent,
		prometheus.GaugeValue,
		dst[0].DataMapHitsPercent,
	)

	ch <- prometheus.MustNewConstMetric(
		c.DataMapPinsTotal,
		prometheus.CounterValue,
		dst[0].DataMapPinsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.DataMapsTotal,
		prometheus.CounterValue,
		dst[0].DataMapsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.DirtyPages,
		prometheus.GaugeValue,
		dst[0].DirtyPages,
	)

	ch <- prometheus.MustNewConstMetric(
		c.DirtyPageThreshold,
		prometheus.GaugeValue,
		dst[0].DirtyPageThreshold,
	)

	ch <- prometheus.MustNewConstMetric(
		c.FastReadNotPossiblesTotal,
		prometheus.CounterValue,
		dst[0].FastReadNotPossiblesTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.FastReadResourceMissesTotal,
		prometheus.CounterValue,
		dst[0].FastReadResourceMissesTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.FastReadsTotal,
		prometheus.CounterValue,
		dst[0].FastReadsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.LazyWriteFlushesTotal,
		prometheus.CounterValue,
		dst[0].LazyWriteFlushesTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.LazyWritePagesTotal,
		prometheus.CounterValue,
		dst[0].LazyWritePagesTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.MDLReadHitsTotal,
		prometheus.CounterValue,
		dst[0].MDLReadHitsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.MDLReadsTotal,
		prometheus.CounterValue,
		dst[0].MDLReadsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.PinReadHitsTotal,
		prometheus.CounterValue,
		dst[0].PinReadHitsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.PinReadsTotal,
		prometheus.CounterValue,
		dst[0].PinReadsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ReadAheadsTotal,
		prometheus.CounterValue,
		dst[0].ReadAheadsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.SyncCopyReadsTotal,
		prometheus.CounterValue,
		dst[0].SyncCopyReadsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.SyncDataMapsTotal,
		prometheus.CounterValue,
		dst[0].SyncDataMapsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.SyncFastReadsTotal,
		prometheus.CounterValue,
		dst[0].SyncFastReadsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.SyncMDLReadsTotal,
		prometheus.CounterValue,
		dst[0].SyncMDLReadsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.SyncPinReadsTotal,
		prometheus.CounterValue,
		dst[0].SyncPinReadsTotal,
	)

	return nil, nil
}

//go:build windows

package cache

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "cache"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for Perflib Cache metrics.
type Collector struct {
	config Config

	perfDataCollector *perfdata.Collector

	asyncCopyReadsTotal         *prometheus.Desc
	asyncDataMapsTotal          *prometheus.Desc
	asyncFastReadsTotal         *prometheus.Desc
	asyncMDLReadsTotal          *prometheus.Desc
	asyncPinReadsTotal          *prometheus.Desc
	copyReadHitsTotal           *prometheus.Desc
	copyReadsTotal              *prometheus.Desc
	dataFlushesTotal            *prometheus.Desc
	dataFlushPagesTotal         *prometheus.Desc
	dataMapHitsPercent          *prometheus.Desc
	dataMapPinsTotal            *prometheus.Desc
	dataMapsTotal               *prometheus.Desc
	dirtyPages                  *prometheus.Desc
	dirtyPageThreshold          *prometheus.Desc
	fastReadNotPossiblesTotal   *prometheus.Desc
	fastReadResourceMissesTotal *prometheus.Desc
	fastReadsTotal              *prometheus.Desc
	lazyWriteFlushesTotal       *prometheus.Desc
	lazyWritePagesTotal         *prometheus.Desc
	mdlReadHitsTotal            *prometheus.Desc
	mdlReadsTotal               *prometheus.Desc
	pinReadHitsTotal            *prometheus.Desc
	pinReadsTotal               *prometheus.Desc
	readAheadsTotal             *prometheus.Desc
	syncCopyReadsTotal          *prometheus.Desc
	syncDataMapsTotal           *prometheus.Desc
	syncFastReadsTotal          *prometheus.Desc
	syncMDLReadsTotal           *prometheus.Desc
	syncPinReadsTotal           *prometheus.Desc
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

func (c *Collector) Close() error {
	c.perfDataCollector.Close()

	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *mi.Session) error {
	var err error

	c.perfDataCollector, err = perfdata.NewCollector("Cache", perfdata.InstanceAll, []string{
		asyncCopyReadsTotal,
		asyncDataMapsTotal,
		asyncFastReadsTotal,
		asyncMDLReadsTotal,
		asyncPinReadsTotal,
		copyReadHitsTotal,
		copyReadsTotal,
		dataFlushesTotal,
		dataFlushPagesTotal,
		dataMapHitsPercent,
		dataMapPinsTotal,
		dataMapsTotal,
		dirtyPages,
		dirtyPageThreshold,
		fastReadNotPossiblesTotal,
		fastReadResourceMissesTotal,
		fastReadsTotal,
		lazyWriteFlushesTotal,
		lazyWritePagesTotal,
		mdlReadHitsTotal,
		mdlReadsTotal,
		pinReadHitsTotal,
		pinReadsTotal,
		readAheadsTotal,
		syncCopyReadsTotal,
		syncDataMapsTotal,
		syncFastReadsTotal,
		syncMDLReadsTotal,
		syncPinReadsTotal,
	})
	if err != nil {
		return fmt.Errorf("failed to create Cache collector: %w", err)
	}

	c.asyncCopyReadsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "async_copy_reads_total"),
		"(AsyncCopyReadsTotal)",
		nil,
		nil,
	)
	c.asyncDataMapsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "async_data_maps_total"),
		"(AsyncDataMapsTotal)",
		nil,
		nil,
	)
	c.asyncFastReadsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "async_fast_reads_total"),
		"(AsyncFastReadsTotal)",
		nil,
		nil,
	)
	c.asyncMDLReadsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "async_mdl_reads_total"),
		"(AsyncMDLReadsTotal)",
		nil,
		nil,
	)
	c.asyncPinReadsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "async_pin_reads_total"),
		"(AsyncPinReadsTotal)",
		nil,
		nil,
	)
	c.copyReadHitsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "copy_read_hits_total"),
		"(CopyReadHitsTotal)",
		nil,
		nil,
	)
	c.copyReadsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "copy_reads_total"),
		"(CopyReadsTotal)",
		nil,
		nil,
	)
	c.dataFlushesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "data_flushes_total"),
		"(DataFlushesTotal)",
		nil,
		nil,
	)
	c.dataFlushPagesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "data_flush_pages_total"),
		"(DataFlushPagesTotal)",
		nil,
		nil,
	)
	c.dataMapHitsPercent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "data_map_hits_percent"),
		"(DataMapHitsPercent)",
		nil,
		nil,
	)
	c.dataMapPinsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "data_map_pins_total"),
		"(DataMapPinsTotal)",
		nil,
		nil,
	)
	c.dataMapsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "data_maps_total"),
		"(DataMapsTotal)",
		nil,
		nil,
	)
	c.dirtyPages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dirty_pages"),
		"(DirtyPages)",
		nil,
		nil,
	)
	c.dirtyPageThreshold = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dirty_page_threshold"),
		"(DirtyPageThreshold)",
		nil,
		nil,
	)
	c.fastReadNotPossiblesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "fast_read_not_possibles_total"),
		"(FastReadNotPossiblesTotal)",
		nil,
		nil,
	)
	c.fastReadResourceMissesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "fast_read_resource_misses_total"),
		"(FastReadResourceMissesTotal)",
		nil,
		nil,
	)
	c.fastReadsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "fast_reads_total"),
		"(FastReadsTotal)",
		nil,
		nil,
	)
	c.lazyWriteFlushesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "lazy_write_flushes_total"),
		"(LazyWriteFlushesTotal)",
		nil,
		nil,
	)
	c.lazyWritePagesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "lazy_write_pages_total"),
		"(LazyWritePagesTotal)",
		nil,
		nil,
	)
	c.mdlReadHitsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mdl_read_hits_total"),
		"(MDLReadHitsTotal)",
		nil,
		nil,
	)
	c.mdlReadsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mdl_reads_total"),
		"(MDLReadsTotal)",
		nil,
		nil,
	)
	c.pinReadHitsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "pin_read_hits_total"),
		"(PinReadHitsTotal)",
		nil,
		nil,
	)
	c.pinReadsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "pin_reads_total"),
		"(PinReadsTotal)",
		nil,
		nil,
	)
	c.readAheadsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "read_aheads_total"),
		"(ReadAheadsTotal)",
		nil,
		nil,
	)
	c.syncCopyReadsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sync_copy_reads_total"),
		"(SyncCopyReadsTotal)",
		nil,
		nil,
	)
	c.syncDataMapsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sync_data_maps_total"),
		"(SyncDataMapsTotal)",
		nil,
		nil,
	)
	c.syncFastReadsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sync_fast_reads_total"),
		"(SyncFastReadsTotal)",
		nil,
		nil,
	)
	c.syncMDLReadsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sync_mdl_reads_total"),
		"(SyncMDLReadsTotal)",
		nil,
		nil,
	)
	c.syncPinReadsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sync_pin_reads_total"),
		"(SyncPinReadsTotal)",
		nil,
		nil,
	)

	return nil
}

// Collect implements the Collector interface.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	data, err := c.perfDataCollector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect Cache metrics: %w", err)
	}

	cacheData, ok := data[perfdata.EmptyInstance]

	if !ok {
		return errors.New("perflib query for Cache returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.asyncCopyReadsTotal,
		prometheus.CounterValue,
		cacheData[asyncCopyReadsTotal].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.asyncDataMapsTotal,
		prometheus.CounterValue,
		cacheData[asyncDataMapsTotal].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.asyncFastReadsTotal,
		prometheus.CounterValue,
		cacheData[asyncFastReadsTotal].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.asyncMDLReadsTotal,
		prometheus.CounterValue,
		cacheData[asyncMDLReadsTotal].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.asyncPinReadsTotal,
		prometheus.CounterValue,
		cacheData[asyncPinReadsTotal].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.copyReadHitsTotal,
		prometheus.GaugeValue,
		cacheData[copyReadHitsTotal].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.copyReadsTotal,
		prometheus.CounterValue,
		cacheData[copyReadsTotal].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.dataFlushesTotal,
		prometheus.CounterValue,
		cacheData[dataFlushesTotal].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.dataFlushPagesTotal,
		prometheus.CounterValue,
		cacheData[dataFlushPagesTotal].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.dataMapHitsPercent,
		prometheus.GaugeValue,
		cacheData[dataMapHitsPercent].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.dataMapPinsTotal,
		prometheus.CounterValue,
		cacheData[dataMapPinsTotal].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.dataMapsTotal,
		prometheus.CounterValue,
		cacheData[dataMapsTotal].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.dirtyPages,
		prometheus.GaugeValue,
		cacheData[dirtyPages].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.dirtyPageThreshold,
		prometheus.GaugeValue,
		cacheData[dirtyPageThreshold].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.fastReadNotPossiblesTotal,
		prometheus.CounterValue,
		cacheData[fastReadNotPossiblesTotal].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.fastReadResourceMissesTotal,
		prometheus.CounterValue,
		cacheData[fastReadResourceMissesTotal].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.fastReadsTotal,
		prometheus.CounterValue,
		cacheData[fastReadsTotal].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.lazyWriteFlushesTotal,
		prometheus.CounterValue,
		cacheData[lazyWriteFlushesTotal].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.lazyWritePagesTotal,
		prometheus.CounterValue,
		cacheData[lazyWritePagesTotal].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.mdlReadHitsTotal,
		prometheus.CounterValue,
		cacheData[mdlReadHitsTotal].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.mdlReadsTotal,
		prometheus.CounterValue,
		cacheData[mdlReadsTotal].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.pinReadHitsTotal,
		prometheus.CounterValue,
		cacheData[pinReadHitsTotal].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.pinReadsTotal,
		prometheus.CounterValue,
		cacheData[pinReadsTotal].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.readAheadsTotal,
		prometheus.CounterValue,
		cacheData[readAheadsTotal].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.syncCopyReadsTotal,
		prometheus.CounterValue,
		cacheData[syncCopyReadsTotal].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.syncDataMapsTotal,
		prometheus.CounterValue,
		cacheData[syncDataMapsTotal].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.syncFastReadsTotal,
		prometheus.CounterValue,
		cacheData[syncFastReadsTotal].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.syncMDLReadsTotal,
		prometheus.CounterValue,
		cacheData[syncMDLReadsTotal].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.syncPinReadsTotal,
		prometheus.CounterValue,
		cacheData[syncPinReadsTotal].FirstValue,
	)

	return nil
}

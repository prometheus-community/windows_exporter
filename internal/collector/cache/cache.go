// Copyright 2024 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows

package cache

import (
	"fmt"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "cache"

type Config struct{}

//nolint:gochecknoglobals
var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for Perflib Cache metrics.
type Collector struct {
	config Config

	perfDataCollector *pdh.Collector
	perfDataObject    []perfDataCounterValues

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

	var err error

	c.perfDataCollector, err = pdh.NewCollector[perfDataCounterValues](pdh.CounterTypeRaw, "Cache", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create Cache collector: %w", err)
	}

	return nil
}

// Collect implements the Collector interface.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollector.Collect(&c.perfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect Cache metrics: %w", err)
	}

	ch <- prometheus.MustNewConstMetric(
		c.asyncCopyReadsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].AsyncCopyReadsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.asyncDataMapsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].AsyncDataMapsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.asyncFastReadsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].AsyncFastReadsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.asyncMDLReadsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].AsyncMDLReadsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.asyncPinReadsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].AsyncPinReadsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.copyReadHitsTotal,
		prometheus.GaugeValue,
		c.perfDataObject[0].CopyReadHitsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.copyReadsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].CopyReadsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.dataFlushesTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DataFlushesTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.dataFlushPagesTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DataFlushPagesTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.dataMapHitsPercent,
		prometheus.GaugeValue,
		c.perfDataObject[0].DataMapHitsPercent,
	)

	ch <- prometheus.MustNewConstMetric(
		c.dataMapPinsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DataMapPinsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.dataMapsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DataMapsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.dirtyPages,
		prometheus.GaugeValue,
		c.perfDataObject[0].DirtyPages,
	)

	ch <- prometheus.MustNewConstMetric(
		c.dirtyPageThreshold,
		prometheus.GaugeValue,
		c.perfDataObject[0].DirtyPageThreshold,
	)

	ch <- prometheus.MustNewConstMetric(
		c.fastReadNotPossiblesTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].FastReadNotPossiblesTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.fastReadResourceMissesTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].FastReadResourceMissesTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.fastReadsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].FastReadsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.lazyWriteFlushesTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].LazyWriteFlushesTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.lazyWritePagesTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].LazyWritePagesTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.mdlReadHitsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].MdlReadHitsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.mdlReadsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].MdlReadsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.pinReadHitsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].PinReadHitsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.pinReadsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].PinReadsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.readAheadsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].ReadAheadsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.syncCopyReadsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].SyncCopyReadsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.syncDataMapsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].SyncDataMapsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.syncFastReadsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].SyncFastReadsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.syncMDLReadsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].SyncMDLReadsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.syncPinReadsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].SyncPinReadsTotal,
	)

	return nil
}

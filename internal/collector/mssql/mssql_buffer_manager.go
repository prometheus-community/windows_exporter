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

package mssql

import (
	"errors"
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorBufferManager struct {
	bufManPerfDataCollectors map[mssqlInstance]*pdh.Collector
	bufManPerfDataObject     []perfDataCounterValuesBufMan

	bufManBackgroundwriterpages         *prometheus.Desc
	bufManBuffercachehits               *prometheus.Desc
	bufManBuffercachelookups            *prometheus.Desc
	bufManCheckpointpages               *prometheus.Desc
	bufManDatabasepages                 *prometheus.Desc
	bufManExtensionallocatedpages       *prometheus.Desc
	bufManExtensionfreepages            *prometheus.Desc
	bufManExtensioninuseaspercentage    *prometheus.Desc
	bufManExtensionoutstandingIOcounter *prometheus.Desc
	bufManExtensionpageevictions        *prometheus.Desc
	bufManExtensionpagereads            *prometheus.Desc
	bufManExtensionpageunreferencedtime *prometheus.Desc
	bufManExtensionpagewrites           *prometheus.Desc
	bufManFreeliststalls                *prometheus.Desc
	bufManIntegralControllerSlope       *prometheus.Desc
	bufManLazywrites                    *prometheus.Desc
	bufManPagelifeexpectancy            *prometheus.Desc
	bufManPagelookups                   *prometheus.Desc
	bufManPagereads                     *prometheus.Desc
	bufManPagewrites                    *prometheus.Desc
	bufManReadaheadpages                *prometheus.Desc
	bufManReadaheadtime                 *prometheus.Desc
	bufManTargetpages                   *prometheus.Desc
}

type perfDataCounterValuesBufMan struct {
	BufManBackgroundWriterPagesPerSec   float64 `perfdata:"Background writer pages/sec"`
	BufManBufferCacheHitRatio           float64 `perfdata:"Buffer cache hit ratio"`
	BufManBufferCacheHitRatioBase       float64 `perfdata:"Buffer cache hit ratio base,secondvalue"`
	BufManCheckpointPagesPerSec         float64 `perfdata:"Checkpoint pages/sec"`
	BufManDatabasePages                 float64 `perfdata:"Database pages"`
	BufManExtensionAllocatedPages       float64 `perfdata:"Extension allocated pages"`
	BufManExtensionFreePages            float64 `perfdata:"Extension free pages"`
	BufManExtensionInUseAsPercentage    float64 `perfdata:"Extension in use as percentage"`
	BufManExtensionOutstandingIOCounter float64 `perfdata:"Extension outstanding IO counter"`
	BufManExtensionPageEvictionsPerSec  float64 `perfdata:"Extension page evictions/sec"`
	BufManExtensionPageReadsPerSec      float64 `perfdata:"Extension page reads/sec"`
	BufManExtensionPageUnreferencedTime float64 `perfdata:"Extension page unreferenced time"`
	BufManExtensionPageWritesPerSec     float64 `perfdata:"Extension page writes/sec"`
	BufManFreeListStallsPerSec          float64 `perfdata:"Free list stalls/sec"`
	BufManIntegralControllerSlope       float64 `perfdata:"Integral Controller Slope"`
	BufManLazyWritesPerSec              float64 `perfdata:"Lazy writes/sec"`
	BufManPageLifeExpectancy            float64 `perfdata:"Page life expectancy"`
	BufManPageLookupsPerSec             float64 `perfdata:"Page lookups/sec"`
	BufManPageReadsPerSec               float64 `perfdata:"Page reads/sec"`
	BufManPageWritesPerSec              float64 `perfdata:"Page writes/sec"`
	BufManReadaheadPagesPerSec          float64 `perfdata:"Readahead pages/sec"`
	BufManReadaheadTimePerSec           float64 `perfdata:"Readahead time/sec"`
	BufManTargetPages                   float64 `perfdata:"Target pages"`
}

func (c *Collector) buildBufferManager() error {
	var err error

	c.bufManPerfDataCollectors = make(map[mssqlInstance]*pdh.Collector, len(c.mssqlInstances))
	errs := make([]error, 0, len(c.mssqlInstances))

	for _, sqlInstance := range c.mssqlInstances {
		c.bufManPerfDataCollectors[sqlInstance], err = pdh.NewCollector[perfDataCounterValuesBufMan](pdh.CounterTypeRaw, c.mssqlGetPerfObjectName(sqlInstance, "Buffer Manager"), nil)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to create Buffer Manager collector for instance %s: %w", sqlInstance.name, err))
		}
	}

	c.bufManBackgroundwriterpages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_background_writer_pages"),
		"(BufferManager.Backgroundwriterpages)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManBuffercachehits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_buffer_cache_hits"),
		"(BufferManager.Buffercachehitratio)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManBuffercachelookups = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_buffer_cache_lookups"),
		"(BufferManager.Buffercachehitratio_Base)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManCheckpointpages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_checkpoint_pages"),
		"(BufferManager.Checkpointpages)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManDatabasepages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_database_pages"),
		"(BufferManager.Databasepages)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManExtensionallocatedpages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_extension_allocated_pages"),
		"(BufferManager.Extensionallocatedpages)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManExtensionfreepages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_extension_free_pages"),
		"(BufferManager.Extensionfreepages)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManExtensioninuseaspercentage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_extension_in_use_as_percentage"),
		"(BufferManager.Extensioninuseaspercentage)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManExtensionoutstandingIOcounter = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_extension_outstanding_io"),
		"(BufferManager.ExtensionoutstandingIOcounter)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManExtensionpageevictions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_extension_page_evictions"),
		"(BufferManager.Extensionpageevictions)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManExtensionpagereads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_extension_page_reads"),
		"(BufferManager.Extensionpagereads)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManExtensionpageunreferencedtime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_extension_page_unreferenced_seconds"),
		"(BufferManager.Extensionpageunreferencedtime)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManExtensionpagewrites = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_extension_page_writes"),
		"(BufferManager.Extensionpagewrites)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManFreeliststalls = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_free_list_stalls"),
		"(BufferManager.Freeliststalls)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManIntegralControllerSlope = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_integral_controller_slope"),
		"(BufferManager.IntegralControllerSlope)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManLazywrites = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_lazywrites"),
		"(BufferManager.Lazywrites)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManPagelifeexpectancy = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_page_life_expectancy_seconds"),
		"(BufferManager.Pagelifeexpectancy)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManPagelookups = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_page_lookups"),
		"(BufferManager.Pagelookups)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManPagereads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_page_reads"),
		"(BufferManager.Pagereads)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManPagewrites = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_page_writes"),
		"(BufferManager.Pagewrites)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManReadaheadpages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_read_ahead_pages"),
		"(BufferManager.Readaheadpages)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManReadaheadtime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_read_ahead_issuing_seconds"),
		"(BufferManager.Readaheadtime)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManTargetpages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_target_pages"),
		"(BufferManager.Targetpages)",
		[]string{"mssql_instance"},
		nil,
	)

	return errors.Join(errs...)
}

func (c *Collector) collectBufferManager(ch chan<- prometheus.Metric) error {
	return c.collect(ch, subCollectorBufferManager, c.bufManPerfDataCollectors, c.collectBufferManagerInstance)
}

func (c *Collector) collectBufferManagerInstance(ch chan<- prometheus.Metric, sqlInstance mssqlInstance, perfDataCollector *pdh.Collector) error {
	err := perfDataCollector.Collect(&c.bufManPerfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect %s metrics: %w", c.mssqlGetPerfObjectName(sqlInstance, "Buffer Manager"), err)
	}

	for _, data := range c.bufManPerfDataObject {
		ch <- prometheus.MustNewConstMetric(
			c.bufManBackgroundwriterpages,
			prometheus.CounterValue,
			data.BufManBackgroundWriterPagesPerSec,
			sqlInstance.name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManBuffercachehits,
			prometheus.GaugeValue,
			data.BufManBufferCacheHitRatio,
			sqlInstance.name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManBuffercachelookups,
			prometheus.GaugeValue,
			data.BufManBufferCacheHitRatioBase,
			sqlInstance.name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManCheckpointpages,
			prometheus.CounterValue,
			data.BufManCheckpointPagesPerSec,
			sqlInstance.name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManDatabasepages,
			prometheus.GaugeValue,
			data.BufManDatabasePages,
			sqlInstance.name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManExtensionallocatedpages,
			prometheus.GaugeValue,
			data.BufManExtensionAllocatedPages,
			sqlInstance.name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManExtensionfreepages,
			prometheus.GaugeValue,
			data.BufManExtensionFreePages,
			sqlInstance.name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManExtensioninuseaspercentage,
			prometheus.GaugeValue,
			data.BufManExtensionInUseAsPercentage,
			sqlInstance.name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManExtensionoutstandingIOcounter,
			prometheus.GaugeValue,
			data.BufManExtensionOutstandingIOCounter,
			sqlInstance.name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManExtensionpageevictions,
			prometheus.CounterValue,
			data.BufManExtensionPageEvictionsPerSec,
			sqlInstance.name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManExtensionpagereads,
			prometheus.CounterValue,
			data.BufManExtensionPageReadsPerSec,
			sqlInstance.name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManExtensionpageunreferencedtime,
			prometheus.GaugeValue,
			data.BufManExtensionPageUnreferencedTime,
			sqlInstance.name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManExtensionpagewrites,
			prometheus.CounterValue,
			data.BufManExtensionPageWritesPerSec,
			sqlInstance.name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManFreeliststalls,
			prometheus.CounterValue,
			data.BufManFreeListStallsPerSec,
			sqlInstance.name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManIntegralControllerSlope,
			prometheus.GaugeValue,
			data.BufManIntegralControllerSlope,
			sqlInstance.name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManLazywrites,
			prometheus.CounterValue,
			data.BufManLazyWritesPerSec,
			sqlInstance.name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManPagelifeexpectancy,
			prometheus.GaugeValue,
			data.BufManPageLifeExpectancy,
			sqlInstance.name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManPagelookups,
			prometheus.CounterValue,
			data.BufManPageLookupsPerSec,
			sqlInstance.name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManPagereads,
			prometheus.CounterValue,
			data.BufManPageReadsPerSec,
			sqlInstance.name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManPagewrites,
			prometheus.CounterValue,
			data.BufManPageWritesPerSec,
			sqlInstance.name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManReadaheadpages,
			prometheus.CounterValue,
			data.BufManReadaheadPagesPerSec,
			sqlInstance.name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManReadaheadtime,
			prometheus.CounterValue,
			data.BufManReadaheadTimePerSec,
			sqlInstance.name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManTargetpages,
			prometheus.GaugeValue,
			data.BufManTargetPages,
			sqlInstance.name,
		)
	}

	return nil
}

func (c *Collector) closeBufferManager() {
	for _, perfDataCollectors := range c.bufManPerfDataCollectors {
		perfDataCollectors.Close()
	}
}

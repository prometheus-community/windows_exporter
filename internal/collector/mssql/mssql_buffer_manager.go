//go:build windows

package mssql

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorBufferManager struct {
	bufManPerfDataCollectors map[string]*perfdata.Collector

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

const (
	bufManBackgroundWriterPagesPerSec   = "Background writer pages/sec"
	bufManBufferCacheHitRatio           = "Buffer cache hit ratio"
	bufManBufferCacheHitRatioBase       = "Buffer cache hit ratio base"
	bufManCheckpointPagesPerSec         = "Checkpoint pages/sec"
	bufManDatabasePages                 = "Database pages"
	bufManExtensionAllocatedPages       = "Extension allocated pages"
	bufManExtensionFreePages            = "Extension free pages"
	bufManExtensionInUseAsPercentage    = "Extension in use as percentage"
	bufManExtensionOutstandingIOCounter = "Extension outstanding IO counter"
	bufManExtensionPageEvictionsPerSec  = "Extension page evictions/sec"
	bufManExtensionPageReadsPerSec      = "Extension page reads/sec"
	bufManExtensionPageUnreferencedTime = "Extension page unreferenced time"
	bufManExtensionPageWritesPerSec     = "Extension page writes/sec"
	bufManFreeListStallsPerSec          = "Free list stalls/sec"
	bufManIntegralControllerSlope       = "Integral Controller Slope"
	bufManLazyWritesPerSec              = "Lazy writes/sec"
	bufManPageLifeExpectancy            = "Page life expectancy"
	bufManPageLookupsPerSec             = "Page lookups/sec"
	bufManPageReadsPerSec               = "Page reads/sec"
	bufManPageWritesPerSec              = "Page writes/sec"
	bufManReadaheadPagesPerSec          = "Readahead pages/sec"
	bufManReadaheadTimePerSec           = "Readahead time/sec"
	bufManTargetPages                   = "Target pages"
)

func (c *Collector) buildBufferManager() error {
	var err error

	c.bufManPerfDataCollectors = make(map[string]*perfdata.Collector, len(c.mssqlInstances))
	counters := []string{
		bufManBackgroundWriterPagesPerSec,
		bufManBufferCacheHitRatio,
		bufManBufferCacheHitRatioBase,
		bufManCheckpointPagesPerSec,
		bufManDatabasePages,
		bufManExtensionAllocatedPages,
		bufManExtensionFreePages,
		bufManExtensionInUseAsPercentage,
		bufManExtensionOutstandingIOCounter,
		bufManExtensionPageEvictionsPerSec,
		bufManExtensionPageReadsPerSec,
		bufManExtensionPageUnreferencedTime,
		bufManExtensionPageWritesPerSec,
		bufManFreeListStallsPerSec,
		bufManIntegralControllerSlope,
		bufManLazyWritesPerSec,
		bufManPageLifeExpectancy,
		bufManPageLookupsPerSec,
		bufManPageReadsPerSec,
		bufManPageWritesPerSec,
		bufManReadaheadPagesPerSec,
		bufManReadaheadTimePerSec,
		bufManTargetPages,
	}

	for sqlInstance := range c.mssqlInstances {
		c.bufManPerfDataCollectors[sqlInstance], err = perfdata.NewCollector(c.mssqlGetPerfObjectName(sqlInstance, "Buffer Manager"), nil, counters)
		if err != nil {
			return fmt.Errorf("failed to create Buffer Manager collector for instance %s: %w", sqlInstance, err)
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

	return nil
}

func (c *Collector) collectBufferManager(ch chan<- prometheus.Metric) error {
	return c.collect(ch, subCollectorBufferManager, c.bufManPerfDataCollectors, c.collectBufferManagerInstance)
}

func (c *Collector) collectBufferManagerInstance(ch chan<- prometheus.Metric, sqlInstance string, perfDataCollector *perfdata.Collector) error {
	perfData, err := perfDataCollector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect %s metrics: %w", c.mssqlGetPerfObjectName(sqlInstance, "Buffer Manager"), err)
	}

	for _, data := range perfData {
		ch <- prometheus.MustNewConstMetric(
			c.bufManBackgroundwriterpages,
			prometheus.CounterValue,
			data[bufManBackgroundWriterPagesPerSec].FirstValue,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManBuffercachehits,
			prometheus.GaugeValue,
			data[bufManBufferCacheHitRatio].FirstValue,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManBuffercachelookups,
			prometheus.GaugeValue,
			data[bufManBufferCacheHitRatioBase].SecondValue,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManCheckpointpages,
			prometheus.CounterValue,
			data[bufManCheckpointPagesPerSec].FirstValue,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManDatabasepages,
			prometheus.GaugeValue,
			data[bufManDatabasePages].FirstValue,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManExtensionallocatedpages,
			prometheus.GaugeValue,
			data[bufManExtensionAllocatedPages].FirstValue,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManExtensionfreepages,
			prometheus.GaugeValue,
			data[bufManExtensionFreePages].FirstValue,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManExtensioninuseaspercentage,
			prometheus.GaugeValue,
			data[bufManExtensionInUseAsPercentage].FirstValue,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManExtensionoutstandingIOcounter,
			prometheus.GaugeValue,
			data[bufManExtensionOutstandingIOCounter].FirstValue,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManExtensionpageevictions,
			prometheus.CounterValue,
			data[bufManExtensionPageEvictionsPerSec].FirstValue,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManExtensionpagereads,
			prometheus.CounterValue,
			data[bufManExtensionPageReadsPerSec].FirstValue,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManExtensionpageunreferencedtime,
			prometheus.GaugeValue,
			data[bufManExtensionPageUnreferencedTime].FirstValue,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManExtensionpagewrites,
			prometheus.CounterValue,
			data[bufManExtensionPageWritesPerSec].FirstValue,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManFreeliststalls,
			prometheus.CounterValue,
			data[bufManFreeListStallsPerSec].FirstValue,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManIntegralControllerSlope,
			prometheus.GaugeValue,
			data[bufManIntegralControllerSlope].FirstValue,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManLazywrites,
			prometheus.CounterValue,
			data[bufManLazyWritesPerSec].FirstValue,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManPagelifeexpectancy,
			prometheus.GaugeValue,
			data[bufManPageLifeExpectancy].FirstValue,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManPagelookups,
			prometheus.CounterValue,
			data[bufManPageLookupsPerSec].FirstValue,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManPagereads,
			prometheus.CounterValue,
			data[bufManPageReadsPerSec].FirstValue,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManPagewrites,
			prometheus.CounterValue,
			data[bufManPageWritesPerSec].FirstValue,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManReadaheadpages,
			prometheus.CounterValue,
			data[bufManReadaheadPagesPerSec].FirstValue,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManReadaheadtime,
			prometheus.CounterValue,
			data[bufManReadaheadTimePerSec].FirstValue,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManTargetpages,
			prometheus.GaugeValue,
			data[bufManTargetPages].FirstValue,
			sqlInstance,
		)
	}

	return nil
}

func (c *Collector) closeBufferManager() {
	for _, perfDataCollectors := range c.bufManPerfDataCollectors {
		perfDataCollectors.Close()
	}
}

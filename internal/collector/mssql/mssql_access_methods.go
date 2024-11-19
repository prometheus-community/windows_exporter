//go:build windows

package mssql

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorAccessMethods struct {
	accessMethodsPerfDataCollectors map[string]*perfdata.Collector

	accessMethodsAUcleanupbatches             *prometheus.Desc
	accessMethodsAUcleanups                   *prometheus.Desc
	accessMethodsByReferenceLobCreateCount    *prometheus.Desc
	accessMethodsByReferenceLobUseCount       *prometheus.Desc
	accessMethodsCountLobReadahead            *prometheus.Desc
	accessMethodsCountPullInRow               *prometheus.Desc
	accessMethodsCountPushOffRow              *prometheus.Desc
	accessMethodsDeferreddroppedAUs           *prometheus.Desc
	accessMethodsDeferredDroppedrowsets       *prometheus.Desc
	accessMethodsDroppedrowsetcleanups        *prometheus.Desc
	accessMethodsDroppedrowsetsskipped        *prometheus.Desc
	accessMethodsExtentDeallocations          *prometheus.Desc
	accessMethodsExtentsAllocated             *prometheus.Desc
	accessMethodsFailedAUcleanupbatches       *prometheus.Desc
	accessMethodsFailedleafpagecookie         *prometheus.Desc
	accessMethodsFailedtreepagecookie         *prometheus.Desc
	accessMethodsForwardedRecords             *prometheus.Desc
	accessMethodsFreeSpacePageFetches         *prometheus.Desc
	accessMethodsFreeSpaceScans               *prometheus.Desc
	accessMethodsFullScans                    *prometheus.Desc
	accessMethodsIndexSearches                *prometheus.Desc
	accessMethodsInSysXactwaits               *prometheus.Desc
	accessMethodsLobHandleCreateCount         *prometheus.Desc
	accessMethodsLobHandleDestroyCount        *prometheus.Desc
	accessMethodsLobSSProviderCreateCount     *prometheus.Desc
	accessMethodsLobSSProviderDestroyCount    *prometheus.Desc
	accessMethodsLobSSProviderTruncationCount *prometheus.Desc
	accessMethodsMixedPageAllocations         *prometheus.Desc
	accessMethodsPageCompressionAttempts      *prometheus.Desc
	accessMethodsPageDeallocations            *prometheus.Desc
	accessMethodsPagesAllocated               *prometheus.Desc
	accessMethodsPagesCompressed              *prometheus.Desc
	accessMethodsPageSplits                   *prometheus.Desc
	accessMethodsProbeScans                   *prometheus.Desc
	accessMethodsRangeScans                   *prometheus.Desc
	accessMethodsScanPointRevalidations       *prometheus.Desc
	accessMethodsSkippedGhostedRecords        *prometheus.Desc
	accessMethodsTableLockEscalations         *prometheus.Desc
	accessMethodsUsedleafpagecookie           *prometheus.Desc
	accessMethodsUsedtreepagecookie           *prometheus.Desc
	accessMethodsWorkfilesCreated             *prometheus.Desc
	accessMethodsWorktablesCreated            *prometheus.Desc
	accessMethodsWorktablesFromCacheHits      *prometheus.Desc
	accessMethodsWorktablesFromCacheLookups   *prometheus.Desc
}

const (
	accessMethodsAUCleanupbatchesPerSec        = "AU cleanup batches/sec"
	accessMethodsAUCleanupsPerSec              = "AU cleanups/sec"
	accessMethodsByReferenceLobCreateCount     = "By-reference Lob Create Count"
	accessMethodsByReferenceLobUseCount        = "By-reference Lob Use Count"
	accessMethodsCountLobReadahead             = "Count Lob Readahead"
	accessMethodsCountPullInRow                = "Count Pull In Row"
	accessMethodsCountPushOffRow               = "Count Push Off Row"
	accessMethodsDeferredDroppedAUs            = "Deferred dropped AUs"
	accessMethodsDeferredDroppedRowsets        = "Deferred Dropped rowsets"
	accessMethodsDroppedRowsetCleanupsPerSec   = "Dropped rowset cleanups/sec"
	accessMethodsDroppedRowsetsSkippedPerSec   = "Dropped rowsets skipped/sec"
	accessMethodsExtentDeallocationsPerSec     = "Extent Deallocations/sec"
	accessMethodsExtentsAllocatedPerSec        = "Extents Allocated/sec"
	accessMethodsFailedAUCleanupBatchesPerSec  = "Failed AU cleanup batches/sec"
	accessMethodsFailedLeafPageCookie          = "Failed leaf page cookie"
	accessMethodsFailedTreePageCookie          = "Failed tree page cookie"
	accessMethodsForwardedRecordsPerSec        = "Forwarded Records/sec"
	accessMethodsFreeSpacePageFetchesPerSec    = "FreeSpace Page Fetches/sec"
	accessMethodsFreeSpaceScansPerSec          = "FreeSpace Scans/sec"
	accessMethodsFullScansPerSec               = "Full Scans/sec"
	accessMethodsIndexSearchesPerSec           = "Index Searches/sec"
	accessMethodsInSysXactWaitsPerSec          = "InSysXact waits/sec"
	accessMethodsLobHandleCreateCount          = "LobHandle Create Count"
	accessMethodsLobHandleDestroyCount         = "LobHandle Destroy Count"
	accessMethodsLobSSProviderCreateCount      = "LobSS Provider Create Count"
	accessMethodsLobSSProviderDestroyCount     = "LobSS Provider Destroy Count"
	accessMethodsLobSSProviderTruncationCount  = "LobSS Provider Truncation Count"
	accessMethodsMixedPageAllocationsPerSec    = "Mixed page allocations/sec"
	accessMethodsPageCompressionAttemptsPerSec = "Page compression attempts/sec"
	accessMethodsPageDeallocationsPerSec       = "Page Deallocations/sec"
	accessMethodsPagesAllocatedPerSec          = "Pages Allocated/sec"
	accessMethodsPagesCompressedPerSec         = "Pages compressed/sec"
	accessMethodsPageSplitsPerSec              = "Page Splits/sec"
	accessMethodsProbeScansPerSec              = "Probe Scans/sec"
	accessMethodsRangeScansPerSec              = "Range Scans/sec"
	accessMethodsScanPointRevalidationsPerSec  = "Scan Point Revalidations/sec"
	accessMethodsSkippedGhostedRecordsPerSec   = "Skipped Ghosted Records/sec"
	accessMethodsTableLockEscalationsPerSec    = "Table Lock Escalations/sec"
	accessMethodsUsedLeafPageCookie            = "Used leaf page cookie"
	accessMethodsUsedTreePageCookie            = "Used tree page cookie"
	accessMethodsWorkfilesCreatedPerSec        = "Workfiles Created/sec"
	accessMethodsWorktablesCreatedPerSec       = "Worktables Created/sec"
	accessMethodsWorktablesFromCacheRatio      = "Worktables From Cache Ratio"
	accessMethodsWorktablesFromCacheRatioBase  = "Worktables From Cache Base"
)

func (c *Collector) buildAccessMethods() error {
	var err error

	c.accessMethodsPerfDataCollectors = make(map[string]*perfdata.Collector, len(c.mssqlInstances))
	counters := []string{
		accessMethodsAUCleanupbatchesPerSec,
		accessMethodsAUCleanupsPerSec,
		accessMethodsByReferenceLobCreateCount,
		accessMethodsByReferenceLobUseCount,
		accessMethodsCountLobReadahead,
		accessMethodsCountPullInRow,
		accessMethodsCountPushOffRow,
		accessMethodsDeferredDroppedAUs,
		accessMethodsDeferredDroppedRowsets,
		accessMethodsDroppedRowsetCleanupsPerSec,
		accessMethodsDroppedRowsetsSkippedPerSec,
		accessMethodsExtentDeallocationsPerSec,
		accessMethodsExtentsAllocatedPerSec,
		accessMethodsFailedAUCleanupBatchesPerSec,
		accessMethodsFailedLeafPageCookie,
		accessMethodsFailedTreePageCookie,
		accessMethodsForwardedRecordsPerSec,
		accessMethodsFreeSpacePageFetchesPerSec,
		accessMethodsFreeSpaceScansPerSec,
		accessMethodsFullScansPerSec,
		accessMethodsIndexSearchesPerSec,
		accessMethodsInSysXactWaitsPerSec,
		accessMethodsLobHandleCreateCount,
		accessMethodsLobHandleDestroyCount,
		accessMethodsLobSSProviderCreateCount,
		accessMethodsLobSSProviderDestroyCount,
		accessMethodsLobSSProviderTruncationCount,
		accessMethodsMixedPageAllocationsPerSec,
		accessMethodsPageCompressionAttemptsPerSec,
		accessMethodsPageDeallocationsPerSec,
		accessMethodsPagesAllocatedPerSec,
		accessMethodsPagesCompressedPerSec,
		accessMethodsPageSplitsPerSec,
		accessMethodsProbeScansPerSec,
		accessMethodsRangeScansPerSec,
		accessMethodsScanPointRevalidationsPerSec,
		accessMethodsSkippedGhostedRecordsPerSec,
		accessMethodsTableLockEscalationsPerSec,
		accessMethodsUsedLeafPageCookie,
		accessMethodsUsedTreePageCookie,
		accessMethodsWorkfilesCreatedPerSec,
		accessMethodsWorktablesCreatedPerSec,
		accessMethodsWorktablesFromCacheRatio,
		accessMethodsWorktablesFromCacheRatioBase,
	}

	for sqlInstance := range c.mssqlInstances {
		c.accessMethodsPerfDataCollectors[sqlInstance], err = perfdata.NewCollector(c.mssqlGetPerfObjectName(sqlInstance, "Access Methods"), nil, counters)
		if err != nil {
			return fmt.Errorf("failed to create AccessMethods collector for instance %s: %w", sqlInstance, err)
		}
	}

	// Win32_PerfRawData_{instance}_SQLServerAccessMethods
	c.accessMethodsAUcleanupbatches = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_au_batch_cleanups"),
		"(AccessMethods.AUcleanupbatches)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsAUcleanups = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_au_cleanups"),
		"(AccessMethods.AUcleanups)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsByReferenceLobCreateCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_by_reference_lob_creates"),
		"(AccessMethods.ByreferenceLobCreateCount)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsByReferenceLobUseCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_by_reference_lob_uses"),
		"(AccessMethods.ByreferenceLobUseCount)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsCountLobReadahead = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_lob_read_aheads"),
		"(AccessMethods.CountLobReadahead)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsCountPullInRow = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_column_value_pulls"),
		"(AccessMethods.CountPullInRow)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsCountPushOffRow = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_column_value_pushes"),
		"(AccessMethods.CountPushOffRow)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsDeferreddroppedAUs = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_deferred_dropped_aus"),
		"(AccessMethods.DeferreddroppedAUs)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsDeferredDroppedrowsets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_deferred_dropped_rowsets"),
		"(AccessMethods.DeferredDroppedrowsets)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsDroppedrowsetcleanups = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_dropped_rowset_cleanups"),
		"(AccessMethods.Droppedrowsetcleanups)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsDroppedrowsetsskipped = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_dropped_rowset_skips"),
		"(AccessMethods.Droppedrowsetsskipped)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsExtentDeallocations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_extent_deallocations"),
		"(AccessMethods.ExtentDeallocations)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsExtentsAllocated = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_extent_allocations"),
		"(AccessMethods.ExtentsAllocated)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsFailedAUcleanupbatches = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_au_batch_cleanup_failures"),
		"(AccessMethods.FailedAUcleanupbatches)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsFailedleafpagecookie = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_leaf_page_cookie_failures"),
		"(AccessMethods.Failedleafpagecookie)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsFailedtreepagecookie = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_tree_page_cookie_failures"),
		"(AccessMethods.Failedtreepagecookie)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsForwardedRecords = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_forwarded_records"),
		"(AccessMethods.ForwardedRecords)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsFreeSpacePageFetches = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_free_space_page_fetches"),
		"(AccessMethods.FreeSpacePageFetches)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsFreeSpaceScans = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_free_space_scans"),
		"(AccessMethods.FreeSpaceScans)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsFullScans = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_full_scans"),
		"(AccessMethods.FullScans)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsIndexSearches = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_index_searches"),
		"(AccessMethods.IndexSearches)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsInSysXactwaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_insysxact_waits"),
		"(AccessMethods.InSysXactwaits)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsLobHandleCreateCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_lob_handle_creates"),
		"(AccessMethods.LobHandleCreateCount)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsLobHandleDestroyCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_lob_handle_destroys"),
		"(AccessMethods.LobHandleDestroyCount)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsLobSSProviderCreateCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_lob_ss_provider_creates"),
		"(AccessMethods.LobSSProviderCreateCount)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsLobSSProviderDestroyCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_lob_ss_provider_destroys"),
		"(AccessMethods.LobSSProviderDestroyCount)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsLobSSProviderTruncationCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_lob_ss_provider_truncations"),
		"(AccessMethods.LobSSProviderTruncationCount)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsMixedPageAllocations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_mixed_page_allocations"),
		"(AccessMethods.MixedpageallocationsPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsPageCompressionAttempts = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_page_compression_attempts"),
		"(AccessMethods.PagecompressionattemptsPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsPageDeallocations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_page_deallocations"),
		"(AccessMethods.PageDeallocationsPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsPagesAllocated = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_page_allocations"),
		"(AccessMethods.PagesAllocatedPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsPagesCompressed = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_page_compressions"),
		"(AccessMethods.PagescompressedPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsPageSplits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_page_splits"),
		"(AccessMethods.PageSplitsPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsProbeScans = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_probe_scans"),
		"(AccessMethods.ProbeScansPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsRangeScans = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_range_scans"),
		"(AccessMethods.RangeScansPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsScanPointRevalidations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_scan_point_revalidations"),
		"(AccessMethods.ScanPointRevalidationsPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsSkippedGhostedRecords = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_ghost_record_skips"),
		"(AccessMethods.SkippedGhostedRecordsPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsTableLockEscalations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_table_lock_escalations"),
		"(AccessMethods.TableLockEscalationsPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsUsedleafpagecookie = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_leaf_page_cookie_uses"),
		"(AccessMethods.Usedleafpagecookie)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsUsedtreepagecookie = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_tree_page_cookie_uses"),
		"(AccessMethods.Usedtreepagecookie)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsWorkfilesCreated = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_workfile_creates"),
		"(AccessMethods.WorkfilesCreatedPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsWorktablesCreated = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_worktables_creates"),
		"(AccessMethods.WorktablesCreatedPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsWorktablesFromCacheHits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_worktables_from_cache_hits"),
		"(AccessMethods.WorktablesFromCacheRatio)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsWorktablesFromCacheLookups = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_worktables_from_cache_lookups"),
		"(AccessMethods.WorktablesFromCacheRatio_Base)",
		[]string{"mssql_instance"},
		nil,
	)

	return nil
}

func (c *Collector) collectAccessMethods(ch chan<- prometheus.Metric) error {
	return c.collect(ch, subCollectorAccessMethods, c.accessMethodsPerfDataCollectors, c.collectAccessMethodsInstance)
}

func (c *Collector) collectAccessMethodsInstance(ch chan<- prometheus.Metric, sqlInstance string, perfDataCollector *perfdata.Collector) error {
	perfData, err := perfDataCollector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect %s metrics: %w", c.mssqlGetPerfObjectName(sqlInstance, "AccessMethods"), err)
	}

	data, ok := perfData[perfdata.EmptyInstance]
	if !ok {
		return fmt.Errorf("perflib query for %s returned empty result set", c.mssqlGetPerfObjectName(sqlInstance, "AccessMethods"))
	}

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsAUcleanupbatches,
		prometheus.CounterValue,
		data[accessMethodsAUCleanupbatchesPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsAUcleanups,
		prometheus.CounterValue,
		data[accessMethodsAUCleanupsPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsByReferenceLobCreateCount,
		prometheus.CounterValue,
		data[accessMethodsByReferenceLobCreateCount].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsByReferenceLobUseCount,
		prometheus.CounterValue,
		data[accessMethodsByReferenceLobUseCount].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsCountLobReadahead,
		prometheus.CounterValue,
		data[accessMethodsCountLobReadahead].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsCountPullInRow,
		prometheus.CounterValue,
		data[accessMethodsCountPullInRow].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsCountPushOffRow,
		prometheus.CounterValue,
		data[accessMethodsCountPushOffRow].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsDeferreddroppedAUs,
		prometheus.GaugeValue,
		data[accessMethodsDeferredDroppedAUs].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsDeferredDroppedrowsets,
		prometheus.GaugeValue,
		data[accessMethodsDeferredDroppedRowsets].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsDroppedrowsetcleanups,
		prometheus.CounterValue,
		data[accessMethodsDroppedRowsetCleanupsPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsDroppedrowsetsskipped,
		prometheus.CounterValue,
		data[accessMethodsDroppedRowsetsSkippedPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsExtentDeallocations,
		prometheus.CounterValue,
		data[accessMethodsExtentDeallocationsPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsExtentsAllocated,
		prometheus.CounterValue,
		data[accessMethodsExtentsAllocatedPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsFailedAUcleanupbatches,
		prometheus.CounterValue,
		data[accessMethodsFailedAUCleanupBatchesPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsFailedleafpagecookie,
		prometheus.CounterValue,
		data[accessMethodsFailedLeafPageCookie].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsFailedtreepagecookie,
		prometheus.CounterValue,
		data[accessMethodsFailedTreePageCookie].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsForwardedRecords,
		prometheus.CounterValue,
		data[accessMethodsForwardedRecordsPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsFreeSpacePageFetches,
		prometheus.CounterValue,
		data[accessMethodsFreeSpacePageFetchesPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsFreeSpaceScans,
		prometheus.CounterValue,
		data[accessMethodsFreeSpaceScansPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsFullScans,
		prometheus.CounterValue,
		data[accessMethodsFullScansPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsIndexSearches,
		prometheus.CounterValue,
		data[accessMethodsIndexSearchesPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsInSysXactwaits,
		prometheus.CounterValue,
		data[accessMethodsInSysXactWaitsPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsLobHandleCreateCount,
		prometheus.CounterValue,
		data[accessMethodsLobHandleCreateCount].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsLobHandleDestroyCount,
		prometheus.CounterValue,
		data[accessMethodsLobHandleDestroyCount].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsLobSSProviderCreateCount,
		prometheus.CounterValue,
		data[accessMethodsLobSSProviderCreateCount].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsLobSSProviderDestroyCount,
		prometheus.CounterValue,
		data[accessMethodsLobSSProviderDestroyCount].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsLobSSProviderTruncationCount,
		prometheus.CounterValue,
		data[accessMethodsLobSSProviderTruncationCount].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsMixedPageAllocations,
		prometheus.CounterValue,
		data[accessMethodsMixedPageAllocationsPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsPageCompressionAttempts,
		prometheus.CounterValue,
		data[accessMethodsPageCompressionAttemptsPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsPageDeallocations,
		prometheus.CounterValue,
		data[accessMethodsPageDeallocationsPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsPagesAllocated,
		prometheus.CounterValue,
		data[accessMethodsPagesAllocatedPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsPagesCompressed,
		prometheus.CounterValue,
		data[accessMethodsPagesCompressedPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsPageSplits,
		prometheus.CounterValue,
		data[accessMethodsPageSplitsPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsProbeScans,
		prometheus.CounterValue,
		data[accessMethodsProbeScansPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsRangeScans,
		prometheus.CounterValue,
		data[accessMethodsRangeScansPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsScanPointRevalidations,
		prometheus.CounterValue,
		data[accessMethodsScanPointRevalidationsPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsSkippedGhostedRecords,
		prometheus.CounterValue,
		data[accessMethodsSkippedGhostedRecordsPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsTableLockEscalations,
		prometheus.CounterValue,
		data[accessMethodsTableLockEscalationsPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsUsedleafpagecookie,
		prometheus.CounterValue,
		data[accessMethodsUsedLeafPageCookie].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsUsedtreepagecookie,
		prometheus.CounterValue,
		data[accessMethodsUsedTreePageCookie].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsWorkfilesCreated,
		prometheus.CounterValue,
		data[accessMethodsWorkfilesCreatedPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsWorktablesCreated,
		prometheus.CounterValue,
		data[accessMethodsWorktablesCreatedPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsWorktablesFromCacheHits,
		prometheus.CounterValue,
		data[accessMethodsWorktablesFromCacheRatio].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsWorktablesFromCacheLookups,
		prometheus.CounterValue,
		data[accessMethodsWorktablesFromCacheRatioBase].SecondValue,
		sqlInstance,
	)

	return nil
}

func (c *Collector) closeAccessMethods() {
	for _, perfDataCollector := range c.accessMethodsPerfDataCollectors {
		perfDataCollector.Close()
	}
}

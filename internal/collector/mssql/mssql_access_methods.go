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

type collectorAccessMethods struct {
	accessMethodsPerfDataCollectors map[mssqlInstance]*pdh.Collector
	accessMethodsPerfDataObject     []perfDataCounterValuesAccessMethods

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

type perfDataCounterValuesAccessMethods struct {
	AccessMethodsAUCleanupbatchesPerSec        float64 `perfdata:"AU cleanup batches/sec"`
	AccessMethodsAUCleanupsPerSec              float64 `perfdata:"AU cleanups/sec"`
	AccessMethodsByReferenceLobCreateCount     float64 `perfdata:"By-reference Lob Create Count"`
	AccessMethodsByReferenceLobUseCount        float64 `perfdata:"By-reference Lob Use Count"`
	AccessMethodsCountLobReadahead             float64 `perfdata:"Count Lob Readahead"`
	AccessMethodsCountPullInRow                float64 `perfdata:"Count Pull In Row"`
	AccessMethodsCountPushOffRow               float64 `perfdata:"Count Push Off Row"`
	AccessMethodsDeferredDroppedAUs            float64 `perfdata:"Deferred dropped AUs"`
	AccessMethodsDeferredDroppedRowsets        float64 `perfdata:"Deferred Dropped rowsets"`
	AccessMethodsDroppedRowsetCleanupsPerSec   float64 `perfdata:"Dropped rowset cleanups/sec"`
	AccessMethodsDroppedRowsetsSkippedPerSec   float64 `perfdata:"Dropped rowsets skipped/sec"`
	AccessMethodsExtentDeallocationsPerSec     float64 `perfdata:"Extent Deallocations/sec"`
	AccessMethodsExtentsAllocatedPerSec        float64 `perfdata:"Extents Allocated/sec"`
	AccessMethodsFailedAUCleanupBatchesPerSec  float64 `perfdata:"Failed AU cleanup batches/sec"`
	AccessMethodsFailedLeafPageCookie          float64 `perfdata:"Failed leaf page cookie"`
	AccessMethodsFailedTreePageCookie          float64 `perfdata:"Failed tree page cookie"`
	AccessMethodsForwardedRecordsPerSec        float64 `perfdata:"Forwarded Records/sec"`
	AccessMethodsFreeSpacePageFetchesPerSec    float64 `perfdata:"FreeSpace Page Fetches/sec"`
	AccessMethodsFreeSpaceScansPerSec          float64 `perfdata:"FreeSpace Scans/sec"`
	AccessMethodsFullScansPerSec               float64 `perfdata:"Full Scans/sec"`
	AccessMethodsIndexSearchesPerSec           float64 `perfdata:"Index Searches/sec"`
	AccessMethodsInSysXactWaitsPerSec          float64 `perfdata:"InSysXact waits/sec"`
	AccessMethodsLobHandleCreateCount          float64 `perfdata:"LobHandle Create Count"`
	AccessMethodsLobHandleDestroyCount         float64 `perfdata:"LobHandle Destroy Count"`
	AccessMethodsLobSSProviderCreateCount      float64 `perfdata:"LobSS Provider Create Count"`
	AccessMethodsLobSSProviderDestroyCount     float64 `perfdata:"LobSS Provider Destroy Count"`
	AccessMethodsLobSSProviderTruncationCount  float64 `perfdata:"LobSS Provider Truncation Count"`
	AccessMethodsMixedPageAllocationsPerSec    float64 `perfdata:"Mixed page allocations/sec"`
	AccessMethodsPageCompressionAttemptsPerSec float64 `perfdata:"Page compression attempts/sec"`
	AccessMethodsPageDeallocationsPerSec       float64 `perfdata:"Page Deallocations/sec"`
	AccessMethodsPagesAllocatedPerSec          float64 `perfdata:"Pages Allocated/sec"`
	AccessMethodsPagesCompressedPerSec         float64 `perfdata:"Pages compressed/sec"`
	AccessMethodsPageSplitsPerSec              float64 `perfdata:"Page Splits/sec"`
	AccessMethodsProbeScansPerSec              float64 `perfdata:"Probe Scans/sec"`
	AccessMethodsRangeScansPerSec              float64 `perfdata:"Range Scans/sec"`
	AccessMethodsScanPointRevalidationsPerSec  float64 `perfdata:"Scan Point Revalidations/sec"`
	AccessMethodsSkippedGhostedRecordsPerSec   float64 `perfdata:"Skipped Ghosted Records/sec"`
	AccessMethodsTableLockEscalationsPerSec    float64 `perfdata:"Table Lock Escalations/sec"`
	AccessMethodsUsedLeafPageCookie            float64 `perfdata:"Used leaf page cookie"`
	AccessMethodsUsedTreePageCookie            float64 `perfdata:"Used tree page cookie"`
	AccessMethodsWorkfilesCreatedPerSec        float64 `perfdata:"Workfiles Created/sec"`
	AccessMethodsWorktablesCreatedPerSec       float64 `perfdata:"Worktables Created/sec"`
	AccessMethodsWorktablesFromCacheRatio      float64 `perfdata:"Worktables From Cache Ratio"`
	AccessMethodsWorktablesFromCacheRatioBase  float64 `perfdata:"Worktables From Cache Base,secondvalue"`
}

func (c *Collector) buildAccessMethods() error {
	var err error

	c.accessMethodsPerfDataCollectors = make(map[mssqlInstance]*pdh.Collector, len(c.mssqlInstances))
	errs := make([]error, 0, len(c.mssqlInstances))

	for _, sqlInstance := range c.mssqlInstances {
		c.accessMethodsPerfDataCollectors[sqlInstance], err = pdh.NewCollector[perfDataCounterValuesAccessMethods](pdh.CounterTypeRaw, c.mssqlGetPerfObjectName(sqlInstance, "Access Methods"), nil)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to create AccessMethods collector for instance %s: %w", sqlInstance.name, err))
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

	return errors.Join(errs...)
}

func (c *Collector) collectAccessMethods(ch chan<- prometheus.Metric) error {
	return c.collect(ch, subCollectorAccessMethods, c.accessMethodsPerfDataCollectors, c.collectAccessMethodsInstance)
}

func (c *Collector) collectAccessMethodsInstance(ch chan<- prometheus.Metric, sqlInstance mssqlInstance, perfDataCollector *pdh.Collector) error {
	err := perfDataCollector.Collect(&c.accessMethodsPerfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect %s metrics: %w", c.mssqlGetPerfObjectName(sqlInstance, "AccessMethods"), err)
	}

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsAUcleanupbatches,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsAUCleanupbatchesPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsAUcleanups,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsAUCleanupsPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsByReferenceLobCreateCount,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsByReferenceLobCreateCount,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsByReferenceLobUseCount,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsByReferenceLobUseCount,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsCountLobReadahead,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsCountLobReadahead,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsCountPullInRow,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsCountPullInRow,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsCountPushOffRow,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsCountPushOffRow,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsDeferreddroppedAUs,
		prometheus.GaugeValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsDeferredDroppedAUs,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsDeferredDroppedrowsets,
		prometheus.GaugeValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsDeferredDroppedRowsets,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsDroppedrowsetcleanups,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsDroppedRowsetCleanupsPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsDroppedrowsetsskipped,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsDroppedRowsetsSkippedPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsExtentDeallocations,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsExtentDeallocationsPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsExtentsAllocated,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsExtentsAllocatedPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsFailedAUcleanupbatches,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsFailedAUCleanupBatchesPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsFailedleafpagecookie,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsFailedLeafPageCookie,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsFailedtreepagecookie,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsFailedTreePageCookie,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsForwardedRecords,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsForwardedRecordsPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsFreeSpacePageFetches,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsFreeSpacePageFetchesPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsFreeSpaceScans,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsFreeSpaceScansPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsFullScans,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsFullScansPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsIndexSearches,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsIndexSearchesPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsInSysXactwaits,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsInSysXactWaitsPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsLobHandleCreateCount,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsLobHandleCreateCount,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsLobHandleDestroyCount,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsLobHandleDestroyCount,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsLobSSProviderCreateCount,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsLobSSProviderCreateCount,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsLobSSProviderDestroyCount,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsLobSSProviderDestroyCount,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsLobSSProviderTruncationCount,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsLobSSProviderTruncationCount,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsMixedPageAllocations,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsMixedPageAllocationsPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsPageCompressionAttempts,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsPageCompressionAttemptsPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsPageDeallocations,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsPageDeallocationsPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsPagesAllocated,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsPagesAllocatedPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsPagesCompressed,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsPagesCompressedPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsPageSplits,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsPageSplitsPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsProbeScans,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsProbeScansPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsRangeScans,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsRangeScansPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsScanPointRevalidations,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsScanPointRevalidationsPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsSkippedGhostedRecords,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsSkippedGhostedRecordsPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsTableLockEscalations,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsTableLockEscalationsPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsUsedleafpagecookie,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsUsedLeafPageCookie,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsUsedtreepagecookie,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsUsedTreePageCookie,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsWorkfilesCreated,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsWorkfilesCreatedPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsWorktablesCreated,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsWorktablesCreatedPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsWorktablesFromCacheHits,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsWorktablesFromCacheRatio,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMethodsWorktablesFromCacheLookups,
		prometheus.CounterValue,
		c.accessMethodsPerfDataObject[0].AccessMethodsWorktablesFromCacheRatioBase,
		sqlInstance.name,
	)

	return nil
}

func (c *Collector) closeAccessMethods() {
	for _, perfDataCollector := range c.accessMethodsPerfDataCollectors {
		perfDataCollector.Close()
	}
}

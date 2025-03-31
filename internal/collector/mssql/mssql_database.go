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

type collectorDatabases struct {
	databasesPerfDataCollectors     map[mssqlInstance]*pdh.Collector
	databasesPerfDataCollectors2019 map[mssqlInstance]*pdh.Collector
	databasesPerfDataObject         []perfDataCounterValuesDatabases
	databasesPerfDataObject2019     []perfDataCounterValuesDatabases2019

	databasesActiveParallelRedoThreads       *prometheus.Desc
	databasesActiveTransactions              *prometheus.Desc
	databasesBackupPerRestoreThroughput      *prometheus.Desc
	databasesBulkCopyRows                    *prometheus.Desc
	databasesBulkCopyThroughput              *prometheus.Desc
	databasesCommitTableEntries              *prometheus.Desc
	databasesDataFilesSizeKB                 *prometheus.Desc
	databasesDBCCLogicalScanBytes            *prometheus.Desc
	databasesGroupCommitTime                 *prometheus.Desc
	databasesLogBytesFlushed                 *prometheus.Desc
	databasesLogCacheHits                    *prometheus.Desc
	databasesLogCacheLookups                 *prometheus.Desc
	databasesLogCacheReads                   *prometheus.Desc
	databasesLogFilesSizeKB                  *prometheus.Desc
	databasesLogFilesUsedSizeKB              *prometheus.Desc
	databasesLogFlushes                      *prometheus.Desc
	databasesLogFlushWaits                   *prometheus.Desc
	databasesLogFlushWaitTime                *prometheus.Desc
	databasesLogFlushWriteTimeMS             *prometheus.Desc
	databasesLogGrowths                      *prometheus.Desc
	databasesLogPoolCacheMisses              *prometheus.Desc
	databasesLogPoolDiskReads                *prometheus.Desc
	databasesLogPoolHashDeletes              *prometheus.Desc
	databasesLogPoolHashInserts              *prometheus.Desc
	databasesLogPoolInvalidHashEntry         *prometheus.Desc
	databasesLogPoolLogScanPushes            *prometheus.Desc
	databasesLogPoolLogWriterPushes          *prometheus.Desc
	databasesLogPoolPushEmptyFreePool        *prometheus.Desc
	databasesLogPoolPushLowMemory            *prometheus.Desc
	databasesLogPoolPushNoFreeBuffer         *prometheus.Desc
	databasesLogPoolReqBehindTrunc           *prometheus.Desc
	databasesLogPoolRequestsOldVLF           *prometheus.Desc
	databasesLogPoolRequests                 *prometheus.Desc
	databasesLogPoolTotalActiveLogSize       *prometheus.Desc
	databasesLogPoolTotalSharedPoolSize      *prometheus.Desc
	databasesLogShrinks                      *prometheus.Desc
	databasesLogTruncations                  *prometheus.Desc
	databasesPercentLogUsed                  *prometheus.Desc
	databasesReplPendingXacts                *prometheus.Desc
	databasesReplTransRate                   *prometheus.Desc
	databasesShrinkDataMovementBytes         *prometheus.Desc
	databasesTrackedTransactions             *prometheus.Desc
	databasesTransactions                    *prometheus.Desc
	databasesWriteTransactions               *prometheus.Desc
	databasesXTPControllerDLCLatencyPerFetch *prometheus.Desc
	databasesXTPControllerDLCPeakLatency     *prometheus.Desc
	databasesXTPControllerLogProcessed       *prometheus.Desc
	databasesXTPMemoryUsedKB                 *prometheus.Desc
}

type perfDataCounterValuesDatabases struct {
	Name string

	DatabasesActiveTransactions               float64 `perfdata:"Active Transactions"`
	DatabasesBackupPerRestoreThroughputPerSec float64 `perfdata:"Backup/Restore Throughput/sec"`
	DatabasesBulkCopyRowsPerSec               float64 `perfdata:"Bulk Copy Rows/sec"`
	DatabasesBulkCopyThroughputPerSec         float64 `perfdata:"Bulk Copy Throughput/sec"`
	DatabasesCommitTableEntries               float64 `perfdata:"Commit table entries"`
	DatabasesDataFilesSizeKB                  float64 `perfdata:"Data File(s) Size (KB)"`
	DatabasesDBCCLogicalScanBytesPerSec       float64 `perfdata:"DBCC Logical Scan Bytes/sec"`
	DatabasesGroupCommitTimePerSec            float64 `perfdata:"Group Commit Time/sec"`
	DatabasesLogBytesFlushedPerSec            float64 `perfdata:"Log Bytes Flushed/sec"`
	DatabasesLogCacheHitRatio                 float64 `perfdata:"Log Cache Hit Ratio"`
	DatabasesLogCacheHitRatioBase             float64 `perfdata:"Log Cache Hit Ratio Base,secondvalue"`
	DatabasesLogCacheReadsPerSec              float64 `perfdata:"Log Cache Reads/sec"`
	DatabasesLogFilesSizeKB                   float64 `perfdata:"Log File(s) Size (KB)"`
	DatabasesLogFilesUsedSizeKB               float64 `perfdata:"Log File(s) Used Size (KB)"`
	DatabasesLogFlushesPerSec                 float64 `perfdata:"Log Flushes/sec"`
	DatabasesLogFlushWaitsPerSec              float64 `perfdata:"Log Flush Waits/sec"`
	DatabasesLogFlushWaitTime                 float64 `perfdata:"Log Flush Wait Time"`
	DatabasesLogFlushWriteTimeMS              float64 `perfdata:"Log Flush Write Time (ms)"`
	DatabasesLogGrowths                       float64 `perfdata:"Log Growths"`
	DatabasesLogPoolCacheMissesPerSec         float64 `perfdata:"Log Pool Cache Misses/sec"`
	DatabasesLogPoolDiskReadsPerSec           float64 `perfdata:"Log Pool Disk Reads/sec"`
	DatabasesLogPoolHashDeletesPerSec         float64 `perfdata:"Log Pool Hash Deletes/sec"`
	DatabasesLogPoolHashInsertsPerSec         float64 `perfdata:"Log Pool Hash Inserts/sec"`
	DatabasesLogPoolInvalidHashEntryPerSec    float64 `perfdata:"Log Pool Invalid Hash Entry/sec"`
	DatabasesLogPoolLogScanPushesPerSec       float64 `perfdata:"Log Pool Log Scan Pushes/sec"`
	DatabasesLogPoolLogWriterPushesPerSec     float64 `perfdata:"Log Pool LogWriter Pushes/sec"`
	DatabasesLogPoolPushEmptyFreePoolPerSec   float64 `perfdata:"Log Pool Push Empty FreePool/sec"`
	DatabasesLogPoolPushLowMemoryPerSec       float64 `perfdata:"Log Pool Push Low Memory/sec"`
	DatabasesLogPoolPushNoFreeBufferPerSec    float64 `perfdata:"Log Pool Push No Free Buffer/sec"`
	DatabasesLogPoolReqBehindTruncPerSec      float64 `perfdata:"Log Pool Req. Behind Trunc/sec"`
	DatabasesLogPoolRequestsOldVLFPerSec      float64 `perfdata:"Log Pool Requests Old VLF/sec"`
	DatabasesLogPoolRequestsPerSec            float64 `perfdata:"Log Pool Requests/sec"`
	DatabasesLogPoolTotalActiveLogSize        float64 `perfdata:"Log Pool Total Active Log Size"`
	DatabasesLogPoolTotalSharedPoolSize       float64 `perfdata:"Log Pool Total Shared Pool Size"`
	DatabasesLogShrinks                       float64 `perfdata:"Log Shrinks"`
	DatabasesLogTruncations                   float64 `perfdata:"Log Truncations"`
	DatabasesPercentLogUsed                   float64 `perfdata:"Percent Log Used"`
	DatabasesReplPendingXacts                 float64 `perfdata:"Repl. Pending Xacts"`
	DatabasesReplTransRate                    float64 `perfdata:"Repl. Trans. Rate"`
	DatabasesShrinkDataMovementBytesPerSec    float64 `perfdata:"Shrink Data Movement Bytes/sec"`
	DatabasesTrackedTransactionsPerSec        float64 `perfdata:"Tracked transactions/sec"`
	DatabasesTransactionsPerSec               float64 `perfdata:"Transactions/sec"`
	DatabasesWriteTransactionsPerSec          float64 `perfdata:"Write Transactions/sec"`
	DatabasesXTPControllerDLCLatencyPerFetch  float64 `perfdata:"XTP Controller DLC Latency/Fetch"`
	DatabasesXTPControllerDLCPeakLatency      float64 `perfdata:"XTP Controller DLC Peak Latency"`
	DatabasesXTPControllerLogProcessedPerSec  float64 `perfdata:"XTP Controller Log Processed/sec"`
	DatabasesXTPMemoryUsedKB                  float64 `perfdata:"XTP Memory Used (KB)"`
}

type perfDataCounterValuesDatabases2019 struct {
	Name string

	DatabasesActiveParallelRedoThreads float64 `perfdata:"Active parallel redo threads"`
}

func (c *Collector) buildDatabases() error {
	var err error

	c.databasesPerfDataCollectors = make(map[mssqlInstance]*pdh.Collector, len(c.mssqlInstances))
	c.databasesPerfDataCollectors2019 = make(map[mssqlInstance]*pdh.Collector, len(c.mssqlInstances))
	errs := make([]error, 0, len(c.mssqlInstances))

	for _, sqlInstance := range c.mssqlInstances {
		c.databasesPerfDataCollectors[sqlInstance], err = pdh.NewCollector[perfDataCounterValuesDatabases](pdh.CounterTypeRaw, c.mssqlGetPerfObjectName(sqlInstance, "Databases"), pdh.InstancesAll)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to create Databases collector for instance %s: %w", sqlInstance.name, err))
		}

		if sqlInstance.isVersionGreaterOrEqualThan(serverVersion2019) {
			c.databasesPerfDataCollectors2019[sqlInstance], err = pdh.NewCollector[perfDataCounterValuesDatabases2019](pdh.CounterTypeRaw, c.mssqlGetPerfObjectName(sqlInstance, "Databases"), pdh.InstancesAll)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to create Databases 2019 collector for instance %s: %w", sqlInstance.name, err))
			}
		}
	}

	c.databasesActiveParallelRedoThreads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_active_parallel_redo_threads"),
		"(Databases.ActiveParallelredothreads)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesActiveTransactions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_active_transactions"),
		"(Databases.ActiveTransactions)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesBackupPerRestoreThroughput = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_backup_restore_operations"),
		"(Databases.BackupPerRestoreThroughput)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesBulkCopyRows = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_bulk_copy_rows"),
		"(Databases.BulkCopyRows)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesBulkCopyThroughput = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_bulk_copy_bytes"),
		"(Databases.BulkCopyThroughput)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesCommitTableEntries = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_commit_table_entries"),
		"(Databases.Committableentries)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesDataFilesSizeKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_data_files_size_bytes"),
		"(Databases.DataFilesSizeKB)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesDBCCLogicalScanBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_dbcc_logical_scan_bytes"),
		"(Databases.DBCCLogicalScanBytes)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesGroupCommitTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_group_commit_stall_seconds"),
		"(Databases.GroupCommitTime)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesLogBytesFlushed = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_flushed_bytes"),
		"(Databases.LogBytesFlushed)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesLogCacheHits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_cache_hits"),
		"(Databases.LogCacheHitRatio)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesLogCacheLookups = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_cache_lookups"),
		"(Databases.LogCacheHitRatio_Base)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesLogCacheReads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_cache_reads"),
		"(Databases.LogCacheReads)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesLogFilesSizeKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_files_size_bytes"),
		"(Databases.LogFilesSizeKB)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesLogFilesUsedSizeKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_files_used_size_bytes"),
		"(Databases.LogFilesUsedSizeKB)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesLogFlushes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_flushes"),
		"(Databases.LogFlushes)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesLogFlushWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_flush_waits"),
		"(Databases.LogFlushWaits)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesLogFlushWaitTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_flush_wait_seconds"),
		"(Databases.LogFlushWaitTime)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesLogFlushWriteTimeMS = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_flush_write_seconds"),
		"(Databases.LogFlushWriteTimems)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesLogGrowths = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_growths"),
		"(Databases.LogGrowths)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesLogPoolCacheMisses = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_pool_cache_misses"),
		"(Databases.LogPoolCacheMisses)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesLogPoolDiskReads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_pool_disk_reads"),
		"(Databases.LogPoolDiskReads)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesLogPoolHashDeletes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_pool_hash_deletes"),
		"(Databases.LogPoolHashDeletes)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesLogPoolHashInserts = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_pool_hash_inserts"),
		"(Databases.LogPoolHashInserts)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesLogPoolInvalidHashEntry = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_pool_invalid_hash_entries"),
		"(Databases.LogPoolInvalidHashEntry)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesLogPoolLogScanPushes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_pool_log_scan_pushes"),
		"(Databases.LogPoolLogScanPushes)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesLogPoolLogWriterPushes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_pool_log_writer_pushes"),
		"(Databases.LogPoolLogWriterPushes)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesLogPoolPushEmptyFreePool = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_pool_empty_free_pool_pushes"),
		"(Databases.LogPoolPushEmptyFreePool)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesLogPoolPushLowMemory = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_pool_low_memory_pushes"),
		"(Databases.LogPoolPushLowMemory)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesLogPoolPushNoFreeBuffer = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_pool_no_free_buffer_pushes"),
		"(Databases.LogPoolPushNoFreeBuffer)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesLogPoolReqBehindTrunc = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_pool_req_behind_trunc"),
		"(Databases.LogPoolReqBehindTrunc)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesLogPoolRequestsOldVLF = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_pool_requests_old_vlf"),
		"(Databases.LogPoolRequestsOldVLF)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesLogPoolRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_pool_requests"),
		"(Databases.LogPoolRequests)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesLogPoolTotalActiveLogSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_pool_total_active_log_bytes"),
		"(Databases.LogPoolTotalActiveLogSize)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesLogPoolTotalSharedPoolSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_pool_total_shared_pool_bytes"),
		"(Databases.LogPoolTotalSharedPoolSize)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesLogShrinks = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_shrinks"),
		"(Databases.LogShrinks)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesLogTruncations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_truncations"),
		"(Databases.LogTruncations)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesPercentLogUsed = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_used_percent"),
		"(Databases.PercentLogUsed)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesReplPendingXacts = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_pending_repl_transactions"),
		"(Databases.ReplPendingTransactions)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesReplTransRate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_repl_transactions"),
		"(Databases.ReplTranactions)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesShrinkDataMovementBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_shrink_data_movement_bytes"),
		"(Databases.ShrinkDataMovementBytes)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesTrackedTransactions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_tracked_transactions"),
		"(Databases.Trackedtransactions)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesTransactions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_transactions"),
		"(Databases.Transactions)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesWriteTransactions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_write_transactions"),
		"(Databases.WriteTransactions)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesXTPControllerDLCLatencyPerFetch = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_xtp_controller_dlc_fetch_latency_seconds"),
		"(Databases.XTPControllerDLCLatencyPerFetch)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesXTPControllerDLCPeakLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_xtp_controller_dlc_peak_latency_seconds"),
		"(Databases.XTPControllerDLCPeakLatency)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesXTPControllerLogProcessed = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_xtp_controller_log_processed_bytes"),
		"(Databases.XTPControllerLogProcessed)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.databasesXTPMemoryUsedKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_xtp_memory_used_bytes"),
		"(Databases.XTPMemoryUsedKB)",
		[]string{"mssql_instance", "database"},
		nil,
	)

	return errors.Join(errs...)
}

func (c *Collector) collectDatabases(ch chan<- prometheus.Metric) error {
	return errors.Join(
		c.collect(ch, subCollectorDatabases, c.databasesPerfDataCollectors, c.collectDatabasesInstance),
		c.collect(ch, "", c.databasesPerfDataCollectors2019, c.collectDatabasesInstance2019),
	)
}

func (c *Collector) collectDatabasesInstance(ch chan<- prometheus.Metric, sqlInstance mssqlInstance, perfDataCollector *pdh.Collector) error {
	err := perfDataCollector.Collect(&c.databasesPerfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect %s metrics: %w", c.mssqlGetPerfObjectName(sqlInstance, "Databases"), err)
	}

	for _, data := range c.databasesPerfDataObject {
		ch <- prometheus.MustNewConstMetric(
			c.databasesActiveTransactions,
			prometheus.GaugeValue,
			data.DatabasesActiveTransactions,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesBackupPerRestoreThroughput,
			prometheus.CounterValue,
			data.DatabasesBackupPerRestoreThroughputPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesBulkCopyRows,
			prometheus.CounterValue,
			data.DatabasesBulkCopyRowsPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesBulkCopyThroughput,
			prometheus.CounterValue,
			data.DatabasesBulkCopyThroughputPerSec*1024,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesCommitTableEntries,
			prometheus.GaugeValue,
			data.DatabasesCommitTableEntries,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesDataFilesSizeKB,
			prometheus.GaugeValue,
			data.DatabasesDataFilesSizeKB*1024,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesDBCCLogicalScanBytes,
			prometheus.CounterValue,
			data.DatabasesDBCCLogicalScanBytesPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesGroupCommitTime,
			prometheus.CounterValue,
			data.DatabasesGroupCommitTimePerSec/1000000.0,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogBytesFlushed,
			prometheus.CounterValue,
			data.DatabasesLogBytesFlushedPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogCacheHits,
			prometheus.GaugeValue,
			data.DatabasesLogCacheHitRatio,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogCacheLookups,
			prometheus.GaugeValue,
			data.DatabasesLogCacheHitRatioBase,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogCacheReads,
			prometheus.CounterValue,
			data.DatabasesLogCacheReadsPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogFilesSizeKB,
			prometheus.GaugeValue,
			data.DatabasesLogFilesSizeKB*1024,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogFilesUsedSizeKB,
			prometheus.GaugeValue,
			data.DatabasesLogFilesUsedSizeKB*1024,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogFlushes,
			prometheus.CounterValue,
			data.DatabasesLogFlushesPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogFlushWaits,
			prometheus.CounterValue,
			data.DatabasesLogFlushWaitsPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogFlushWaitTime,
			prometheus.GaugeValue,
			data.DatabasesLogFlushWaitTime/1000.0,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogFlushWriteTimeMS,
			prometheus.GaugeValue,
			data.DatabasesLogFlushWriteTimeMS/1000.0,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogGrowths,
			prometheus.GaugeValue,
			data.DatabasesLogGrowths,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolCacheMisses,
			prometheus.CounterValue,
			data.DatabasesLogPoolCacheMissesPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolDiskReads,
			prometheus.CounterValue,
			data.DatabasesLogPoolDiskReadsPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolHashDeletes,
			prometheus.CounterValue,
			data.DatabasesLogPoolHashDeletesPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolHashInserts,
			prometheus.CounterValue,
			data.DatabasesLogPoolHashInsertsPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolInvalidHashEntry,
			prometheus.CounterValue,
			data.DatabasesLogPoolInvalidHashEntryPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolLogScanPushes,
			prometheus.CounterValue,
			data.DatabasesLogPoolLogScanPushesPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolLogWriterPushes,
			prometheus.CounterValue,
			data.DatabasesLogPoolLogWriterPushesPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolPushEmptyFreePool,
			prometheus.CounterValue,
			data.DatabasesLogPoolPushEmptyFreePoolPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolPushLowMemory,
			prometheus.CounterValue,
			data.DatabasesLogPoolPushLowMemoryPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolPushNoFreeBuffer,
			prometheus.CounterValue,
			data.DatabasesLogPoolPushNoFreeBufferPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolReqBehindTrunc,
			prometheus.CounterValue,
			data.DatabasesLogPoolReqBehindTruncPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolRequestsOldVLF,
			prometheus.CounterValue,
			data.DatabasesLogPoolRequestsOldVLFPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolRequests,
			prometheus.CounterValue,
			data.DatabasesLogPoolRequestsPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolTotalActiveLogSize,
			prometheus.GaugeValue,
			data.DatabasesLogPoolTotalActiveLogSize,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolTotalSharedPoolSize,
			prometheus.GaugeValue,
			data.DatabasesLogPoolTotalSharedPoolSize,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogShrinks,
			prometheus.GaugeValue,
			data.DatabasesLogShrinks,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogTruncations,
			prometheus.GaugeValue,
			data.DatabasesLogTruncations,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesPercentLogUsed,
			prometheus.GaugeValue,
			data.DatabasesPercentLogUsed,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesReplPendingXacts,
			prometheus.GaugeValue,
			data.DatabasesReplPendingXacts,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesReplTransRate,
			prometheus.CounterValue,
			data.DatabasesReplTransRate,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesShrinkDataMovementBytes,
			prometheus.CounterValue,
			data.DatabasesShrinkDataMovementBytesPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesTrackedTransactions,
			prometheus.CounterValue,
			data.DatabasesTrackedTransactionsPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesTransactions,
			prometheus.CounterValue,
			data.DatabasesTransactionsPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesWriteTransactions,
			prometheus.CounterValue,
			data.DatabasesWriteTransactionsPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesXTPControllerDLCLatencyPerFetch,
			prometheus.GaugeValue,
			data.DatabasesXTPControllerDLCLatencyPerFetch,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesXTPControllerDLCPeakLatency,
			prometheus.GaugeValue,
			data.DatabasesXTPControllerDLCPeakLatency*1000000.0,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesXTPControllerLogProcessed,
			prometheus.CounterValue,
			data.DatabasesXTPControllerLogProcessedPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesXTPMemoryUsedKB,
			prometheus.GaugeValue,
			data.DatabasesXTPMemoryUsedKB*1024,
			sqlInstance.name, data.Name,
		)
	}

	return nil
}

func (c *Collector) collectDatabasesInstance2019(ch chan<- prometheus.Metric, sqlInstance mssqlInstance, perfDataCollector *pdh.Collector) error {
	err := perfDataCollector.Collect(&c.databasesPerfDataObject2019)
	if err != nil {
		return fmt.Errorf("failed to collect %s metrics: %w", c.mssqlGetPerfObjectName(sqlInstance, "Databases"), err)
	}

	for _, data := range c.databasesPerfDataObject2019 {
		ch <- prometheus.MustNewConstMetric(
			c.databasesActiveParallelRedoThreads,
			prometheus.GaugeValue,
			data.DatabasesActiveParallelRedoThreads,
			sqlInstance.name, data.Name,
		)
	}

	return nil
}

func (c *Collector) closeDatabases() {
	for _, collector := range c.databasesPerfDataCollectors {
		collector.Close()
	}
}

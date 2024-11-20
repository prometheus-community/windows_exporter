//go:build windows

package mssql

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorDatabases struct {
	databasesPerfDataCollectors map[string]*perfdata.Collector

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

const (
	databasesActiveParallelRedoThreads        = "Active parallel redo threads"
	databasesActiveTransactions               = "Active Transactions"
	databasesBackupPerRestoreThroughputPerSec = "Backup/Restore Throughput/sec"
	databasesBulkCopyRowsPerSec               = "Bulk Copy Rows/sec"
	databasesBulkCopyThroughputPerSec         = "Bulk Copy Throughput/sec"
	databasesCommitTableEntries               = "Commit table entries"
	databasesDataFilesSizeKB                  = "Data File(s) Size (KB)"
	databasesDBCCLogicalScanBytesPerSec       = "DBCC Logical Scan Bytes/sec"
	databasesGroupCommitTimePerSec            = "Group Commit Time/sec"
	databasesLogBytesFlushedPerSec            = "Log Bytes Flushed/sec"
	databasesLogCacheHitRatio                 = "Log Cache Hit Ratio"
	databasesLogCacheHitRatioBase             = "Log Cache Hit Ratio Base"
	databasesLogCacheReadsPerSec              = "Log Cache Reads/sec"
	databasesLogFilesSizeKB                   = "Log File(s) Size (KB)"
	databasesLogFilesUsedSizeKB               = "Log File(s) Used Size (KB)"
	databasesLogFlushesPerSec                 = "Log Flushes/sec"
	databasesLogFlushWaitsPerSec              = "Log Flush Waits/sec"
	databasesLogFlushWaitTime                 = "Log Flush Wait Time"
	databasesLogFlushWriteTimeMS              = "Log Flush Write Time (ms)"
	databasesLogGrowths                       = "Log Growths"
	databasesLogPoolCacheMissesPerSec         = "Log Pool Cache Misses/sec"
	databasesLogPoolDiskReadsPerSec           = "Log Pool Disk Reads/sec"
	databasesLogPoolHashDeletesPerSec         = "Log Pool Hash Deletes/sec"
	databasesLogPoolHashInsertsPerSec         = "Log Pool Hash Inserts/sec"
	databasesLogPoolInvalidHashEntryPerSec    = "Log Pool Invalid Hash Entry/sec"
	databasesLogPoolLogScanPushesPerSec       = "Log Pool Log Scan Pushes/sec"
	databasesLogPoolLogWriterPushesPerSec     = "Log Pool LogWriter Pushes/sec"
	databasesLogPoolPushEmptyFreePoolPerSec   = "Log Pool Push Empty FreePool/sec"
	databasesLogPoolPushLowMemoryPerSec       = "Log Pool Push Low Memory/sec"
	databasesLogPoolPushNoFreeBufferPerSec    = "Log Pool Push No Free Buffer/sec"
	databasesLogPoolReqBehindTruncPerSec      = "Log Pool Req. Behind Trunc/sec"
	databasesLogPoolRequestsOldVLFPerSec      = "Log Pool Requests Old VLF/sec"
	databasesLogPoolRequestsPerSec            = "Log Pool Requests/sec"
	databasesLogPoolTotalActiveLogSize        = "Log Pool Total Active Log Size"
	databasesLogPoolTotalSharedPoolSize       = "Log Pool Total Shared Pool Size"
	databasesLogShrinks                       = "Log Shrinks"
	databasesLogTruncations                   = "Log Truncations"
	databasesPercentLogUsed                   = "Percent Log Used"
	databasesReplPendingXacts                 = "Repl. Pending Xacts"
	databasesReplTransRate                    = "Repl. Trans. Rate"
	databasesShrinkDataMovementBytesPerSec    = "Shrink Data Movement Bytes/sec"
	databasesTrackedTransactionsPerSec        = "Tracked transactions/sec"
	databasesTransactionsPerSec               = "Transactions/sec"
	databasesWriteTransactionsPerSec          = "Write Transactions/sec"
	databasesXTPControllerDLCLatencyPerFetch  = "XTP Controller DLC Latency/Fetch"
	databasesXTPControllerDLCPeakLatency      = "XTP Controller DLC Peak Latency"
	databasesXTPControllerLogProcessedPerSec  = "XTP Controller Log Processed/sec"
	databasesXTPMemoryUsedKB                  = "XTP Memory Used (KB)"
)

func (c *Collector) buildDatabases() error {
	var err error

	c.databasesPerfDataCollectors = make(map[string]*perfdata.Collector, len(c.mssqlInstances))
	counters := []string{
		databasesActiveParallelRedoThreads,
		databasesActiveTransactions,
		databasesBackupPerRestoreThroughputPerSec,
		databasesBulkCopyRowsPerSec,
		databasesBulkCopyThroughputPerSec,
		databasesCommitTableEntries,
		databasesDataFilesSizeKB,
		databasesDBCCLogicalScanBytesPerSec,
		databasesGroupCommitTimePerSec,
		databasesLogBytesFlushedPerSec,
		databasesLogCacheHitRatio,
		databasesLogCacheHitRatioBase,
		databasesLogCacheReadsPerSec,
		databasesLogFilesSizeKB,
		databasesLogFilesUsedSizeKB,
		databasesLogFlushesPerSec,
		databasesLogFlushWaitsPerSec,
		databasesLogFlushWaitTime,
		databasesLogFlushWriteTimeMS,
		databasesLogGrowths,
		databasesLogPoolCacheMissesPerSec,
		databasesLogPoolDiskReadsPerSec,
		databasesLogPoolHashDeletesPerSec,
		databasesLogPoolHashInsertsPerSec,
		databasesLogPoolInvalidHashEntryPerSec,
		databasesLogPoolLogScanPushesPerSec,
		databasesLogPoolLogWriterPushesPerSec,
		databasesLogPoolPushEmptyFreePoolPerSec,
		databasesLogPoolPushLowMemoryPerSec,
		databasesLogPoolPushNoFreeBufferPerSec,
		databasesLogPoolReqBehindTruncPerSec,
		databasesLogPoolRequestsOldVLFPerSec,
		databasesLogPoolRequestsPerSec,
		databasesLogPoolTotalActiveLogSize,
		databasesLogPoolTotalSharedPoolSize,
		databasesLogShrinks,
		databasesLogTruncations,
		databasesPercentLogUsed,
		databasesReplPendingXacts,
		databasesReplTransRate,
		databasesShrinkDataMovementBytesPerSec,
		databasesTrackedTransactionsPerSec,
		databasesTransactionsPerSec,
		databasesWriteTransactionsPerSec,
		databasesXTPControllerDLCLatencyPerFetch,
		databasesXTPControllerDLCPeakLatency,
		databasesXTPControllerLogProcessedPerSec,
		databasesXTPMemoryUsedKB,
	}

	for sqlInstance := range c.mssqlInstances {
		c.databasesPerfDataCollectors[sqlInstance], err = perfdata.NewCollector(c.mssqlGetPerfObjectName(sqlInstance, "Databases"), perfdata.InstanceAll, counters)
		if err != nil {
			return fmt.Errorf("failed to create Databases collector for instance %s: %w", sqlInstance, err)
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

	return nil
}

func (c *Collector) collectDatabases(ch chan<- prometheus.Metric) error {
	return c.collect(ch, subCollectorDatabases, c.databasesPerfDataCollectors, c.collectDatabasesInstance)
}

func (c *Collector) collectDatabasesInstance(ch chan<- prometheus.Metric, sqlInstance string, perfDataCollector *perfdata.Collector) error {
	perfData, err := perfDataCollector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect %s metrics: %w", c.mssqlGetPerfObjectName(sqlInstance, "Databases"), err)
	}

	for dbName, data := range perfData {
		ch <- prometheus.MustNewConstMetric(
			c.databasesActiveParallelRedoThreads,
			prometheus.GaugeValue,
			data[databasesActiveParallelRedoThreads].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesActiveTransactions,
			prometheus.GaugeValue,
			data[databasesActiveTransactions].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesBackupPerRestoreThroughput,
			prometheus.CounterValue,
			data[databasesBackupPerRestoreThroughputPerSec].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesBulkCopyRows,
			prometheus.CounterValue,
			data[databasesBulkCopyRowsPerSec].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesBulkCopyThroughput,
			prometheus.CounterValue,
			data[databasesBulkCopyThroughputPerSec].FirstValue*1024,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesCommitTableEntries,
			prometheus.GaugeValue,
			data[databasesCommitTableEntries].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesDataFilesSizeKB,
			prometheus.GaugeValue,
			data[databasesDataFilesSizeKB].FirstValue*1024,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesDBCCLogicalScanBytes,
			prometheus.CounterValue,
			data[databasesDBCCLogicalScanBytesPerSec].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesGroupCommitTime,
			prometheus.CounterValue,
			data[databasesGroupCommitTimePerSec].FirstValue/1000000.0,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogBytesFlushed,
			prometheus.CounterValue,
			data[databasesLogBytesFlushedPerSec].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogCacheHits,
			prometheus.GaugeValue,
			data[databasesLogCacheHitRatio].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogCacheLookups,
			prometheus.GaugeValue,
			data[databasesLogCacheHitRatioBase].SecondValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogCacheReads,
			prometheus.CounterValue,
			data[databasesLogCacheReadsPerSec].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogFilesSizeKB,
			prometheus.GaugeValue,
			data[databasesLogFilesSizeKB].FirstValue*1024,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogFilesUsedSizeKB,
			prometheus.GaugeValue,
			data[databasesLogFilesUsedSizeKB].FirstValue*1024,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogFlushes,
			prometheus.CounterValue,
			data[databasesLogFlushesPerSec].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogFlushWaits,
			prometheus.CounterValue,
			data[databasesLogFlushWaitsPerSec].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogFlushWaitTime,
			prometheus.GaugeValue,
			data[databasesLogFlushWaitTime].FirstValue/1000.0,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogFlushWriteTimeMS,
			prometheus.GaugeValue,
			data[databasesLogFlushWriteTimeMS].FirstValue/1000.0,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogGrowths,
			prometheus.GaugeValue,
			data[databasesLogGrowths].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolCacheMisses,
			prometheus.CounterValue,
			data[databasesLogPoolCacheMissesPerSec].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolDiskReads,
			prometheus.CounterValue,
			data[databasesLogPoolDiskReadsPerSec].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolHashDeletes,
			prometheus.CounterValue,
			data[databasesLogPoolHashDeletesPerSec].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolHashInserts,
			prometheus.CounterValue,
			data[databasesLogPoolHashInsertsPerSec].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolInvalidHashEntry,
			prometheus.CounterValue,
			data[databasesLogPoolInvalidHashEntryPerSec].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolLogScanPushes,
			prometheus.CounterValue,
			data[databasesLogPoolLogScanPushesPerSec].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolLogWriterPushes,
			prometheus.CounterValue,
			data[databasesLogPoolLogWriterPushesPerSec].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolPushEmptyFreePool,
			prometheus.CounterValue,
			data[databasesLogPoolPushEmptyFreePoolPerSec].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolPushLowMemory,
			prometheus.CounterValue,
			data[databasesLogPoolPushLowMemoryPerSec].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolPushNoFreeBuffer,
			prometheus.CounterValue,
			data[databasesLogPoolPushNoFreeBufferPerSec].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolReqBehindTrunc,
			prometheus.CounterValue,
			data[databasesLogPoolReqBehindTruncPerSec].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolRequestsOldVLF,
			prometheus.CounterValue,
			data[databasesLogPoolRequestsOldVLFPerSec].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolRequests,
			prometheus.CounterValue,
			data[databasesLogPoolRequestsPerSec].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolTotalActiveLogSize,
			prometheus.GaugeValue,
			data[databasesLogPoolTotalActiveLogSize].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolTotalSharedPoolSize,
			prometheus.GaugeValue,
			data[databasesLogPoolTotalSharedPoolSize].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogShrinks,
			prometheus.GaugeValue,
			data[databasesLogShrinks].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogTruncations,
			prometheus.GaugeValue,
			data[databasesLogTruncations].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesPercentLogUsed,
			prometheus.GaugeValue,
			data[databasesPercentLogUsed].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesReplPendingXacts,
			prometheus.GaugeValue,
			data[databasesReplPendingXacts].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesReplTransRate,
			prometheus.CounterValue,
			data[databasesReplTransRate].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesShrinkDataMovementBytes,
			prometheus.CounterValue,
			data[databasesShrinkDataMovementBytesPerSec].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesTrackedTransactions,
			prometheus.CounterValue,
			data[databasesTrackedTransactionsPerSec].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesTransactions,
			prometheus.CounterValue,
			data[databasesTransactionsPerSec].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesWriteTransactions,
			prometheus.CounterValue,
			data[databasesWriteTransactionsPerSec].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesXTPControllerDLCLatencyPerFetch,
			prometheus.GaugeValue,
			data[databasesXTPControllerDLCLatencyPerFetch].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesXTPControllerDLCPeakLatency,
			prometheus.GaugeValue,
			data[databasesXTPControllerDLCPeakLatency].FirstValue*1000000.0,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesXTPControllerLogProcessed,
			prometheus.CounterValue,
			data[databasesXTPControllerLogProcessedPerSec].FirstValue,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesXTPMemoryUsedKB,
			prometheus.GaugeValue,
			data[databasesXTPMemoryUsedKB].FirstValue*1024,
			sqlInstance, dbName,
		)
	}

	return nil
}

func (c *Collector) closeDatabases() {
	for _, collector := range c.databasesPerfDataCollectors {
		collector.Close()
	}
}

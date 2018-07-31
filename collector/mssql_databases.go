// returns data points from the following classes:
// - Win32_PerfRawData_MSSQLSERVER_SQLServerDatabases
//   https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-databases-object?view=sql-server-2017

package collector

import (
	"fmt"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	Factories["mssql_databases"] = NewMSSQLDatabasesCollector
}

// MSSQLDatabasesCollector is a Prometheus collector for Win32_PerfRawData_{instance}_SQLServerDatabases metrics
type MSSQLDatabasesCollector struct {
	ActiveTransactions               *prometheus.Desc
	BackupPerRestoreThroughputPersec *prometheus.Desc
	BulkCopyRowsPersec               *prometheus.Desc
	BulkCopyThroughputPersec         *prometheus.Desc
	Committableentries               *prometheus.Desc
	DataFilesSizeKB                  *prometheus.Desc
	DBCCLogicalScanBytesPersec       *prometheus.Desc
	GroupCommitTimePersec            *prometheus.Desc
	LogBytesFlushedPersec            *prometheus.Desc
	LogCacheHitRatio                 *prometheus.Desc
	LogCacheReadsPersec              *prometheus.Desc
	LogFilesSizeKB                   *prometheus.Desc
	LogFilesUsedSizeKB               *prometheus.Desc
	LogFlushesPersec                 *prometheus.Desc
	LogFlushWaitsPersec              *prometheus.Desc
	LogFlushWaitTime                 *prometheus.Desc
	LogFlushWriteTimems              *prometheus.Desc
	LogGrowths                       *prometheus.Desc
	LogPoolCacheMissesPersec         *prometheus.Desc
	LogPoolDiskReadsPersec           *prometheus.Desc
	LogPoolHashDeletesPersec         *prometheus.Desc
	LogPoolHashInsertsPersec         *prometheus.Desc
	LogPoolInvalidHashEntryPersec    *prometheus.Desc
	LogPoolLogScanPushesPersec       *prometheus.Desc
	LogPoolLogWriterPushesPersec     *prometheus.Desc
	LogPoolPushEmptyFreePoolPersec   *prometheus.Desc
	LogPoolPushLowMemoryPersec       *prometheus.Desc
	LogPoolPushNoFreeBufferPersec    *prometheus.Desc
	LogPoolReqBehindTruncPersec      *prometheus.Desc
	LogPoolRequestsOldVLFPersec      *prometheus.Desc
	LogPoolRequestsPersec            *prometheus.Desc
	LogPoolTotalActiveLogSize        *prometheus.Desc
	LogPoolTotalSharedPoolSize       *prometheus.Desc
	LogShrinks                       *prometheus.Desc
	LogTruncations                   *prometheus.Desc
	PercentLogUsed                   *prometheus.Desc
	ReplPendingXacts                 *prometheus.Desc
	ReplTransRate                    *prometheus.Desc
	ShrinkDataMovementBytesPersec    *prometheus.Desc
	TrackedtransactionsPersec        *prometheus.Desc
	TransactionsPersec               *prometheus.Desc
	WriteTransactionsPersec          *prometheus.Desc
	XTPControllerDLCLatencyPerFetch  *prometheus.Desc
	XTPControllerDLCPeakLatency      *prometheus.Desc
	XTPControllerLogProcessedPersec  *prometheus.Desc
	XTPMemoryUsedKB                  *prometheus.Desc

	sqlInstances sqlInstancesType
}

// NewMSSQLDatabasesCollector ...
func NewMSSQLDatabasesCollector() (Collector, error) {

	const subsystem = "mssql_databases"
	return &MSSQLDatabasesCollector{

		// Win32_PerfRawData_{instance}_SQLServerDatabases
		ActiveTransactions: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "active_transactions"),
			"(Databases.ActiveTransactions)",
			[]string{"instance", "database"},
			nil,
		),
		BackupPerRestoreThroughputPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "backup_restore_operations"),
			"(Databases.BackupPerRestoreThroughput)",
			[]string{"instance", "database"},
			nil,
		),
		BulkCopyRowsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bulk_copy_rows"),
			"(Databases.BulkCopyRows)",
			[]string{"instance", "database"},
			nil,
		),
		BulkCopyThroughputPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bulk_copy_bytes"),
			"(Databases.BulkCopyThroughput)",
			[]string{"instance", "database"},
			nil,
		),
		Committableentries: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "commit_table_entries"),
			"(Databases.Committableentries)",
			[]string{"instance", "database"},
			nil,
		),
		DataFilesSizeKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "data_files_size_bytes"),
			"(Databases.DataFilesSizeKB)",
			[]string{"instance", "database"},
			nil,
		),
		DBCCLogicalScanBytesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbcc_logical_scan_bytes"),
			"(Databases.DBCCLogicalScanBytes)",
			[]string{"instance", "database"},
			nil,
		),
		GroupCommitTimePersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "group_commit_stall_seconds"),
			"(Databases.GroupCommitTime)",
			[]string{"instance", "database"},
			nil,
		),
		LogBytesFlushedPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_flushed_bytes"),
			"(Databases.LogBytesFlushed)",
			[]string{"instance", "database"},
			nil,
		),
		LogCacheHitRatio: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_cache_hit_ratio"),
			"(Databases.LogCacheHitRatio)",
			[]string{"instance", "database"},
			nil,
		),
		LogCacheReadsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_cache_reads"),
			"(Databases.LogCacheReads)",
			[]string{"instance", "database"},
			nil,
		),
		LogFilesSizeKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_files_size_bytes"),
			"(Databases.LogFilesSizeKB)",
			[]string{"instance", "database"},
			nil,
		),
		LogFilesUsedSizeKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_files_used_size_bytes"),
			"(Databases.LogFilesUsedSizeKB)",
			[]string{"instance", "database"},
			nil,
		),
		LogFlushesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_flushes"),
			"(Databases.LogFlushes)",
			[]string{"instance", "database"},
			nil,
		),
		LogFlushWaitsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_flush_waits"),
			"(Databases.LogFlushWaits)",
			[]string{"instance", "database"},
			nil,
		),
		LogFlushWaitTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_flush_wait_seconds"),
			"(Databases.LogFlushWaitTime)",
			[]string{"instance", "database"},
			nil,
		),
		LogFlushWriteTimems: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_flush_write_seconds"),
			"(Databases.LogFlushWriteTimems)",
			[]string{"instance", "database"},
			nil,
		),
		LogGrowths: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_growths"),
			"(Databases.LogGrowths)",
			[]string{"instance", "database"},
			nil,
		),
		LogPoolCacheMissesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_cache_misses"),
			"(Databases.LogPoolCacheMisses)",
			[]string{"instance", "database"},
			nil,
		),
		LogPoolDiskReadsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_disk_reads"),
			"(Databases.LogPoolDiskReads)",
			[]string{"instance", "database"},
			nil,
		),
		LogPoolHashDeletesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_hash_deletes"),
			"(Databases.LogPoolHashDeletes)",
			[]string{"instance", "database"},
			nil,
		),
		LogPoolHashInsertsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_hash_inserts"),
			"(Databases.LogPoolHashInserts)",
			[]string{"instance", "database"},
			nil,
		),
		LogPoolInvalidHashEntryPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_invalid_hash_entries"),
			"(Databases.LogPoolInvalidHashEntry)",
			[]string{"instance", "database"},
			nil,
		),
		LogPoolLogScanPushesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_log_scan_pushes"),
			"(Databases.LogPoolLogScanPushes)",
			[]string{"instance", "database"},
			nil,
		),
		LogPoolLogWriterPushesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_log_writer_pushes"),
			"(Databases.LogPoolLogWriterPushes)",
			[]string{"instance", "database"},
			nil,
		),
		LogPoolPushEmptyFreePoolPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_empty_free_pool_pushes"),
			"(Databases.LogPoolPushEmptyFreePool)",
			[]string{"instance", "database"},
			nil,
		),
		LogPoolPushLowMemoryPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_low_memory_pushes"),
			"(Databases.LogPoolPushLowMemory)",
			[]string{"instance", "database"},
			nil,
		),
		LogPoolPushNoFreeBufferPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_no_free_buffer_pushes"),
			"(Databases.LogPoolPushNoFreeBuffer)",
			[]string{"instance", "database"},
			nil,
		),
		LogPoolReqBehindTruncPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_req_behind_trunc"),
			"(Databases.LogPoolReqBehindTrunc)",
			[]string{"instance", "database"},
			nil,
		),
		LogPoolRequestsOldVLFPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_requests_old_vlf"),
			"(Databases.LogPoolRequestsOldVLF)",
			[]string{"instance", "database"},
			nil,
		),
		LogPoolRequestsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_requests"),
			"(Databases.LogPoolRequests)",
			[]string{"instance", "database"},
			nil,
		),
		LogPoolTotalActiveLogSize: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_total_active_log_bytes"),
			"(Databases.LogPoolTotalActiveLogSize)",
			[]string{"instance", "database"},
			nil,
		),
		LogPoolTotalSharedPoolSize: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_total_shared_pool_bytes"),
			"(Databases.LogPoolTotalSharedPoolSize)",
			[]string{"instance", "database"},
			nil,
		),
		LogShrinks: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_shrinks"),
			"(Databases.LogShrinks)",
			[]string{"instance", "database"},
			nil,
		),
		LogTruncations: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_truncations"),
			"(Databases.LogTruncations)",
			[]string{"instance", "database"},
			nil,
		),
		PercentLogUsed: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_used_percent"),
			"(Databases.PercentLogUsed)",
			[]string{"instance", "database"},
			nil,
		),
		ReplPendingXacts: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pending_repl_transactions"),
			"(Databases.ReplPendingTransactions)",
			[]string{"instance", "database"},
			nil,
		),
		ReplTransRate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "repl_transactions"),
			"(Databases.ReplTranactions)",
			[]string{"instance", "database"},
			nil,
		),
		ShrinkDataMovementBytesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "shrink_data_movement_bytes"),
			"(Databases.ShrinkDataMovementBytes)",
			[]string{"instance", "database"},
			nil,
		),
		TrackedtransactionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "tracked_transactions"),
			"(Databases.Trackedtransactions)",
			[]string{"instance", "database"},
			nil,
		),
		TransactionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "transactions"),
			"(Databases.Transactions)",
			[]string{"instance", "database"},
			nil,
		),
		WriteTransactionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "write_transactions"),
			"(Databases.WriteTransactions)",
			[]string{"instance", "database"},
			nil,
		),
		XTPControllerDLCLatencyPerFetch: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "xtp_controller_dlc_fetch_latency_seconds"),
			"(Databases.XTPControllerDLCLatencyPerFetch)",
			[]string{"instance", "database"},
			nil,
		),
		XTPControllerDLCPeakLatency: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "xtp_controller_dlc_peak_latency_seconds"),
			"(Databases.XTPControllerDLCPeakLatency)",
			[]string{"instance", "database"},
			nil,
		),
		XTPControllerLogProcessedPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "xtp_controller_log_processed_bytes"),
			"(Databases.XTPControllerLogProcessed)",
			[]string{"instance", "database"},
			nil,
		),
		XTPMemoryUsedKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "xtp_memory_used_bytes"),
			"(Databases.XTPMemoryUsedKB)",
			[]string{"instance", "database"},
			nil,
		),

		sqlInstances: getMSSQLInstances(),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *MSSQLDatabasesCollector) Collect(ch chan<- prometheus.Metric) error {
	for instance := range c.sqlInstances {
		log.Debugf("mssql collector iterating sql instance %s.", instance)

		// Win32_PerfRawData_{instance}_SQLServerDatabases
		if desc, err := c.collectDatabases(ch, instance); err != nil {
			log.Error("failed collecting MSSQL Databases metrics:", desc, err)
			return err
		}
	}

	return nil
}

type win32PerfRawDataSQLServerDatabases struct {
	Name                             string
	ActiveTransactions               uint64
	BackupPerRestoreThroughputPersec uint64
	BulkCopyRowsPersec               uint64
	BulkCopyThroughputPersec         uint64
	Committableentries               uint64
	DataFilesSizeKB                  uint64
	DBCCLogicalScanBytesPersec       uint64
	GroupCommitTimePersec            uint64
	LogBytesFlushedPersec            uint64
	LogCacheHitRatio                 uint64
	LogCacheReadsPersec              uint64
	LogFilesSizeKB                   uint64
	LogFilesUsedSizeKB               uint64
	LogFlushesPersec                 uint64
	LogFlushWaitsPersec              uint64
	LogFlushWaitTime                 uint64
	LogFlushWriteTimems              uint64
	LogGrowths                       uint64
	LogPoolCacheMissesPersec         uint64
	LogPoolDiskReadsPersec           uint64
	LogPoolHashDeletesPersec         uint64
	LogPoolHashInsertsPersec         uint64
	LogPoolInvalidHashEntryPersec    uint64
	LogPoolLogScanPushesPersec       uint64
	LogPoolLogWriterPushesPersec     uint64
	LogPoolPushEmptyFreePoolPersec   uint64
	LogPoolPushLowMemoryPersec       uint64
	LogPoolPushNoFreeBufferPersec    uint64
	LogPoolReqBehindTruncPersec      uint64
	LogPoolRequestsOldVLFPersec      uint64
	LogPoolRequestsPersec            uint64
	LogPoolTotalActiveLogSize        uint64
	LogPoolTotalSharedPoolSize       uint64
	LogShrinks                       uint64
	LogTruncations                   uint64
	PercentLogUsed                   uint64
	ReplPendingXacts                 uint64
	ReplTransRate                    uint64
	ShrinkDataMovementBytesPersec    uint64
	TrackedtransactionsPersec        uint64
	TransactionsPersec               uint64
	WriteTransactionsPersec          uint64
	XTPControllerDLCLatencyPerFetch  uint64
	XTPControllerDLCPeakLatency      uint64
	XTPControllerLogProcessedPersec  uint64
	XTPMemoryUsedKB                  uint64
}

func (c *MSSQLDatabasesCollector) collectDatabases(ch chan<- prometheus.Metric, sqlInstance string) (*prometheus.Desc, error) {
	var dst []win32PerfRawDataSQLServerDatabases
	class := fmt.Sprintf("Win32_PerfRawData_%s_SQLServerDatabases", sqlInstance)

	q := queryAllForClassWhere(&dst, class, `Name <> '_Total'`)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	for _, v := range dst {
		dbName := v.Name

		ch <- prometheus.MustNewConstMetric(
			c.ActiveTransactions,
			prometheus.GaugeValue,
			float64(v.ActiveTransactions),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BackupPerRestoreThroughputPersec,
			prometheus.CounterValue,
			float64(v.BackupPerRestoreThroughputPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BulkCopyRowsPersec,
			prometheus.CounterValue,
			float64(v.BulkCopyRowsPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BulkCopyThroughputPersec,
			prometheus.CounterValue,
			float64(v.BulkCopyThroughputPersec)*1024,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Committableentries,
			prometheus.GaugeValue,
			float64(v.Committableentries),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DataFilesSizeKB,
			prometheus.GaugeValue,
			float64(v.DataFilesSizeKB*1024),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBCCLogicalScanBytesPersec,
			prometheus.CounterValue,
			float64(v.DBCCLogicalScanBytesPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GroupCommitTimePersec,
			prometheus.CounterValue,
			float64(v.GroupCommitTimePersec)/1000000.0,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogBytesFlushedPersec,
			prometheus.CounterValue,
			float64(v.LogBytesFlushedPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogCacheHitRatio,
			prometheus.GaugeValue,
			float64(v.LogCacheHitRatio),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogCacheReadsPersec,
			prometheus.CounterValue,
			float64(v.LogCacheReadsPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogFilesSizeKB,
			prometheus.GaugeValue,
			float64(v.LogFilesSizeKB*1024),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogFilesUsedSizeKB,
			prometheus.GaugeValue,
			float64(v.LogFilesUsedSizeKB*1024),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogFlushesPersec,
			prometheus.CounterValue,
			float64(v.LogFlushesPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogFlushWaitsPersec,
			prometheus.CounterValue,
			float64(v.LogFlushWaitsPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogFlushWaitTime,
			prometheus.GaugeValue,
			float64(v.LogFlushWaitTime)/1000.0,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogFlushWriteTimems,
			prometheus.GaugeValue,
			float64(v.LogFlushWriteTimems)/1000.0,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogGrowths,
			prometheus.GaugeValue,
			float64(v.LogGrowths),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolCacheMissesPersec,
			prometheus.CounterValue,
			float64(v.LogPoolCacheMissesPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolDiskReadsPersec,
			prometheus.CounterValue,
			float64(v.LogPoolDiskReadsPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolHashDeletesPersec,
			prometheus.CounterValue,
			float64(v.LogPoolHashDeletesPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolHashInsertsPersec,
			prometheus.CounterValue,
			float64(v.LogPoolHashInsertsPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolInvalidHashEntryPersec,
			prometheus.CounterValue,
			float64(v.LogPoolInvalidHashEntryPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolLogScanPushesPersec,
			prometheus.CounterValue,
			float64(v.LogPoolLogScanPushesPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolLogWriterPushesPersec,
			prometheus.CounterValue,
			float64(v.LogPoolLogWriterPushesPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolPushEmptyFreePoolPersec,
			prometheus.CounterValue,
			float64(v.LogPoolPushEmptyFreePoolPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolPushLowMemoryPersec,
			prometheus.CounterValue,
			float64(v.LogPoolPushLowMemoryPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolPushNoFreeBufferPersec,
			prometheus.CounterValue,
			float64(v.LogPoolPushNoFreeBufferPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolReqBehindTruncPersec,
			prometheus.CounterValue,
			float64(v.LogPoolReqBehindTruncPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolRequestsOldVLFPersec,
			prometheus.CounterValue,
			float64(v.LogPoolRequestsOldVLFPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolRequestsPersec,
			prometheus.CounterValue,
			float64(v.LogPoolRequestsPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolTotalActiveLogSize,
			prometheus.GaugeValue,
			float64(v.LogPoolTotalActiveLogSize),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolTotalSharedPoolSize,
			prometheus.GaugeValue,
			float64(v.LogPoolTotalSharedPoolSize),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogShrinks,
			prometheus.GaugeValue,
			float64(v.LogShrinks),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogTruncations,
			prometheus.GaugeValue,
			float64(v.LogTruncations),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PercentLogUsed,
			prometheus.GaugeValue,
			float64(v.PercentLogUsed),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReplPendingXacts,
			prometheus.GaugeValue,
			float64(v.ReplPendingXacts),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReplTransRate,
			prometheus.CounterValue,
			float64(v.ReplTransRate),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ShrinkDataMovementBytesPersec,
			prometheus.CounterValue,
			float64(v.ShrinkDataMovementBytesPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TrackedtransactionsPersec,
			prometheus.CounterValue,
			float64(v.TrackedtransactionsPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TransactionsPersec,
			prometheus.CounterValue,
			float64(v.TransactionsPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WriteTransactionsPersec,
			prometheus.CounterValue,
			float64(v.WriteTransactionsPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.XTPControllerDLCLatencyPerFetch,
			prometheus.GaugeValue,
			float64(v.XTPControllerDLCLatencyPerFetch),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.XTPControllerDLCPeakLatency,
			prometheus.GaugeValue,
			float64(v.XTPControllerDLCPeakLatency)*1000000.0,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.XTPControllerLogProcessedPersec,
			prometheus.CounterValue,
			float64(v.XTPControllerLogProcessedPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.XTPMemoryUsedKB,
			prometheus.GaugeValue,
			float64(v.XTPMemoryUsedKB*1024),
			sqlInstance, dbName,
		)
	}

	return nil, nil
}

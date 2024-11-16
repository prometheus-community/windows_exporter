//go:build windows

package mssql

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorDatabaseReplica struct {
	dbReplicaPerfDataCollectors map[string]*perfdata.Collector

	dbReplicaDatabaseFlowControlDelay  *prometheus.Desc
	dbReplicaDatabaseFlowControls      *prometheus.Desc
	dbReplicaFileBytesReceived         *prometheus.Desc
	dbReplicaGroupCommits              *prometheus.Desc
	dbReplicaGroupCommitTime           *prometheus.Desc
	dbReplicaLogApplyPendingQueue      *prometheus.Desc
	dbReplicaLogApplyReadyQueue        *prometheus.Desc
	dbReplicaLogBytesCompressed        *prometheus.Desc
	dbReplicaLogBytesDecompressed      *prometheus.Desc
	dbReplicaLogBytesReceived          *prometheus.Desc
	dbReplicaLogCompressionCachehits   *prometheus.Desc
	dbReplicaLogCompressionCachemisses *prometheus.Desc
	dbReplicaLogCompressions           *prometheus.Desc
	dbReplicaLogDecompressions         *prometheus.Desc
	dbReplicaLogremainingforundo       *prometheus.Desc
	dbReplicaLogSendQueue              *prometheus.Desc
	dbReplicaMirroredWritetransactions *prometheus.Desc
	dbReplicaRecoveryQueue             *prometheus.Desc
	dbReplicaRedoblocked               *prometheus.Desc
	dbReplicaRedoBytesRemaining        *prometheus.Desc
	dbReplicaRedoneBytes               *prometheus.Desc
	dbReplicaRedones                   *prometheus.Desc
	dbReplicaTotalLogrequiringundo     *prometheus.Desc
	dbReplicaTransactionDelay          *prometheus.Desc
}

const (
	dbReplicaDatabaseFlowControlDelay        = "Database Flow Control Delay"
	dbReplicaDatabaseFlowControlsPerSec      = "Database Flow Controls/sec"
	dbReplicaFileBytesReceivedPerSec         = "File Bytes Received/sec"
	dbReplicaGroupCommitsPerSec              = "Group Commits/Sec"
	dbReplicaGroupCommitTime                 = "Group Commit Time"
	dbReplicaLogApplyPendingQueue            = "Log Apply Pending Queue"
	dbReplicaLogApplyReadyQueue              = "Log Apply Ready Queue"
	dbReplicaLogBytesCompressedPerSec        = "Log Bytes Compressed/sec"
	dbReplicaLogBytesDecompressedPerSec      = "Log Bytes Decompressed/sec"
	dbReplicaLogBytesReceivedPerSec          = "Log Bytes Received/sec"
	dbReplicaLogCompressionCacheHitsPerSec   = "Log Compression Cache hits/sec"
	dbReplicaLogCompressionCacheMissesPerSec = "Log Compression Cache misses/sec"
	dbReplicaLogCompressionsPerSec           = "Log Compressions/sec"
	dbReplicaLogDecompressionsPerSec         = "Log Decompressions/sec"
	dbReplicaLogRemainingForUndo             = "Log remaining for undo"
	dbReplicaLogSendQueue                    = "Log Send Queue"
	dbReplicaMirroredWriteTransactionsPerSec = "Mirrored Write Transactions/sec"
	dbReplicaRecoveryQueue                   = "Recovery Queue"
	dbReplicaRedoBlockedPerSec               = "Redo blocked/sec"
	dbReplicaRedoBytesRemaining              = "Redo Bytes Remaining"
	dbReplicaRedoneBytesPerSec               = "Redone Bytes/sec"
	dbReplicaRedonesPerSec                   = "Redones/sec"
	dbReplicaTotalLogRequiringUndo           = "Total Log requiring undo"
	dbReplicaTransactionDelay                = "Transaction Delay"
)

func (c *Collector) buildDatabaseReplica() error {
	var err error

	c.dbReplicaPerfDataCollectors = make(map[string]*perfdata.Collector, len(c.mssqlInstances))
	counters := []string{
		dbReplicaDatabaseFlowControlDelay,
		dbReplicaDatabaseFlowControlsPerSec,
		dbReplicaFileBytesReceivedPerSec,
		dbReplicaGroupCommitsPerSec,
		dbReplicaGroupCommitTime,
		dbReplicaLogApplyPendingQueue,
		dbReplicaLogApplyReadyQueue,
		dbReplicaLogBytesCompressedPerSec,
		dbReplicaLogBytesDecompressedPerSec,
		dbReplicaLogBytesReceivedPerSec,
		dbReplicaLogCompressionCacheHitsPerSec,
		dbReplicaLogCompressionCacheMissesPerSec,
		dbReplicaLogCompressionsPerSec,
		dbReplicaLogDecompressionsPerSec,
		dbReplicaLogRemainingForUndo,
		dbReplicaLogSendQueue,
		dbReplicaMirroredWriteTransactionsPerSec,
		dbReplicaRecoveryQueue,
		dbReplicaRedoBlockedPerSec,
		dbReplicaRedoBytesRemaining,
		dbReplicaRedoneBytesPerSec,
		dbReplicaRedonesPerSec,
		dbReplicaTotalLogRequiringUndo,
		dbReplicaTransactionDelay,
	}

	for sqlInstance := range c.mssqlInstances {
		c.dbReplicaPerfDataCollectors[sqlInstance], err = perfdata.NewCollector(c.mssqlGetPerfObjectName(sqlInstance, "Database Replica"), perfdata.InstanceAll, counters)
		if err != nil {
			return fmt.Errorf("failed to create Database Replica collector for instance %s: %w", sqlInstance, err)
		}
	}

	// Win32_PerfRawData_{instance}_SQLServerDatabaseReplica
	c.dbReplicaDatabaseFlowControlDelay = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_database_flow_control_wait_seconds"),
		"(DatabaseReplica.DatabaseFlowControlDelay)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.dbReplicaDatabaseFlowControls = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_database_initiated_flow_controls"),
		"(DatabaseReplica.DatabaseFlowControls)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.dbReplicaFileBytesReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_received_file_bytes"),
		"(DatabaseReplica.FileBytesReceived)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.dbReplicaGroupCommits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_group_commits"),
		"(DatabaseReplica.GroupCommits)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.dbReplicaGroupCommitTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_group_commit_stall_seconds"),
		"(DatabaseReplica.GroupCommitTime)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.dbReplicaLogApplyPendingQueue = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_log_apply_pending_queue"),
		"(DatabaseReplica.LogApplyPendingQueue)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.dbReplicaLogApplyReadyQueue = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_log_apply_ready_queue"),
		"(DatabaseReplica.LogApplyReadyQueue)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.dbReplicaLogBytesCompressed = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_log_compressed_bytes"),
		"(DatabaseReplica.LogBytesCompressed)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.dbReplicaLogBytesDecompressed = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_log_decompressed_bytes"),
		"(DatabaseReplica.LogBytesDecompressed)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.dbReplicaLogBytesReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_log_received_bytes"),
		"(DatabaseReplica.LogBytesReceived)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.dbReplicaLogCompressionCachehits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_log_compression_cachehits"),
		"(DatabaseReplica.LogCompressionCachehits)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.dbReplicaLogCompressionCachemisses = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_log_compression_cachemisses"),
		"(DatabaseReplica.LogCompressionCachemisses)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.dbReplicaLogCompressions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_log_compressions"),
		"(DatabaseReplica.LogCompressions)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.dbReplicaLogDecompressions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_log_decompressions"),
		"(DatabaseReplica.LogDecompressions)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.dbReplicaLogremainingforundo = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_log_remaining_for_undo"),
		"(DatabaseReplica.Logremainingforundo)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.dbReplicaLogSendQueue = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_log_send_queue"),
		"(DatabaseReplica.LogSendQueue)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.dbReplicaMirroredWritetransactions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_mirrored_write_transactions"),
		"(DatabaseReplica.MirroredWriteTransactions)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.dbReplicaRecoveryQueue = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_recovery_queue_records"),
		"(DatabaseReplica.RecoveryQueue)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.dbReplicaRedoblocked = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_redo_blocks"),
		"(DatabaseReplica.Redoblocked)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.dbReplicaRedoBytesRemaining = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_redo_remaining_bytes"),
		"(DatabaseReplica.RedoBytesRemaining)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.dbReplicaRedoneBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_redone_bytes"),
		"(DatabaseReplica.RedoneBytes)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.dbReplicaRedones = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_redones"),
		"(DatabaseReplica.Redones)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.dbReplicaTotalLogrequiringundo = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_total_log_requiring_undo"),
		"(DatabaseReplica.TotalLogrequiringundo)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.dbReplicaTransactionDelay = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_transaction_delay_seconds"),
		"(DatabaseReplica.TransactionDelay)",
		[]string{"mssql_instance", "replica"},
		nil,
	)

	return nil
}

func (c *Collector) collectDatabaseReplica(ch chan<- prometheus.Metric) error {
	return c.collect(ch, subCollectorDatabaseReplica, c.dbReplicaPerfDataCollectors, c.collectDatabaseReplicaInstance)
}

func (c *Collector) collectDatabaseReplicaInstance(ch chan<- prometheus.Metric, sqlInstance string, perfDataCollector *perfdata.Collector) error {
	perfData, err := perfDataCollector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect %s metrics: %w", c.mssqlGetPerfObjectName(sqlInstance, "Database Replica"), err)
	}

	for replicaName, data := range perfData {
		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaDatabaseFlowControlDelay,
			prometheus.GaugeValue,
			data[dbReplicaDatabaseFlowControlDelay].FirstValue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaDatabaseFlowControls,
			prometheus.CounterValue,
			data[dbReplicaDatabaseFlowControlsPerSec].FirstValue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaFileBytesReceived,
			prometheus.CounterValue,
			data[dbReplicaFileBytesReceivedPerSec].FirstValue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaGroupCommits,
			prometheus.CounterValue,
			data[dbReplicaGroupCommitsPerSec].FirstValue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaGroupCommitTime,
			prometheus.GaugeValue,
			data[dbReplicaGroupCommitTime].FirstValue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogApplyPendingQueue,
			prometheus.GaugeValue,
			data[dbReplicaLogApplyPendingQueue].FirstValue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogApplyReadyQueue,
			prometheus.GaugeValue,
			data[dbReplicaLogApplyReadyQueue].FirstValue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogBytesCompressed,
			prometheus.CounterValue,
			data[dbReplicaLogBytesCompressedPerSec].FirstValue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogBytesDecompressed,
			prometheus.CounterValue,
			data[dbReplicaLogBytesDecompressedPerSec].FirstValue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogBytesReceived,
			prometheus.CounterValue,
			data[dbReplicaLogBytesReceivedPerSec].FirstValue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogCompressionCachehits,
			prometheus.CounterValue,
			data[dbReplicaLogCompressionCacheHitsPerSec].FirstValue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogCompressionCachemisses,
			prometheus.CounterValue,
			data[dbReplicaLogCompressionCacheMissesPerSec].FirstValue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogCompressions,
			prometheus.CounterValue,
			data[dbReplicaLogCompressionsPerSec].FirstValue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogDecompressions,
			prometheus.CounterValue,
			data[dbReplicaLogDecompressionsPerSec].FirstValue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogremainingforundo,
			prometheus.GaugeValue,
			data[dbReplicaLogRemainingForUndo].FirstValue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogSendQueue,
			prometheus.GaugeValue,
			data[dbReplicaLogSendQueue].FirstValue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaMirroredWritetransactions,
			prometheus.CounterValue,
			data[dbReplicaMirroredWriteTransactionsPerSec].FirstValue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaRecoveryQueue,
			prometheus.GaugeValue,
			data[dbReplicaRecoveryQueue].FirstValue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaRedoblocked,
			prometheus.CounterValue,
			data[dbReplicaRedoBlockedPerSec].FirstValue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaRedoBytesRemaining,
			prometheus.GaugeValue,
			data[dbReplicaRedoBytesRemaining].FirstValue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaRedoneBytes,
			prometheus.CounterValue,
			data[dbReplicaRedoneBytesPerSec].FirstValue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaRedones,
			prometheus.CounterValue,
			data[dbReplicaRedonesPerSec].FirstValue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaTotalLogrequiringundo,
			prometheus.GaugeValue,
			data[dbReplicaTotalLogRequiringUndo].FirstValue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaTransactionDelay,
			prometheus.GaugeValue,
			data[dbReplicaTransactionDelay].FirstValue/1000.0,
			sqlInstance, replicaName,
		)
	}

	return nil
}

func (c *Collector) closeDatabaseReplica() {
	for _, collector := range c.dbReplicaPerfDataCollectors {
		collector.Close()
	}
}

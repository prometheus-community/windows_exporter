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

type collectorDatabaseReplica struct {
	dbReplicaPerfDataCollectors map[mssqlInstance]*pdh.Collector
	dbReplicaPerfDataObject     []perfDataCounterValuesDBReplica

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

type perfDataCounterValuesDBReplica struct {
	Name string

	DbReplicaDatabaseFlowControlDelay        float64 `perfdata:"Database Flow Control Delay"`
	DbReplicaDatabaseFlowControlsPerSec      float64 `perfdata:"Database Flow Controls/sec"`
	DbReplicaFileBytesReceivedPerSec         float64 `perfdata:"File Bytes Received/sec"`
	DbReplicaGroupCommitsPerSec              float64 `perfdata:"Group Commits/Sec"`
	DbReplicaGroupCommitTime                 float64 `perfdata:"Group Commit Time"`
	DbReplicaLogApplyPendingQueue            float64 `perfdata:"Log Apply Pending Queue"`
	DbReplicaLogApplyReadyQueue              float64 `perfdata:"Log Apply Ready Queue"`
	DbReplicaLogBytesCompressedPerSec        float64 `perfdata:"Log Bytes Compressed/sec"`
	DbReplicaLogBytesDecompressedPerSec      float64 `perfdata:"Log Bytes Decompressed/sec"`
	DbReplicaLogBytesReceivedPerSec          float64 `perfdata:"Log Bytes Received/sec"`
	DbReplicaLogCompressionCacheHitsPerSec   float64 `perfdata:"Log Compression Cache hits/sec"`
	DbReplicaLogCompressionCacheMissesPerSec float64 `perfdata:"Log Compression Cache misses/sec"`
	DbReplicaLogCompressionsPerSec           float64 `perfdata:"Log Compressions/sec"`
	DbReplicaLogDecompressionsPerSec         float64 `perfdata:"Log Decompressions/sec"`
	DbReplicaLogRemainingForUndo             float64 `perfdata:"Log remaining for undo"`
	DbReplicaLogSendQueue                    float64 `perfdata:"Log Send Queue"`
	DbReplicaMirroredWriteTransactionsPerSec float64 `perfdata:"Mirrored Write Transactions/sec"`
	DbReplicaRecoveryQueue                   float64 `perfdata:"Recovery Queue"`
	DbReplicaRedoBlockedPerSec               float64 `perfdata:"Redo blocked/sec"`
	DbReplicaRedoBytesRemaining              float64 `perfdata:"Redo Bytes Remaining"`
	DbReplicaRedoneBytesPerSec               float64 `perfdata:"Redone Bytes/sec"`
	DbReplicaRedonesPerSec                   float64 `perfdata:"Redones/sec"`
	DbReplicaTotalLogRequiringUndo           float64 `perfdata:"Total Log requiring undo"`
	DbReplicaTransactionDelay                float64 `perfdata:"Transaction Delay"`
}

func (c *Collector) buildDatabaseReplica() error {
	var err error

	c.dbReplicaPerfDataCollectors = make(map[mssqlInstance]*pdh.Collector, len(c.mssqlInstances))
	errs := make([]error, 0, len(c.mssqlInstances))

	for _, sqlInstance := range c.mssqlInstances {
		c.dbReplicaPerfDataCollectors[sqlInstance], err = pdh.NewCollector[perfDataCounterValuesDBReplica](pdh.CounterTypeRaw, c.mssqlGetPerfObjectName(sqlInstance, "Database Replica"), pdh.InstancesAll)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to create Database Replica collector for instance %s: %w", sqlInstance.name, err))
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

	return errors.Join(errs...)
}

func (c *Collector) collectDatabaseReplica(ch chan<- prometheus.Metric) error {
	return c.collect(ch, subCollectorDatabaseReplica, c.dbReplicaPerfDataCollectors, c.collectDatabaseReplicaInstance)
}

func (c *Collector) collectDatabaseReplicaInstance(ch chan<- prometheus.Metric, sqlInstance mssqlInstance, perfDataCollector *pdh.Collector) error {
	err := perfDataCollector.Collect(&c.dbReplicaPerfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect %s metrics: %w", c.mssqlGetPerfObjectName(sqlInstance, "Database Replica"), err)
	}

	for _, data := range c.dbReplicaPerfDataObject {
		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaDatabaseFlowControlDelay,
			prometheus.GaugeValue,
			data.DbReplicaDatabaseFlowControlDelay,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaDatabaseFlowControls,
			prometheus.CounterValue,
			data.DbReplicaDatabaseFlowControlsPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaFileBytesReceived,
			prometheus.CounterValue,
			data.DbReplicaFileBytesReceivedPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaGroupCommits,
			prometheus.CounterValue,
			data.DbReplicaGroupCommitsPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaGroupCommitTime,
			prometheus.GaugeValue,
			data.DbReplicaGroupCommitTime,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogApplyPendingQueue,
			prometheus.GaugeValue,
			data.DbReplicaLogApplyPendingQueue,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogApplyReadyQueue,
			prometheus.GaugeValue,
			data.DbReplicaLogApplyReadyQueue,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogBytesCompressed,
			prometheus.CounterValue,
			data.DbReplicaLogBytesCompressedPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogBytesDecompressed,
			prometheus.CounterValue,
			data.DbReplicaLogBytesDecompressedPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogBytesReceived,
			prometheus.CounterValue,
			data.DbReplicaLogBytesReceivedPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogCompressionCachehits,
			prometheus.CounterValue,
			data.DbReplicaLogCompressionCacheHitsPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogCompressionCachemisses,
			prometheus.CounterValue,
			data.DbReplicaLogCompressionCacheMissesPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogCompressions,
			prometheus.CounterValue,
			data.DbReplicaLogCompressionsPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogDecompressions,
			prometheus.CounterValue,
			data.DbReplicaLogDecompressionsPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogremainingforundo,
			prometheus.GaugeValue,
			data.DbReplicaLogRemainingForUndo,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogSendQueue,
			prometheus.GaugeValue,
			data.DbReplicaLogSendQueue,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaMirroredWritetransactions,
			prometheus.CounterValue,
			data.DbReplicaMirroredWriteTransactionsPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaRecoveryQueue,
			prometheus.GaugeValue,
			data.DbReplicaRecoveryQueue,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaRedoblocked,
			prometheus.CounterValue,
			data.DbReplicaRedoBlockedPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaRedoBytesRemaining,
			prometheus.GaugeValue,
			data.DbReplicaRedoBytesRemaining,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaRedoneBytes,
			prometheus.CounterValue,
			data.DbReplicaRedoneBytesPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaRedones,
			prometheus.CounterValue,
			data.DbReplicaRedonesPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaTotalLogrequiringundo,
			prometheus.GaugeValue,
			data.DbReplicaTotalLogRequiringUndo,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaTransactionDelay,
			prometheus.GaugeValue,
			data.DbReplicaTransactionDelay/1000.0,
			sqlInstance.name, data.Name,
		)
	}

	return nil
}

func (c *Collector) closeDatabaseReplica() {
	for _, collector := range c.dbReplicaPerfDataCollectors {
		collector.Close()
	}
}

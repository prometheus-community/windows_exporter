// returns data points from the following classes:
// - Win32_PerfRawData_MSSQLSERVER_SQLServerDatabaseReplica
//   https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-database-replica

package collector

import (
	"fmt"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	Factories["mssql_dbreplica"] = NewMSSQLDBReplicaCollector
}

// MSSQLDBReplicaCollector is a Prometheus collector for Win32_PerfRawData_{instance}_SQLServerDatabaseReplica metrics
type MSSQLDBReplicaCollector struct {
	DatabaseFlowControlDelay        *prometheus.Desc
	DatabaseFlowControlsPersec      *prometheus.Desc
	FileBytesReceivedPersec         *prometheus.Desc
	GroupCommitsPerSec              *prometheus.Desc
	GroupCommitTime                 *prometheus.Desc
	LogApplyPendingQueue            *prometheus.Desc
	LogApplyReadyQueue              *prometheus.Desc
	LogBytesCompressedPersec        *prometheus.Desc
	LogBytesDecompressedPersec      *prometheus.Desc
	LogBytesReceivedPersec          *prometheus.Desc
	LogCompressionCachehitsPersec   *prometheus.Desc
	LogCompressionCachemissesPersec *prometheus.Desc
	LogCompressionsPersec           *prometheus.Desc
	LogDecompressionsPersec         *prometheus.Desc
	Logremainingforundo             *prometheus.Desc
	LogSendQueue                    *prometheus.Desc
	MirroredWriteTransactionsPersec *prometheus.Desc
	RecoveryQueue                   *prometheus.Desc
	RedoblockedPersec               *prometheus.Desc
	RedoBytesRemaining              *prometheus.Desc
	RedoneBytesPersec               *prometheus.Desc
	RedonesPersec                   *prometheus.Desc
	TotalLogrequiringundo           *prometheus.Desc
	TransactionDelay                *prometheus.Desc

	sqlInstances sqlInstancesType
}

// NewMSSQLDBReplicaCollector ...
func NewMSSQLDBReplicaCollector() (Collector, error) {

	const subsystem = "mssql_dbreplica"
	return &MSSQLDBReplicaCollector{

		// Win32_PerfRawData_{instance}_SQLServerDatabaseReplica
		DatabaseFlowControlDelay: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "database_flow_control_wait_seconds"),
			"(DatabaseReplica.DatabaseFlowControlDelay)",
			[]string{"instance", "replica"},
			nil,
		),
		DatabaseFlowControlsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "database_initiated_flow_controls"),
			"(DatabaseReplica.DatabaseFlowControls)",
			[]string{"instance", "replica"},
			nil,
		),
		FileBytesReceivedPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "received_file_bytes"),
			"(DatabaseReplica.FileBytesReceived)",
			[]string{"instance", "replica"},
			nil,
		),
		GroupCommitsPerSec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "group_commits"),
			"(DatabaseReplica.GroupCommits)",
			[]string{"instance", "replica"},
			nil,
		),
		GroupCommitTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "group_commit_stall_seconds"),
			"(DatabaseReplica.GroupCommitTime)",
			[]string{"instance", "replica"},
			nil,
		),
		LogApplyPendingQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_apply_pending_queue"),
			"(DatabaseReplica.LogApplyPendingQueue)",
			[]string{"instance", "replica"},
			nil,
		),
		LogApplyReadyQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_apply_ready_queue"),
			"(DatabaseReplica.LogApplyReadyQueue)",
			[]string{"instance", "replica"},
			nil,
		),
		LogBytesCompressedPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_compressed_bytes"),
			"(DatabaseReplica.LogBytesCompressed)",
			[]string{"instance", "replica"},
			nil,
		),
		LogBytesDecompressedPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_decompressed_bytes"),
			"(DatabaseReplica.LogBytesDecompressed)",
			[]string{"instance", "replica"},
			nil,
		),
		LogBytesReceivedPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_received_bytes"),
			"(DatabaseReplica.LogBytesReceived)",
			[]string{"instance", "replica"},
			nil,
		),
		LogCompressionCachehitsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_compression_cachehits"),
			"(DatabaseReplica.LogCompressionCachehits)",
			[]string{"instance", "replica"},
			nil,
		),
		LogCompressionCachemissesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_compression_cachemisses"),
			"(DatabaseReplica.LogCompressionCachemisses)",
			[]string{"instance", "replica"},
			nil,
		),
		LogCompressionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_compressions"),
			"(DatabaseReplica.LogCompressions)",
			[]string{"instance", "replica"},
			nil,
		),
		LogDecompressionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_decompressions"),
			"(DatabaseReplica.LogDecompressions)",
			[]string{"instance", "replica"},
			nil,
		),
		Logremainingforundo: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_remaining_for_undo"),
			"(DatabaseReplica.Logremainingforundo)",
			[]string{"instance", "replica"},
			nil,
		),
		LogSendQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_send_queue"),
			"(DatabaseReplica.LogSendQueue)",
			[]string{"instance", "replica"},
			nil,
		),
		MirroredWriteTransactionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "mirrored_write_transactions"),
			"(DatabaseReplica.MirroredWriteTransactions)",
			[]string{"instance", "replica"},
			nil,
		),
		RecoveryQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "recovery_queue_records"),
			"(DatabaseReplica.RecoveryQueue)",
			[]string{"instance", "replica"},
			nil,
		),
		RedoblockedPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "redo_blocks"),
			"(DatabaseReplica.Redoblocked)",
			[]string{"instance", "replica"},
			nil,
		),
		RedoBytesRemaining: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "redo_remaining_bytes"),
			"(DatabaseReplica.RedoBytesRemaining)",
			[]string{"instance", "replica"},
			nil,
		),
		RedoneBytesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "redone_bytes"),
			"(DatabaseReplica.RedoneBytes)",
			[]string{"instance", "replica"},
			nil,
		),
		RedonesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "redones"),
			"(DatabaseReplica.Redones)",
			[]string{"instance", "replica"},
			nil,
		),
		TotalLogrequiringundo: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "total_log_requiring_undo"),
			"(DatabaseReplica.TotalLogrequiringundo)",
			[]string{"instance", "replica"},
			nil,
		),
		TransactionDelay: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "transaction_delay_seconds"),
			"(DatabaseReplica.TransactionDelay)",
			[]string{"instance", "replica"},
			nil,
		),

		sqlInstances: getMSSQLInstances(),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *MSSQLDBReplicaCollector) Collect(ch chan<- prometheus.Metric) error {
	for instance := range c.sqlInstances {
		log.Debugf("mssql_dbreplica collector iterating sql instance %s.", instance)

		if desc, err := c.collectDatabaseReplica(ch, instance); err != nil {
			log.Error("failed collecting MSSQL DatabaseReplica metrics:", desc, err)
			return err
		}
	}

	return nil
}

type win32PerfRawDataSQLServerDatabaseReplica struct {
	Name                            string
	DatabaseFlowControlDelay        uint64
	DatabaseFlowControlsPersec      uint64
	FileBytesReceivedPersec         uint64
	GroupCommitsPerSec              uint64
	GroupCommitTime                 uint64
	LogApplyPendingQueue            uint64
	LogApplyReadyQueue              uint64
	LogBytesCompressedPersec        uint64
	LogBytesDecompressedPersec      uint64
	LogBytesReceivedPersec          uint64
	LogCompressionCachehitsPersec   uint64
	LogCompressionCachemissesPersec uint64
	LogCompressionsPersec           uint64
	LogDecompressionsPersec         uint64
	Logremainingforundo             uint64
	LogSendQueue                    uint64
	MirroredWriteTransactionsPersec uint64
	RecoveryQueue                   uint64
	RedoblockedPersec               uint64
	RedoBytesRemaining              uint64
	RedoneBytesPersec               uint64
	RedonesPersec                   uint64
	TotalLogrequiringundo           uint64
	TransactionDelay                uint64
}

func (c *MSSQLDBReplicaCollector) collectDatabaseReplica(ch chan<- prometheus.Metric, sqlInstance string) (*prometheus.Desc, error) {
	var dst []win32PerfRawDataSQLServerDatabaseReplica
	class := fmt.Sprintf("Win32_PerfRawData_%s_SQLServerDatabaseReplica", sqlInstance)
	q := queryAllForClassWhere(&dst, class, `Name <> '_Total'`)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	for _, v := range dst {
		replicaName := v.Name

		ch <- prometheus.MustNewConstMetric(
			c.DatabaseFlowControlDelay,
			prometheus.GaugeValue,
			float64(v.DatabaseFlowControlDelay),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabaseFlowControlsPersec,
			prometheus.CounterValue,
			float64(v.DatabaseFlowControlsPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FileBytesReceivedPersec,
			prometheus.CounterValue,
			float64(v.FileBytesReceivedPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GroupCommitsPerSec,
			prometheus.CounterValue,
			float64(v.GroupCommitsPerSec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GroupCommitTime,
			prometheus.GaugeValue,
			float64(v.GroupCommitTime),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogApplyPendingQueue,
			prometheus.GaugeValue,
			float64(v.LogApplyPendingQueue),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogApplyReadyQueue,
			prometheus.GaugeValue,
			float64(v.LogApplyReadyQueue),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogBytesCompressedPersec,
			prometheus.CounterValue,
			float64(v.LogBytesCompressedPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogBytesDecompressedPersec,
			prometheus.CounterValue,
			float64(v.LogBytesDecompressedPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogBytesReceivedPersec,
			prometheus.CounterValue,
			float64(v.LogBytesReceivedPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogCompressionCachehitsPersec,
			prometheus.CounterValue,
			float64(v.LogCompressionCachehitsPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogCompressionCachemissesPersec,
			prometheus.CounterValue,
			float64(v.LogCompressionCachemissesPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogCompressionsPersec,
			prometheus.CounterValue,
			float64(v.LogCompressionsPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogDecompressionsPersec,
			prometheus.CounterValue,
			float64(v.LogDecompressionsPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Logremainingforundo,
			prometheus.GaugeValue,
			float64(v.Logremainingforundo),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogSendQueue,
			prometheus.GaugeValue,
			float64(v.LogSendQueue),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MirroredWriteTransactionsPersec,
			prometheus.CounterValue,
			float64(v.MirroredWriteTransactionsPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RecoveryQueue,
			prometheus.GaugeValue,
			float64(v.RecoveryQueue),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RedoblockedPersec,
			prometheus.CounterValue,
			float64(v.RedoblockedPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RedoBytesRemaining,
			prometheus.GaugeValue,
			float64(v.RedoBytesRemaining),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RedoneBytesPersec,
			prometheus.CounterValue,
			float64(v.RedoneBytesPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RedonesPersec,
			prometheus.CounterValue,
			float64(v.RedonesPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalLogrequiringundo,
			prometheus.GaugeValue,
			float64(v.TotalLogrequiringundo),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TransactionDelay,
			prometheus.GaugeValue,
			float64(v.TransactionDelay)*1000.0,
			sqlInstance, replicaName,
		)
	}

	return nil, nil
}

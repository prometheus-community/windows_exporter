// returns data points from the following classes:
// - Win32_PerfRawData_MSSQLSERVER_SQLServerAvailabilityReplica
//   https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-availability-replica
// - Win32_PerfRawData_MSSQLSERVER_SQLServerBufferManager
//   https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-buffer-manager-object
// - Win32_PerfRawData_MSSQLSERVER_SQLServerDatabaseReplica
//   https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-database-replica
// - Win32_PerfRawData_MSSQLSERVER_SQLServerDatabases
//   https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-databases-object?view=sql-server-2017
// - Win32_PerfRawData_MSSQLSERVER_SQLServerGeneralStatistics
//   https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-general-statistics-object
// - Win32_PerfRawData_MSSQLSERVER_SQLServerLocks
//   https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-locks-object
// - Win32_PerfRawData_MSSQLSERVER_SQLServerMemoryManager
//   https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-memory-manager-object
// - Win32_PerfRawData_MSSQLSERVER_SQLServerSQLStatistics
//   https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-sql-statistics-object

package collector

import (
	"fmt"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"golang.org/x/sys/windows/registry"
)

type sqlInstancesType map[string]string

var sqlInstances sqlInstancesType

func getMSSQLInstances() sqlInstancesType {
	sqlInstances := make(sqlInstancesType)

	// in case querying the registry fails, initialize list to the default instance
	sqlInstances["MSSQLSERVER"] = ""

	regkey := `Software\Microsoft\Microsoft SQL Server\Instance Names\SQL`
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, regkey, registry.QUERY_VALUE)
	if err != nil {
		log.Warn("Couldn't open registry to determine SQL instances:", err)
		return sqlInstances
	}
	defer k.Close()

	params, err := k.ReadValueNames(0)
	if err != nil {
		log.Warn("Can't ReadSubKeyNames %#v", err)
		return sqlInstances
	}

	for _, param := range params {
		if val, _, err := k.GetStringValue(param); err == nil {
			sqlInstances[param] = val
		}
	}

	log.Debugf("Detected MSSql Instnaces: %#v\n", sqlInstances)

	return sqlInstances
}

func init() {
	Factories["mssql"] = NewMSSQLCollector
}

// A MSSQLCollector is a Prometheus collector for various WMI Win32_PerfRawData_MSSQLSERVER_* metrics
type MSSQLCollector struct {
	// Win32_PerfRawData_{instance}_SQLServerAvailabilityReplica
	BytesReceivedfromReplicaPersec *prometheus.Desc
	BytesSenttoReplicaPersec       *prometheus.Desc
	BytesSenttoTransportPersec     *prometheus.Desc
	FlowControlPersec              *prometheus.Desc
	FlowControlTimemsPersec        *prometheus.Desc
	ReceivesfromReplicaPersec      *prometheus.Desc
	ResentMessagesPersec           *prometheus.Desc
	SendstoReplicaPersec           *prometheus.Desc
	SendstoTransportPersec         *prometheus.Desc

	// Win32_PerfRawData_{instance}_SQLServerBufferManager
	BackgroundwriterpagesPersec   *prometheus.Desc
	Buffercachehitratio           *prometheus.Desc
	CheckpointpagesPersec         *prometheus.Desc
	Databasepages                 *prometheus.Desc
	Extensionallocatedpages       *prometheus.Desc
	Extensionfreepages            *prometheus.Desc
	Extensioninuseaspercentage    *prometheus.Desc
	ExtensionoutstandingIOcounter *prometheus.Desc
	ExtensionpageevictionsPersec  *prometheus.Desc
	ExtensionpagereadsPersec      *prometheus.Desc
	Extensionpageunreferencedtime *prometheus.Desc
	ExtensionpagewritesPersec     *prometheus.Desc
	FreeliststallsPersec          *prometheus.Desc
	IntegralControllerSlope       *prometheus.Desc
	LazywritesPersec              *prometheus.Desc
	Pagelifeexpectancy            *prometheus.Desc
	PagelookupsPersec             *prometheus.Desc
	PagereadsPersec               *prometheus.Desc
	PagewritesPersec              *prometheus.Desc
	ReadaheadpagesPersec          *prometheus.Desc
	ReadaheadtimePersec           *prometheus.Desc
	Targetpages                   *prometheus.Desc

	// Win32_PerfRawData_{instance}_SQLServerDatabaseReplica
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

	// Win32_PerfRawData_{instance}_SQLServerDatabases
	ActiveTransactions               *prometheus.Desc
	AvgDistFromEOLLPRequest          *prometheus.Desc
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

	// Win32_PerfRawData_{instance}_SQLServerGeneralStatistics
	ActiveTempTables              *prometheus.Desc
	ConnectionResetPersec         *prometheus.Desc
	EventNotificationsDelayedDrop *prometheus.Desc
	HTTPAuthenticatedRequests     *prometheus.Desc
	LogicalConnections            *prometheus.Desc
	LoginsPersec                  *prometheus.Desc
	LogoutsPersec                 *prometheus.Desc
	MarsDeadlocks                 *prometheus.Desc
	Nonatomicyieldrate            *prometheus.Desc
	Processesblocked              *prometheus.Desc
	SOAPEmptyRequests             *prometheus.Desc
	SOAPMethodInvocations         *prometheus.Desc
	SOAPSessionInitiateRequests   *prometheus.Desc
	SOAPSessionTerminateRequests  *prometheus.Desc
	SOAPSQLRequests               *prometheus.Desc
	SOAPWSDLRequests              *prometheus.Desc
	SQLTraceIOProviderLockWaits   *prometheus.Desc
	Tempdbrecoveryunitid          *prometheus.Desc
	Tempdbrowsetid                *prometheus.Desc
	TempTablesCreationRate        *prometheus.Desc
	TempTablesForDestruction      *prometheus.Desc
	TraceEventNotificationQueue   *prometheus.Desc
	Transactions                  *prometheus.Desc
	UserConnections               *prometheus.Desc

	// Win32_PerfRawData_{instance}_SQLServerLocks
	AverageWaitTimems          *prometheus.Desc
	LockRequestsPersec         *prometheus.Desc
	LockTimeoutsPersec         *prometheus.Desc
	LockTimeoutstimeout0Persec *prometheus.Desc
	LockWaitsPersec            *prometheus.Desc
	LockWaitTimems             *prometheus.Desc
	NumberofDeadlocksPersec    *prometheus.Desc

	// Win32_PerfRawData_{instance}_SQLServerMemoryManager
	ConnectionMemoryKB       *prometheus.Desc
	DatabaseCacheMemoryKB    *prometheus.Desc
	Externalbenefitofmemory  *prometheus.Desc
	FreeMemoryKB             *prometheus.Desc
	GrantedWorkspaceMemoryKB *prometheus.Desc
	LockBlocks               *prometheus.Desc
	LockBlocksAllocated      *prometheus.Desc
	LockMemoryKB             *prometheus.Desc
	LockOwnerBlocks          *prometheus.Desc
	LockOwnerBlocksAllocated *prometheus.Desc
	LogPoolMemoryKB          *prometheus.Desc
	MaximumWorkspaceMemoryKB *prometheus.Desc
	MemoryGrantsOutstanding  *prometheus.Desc
	MemoryGrantsPending      *prometheus.Desc
	OptimizerMemoryKB        *prometheus.Desc
	ReservedServerMemoryKB   *prometheus.Desc
	SQLCacheMemoryKB         *prometheus.Desc
	StolenServerMemoryKB     *prometheus.Desc
	TargetServerMemoryKB     *prometheus.Desc
	TotalServerMemoryKB      *prometheus.Desc

	// Win32_PerfRawData_{instance}_SQLServerSQLStatistics
	AutoParamAttemptsPersec       *prometheus.Desc
	BatchRequestsPersec           *prometheus.Desc
	FailedAutoParamsPersec        *prometheus.Desc
	ForcedParameterizationsPersec *prometheus.Desc
	GuidedplanexecutionsPersec    *prometheus.Desc
	MisguidedplanexecutionsPersec *prometheus.Desc
	SafeAutoParamsPersec          *prometheus.Desc
	SQLAttentionrate              *prometheus.Desc
	SQLCompilationsPersec         *prometheus.Desc
	SQLReCompilationsPersec       *prometheus.Desc
	UnsafeAutoParamsPersec        *prometheus.Desc
}

// NewMSSQLCollector ...
func NewMSSQLCollector() (Collector, error) {
	sqlInstances = getMSSQLInstances()

	const subsystem = "mssql"
	return &MSSQLCollector{
		// Win32_PerfRawData_{instance}_SQLServerAvailabilityReplica
		BytesReceivedfromReplicaPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bytes_receivedfrom_replica_persec"),
			"(BytesReceivedfromReplicaPersec)",
			[]string{"instance"},
			nil,
		),
		BytesSenttoReplicaPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bytes_sentto_replica_persec"),
			"(BytesSenttoReplicaPersec)",
			[]string{"instance"},
			nil,
		),
		BytesSenttoTransportPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bytes_sentto_transport_persec"),
			"(BytesSenttoTransportPersec)",
			[]string{"instance"},
			nil,
		),
		FlowControlPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "flow_control_persec"),
			"(FlowControlPersec)",
			[]string{"instance"},
			nil,
		),
		FlowControlTimemsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "flow_control_timems_persec"),
			"(FlowControlTimemsPersec)",
			[]string{"instance"},
			nil,
		),
		ReceivesfromReplicaPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "receivesfrom_replica_persec"),
			"(ReceivesfromReplicaPersec)",
			[]string{"instance"},
			nil,
		),
		ResentMessagesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "resent_messages_persec"),
			"(ResentMessagesPersec)",
			[]string{"instance"},
			nil,
		),
		SendstoReplicaPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sendsto_replica_persec"),
			"(SendstoReplicaPersec)",
			[]string{"instance"},
			nil,
		),
		SendstoTransportPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sendsto_transport_persec"),
			"(SendstoTransportPersec)",
			[]string{"instance"},
			nil,
		),

		// Win32_PerfRawData_{instance}_SQLServerBufferManager
		BackgroundwriterpagesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "backgroundwriterpages_persec"),
			"(BackgroundwriterpagesPersec)",
			[]string{"instance"},
			nil,
		),
		Buffercachehitratio: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "buffercachehitratio"),
			"(Buffercachehitratio)",
			[]string{"instance"},
			nil,
		),
		CheckpointpagesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "checkpointpages_persec"),
			"(CheckpointpagesPersec)",
			[]string{"instance"},
			nil,
		),
		Databasepages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databasepages"),
			"(Databasepages)",
			[]string{"instance"},
			nil,
		),
		Extensionallocatedpages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "extensionallocatedpages"),
			"(Extensionallocatedpages)",
			[]string{"instance"},
			nil,
		),
		Extensionfreepages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "extensionfreepages"),
			"(Extensionfreepages)",
			[]string{"instance"},
			nil,
		),
		Extensioninuseaspercentage: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "extensioninuseaspercentage"),
			"(Extensioninuseaspercentage)",
			[]string{"instance"},
			nil,
		),
		ExtensionoutstandingIOcounter: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "extensionoutstanding_i_ocounter"),
			"(ExtensionoutstandingIOcounter)",
			[]string{"instance"},
			nil,
		),
		ExtensionpageevictionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "extensionpageevictions_persec"),
			"(ExtensionpageevictionsPersec)",
			[]string{"instance"},
			nil,
		),
		ExtensionpagereadsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "extensionpagereads_persec"),
			"(ExtensionpagereadsPersec)",
			[]string{"instance"},
			nil,
		),
		Extensionpageunreferencedtime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "extensionpageunreferencedtime"),
			"(Extensionpageunreferencedtime)",
			[]string{"instance"},
			nil,
		),
		ExtensionpagewritesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "extensionpagewrites_persec"),
			"(ExtensionpagewritesPersec)",
			[]string{"instance"},
			nil,
		),
		FreeliststallsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "freeliststalls_persec"),
			"(FreeliststallsPersec)",
			[]string{"instance"},
			nil,
		),
		IntegralControllerSlope: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "integral_controller_slope"),
			"(IntegralControllerSlope)",
			[]string{"instance"},
			nil,
		),
		LazywritesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lazywrites_persec"),
			"(LazywritesPersec)",
			[]string{"instance"},
			nil,
		),
		Pagelifeexpectancy: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pagelifeexpectancy"),
			"(Pagelifeexpectancy)",
			[]string{"instance"},
			nil,
		),
		PagelookupsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pagelookups_persec"),
			"(PagelookupsPersec)",
			[]string{"instance"},
			nil,
		),
		PagereadsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pagereads_persec"),
			"(PagereadsPersec)",
			[]string{"instance"},
			nil,
		),
		PagewritesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pagewrites_persec"),
			"(PagewritesPersec)",
			[]string{"instance"},
			nil,
		),
		ReadaheadpagesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "readaheadpages_persec"),
			"(ReadaheadpagesPersec)",
			[]string{"instance"},
			nil,
		),
		ReadaheadtimePersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "readaheadtime_persec"),
			"(ReadaheadtimePersec)",
			[]string{"instance"},
			nil,
		),
		Targetpages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "targetpages"),
			"(Targetpages)",
			[]string{"instance"},
			nil,
		),

		// Win32_PerfRawData_{instance}_SQLServerDatabaseReplica
		DatabaseFlowControlDelay: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "database_flow_control_delay"),
			"(DatabaseFlowControlDelay)",
			[]string{"instance"},
			nil,
		),
		DatabaseFlowControlsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "database_flow_controls_persec"),
			"(DatabaseFlowControlsPersec)",
			[]string{"instance"},
			nil,
		),
		FileBytesReceivedPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "file_bytes_received_persec"),
			"(FileBytesReceivedPersec)",
			[]string{"instance"},
			nil,
		),
		GroupCommitsPerSec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "group_commits_per_sec"),
			"(GroupCommitsPerSec)",
			[]string{"instance"},
			nil,
		),
		GroupCommitTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "group_commit_time"),
			"(GroupCommitTime)",
			[]string{"instance"},
			nil,
		),
		LogApplyPendingQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_apply_pending_queue"),
			"(LogApplyPendingQueue)",
			[]string{"instance"},
			nil,
		),
		LogApplyReadyQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_apply_ready_queue"),
			"(LogApplyReadyQueue)",
			[]string{"instance"},
			nil,
		),
		LogBytesCompressedPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_bytes_compressed_persec"),
			"(LogBytesCompressedPersec)",
			[]string{"instance"},
			nil,
		),
		LogBytesDecompressedPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_bytes_decompressed_persec"),
			"(LogBytesDecompressedPersec)",
			[]string{"instance"},
			nil,
		),
		LogBytesReceivedPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_bytes_received_persec"),
			"(LogBytesReceivedPersec)",
			[]string{"instance"},
			nil,
		),
		LogCompressionCachehitsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_compression_cachehits_persec"),
			"(LogCompressionCachehitsPersec)",
			[]string{"instance"},
			nil,
		),
		LogCompressionCachemissesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_compression_cachemisses_persec"),
			"(LogCompressionCachemissesPersec)",
			[]string{"instance"},
			nil,
		),
		LogCompressionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_compressions_persec"),
			"(LogCompressionsPersec)",
			[]string{"instance"},
			nil,
		),
		LogDecompressionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_decompressions_persec"),
			"(LogDecompressionsPersec)",
			[]string{"instance"},
			nil,
		),
		Logremainingforundo: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "logremainingforundo"),
			"(Logremainingforundo)",
			[]string{"instance"},
			nil,
		),
		LogSendQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_send_queue"),
			"(LogSendQueue)",
			[]string{"instance"},
			nil,
		),
		MirroredWriteTransactionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "mirrored_write_transactions_persec"),
			"(MirroredWriteTransactionsPersec)",
			[]string{"instance"},
			nil,
		),
		RecoveryQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "recovery_queue"),
			"(RecoveryQueue)",
			[]string{"instance"},
			nil,
		),
		RedoblockedPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "redoblocked_persec"),
			"(RedoblockedPersec)",
			[]string{"instance"},
			nil,
		),
		RedoBytesRemaining: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "redo_bytes_remaining"),
			"(RedoBytesRemaining)",
			[]string{"instance"},
			nil,
		),
		RedoneBytesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "redone_bytes_persec"),
			"(RedoneBytesPersec)",
			[]string{"instance"},
			nil,
		),
		RedonesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "redones_persec"),
			"(RedonesPersec)",
			[]string{"instance"},
			nil,
		),
		TotalLogrequiringundo: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "total_logrequiringundo"),
			"(TotalLogrequiringundo)",
			[]string{"instance"},
			nil,
		),
		TransactionDelay: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "transaction_delay"),
			"(TransactionDelay)",
			[]string{"instance"},
			nil,
		),

		// Win32_PerfRawData_{instance}_SQLServerDatabases
		ActiveTransactions: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "active_transactions"),
			"(ActiveTransactions)",
			[]string{"instance"},
			nil,
		),
		AvgDistFromEOLLPRequest: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "avg_dist_from_eollp_request"),
			"(AvgDistFromEOLLPRequest)",
			[]string{"instance"},
			nil,
		),
		BackupPerRestoreThroughputPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "backup_per_restore_throughput_persec"),
			"(BackupPerRestoreThroughputPersec)",
			[]string{"instance"},
			nil,
		),
		BulkCopyRowsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bulk_copy_rows_persec"),
			"(BulkCopyRowsPersec)",
			[]string{"instance"},
			nil,
		),
		BulkCopyThroughputPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bulk_copy_throughput_persec"),
			"(BulkCopyThroughputPersec)",
			[]string{"instance"},
			nil,
		),
		Committableentries: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "committableentries"),
			"(Committableentries)",
			[]string{"instance"},
			nil,
		),
		DataFilesSizeKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "data_files_size_kb"),
			"(DataFilesSizeKB)",
			[]string{"instance"},
			nil,
		),
		DBCCLogicalScanBytesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbcc_logical_scan_bytes_persec"),
			"(DBCCLogicalScanBytesPersec)",
			[]string{"instance"},
			nil,
		),
		GroupCommitTimePersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "group_commit_time_persec"),
			"(GroupCommitTimePersec)",
			[]string{"instance"},
			nil,
		),
		LogBytesFlushedPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_bytes_flushed_persec"),
			"(LogBytesFlushedPersec)",
			[]string{"instance"},
			nil,
		),
		LogCacheHitRatio: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_cache_hit_ratio"),
			"(LogCacheHitRatio)",
			[]string{"instance"},
			nil,
		),
		LogCacheReadsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_cache_reads_persec"),
			"(LogCacheReadsPersec)",
			[]string{"instance"},
			nil,
		),
		LogFilesSizeKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_files_size_kb"),
			"(LogFilesSizeKB)",
			[]string{"instance"},
			nil,
		),
		LogFilesUsedSizeKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_files_used_size_kb"),
			"(LogFilesUsedSizeKB)",
			[]string{"instance"},
			nil,
		),
		LogFlushesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_flushes_persec"),
			"(LogFlushesPersec)",
			[]string{"instance"},
			nil,
		),
		LogFlushWaitsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_flush_waits_persec"),
			"(LogFlushWaitsPersec)",
			[]string{"instance"},
			nil,
		),
		LogFlushWaitTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_flush_wait_time"),
			"(LogFlushWaitTime)",
			[]string{"instance"},
			nil,
		),
		LogFlushWriteTimems: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_flush_write_timems"),
			"(LogFlushWriteTimems)",
			[]string{"instance"},
			nil,
		),
		LogGrowths: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_growths"),
			"(LogGrowths)",
			[]string{"instance"},
			nil,
		),
		LogPoolCacheMissesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_cache_misses_persec"),
			"(LogPoolCacheMissesPersec)",
			[]string{"instance"},
			nil,
		),
		LogPoolDiskReadsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_disk_reads_persec"),
			"(LogPoolDiskReadsPersec)",
			[]string{"instance"},
			nil,
		),
		LogPoolHashDeletesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_hash_deletes_persec"),
			"(LogPoolHashDeletesPersec)",
			[]string{"instance"},
			nil,
		),
		LogPoolHashInsertsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_hash_inserts_persec"),
			"(LogPoolHashInsertsPersec)",
			[]string{"instance"},
			nil,
		),
		LogPoolInvalidHashEntryPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_invalid_hash_entry_persec"),
			"(LogPoolInvalidHashEntryPersec)",
			[]string{"instance"},
			nil,
		),
		LogPoolLogScanPushesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_log_scan_pushes_persec"),
			"(LogPoolLogScanPushesPersec)",
			[]string{"instance"},
			nil,
		),
		LogPoolLogWriterPushesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_log_writer_pushes_persec"),
			"(LogPoolLogWriterPushesPersec)",
			[]string{"instance"},
			nil,
		),
		LogPoolPushEmptyFreePoolPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_push_empty_free_pool_persec"),
			"(LogPoolPushEmptyFreePoolPersec)",
			[]string{"instance"},
			nil,
		),
		LogPoolPushLowMemoryPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_push_low_memory_persec"),
			"(LogPoolPushLowMemoryPersec)",
			[]string{"instance"},
			nil,
		),
		LogPoolPushNoFreeBufferPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_push_no_free_buffer_persec"),
			"(LogPoolPushNoFreeBufferPersec)",
			[]string{"instance"},
			nil,
		),
		LogPoolReqBehindTruncPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_req_behind_trunc_persec"),
			"(LogPoolReqBehindTruncPersec)",
			[]string{"instance"},
			nil,
		),
		LogPoolRequestsOldVLFPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_requests_old_vlf_persec"),
			"(LogPoolRequestsOldVLFPersec)",
			[]string{"instance"},
			nil,
		),
		LogPoolRequestsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_requests_persec"),
			"(LogPoolRequestsPersec)",
			[]string{"instance"},
			nil,
		),
		LogPoolTotalActiveLogSize: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_total_active_log_size"),
			"(LogPoolTotalActiveLogSize)",
			[]string{"instance"},
			nil,
		),
		LogPoolTotalSharedPoolSize: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_total_shared_pool_size"),
			"(LogPoolTotalSharedPoolSize)",
			[]string{"instance"},
			nil,
		),
		LogShrinks: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_shrinks"),
			"(LogShrinks)",
			[]string{"instance"},
			nil,
		),
		LogTruncations: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_truncations"),
			"(LogTruncations)",
			[]string{"instance"},
			nil,
		),
		PercentLogUsed: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "percent_log_used"),
			"(PercentLogUsed)",
			[]string{"instance"},
			nil,
		),
		ReplPendingXacts: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "repl_pending_xacts"),
			"(ReplPendingXacts)",
			[]string{"instance"},
			nil,
		),
		ReplTransRate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "repl_trans_rate"),
			"(ReplTransRate)",
			[]string{"instance"},
			nil,
		),
		ShrinkDataMovementBytesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "shrink_data_movement_bytes_persec"),
			"(ShrinkDataMovementBytesPersec)",
			[]string{"instance"},
			nil,
		),
		TrackedtransactionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "trackedtransactions_persec"),
			"(TrackedtransactionsPersec)",
			[]string{"instance"},
			nil,
		),
		TransactionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "transactions_persec"),
			"(TransactionsPersec)",
			[]string{"instance"},
			nil,
		),
		WriteTransactionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "write_transactions_persec"),
			"(WriteTransactionsPersec)",
			[]string{"instance"},
			nil,
		),
		XTPControllerDLCLatencyPerFetch: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "xtp_controller_dlc_latency_per_fetch"),
			"(XTPControllerDLCLatencyPerFetch)",
			[]string{"instance"},
			nil,
		),
		XTPControllerDLCPeakLatency: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "xtp_controller_dlc_peak_latency"),
			"(XTPControllerDLCPeakLatency)",
			[]string{"instance"},
			nil,
		),
		XTPControllerLogProcessedPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "xtp_controller_log_processed_persec"),
			"(XTPControllerLogProcessedPersec)",
			[]string{"instance"},
			nil,
		),
		XTPMemoryUsedKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "xtp_memory_used_kb"),
			"(XTPMemoryUsedKB)",
			[]string{"instance"},
			nil,
		),

		// Win32_PerfRawData_{instance}_SQLServerGeneralStatistics
		ActiveTempTables: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "active_temp_tables"),
			"(ActiveTempTables)",
			[]string{"instance"},
			nil,
		),
		ConnectionResetPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "connection_reset_persec"),
			"(ConnectionResetPersec)",
			[]string{"instance"},
			nil,
		),
		EventNotificationsDelayedDrop: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "event_notifications_delayed_drop"),
			"(EventNotificationsDelayedDrop)",
			[]string{"instance"},
			nil,
		),
		HTTPAuthenticatedRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "http_authenticated_requests"),
			"(HTTPAuthenticatedRequests)",
			[]string{"instance"},
			nil,
		),
		LogicalConnections: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "logical_connections"),
			"(LogicalConnections)",
			[]string{"instance"},
			nil,
		),
		LoginsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "logins_persec"),
			"(LoginsPersec)",
			[]string{"instance"},
			nil,
		),
		LogoutsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "logouts_persec"),
			"(LogoutsPersec)",
			[]string{"instance"},
			nil,
		),
		MarsDeadlocks: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "mars_deadlocks"),
			"(MarsDeadlocks)",
			[]string{"instance"},
			nil,
		),
		Nonatomicyieldrate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "nonatomicyieldrate"),
			"(Nonatomicyieldrate)",
			[]string{"instance"},
			nil,
		),
		Processesblocked: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "processesblocked"),
			"(Processesblocked)",
			[]string{"instance"},
			nil,
		),
		SOAPEmptyRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "soap_empty_requests"),
			"(SOAPEmptyRequests)",
			[]string{"instance"},
			nil,
		),
		SOAPMethodInvocations: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "soap_method_invocations"),
			"(SOAPMethodInvocations)",
			[]string{"instance"},
			nil,
		),
		SOAPSessionInitiateRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "soap_session_initiate_requests"),
			"(SOAPSessionInitiateRequests)",
			[]string{"instance"},
			nil,
		),
		SOAPSessionTerminateRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "soap_session_terminate_requests"),
			"(SOAPSessionTerminateRequests)",
			[]string{"instance"},
			nil,
		),
		SOAPSQLRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "soapsql_requests"),
			"(SOAPSQLRequests)",
			[]string{"instance"},
			nil,
		),
		SOAPWSDLRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "soapwsdl_requests"),
			"(SOAPWSDLRequests)",
			[]string{"instance"},
			nil,
		),
		SQLTraceIOProviderLockWaits: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sql_trace_io_provider_lock_waits"),
			"(SQLTraceIOProviderLockWaits)",
			[]string{"instance"},
			nil,
		),
		Tempdbrecoveryunitid: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "tempdbrecoveryunitid"),
			"(Tempdbrecoveryunitid)",
			[]string{"instance"},
			nil,
		),
		Tempdbrowsetid: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "tempdbrowsetid"),
			"(Tempdbrowsetid)",
			[]string{"instance"},
			nil,
		),
		TempTablesCreationRate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "temp_tables_creation_rate"),
			"(TempTablesCreationRate)",
			[]string{"instance"},
			nil,
		),
		TempTablesForDestruction: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "temp_tables_for_destruction"),
			"(TempTablesForDestruction)",
			[]string{"instance"},
			nil,
		),
		TraceEventNotificationQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "trace_event_notification_queue"),
			"(TraceEventNotificationQueue)",
			[]string{"instance"},
			nil,
		),
		Transactions: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "transactions"),
			"(Transactions)",
			[]string{"instance"},
			nil,
		),
		UserConnections: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "user_connections"),
			"(UserConnections)",
			[]string{"instance"},
			nil,
		),

		// Win32_PerfRawData_{instance}_SQLServerLocks
		AverageWaitTimems: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "average_wait_timems"),
			"(AverageWaitTimems)",
			[]string{"instance"},
			nil,
		),
		LockRequestsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_requests_persec"),
			"(LockRequestsPersec)",
			[]string{"instance"},
			nil,
		),
		LockTimeoutsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_timeouts_persec"),
			"(LockTimeoutsPersec)",
			[]string{"instance"},
			nil,
		),
		LockTimeoutstimeout0Persec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_timeoutstimeout0_persec"),
			"(LockTimeoutstimeout0Persec)",
			[]string{"instance"},
			nil,
		),
		LockWaitsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_waits_persec"),
			"(LockWaitsPersec)",
			[]string{"instance"},
			nil,
		),
		LockWaitTimems: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_wait_timems"),
			"(LockWaitTimems)",
			[]string{"instance"},
			nil,
		),
		NumberofDeadlocksPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "numberof_deadlocks_persec"),
			"(NumberofDeadlocksPersec)",
			[]string{"instance"},
			nil,
		),

		// Win32_PerfRawData_{instance}_SQLServerMemoryManager
		ConnectionMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "connection_memory_kb"),
			"(ConnectionMemoryKB)",
			[]string{"instance"},
			nil,
		),
		DatabaseCacheMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "database_cache_memory_kb"),
			"(DatabaseCacheMemoryKB)",
			[]string{"instance"},
			nil,
		),
		Externalbenefitofmemory: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "externalbenefitofmemory"),
			"(Externalbenefitofmemory)",
			[]string{"instance"},
			nil,
		),
		FreeMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "free_memory_kb"),
			"(FreeMemoryKB)",
			[]string{"instance"},
			nil,
		),
		GrantedWorkspaceMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "granted_workspace_memory_kb"),
			"(GrantedWorkspaceMemoryKB)",
			[]string{"instance"},
			nil,
		),
		LockBlocks: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_blocks"),
			"(LockBlocks)",
			[]string{"instance"},
			nil,
		),
		LockBlocksAllocated: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_blocks_allocated"),
			"(LockBlocksAllocated)",
			[]string{"instance"},
			nil,
		),
		LockMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_memory_kb"),
			"(LockMemoryKB)",
			[]string{"instance"},
			nil,
		),
		LockOwnerBlocks: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_owner_blocks"),
			"(LockOwnerBlocks)",
			[]string{"instance"},
			nil,
		),
		LockOwnerBlocksAllocated: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_owner_blocks_allocated"),
			"(LockOwnerBlocksAllocated)",
			[]string{"instance"},
			nil,
		),
		LogPoolMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_memory_kb"),
			"(LogPoolMemoryKB)",
			[]string{"instance"},
			nil,
		),
		MaximumWorkspaceMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "maximum_workspace_memory_kb"),
			"(MaximumWorkspaceMemoryKB)",
			[]string{"instance"},
			nil,
		),
		MemoryGrantsOutstanding: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memory_grants_outstanding"),
			"(MemoryGrantsOutstanding)",
			[]string{"instance"},
			nil,
		),
		MemoryGrantsPending: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memory_grants_pending"),
			"(MemoryGrantsPending)",
			[]string{"instance"},
			nil,
		),
		OptimizerMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "optimizer_memory_kb"),
			"(OptimizerMemoryKB)",
			[]string{"instance"},
			nil,
		),
		ReservedServerMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "reserved_server_memory_kb"),
			"(ReservedServerMemoryKB)",
			[]string{"instance"},
			nil,
		),
		SQLCacheMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sql_cache_memory_kb"),
			"(SQLCacheMemoryKB)",
			[]string{"instance"},
			nil,
		),
		StolenServerMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "stolen_server_memory_kb"),
			"(StolenServerMemoryKB)",
			[]string{"instance"},
			nil,
		),
		TargetServerMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "target_server_memory_kb"),
			"(TargetServerMemoryKB)",
			[]string{"instance"},
			nil,
		),
		TotalServerMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "total_server_memory_kb"),
			"(TotalServerMemoryKB)",
			[]string{"instance"},
			nil,
		),

		// Win32_PerfRawData_{instance}_SQLServerSQLStatistics
		AutoParamAttemptsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "auto_param_attempts_persec"),
			"(AutoParamAttemptsPersec)",
			[]string{"instance"},
			nil,
		),
		BatchRequestsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "batch_requests_persec"),
			"(BatchRequestsPersec)",
			[]string{"instance"},
			nil,
		),
		FailedAutoParamsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "failed_auto_params_persec"),
			"(FailedAutoParamsPersec)",
			[]string{"instance"},
			nil,
		),
		ForcedParameterizationsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "forced_parameterizations_persec"),
			"(ForcedParameterizationsPersec)",
			[]string{"instance"},
			nil,
		),
		GuidedplanexecutionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "guidedplanexecutions_persec"),
			"(GuidedplanexecutionsPersec)",
			[]string{"instance"},
			nil,
		),
		MisguidedplanexecutionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "misguidedplanexecutions_persec"),
			"(MisguidedplanexecutionsPersec)",
			[]string{"instance"},
			nil,
		),
		SafeAutoParamsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "safe_auto_params_persec"),
			"(SafeAutoParamsPersec)",
			[]string{"instance"},
			nil,
		),
		SQLAttentionrate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sql_attentionrate"),
			"(SQLAttentionrate)",
			[]string{"instance"},
			nil,
		),
		SQLCompilationsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sql_compilations_persec"),
			"(SQLCompilationsPersec)",
			[]string{"instance"},
			nil,
		),
		SQLReCompilationsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sql_re_compilations_persec"),
			"(SQLReCompilationsPersec)",
			[]string{"instance"},
			nil,
		),
		UnsafeAutoParamsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "unsafe_auto_params_persec"),
			"(UnsafeAutoParamsPersec)",
			[]string{"instance"},
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *MSSQLCollector) Collect(ch chan<- prometheus.Metric) error {
	for instance := range sqlInstances {
		log.Debugf("mssql collector iterating sql instance %s.", instance)

		// Win32_PerfRawData_{instance}_SQLServerAvailabilityReplica
		if desc, err := c.collectAvailabilityReplica(ch, instance); err != nil {
			log.Error("failed collecting MSSQL GeneralStatistics metrics:", desc, err)
			return err
		}

		// Win32_PerfRawData_{instance}_SQLServerBufferManager
		if desc, err := c.collectBufferManager(ch, instance); err != nil {
			log.Error("failed collecting MSSQL BufferManager metrics:", desc, err)
			return err
		}

		// Win32_PerfRawData_{instance}_SQLServerDatabaseReplica
		if desc, err := c.collectDatabaseReplica(ch, instance); err != nil {
			log.Error("failed collecting MSSQL DatabaseReplica metrics:", desc, err)
			return err
		}

		// Win32_PerfRawData_{instance}_SQLServerDatabases
		if desc, err := c.collectDatabases(ch, instance); err != nil {
			log.Error("failed collecting MSSQL Databases metrics:", desc, err)
			return err
		}

		// Win32_PerfRawData_{instance}_SQLServerGeneralStatistics
		if desc, err := c.collectGeneralStatistics(ch, instance); err != nil {
			log.Error("failed collecting MSSQL GeneralStatistics metrics:", desc, instance, err)
			return err
		}

		// Win32_PerfRawData_{instance}_SQLServerLocks
		if desc, err := c.collectLocks(ch, instance); err != nil {
			log.Error("failed collecting MSSQL Locks metrics:", desc, err)
			return err
		}

		// Win32_PerfRawData_{instance}_SQLServerMemoryManager
		if desc, err := c.collectMemoryManager(ch, instance); err != nil {
			log.Error("failed collecting MSSQL MemoryManager metrics:", desc, err)
			return err
		}

		// Win32_PerfRawData_{instance}_SQLServerSQLStatistics
		if desc, err := c.collectSQLStats(ch, instance); err != nil {
			log.Error("failed collecting MSSQL SQLStats metrics:", desc, err)
			return err
		}
	}

	return nil
}

type win32PerfRawDataSQLServerAvailabilityReplica struct {
	BytesReceivedfromReplicaPersec uint64
	BytesSenttoReplicaPersec       uint64
	BytesSenttoTransportPersec     uint64
	FlowControlPersec              uint64
	FlowControlTimemsPersec        uint64
	ReceivesfromReplicaPersec      uint64
	ResentMessagesPersec           uint64
	SendstoReplicaPersec           uint64
	SendstoTransportPersec         uint64
}

func (c *MSSQLCollector) collectAvailabilityReplica(ch chan<- prometheus.Metric, sqlInstance string) (*prometheus.Desc, error) {
	var dst []win32PerfRawDataSQLServerAvailabilityReplica
	class := fmt.Sprintf("Win32_PerfRawData_%s_SQLServerAvailabilityReplica", sqlInstance)
	q := queryAllForClass(&dst, class)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	if len(dst) > 0 {
		ch <- prometheus.MustNewConstMetric(
			c.BytesReceivedfromReplicaPersec,
			prometheus.GaugeValue,
			float64(dst[0].BytesReceivedfromReplicaPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BytesSenttoReplicaPersec,
			prometheus.GaugeValue,
			float64(dst[0].BytesSenttoReplicaPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BytesSenttoTransportPersec,
			prometheus.GaugeValue,
			float64(dst[0].BytesSenttoTransportPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FlowControlPersec,
			prometheus.GaugeValue,
			float64(dst[0].FlowControlPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FlowControlTimemsPersec,
			prometheus.GaugeValue,
			float64(dst[0].FlowControlTimemsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReceivesfromReplicaPersec,
			prometheus.GaugeValue,
			float64(dst[0].ReceivesfromReplicaPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ResentMessagesPersec,
			prometheus.GaugeValue,
			float64(dst[0].ResentMessagesPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SendstoReplicaPersec,
			prometheus.GaugeValue,
			float64(dst[0].SendstoReplicaPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SendstoTransportPersec,
			prometheus.GaugeValue,
			float64(dst[0].SendstoTransportPersec),
			sqlInstance,
		)
	}

	return nil, nil
}

type win32PerfRawDataSQLServerBufferManager struct {
	BackgroundwriterpagesPersec   uint64
	Buffercachehitratio           uint64
	CheckpointpagesPersec         uint64
	Databasepages                 uint64
	Extensionallocatedpages       uint64
	Extensionfreepages            uint64
	Extensioninuseaspercentage    uint64
	ExtensionoutstandingIOcounter uint64
	ExtensionpageevictionsPersec  uint64
	ExtensionpagereadsPersec      uint64
	Extensionpageunreferencedtime uint64
	ExtensionpagewritesPersec     uint64
	FreeliststallsPersec          uint64
	IntegralControllerSlope       uint64
	LazywritesPersec              uint64
	Pagelifeexpectancy            uint64
	PagelookupsPersec             uint64
	PagereadsPersec               uint64
	PagewritesPersec              uint64
	ReadaheadpagesPersec          uint64
	ReadaheadtimePersec           uint64
	Targetpages                   uint64
}

func (c *MSSQLCollector) collectBufferManager(ch chan<- prometheus.Metric, sqlInstance string) (*prometheus.Desc, error) {
	var dst []win32PerfRawDataSQLServerBufferManager
	class := fmt.Sprintf("Win32_PerfRawData_%s_SQLServerBufferManager", sqlInstance)
	q := queryAllForClass(&dst, class)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	if len(dst) > 0 {
		ch <- prometheus.MustNewConstMetric(
			c.BackgroundwriterpagesPersec,
			prometheus.GaugeValue,
			float64(dst[0].BackgroundwriterpagesPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Buffercachehitratio,
			prometheus.GaugeValue,
			float64(dst[0].Buffercachehitratio),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.CheckpointpagesPersec,
			prometheus.GaugeValue,
			float64(dst[0].CheckpointpagesPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Databasepages,
			prometheus.GaugeValue,
			float64(dst[0].Databasepages),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Extensionallocatedpages,
			prometheus.GaugeValue,
			float64(dst[0].Extensionallocatedpages),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Extensionfreepages,
			prometheus.GaugeValue,
			float64(dst[0].Extensionfreepages),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Extensioninuseaspercentage,
			prometheus.GaugeValue,
			float64(dst[0].Extensioninuseaspercentage),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ExtensionoutstandingIOcounter,
			prometheus.GaugeValue,
			float64(dst[0].ExtensionoutstandingIOcounter),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ExtensionpageevictionsPersec,
			prometheus.GaugeValue,
			float64(dst[0].ExtensionpageevictionsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ExtensionpagereadsPersec,
			prometheus.GaugeValue,
			float64(dst[0].ExtensionpagereadsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Extensionpageunreferencedtime,
			prometheus.GaugeValue,
			float64(dst[0].Extensionpageunreferencedtime),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ExtensionpagewritesPersec,
			prometheus.GaugeValue,
			float64(dst[0].ExtensionpagewritesPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FreeliststallsPersec,
			prometheus.GaugeValue,
			float64(dst[0].FreeliststallsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.IntegralControllerSlope,
			prometheus.GaugeValue,
			float64(dst[0].IntegralControllerSlope),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LazywritesPersec,
			prometheus.GaugeValue,
			float64(dst[0].LazywritesPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Pagelifeexpectancy,
			prometheus.GaugeValue,
			float64(dst[0].Pagelifeexpectancy),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PagelookupsPersec,
			prometheus.GaugeValue,
			float64(dst[0].PagelookupsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PagereadsPersec,
			prometheus.GaugeValue,
			float64(dst[0].PagereadsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PagewritesPersec,
			prometheus.GaugeValue,
			float64(dst[0].PagewritesPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReadaheadpagesPersec,
			prometheus.GaugeValue,
			float64(dst[0].ReadaheadpagesPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReadaheadtimePersec,
			prometheus.GaugeValue,
			float64(dst[0].ReadaheadtimePersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Targetpages,
			prometheus.GaugeValue,
			float64(dst[0].Targetpages),
			sqlInstance,
		)
	}

	return nil, nil
}

type win32PerfRawDataSQLServerDatabaseReplica struct {
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

func (c *MSSQLCollector) collectDatabaseReplica(ch chan<- prometheus.Metric, sqlInstance string) (*prometheus.Desc, error) {
	var dst []win32PerfRawDataSQLServerDatabaseReplica
	class := fmt.Sprintf("Win32_PerfRawData_%s_SQLServerDatabaseReplica", sqlInstance)
	q := queryAllForClass(&dst, class)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	if len(dst) > 0 {
		ch <- prometheus.MustNewConstMetric(
			c.DatabaseFlowControlDelay,
			prometheus.GaugeValue,
			float64(dst[0].DatabaseFlowControlDelay),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabaseFlowControlsPersec,
			prometheus.GaugeValue,
			float64(dst[0].DatabaseFlowControlsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FileBytesReceivedPersec,
			prometheus.GaugeValue,
			float64(dst[0].FileBytesReceivedPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GroupCommitsPerSec,
			prometheus.GaugeValue,
			float64(dst[0].GroupCommitsPerSec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GroupCommitTime,
			prometheus.GaugeValue,
			float64(dst[0].GroupCommitTime),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogApplyPendingQueue,
			prometheus.GaugeValue,
			float64(dst[0].LogApplyPendingQueue),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogApplyReadyQueue,
			prometheus.GaugeValue,
			float64(dst[0].LogApplyReadyQueue),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogBytesCompressedPersec,
			prometheus.GaugeValue,
			float64(dst[0].LogBytesCompressedPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogBytesDecompressedPersec,
			prometheus.GaugeValue,
			float64(dst[0].LogBytesDecompressedPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogBytesReceivedPersec,
			prometheus.GaugeValue,
			float64(dst[0].LogBytesReceivedPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogCompressionCachehitsPersec,
			prometheus.GaugeValue,
			float64(dst[0].LogCompressionCachehitsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogCompressionCachemissesPersec,
			prometheus.GaugeValue,
			float64(dst[0].LogCompressionCachemissesPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogCompressionsPersec,
			prometheus.GaugeValue,
			float64(dst[0].LogCompressionsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogDecompressionsPersec,
			prometheus.GaugeValue,
			float64(dst[0].LogDecompressionsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Logremainingforundo,
			prometheus.GaugeValue,
			float64(dst[0].Logremainingforundo),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogSendQueue,
			prometheus.GaugeValue,
			float64(dst[0].LogSendQueue),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MirroredWriteTransactionsPersec,
			prometheus.GaugeValue,
			float64(dst[0].MirroredWriteTransactionsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RecoveryQueue,
			prometheus.GaugeValue,
			float64(dst[0].RecoveryQueue),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RedoblockedPersec,
			prometheus.GaugeValue,
			float64(dst[0].RedoblockedPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RedoBytesRemaining,
			prometheus.GaugeValue,
			float64(dst[0].RedoBytesRemaining),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RedoneBytesPersec,
			prometheus.GaugeValue,
			float64(dst[0].RedoneBytesPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RedonesPersec,
			prometheus.GaugeValue,
			float64(dst[0].RedonesPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalLogrequiringundo,
			prometheus.GaugeValue,
			float64(dst[0].TotalLogrequiringundo),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TransactionDelay,
			prometheus.GaugeValue,
			float64(dst[0].TransactionDelay),
			sqlInstance,
		)
	}

	return nil, nil
}

type win32PerfRawDataSQLServerDatabases struct {
	ActiveTransactions               uint64
	AvgDistFromEOLLPRequest          uint64
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

func (c *MSSQLCollector) collectDatabases(ch chan<- prometheus.Metric, sqlInstance string) (*prometheus.Desc, error) {
	var dst []win32PerfRawDataSQLServerDatabases
	class := fmt.Sprintf("Win32_PerfRawData_%s_SQLServerDatabases", sqlInstance)
	q := queryAllForClass(&dst, class)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	if len(dst) > 0 {
		ch <- prometheus.MustNewConstMetric(
			c.ActiveTransactions,
			prometheus.GaugeValue,
			float64(dst[0].ActiveTransactions),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AvgDistFromEOLLPRequest,
			prometheus.GaugeValue,
			float64(dst[0].AvgDistFromEOLLPRequest),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BackupPerRestoreThroughputPersec,
			prometheus.GaugeValue,
			float64(dst[0].BackupPerRestoreThroughputPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BulkCopyRowsPersec,
			prometheus.GaugeValue,
			float64(dst[0].BulkCopyRowsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BulkCopyThroughputPersec,
			prometheus.GaugeValue,
			float64(dst[0].BulkCopyThroughputPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Committableentries,
			prometheus.GaugeValue,
			float64(dst[0].Committableentries),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DataFilesSizeKB,
			prometheus.GaugeValue,
			float64(dst[0].DataFilesSizeKB),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBCCLogicalScanBytesPersec,
			prometheus.GaugeValue,
			float64(dst[0].DBCCLogicalScanBytesPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GroupCommitTimePersec,
			prometheus.GaugeValue,
			float64(dst[0].GroupCommitTimePersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogBytesFlushedPersec,
			prometheus.GaugeValue,
			float64(dst[0].LogBytesFlushedPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogCacheHitRatio,
			prometheus.GaugeValue,
			float64(dst[0].LogCacheHitRatio),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogCacheReadsPersec,
			prometheus.GaugeValue,
			float64(dst[0].LogCacheReadsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogFilesSizeKB,
			prometheus.GaugeValue,
			float64(dst[0].LogFilesSizeKB),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogFilesUsedSizeKB,
			prometheus.GaugeValue,
			float64(dst[0].LogFilesUsedSizeKB),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogFlushesPersec,
			prometheus.GaugeValue,
			float64(dst[0].LogFlushesPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogFlushWaitsPersec,
			prometheus.GaugeValue,
			float64(dst[0].LogFlushWaitsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogFlushWaitTime,
			prometheus.GaugeValue,
			float64(dst[0].LogFlushWaitTime),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogFlushWriteTimems,
			prometheus.GaugeValue,
			float64(dst[0].LogFlushWriteTimems),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogGrowths,
			prometheus.GaugeValue,
			float64(dst[0].LogGrowths),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolCacheMissesPersec,
			prometheus.GaugeValue,
			float64(dst[0].LogPoolCacheMissesPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolDiskReadsPersec,
			prometheus.GaugeValue,
			float64(dst[0].LogPoolDiskReadsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolHashDeletesPersec,
			prometheus.GaugeValue,
			float64(dst[0].LogPoolHashDeletesPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolHashInsertsPersec,
			prometheus.GaugeValue,
			float64(dst[0].LogPoolHashInsertsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolInvalidHashEntryPersec,
			prometheus.GaugeValue,
			float64(dst[0].LogPoolInvalidHashEntryPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolLogScanPushesPersec,
			prometheus.GaugeValue,
			float64(dst[0].LogPoolLogScanPushesPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolLogWriterPushesPersec,
			prometheus.GaugeValue,
			float64(dst[0].LogPoolLogWriterPushesPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolPushEmptyFreePoolPersec,
			prometheus.GaugeValue,
			float64(dst[0].LogPoolPushEmptyFreePoolPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolPushLowMemoryPersec,
			prometheus.GaugeValue,
			float64(dst[0].LogPoolPushLowMemoryPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolPushNoFreeBufferPersec,
			prometheus.GaugeValue,
			float64(dst[0].LogPoolPushNoFreeBufferPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolReqBehindTruncPersec,
			prometheus.GaugeValue,
			float64(dst[0].LogPoolReqBehindTruncPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolRequestsOldVLFPersec,
			prometheus.GaugeValue,
			float64(dst[0].LogPoolRequestsOldVLFPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolRequestsPersec,
			prometheus.GaugeValue,
			float64(dst[0].LogPoolRequestsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolTotalActiveLogSize,
			prometheus.GaugeValue,
			float64(dst[0].LogPoolTotalActiveLogSize),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolTotalSharedPoolSize,
			prometheus.GaugeValue,
			float64(dst[0].LogPoolTotalSharedPoolSize),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogShrinks,
			prometheus.GaugeValue,
			float64(dst[0].LogShrinks),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogTruncations,
			prometheus.GaugeValue,
			float64(dst[0].LogTruncations),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PercentLogUsed,
			prometheus.GaugeValue,
			float64(dst[0].PercentLogUsed),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReplPendingXacts,
			prometheus.GaugeValue,
			float64(dst[0].ReplPendingXacts),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReplTransRate,
			prometheus.GaugeValue,
			float64(dst[0].ReplTransRate),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ShrinkDataMovementBytesPersec,
			prometheus.GaugeValue,
			float64(dst[0].ShrinkDataMovementBytesPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TrackedtransactionsPersec,
			prometheus.GaugeValue,
			float64(dst[0].TrackedtransactionsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TransactionsPersec,
			prometheus.GaugeValue,
			float64(dst[0].TransactionsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WriteTransactionsPersec,
			prometheus.GaugeValue,
			float64(dst[0].WriteTransactionsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.XTPControllerDLCLatencyPerFetch,
			prometheus.GaugeValue,
			float64(dst[0].XTPControllerDLCLatencyPerFetch),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.XTPControllerDLCPeakLatency,
			prometheus.GaugeValue,
			float64(dst[0].XTPControllerDLCPeakLatency),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.XTPControllerLogProcessedPersec,
			prometheus.GaugeValue,
			float64(dst[0].XTPControllerLogProcessedPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.XTPMemoryUsedKB,
			prometheus.GaugeValue,
			float64(dst[0].XTPMemoryUsedKB),
			sqlInstance,
		)
	}

	return nil, nil
}

type win32PerfRawDataSQLServerGeneralStatistics struct {
	ActiveTempTables              uint64
	ConnectionResetPersec         uint64
	EventNotificationsDelayedDrop uint64
	HTTPAuthenticatedRequests     uint64
	LogicalConnections            uint64
	LoginsPersec                  uint64
	LogoutsPersec                 uint64
	MarsDeadlocks                 uint64
	Nonatomicyieldrate            uint64
	Processesblocked              uint64
	SOAPEmptyRequests             uint64
	SOAPMethodInvocations         uint64
	SOAPSessionInitiateRequests   uint64
	SOAPSessionTerminateRequests  uint64
	SOAPSQLRequests               uint64
	SOAPWSDLRequests              uint64
	SQLTraceIOProviderLockWaits   uint64
	Tempdbrecoveryunitid          uint64
	Tempdbrowsetid                uint64
	TempTablesCreationRate        uint64
	TempTablesForDestruction      uint64
	TraceEventNotificationQueue   uint64
	Transactions                  uint64
	UserConnections               uint64
}

func (c *MSSQLCollector) collectGeneralStatistics(ch chan<- prometheus.Metric, sqlInstance string) (*prometheus.Desc, error) {
	var dst []win32PerfRawDataSQLServerGeneralStatistics
	class := fmt.Sprintf("Win32_PerfRawData_%s_SQLServerGeneralStatistics", sqlInstance)
	q := queryAllForClass(&dst, class)

	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	if len(dst) > 0 {
		ch <- prometheus.MustNewConstMetric(
			c.ActiveTempTables,
			prometheus.GaugeValue,
			float64(dst[0].ActiveTempTables),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ConnectionResetPersec,
			prometheus.GaugeValue,
			float64(dst[0].ConnectionResetPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.EventNotificationsDelayedDrop,
			prometheus.GaugeValue,
			float64(dst[0].EventNotificationsDelayedDrop),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.HTTPAuthenticatedRequests,
			prometheus.GaugeValue,
			float64(dst[0].HTTPAuthenticatedRequests),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogicalConnections,
			prometheus.GaugeValue,
			float64(dst[0].LogicalConnections),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LoginsPersec,
			prometheus.GaugeValue,
			float64(dst[0].LoginsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogoutsPersec,
			prometheus.GaugeValue,
			float64(dst[0].LogoutsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MarsDeadlocks,
			prometheus.GaugeValue,
			float64(dst[0].MarsDeadlocks),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Nonatomicyieldrate,
			prometheus.GaugeValue,
			float64(dst[0].Nonatomicyieldrate),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Processesblocked,
			prometheus.GaugeValue,
			float64(dst[0].Processesblocked),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SOAPEmptyRequests,
			prometheus.GaugeValue,
			float64(dst[0].SOAPEmptyRequests),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SOAPMethodInvocations,
			prometheus.GaugeValue,
			float64(dst[0].SOAPMethodInvocations),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SOAPSessionInitiateRequests,
			prometheus.GaugeValue,
			float64(dst[0].SOAPSessionInitiateRequests),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SOAPSessionTerminateRequests,
			prometheus.GaugeValue,
			float64(dst[0].SOAPSessionTerminateRequests),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SOAPSQLRequests,
			prometheus.GaugeValue,
			float64(dst[0].SOAPSQLRequests),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SOAPWSDLRequests,
			prometheus.GaugeValue,
			float64(dst[0].SOAPWSDLRequests),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SQLTraceIOProviderLockWaits,
			prometheus.GaugeValue,
			float64(dst[0].SQLTraceIOProviderLockWaits),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Tempdbrecoveryunitid,
			prometheus.GaugeValue,
			float64(dst[0].Tempdbrecoveryunitid),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Tempdbrowsetid,
			prometheus.GaugeValue,
			float64(dst[0].Tempdbrowsetid),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TempTablesCreationRate,
			prometheus.GaugeValue,
			float64(dst[0].TempTablesCreationRate),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TempTablesForDestruction,
			prometheus.GaugeValue,
			float64(dst[0].TempTablesForDestruction),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TraceEventNotificationQueue,
			prometheus.GaugeValue,
			float64(dst[0].TraceEventNotificationQueue),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Transactions,
			prometheus.GaugeValue,
			float64(dst[0].Transactions),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.UserConnections,
			prometheus.GaugeValue,
			float64(dst[0].UserConnections),
			sqlInstance,
		)
	}

	return nil, nil
}

type win32PerfRawDataSQLServerLocks struct {
	AverageWaitTimems          uint64
	LockRequestsPersec         uint64
	LockTimeoutsPersec         uint64
	LockTimeoutstimeout0Persec uint64
	LockWaitsPersec            uint64
	LockWaitTimems             uint64
	NumberofDeadlocksPersec    uint64
}

func (c *MSSQLCollector) collectLocks(ch chan<- prometheus.Metric, sqlInstance string) (*prometheus.Desc, error) {
	var dst []win32PerfRawDataSQLServerLocks
	class := fmt.Sprintf("Win32_PerfRawData_%s_SQLServerLocks", sqlInstance)
	q := queryAllForClass(&dst, class)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	if len(dst) > 0 {
		ch <- prometheus.MustNewConstMetric(
			c.AverageWaitTimems,
			prometheus.GaugeValue,
			float64(dst[0].AverageWaitTimems),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockRequestsPersec,
			prometheus.GaugeValue,
			float64(dst[0].LockRequestsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockTimeoutsPersec,
			prometheus.GaugeValue,
			float64(dst[0].LockTimeoutsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockTimeoutstimeout0Persec,
			prometheus.GaugeValue,
			float64(dst[0].LockTimeoutstimeout0Persec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockWaitsPersec,
			prometheus.GaugeValue,
			float64(dst[0].LockWaitsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockWaitTimems,
			prometheus.GaugeValue,
			float64(dst[0].LockWaitTimems),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberofDeadlocksPersec,
			prometheus.GaugeValue,
			float64(dst[0].NumberofDeadlocksPersec),
			sqlInstance,
		)
	}

	return nil, nil
}

type win32PerfRawDataSQLServerMemoryManager struct {
	ConnectionMemoryKB       uint64
	DatabaseCacheMemoryKB    uint64
	Externalbenefitofmemory  uint64
	FreeMemoryKB             uint64
	GrantedWorkspaceMemoryKB uint64
	LockBlocks               uint64
	LockBlocksAllocated      uint64
	LockMemoryKB             uint64
	LockOwnerBlocks          uint64
	LockOwnerBlocksAllocated uint64
	LogPoolMemoryKB          uint64
	MaximumWorkspaceMemoryKB uint64
	MemoryGrantsOutstanding  uint64
	MemoryGrantsPending      uint64
	OptimizerMemoryKB        uint64
	ReservedServerMemoryKB   uint64
	SQLCacheMemoryKB         uint64
	StolenServerMemoryKB     uint64
	TargetServerMemoryKB     uint64
	TotalServerMemoryKB      uint64
}

func (c *MSSQLCollector) collectMemoryManager(ch chan<- prometheus.Metric, sqlInstance string) (*prometheus.Desc, error) {
	var dst []win32PerfRawDataSQLServerMemoryManager
	class := fmt.Sprintf("Win32_PerfRawData_%s_SQLServerMemoryManager", sqlInstance)
	q := queryAllForClass(&dst, class)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	if len(dst) > 0 {
		ch <- prometheus.MustNewConstMetric(
			c.ConnectionMemoryKB,
			prometheus.GaugeValue,
			float64(dst[0].ConnectionMemoryKB),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabaseCacheMemoryKB,
			prometheus.GaugeValue,
			float64(dst[0].DatabaseCacheMemoryKB),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Externalbenefitofmemory,
			prometheus.GaugeValue,
			float64(dst[0].Externalbenefitofmemory),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FreeMemoryKB,
			prometheus.GaugeValue,
			float64(dst[0].FreeMemoryKB),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GrantedWorkspaceMemoryKB,
			prometheus.GaugeValue,
			float64(dst[0].GrantedWorkspaceMemoryKB),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockBlocks,
			prometheus.GaugeValue,
			float64(dst[0].LockBlocks),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockBlocksAllocated,
			prometheus.GaugeValue,
			float64(dst[0].LockBlocksAllocated),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockMemoryKB,
			prometheus.GaugeValue,
			float64(dst[0].LockMemoryKB),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockOwnerBlocks,
			prometheus.GaugeValue,
			float64(dst[0].LockOwnerBlocks),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockOwnerBlocksAllocated,
			prometheus.GaugeValue,
			float64(dst[0].LockOwnerBlocksAllocated),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolMemoryKB,
			prometheus.GaugeValue,
			float64(dst[0].LogPoolMemoryKB),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MaximumWorkspaceMemoryKB,
			prometheus.GaugeValue,
			float64(dst[0].MaximumWorkspaceMemoryKB),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MemoryGrantsOutstanding,
			prometheus.GaugeValue,
			float64(dst[0].MemoryGrantsOutstanding),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MemoryGrantsPending,
			prometheus.GaugeValue,
			float64(dst[0].MemoryGrantsPending),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.OptimizerMemoryKB,
			prometheus.GaugeValue,
			float64(dst[0].OptimizerMemoryKB),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReservedServerMemoryKB,
			prometheus.GaugeValue,
			float64(dst[0].ReservedServerMemoryKB),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SQLCacheMemoryKB,
			prometheus.GaugeValue,
			float64(dst[0].SQLCacheMemoryKB),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.StolenServerMemoryKB,
			prometheus.GaugeValue,
			float64(dst[0].StolenServerMemoryKB),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TargetServerMemoryKB,
			prometheus.GaugeValue,
			float64(dst[0].TargetServerMemoryKB),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalServerMemoryKB,
			prometheus.GaugeValue,
			float64(dst[0].TotalServerMemoryKB),
			sqlInstance,
		)
	}

	return nil, nil
}

type win32PerfRawDataSQLServerSQLStatistics struct {
	AutoParamAttemptsPersec       uint64
	BatchRequestsPersec           uint64
	FailedAutoParamsPersec        uint64
	ForcedParameterizationsPersec uint64
	GuidedplanexecutionsPersec    uint64
	MisguidedplanexecutionsPersec uint64
	SafeAutoParamsPersec          uint64
	SQLAttentionrate              uint64
	SQLCompilationsPersec         uint64
	SQLReCompilationsPersec       uint64
	UnsafeAutoParamsPersec        uint64
}

func (c *MSSQLCollector) collectSQLStats(ch chan<- prometheus.Metric, sqlInstance string) (*prometheus.Desc, error) {
	var dst []win32PerfRawDataSQLServerSQLStatistics
	class := fmt.Sprintf("Win32_PerfRawData_%s_SQLServerSQLStatistics", sqlInstance)
	q := queryAllForClass(&dst, class)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	if len(dst) > 0 {
		ch <- prometheus.MustNewConstMetric(
			c.AutoParamAttemptsPersec,
			prometheus.GaugeValue,
			float64(dst[0].AutoParamAttemptsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BatchRequestsPersec,
			prometheus.GaugeValue,
			float64(dst[0].BatchRequestsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FailedAutoParamsPersec,
			prometheus.GaugeValue,
			float64(dst[0].FailedAutoParamsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ForcedParameterizationsPersec,
			prometheus.GaugeValue,
			float64(dst[0].ForcedParameterizationsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GuidedplanexecutionsPersec,
			prometheus.GaugeValue,
			float64(dst[0].GuidedplanexecutionsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MisguidedplanexecutionsPersec,
			prometheus.GaugeValue,
			float64(dst[0].MisguidedplanexecutionsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SafeAutoParamsPersec,
			prometheus.GaugeValue,
			float64(dst[0].SafeAutoParamsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SQLAttentionrate,
			prometheus.GaugeValue,
			float64(dst[0].SQLAttentionrate),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SQLCompilationsPersec,
			prometheus.GaugeValue,
			float64(dst[0].SQLCompilationsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SQLReCompilationsPersec,
			prometheus.GaugeValue,
			float64(dst[0].SQLReCompilationsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.UnsafeAutoParamsPersec,
			prometheus.GaugeValue,
			float64(dst[0].UnsafeAutoParamsPersec),
			sqlInstance,
		)
	}

	return nil, nil
}

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
	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	Factories["mssql"] = NewMSSQLCollector
}

// A MSSQLCollector is a Prometheus collector for various WMI Win32_PerfRawData_MSSQLSERVER_* metrics
type MSSQLCollector struct {
	// Win32_PerfRawData_MSSQLSERVER_SQLServerAvailabilityReplica
	BytesReceivedfromReplicaPersec *prometheus.Desc
	BytesSenttoReplicaPersec       *prometheus.Desc
	BytesSenttoTransportPersec     *prometheus.Desc
	FlowControlPersec              *prometheus.Desc
	FlowControlTimemsPersec        *prometheus.Desc
	ReceivesfromReplicaPersec      *prometheus.Desc
	ResentMessagesPersec           *prometheus.Desc
	SendstoReplicaPersec           *prometheus.Desc
	SendstoTransportPersec         *prometheus.Desc

	// Win32_PerfRawData_MSSQLSERVER_SQLServerBufferManager
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

	// Win32_PerfRawData_MSSQLSERVER_SQLServerDatabaseReplica
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

	// Win32_PerfRawData_MSSQLSERVER_SQLServerDatabases
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

	// Win32_PerfRawData_MSSQLSERVER_SQLServerGeneralStatistics
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

	// Win32_PerfRawData_MSSQLSERVER_SQLServerLocks
	AverageWaitTimems          *prometheus.Desc
	LockRequestsPersec         *prometheus.Desc
	LockTimeoutsPersec         *prometheus.Desc
	LockTimeoutstimeout0Persec *prometheus.Desc
	LockWaitsPersec            *prometheus.Desc
	LockWaitTimems             *prometheus.Desc
	NumberofDeadlocksPersec    *prometheus.Desc

	// Win32_PerfRawData_MSSQLSERVER_SQLServerMemoryManager
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

	// Win32_PerfRawData_MSSQLSERVER_SQLServerSQLStatistics
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

// NewMSSQLGeneralStatisticsCollector ...
func NewMSSQLCollector() (Collector, error) {
	const subsystem = "mssql"
	return &MSSQLCollector{
		// Win32_PerfRawData_MSSQLSERVER_SQLServerAvailabilityReplica
		BytesReceivedfromReplicaPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bytes_receivedfrom_replica_persec"),
			"(BytesReceivedfromReplicaPersec)",
			nil,
			nil,
		),
		BytesSenttoReplicaPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bytes_sentto_replica_persec"),
			"(BytesSenttoReplicaPersec)",
			nil,
			nil,
		),
		BytesSenttoTransportPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bytes_sentto_transport_persec"),
			"(BytesSenttoTransportPersec)",
			nil,
			nil,
		),
		FlowControlPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "flow_control_persec"),
			"(FlowControlPersec)",
			nil,
			nil,
		),
		FlowControlTimemsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "flow_control_timems_persec"),
			"(FlowControlTimemsPersec)",
			nil,
			nil,
		),
		ReceivesfromReplicaPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "receivesfrom_replica_persec"),
			"(ReceivesfromReplicaPersec)",
			nil,
			nil,
		),
		ResentMessagesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "resent_messages_persec"),
			"(ResentMessagesPersec)",
			nil,
			nil,
		),
		SendstoReplicaPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sendsto_replica_persec"),
			"(SendstoReplicaPersec)",
			nil,
			nil,
		),
		SendstoTransportPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sendsto_transport_persec"),
			"(SendstoTransportPersec)",
			nil,
			nil,
		),

		// Win32_PerfRawData_MSSQLSERVER_SQLServerBufferManager
		BackgroundwriterpagesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "backgroundwriterpages_persec"),
			"(BackgroundwriterpagesPersec)",
			nil,
			nil,
		),
		Buffercachehitratio: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "buffercachehitratio"),
			"(Buffercachehitratio)",
			nil,
			nil,
		),
		CheckpointpagesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "checkpointpages_persec"),
			"(CheckpointpagesPersec)",
			nil,
			nil,
		),
		Databasepages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databasepages"),
			"(Databasepages)",
			nil,
			nil,
		),
		Extensionallocatedpages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "extensionallocatedpages"),
			"(Extensionallocatedpages)",
			nil,
			nil,
		),
		Extensionfreepages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "extensionfreepages"),
			"(Extensionfreepages)",
			nil,
			nil,
		),
		Extensioninuseaspercentage: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "extensioninuseaspercentage"),
			"(Extensioninuseaspercentage)",
			nil,
			nil,
		),
		ExtensionoutstandingIOcounter: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "extensionoutstanding_i_ocounter"),
			"(ExtensionoutstandingIOcounter)",
			nil,
			nil,
		),
		ExtensionpageevictionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "extensionpageevictions_persec"),
			"(ExtensionpageevictionsPersec)",
			nil,
			nil,
		),
		ExtensionpagereadsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "extensionpagereads_persec"),
			"(ExtensionpagereadsPersec)",
			nil,
			nil,
		),
		Extensionpageunreferencedtime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "extensionpageunreferencedtime"),
			"(Extensionpageunreferencedtime)",
			nil,
			nil,
		),
		ExtensionpagewritesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "extensionpagewrites_persec"),
			"(ExtensionpagewritesPersec)",
			nil,
			nil,
		),
		FreeliststallsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "freeliststalls_persec"),
			"(FreeliststallsPersec)",
			nil,
			nil,
		),
		IntegralControllerSlope: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "integral_controller_slope"),
			"(IntegralControllerSlope)",
			nil,
			nil,
		),
		LazywritesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lazywrites_persec"),
			"(LazywritesPersec)",
			nil,
			nil,
		),
		Pagelifeexpectancy: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pagelifeexpectancy"),
			"(Pagelifeexpectancy)",
			nil,
			nil,
		),
		PagelookupsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pagelookups_persec"),
			"(PagelookupsPersec)",
			nil,
			nil,
		),
		PagereadsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pagereads_persec"),
			"(PagereadsPersec)",
			nil,
			nil,
		),
		PagewritesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pagewrites_persec"),
			"(PagewritesPersec)",
			nil,
			nil,
		),
		ReadaheadpagesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "readaheadpages_persec"),
			"(ReadaheadpagesPersec)",
			nil,
			nil,
		),
		ReadaheadtimePersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "readaheadtime_persec"),
			"(ReadaheadtimePersec)",
			nil,
			nil,
		),
		Targetpages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "targetpages"),
			"(Targetpages)",
			nil,
			nil,
		),

		// Win32_PerfRawData_MSSQLSERVER_SQLServerDatabaseReplica
		DatabaseFlowControlDelay: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "database_flow_control_delay"),
			"(DatabaseFlowControlDelay)",
			nil,
			nil,
		),
		DatabaseFlowControlsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "database_flow_controls_persec"),
			"(DatabaseFlowControlsPersec)",
			nil,
			nil,
		),
		FileBytesReceivedPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "file_bytes_received_persec"),
			"(FileBytesReceivedPersec)",
			nil,
			nil,
		),
		GroupCommitsPerSec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "group_commits_per_sec"),
			"(GroupCommitsPerSec)",
			nil,
			nil,
		),
		GroupCommitTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "group_commit_time"),
			"(GroupCommitTime)",
			nil,
			nil,
		),
		LogApplyPendingQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_apply_pending_queue"),
			"(LogApplyPendingQueue)",
			nil,
			nil,
		),
		LogApplyReadyQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_apply_ready_queue"),
			"(LogApplyReadyQueue)",
			nil,
			nil,
		),
		LogBytesCompressedPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_bytes_compressed_persec"),
			"(LogBytesCompressedPersec)",
			nil,
			nil,
		),
		LogBytesDecompressedPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_bytes_decompressed_persec"),
			"(LogBytesDecompressedPersec)",
			nil,
			nil,
		),
		LogBytesReceivedPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_bytes_received_persec"),
			"(LogBytesReceivedPersec)",
			nil,
			nil,
		),
		LogCompressionCachehitsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_compression_cachehits_persec"),
			"(LogCompressionCachehitsPersec)",
			nil,
			nil,
		),
		LogCompressionCachemissesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_compression_cachemisses_persec"),
			"(LogCompressionCachemissesPersec)",
			nil,
			nil,
		),
		LogCompressionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_compressions_persec"),
			"(LogCompressionsPersec)",
			nil,
			nil,
		),
		LogDecompressionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_decompressions_persec"),
			"(LogDecompressionsPersec)",
			nil,
			nil,
		),
		Logremainingforundo: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "logremainingforundo"),
			"(Logremainingforundo)",
			nil,
			nil,
		),
		LogSendQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_send_queue"),
			"(LogSendQueue)",
			nil,
			nil,
		),
		MirroredWriteTransactionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "mirrored_write_transactions_persec"),
			"(MirroredWriteTransactionsPersec)",
			nil,
			nil,
		),
		RecoveryQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "recovery_queue"),
			"(RecoveryQueue)",
			nil,
			nil,
		),
		RedoblockedPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "redoblocked_persec"),
			"(RedoblockedPersec)",
			nil,
			nil,
		),
		RedoBytesRemaining: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "redo_bytes_remaining"),
			"(RedoBytesRemaining)",
			nil,
			nil,
		),
		RedoneBytesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "redone_bytes_persec"),
			"(RedoneBytesPersec)",
			nil,
			nil,
		),
		RedonesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "redones_persec"),
			"(RedonesPersec)",
			nil,
			nil,
		),
		TotalLogrequiringundo: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "total_logrequiringundo"),
			"(TotalLogrequiringundo)",
			nil,
			nil,
		),
		TransactionDelay: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "transaction_delay"),
			"(TransactionDelay)",
			nil,
			nil,
		),

		// Win32_PerfRawData_MSSQLSERVER_SQLServerDatabases
		ActiveTransactions: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "active_transactions"),
			"(ActiveTransactions)",
			nil,
			nil,
		),
		AvgDistFromEOLLPRequest: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "avg_dist_from_eollp_request"),
			"(AvgDistFromEOLLPRequest)",
			nil,
			nil,
		),
		BackupPerRestoreThroughputPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "backup_per_restore_throughput_persec"),
			"(BackupPerRestoreThroughputPersec)",
			nil,
			nil,
		),
		BulkCopyRowsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bulk_copy_rows_persec"),
			"(BulkCopyRowsPersec)",
			nil,
			nil,
		),
		BulkCopyThroughputPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bulk_copy_throughput_persec"),
			"(BulkCopyThroughputPersec)",
			nil,
			nil,
		),
		Committableentries: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "committableentries"),
			"(Committableentries)",
			nil,
			nil,
		),
		DataFilesSizeKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "data_files_size_kb"),
			"(DataFilesSizeKB)",
			nil,
			nil,
		),
		DBCCLogicalScanBytesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbcc_logical_scan_bytes_persec"),
			"(DBCCLogicalScanBytesPersec)",
			nil,
			nil,
		),
		GroupCommitTimePersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "group_commit_time_persec"),
			"(GroupCommitTimePersec)",
			nil,
			nil,
		),
		LogBytesFlushedPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_bytes_flushed_persec"),
			"(LogBytesFlushedPersec)",
			nil,
			nil,
		),
		LogCacheHitRatio: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_cache_hit_ratio"),
			"(LogCacheHitRatio)",
			nil,
			nil,
		),
		LogCacheReadsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_cache_reads_persec"),
			"(LogCacheReadsPersec)",
			nil,
			nil,
		),
		LogFilesSizeKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_files_size_kb"),
			"(LogFilesSizeKB)",
			nil,
			nil,
		),
		LogFilesUsedSizeKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_files_used_size_kb"),
			"(LogFilesUsedSizeKB)",
			nil,
			nil,
		),
		LogFlushesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_flushes_persec"),
			"(LogFlushesPersec)",
			nil,
			nil,
		),
		LogFlushWaitsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_flush_waits_persec"),
			"(LogFlushWaitsPersec)",
			nil,
			nil,
		),
		LogFlushWaitTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_flush_wait_time"),
			"(LogFlushWaitTime)",
			nil,
			nil,
		),
		LogFlushWriteTimems: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_flush_write_timems"),
			"(LogFlushWriteTimems)",
			nil,
			nil,
		),
		LogGrowths: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_growths"),
			"(LogGrowths)",
			nil,
			nil,
		),
		LogPoolCacheMissesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_cache_misses_persec"),
			"(LogPoolCacheMissesPersec)",
			nil,
			nil,
		),
		LogPoolDiskReadsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_disk_reads_persec"),
			"(LogPoolDiskReadsPersec)",
			nil,
			nil,
		),
		LogPoolHashDeletesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_hash_deletes_persec"),
			"(LogPoolHashDeletesPersec)",
			nil,
			nil,
		),
		LogPoolHashInsertsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_hash_inserts_persec"),
			"(LogPoolHashInsertsPersec)",
			nil,
			nil,
		),
		LogPoolInvalidHashEntryPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_invalid_hash_entry_persec"),
			"(LogPoolInvalidHashEntryPersec)",
			nil,
			nil,
		),
		LogPoolLogScanPushesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_log_scan_pushes_persec"),
			"(LogPoolLogScanPushesPersec)",
			nil,
			nil,
		),
		LogPoolLogWriterPushesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_log_writer_pushes_persec"),
			"(LogPoolLogWriterPushesPersec)",
			nil,
			nil,
		),
		LogPoolPushEmptyFreePoolPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_push_empty_free_pool_persec"),
			"(LogPoolPushEmptyFreePoolPersec)",
			nil,
			nil,
		),
		LogPoolPushLowMemoryPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_push_low_memory_persec"),
			"(LogPoolPushLowMemoryPersec)",
			nil,
			nil,
		),
		LogPoolPushNoFreeBufferPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_push_no_free_buffer_persec"),
			"(LogPoolPushNoFreeBufferPersec)",
			nil,
			nil,
		),
		LogPoolReqBehindTruncPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_req_behind_trunc_persec"),
			"(LogPoolReqBehindTruncPersec)",
			nil,
			nil,
		),
		LogPoolRequestsOldVLFPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_requests_old_vlf_persec"),
			"(LogPoolRequestsOldVLFPersec)",
			nil,
			nil,
		),
		LogPoolRequestsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_requests_persec"),
			"(LogPoolRequestsPersec)",
			nil,
			nil,
		),
		LogPoolTotalActiveLogSize: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_total_active_log_size"),
			"(LogPoolTotalActiveLogSize)",
			nil,
			nil,
		),
		LogPoolTotalSharedPoolSize: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_total_shared_pool_size"),
			"(LogPoolTotalSharedPoolSize)",
			nil,
			nil,
		),
		LogShrinks: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_shrinks"),
			"(LogShrinks)",
			nil,
			nil,
		),
		LogTruncations: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_truncations"),
			"(LogTruncations)",
			nil,
			nil,
		),
		PercentLogUsed: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "percent_log_used"),
			"(PercentLogUsed)",
			nil,
			nil,
		),
		ReplPendingXacts: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "repl_pending_xacts"),
			"(ReplPendingXacts)",
			nil,
			nil,
		),
		ReplTransRate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "repl_trans_rate"),
			"(ReplTransRate)",
			nil,
			nil,
		),
		ShrinkDataMovementBytesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "shrink_data_movement_bytes_persec"),
			"(ShrinkDataMovementBytesPersec)",
			nil,
			nil,
		),
		TrackedtransactionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "trackedtransactions_persec"),
			"(TrackedtransactionsPersec)",
			nil,
			nil,
		),
		TransactionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "transactions_persec"),
			"(TransactionsPersec)",
			nil,
			nil,
		),
		WriteTransactionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "write_transactions_persec"),
			"(WriteTransactionsPersec)",
			nil,
			nil,
		),
		XTPControllerDLCLatencyPerFetch: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "xtp_controller_dlc_latency_per_fetch"),
			"(XTPControllerDLCLatencyPerFetch)",
			nil,
			nil,
		),
		XTPControllerDLCPeakLatency: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "xtp_controller_dlc_peak_latency"),
			"(XTPControllerDLCPeakLatency)",
			nil,
			nil,
		),
		XTPControllerLogProcessedPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "xtp_controller_log_processed_persec"),
			"(XTPControllerLogProcessedPersec)",
			nil,
			nil,
		),
		XTPMemoryUsedKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "xtp_memory_used_kb"),
			"(XTPMemoryUsedKB)",
			nil,
			nil,
		),

		// Win32_PerfRawData_MSSQLSERVER_SQLServerGeneralStatistics
		ActiveTempTables: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "active_temp_tables"),
			"(ActiveTempTables)",
			nil,
			nil,
		),
		ConnectionResetPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "connection_reset_persec"),
			"(ConnectionResetPersec)",
			nil,
			nil,
		),
		EventNotificationsDelayedDrop: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "event_notifications_delayed_drop"),
			"(EventNotificationsDelayedDrop)",
			nil,
			nil,
		),
		HTTPAuthenticatedRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "http_authenticated_requests"),
			"(HTTPAuthenticatedRequests)",
			nil,
			nil,
		),
		LogicalConnections: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "logical_connections"),
			"(LogicalConnections)",
			nil,
			nil,
		),
		LoginsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "logins_persec"),
			"(LoginsPersec)",
			nil,
			nil,
		),
		LogoutsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "logouts_persec"),
			"(LogoutsPersec)",
			nil,
			nil,
		),
		MarsDeadlocks: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "mars_deadlocks"),
			"(MarsDeadlocks)",
			nil,
			nil,
		),
		Nonatomicyieldrate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "nonatomicyieldrate"),
			"(Nonatomicyieldrate)",
			nil,
			nil,
		),
		Processesblocked: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "processesblocked"),
			"(Processesblocked)",
			nil,
			nil,
		),
		SOAPEmptyRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "soap_empty_requests"),
			"(SOAPEmptyRequests)",
			nil,
			nil,
		),
		SOAPMethodInvocations: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "soap_method_invocations"),
			"(SOAPMethodInvocations)",
			nil,
			nil,
		),
		SOAPSessionInitiateRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "soap_session_initiate_requests"),
			"(SOAPSessionInitiateRequests)",
			nil,
			nil,
		),
		SOAPSessionTerminateRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "soap_session_terminate_requests"),
			"(SOAPSessionTerminateRequests)",
			nil,
			nil,
		),
		SOAPSQLRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "soapsql_requests"),
			"(SOAPSQLRequests)",
			nil,
			nil,
		),
		SOAPWSDLRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "soapwsdl_requests"),
			"(SOAPWSDLRequests)",
			nil,
			nil,
		),
		SQLTraceIOProviderLockWaits: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sql_trace_io_provider_lock_waits"),
			"(SQLTraceIOProviderLockWaits)",
			nil,
			nil,
		),
		Tempdbrecoveryunitid: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "tempdbrecoveryunitid"),
			"(Tempdbrecoveryunitid)",
			nil,
			nil,
		),
		Tempdbrowsetid: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "tempdbrowsetid"),
			"(Tempdbrowsetid)",
			nil,
			nil,
		),
		TempTablesCreationRate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "temp_tables_creation_rate"),
			"(TempTablesCreationRate)",
			nil,
			nil,
		),
		TempTablesForDestruction: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "temp_tables_for_destruction"),
			"(TempTablesForDestruction)",
			nil,
			nil,
		),
		TraceEventNotificationQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "trace_event_notification_queue"),
			"(TraceEventNotificationQueue)",
			nil,
			nil,
		),
		Transactions: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "transactions"),
			"(Transactions)",
			nil,
			nil,
		),
		UserConnections: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "user_connections"),
			"(UserConnections)",
			nil,
			nil,
		),

		// Win32_PerfRawData_MSSQLSERVER_SQLServerLocks
		AverageWaitTimems: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "average_wait_timems"),
			"(AverageWaitTimems)",
			nil,
			nil,
		),
		LockRequestsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_requests_persec"),
			"(LockRequestsPersec)",
			nil,
			nil,
		),
		LockTimeoutsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_timeouts_persec"),
			"(LockTimeoutsPersec)",
			nil,
			nil,
		),
		LockTimeoutstimeout0Persec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_timeoutstimeout0_persec"),
			"(LockTimeoutstimeout0Persec)",
			nil,
			nil,
		),
		LockWaitsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_waits_persec"),
			"(LockWaitsPersec)",
			nil,
			nil,
		),
		LockWaitTimems: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_wait_timems"),
			"(LockWaitTimems)",
			nil,
			nil,
		),
		NumberofDeadlocksPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "numberof_deadlocks_persec"),
			"(NumberofDeadlocksPersec)",
			nil,
			nil,
		),

		// Win32_PerfRawData_MSSQLSERVER_SQLServerMemoryManager
		ConnectionMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "connection_memory_kb"),
			"(ConnectionMemoryKB)",
			nil,
			nil,
		),
		DatabaseCacheMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "database_cache_memory_kb"),
			"(DatabaseCacheMemoryKB)",
			nil,
			nil,
		),
		Externalbenefitofmemory: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "externalbenefitofmemory"),
			"(Externalbenefitofmemory)",
			nil,
			nil,
		),
		FreeMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "free_memory_kb"),
			"(FreeMemoryKB)",
			nil,
			nil,
		),
		GrantedWorkspaceMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "granted_workspace_memory_kb"),
			"(GrantedWorkspaceMemoryKB)",
			nil,
			nil,
		),
		LockBlocks: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_blocks"),
			"(LockBlocks)",
			nil,
			nil,
		),
		LockBlocksAllocated: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_blocks_allocated"),
			"(LockBlocksAllocated)",
			nil,
			nil,
		),
		LockMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_memory_kb"),
			"(LockMemoryKB)",
			nil,
			nil,
		),
		LockOwnerBlocks: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_owner_blocks"),
			"(LockOwnerBlocks)",
			nil,
			nil,
		),
		LockOwnerBlocksAllocated: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_owner_blocks_allocated"),
			"(LockOwnerBlocksAllocated)",
			nil,
			nil,
		),
		LogPoolMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_memory_kb"),
			"(LogPoolMemoryKB)",
			nil,
			nil,
		),
		MaximumWorkspaceMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "maximum_workspace_memory_kb"),
			"(MaximumWorkspaceMemoryKB)",
			nil,
			nil,
		),
		MemoryGrantsOutstanding: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memory_grants_outstanding"),
			"(MemoryGrantsOutstanding)",
			nil,
			nil,
		),
		MemoryGrantsPending: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memory_grants_pending"),
			"(MemoryGrantsPending)",
			nil,
			nil,
		),
		OptimizerMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "optimizer_memory_kb"),
			"(OptimizerMemoryKB)",
			nil,
			nil,
		),
		ReservedServerMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "reserved_server_memory_kb"),
			"(ReservedServerMemoryKB)",
			nil,
			nil,
		),
		SQLCacheMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sql_cache_memory_kb"),
			"(SQLCacheMemoryKB)",
			nil,
			nil,
		),
		StolenServerMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "stolen_server_memory_kb"),
			"(StolenServerMemoryKB)",
			nil,
			nil,
		),
		TargetServerMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "target_server_memory_kb"),
			"(TargetServerMemoryKB)",
			nil,
			nil,
		),
		TotalServerMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "total_server_memory_kb"),
			"(TotalServerMemoryKB)",
			nil,
			nil,
		),

		// Win32_PerfRawData_MSSQLSERVER_SQLServerSQLStatistics
		AutoParamAttemptsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "auto_param_attempts_persec"),
			"(AutoParamAttemptsPersec)",
			nil,
			nil,
		),
		BatchRequestsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "batch_requests_persec"),
			"(BatchRequestsPersec)",
			nil,
			nil,
		),
		FailedAutoParamsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "failed_auto_params_persec"),
			"(FailedAutoParamsPersec)",
			nil,
			nil,
		),
		ForcedParameterizationsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "forced_parameterizations_persec"),
			"(ForcedParameterizationsPersec)",
			nil,
			nil,
		),
		GuidedplanexecutionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "guidedplanexecutions_persec"),
			"(GuidedplanexecutionsPersec)",
			nil,
			nil,
		),
		MisguidedplanexecutionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "misguidedplanexecutions_persec"),
			"(MisguidedplanexecutionsPersec)",
			nil,
			nil,
		),
		SafeAutoParamsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "safe_auto_params_persec"),
			"(SafeAutoParamsPersec)",
			nil,
			nil,
		),
		SQLAttentionrate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sql_attentionrate"),
			"(SQLAttentionrate)",
			nil,
			nil,
		),
		SQLCompilationsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sql_compilations_persec"),
			"(SQLCompilationsPersec)",
			nil,
			nil,
		),
		SQLReCompilationsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sql_re_compilations_persec"),
			"(SQLReCompilationsPersec)",
			nil,
			nil,
		),
		UnsafeAutoParamsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "unsafe_auto_params_persec"),
			"(UnsafeAutoParamsPersec)",
			nil,
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *MSSQLCollector) Collect(ch chan<- prometheus.Metric) error {
	// Win32_PerfRawData_MSSQLSERVER_SQLServerAvailabilityReplica
	if desc, err := c.collectAvailabilityReplica(ch); err != nil {
		log.Error("failed collecting MSSQL GeneralStatistics metrics:", desc, err)
		return err
	}
	// Win32_PerfRawData_MSSQLSERVER_SQLServerBufferManager
	if desc, err := c.collectBufferManager(ch); err != nil {
		log.Error("failed collecting MSSQL BufferManager metrics:", desc, err)
		return err
	}

	// Win32_PerfRawData_MSSQLSERVER_SQLServerDatabaseReplica
	if desc, err := c.collectDatabaseReplica(ch); err != nil {
		log.Error("failed collecting MSSQL DatabaseReplica metrics:", desc, err)
		return err
	}

	// Win32_PerfRawData_MSSQLSERVER_SQLServerDatabases
	if desc, err := c.collectDatabases(ch); err != nil {
		log.Error("failed collecting MSSQL Databases metrics:", desc, err)
		return err
	}

	// Win32_PerfRawData_MSSQLSERVER_SQLServerGeneralStatistics
	if desc, err := c.collectGeneralStatistics(ch); err != nil {
		log.Error("failed collecting MSSQL GeneralStatistics metrics:", desc, err)
		return err
	}
	// Win32_PerfRawData_MSSQLSERVER_SQLServerLocks
	if desc, err := c.collectLocks(ch); err != nil {
		log.Error("failed collecting MSSQL Locks metrics:", desc, err)
		return err
	}
	// Win32_PerfRawData_MSSQLSERVER_SQLServerMemoryManager
	if desc, err := c.collectMemoryManager(ch); err != nil {
		log.Error("failed collecting MSSQL MemoryManager metrics:", desc, err)
		return err
	}
	// Win32_PerfRawData_MSSQLSERVER_SQLServerSQLStatistics
	if desc, err := c.collectSQLStats(ch); err != nil {
		log.Error("failed collecting MSSQL SQLStats metrics:", desc, err)
		return err
	}

	return nil
}

type Win32_PerfRawData_MSSQLSERVER_SQLServerAvailabilityReplica struct {
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

func (c *MSSQLCollector) collectAvailabilityReplica(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_MSSQLSERVER_SQLServerAvailabilityReplica
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	ch <- prometheus.MustNewConstMetric(
		c.BytesReceivedfromReplicaPersec,
		prometheus.GaugeValue,
		float64(dst[0].BytesReceivedfromReplicaPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.BytesSenttoReplicaPersec,
		prometheus.GaugeValue,
		float64(dst[0].BytesSenttoReplicaPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.BytesSenttoTransportPersec,
		prometheus.GaugeValue,
		float64(dst[0].BytesSenttoTransportPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.FlowControlPersec,
		prometheus.GaugeValue,
		float64(dst[0].FlowControlPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.FlowControlTimemsPersec,
		prometheus.GaugeValue,
		float64(dst[0].FlowControlTimemsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ReceivesfromReplicaPersec,
		prometheus.GaugeValue,
		float64(dst[0].ReceivesfromReplicaPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ResentMessagesPersec,
		prometheus.GaugeValue,
		float64(dst[0].ResentMessagesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SendstoReplicaPersec,
		prometheus.GaugeValue,
		float64(dst[0].SendstoReplicaPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SendstoTransportPersec,
		prometheus.GaugeValue,
		float64(dst[0].SendstoTransportPersec),
	)

	return nil, nil
}

type Win32_PerfRawData_MSSQLSERVER_SQLServerBufferManager struct {
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

func (c *MSSQLCollector) collectBufferManager(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_MSSQLSERVER_SQLServerBufferManager
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	ch <- prometheus.MustNewConstMetric(
		c.BackgroundwriterpagesPersec,
		prometheus.GaugeValue,
		float64(dst[0].BackgroundwriterpagesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.Buffercachehitratio,
		prometheus.GaugeValue,
		float64(dst[0].Buffercachehitratio),
	)

	ch <- prometheus.MustNewConstMetric(
		c.CheckpointpagesPersec,
		prometheus.GaugeValue,
		float64(dst[0].CheckpointpagesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.Databasepages,
		prometheus.GaugeValue,
		float64(dst[0].Databasepages),
	)

	ch <- prometheus.MustNewConstMetric(
		c.Extensionallocatedpages,
		prometheus.GaugeValue,
		float64(dst[0].Extensionallocatedpages),
	)

	ch <- prometheus.MustNewConstMetric(
		c.Extensionfreepages,
		prometheus.GaugeValue,
		float64(dst[0].Extensionfreepages),
	)

	ch <- prometheus.MustNewConstMetric(
		c.Extensioninuseaspercentage,
		prometheus.GaugeValue,
		float64(dst[0].Extensioninuseaspercentage),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ExtensionoutstandingIOcounter,
		prometheus.GaugeValue,
		float64(dst[0].ExtensionoutstandingIOcounter),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ExtensionpageevictionsPersec,
		prometheus.GaugeValue,
		float64(dst[0].ExtensionpageevictionsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ExtensionpagereadsPersec,
		prometheus.GaugeValue,
		float64(dst[0].ExtensionpagereadsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.Extensionpageunreferencedtime,
		prometheus.GaugeValue,
		float64(dst[0].Extensionpageunreferencedtime),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ExtensionpagewritesPersec,
		prometheus.GaugeValue,
		float64(dst[0].ExtensionpagewritesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.FreeliststallsPersec,
		prometheus.GaugeValue,
		float64(dst[0].FreeliststallsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.IntegralControllerSlope,
		prometheus.GaugeValue,
		float64(dst[0].IntegralControllerSlope),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LazywritesPersec,
		prometheus.GaugeValue,
		float64(dst[0].LazywritesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.Pagelifeexpectancy,
		prometheus.GaugeValue,
		float64(dst[0].Pagelifeexpectancy),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PagelookupsPersec,
		prometheus.GaugeValue,
		float64(dst[0].PagelookupsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PagereadsPersec,
		prometheus.GaugeValue,
		float64(dst[0].PagereadsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PagewritesPersec,
		prometheus.GaugeValue,
		float64(dst[0].PagewritesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ReadaheadpagesPersec,
		prometheus.GaugeValue,
		float64(dst[0].ReadaheadpagesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ReadaheadtimePersec,
		prometheus.GaugeValue,
		float64(dst[0].ReadaheadtimePersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.Targetpages,
		prometheus.GaugeValue,
		float64(dst[0].Targetpages),
	)

	return nil, nil
}

type Win32_PerfRawData_MSSQLSERVER_SQLServerDatabaseReplica struct {
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

func (c *MSSQLCollector) collectDatabaseReplica(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_MSSQLSERVER_SQLServerDatabaseReplica
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	ch <- prometheus.MustNewConstMetric(
		c.DatabaseFlowControlDelay,
		prometheus.GaugeValue,
		float64(dst[0].DatabaseFlowControlDelay),
	)

	ch <- prometheus.MustNewConstMetric(
		c.DatabaseFlowControlsPersec,
		prometheus.GaugeValue,
		float64(dst[0].DatabaseFlowControlsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.FileBytesReceivedPersec,
		prometheus.GaugeValue,
		float64(dst[0].FileBytesReceivedPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.GroupCommitsPerSec,
		prometheus.GaugeValue,
		float64(dst[0].GroupCommitsPerSec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.GroupCommitTime,
		prometheus.GaugeValue,
		float64(dst[0].GroupCommitTime),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogApplyPendingQueue,
		prometheus.GaugeValue,
		float64(dst[0].LogApplyPendingQueue),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogApplyReadyQueue,
		prometheus.GaugeValue,
		float64(dst[0].LogApplyReadyQueue),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogBytesCompressedPersec,
		prometheus.GaugeValue,
		float64(dst[0].LogBytesCompressedPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogBytesDecompressedPersec,
		prometheus.GaugeValue,
		float64(dst[0].LogBytesDecompressedPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogBytesReceivedPersec,
		prometheus.GaugeValue,
		float64(dst[0].LogBytesReceivedPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogCompressionCachehitsPersec,
		prometheus.GaugeValue,
		float64(dst[0].LogCompressionCachehitsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogCompressionCachemissesPersec,
		prometheus.GaugeValue,
		float64(dst[0].LogCompressionCachemissesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogCompressionsPersec,
		prometheus.GaugeValue,
		float64(dst[0].LogCompressionsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogDecompressionsPersec,
		prometheus.GaugeValue,
		float64(dst[0].LogDecompressionsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.Logremainingforundo,
		prometheus.GaugeValue,
		float64(dst[0].Logremainingforundo),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogSendQueue,
		prometheus.GaugeValue,
		float64(dst[0].LogSendQueue),
	)

	ch <- prometheus.MustNewConstMetric(
		c.MirroredWriteTransactionsPersec,
		prometheus.GaugeValue,
		float64(dst[0].MirroredWriteTransactionsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RecoveryQueue,
		prometheus.GaugeValue,
		float64(dst[0].RecoveryQueue),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RedoblockedPersec,
		prometheus.GaugeValue,
		float64(dst[0].RedoblockedPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RedoBytesRemaining,
		prometheus.GaugeValue,
		float64(dst[0].RedoBytesRemaining),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RedoneBytesPersec,
		prometheus.GaugeValue,
		float64(dst[0].RedoneBytesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RedonesPersec,
		prometheus.GaugeValue,
		float64(dst[0].RedonesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.TotalLogrequiringundo,
		prometheus.GaugeValue,
		float64(dst[0].TotalLogrequiringundo),
	)

	ch <- prometheus.MustNewConstMetric(
		c.TransactionDelay,
		prometheus.GaugeValue,
		float64(dst[0].TransactionDelay),
	)

	return nil, nil
}

type Win32_PerfRawData_MSSQLSERVER_SQLServerDatabases struct {
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

func (c *MSSQLCollector) collectDatabases(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_MSSQLSERVER_SQLServerDatabases
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	ch <- prometheus.MustNewConstMetric(
		c.ActiveTransactions,
		prometheus.GaugeValue,
		float64(dst[0].ActiveTransactions),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AvgDistFromEOLLPRequest,
		prometheus.GaugeValue,
		float64(dst[0].AvgDistFromEOLLPRequest),
	)

	ch <- prometheus.MustNewConstMetric(
		c.BackupPerRestoreThroughputPersec,
		prometheus.GaugeValue,
		float64(dst[0].BackupPerRestoreThroughputPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.BulkCopyRowsPersec,
		prometheus.GaugeValue,
		float64(dst[0].BulkCopyRowsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.BulkCopyThroughputPersec,
		prometheus.GaugeValue,
		float64(dst[0].BulkCopyThroughputPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.Committableentries,
		prometheus.GaugeValue,
		float64(dst[0].Committableentries),
	)

	ch <- prometheus.MustNewConstMetric(
		c.DataFilesSizeKB,
		prometheus.GaugeValue,
		float64(dst[0].DataFilesSizeKB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.DBCCLogicalScanBytesPersec,
		prometheus.GaugeValue,
		float64(dst[0].DBCCLogicalScanBytesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.GroupCommitTimePersec,
		prometheus.GaugeValue,
		float64(dst[0].GroupCommitTimePersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogBytesFlushedPersec,
		prometheus.GaugeValue,
		float64(dst[0].LogBytesFlushedPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogCacheHitRatio,
		prometheus.GaugeValue,
		float64(dst[0].LogCacheHitRatio),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogCacheReadsPersec,
		prometheus.GaugeValue,
		float64(dst[0].LogCacheReadsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogFilesSizeKB,
		prometheus.GaugeValue,
		float64(dst[0].LogFilesSizeKB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogFilesUsedSizeKB,
		prometheus.GaugeValue,
		float64(dst[0].LogFilesUsedSizeKB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogFlushesPersec,
		prometheus.GaugeValue,
		float64(dst[0].LogFlushesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogFlushWaitsPersec,
		prometheus.GaugeValue,
		float64(dst[0].LogFlushWaitsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogFlushWaitTime,
		prometheus.GaugeValue,
		float64(dst[0].LogFlushWaitTime),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogFlushWriteTimems,
		prometheus.GaugeValue,
		float64(dst[0].LogFlushWriteTimems),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogGrowths,
		prometheus.GaugeValue,
		float64(dst[0].LogGrowths),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogPoolCacheMissesPersec,
		prometheus.GaugeValue,
		float64(dst[0].LogPoolCacheMissesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogPoolDiskReadsPersec,
		prometheus.GaugeValue,
		float64(dst[0].LogPoolDiskReadsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogPoolHashDeletesPersec,
		prometheus.GaugeValue,
		float64(dst[0].LogPoolHashDeletesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogPoolHashInsertsPersec,
		prometheus.GaugeValue,
		float64(dst[0].LogPoolHashInsertsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogPoolInvalidHashEntryPersec,
		prometheus.GaugeValue,
		float64(dst[0].LogPoolInvalidHashEntryPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogPoolLogScanPushesPersec,
		prometheus.GaugeValue,
		float64(dst[0].LogPoolLogScanPushesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogPoolLogWriterPushesPersec,
		prometheus.GaugeValue,
		float64(dst[0].LogPoolLogWriterPushesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogPoolPushEmptyFreePoolPersec,
		prometheus.GaugeValue,
		float64(dst[0].LogPoolPushEmptyFreePoolPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogPoolPushLowMemoryPersec,
		prometheus.GaugeValue,
		float64(dst[0].LogPoolPushLowMemoryPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogPoolPushNoFreeBufferPersec,
		prometheus.GaugeValue,
		float64(dst[0].LogPoolPushNoFreeBufferPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogPoolReqBehindTruncPersec,
		prometheus.GaugeValue,
		float64(dst[0].LogPoolReqBehindTruncPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogPoolRequestsOldVLFPersec,
		prometheus.GaugeValue,
		float64(dst[0].LogPoolRequestsOldVLFPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogPoolRequestsPersec,
		prometheus.GaugeValue,
		float64(dst[0].LogPoolRequestsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogPoolTotalActiveLogSize,
		prometheus.GaugeValue,
		float64(dst[0].LogPoolTotalActiveLogSize),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogPoolTotalSharedPoolSize,
		prometheus.GaugeValue,
		float64(dst[0].LogPoolTotalSharedPoolSize),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogShrinks,
		prometheus.GaugeValue,
		float64(dst[0].LogShrinks),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogTruncations,
		prometheus.GaugeValue,
		float64(dst[0].LogTruncations),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PercentLogUsed,
		prometheus.GaugeValue,
		float64(dst[0].PercentLogUsed),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ReplPendingXacts,
		prometheus.GaugeValue,
		float64(dst[0].ReplPendingXacts),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ReplTransRate,
		prometheus.GaugeValue,
		float64(dst[0].ReplTransRate),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ShrinkDataMovementBytesPersec,
		prometheus.GaugeValue,
		float64(dst[0].ShrinkDataMovementBytesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.TrackedtransactionsPersec,
		prometheus.GaugeValue,
		float64(dst[0].TrackedtransactionsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.TransactionsPersec,
		prometheus.GaugeValue,
		float64(dst[0].TransactionsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.WriteTransactionsPersec,
		prometheus.GaugeValue,
		float64(dst[0].WriteTransactionsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.XTPControllerDLCLatencyPerFetch,
		prometheus.GaugeValue,
		float64(dst[0].XTPControllerDLCLatencyPerFetch),
	)

	ch <- prometheus.MustNewConstMetric(
		c.XTPControllerDLCPeakLatency,
		prometheus.GaugeValue,
		float64(dst[0].XTPControllerDLCPeakLatency),
	)

	ch <- prometheus.MustNewConstMetric(
		c.XTPControllerLogProcessedPersec,
		prometheus.GaugeValue,
		float64(dst[0].XTPControllerLogProcessedPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.XTPMemoryUsedKB,
		prometheus.GaugeValue,
		float64(dst[0].XTPMemoryUsedKB),
	)

	return nil, nil
}

type Win32_PerfRawData_MSSQLSERVER_SQLServerGeneralStatistics struct {
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

func (c *MSSQLCollector) collectGeneralStatistics(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_MSSQLSERVER_SQLServerGeneralStatistics
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	if len(dst) > 0 {
		ch <- prometheus.MustNewConstMetric(
			c.ActiveTempTables,
			prometheus.GaugeValue,
			float64(dst[0].ActiveTempTables),
		)

		ch <- prometheus.MustNewConstMetric(
			c.ConnectionResetPersec,
			prometheus.GaugeValue,
			float64(dst[0].ConnectionResetPersec),
		)

		ch <- prometheus.MustNewConstMetric(
			c.EventNotificationsDelayedDrop,
			prometheus.GaugeValue,
			float64(dst[0].EventNotificationsDelayedDrop),
		)

		ch <- prometheus.MustNewConstMetric(
			c.HTTPAuthenticatedRequests,
			prometheus.GaugeValue,
			float64(dst[0].HTTPAuthenticatedRequests),
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogicalConnections,
			prometheus.GaugeValue,
			float64(dst[0].LogicalConnections),
		)

		ch <- prometheus.MustNewConstMetric(
			c.LoginsPersec,
			prometheus.GaugeValue,
			float64(dst[0].LoginsPersec),
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogoutsPersec,
			prometheus.GaugeValue,
			float64(dst[0].LogoutsPersec),
		)

		ch <- prometheus.MustNewConstMetric(
			c.MarsDeadlocks,
			prometheus.GaugeValue,
			float64(dst[0].MarsDeadlocks),
		)

		ch <- prometheus.MustNewConstMetric(
			c.Nonatomicyieldrate,
			prometheus.GaugeValue,
			float64(dst[0].Nonatomicyieldrate),
		)

		ch <- prometheus.MustNewConstMetric(
			c.Processesblocked,
			prometheus.GaugeValue,
			float64(dst[0].Processesblocked),
		)

		ch <- prometheus.MustNewConstMetric(
			c.SOAPEmptyRequests,
			prometheus.GaugeValue,
			float64(dst[0].SOAPEmptyRequests),
		)

		ch <- prometheus.MustNewConstMetric(
			c.SOAPMethodInvocations,
			prometheus.GaugeValue,
			float64(dst[0].SOAPMethodInvocations),
		)

		ch <- prometheus.MustNewConstMetric(
			c.SOAPSessionInitiateRequests,
			prometheus.GaugeValue,
			float64(dst[0].SOAPSessionInitiateRequests),
		)

		ch <- prometheus.MustNewConstMetric(
			c.SOAPSessionTerminateRequests,
			prometheus.GaugeValue,
			float64(dst[0].SOAPSessionTerminateRequests),
		)

		ch <- prometheus.MustNewConstMetric(
			c.SOAPSQLRequests,
			prometheus.GaugeValue,
			float64(dst[0].SOAPSQLRequests),
		)

		ch <- prometheus.MustNewConstMetric(
			c.SOAPWSDLRequests,
			prometheus.GaugeValue,
			float64(dst[0].SOAPWSDLRequests),
		)

		ch <- prometheus.MustNewConstMetric(
			c.SQLTraceIOProviderLockWaits,
			prometheus.GaugeValue,
			float64(dst[0].SQLTraceIOProviderLockWaits),
		)

		ch <- prometheus.MustNewConstMetric(
			c.Tempdbrecoveryunitid,
			prometheus.GaugeValue,
			float64(dst[0].Tempdbrecoveryunitid),
		)

		ch <- prometheus.MustNewConstMetric(
			c.Tempdbrowsetid,
			prometheus.GaugeValue,
			float64(dst[0].Tempdbrowsetid),
		)

		ch <- prometheus.MustNewConstMetric(
			c.TempTablesCreationRate,
			prometheus.GaugeValue,
			float64(dst[0].TempTablesCreationRate),
		)

		ch <- prometheus.MustNewConstMetric(
			c.TempTablesForDestruction,
			prometheus.GaugeValue,
			float64(dst[0].TempTablesForDestruction),
		)

		ch <- prometheus.MustNewConstMetric(
			c.TraceEventNotificationQueue,
			prometheus.GaugeValue,
			float64(dst[0].TraceEventNotificationQueue),
		)

		ch <- prometheus.MustNewConstMetric(
			c.Transactions,
			prometheus.GaugeValue,
			float64(dst[0].Transactions),
		)

		ch <- prometheus.MustNewConstMetric(
			c.UserConnections,
			prometheus.GaugeValue,
			float64(dst[0].UserConnections),
		)
	}
	return nil, nil
}

type Win32_PerfRawData_MSSQLSERVER_SQLServerLocks struct {
	AverageWaitTimems          uint64
	LockRequestsPersec         uint64
	LockTimeoutsPersec         uint64
	LockTimeoutstimeout0Persec uint64
	LockWaitsPersec            uint64
	LockWaitTimems             uint64
	NumberofDeadlocksPersec    uint64
}

func (c *MSSQLCollector) collectLocks(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_MSSQLSERVER_SQLServerLocks
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	if len(dst) > 0 {
		ch <- prometheus.MustNewConstMetric(
			c.AverageWaitTimems,
			prometheus.GaugeValue,
			float64(dst[0].AverageWaitTimems),
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockRequestsPersec,
			prometheus.GaugeValue,
			float64(dst[0].LockRequestsPersec),
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockTimeoutsPersec,
			prometheus.GaugeValue,
			float64(dst[0].LockTimeoutsPersec),
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockTimeoutstimeout0Persec,
			prometheus.GaugeValue,
			float64(dst[0].LockTimeoutstimeout0Persec),
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockWaitsPersec,
			prometheus.GaugeValue,
			float64(dst[0].LockWaitsPersec),
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockWaitTimems,
			prometheus.GaugeValue,
			float64(dst[0].LockWaitTimems),
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberofDeadlocksPersec,
			prometheus.GaugeValue,
			float64(dst[0].NumberofDeadlocksPersec),
		)
	}

	return nil, nil
}

type Win32_PerfRawData_MSSQLSERVER_SQLServerMemoryManager struct {
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

func (c *MSSQLCollector) collectMemoryManager(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_MSSQLSERVER_SQLServerMemoryManager
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	ch <- prometheus.MustNewConstMetric(
		c.ConnectionMemoryKB,
		prometheus.GaugeValue,
		float64(dst[0].ConnectionMemoryKB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.DatabaseCacheMemoryKB,
		prometheus.GaugeValue,
		float64(dst[0].DatabaseCacheMemoryKB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.Externalbenefitofmemory,
		prometheus.GaugeValue,
		float64(dst[0].Externalbenefitofmemory),
	)

	ch <- prometheus.MustNewConstMetric(
		c.FreeMemoryKB,
		prometheus.GaugeValue,
		float64(dst[0].FreeMemoryKB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.GrantedWorkspaceMemoryKB,
		prometheus.GaugeValue,
		float64(dst[0].GrantedWorkspaceMemoryKB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LockBlocks,
		prometheus.GaugeValue,
		float64(dst[0].LockBlocks),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LockBlocksAllocated,
		prometheus.GaugeValue,
		float64(dst[0].LockBlocksAllocated),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LockMemoryKB,
		prometheus.GaugeValue,
		float64(dst[0].LockMemoryKB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LockOwnerBlocks,
		prometheus.GaugeValue,
		float64(dst[0].LockOwnerBlocks),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LockOwnerBlocksAllocated,
		prometheus.GaugeValue,
		float64(dst[0].LockOwnerBlocksAllocated),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LogPoolMemoryKB,
		prometheus.GaugeValue,
		float64(dst[0].LogPoolMemoryKB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.MaximumWorkspaceMemoryKB,
		prometheus.GaugeValue,
		float64(dst[0].MaximumWorkspaceMemoryKB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemoryGrantsOutstanding,
		prometheus.GaugeValue,
		float64(dst[0].MemoryGrantsOutstanding),
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemoryGrantsPending,
		prometheus.GaugeValue,
		float64(dst[0].MemoryGrantsPending),
	)

	ch <- prometheus.MustNewConstMetric(
		c.OptimizerMemoryKB,
		prometheus.GaugeValue,
		float64(dst[0].OptimizerMemoryKB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ReservedServerMemoryKB,
		prometheus.GaugeValue,
		float64(dst[0].ReservedServerMemoryKB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SQLCacheMemoryKB,
		prometheus.GaugeValue,
		float64(dst[0].SQLCacheMemoryKB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.StolenServerMemoryKB,
		prometheus.GaugeValue,
		float64(dst[0].StolenServerMemoryKB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.TargetServerMemoryKB,
		prometheus.GaugeValue,
		float64(dst[0].TargetServerMemoryKB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.TotalServerMemoryKB,
		prometheus.GaugeValue,
		float64(dst[0].TotalServerMemoryKB),
	)

	return nil, nil
}

// Win32_PerfRawData_MSSQLSERVER_SQLServerSQLStatistics
type Win32_PerfRawData_MSSQLSERVER_SQLServerSQLStatistics struct {
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

func (c *MSSQLCollector) collectSQLStats(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_MSSQLSERVER_SQLServerSQLStatistics
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	ch <- prometheus.MustNewConstMetric(
		c.AutoParamAttemptsPersec,
		prometheus.GaugeValue,
		float64(dst[0].AutoParamAttemptsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.BatchRequestsPersec,
		prometheus.GaugeValue,
		float64(dst[0].BatchRequestsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.FailedAutoParamsPersec,
		prometheus.GaugeValue,
		float64(dst[0].FailedAutoParamsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ForcedParameterizationsPersec,
		prometheus.GaugeValue,
		float64(dst[0].ForcedParameterizationsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.GuidedplanexecutionsPersec,
		prometheus.GaugeValue,
		float64(dst[0].GuidedplanexecutionsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.MisguidedplanexecutionsPersec,
		prometheus.GaugeValue,
		float64(dst[0].MisguidedplanexecutionsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SafeAutoParamsPersec,
		prometheus.GaugeValue,
		float64(dst[0].SafeAutoParamsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SQLAttentionrate,
		prometheus.GaugeValue,
		float64(dst[0].SQLAttentionrate),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SQLCompilationsPersec,
		prometheus.GaugeValue,
		float64(dst[0].SQLCompilationsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SQLReCompilationsPersec,
		prometheus.GaugeValue,
		float64(dst[0].SQLReCompilationsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.UnsafeAutoParamsPersec,
		prometheus.GaugeValue,
		float64(dst[0].UnsafeAutoParamsPersec),
	)

	return nil, nil
}

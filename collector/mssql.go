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

	instanceNames, err := k.ReadValueNames(0)
	if err != nil {
		log.Warn("Can't ReadSubKeyNames %#v", err)
		return sqlInstances
	}

	for _, instanceName := range instanceNames {
		if instanceVersion, _, err := k.GetStringValue(instanceName); err == nil {
			sqlInstances[instanceName] = instanceVersion
		}
	}

	log.Debugf("Detected MSSQL Instances: %#v\n", sqlInstances)

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

	UnsafeAutoParamsPersec *prometheus.Desc

	sqlInstances sqlInstancesType
}

// NewMSSQLCollector ...
func NewMSSQLCollector() (Collector, error) {

	const subsystem = "mssql"
	return &MSSQLCollector{

		// Win32_PerfRawData_{instance}_SQLServerAvailabilityReplica
		BytesReceivedfromReplicaPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "availreplica_bytes_received_from_replica"),
			"(AvailabilityReplica.BytesReceivedfromReplicaPersec)",
			[]string{"instance", "replica"},
			nil,
		),
		BytesSenttoReplicaPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "availreplica_bytes_sent_to_replica"),
			"(AvailabilityReplica.BytesSenttoReplicaPersec)",
			[]string{"instance", "replica"},
			nil,
		),
		BytesSenttoTransportPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "availreplica_bytes_sent_to_transport"),
			"(AvailabilityReplica.BytesSenttoTransportPersec)",
			[]string{"instance", "replica"},
			nil,
		),
		FlowControlPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "availreplica_flow_control"),
			"(AvailabilityReplica.FlowControlPersec)",
			[]string{"instance", "replica"},
			nil,
		),
		FlowControlTimemsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "availreplica_flow_control_time"),
			"(AvailabilityReplica.FlowControlTimemsPersec)",
			[]string{"instance", "replica"},
			nil,
		),
		ReceivesfromReplicaPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "availreplica_receives_from_replica"),
			"(AvailabilityReplica.ReceivesfromReplicaPersec)",
			[]string{"instance", "replica"},
			nil,
		),
		ResentMessagesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "availreplica_resent_messages"),
			"(AvailabilityReplica.ResentMessagesPersec)",
			[]string{"instance", "replica"},
			nil,
		),
		SendstoReplicaPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "availreplica_sends_to_replica"),
			"(AvailabilityReplica.SendstoReplicaPersec)",
			[]string{"instance", "replica"},
			nil,
		),
		SendstoTransportPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "availreplica_sends_to_transport"),
			"(AvailabilityReplica.SendstoTransportPersec)",
			[]string{"instance", "replica"},
			nil,
		),

		// Win32_PerfRawData_{instance}_SQLServerBufferManager
		BackgroundwriterpagesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_background_writer_pages"),
			"(BufferManager.BackgroundwriterpagesPersec)",
			[]string{"instance"},
			nil,
		),
		Buffercachehitratio: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_buffer_cache_hit_ratio"),
			"(BufferManager.Buffercachehitratio)",
			[]string{"instance"},
			nil,
		),
		CheckpointpagesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_checkpoint_pages"),
			"(BufferManager.CheckpointpagesPersec)",
			[]string{"instance"},
			nil,
		),
		Databasepages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_database_pages"),
			"(BufferManager.Databasepages)",
			[]string{"instance"},
			nil,
		),
		Extensionallocatedpages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_extension_allocated_pages"),
			"(BufferManager.Extensionallocatedpages)",
			[]string{"instance"},
			nil,
		),
		Extensionfreepages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_extension_free_pages"),
			"(BufferManager.Extensionfreepages)",
			[]string{"instance"},
			nil,
		),
		Extensioninuseaspercentage: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_extension_in_use_as_percentage"),
			"(BufferManager.Extensioninuseaspercentage)",
			[]string{"instance"},
			nil,
		),
		ExtensionoutstandingIOcounter: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_extension_outstanding_io_counter"),
			"(BufferManager.ExtensionoutstandingIOcounter)",
			[]string{"instance"},
			nil,
		),
		ExtensionpageevictionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_extension_page_evictions"),
			"(BufferManager.ExtensionpageevictionsPersec)",
			[]string{"instance"},
			nil,
		),
		ExtensionpagereadsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_extension_page_reads"),
			"(BufferManager.ExtensionpagereadsPersec)",
			[]string{"instance"},
			nil,
		),
		Extensionpageunreferencedtime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_extension_page_unreferenced_time"),
			"(BufferManager.Extensionpageunreferencedtime)",
			[]string{"instance"},
			nil,
		),
		ExtensionpagewritesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_extension_page_writes"),
			"(BufferManager.ExtensionpagewritesPersec)",
			[]string{"instance"},
			nil,
		),
		FreeliststallsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_free_list_stalls"),
			"(BufferManager.FreeliststallsPersec)",
			[]string{"instance"},
			nil,
		),
		IntegralControllerSlope: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_integral_controller_slope"),
			"(BufferManager.IntegralControllerSlope)",
			[]string{"instance"},
			nil,
		),
		LazywritesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_lazywrites"),
			"(BufferManager.LazywritesPersec)",
			[]string{"instance"},
			nil,
		),
		Pagelifeexpectancy: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_page_life_expectancy"),
			"(BufferManager.Pagelifeexpectancy)",
			[]string{"instance"},
			nil,
		),
		PagelookupsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_page_lookups"),
			"(BufferManager.PagelookupsPersec)",
			[]string{"instance"},
			nil,
		),
		PagereadsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_page_reads"),
			"(BufferManager.PagereadsPersec)",
			[]string{"instance"},
			nil,
		),
		PagewritesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_page_writes"),
			"(BufferManager.PagewritesPersec)",
			[]string{"instance"},
			nil,
		),
		ReadaheadpagesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_read_ahead_pages"),
			"(BufferManager.ReadaheadpagesPersec)",
			[]string{"instance"},
			nil,
		),
		ReadaheadtimePersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_read_ahead_time"),
			"(BufferManager.ReadaheadtimePersec)",
			[]string{"instance"},
			nil,
		),
		Targetpages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_target_pages"),
			"(BufferManager.Targetpages)",
			[]string{"instance"},
			nil,
		),

		// Win32_PerfRawData_{instance}_SQLServerDatabaseReplica
		DatabaseFlowControlDelay: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_database_flow_control_delay"),
			"(DatabaseReplica.DatabaseFlowControlDelay)",
			[]string{"instance", "replica"},
			nil,
		),
		DatabaseFlowControlsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_database_flow_controls"),
			"(DatabaseReplica.DatabaseFlowControlsPersec)",
			[]string{"instance", "replica"},
			nil,
		),
		FileBytesReceivedPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_file_bytes_received"),
			"(DatabaseReplica.FileBytesReceivedPersec)",
			[]string{"instance", "replica"},
			nil,
		),
		GroupCommitsPerSec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_group_commits"),
			"(DatabaseReplica.GroupCommitsPerSec)",
			[]string{"instance", "replica"},
			nil,
		),
		GroupCommitTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_group_commit_time"),
			"(DatabaseReplica.GroupCommitTime)",
			[]string{"instance", "replica"},
			nil,
		),
		LogApplyPendingQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_log_apply_pending_queue"),
			"(DatabaseReplica.LogApplyPendingQueue)",
			[]string{"instance", "replica"},
			nil,
		),
		LogApplyReadyQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_log_apply_ready_queue"),
			"(DatabaseReplica.LogApplyReadyQueue)",
			[]string{"instance", "replica"},
			nil,
		),
		LogBytesCompressedPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_log_bytes_compressed"),
			"(DatabaseReplica.LogBytesCompressedPersec)",
			[]string{"instance", "replica"},
			nil,
		),
		LogBytesDecompressedPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_log_bytes_decompressed"),
			"(DatabaseReplica.LogBytesDecompressedPersec)",
			[]string{"instance", "replica"},
			nil,
		),
		LogBytesReceivedPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_log_bytes_received"),
			"(DatabaseReplica.LogBytesReceivedPersec)",
			[]string{"instance", "replica"},
			nil,
		),
		LogCompressionCachehitsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_log_compression_cachehits"),
			"(DatabaseReplica.LogCompressionCachehitsPersec)",
			[]string{"instance", "replica"},
			nil,
		),
		LogCompressionCachemissesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_log_compression_cachemisses"),
			"(DatabaseReplica.LogCompressionCachemissesPersec)",
			[]string{"instance", "replica"},
			nil,
		),
		LogCompressionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_log_compressions"),
			"(DatabaseReplica.LogCompressionsPersec)",
			[]string{"instance", "replica"},
			nil,
		),
		LogDecompressionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_log_decompressions"),
			"(DatabaseReplica.LogDecompressionsPersec)",
			[]string{"instance", "replica"},
			nil,
		),
		Logremainingforundo: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_log_remaining_for_undo"),
			"(DatabaseReplica.Logremainingforundo)",
			[]string{"instance", "replica"},
			nil,
		),
		LogSendQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_log_send_queue"),
			"(DatabaseReplica.LogSendQueue)",
			[]string{"instance", "replica"},
			nil,
		),
		MirroredWriteTransactionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_mirrored_write_transactions"),
			"(DatabaseReplica.MirroredWriteTransactionsPersec)",
			[]string{"instance", "replica"},
			nil,
		),
		RecoveryQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_recovery_queue"),
			"(DatabaseReplica.RecoveryQueue)",
			[]string{"instance", "replica"},
			nil,
		),
		RedoblockedPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_redoblocked"),
			"(DatabaseReplica.RedoblockedPersec)",
			[]string{"instance", "replica"},
			nil,
		),
		RedoBytesRemaining: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_redo_bytes_remaining"),
			"(DatabaseReplica.RedoBytesRemaining)",
			[]string{"instance", "replica"},
			nil,
		),
		RedoneBytesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_redone_bytes"),
			"(DatabaseReplica.RedoneBytesPersec)",
			[]string{"instance", "replica"},
			nil,
		),
		RedonesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_redones"),
			"(DatabaseReplica.RedonesPersec)",
			[]string{"instance", "replica"},
			nil,
		),
		TotalLogrequiringundo: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_total_log_requiring_undo"),
			"(DatabaseReplica.TotalLogrequiringundo)",
			[]string{"instance", "replica"},
			nil,
		),
		TransactionDelay: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_transaction_delay"),
			"(DatabaseReplica.TransactionDelay)",
			[]string{"instance", "replica"},
			nil,
		),

		// Win32_PerfRawData_{instance}_SQLServerDatabases
		ActiveTransactions: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_active_transactions"),
			"(Databases.ActiveTransactions)",
			[]string{"instance", "database"},
			nil,
		),
		AvgDistFromEOLLPRequest: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_avg_dist_from_eollp_request"),
			"(Databases.AvgDistFromEOLLPRequest)",
			[]string{"instance", "database"},
			nil,
		),
		BackupPerRestoreThroughputPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_backup_per_restore_throughput"),
			"(Databases.BackupPerRestoreThroughputPersec)",
			[]string{"instance", "database"},
			nil,
		),
		BulkCopyRowsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_bulk_copy_rows"),
			"(Databases.BulkCopyRowsPersec)",
			[]string{"instance", "database"},
			nil,
		),
		BulkCopyThroughputPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_bulk_copy_throughput"),
			"(Databases.BulkCopyThroughputPersec)",
			[]string{"instance", "database"},
			nil,
		),
		Committableentries: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_commit_table_entries"),
			"(Databases.Committableentries)",
			[]string{"instance", "database"},
			nil,
		),
		DataFilesSizeKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_data_files_size"),
			"(Databases.DataFilesSizeKB)",
			[]string{"instance", "database"},
			nil,
		),
		DBCCLogicalScanBytesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_dbcc_logical_scan_bytes"),
			"(Databases.DBCCLogicalScanBytesPersec)",
			[]string{"instance", "database"},
			nil,
		),
		GroupCommitTimePersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_group_commit_time"),
			"(Databases.GroupCommitTimePersec)",
			[]string{"instance", "database"},
			nil,
		),
		LogBytesFlushedPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_bytes_flushed"),
			"(Databases.LogBytesFlushedPersec)",
			[]string{"instance", "database"},
			nil,
		),
		LogCacheHitRatio: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_cache_hit_ratio"),
			"(Databases.LogCacheHitRatio)",
			[]string{"instance", "database"},
			nil,
		),
		LogCacheReadsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_cache_reads"),
			"(Databases.LogCacheReadsPersec)",
			[]string{"instance", "database"},
			nil,
		),
		LogFilesSizeKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_files_size"),
			"(Databases.LogFilesSizeKB)",
			[]string{"instance", "database"},
			nil,
		),
		LogFilesUsedSizeKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_files_used_size"),
			"(Databases.LogFilesUsedSizeKB)",
			[]string{"instance", "database"},
			nil,
		),
		LogFlushesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_flushes"),
			"(Databases.LogFlushesPersec)",
			[]string{"instance", "database"},
			nil,
		),
		LogFlushWaitsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_flush_waits"),
			"(Databases.LogFlushWaitsPersec)",
			[]string{"instance", "database"},
			nil,
		),
		LogFlushWaitTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_flush_wait_time"),
			"(Databases.LogFlushWaitTime)",
			[]string{"instance", "database"},
			nil,
		),
		LogFlushWriteTimems: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_flush_write_time"),
			"(Databases.LogFlushWriteTimems)",
			[]string{"instance", "database"},
			nil,
		),
		LogGrowths: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_growths"),
			"(Databases.LogGrowths)",
			[]string{"instance", "database"},
			nil,
		),
		LogPoolCacheMissesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_pool_cache_misses"),
			"(Databases.LogPoolCacheMissesPersec)",
			[]string{"instance", "database"},
			nil,
		),
		LogPoolDiskReadsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_pool_disk_reads"),
			"(Databases.LogPoolDiskReadsPersec)",
			[]string{"instance", "database"},
			nil,
		),
		LogPoolHashDeletesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_pool_hash_deletes"),
			"(Databases.LogPoolHashDeletesPersec)",
			[]string{"instance", "database"},
			nil,
		),
		LogPoolHashInsertsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_pool_hash_inserts"),
			"(Databases.LogPoolHashInsertsPersec)",
			[]string{"instance", "database"},
			nil,
		),
		LogPoolInvalidHashEntryPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_pool_invalid_hash_entry"),
			"(Databases.LogPoolInvalidHashEntryPersec)",
			[]string{"instance", "database"},
			nil,
		),
		LogPoolLogScanPushesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_pool_log_scan_pushes"),
			"(Databases.LogPoolLogScanPushesPersec)",
			[]string{"instance", "database"},
			nil,
		),
		LogPoolLogWriterPushesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_pool_log_writer_pushes"),
			"(Databases.LogPoolLogWriterPushesPersec)",
			[]string{"instance", "database"},
			nil,
		),
		LogPoolPushEmptyFreePoolPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_pool_push_empty_free_pool"),
			"(Databases.LogPoolPushEmptyFreePoolPersec)",
			[]string{"instance", "database"},
			nil,
		),
		LogPoolPushLowMemoryPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_pool_push_low_memory"),
			"(Databases.LogPoolPushLowMemoryPersec)",
			[]string{"instance", "database"},
			nil,
		),
		LogPoolPushNoFreeBufferPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_pool_push_no_free_buffer"),
			"(Databases.LogPoolPushNoFreeBufferPersec)",
			[]string{"instance", "database"},
			nil,
		),
		LogPoolReqBehindTruncPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_pool_req_behind_trunc"),
			"(Databases.LogPoolReqBehindTruncPersec)",
			[]string{"instance", "database"},
			nil,
		),
		LogPoolRequestsOldVLFPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_pool_requests_old_vlf"),
			"(Databases.LogPoolRequestsOldVLFPersec)",
			[]string{"instance", "database"},
			nil,
		),
		LogPoolRequestsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_pool_requests"),
			"(Databases.LogPoolRequestsPersec)",
			[]string{"instance", "database"},
			nil,
		),
		LogPoolTotalActiveLogSize: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_pool_total_active_log_size"),
			"(Databases.LogPoolTotalActiveLogSize)",
			[]string{"instance", "database"},
			nil,
		),
		LogPoolTotalSharedPoolSize: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_pool_total_shared_pool_size"),
			"(Databases.LogPoolTotalSharedPoolSize)",
			[]string{"instance", "database"},
			nil,
		),
		LogShrinks: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_shrinks"),
			"(Databases.LogShrinks)",
			[]string{"instance", "database"},
			nil,
		),
		LogTruncations: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_truncations"),
			"(Databases.LogTruncations)",
			[]string{"instance", "database"},
			nil,
		),
		PercentLogUsed: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_percent_log_used"),
			"(Databases.PercentLogUsed)",
			[]string{"instance", "database"},
			nil,
		),
		ReplPendingXacts: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_repl_pending_xacts"),
			"(Databases.ReplPendingXacts)",
			[]string{"instance", "database"},
			nil,
		),
		ReplTransRate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_repl_trans_rate"),
			"(Databases.ReplTransRate)",
			[]string{"instance", "database"},
			nil,
		),
		ShrinkDataMovementBytesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_shrink_data_movement_bytes"),
			"(Databases.ShrinkDataMovementBytesPersec)",
			[]string{"instance", "database"},
			nil,
		),
		TrackedtransactionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_tracked_transactions"),
			"(Databases.TrackedtransactionsPersec)",
			[]string{"instance", "database"},
			nil,
		),
		TransactionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_transactions"),
			"(Databases.TransactionsPersec)",
			[]string{"instance", "database"},
			nil,
		),
		WriteTransactionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_write_transactions"),
			"(Databases.WriteTransactionsPersec)",
			[]string{"instance", "database"},
			nil,
		),
		XTPControllerDLCLatencyPerFetch: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_xtp_controller_dlc_latency_per_fetch"),
			"(Databases.XTPControllerDLCLatencyPerFetch)",
			[]string{"instance", "database"},
			nil,
		),
		XTPControllerDLCPeakLatency: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_xtp_controller_dlc_peak_latency"),
			"(Databases.XTPControllerDLCPeakLatency)",
			[]string{"instance", "database"},
			nil,
		),
		XTPControllerLogProcessedPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_xtp_controller_log_processed"),
			"(Databases.XTPControllerLogProcessedPersec)",
			[]string{"instance", "database"},
			nil,
		),
		XTPMemoryUsedKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_xtp_memory_used"),
			"(Databases.XTPMemoryUsedKB)",
			[]string{"instance", "database"},
			nil,
		),

		// Win32_PerfRawData_{instance}_SQLServerGeneralStatistics
		ActiveTempTables: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_active_temp_tables"),
			"(GeneralStatistics.ActiveTempTables)",
			[]string{"instance"},
			nil,
		),
		ConnectionResetPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_connection_reset"),
			"(GeneralStatistics.ConnectionResetPersec)",
			[]string{"instance"},
			nil,
		),
		EventNotificationsDelayedDrop: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_event_notifications_delayed_drop"),
			"(GeneralStatistics.EventNotificationsDelayedDrop)",
			[]string{"instance"},
			nil,
		),
		HTTPAuthenticatedRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_http_authenticated_requests"),
			"(GeneralStatistics.HTTPAuthenticatedRequests)",
			[]string{"instance"},
			nil,
		),
		LogicalConnections: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_logical_connections"),
			"(GeneralStatistics.LogicalConnections)",
			[]string{"instance"},
			nil,
		),
		LoginsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_logins"),
			"(GeneralStatistics.LoginsPersec)",
			[]string{"instance"},
			nil,
		),
		LogoutsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_logouts"),
			"(GeneralStatistics.LogoutsPersec)",
			[]string{"instance"},
			nil,
		),
		MarsDeadlocks: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_mars_deadlocks"),
			"(GeneralStatistics.MarsDeadlocks)",
			[]string{"instance"},
			nil,
		),
		Nonatomicyieldrate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_non_atomic_yield_rate"),
			"(GeneralStatistics.Nonatomicyieldrate)",
			[]string{"instance"},
			nil,
		),
		Processesblocked: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_processes_blocked"),
			"(GeneralStatistics.Processesblocked)",
			[]string{"instance"},
			nil,
		),
		SOAPEmptyRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_soap_empty_requests"),
			"(GeneralStatistics.SOAPEmptyRequests)",
			[]string{"instance"},
			nil,
		),
		SOAPMethodInvocations: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_soap_method_invocations"),
			"(GeneralStatistics.SOAPMethodInvocations)",
			[]string{"instance"},
			nil,
		),
		SOAPSessionInitiateRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_soap_session_initiate_requests"),
			"(GeneralStatistics.SOAPSessionInitiateRequests)",
			[]string{"instance"},
			nil,
		),
		SOAPSessionTerminateRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_soap_session_terminate_requests"),
			"(GeneralStatistics.SOAPSessionTerminateRequests)",
			[]string{"instance"},
			nil,
		),
		SOAPSQLRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_soapsql_requests"),
			"(GeneralStatistics.SOAPSQLRequests)",
			[]string{"instance"},
			nil,
		),
		SOAPWSDLRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_soapwsdl_requests"),
			"(GeneralStatistics.SOAPWSDLRequests)",
			[]string{"instance"},
			nil,
		),
		SQLTraceIOProviderLockWaits: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_sql_trace_io_provider_lock_waits"),
			"(GeneralStatistics.SQLTraceIOProviderLockWaits)",
			[]string{"instance"},
			nil,
		),
		Tempdbrecoveryunitid: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_tempdb_recovery_unit_id"),
			"(GeneralStatistics.Tempdbrecoveryunitid)",
			[]string{"instance"},
			nil,
		),
		Tempdbrowsetid: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_tempdb_rowsetid"),
			"(GeneralStatistics.Tempdbrowsetid)",
			[]string{"instance"},
			nil,
		),
		TempTablesCreationRate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_temp_tables_creation_rate"),
			"(GeneralStatistics.TempTablesCreationRate)",
			[]string{"instance"},
			nil,
		),
		TempTablesForDestruction: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_temp_tables_for_destruction"),
			"(GeneralStatistics.TempTablesForDestruction)",
			[]string{"instance"},
			nil,
		),
		TraceEventNotificationQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_trace_event_notification_queue"),
			"(GeneralStatistics.TraceEventNotificationQueue)",
			[]string{"instance"},
			nil,
		),
		Transactions: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_transactions"),
			"(GeneralStatistics.Transactions)",
			[]string{"instance"},
			nil,
		),
		UserConnections: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_user_connections"),
			"(GeneralStatistics.UserConnections)",
			[]string{"instance"},
			nil,
		),

		// Win32_PerfRawData_{instance}_SQLServerLocks
		AverageWaitTimems: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "locks_average_wait_time"),
			"(Locks.AverageWaitTimems)",
			[]string{"instance", "resource"},
			nil,
		),
		LockRequestsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "locks_lock_requests"),
			"(Locks.LockRequestsPersec)",
			[]string{"instance", "resource"},
			nil,
		),
		LockTimeoutsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "locks_lock_timeouts"),
			"(Locks.LockTimeoutsPersec)",
			[]string{"instance", "resource"},
			nil,
		),
		LockTimeoutstimeout0Persec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "locks_lock_timeouts_timeout0"),
			"(Locks.LockTimeoutstimeout0Persec)",
			[]string{"instance", "resource"},
			nil,
		),
		LockWaitsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "locks_lock_waits"),
			"(Locks.LockWaitsPersec)",
			[]string{"instance", "resource"},
			nil,
		),
		LockWaitTimems: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "locks_lock_wait_time"),
			"(Locks.LockWaitTimems)",
			[]string{"instance", "resource"},
			nil,
		),
		NumberofDeadlocksPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "locks_number_of_deadlocks"),
			"(Locks.NumberofDeadlocksPersec)",
			[]string{"instance", "resource"},
			nil,
		),

		// Win32_PerfRawData_{instance}_SQLServerMemoryManager
		ConnectionMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_connection_memory"),
			"(MemoryManager.ConnectionMemoryKB)",
			[]string{"instance"},
			nil,
		),
		DatabaseCacheMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_database_cache_memory"),
			"(MemoryManager.DatabaseCacheMemoryKB)",
			[]string{"instance"},
			nil,
		),
		Externalbenefitofmemory: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_external_benefit_of_memory"),
			"(MemoryManager.Externalbenefitofmemory)",
			[]string{"instance"},
			nil,
		),
		FreeMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_free_memory"),
			"(MemoryManager.FreeMemoryKB)",
			[]string{"instance"},
			nil,
		),
		GrantedWorkspaceMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_granted_workspace_memory"),
			"(MemoryManager.GrantedWorkspaceMemoryKB)",
			[]string{"instance"},
			nil,
		),
		LockBlocks: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_lock_blocks"),
			"(MemoryManager.LockBlocks)",
			[]string{"instance"},
			nil,
		),
		LockBlocksAllocated: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_lock_blocks_allocated"),
			"(MemoryManager.LockBlocksAllocated)",
			[]string{"instance"},
			nil,
		),
		LockMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_lock_memory"),
			"(MemoryManager.LockMemoryKB)",
			[]string{"instance"},
			nil,
		),
		LockOwnerBlocks: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_lock_owner_blocks"),
			"(MemoryManager.LockOwnerBlocks)",
			[]string{"instance"},
			nil,
		),
		LockOwnerBlocksAllocated: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_lock_owner_blocks_allocated"),
			"(MemoryManager.LockOwnerBlocksAllocated)",
			[]string{"instance"},
			nil,
		),
		LogPoolMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_log_pool_memory"),
			"(MemoryManager.LogPoolMemoryKB)",
			[]string{"instance"},
			nil,
		),
		MaximumWorkspaceMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_maximum_workspace_memory"),
			"(MemoryManager.MaximumWorkspaceMemoryKB)",
			[]string{"instance"},
			nil,
		),
		MemoryGrantsOutstanding: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_memory_grants_outstanding"),
			"(MemoryManager.MemoryGrantsOutstanding)",
			[]string{"instance"},
			nil,
		),
		MemoryGrantsPending: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_memory_grants_pending"),
			"(MemoryManager.MemoryGrantsPending)",
			[]string{"instance"},
			nil,
		),
		OptimizerMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_optimizer_memory"),
			"(MemoryManager.OptimizerMemoryKB)",
			[]string{"instance"},
			nil,
		),
		ReservedServerMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_reserved_server_memory"),
			"(MemoryManager.ReservedServerMemoryKB)",
			[]string{"instance"},
			nil,
		),
		SQLCacheMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_sql_cache_memory"),
			"(MemoryManager.SQLCacheMemoryKB)",
			[]string{"instance"},
			nil,
		),
		StolenServerMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_stolen_server_memory"),
			"(MemoryManager.StolenServerMemoryKB)",
			[]string{"instance"},
			nil,
		),
		TargetServerMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_target_server_memory"),
			"(MemoryManager.TargetServerMemoryKB)",
			[]string{"instance"},
			nil,
		),
		TotalServerMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_total_server_memory"),
			"(MemoryManager.TotalServerMemoryKB)",
			[]string{"instance"},
			nil,
		),

		// Win32_PerfRawData_{instance}_SQLServerSQLStatistics
		AutoParamAttemptsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sqlstats_auto_param_attempts"),
			"(SQLStatistics.AutoParamAttemptsPersec)",
			[]string{"instance"},
			nil,
		),
		BatchRequestsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sqlstats_batch_requests"),
			"(SQLStatistics.BatchRequestsPersec)",
			[]string{"instance"},
			nil,
		),
		FailedAutoParamsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sqlstats_failed_auto_params"),
			"(SQLStatistics.FailedAutoParamsPersec)",
			[]string{"instance"},
			nil,
		),
		ForcedParameterizationsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sqlstats_forced_parameterizations"),
			"(SQLStatistics.ForcedParameterizationsPersec)",
			[]string{"instance"},
			nil,
		),
		GuidedplanexecutionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sqlstats_guided_plan_executions"),
			"(SQLStatistics.GuidedplanexecutionsPersec)",
			[]string{"instance"},
			nil,
		),
		MisguidedplanexecutionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sqlstats_misguided_plan_executions"),
			"(SQLStatistics.MisguidedplanexecutionsPersec)",
			[]string{"instance"},
			nil,
		),
		SafeAutoParamsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sqlstats_safe_auto_params"),
			"(SQLStatistics.SafeAutoParamsPersec)",
			[]string{"instance"},
			nil,
		),
		SQLAttentionrate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sqlstats_sql_attention_rate"),
			"(SQLStatistics.SQLAttentionrate)",
			[]string{"instance"},
			nil,
		),
		SQLCompilationsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sqlstats_sql_compilations"),
			"(SQLStatistics.SQLCompilationsPersec)",
			[]string{"instance"},
			nil,
		),
		SQLReCompilationsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sqlstats_sql_recompilations"),
			"(SQLStatistics.SQLReCompilationsPersec)",
			[]string{"instance"},
			nil,
		),
		UnsafeAutoParamsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sqlstats_unsafe_auto_params"),
			"(SQLStatistics.UnsafeAutoParamsPersec)",
			[]string{"instance"},
			nil,
		),

		sqlInstances: getMSSQLInstances(),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *MSSQLCollector) Collect(ch chan<- prometheus.Metric) error {
	for instance := range c.sqlInstances {
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
	Name                           string
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
	q := queryAllForClassWhere(&dst, class, `Name <> '_Total'`)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	for _, v := range dst {
		replicaName := v.Name

		ch <- prometheus.MustNewConstMetric(
			c.BytesReceivedfromReplicaPersec,
			prometheus.CounterValue,
			float64(v.BytesReceivedfromReplicaPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BytesSenttoReplicaPersec,
			prometheus.CounterValue,
			float64(v.BytesSenttoReplicaPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BytesSenttoTransportPersec,
			prometheus.CounterValue,
			float64(v.BytesSenttoTransportPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FlowControlPersec,
			prometheus.CounterValue,
			float64(v.FlowControlPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FlowControlTimemsPersec,
			prometheus.CounterValue,
			float64(v.FlowControlTimemsPersec)/1000.0,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReceivesfromReplicaPersec,
			prometheus.CounterValue,
			float64(v.ReceivesfromReplicaPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ResentMessagesPersec,
			prometheus.CounterValue,
			float64(v.ResentMessagesPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SendstoReplicaPersec,
			prometheus.CounterValue,
			float64(v.SendstoReplicaPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SendstoTransportPersec,
			prometheus.CounterValue,
			float64(v.SendstoTransportPersec),
			sqlInstance, replicaName,
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
		v := dst[0]

		ch <- prometheus.MustNewConstMetric(
			c.BackgroundwriterpagesPersec,
			prometheus.CounterValue,
			float64(v.BackgroundwriterpagesPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Buffercachehitratio,
			prometheus.GaugeValue,
			float64(v.Buffercachehitratio),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.CheckpointpagesPersec,
			prometheus.CounterValue,
			float64(v.CheckpointpagesPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Databasepages,
			prometheus.GaugeValue,
			float64(v.Databasepages),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Extensionallocatedpages,
			prometheus.GaugeValue,
			float64(v.Extensionallocatedpages),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Extensionfreepages,
			prometheus.GaugeValue,
			float64(v.Extensionfreepages),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Extensioninuseaspercentage,
			prometheus.GaugeValue,
			float64(v.Extensioninuseaspercentage),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ExtensionoutstandingIOcounter,
			prometheus.GaugeValue,
			float64(v.ExtensionoutstandingIOcounter),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ExtensionpageevictionsPersec,
			prometheus.CounterValue,
			float64(v.ExtensionpageevictionsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ExtensionpagereadsPersec,
			prometheus.CounterValue,
			float64(v.ExtensionpagereadsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Extensionpageunreferencedtime,
			prometheus.GaugeValue,
			float64(v.Extensionpageunreferencedtime),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ExtensionpagewritesPersec,
			prometheus.CounterValue,
			float64(v.ExtensionpagewritesPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FreeliststallsPersec,
			prometheus.CounterValue,
			float64(v.FreeliststallsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.IntegralControllerSlope,
			prometheus.GaugeValue,
			float64(v.IntegralControllerSlope),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LazywritesPersec,
			prometheus.CounterValue,
			float64(v.LazywritesPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Pagelifeexpectancy,
			prometheus.GaugeValue,
			float64(v.Pagelifeexpectancy),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PagelookupsPersec,
			prometheus.CounterValue,
			float64(v.PagelookupsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PagereadsPersec,
			prometheus.CounterValue,
			float64(v.PagereadsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PagewritesPersec,
			prometheus.CounterValue,
			float64(v.PagewritesPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReadaheadpagesPersec,
			prometheus.CounterValue,
			float64(v.ReadaheadpagesPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReadaheadtimePersec,
			prometheus.CounterValue,
			float64(v.ReadaheadtimePersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Targetpages,
			prometheus.GaugeValue,
			float64(v.Targetpages),
			sqlInstance,
		)
	}

	return nil, nil
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

func (c *MSSQLCollector) collectDatabaseReplica(ch chan<- prometheus.Metric, sqlInstance string) (*prometheus.Desc, error) {
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
			float64(v.TransactionDelay),
			sqlInstance, replicaName,
		)
	}

	return nil, nil
}

type win32PerfRawDataSQLServerDatabases struct {
	Name                             string
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
			c.AvgDistFromEOLLPRequest,
			prometheus.GaugeValue,
			float64(v.AvgDistFromEOLLPRequest),
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
			float64(v.BulkCopyThroughputPersec),
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
			float64(v.GroupCommitTimePersec),
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
			float64(v.LogFlushWaitTime),
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
			prometheus.GaugeValue,
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
			float64(v.XTPControllerDLCPeakLatency),
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
		v := dst[0]
		ch <- prometheus.MustNewConstMetric(
			c.ActiveTempTables,
			prometheus.GaugeValue,
			float64(v.ActiveTempTables),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ConnectionResetPersec,
			prometheus.CounterValue,
			float64(v.ConnectionResetPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.EventNotificationsDelayedDrop,
			prometheus.GaugeValue,
			float64(v.EventNotificationsDelayedDrop),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.HTTPAuthenticatedRequests,
			prometheus.GaugeValue,
			float64(v.HTTPAuthenticatedRequests),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogicalConnections,
			prometheus.GaugeValue,
			float64(v.LogicalConnections),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LoginsPersec,
			prometheus.CounterValue,
			float64(v.LoginsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogoutsPersec,
			prometheus.CounterValue,
			float64(v.LogoutsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MarsDeadlocks,
			prometheus.GaugeValue,
			float64(v.MarsDeadlocks),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Nonatomicyieldrate,
			prometheus.GaugeValue,
			float64(v.Nonatomicyieldrate),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Processesblocked,
			prometheus.GaugeValue,
			float64(v.Processesblocked),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SOAPEmptyRequests,
			prometheus.GaugeValue,
			float64(v.SOAPEmptyRequests),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SOAPMethodInvocations,
			prometheus.GaugeValue,
			float64(v.SOAPMethodInvocations),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SOAPSessionInitiateRequests,
			prometheus.GaugeValue,
			float64(v.SOAPSessionInitiateRequests),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SOAPSessionTerminateRequests,
			prometheus.GaugeValue,
			float64(v.SOAPSessionTerminateRequests),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SOAPSQLRequests,
			prometheus.GaugeValue,
			float64(v.SOAPSQLRequests),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SOAPWSDLRequests,
			prometheus.GaugeValue,
			float64(v.SOAPWSDLRequests),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SQLTraceIOProviderLockWaits,
			prometheus.GaugeValue,
			float64(v.SQLTraceIOProviderLockWaits),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Tempdbrecoveryunitid,
			prometheus.GaugeValue,
			float64(v.Tempdbrecoveryunitid),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Tempdbrowsetid,
			prometheus.GaugeValue,
			float64(v.Tempdbrowsetid),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TempTablesCreationRate,
			prometheus.GaugeValue,
			float64(v.TempTablesCreationRate),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TempTablesForDestruction,
			prometheus.GaugeValue,
			float64(v.TempTablesForDestruction),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TraceEventNotificationQueue,
			prometheus.GaugeValue,
			float64(v.TraceEventNotificationQueue),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Transactions,
			prometheus.GaugeValue,
			float64(v.Transactions),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.UserConnections,
			prometheus.GaugeValue,
			float64(v.UserConnections),
			sqlInstance,
		)
	}

	return nil, nil
}

type win32PerfRawDataSQLServerLocks struct {
	Name                       string
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
	q := queryAllForClassWhere(&dst, class, `Name <> '_Total'`)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	for _, v := range dst {
		lockResourceName := v.Name

		ch <- prometheus.MustNewConstMetric(
			c.AverageWaitTimems,
			prometheus.GaugeValue,
			float64(v.AverageWaitTimems)/1000.0,
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockRequestsPersec,
			prometheus.CounterValue,
			float64(v.LockRequestsPersec),
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockTimeoutsPersec,
			prometheus.CounterValue,
			float64(v.LockTimeoutsPersec),
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockTimeoutstimeout0Persec,
			prometheus.CounterValue,
			float64(v.LockTimeoutstimeout0Persec),
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockWaitsPersec,
			prometheus.CounterValue,
			float64(v.LockWaitsPersec),
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockWaitTimems,
			prometheus.GaugeValue,
			float64(v.LockWaitTimems)/1000.0,
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberofDeadlocksPersec,
			prometheus.CounterValue,
			float64(v.NumberofDeadlocksPersec),
			sqlInstance, lockResourceName,
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
		v := dst[0]

		ch <- prometheus.MustNewConstMetric(
			c.ConnectionMemoryKB,
			prometheus.GaugeValue,
			float64(v.ConnectionMemoryKB*1024),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabaseCacheMemoryKB,
			prometheus.GaugeValue,
			float64(v.DatabaseCacheMemoryKB*1024),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Externalbenefitofmemory,
			prometheus.GaugeValue,
			float64(v.Externalbenefitofmemory),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FreeMemoryKB,
			prometheus.GaugeValue,
			float64(v.FreeMemoryKB*1024),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GrantedWorkspaceMemoryKB,
			prometheus.GaugeValue,
			float64(v.GrantedWorkspaceMemoryKB*1024),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockBlocks,
			prometheus.GaugeValue,
			float64(v.LockBlocks),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockBlocksAllocated,
			prometheus.GaugeValue,
			float64(v.LockBlocksAllocated),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockMemoryKB,
			prometheus.GaugeValue,
			float64(v.LockMemoryKB*1024),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockOwnerBlocks,
			prometheus.GaugeValue,
			float64(v.LockOwnerBlocks),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockOwnerBlocksAllocated,
			prometheus.GaugeValue,
			float64(v.LockOwnerBlocksAllocated),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolMemoryKB,
			prometheus.GaugeValue,
			float64(v.LogPoolMemoryKB*1024),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MaximumWorkspaceMemoryKB,
			prometheus.GaugeValue,
			float64(v.MaximumWorkspaceMemoryKB*1024),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MemoryGrantsOutstanding,
			prometheus.GaugeValue,
			float64(v.MemoryGrantsOutstanding),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MemoryGrantsPending,
			prometheus.GaugeValue,
			float64(v.MemoryGrantsPending),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.OptimizerMemoryKB,
			prometheus.GaugeValue,
			float64(v.OptimizerMemoryKB*1024),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReservedServerMemoryKB,
			prometheus.GaugeValue,
			float64(v.ReservedServerMemoryKB*1024),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SQLCacheMemoryKB,
			prometheus.GaugeValue,
			float64(v.SQLCacheMemoryKB*1024),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.StolenServerMemoryKB,
			prometheus.GaugeValue,
			float64(v.StolenServerMemoryKB*1024),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TargetServerMemoryKB,
			prometheus.GaugeValue,
			float64(v.TargetServerMemoryKB*1024),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalServerMemoryKB,
			prometheus.GaugeValue,
			float64(v.TotalServerMemoryKB*1024),
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
		v := dst[0]

		ch <- prometheus.MustNewConstMetric(
			c.AutoParamAttemptsPersec,
			prometheus.CounterValue,
			float64(v.AutoParamAttemptsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BatchRequestsPersec,
			prometheus.CounterValue,
			float64(v.BatchRequestsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FailedAutoParamsPersec,
			prometheus.CounterValue,
			float64(v.FailedAutoParamsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ForcedParameterizationsPersec,
			prometheus.CounterValue,
			float64(v.ForcedParameterizationsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GuidedplanexecutionsPersec,
			prometheus.CounterValue,
			float64(v.GuidedplanexecutionsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MisguidedplanexecutionsPersec,
			prometheus.CounterValue,
			float64(v.MisguidedplanexecutionsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SafeAutoParamsPersec,
			prometheus.CounterValue,
			float64(v.SafeAutoParamsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SQLAttentionrate,
			prometheus.GaugeValue,
			float64(v.SQLAttentionrate),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SQLCompilationsPersec,
			prometheus.CounterValue,
			float64(v.SQLCompilationsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SQLReCompilationsPersec,
			prometheus.CounterValue,
			float64(v.SQLReCompilationsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.UnsafeAutoParamsPersec,
			prometheus.CounterValue,
			float64(v.UnsafeAutoParamsPersec),
			sqlInstance,
		)
	}

	return nil, nil
}

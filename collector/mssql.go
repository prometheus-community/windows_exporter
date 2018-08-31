// returns data points from the following classes:
// - Win32_PerfRawData_MSSQLSERVER_SQLServerAccessMethods
//   https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-access-methods-object
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
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"golang.org/x/sys/windows/registry"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	mssqlEnabledCollectors = kingpin.Flag(
		"collectors.mssql.classes-enabled",
		"Comma-separated list of mssql WMI classes to use.").
		Default(mssqlAvailableClassCollectors()).String()

	mssqlPrintCollectors = kingpin.Flag(
		"collectors.mssql.class-print",
		"If true, print available mssql WMI classes and exit.  Only displays if the mssql collector is enabled.",
	).Bool()
)

type mssqlInstancesType map[string]string

func getMSSQLInstances() mssqlInstancesType {
	sqlInstances := make(mssqlInstancesType)

	// in case querying the registry fails, return the default instance
	sqlDefaultInstance := make(mssqlInstancesType)
	sqlDefaultInstance["MSSQLSERVER"] = ""

	regkey := `Software\Microsoft\Microsoft SQL Server\Instance Names\SQL`
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, regkey, registry.QUERY_VALUE)
	if err != nil {
		log.Warn("Couldn't open registry to determine SQL instances:", err)
		return sqlDefaultInstance
	}
	defer k.Close()

	instanceNames, err := k.ReadValueNames(0)
	if err != nil {
		log.Warnf("Can't ReadSubKeyNames %#v", err)
		return sqlDefaultInstance
	}

	for _, instanceName := range instanceNames {
		if instanceVersion, _, err := k.GetStringValue(instanceName); err == nil {
			sqlInstances[instanceName] = instanceVersion
		}
	}

	log.Debugf("Detected MSSQL Instances: %#v\n", sqlInstances)

	return sqlInstances
}

// mssqlBuildWMIInstanceClass - a helper function to build the correct WMI class name
// if default instance, class looks like `Win32_PerfRawData_MSSQLSERVER_SQLServerGeneralStatistics`
// if instance is 'SQLA` class looks like `Win32_PerfRawData_MSSQLSQLA_MSSQLSQLAGeneralStatistics`
func mssqlBuildWMIInstanceClass(suffix string, instance string) string {
	instancePart := "MSSQLSERVER_SQLServer"
	if instance != "MSSQLSERVER" {
		// Instance names can contain some special characters, which are not supported in the WMI class name.
		// We strip those out.
		cleanedName := strings.Map(func(r rune) rune {
			if r == '_' || r == '$' || r == '#' {
				return -1
			}
			return r
		}, instance)
		instancePart = fmt.Sprintf("MSSQL%s_MSSQL%s", cleanedName, cleanedName)
	}

	return fmt.Sprintf("Win32_PerfRawData_%s%s", instancePart, suffix)
}

type mssqlCollectorsMap map[string]mssqlCollectorFunc

func mssqlAvailableClassCollectors() string {
	return "accessmethods,availreplica,bufman,databases,dbreplica,genstats,locks,memmgr,sqlstats"
}

func (c *MSSQLCollector) getMSSQLCollectors() mssqlCollectorsMap {
	mssqlCollectors := make(mssqlCollectorsMap)
	mssqlCollectors["accessmethods"] = c.collectAccessMethods
	mssqlCollectors["availreplica"] = c.collectAvailabilityReplica
	mssqlCollectors["bufman"] = c.collectBufferManager
	mssqlCollectors["databases"] = c.collectDatabases
	mssqlCollectors["dbreplica"] = c.collectDatabaseReplica
	mssqlCollectors["genstats"] = c.collectGeneralStatistics
	mssqlCollectors["locks"] = c.collectLocks
	mssqlCollectors["memmgr"] = c.collectMemoryManager
	mssqlCollectors["sqlstats"] = c.collectSQLStats

	return mssqlCollectors
}

func mssqlExpandEnabledCollectors(enabled string) []string {
	separated := strings.Split(enabled, ",")
	unique := map[string]bool{}
	for _, s := range separated {
		if s != "" {
			unique[s] = true
		}
	}
	result := make([]string, 0, len(unique))
	for s := range unique {
		result = append(result, s)
	}
	return result
}

func init() {
	Factories["mssql"] = NewMSSQLCollector
}

// A MSSQLCollector is a Prometheus collector for various WMI Win32_PerfRawData_MSSQLSERVER_* metrics
type MSSQLCollector struct {
	// meta
	mssqlScrapeDurationDesc *prometheus.Desc
	mssqlScrapeSuccessDesc  *prometheus.Desc

	// Win32_PerfRawData_{instance}_SQLServerAccessMethods
	AccessMethodsAUcleanupbatches             *prometheus.Desc
	AccessMethodsAUcleanups                   *prometheus.Desc
	AccessMethodsByreferenceLobCreateCount    *prometheus.Desc
	AccessMethodsByreferenceLobUseCount       *prometheus.Desc
	AccessMethodsCountLobReadahead            *prometheus.Desc
	AccessMethodsCountPullInRow               *prometheus.Desc
	AccessMethodsCountPushOffRow              *prometheus.Desc
	AccessMethodsDeferreddroppedAUs           *prometheus.Desc
	AccessMethodsDeferredDroppedrowsets       *prometheus.Desc
	AccessMethodsDroppedrowsetcleanups        *prometheus.Desc
	AccessMethodsDroppedrowsetsskipped        *prometheus.Desc
	AccessMethodsExtentDeallocations          *prometheus.Desc
	AccessMethodsExtentsAllocated             *prometheus.Desc
	AccessMethodsFailedAUcleanupbatches       *prometheus.Desc
	AccessMethodsFailedleafpagecookie         *prometheus.Desc
	AccessMethodsFailedtreepagecookie         *prometheus.Desc
	AccessMethodsForwardedRecords             *prometheus.Desc
	AccessMethodsFreeSpacePageFetches         *prometheus.Desc
	AccessMethodsFreeSpaceScans               *prometheus.Desc
	AccessMethodsFullScans                    *prometheus.Desc
	AccessMethodsIndexSearches                *prometheus.Desc
	AccessMethodsInSysXactwaits               *prometheus.Desc
	AccessMethodsLobHandleCreateCount         *prometheus.Desc
	AccessMethodsLobHandleDestroyCount        *prometheus.Desc
	AccessMethodsLobSSProviderCreateCount     *prometheus.Desc
	AccessMethodsLobSSProviderDestroyCount    *prometheus.Desc
	AccessMethodsLobSSProviderTruncationCount *prometheus.Desc
	AccessMethodsMixedpageallocations         *prometheus.Desc
	AccessMethodsPagecompressionattempts      *prometheus.Desc
	AccessMethodsPageDeallocations            *prometheus.Desc
	AccessMethodsPagesAllocated               *prometheus.Desc
	AccessMethodsPagescompressed              *prometheus.Desc
	AccessMethodsPageSplits                   *prometheus.Desc
	AccessMethodsProbeScans                   *prometheus.Desc
	AccessMethodsRangeScans                   *prometheus.Desc
	AccessMethodsScanPointRevalidations       *prometheus.Desc
	AccessMethodsSkippedGhostedRecords        *prometheus.Desc
	AccessMethodsTableLockEscalations         *prometheus.Desc
	AccessMethodsUsedleafpagecookie           *prometheus.Desc
	AccessMethodsUsedtreepagecookie           *prometheus.Desc
	AccessMethodsWorkfilesCreated             *prometheus.Desc
	AccessMethodsWorktablesCreated            *prometheus.Desc
	AccessMethodsWorktablesFromCacheRatio     *prometheus.Desc

	// Win32_PerfRawData_{instance}_SQLServerAvailabilityReplica
	AvailReplicaBytesReceivedfromReplica *prometheus.Desc
	AvailReplicaBytesSenttoReplica       *prometheus.Desc
	AvailReplicaBytesSenttoTransport     *prometheus.Desc
	AvailReplicaFlowControl              *prometheus.Desc
	AvailReplicaFlowControlTimems        *prometheus.Desc
	AvailReplicaReceivesfromReplica      *prometheus.Desc
	AvailReplicaResentMessages           *prometheus.Desc
	AvailReplicaSendstoReplica           *prometheus.Desc
	AvailReplicaSendstoTransport         *prometheus.Desc

	// Win32_PerfRawData_{instance}_SQLServerBufferManager
	BufManBackgroundwriterpages         *prometheus.Desc
	BufManBuffercachehitratio           *prometheus.Desc
	BufManCheckpointpages               *prometheus.Desc
	BufManDatabasepages                 *prometheus.Desc
	BufManExtensionallocatedpages       *prometheus.Desc
	BufManExtensionfreepages            *prometheus.Desc
	BufManExtensioninuseaspercentage    *prometheus.Desc
	BufManExtensionoutstandingIOcounter *prometheus.Desc
	BufManExtensionpageevictions        *prometheus.Desc
	BufManExtensionpagereads            *prometheus.Desc
	BufManExtensionpageunreferencedtime *prometheus.Desc
	BufManExtensionpagewrites           *prometheus.Desc
	BufManFreeliststalls                *prometheus.Desc
	BufManIntegralControllerSlope       *prometheus.Desc
	BufManLazywrites                    *prometheus.Desc
	BufManPagelifeexpectancy            *prometheus.Desc
	BufManPagelookups                   *prometheus.Desc
	BufManPagereads                     *prometheus.Desc
	BufManPagewrites                    *prometheus.Desc
	BufManReadaheadpages                *prometheus.Desc
	BufManReadaheadtime                 *prometheus.Desc
	BufManTargetpages                   *prometheus.Desc

	// Win32_PerfRawData_{instance}_SQLServerDatabaseReplica
	DBReplicaDatabaseFlowControlDelay  *prometheus.Desc
	DBReplicaDatabaseFlowControls      *prometheus.Desc
	DBReplicaFileBytesReceived         *prometheus.Desc
	DBReplicaGroupCommits              *prometheus.Desc
	DBReplicaGroupCommitTime           *prometheus.Desc
	DBReplicaLogApplyPendingQueue      *prometheus.Desc
	DBReplicaLogApplyReadyQueue        *prometheus.Desc
	DBReplicaLogBytesCompressed        *prometheus.Desc
	DBReplicaLogBytesDecompressed      *prometheus.Desc
	DBReplicaLogBytesReceived          *prometheus.Desc
	DBReplicaLogCompressionCachehits   *prometheus.Desc
	DBReplicaLogCompressionCachemisses *prometheus.Desc
	DBReplicaLogCompressions           *prometheus.Desc
	DBReplicaLogDecompressions         *prometheus.Desc
	DBReplicaLogremainingforundo       *prometheus.Desc
	DBReplicaLogSendQueue              *prometheus.Desc
	DBReplicaMirroredWriteTransactions *prometheus.Desc
	DBReplicaRecoveryQueue             *prometheus.Desc
	DBReplicaRedoblocked               *prometheus.Desc
	DBReplicaRedoBytesRemaining        *prometheus.Desc
	DBReplicaRedoneBytes               *prometheus.Desc
	DBReplicaRedones                   *prometheus.Desc
	DBReplicaTotalLogrequiringundo     *prometheus.Desc
	DBReplicaTransactionDelay          *prometheus.Desc

	// Win32_PerfRawData_{instance}_SQLServerDatabases
	DatabasesActiveTransactions              *prometheus.Desc
	DatabasesBackupPerRestoreThroughput      *prometheus.Desc
	DatabasesBulkCopyRows                    *prometheus.Desc
	DatabasesBulkCopyThroughput              *prometheus.Desc
	DatabasesCommittableentries              *prometheus.Desc
	DatabasesDataFilesSizeKB                 *prometheus.Desc
	DatabasesDBCCLogicalScanBytes            *prometheus.Desc
	DatabasesGroupCommitTime                 *prometheus.Desc
	DatabasesLogBytesFlushed                 *prometheus.Desc
	DatabasesLogCacheHitRatio                *prometheus.Desc
	DatabasesLogCacheReads                   *prometheus.Desc
	DatabasesLogFilesSizeKB                  *prometheus.Desc
	DatabasesLogFilesUsedSizeKB              *prometheus.Desc
	DatabasesLogFlushes                      *prometheus.Desc
	DatabasesLogFlushWaits                   *prometheus.Desc
	DatabasesLogFlushWaitTime                *prometheus.Desc
	DatabasesLogFlushWriteTimems             *prometheus.Desc
	DatabasesLogGrowths                      *prometheus.Desc
	DatabasesLogPoolCacheMisses              *prometheus.Desc
	DatabasesLogPoolDiskReads                *prometheus.Desc
	DatabasesLogPoolHashDeletes              *prometheus.Desc
	DatabasesLogPoolHashInserts              *prometheus.Desc
	DatabasesLogPoolInvalidHashEntry         *prometheus.Desc
	DatabasesLogPoolLogScanPushes            *prometheus.Desc
	DatabasesLogPoolLogWriterPushes          *prometheus.Desc
	DatabasesLogPoolPushEmptyFreePool        *prometheus.Desc
	DatabasesLogPoolPushLowMemory            *prometheus.Desc
	DatabasesLogPoolPushNoFreeBuffer         *prometheus.Desc
	DatabasesLogPoolReqBehindTrunc           *prometheus.Desc
	DatabasesLogPoolRequestsOldVLF           *prometheus.Desc
	DatabasesLogPoolRequests                 *prometheus.Desc
	DatabasesLogPoolTotalActiveLogSize       *prometheus.Desc
	DatabasesLogPoolTotalSharedPoolSize      *prometheus.Desc
	DatabasesLogShrinks                      *prometheus.Desc
	DatabasesLogTruncations                  *prometheus.Desc
	DatabasesPercentLogUsed                  *prometheus.Desc
	DatabasesReplPendingXacts                *prometheus.Desc
	DatabasesReplTransRate                   *prometheus.Desc
	DatabasesShrinkDataMovementBytes         *prometheus.Desc
	DatabasesTrackedtransactions             *prometheus.Desc
	DatabasesTransactions                    *prometheus.Desc
	DatabasesWriteTransactions               *prometheus.Desc
	DatabasesXTPControllerDLCLatencyPerFetch *prometheus.Desc
	DatabasesXTPControllerDLCPeakLatency     *prometheus.Desc
	DatabasesXTPControllerLogProcessed       *prometheus.Desc
	DatabasesXTPMemoryUsedKB                 *prometheus.Desc

	// Win32_PerfRawData_{instance}_SQLServerGeneralStatistics
	GenStatsActiveTempTables              *prometheus.Desc
	GenStatsConnectionReset               *prometheus.Desc
	GenStatsEventNotificationsDelayedDrop *prometheus.Desc
	GenStatsHTTPAuthenticatedRequests     *prometheus.Desc
	GenStatsLogicalConnections            *prometheus.Desc
	GenStatsLogins                        *prometheus.Desc
	GenStatsLogouts                       *prometheus.Desc
	GenStatsMarsDeadlocks                 *prometheus.Desc
	GenStatsNonatomicyieldrate            *prometheus.Desc
	GenStatsProcessesblocked              *prometheus.Desc
	GenStatsSOAPEmptyRequests             *prometheus.Desc
	GenStatsSOAPMethodInvocations         *prometheus.Desc
	GenStatsSOAPSessionInitiateRequests   *prometheus.Desc
	GenStatsSOAPSessionTerminateRequests  *prometheus.Desc
	GenStatsSOAPSQLRequests               *prometheus.Desc
	GenStatsSOAPWSDLRequests              *prometheus.Desc
	GenStatsSQLTraceIOProviderLockWaits   *prometheus.Desc
	GenStatsTempdbrecoveryunitid          *prometheus.Desc
	GenStatsTempdbrowsetid                *prometheus.Desc
	GenStatsTempTablesCreationRate        *prometheus.Desc
	GenStatsTempTablesForDestruction      *prometheus.Desc
	GenStatsTraceEventNotificationQueue   *prometheus.Desc
	GenStatsTransactions                  *prometheus.Desc
	GenStatsUserConnections               *prometheus.Desc

	// Win32_PerfRawData_{instance}_SQLServerLocks
	LocksAverageWaitTimems    *prometheus.Desc
	LocksLockRequests         *prometheus.Desc
	LocksLockTimeouts         *prometheus.Desc
	LocksLockTimeoutstimeout0 *prometheus.Desc
	LocksLockWaits            *prometheus.Desc
	LocksLockWaitTimems       *prometheus.Desc
	LocksNumberofDeadlocks    *prometheus.Desc

	// Win32_PerfRawData_{instance}_SQLServerMemoryManager
	MemMgrConnectionMemoryKB       *prometheus.Desc
	MemMgrDatabaseCacheMemoryKB    *prometheus.Desc
	MemMgrExternalbenefitofmemory  *prometheus.Desc
	MemMgrFreeMemoryKB             *prometheus.Desc
	MemMgrGrantedWorkspaceMemoryKB *prometheus.Desc
	MemMgrLockBlocks               *prometheus.Desc
	MemMgrLockBlocksAllocated      *prometheus.Desc
	MemMgrLockMemoryKB             *prometheus.Desc
	MemMgrLockOwnerBlocks          *prometheus.Desc
	MemMgrLockOwnerBlocksAllocated *prometheus.Desc
	MemMgrLogPoolMemoryKB          *prometheus.Desc
	MemMgrMaximumWorkspaceMemoryKB *prometheus.Desc
	MemMgrMemoryGrantsOutstanding  *prometheus.Desc
	MemMgrMemoryGrantsPending      *prometheus.Desc
	MemMgrOptimizerMemoryKB        *prometheus.Desc
	MemMgrReservedServerMemoryKB   *prometheus.Desc
	MemMgrSQLCacheMemoryKB         *prometheus.Desc
	MemMgrStolenServerMemoryKB     *prometheus.Desc
	MemMgrTargetServerMemoryKB     *prometheus.Desc
	MemMgrTotalServerMemoryKB      *prometheus.Desc

	// Win32_PerfRawData_{instance}_SQLServerSQLStatistics
	SQLStatsAutoParamAttempts       *prometheus.Desc
	SQLStatsBatchRequests           *prometheus.Desc
	SQLStatsFailedAutoParams        *prometheus.Desc
	SQLStatsForcedParameterizations *prometheus.Desc
	SQLStatsGuidedplanexecutions    *prometheus.Desc
	SQLStatsMisguidedplanexecutions *prometheus.Desc
	SQLStatsSafeAutoParams          *prometheus.Desc
	SQLStatsSQLAttentionrate        *prometheus.Desc
	SQLStatsSQLCompilations         *prometheus.Desc
	SQLStatsSQLReCompilations       *prometheus.Desc
	SQLStatsUnsafeAutoParams        *prometheus.Desc

	mssqlInstances             mssqlInstancesType
	mssqlCollectors            mssqlCollectorsMap
	mssqlChildCollectorFailure int
}

// NewMSSQLCollector ...
func NewMSSQLCollector() (Collector, error) {

	const subsystem = "mssql"

	MSSQLCollector := MSSQLCollector{
		// meta
		mssqlScrapeDurationDesc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "collector_duration_seconds"),
			"wmi_exporter: Duration of an mssql child collection.",
			[]string{"collector", "instance"},
			nil,
		),
		mssqlScrapeSuccessDesc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "collector_success"),
			"wmi_exporter: Whether a mssql child collector was successful.",
			[]string{"collector", "instance"},
			nil,
		),

		// Win32_PerfRawData_{instance}_SQLServerAccessMethods
		AccessMethodsAUcleanupbatches: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_au_batch_cleanups"),
			"(AccessMethods.AUcleanupbatches)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsAUcleanups: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_au_cleanups"),
			"(AccessMethods.AUcleanups)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsByreferenceLobCreateCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_by_reference_lob_creates"),
			"(AccessMethods.ByreferenceLobCreateCount)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsByreferenceLobUseCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_by_reference_lob_uses"),
			"(AccessMethods.ByreferenceLobUseCount)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsCountLobReadahead: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_lob_read_aheads"),
			"(AccessMethods.CountLobReadahead)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsCountPullInRow: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_column_value_pulls"),
			"(AccessMethods.CountPullInRow)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsCountPushOffRow: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_column_value_pushes"),
			"(AccessMethods.CountPushOffRow)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsDeferreddroppedAUs: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_deferred_dropped_aus"),
			"(AccessMethods.DeferreddroppedAUs)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsDeferredDroppedrowsets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_deferred_dropped_rowsets"),
			"(AccessMethods.DeferredDroppedrowsets)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsDroppedrowsetcleanups: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_dropped_rowset_cleanups"),
			"(AccessMethods.Droppedrowsetcleanups)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsDroppedrowsetsskipped: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_dropped_rowset_skips"),
			"(AccessMethods.Droppedrowsetsskipped)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsExtentDeallocations: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_extent_deallocations"),
			"(AccessMethods.ExtentDeallocations)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsExtentsAllocated: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_extent_allocations"),
			"(AccessMethods.ExtentsAllocated)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsFailedAUcleanupbatches: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_au_batch_cleanup_failures"),
			"(AccessMethods.FailedAUcleanupbatches)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsFailedleafpagecookie: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_leaf_page_cookie_failures"),
			"(AccessMethods.Failedleafpagecookie)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsFailedtreepagecookie: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_tree_page_cookie_failures"),
			"(AccessMethods.Failedtreepagecookie)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsForwardedRecords: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_forwarded_records"),
			"(AccessMethods.ForwardedRecords)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsFreeSpacePageFetches: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_free_space_page_fetches"),
			"(AccessMethods.FreeSpacePageFetches)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsFreeSpaceScans: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_free_space_scans"),
			"(AccessMethods.FreeSpaceScans)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsFullScans: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_full_scans"),
			"(AccessMethods.FullScans)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsIndexSearches: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_index_searches"),
			"(AccessMethods.IndexSearches)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsInSysXactwaits: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_insysxact_waits"),
			"(AccessMethods.InSysXactwaits)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsLobHandleCreateCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_lob_handle_creates"),
			"(AccessMethods.LobHandleCreateCount)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsLobHandleDestroyCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_lob_handle_destroys"),
			"(AccessMethods.LobHandleDestroyCount)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsLobSSProviderCreateCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_lob_ss_provider_creates"),
			"(AccessMethods.LobSSProviderCreateCount)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsLobSSProviderDestroyCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_lob_ss_provider_destroys"),
			"(AccessMethods.LobSSProviderDestroyCount)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsLobSSProviderTruncationCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_lob_ss_provider_truncations"),
			"(AccessMethods.LobSSProviderTruncationCount)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsMixedpageallocations: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_mixed_page_allocations"),
			"(AccessMethods.MixedpageallocationsPersec)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsPagecompressionattempts: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_page_compression_attempts"),
			"(AccessMethods.PagecompressionattemptsPersec)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsPageDeallocations: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_page_deallocations"),
			"(AccessMethods.PageDeallocationsPersec)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsPagesAllocated: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_page_allocations"),
			"(AccessMethods.PagesAllocatedPersec)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsPagescompressed: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_page_compressions"),
			"(AccessMethods.PagescompressedPersec)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsPageSplits: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_page_splits"),
			"(AccessMethods.PageSplitsPersec)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsProbeScans: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_probe_scans"),
			"(AccessMethods.ProbeScansPersec)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsRangeScans: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_range_scans"),
			"(AccessMethods.RangeScansPersec)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsScanPointRevalidations: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_scan_point_revalidations"),
			"(AccessMethods.ScanPointRevalidationsPersec)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsSkippedGhostedRecords: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_ghost_record_skips"),
			"(AccessMethods.SkippedGhostedRecordsPersec)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsTableLockEscalations: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_table_lock_escalations"),
			"(AccessMethods.TableLockEscalationsPersec)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsUsedleafpagecookie: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_leaf_page_cookie_uses"),
			"(AccessMethods.Usedleafpagecookie)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsUsedtreepagecookie: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_tree_page_cookie_uses"),
			"(AccessMethods.Usedtreepagecookie)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsWorkfilesCreated: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_workfile_creates"),
			"(AccessMethods.WorkfilesCreatedPersec)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsWorktablesCreated: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_worktables_creates"),
			"(AccessMethods.WorktablesCreatedPersec)",
			[]string{"instance"},
			nil,
		),
		AccessMethodsWorktablesFromCacheRatio: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accessmethods_worktables_from_cache_ratio"),
			"(AccessMethods.WorktablesFromCacheRatio)",
			[]string{"instance"},
			nil,
		),

		// Win32_PerfRawData_{instance}_SQLServerAvailabilityReplica
		AvailReplicaBytesReceivedfromReplica: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "availreplica_received_from_replica_bytes"),
			"(AvailabilityReplica.BytesReceivedfromReplica)",
			[]string{"instance", "replica"},
			nil,
		),
		AvailReplicaBytesSenttoReplica: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "availreplica_sent_to_replica_bytes"),
			"(AvailabilityReplica.BytesSenttoReplica)",
			[]string{"instance", "replica"},
			nil,
		),
		AvailReplicaBytesSenttoTransport: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "availreplica_sent_to_transport_bytes"),
			"(AvailabilityReplica.BytesSenttoTransport)",
			[]string{"instance", "replica"},
			nil,
		),
		AvailReplicaFlowControl: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "availreplica_initiated_flow_controls"),
			"(AvailabilityReplica.FlowControl)",
			[]string{"instance", "replica"},
			nil,
		),
		AvailReplicaFlowControlTimems: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "availreplica_flow_control_wait_seconds"),
			"(AvailabilityReplica.FlowControlTimems)",
			[]string{"instance", "replica"},
			nil,
		),
		AvailReplicaReceivesfromReplica: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "availreplica_receives_from_replica"),
			"(AvailabilityReplica.ReceivesfromReplica)",
			[]string{"instance", "replica"},
			nil,
		),
		AvailReplicaResentMessages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "availreplica_resent_messages"),
			"(AvailabilityReplica.ResentMessages)",
			[]string{"instance", "replica"},
			nil,
		),
		AvailReplicaSendstoReplica: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "availreplica_sends_to_replica"),
			"(AvailabilityReplica.SendstoReplica)",
			[]string{"instance", "replica"},
			nil,
		),
		AvailReplicaSendstoTransport: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "availreplica_sends_to_transport"),
			"(AvailabilityReplica.SendstoTransport)",
			[]string{"instance", "replica"},
			nil,
		),

		// Win32_PerfRawData_{instance}_SQLServerBufferManager
		BufManBackgroundwriterpages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_background_writer_pages"),
			"(BufferManager.Backgroundwriterpages)",
			[]string{"instance"},
			nil,
		),
		BufManBuffercachehitratio: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_buffer_cache_hit_ratio"),
			"(BufferManager.Buffercachehitratio)",
			[]string{"instance"},
			nil,
		),
		BufManCheckpointpages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_checkpoint_pages"),
			"(BufferManager.Checkpointpages)",
			[]string{"instance"},
			nil,
		),
		BufManDatabasepages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_database_pages"),
			"(BufferManager.Databasepages)",
			[]string{"instance"},
			nil,
		),
		BufManExtensionallocatedpages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_extension_allocated_pages"),
			"(BufferManager.Extensionallocatedpages)",
			[]string{"instance"},
			nil,
		),
		BufManExtensionfreepages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_extension_free_pages"),
			"(BufferManager.Extensionfreepages)",
			[]string{"instance"},
			nil,
		),
		BufManExtensioninuseaspercentage: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_extension_in_use_as_percentage"),
			"(BufferManager.Extensioninuseaspercentage)",
			[]string{"instance"},
			nil,
		),
		BufManExtensionoutstandingIOcounter: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_extension_outstanding_io"),
			"(BufferManager.ExtensionoutstandingIOcounter)",
			[]string{"instance"},
			nil,
		),
		BufManExtensionpageevictions: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_extension_page_evictions"),
			"(BufferManager.Extensionpageevictions)",
			[]string{"instance"},
			nil,
		),
		BufManExtensionpagereads: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_extension_page_reads"),
			"(BufferManager.Extensionpagereads)",
			[]string{"instance"},
			nil,
		),
		BufManExtensionpageunreferencedtime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_extension_page_unreferenced_seconds"),
			"(BufferManager.Extensionpageunreferencedtime)",
			[]string{"instance"},
			nil,
		),
		BufManExtensionpagewrites: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_extension_page_writes"),
			"(BufferManager.Extensionpagewrites)",
			[]string{"instance"},
			nil,
		),
		BufManFreeliststalls: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_free_list_stalls"),
			"(BufferManager.Freeliststalls)",
			[]string{"instance"},
			nil,
		),
		BufManIntegralControllerSlope: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_integral_controller_slope"),
			"(BufferManager.IntegralControllerSlope)",
			[]string{"instance"},
			nil,
		),
		BufManLazywrites: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_lazywrites"),
			"(BufferManager.Lazywrites)",
			[]string{"instance"},
			nil,
		),
		BufManPagelifeexpectancy: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_page_life_expectancy_seconds"),
			"(BufferManager.Pagelifeexpectancy)",
			[]string{"instance"},
			nil,
		),
		BufManPagelookups: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_page_lookups"),
			"(BufferManager.Pagelookups)",
			[]string{"instance"},
			nil,
		),
		BufManPagereads: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_page_reads"),
			"(BufferManager.Pagereads)",
			[]string{"instance"},
			nil,
		),
		BufManPagewrites: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_page_writes"),
			"(BufferManager.Pagewrites)",
			[]string{"instance"},
			nil,
		),
		BufManReadaheadpages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_read_ahead_pages"),
			"(BufferManager.Readaheadpages)",
			[]string{"instance"},
			nil,
		),
		BufManReadaheadtime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_read_ahead_issuing_seconds"),
			"(BufferManager.Readaheadtime)",
			[]string{"instance"},
			nil,
		),
		BufManTargetpages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bufman_target_pages"),
			"(BufferManager.Targetpages)",
			[]string{"instance"},
			nil,
		),

		// Win32_PerfRawData_{instance}_SQLServerDatabaseReplica
		DBReplicaDatabaseFlowControlDelay: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_database_flow_control_wait_seconds"),
			"(DatabaseReplica.DatabaseFlowControlDelay)",
			[]string{"instance", "replica"},
			nil,
		),
		DBReplicaDatabaseFlowControls: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_database_initiated_flow_controls"),
			"(DatabaseReplica.DatabaseFlowControls)",
			[]string{"instance", "replica"},
			nil,
		),
		DBReplicaFileBytesReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_received_file_bytes"),
			"(DatabaseReplica.FileBytesReceived)",
			[]string{"instance", "replica"},
			nil,
		),
		DBReplicaGroupCommits: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_group_commits"),
			"(DatabaseReplica.GroupCommits)",
			[]string{"instance", "replica"},
			nil,
		),
		DBReplicaGroupCommitTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_group_commit_stall_seconds"),
			"(DatabaseReplica.GroupCommitTime)",
			[]string{"instance", "replica"},
			nil,
		),
		DBReplicaLogApplyPendingQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_log_apply_pending_queue"),
			"(DatabaseReplica.LogApplyPendingQueue)",
			[]string{"instance", "replica"},
			nil,
		),
		DBReplicaLogApplyReadyQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_log_apply_ready_queue"),
			"(DatabaseReplica.LogApplyReadyQueue)",
			[]string{"instance", "replica"},
			nil,
		),
		DBReplicaLogBytesCompressed: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_log_compressed_bytes"),
			"(DatabaseReplica.LogBytesCompressed)",
			[]string{"instance", "replica"},
			nil,
		),
		DBReplicaLogBytesDecompressed: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_log_decompressed_bytes"),
			"(DatabaseReplica.LogBytesDecompressed)",
			[]string{"instance", "replica"},
			nil,
		),
		DBReplicaLogBytesReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_log_received_bytes"),
			"(DatabaseReplica.LogBytesReceived)",
			[]string{"instance", "replica"},
			nil,
		),
		DBReplicaLogCompressionCachehits: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_log_compression_cachehits"),
			"(DatabaseReplica.LogCompressionCachehits)",
			[]string{"instance", "replica"},
			nil,
		),
		DBReplicaLogCompressionCachemisses: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_log_compression_cachemisses"),
			"(DatabaseReplica.LogCompressionCachemisses)",
			[]string{"instance", "replica"},
			nil,
		),
		DBReplicaLogCompressions: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_log_compressions"),
			"(DatabaseReplica.LogCompressions)",
			[]string{"instance", "replica"},
			nil,
		),
		DBReplicaLogDecompressions: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_log_decompressions"),
			"(DatabaseReplica.LogDecompressions)",
			[]string{"instance", "replica"},
			nil,
		),
		DBReplicaLogremainingforundo: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_log_remaining_for_undo"),
			"(DatabaseReplica.Logremainingforundo)",
			[]string{"instance", "replica"},
			nil,
		),
		DBReplicaLogSendQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_log_send_queue"),
			"(DatabaseReplica.LogSendQueue)",
			[]string{"instance", "replica"},
			nil,
		),
		DBReplicaMirroredWriteTransactions: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_mirrored_write_transactions"),
			"(DatabaseReplica.MirroredWriteTransactions)",
			[]string{"instance", "replica"},
			nil,
		),
		DBReplicaRecoveryQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_recovery_queue_records"),
			"(DatabaseReplica.RecoveryQueue)",
			[]string{"instance", "replica"},
			nil,
		),
		DBReplicaRedoblocked: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_redo_blocks"),
			"(DatabaseReplica.Redoblocked)",
			[]string{"instance", "replica"},
			nil,
		),
		DBReplicaRedoBytesRemaining: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_redo_remaining_bytes"),
			"(DatabaseReplica.RedoBytesRemaining)",
			[]string{"instance", "replica"},
			nil,
		),
		DBReplicaRedoneBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_redone_bytes"),
			"(DatabaseReplica.RedoneBytes)",
			[]string{"instance", "replica"},
			nil,
		),
		DBReplicaRedones: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_redones"),
			"(DatabaseReplica.Redones)",
			[]string{"instance", "replica"},
			nil,
		),
		DBReplicaTotalLogrequiringundo: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_total_log_requiring_undo"),
			"(DatabaseReplica.TotalLogrequiringundo)",
			[]string{"instance", "replica"},
			nil,
		),
		DBReplicaTransactionDelay: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dbreplica_transaction_delay_seconds"),
			"(DatabaseReplica.TransactionDelay)",
			[]string{"instance", "replica"},
			nil,
		),

		// Win32_PerfRawData_{instance}_SQLServerDatabases
		DatabasesActiveTransactions: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_active_transactions"),
			"(Databases.ActiveTransactions)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesBackupPerRestoreThroughput: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_backup_restore_operations"),
			"(Databases.BackupPerRestoreThroughput)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesBulkCopyRows: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_bulk_copy_rows"),
			"(Databases.BulkCopyRows)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesBulkCopyThroughput: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_bulk_copy_bytes"),
			"(Databases.BulkCopyThroughput)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesCommittableentries: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_commit_table_entries"),
			"(Databases.Committableentries)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesDataFilesSizeKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_data_files_size_bytes"),
			"(Databases.DataFilesSizeKB)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesDBCCLogicalScanBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_dbcc_logical_scan_bytes"),
			"(Databases.DBCCLogicalScanBytes)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesGroupCommitTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_group_commit_stall_seconds"),
			"(Databases.GroupCommitTime)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesLogBytesFlushed: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_flushed_bytes"),
			"(Databases.LogBytesFlushed)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesLogCacheHitRatio: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_cache_hit_ratio"),
			"(Databases.LogCacheHitRatio)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesLogCacheReads: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_cache_reads"),
			"(Databases.LogCacheReads)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesLogFilesSizeKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_files_size_bytes"),
			"(Databases.LogFilesSizeKB)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesLogFilesUsedSizeKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_files_used_size_bytes"),
			"(Databases.LogFilesUsedSizeKB)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesLogFlushes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_flushes"),
			"(Databases.LogFlushes)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesLogFlushWaits: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_flush_waits"),
			"(Databases.LogFlushWaits)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesLogFlushWaitTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_flush_wait_seconds"),
			"(Databases.LogFlushWaitTime)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesLogFlushWriteTimems: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_flush_write_seconds"),
			"(Databases.LogFlushWriteTimems)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesLogGrowths: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_growths"),
			"(Databases.LogGrowths)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesLogPoolCacheMisses: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_pool_cache_misses"),
			"(Databases.LogPoolCacheMisses)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesLogPoolDiskReads: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_pool_disk_reads"),
			"(Databases.LogPoolDiskReads)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesLogPoolHashDeletes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_pool_hash_deletes"),
			"(Databases.LogPoolHashDeletes)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesLogPoolHashInserts: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_pool_hash_inserts"),
			"(Databases.LogPoolHashInserts)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesLogPoolInvalidHashEntry: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_pool_invalid_hash_entries"),
			"(Databases.LogPoolInvalidHashEntry)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesLogPoolLogScanPushes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_pool_log_scan_pushes"),
			"(Databases.LogPoolLogScanPushes)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesLogPoolLogWriterPushes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_pool_log_writer_pushes"),
			"(Databases.LogPoolLogWriterPushes)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesLogPoolPushEmptyFreePool: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_pool_empty_free_pool_pushes"),
			"(Databases.LogPoolPushEmptyFreePool)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesLogPoolPushLowMemory: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_pool_low_memory_pushes"),
			"(Databases.LogPoolPushLowMemory)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesLogPoolPushNoFreeBuffer: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_pool_no_free_buffer_pushes"),
			"(Databases.LogPoolPushNoFreeBuffer)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesLogPoolReqBehindTrunc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_pool_req_behind_trunc"),
			"(Databases.LogPoolReqBehindTrunc)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesLogPoolRequestsOldVLF: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_pool_requests_old_vlf"),
			"(Databases.LogPoolRequestsOldVLF)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesLogPoolRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_pool_requests"),
			"(Databases.LogPoolRequests)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesLogPoolTotalActiveLogSize: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_pool_total_active_log_bytes"),
			"(Databases.LogPoolTotalActiveLogSize)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesLogPoolTotalSharedPoolSize: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_pool_total_shared_pool_bytes"),
			"(Databases.LogPoolTotalSharedPoolSize)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesLogShrinks: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_shrinks"),
			"(Databases.LogShrinks)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesLogTruncations: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_truncations"),
			"(Databases.LogTruncations)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesPercentLogUsed: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_log_used_percent"),
			"(Databases.PercentLogUsed)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesReplPendingXacts: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_pending_repl_transactions"),
			"(Databases.ReplPendingTransactions)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesReplTransRate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_repl_transactions"),
			"(Databases.ReplTranactions)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesShrinkDataMovementBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_shrink_data_movement_bytes"),
			"(Databases.ShrinkDataMovementBytes)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesTrackedtransactions: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_tracked_transactions"),
			"(Databases.Trackedtransactions)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesTransactions: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_transactions"),
			"(Databases.Transactions)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesWriteTransactions: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_write_transactions"),
			"(Databases.WriteTransactions)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesXTPControllerDLCLatencyPerFetch: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_xtp_controller_dlc_fetch_latency_seconds"),
			"(Databases.XTPControllerDLCLatencyPerFetch)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesXTPControllerDLCPeakLatency: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_xtp_controller_dlc_peak_latency_seconds"),
			"(Databases.XTPControllerDLCPeakLatency)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesXTPControllerLogProcessed: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_xtp_controller_log_processed_bytes"),
			"(Databases.XTPControllerLogProcessed)",
			[]string{"instance", "database"},
			nil,
		),
		DatabasesXTPMemoryUsedKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "databases_xtp_memory_used_bytes"),
			"(Databases.XTPMemoryUsedKB)",
			[]string{"instance", "database"},
			nil,
		),

		// Win32_PerfRawData_{instance}_SQLServerGeneralStatistics
		GenStatsActiveTempTables: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_active_temp_tables"),
			"(GeneralStatistics.ActiveTempTables)",
			[]string{"instance"},
			nil,
		),
		GenStatsConnectionReset: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_connection_resets"),
			"(GeneralStatistics.ConnectionReset)",
			[]string{"instance"},
			nil,
		),
		GenStatsEventNotificationsDelayedDrop: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_event_notifications_delayed_drop"),
			"(GeneralStatistics.EventNotificationsDelayedDrop)",
			[]string{"instance"},
			nil,
		),
		GenStatsHTTPAuthenticatedRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_http_authenticated_requests"),
			"(GeneralStatistics.HTTPAuthenticatedRequests)",
			[]string{"instance"},
			nil,
		),
		GenStatsLogicalConnections: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_logical_connections"),
			"(GeneralStatistics.LogicalConnections)",
			[]string{"instance"},
			nil,
		),
		GenStatsLogins: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_logins"),
			"(GeneralStatistics.Logins)",
			[]string{"instance"},
			nil,
		),
		GenStatsLogouts: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_logouts"),
			"(GeneralStatistics.Logouts)",
			[]string{"instance"},
			nil,
		),
		GenStatsMarsDeadlocks: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_mars_deadlocks"),
			"(GeneralStatistics.MarsDeadlocks)",
			[]string{"instance"},
			nil,
		),
		GenStatsNonatomicyieldrate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_non_atomic_yields"),
			"(GeneralStatistics.Nonatomicyields)",
			[]string{"instance"},
			nil,
		),
		GenStatsProcessesblocked: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_blocked_processes"),
			"(GeneralStatistics.Processesblocked)",
			[]string{"instance"},
			nil,
		),
		GenStatsSOAPEmptyRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_soap_empty_requests"),
			"(GeneralStatistics.SOAPEmptyRequests)",
			[]string{"instance"},
			nil,
		),
		GenStatsSOAPMethodInvocations: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_soap_method_invocations"),
			"(GeneralStatistics.SOAPMethodInvocations)",
			[]string{"instance"},
			nil,
		),
		GenStatsSOAPSessionInitiateRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_soap_session_initiate_requests"),
			"(GeneralStatistics.SOAPSessionInitiateRequests)",
			[]string{"instance"},
			nil,
		),
		GenStatsSOAPSessionTerminateRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_soap_session_terminate_requests"),
			"(GeneralStatistics.SOAPSessionTerminateRequests)",
			[]string{"instance"},
			nil,
		),
		GenStatsSOAPSQLRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_soapsql_requests"),
			"(GeneralStatistics.SOAPSQLRequests)",
			[]string{"instance"},
			nil,
		),
		GenStatsSOAPWSDLRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_soapwsdl_requests"),
			"(GeneralStatistics.SOAPWSDLRequests)",
			[]string{"instance"},
			nil,
		),
		GenStatsSQLTraceIOProviderLockWaits: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_sql_trace_io_provider_lock_waits"),
			"(GeneralStatistics.SQLTraceIOProviderLockWaits)",
			[]string{"instance"},
			nil,
		),
		GenStatsTempdbrecoveryunitid: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_tempdb_recovery_unit_ids_generated"),
			"(GeneralStatistics.Tempdbrecoveryunitid)",
			[]string{"instance"},
			nil,
		),
		GenStatsTempdbrowsetid: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_tempdb_rowset_ids_generated"),
			"(GeneralStatistics.Tempdbrowsetid)",
			[]string{"instance"},
			nil,
		),
		GenStatsTempTablesCreationRate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_temp_tables_creations"),
			"(GeneralStatistics.TempTablesCreations)",
			[]string{"instance"},
			nil,
		),
		GenStatsTempTablesForDestruction: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_temp_tables_awaiting_destruction"),
			"(GeneralStatistics.TempTablesForDestruction)",
			[]string{"instance"},
			nil,
		),
		GenStatsTraceEventNotificationQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_trace_event_notification_queue_size"),
			"(GeneralStatistics.TraceEventNotificationQueue)",
			[]string{"instance"},
			nil,
		),
		GenStatsTransactions: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_transactions"),
			"(GeneralStatistics.Transactions)",
			[]string{"instance"},
			nil,
		),
		GenStatsUserConnections: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstats_user_connections"),
			"(GeneralStatistics.UserConnections)",
			[]string{"instance"},
			nil,
		),

		// Win32_PerfRawData_{instance}_SQLServerLocks
		LocksAverageWaitTimems: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "locks_average_wait_seconds"),
			"(Locks.AverageWaitTimems)",
			[]string{"instance", "resource"},
			nil,
		),
		LocksLockRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "locks_lock_requests"),
			"(Locks.LockRequests)",
			[]string{"instance", "resource"},
			nil,
		),
		LocksLockTimeouts: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "locks_lock_timeouts"),
			"(Locks.LockTimeouts)",
			[]string{"instance", "resource"},
			nil,
		),
		LocksLockTimeoutstimeout0: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "locks_lock_timeouts_excluding_NOWAIT"),
			"(Locks.LockTimeoutstimeout0)",
			[]string{"instance", "resource"},
			nil,
		),
		LocksLockWaits: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "locks_lock_waits"),
			"(Locks.LockWaits)",
			[]string{"instance", "resource"},
			nil,
		),
		LocksLockWaitTimems: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "locks_lock_wait_seconds"),
			"(Locks.LockWaitTimems)",
			[]string{"instance", "resource"},
			nil,
		),
		LocksNumberofDeadlocks: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "locks_deadlocks"),
			"(Locks.NumberofDeadlocks)",
			[]string{"instance", "resource"},
			nil,
		),

		// Win32_PerfRawData_{instance}_SQLServerMemoryManager
		MemMgrConnectionMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_connection_memory_bytes"),
			"(MemoryManager.ConnectionMemoryKB)",
			[]string{"instance"},
			nil,
		),
		MemMgrDatabaseCacheMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_database_cache_memory_bytes"),
			"(MemoryManager.DatabaseCacheMemoryKB)",
			[]string{"instance"},
			nil,
		),
		MemMgrExternalbenefitofmemory: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_external_benefit_of_memory"),
			"(MemoryManager.Externalbenefitofmemory)",
			[]string{"instance"},
			nil,
		),
		MemMgrFreeMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_free_memory_bytes"),
			"(MemoryManager.FreeMemoryKB)",
			[]string{"instance"},
			nil,
		),
		MemMgrGrantedWorkspaceMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_granted_workspace_memory_bytes"),
			"(MemoryManager.GrantedWorkspaceMemoryKB)",
			[]string{"instance"},
			nil,
		),
		MemMgrLockBlocks: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_lock_blocks"),
			"(MemoryManager.LockBlocks)",
			[]string{"instance"},
			nil,
		),
		MemMgrLockBlocksAllocated: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_allocated_lock_blocks"),
			"(MemoryManager.LockBlocksAllocated)",
			[]string{"instance"},
			nil,
		),
		MemMgrLockMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_lock_memory_bytes"),
			"(MemoryManager.LockMemoryKB)",
			[]string{"instance"},
			nil,
		),
		MemMgrLockOwnerBlocks: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_lock_owner_blocks"),
			"(MemoryManager.LockOwnerBlocks)",
			[]string{"instance"},
			nil,
		),
		MemMgrLockOwnerBlocksAllocated: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_allocated_lock_owner_blocks"),
			"(MemoryManager.LockOwnerBlocksAllocated)",
			[]string{"instance"},
			nil,
		),
		MemMgrLogPoolMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_log_pool_memory_bytes"),
			"(MemoryManager.LogPoolMemoryKB)",
			[]string{"instance"},
			nil,
		),
		MemMgrMaximumWorkspaceMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_maximum_workspace_memory_bytes"),
			"(MemoryManager.MaximumWorkspaceMemoryKB)",
			[]string{"instance"},
			nil,
		),
		MemMgrMemoryGrantsOutstanding: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_outstanding_memory_grants"),
			"(MemoryManager.MemoryGrantsOutstanding)",
			[]string{"instance"},
			nil,
		),
		MemMgrMemoryGrantsPending: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_pending_memory_grants"),
			"(MemoryManager.MemoryGrantsPending)",
			[]string{"instance"},
			nil,
		),
		MemMgrOptimizerMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_optimizer_memory_bytes"),
			"(MemoryManager.OptimizerMemoryKB)",
			[]string{"instance"},
			nil,
		),
		MemMgrReservedServerMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_reserved_server_memory_bytes"),
			"(MemoryManager.ReservedServerMemoryKB)",
			[]string{"instance"},
			nil,
		),
		MemMgrSQLCacheMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_sql_cache_memory_bytes"),
			"(MemoryManager.SQLCacheMemoryKB)",
			[]string{"instance"},
			nil,
		),
		MemMgrStolenServerMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_stolen_server_memory_bytes"),
			"(MemoryManager.StolenServerMemoryKB)",
			[]string{"instance"},
			nil,
		),
		MemMgrTargetServerMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_target_server_memory_bytes"),
			"(MemoryManager.TargetServerMemoryKB)",
			[]string{"instance"},
			nil,
		),
		MemMgrTotalServerMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memmgr_total_server_memory_bytes"),
			"(MemoryManager.TotalServerMemoryKB)",
			[]string{"instance"},
			nil,
		),

		// Win32_PerfRawData_{instance}_SQLServerSQLStatistics
		SQLStatsAutoParamAttempts: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sqlstats_auto_parameterization_attempts"),
			"(SQLStatistics.AutoParamAttempts)",
			[]string{"instance"},
			nil,
		),
		SQLStatsBatchRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sqlstats_batch_requests"),
			"(SQLStatistics.BatchRequests)",
			[]string{"instance"},
			nil,
		),
		SQLStatsFailedAutoParams: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sqlstats_failed_auto_parameterization_attempts"),
			"(SQLStatistics.FailedAutoParams)",
			[]string{"instance"},
			nil,
		),
		SQLStatsForcedParameterizations: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sqlstats_forced_parameterizations"),
			"(SQLStatistics.ForcedParameterizations)",
			[]string{"instance"},
			nil,
		),
		SQLStatsGuidedplanexecutions: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sqlstats_guided_plan_executions"),
			"(SQLStatistics.Guidedplanexecutions)",
			[]string{"instance"},
			nil,
		),
		SQLStatsMisguidedplanexecutions: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sqlstats_misguided_plan_executions"),
			"(SQLStatistics.Misguidedplanexecutions)",
			[]string{"instance"},
			nil,
		),
		SQLStatsSafeAutoParams: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sqlstats_safe_auto_parameterization_attempts"),
			"(SQLStatistics.SafeAutoParams)",
			[]string{"instance"},
			nil,
		),
		SQLStatsSQLAttentionrate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sqlstats_sql_attentions"),
			"(SQLStatistics.SQLAttentions)",
			[]string{"instance"},
			nil,
		),
		SQLStatsSQLCompilations: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sqlstats_sql_compilations"),
			"(SQLStatistics.SQLCompilations)",
			[]string{"instance"},
			nil,
		),
		SQLStatsSQLReCompilations: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sqlstats_sql_recompilations"),
			"(SQLStatistics.SQLReCompilations)",
			[]string{"instance"},
			nil,
		),
		SQLStatsUnsafeAutoParams: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sqlstats_unsafe_auto_parameterization_attempts"),
			"(SQLStatistics.UnsafeAutoParams)",
			[]string{"instance"},
			nil,
		),

		mssqlInstances: getMSSQLInstances(),
	}

	MSSQLCollector.mssqlCollectors = MSSQLCollector.getMSSQLCollectors()

	if *mssqlPrintCollectors {
		fmt.Printf("Available SQLServer Classes:\n")
		for name := range MSSQLCollector.mssqlCollectors {
			fmt.Printf(" - %s\n", name)
		}
		os.Exit(0)
	}

	return &MSSQLCollector, nil
}

type mssqlCollectorFunc func(ch chan<- prometheus.Metric, sqlInstance string) (*prometheus.Desc, error)

func (c *MSSQLCollector) execute(name string, fn mssqlCollectorFunc, ch chan<- prometheus.Metric, sqlInstance string, wg *sync.WaitGroup) {
	defer wg.Done()

	begin := time.Now()
	_, err := fn(ch, sqlInstance)
	duration := time.Since(begin)
	var success float64

	if err != nil {
		log.Errorf("mssql class collector %s failed after %fs: %s", name, duration.Seconds(), err)
		success = 0
		c.mssqlChildCollectorFailure++
	} else {
		log.Debugf("mssql class collector %s succeeded after %fs.", name, duration.Seconds())
		success = 1
	}
	ch <- prometheus.MustNewConstMetric(
		c.mssqlScrapeDurationDesc,
		prometheus.GaugeValue,
		duration.Seconds(),
		name, sqlInstance,
	)
	ch <- prometheus.MustNewConstMetric(
		c.mssqlScrapeSuccessDesc,
		prometheus.GaugeValue,
		success,
		name, sqlInstance,
	)
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *MSSQLCollector) Collect(ch chan<- prometheus.Metric) error {
	wg := sync.WaitGroup{}

	enabled := mssqlExpandEnabledCollectors(*mssqlEnabledCollectors)
	for sqlInstance := range c.mssqlInstances {
		for _, name := range enabled {
			function := c.mssqlCollectors[name]

			wg.Add(1)
			go c.execute(name, function, ch, sqlInstance, &wg)
		}
	}
	wg.Wait()

	// this shoud return an error if any? some? children errord.
	if c.mssqlChildCollectorFailure > 0 {
		return errors.New("at least one child collector failed")
	}
	return nil
}

type win32PerfRawDataSQLServerAccessMethods struct {
	AUcleanupbatchesPersec        uint64
	AUcleanupsPersec              uint64
	ByreferenceLobCreateCount     uint64
	ByreferenceLobUseCount        uint64
	CountLobReadahead             uint64
	CountPullInRow                uint64
	CountPushOffRow               uint64
	DeferreddroppedAUs            uint64
	DeferredDroppedrowsets        uint64
	DroppedrowsetcleanupsPersec   uint64
	DroppedrowsetsskippedPersec   uint64
	ExtentDeallocationsPersec     uint64
	ExtentsAllocatedPersec        uint64
	FailedAUcleanupbatchesPersec  uint64
	Failedleafpagecookie          uint64
	Failedtreepagecookie          uint64
	ForwardedRecordsPersec        uint64
	FreeSpacePageFetchesPersec    uint64
	FreeSpaceScansPersec          uint64
	FullScansPersec               uint64
	IndexSearchesPersec           uint64
	InSysXactwaitsPersec          uint64
	LobHandleCreateCount          uint64
	LobHandleDestroyCount         uint64
	LobSSProviderCreateCount      uint64
	LobSSProviderDestroyCount     uint64
	LobSSProviderTruncationCount  uint64
	MixedpageallocationsPersec    uint64
	PagecompressionattemptsPersec uint64
	PageDeallocationsPersec       uint64
	PagesAllocatedPersec          uint64
	PagescompressedPersec         uint64
	PageSplitsPersec              uint64
	ProbeScansPersec              uint64
	RangeScansPersec              uint64
	ScanPointRevalidationsPersec  uint64
	SkippedGhostedRecordsPersec   uint64
	TableLockEscalationsPersec    uint64
	Usedleafpagecookie            uint64
	Usedtreepagecookie            uint64
	WorkfilesCreatedPersec        uint64
	WorktablesCreatedPersec       uint64
	WorktablesFromCacheRatio      uint64
}

func (c *MSSQLCollector) collectAccessMethods(ch chan<- prometheus.Metric, sqlInstance string) (*prometheus.Desc, error) {
	var dst []win32PerfRawDataSQLServerAccessMethods
	log.Debugf("mssql_accessmethods collector iterating sql instance %s.", sqlInstance)

	class := mssqlBuildWMIInstanceClass("AccessMethods", sqlInstance)
	q := queryAllForClass(&dst, class)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	if len(dst) == 0 {
		return nil, errors.New("WMI query returned empty result set")
	}

	v := dst[0]
	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsAUcleanupbatches,
		prometheus.CounterValue,
		float64(v.AUcleanupbatchesPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsAUcleanups,
		prometheus.CounterValue,
		float64(v.AUcleanupsPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsByreferenceLobCreateCount,
		prometheus.CounterValue,
		float64(v.ByreferenceLobCreateCount),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsByreferenceLobUseCount,
		prometheus.CounterValue,
		float64(v.ByreferenceLobUseCount),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsCountLobReadahead,
		prometheus.CounterValue,
		float64(v.CountLobReadahead),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsCountPullInRow,
		prometheus.CounterValue,
		float64(v.CountPullInRow),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsCountPushOffRow,
		prometheus.CounterValue,
		float64(v.CountPushOffRow),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsDeferreddroppedAUs,
		prometheus.GaugeValue,
		float64(v.DeferreddroppedAUs),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsDeferredDroppedrowsets,
		prometheus.GaugeValue,
		float64(v.DeferredDroppedrowsets),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsDroppedrowsetcleanups,
		prometheus.CounterValue,
		float64(v.DroppedrowsetcleanupsPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsDroppedrowsetsskipped,
		prometheus.CounterValue,
		float64(v.DroppedrowsetsskippedPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsExtentDeallocations,
		prometheus.CounterValue,
		float64(v.ExtentDeallocationsPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsExtentsAllocated,
		prometheus.CounterValue,
		float64(v.ExtentsAllocatedPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsFailedAUcleanupbatches,
		prometheus.CounterValue,
		float64(v.FailedAUcleanupbatchesPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsFailedleafpagecookie,
		prometheus.CounterValue,
		float64(v.Failedleafpagecookie),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsFailedtreepagecookie,
		prometheus.CounterValue,
		float64(v.Failedtreepagecookie),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsForwardedRecords,
		prometheus.CounterValue,
		float64(v.ForwardedRecordsPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsFreeSpacePageFetches,
		prometheus.CounterValue,
		float64(v.FreeSpacePageFetchesPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsFreeSpaceScans,
		prometheus.CounterValue,
		float64(v.FreeSpaceScansPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsFullScans,
		prometheus.CounterValue,
		float64(v.FullScansPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsIndexSearches,
		prometheus.CounterValue,
		float64(v.IndexSearchesPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsInSysXactwaits,
		prometheus.CounterValue,
		float64(v.InSysXactwaitsPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsLobHandleCreateCount,
		prometheus.CounterValue,
		float64(v.LobHandleCreateCount),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsLobHandleDestroyCount,
		prometheus.CounterValue,
		float64(v.LobHandleDestroyCount),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsLobSSProviderCreateCount,
		prometheus.CounterValue,
		float64(v.LobSSProviderCreateCount),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsLobSSProviderDestroyCount,
		prometheus.CounterValue,
		float64(v.LobSSProviderDestroyCount),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsLobSSProviderTruncationCount,
		prometheus.CounterValue,
		float64(v.LobSSProviderTruncationCount),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsMixedpageallocations,
		prometheus.CounterValue,
		float64(v.MixedpageallocationsPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsPagecompressionattempts,
		prometheus.CounterValue,
		float64(v.PagecompressionattemptsPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsPageDeallocations,
		prometheus.CounterValue,
		float64(v.PageDeallocationsPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsPagesAllocated,
		prometheus.CounterValue,
		float64(v.PagesAllocatedPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsPagescompressed,
		prometheus.CounterValue,
		float64(v.PagescompressedPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsPageSplits,
		prometheus.CounterValue,
		float64(v.PageSplitsPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsProbeScans,
		prometheus.CounterValue,
		float64(v.ProbeScansPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsRangeScans,
		prometheus.CounterValue,
		float64(v.RangeScansPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsScanPointRevalidations,
		prometheus.CounterValue,
		float64(v.ScanPointRevalidationsPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsSkippedGhostedRecords,
		prometheus.CounterValue,
		float64(v.SkippedGhostedRecordsPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsTableLockEscalations,
		prometheus.CounterValue,
		float64(v.TableLockEscalationsPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsUsedleafpagecookie,
		prometheus.CounterValue,
		float64(v.Usedleafpagecookie),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsUsedtreepagecookie,
		prometheus.CounterValue,
		float64(v.Usedtreepagecookie),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsWorkfilesCreated,
		prometheus.CounterValue,
		float64(v.WorkfilesCreatedPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsWorktablesCreated,
		prometheus.CounterValue,
		float64(v.WorktablesCreatedPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMethodsWorktablesFromCacheRatio,
		prometheus.CounterValue,
		float64(v.WorktablesFromCacheRatio),
		sqlInstance,
	)
	return nil, nil
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
	log.Debugf("mssql_availreplica collector iterating sql instance %s.", sqlInstance)

	class := mssqlBuildWMIInstanceClass("AvailabilityReplica", sqlInstance)
	q := queryAllForClassWhere(&dst, class, `Name <> '_Total'`)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	for _, v := range dst {
		replicaName := v.Name

		ch <- prometheus.MustNewConstMetric(
			c.AvailReplicaBytesReceivedfromReplica,
			prometheus.CounterValue,
			float64(v.BytesReceivedfromReplicaPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AvailReplicaBytesSenttoReplica,
			prometheus.CounterValue,
			float64(v.BytesSenttoReplicaPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AvailReplicaBytesSenttoTransport,
			prometheus.CounterValue,
			float64(v.BytesSenttoTransportPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AvailReplicaFlowControl,
			prometheus.CounterValue,
			float64(v.FlowControlPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AvailReplicaFlowControlTimems,
			prometheus.CounterValue,
			float64(v.FlowControlTimemsPersec)/1000.0,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AvailReplicaReceivesfromReplica,
			prometheus.CounterValue,
			float64(v.ReceivesfromReplicaPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AvailReplicaResentMessages,
			prometheus.CounterValue,
			float64(v.ResentMessagesPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AvailReplicaSendstoReplica,
			prometheus.CounterValue,
			float64(v.SendstoReplicaPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AvailReplicaSendstoTransport,
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
	log.Debugf("mssql_bufman collector iterating sql instance %s.", sqlInstance)

	class := mssqlBuildWMIInstanceClass("BufferManager", sqlInstance)
	q := queryAllForClass(&dst, class)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}
	if len(dst) == 0 {
		return nil, errors.New("WMI query returned empty result set")
	}

	v := dst[0]

	ch <- prometheus.MustNewConstMetric(
		c.BufManBackgroundwriterpages,
		prometheus.CounterValue,
		float64(v.BackgroundwriterpagesPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.BufManBuffercachehitratio,
		prometheus.GaugeValue,
		float64(v.Buffercachehitratio),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.BufManCheckpointpages,
		prometheus.CounterValue,
		float64(v.CheckpointpagesPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.BufManDatabasepages,
		prometheus.GaugeValue,
		float64(v.Databasepages),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.BufManExtensionallocatedpages,
		prometheus.GaugeValue,
		float64(v.Extensionallocatedpages),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.BufManExtensionfreepages,
		prometheus.GaugeValue,
		float64(v.Extensionfreepages),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.BufManExtensioninuseaspercentage,
		prometheus.GaugeValue,
		float64(v.Extensioninuseaspercentage),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.BufManExtensionoutstandingIOcounter,
		prometheus.GaugeValue,
		float64(v.ExtensionoutstandingIOcounter),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.BufManExtensionpageevictions,
		prometheus.CounterValue,
		float64(v.ExtensionpageevictionsPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.BufManExtensionpagereads,
		prometheus.CounterValue,
		float64(v.ExtensionpagereadsPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.BufManExtensionpageunreferencedtime,
		prometheus.GaugeValue,
		float64(v.Extensionpageunreferencedtime),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.BufManExtensionpagewrites,
		prometheus.CounterValue,
		float64(v.ExtensionpagewritesPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.BufManFreeliststalls,
		prometheus.CounterValue,
		float64(v.FreeliststallsPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.BufManIntegralControllerSlope,
		prometheus.GaugeValue,
		float64(v.IntegralControllerSlope),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.BufManLazywrites,
		prometheus.CounterValue,
		float64(v.LazywritesPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.BufManPagelifeexpectancy,
		prometheus.GaugeValue,
		float64(v.Pagelifeexpectancy),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.BufManPagelookups,
		prometheus.CounterValue,
		float64(v.PagelookupsPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.BufManPagereads,
		prometheus.CounterValue,
		float64(v.PagereadsPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.BufManPagewrites,
		prometheus.CounterValue,
		float64(v.PagewritesPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.BufManReadaheadpages,
		prometheus.CounterValue,
		float64(v.ReadaheadpagesPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.BufManReadaheadtime,
		prometheus.CounterValue,
		float64(v.ReadaheadtimePersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.BufManTargetpages,
		prometheus.GaugeValue,
		float64(v.Targetpages),
		sqlInstance,
	)

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
	log.Debugf("mssql_dbreplica collector iterating sql instance %s.", sqlInstance)

	class := mssqlBuildWMIInstanceClass("DatabaseReplica", sqlInstance)
	q := queryAllForClassWhere(&dst, class, `Name <> '_Total'`)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	for _, v := range dst {
		replicaName := v.Name

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaDatabaseFlowControlDelay,
			prometheus.GaugeValue,
			float64(v.DatabaseFlowControlDelay),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaDatabaseFlowControls,
			prometheus.CounterValue,
			float64(v.DatabaseFlowControlsPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaFileBytesReceived,
			prometheus.CounterValue,
			float64(v.FileBytesReceivedPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaGroupCommits,
			prometheus.CounterValue,
			float64(v.GroupCommitsPerSec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaGroupCommitTime,
			prometheus.GaugeValue,
			float64(v.GroupCommitTime),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaLogApplyPendingQueue,
			prometheus.GaugeValue,
			float64(v.LogApplyPendingQueue),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaLogApplyReadyQueue,
			prometheus.GaugeValue,
			float64(v.LogApplyReadyQueue),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaLogBytesCompressed,
			prometheus.CounterValue,
			float64(v.LogBytesCompressedPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaLogBytesDecompressed,
			prometheus.CounterValue,
			float64(v.LogBytesDecompressedPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaLogBytesReceived,
			prometheus.CounterValue,
			float64(v.LogBytesReceivedPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaLogCompressionCachehits,
			prometheus.CounterValue,
			float64(v.LogCompressionCachehitsPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaLogCompressionCachemisses,
			prometheus.CounterValue,
			float64(v.LogCompressionCachemissesPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaLogCompressions,
			prometheus.CounterValue,
			float64(v.LogCompressionsPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaLogDecompressions,
			prometheus.CounterValue,
			float64(v.LogDecompressionsPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaLogremainingforundo,
			prometheus.GaugeValue,
			float64(v.Logremainingforundo),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaLogSendQueue,
			prometheus.GaugeValue,
			float64(v.LogSendQueue),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaMirroredWriteTransactions,
			prometheus.CounterValue,
			float64(v.MirroredWriteTransactionsPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaRecoveryQueue,
			prometheus.GaugeValue,
			float64(v.RecoveryQueue),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaRedoblocked,
			prometheus.CounterValue,
			float64(v.RedoblockedPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaRedoBytesRemaining,
			prometheus.GaugeValue,
			float64(v.RedoBytesRemaining),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaRedoneBytes,
			prometheus.CounterValue,
			float64(v.RedoneBytesPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaRedones,
			prometheus.CounterValue,
			float64(v.RedonesPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaTotalLogrequiringundo,
			prometheus.GaugeValue,
			float64(v.TotalLogrequiringundo),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaTransactionDelay,
			prometheus.GaugeValue,
			float64(v.TransactionDelay)*1000.0,
			sqlInstance, replicaName,
		)
	}
	return nil, nil
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

func (c *MSSQLCollector) collectDatabases(ch chan<- prometheus.Metric, sqlInstance string) (*prometheus.Desc, error) {
	var dst []win32PerfRawDataSQLServerDatabases
	log.Debugf("mssql_databases collector iterating sql instance %s.", sqlInstance)

	class := mssqlBuildWMIInstanceClass("Databases", sqlInstance)
	q := queryAllForClassWhere(&dst, class, `Name <> '_Total'`)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	for _, v := range dst {
		dbName := v.Name

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesActiveTransactions,
			prometheus.GaugeValue,
			float64(v.ActiveTransactions),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesBackupPerRestoreThroughput,
			prometheus.CounterValue,
			float64(v.BackupPerRestoreThroughputPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesBulkCopyRows,
			prometheus.CounterValue,
			float64(v.BulkCopyRowsPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesBulkCopyThroughput,
			prometheus.CounterValue,
			float64(v.BulkCopyThroughputPersec)*1024,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesCommittableentries,
			prometheus.GaugeValue,
			float64(v.Committableentries),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesDataFilesSizeKB,
			prometheus.GaugeValue,
			float64(v.DataFilesSizeKB*1024),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesDBCCLogicalScanBytes,
			prometheus.CounterValue,
			float64(v.DBCCLogicalScanBytesPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesGroupCommitTime,
			prometheus.CounterValue,
			float64(v.GroupCommitTimePersec)/1000000.0,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogBytesFlushed,
			prometheus.CounterValue,
			float64(v.LogBytesFlushedPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogCacheHitRatio,
			prometheus.GaugeValue,
			float64(v.LogCacheHitRatio),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogCacheReads,
			prometheus.CounterValue,
			float64(v.LogCacheReadsPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogFilesSizeKB,
			prometheus.GaugeValue,
			float64(v.LogFilesSizeKB*1024),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogFilesUsedSizeKB,
			prometheus.GaugeValue,
			float64(v.LogFilesUsedSizeKB*1024),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogFlushes,
			prometheus.CounterValue,
			float64(v.LogFlushesPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogFlushWaits,
			prometheus.CounterValue,
			float64(v.LogFlushWaitsPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogFlushWaitTime,
			prometheus.GaugeValue,
			float64(v.LogFlushWaitTime)/1000.0,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogFlushWriteTimems,
			prometheus.GaugeValue,
			float64(v.LogFlushWriteTimems)/1000.0,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogGrowths,
			prometheus.GaugeValue,
			float64(v.LogGrowths),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogPoolCacheMisses,
			prometheus.CounterValue,
			float64(v.LogPoolCacheMissesPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogPoolDiskReads,
			prometheus.CounterValue,
			float64(v.LogPoolDiskReadsPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogPoolHashDeletes,
			prometheus.CounterValue,
			float64(v.LogPoolHashDeletesPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogPoolHashInserts,
			prometheus.CounterValue,
			float64(v.LogPoolHashInsertsPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogPoolInvalidHashEntry,
			prometheus.CounterValue,
			float64(v.LogPoolInvalidHashEntryPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogPoolLogScanPushes,
			prometheus.CounterValue,
			float64(v.LogPoolLogScanPushesPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogPoolLogWriterPushes,
			prometheus.CounterValue,
			float64(v.LogPoolLogWriterPushesPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogPoolPushEmptyFreePool,
			prometheus.CounterValue,
			float64(v.LogPoolPushEmptyFreePoolPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogPoolPushLowMemory,
			prometheus.CounterValue,
			float64(v.LogPoolPushLowMemoryPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogPoolPushNoFreeBuffer,
			prometheus.CounterValue,
			float64(v.LogPoolPushNoFreeBufferPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogPoolReqBehindTrunc,
			prometheus.CounterValue,
			float64(v.LogPoolReqBehindTruncPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogPoolRequestsOldVLF,
			prometheus.CounterValue,
			float64(v.LogPoolRequestsOldVLFPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogPoolRequests,
			prometheus.CounterValue,
			float64(v.LogPoolRequestsPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogPoolTotalActiveLogSize,
			prometheus.GaugeValue,
			float64(v.LogPoolTotalActiveLogSize),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogPoolTotalSharedPoolSize,
			prometheus.GaugeValue,
			float64(v.LogPoolTotalSharedPoolSize),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogShrinks,
			prometheus.GaugeValue,
			float64(v.LogShrinks),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogTruncations,
			prometheus.GaugeValue,
			float64(v.LogTruncations),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesPercentLogUsed,
			prometheus.GaugeValue,
			float64(v.PercentLogUsed),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesReplPendingXacts,
			prometheus.GaugeValue,
			float64(v.ReplPendingXacts),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesReplTransRate,
			prometheus.CounterValue,
			float64(v.ReplTransRate),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesShrinkDataMovementBytes,
			prometheus.CounterValue,
			float64(v.ShrinkDataMovementBytesPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesTrackedtransactions,
			prometheus.CounterValue,
			float64(v.TrackedtransactionsPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesTransactions,
			prometheus.CounterValue,
			float64(v.TransactionsPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesWriteTransactions,
			prometheus.CounterValue,
			float64(v.WriteTransactionsPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesXTPControllerDLCLatencyPerFetch,
			prometheus.GaugeValue,
			float64(v.XTPControllerDLCLatencyPerFetch),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesXTPControllerDLCPeakLatency,
			prometheus.GaugeValue,
			float64(v.XTPControllerDLCPeakLatency)*1000000.0,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesXTPControllerLogProcessed,
			prometheus.CounterValue,
			float64(v.XTPControllerLogProcessedPersec),
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesXTPMemoryUsedKB,
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
	log.Debugf("mssql_genstats collector iterating sql instance %s.", sqlInstance)

	class := mssqlBuildWMIInstanceClass("GeneralStatistics", sqlInstance)
	q := queryAllForClass(&dst, class)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}
	if len(dst) == 0 {
		return nil, errors.New("WMI query returned empty result set")
	}

	v := dst[0]
	ch <- prometheus.MustNewConstMetric(
		c.GenStatsActiveTempTables,
		prometheus.GaugeValue,
		float64(v.ActiveTempTables),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.GenStatsConnectionReset,
		prometheus.CounterValue,
		float64(v.ConnectionResetPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.GenStatsEventNotificationsDelayedDrop,
		prometheus.GaugeValue,
		float64(v.EventNotificationsDelayedDrop),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.GenStatsHTTPAuthenticatedRequests,
		prometheus.GaugeValue,
		float64(v.HTTPAuthenticatedRequests),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.GenStatsLogicalConnections,
		prometheus.GaugeValue,
		float64(v.LogicalConnections),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.GenStatsLogins,
		prometheus.CounterValue,
		float64(v.LoginsPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.GenStatsLogouts,
		prometheus.CounterValue,
		float64(v.LogoutsPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.GenStatsMarsDeadlocks,
		prometheus.GaugeValue,
		float64(v.MarsDeadlocks),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.GenStatsNonatomicyieldrate,
		prometheus.CounterValue,
		float64(v.Nonatomicyieldrate),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.GenStatsProcessesblocked,
		prometheus.GaugeValue,
		float64(v.Processesblocked),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.GenStatsSOAPEmptyRequests,
		prometheus.GaugeValue,
		float64(v.SOAPEmptyRequests),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.GenStatsSOAPMethodInvocations,
		prometheus.GaugeValue,
		float64(v.SOAPMethodInvocations),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.GenStatsSOAPSessionInitiateRequests,
		prometheus.GaugeValue,
		float64(v.SOAPSessionInitiateRequests),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.GenStatsSOAPSessionTerminateRequests,
		prometheus.GaugeValue,
		float64(v.SOAPSessionTerminateRequests),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.GenStatsSOAPSQLRequests,
		prometheus.GaugeValue,
		float64(v.SOAPSQLRequests),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.GenStatsSOAPWSDLRequests,
		prometheus.GaugeValue,
		float64(v.SOAPWSDLRequests),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.GenStatsSQLTraceIOProviderLockWaits,
		prometheus.GaugeValue,
		float64(v.SQLTraceIOProviderLockWaits),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.GenStatsTempdbrecoveryunitid,
		prometheus.GaugeValue,
		float64(v.Tempdbrecoveryunitid),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.GenStatsTempdbrowsetid,
		prometheus.GaugeValue,
		float64(v.Tempdbrowsetid),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.GenStatsTempTablesCreationRate,
		prometheus.CounterValue,
		float64(v.TempTablesCreationRate),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.GenStatsTempTablesForDestruction,
		prometheus.GaugeValue,
		float64(v.TempTablesForDestruction),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.GenStatsTraceEventNotificationQueue,
		prometheus.GaugeValue,
		float64(v.TraceEventNotificationQueue),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.GenStatsTransactions,
		prometheus.GaugeValue,
		float64(v.Transactions),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.GenStatsUserConnections,
		prometheus.GaugeValue,
		float64(v.UserConnections),
		sqlInstance,
	)

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
	log.Debugf("mssql_locks collector iterating sql instance %s.", sqlInstance)

	class := mssqlBuildWMIInstanceClass("Locks", sqlInstance)
	q := queryAllForClassWhere(&dst, class, `Name <> '_Total'`)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	for _, v := range dst {
		lockResourceName := v.Name

		ch <- prometheus.MustNewConstMetric(
			c.LocksAverageWaitTimems,
			prometheus.GaugeValue,
			float64(v.AverageWaitTimems)/1000.0,
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LocksLockRequests,
			prometheus.CounterValue,
			float64(v.LockRequestsPersec),
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LocksLockTimeouts,
			prometheus.CounterValue,
			float64(v.LockTimeoutsPersec),
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LocksLockTimeoutstimeout0,
			prometheus.CounterValue,
			float64(v.LockTimeoutstimeout0Persec),
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LocksLockWaits,
			prometheus.CounterValue,
			float64(v.LockWaitsPersec),
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LocksLockWaitTimems,
			prometheus.GaugeValue,
			float64(v.LockWaitTimems)/1000.0,
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LocksNumberofDeadlocks,
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
	log.Debugf("mssql_memmgr collector iterating sql instance %s.", sqlInstance)

	class := mssqlBuildWMIInstanceClass("MemoryManager", sqlInstance)
	q := queryAllForClass(&dst, class)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}
	if len(dst) == 0 {
		return nil, errors.New("WMI query returned empty result set")
	}

	v := dst[0]

	ch <- prometheus.MustNewConstMetric(
		c.MemMgrConnectionMemoryKB,
		prometheus.GaugeValue,
		float64(v.ConnectionMemoryKB*1024),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemMgrDatabaseCacheMemoryKB,
		prometheus.GaugeValue,
		float64(v.DatabaseCacheMemoryKB*1024),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemMgrExternalbenefitofmemory,
		prometheus.GaugeValue,
		float64(v.Externalbenefitofmemory),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemMgrFreeMemoryKB,
		prometheus.GaugeValue,
		float64(v.FreeMemoryKB*1024),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemMgrGrantedWorkspaceMemoryKB,
		prometheus.GaugeValue,
		float64(v.GrantedWorkspaceMemoryKB*1024),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemMgrLockBlocks,
		prometheus.GaugeValue,
		float64(v.LockBlocks),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemMgrLockBlocksAllocated,
		prometheus.GaugeValue,
		float64(v.LockBlocksAllocated),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemMgrLockMemoryKB,
		prometheus.GaugeValue,
		float64(v.LockMemoryKB*1024),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemMgrLockOwnerBlocks,
		prometheus.GaugeValue,
		float64(v.LockOwnerBlocks),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemMgrLockOwnerBlocksAllocated,
		prometheus.GaugeValue,
		float64(v.LockOwnerBlocksAllocated),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemMgrLogPoolMemoryKB,
		prometheus.GaugeValue,
		float64(v.LogPoolMemoryKB*1024),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemMgrMaximumWorkspaceMemoryKB,
		prometheus.GaugeValue,
		float64(v.MaximumWorkspaceMemoryKB*1024),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemMgrMemoryGrantsOutstanding,
		prometheus.GaugeValue,
		float64(v.MemoryGrantsOutstanding),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemMgrMemoryGrantsPending,
		prometheus.GaugeValue,
		float64(v.MemoryGrantsPending),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemMgrOptimizerMemoryKB,
		prometheus.GaugeValue,
		float64(v.OptimizerMemoryKB*1024),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemMgrReservedServerMemoryKB,
		prometheus.GaugeValue,
		float64(v.ReservedServerMemoryKB*1024),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemMgrSQLCacheMemoryKB,
		prometheus.GaugeValue,
		float64(v.SQLCacheMemoryKB*1024),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemMgrStolenServerMemoryKB,
		prometheus.GaugeValue,
		float64(v.StolenServerMemoryKB*1024),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemMgrTargetServerMemoryKB,
		prometheus.GaugeValue,
		float64(v.TargetServerMemoryKB*1024),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemMgrTotalServerMemoryKB,
		prometheus.GaugeValue,
		float64(v.TotalServerMemoryKB*1024),
		sqlInstance,
	)

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
	log.Debugf("mssql_sqlstats collector iterating sql instance %s.", sqlInstance)

	class := mssqlBuildWMIInstanceClass("SQLStatistics", sqlInstance)
	q := queryAllForClass(&dst, class)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	if len(dst) == 0 {
		return nil, errors.New("WMI query returned empty result set")
	}

	v := dst[0]

	ch <- prometheus.MustNewConstMetric(
		c.SQLStatsAutoParamAttempts,
		prometheus.CounterValue,
		float64(v.AutoParamAttemptsPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.SQLStatsBatchRequests,
		prometheus.CounterValue,
		float64(v.BatchRequestsPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.SQLStatsFailedAutoParams,
		prometheus.CounterValue,
		float64(v.FailedAutoParamsPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.SQLStatsForcedParameterizations,
		prometheus.CounterValue,
		float64(v.ForcedParameterizationsPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.SQLStatsGuidedplanexecutions,
		prometheus.CounterValue,
		float64(v.GuidedplanexecutionsPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.SQLStatsMisguidedplanexecutions,
		prometheus.CounterValue,
		float64(v.MisguidedplanexecutionsPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.SQLStatsSafeAutoParams,
		prometheus.CounterValue,
		float64(v.SafeAutoParamsPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.SQLStatsSQLAttentionrate,
		prometheus.CounterValue,
		float64(v.SQLAttentionrate),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.SQLStatsSQLCompilations,
		prometheus.CounterValue,
		float64(v.SQLCompilationsPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.SQLStatsSQLReCompilations,
		prometheus.CounterValue,
		float64(v.SQLReCompilationsPersec),
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.SQLStatsUnsafeAutoParams,
		prometheus.CounterValue,
		float64(v.UnsafeAutoParamsPersec),
		sqlInstance,
	)

	return nil, nil
}

//go:build windows

package mssql

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows/registry"
)

const Name = "mssql"

type Config struct {
	EnabledCollectors string `yaml:"collectors_enabled"`
}

var ConfigDefaults = Config{
	EnabledCollectors: "accessmethods,availreplica,bufman,databases,dbreplica,genstats,locks,memmgr,sqlstats,sqlerrors,transactions,waitstats",
}

type mssqlInstancesType map[string]string

func getMSSQLInstances(logger log.Logger) mssqlInstancesType {
	sqlInstances := make(mssqlInstancesType)

	// in case querying the registry fails, return the default instance
	sqlDefaultInstance := make(mssqlInstancesType)
	sqlDefaultInstance["MSSQLSERVER"] = ""

	regkey := `Software\Microsoft\Microsoft SQL Server\Instance Names\SQL`
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, regkey, registry.QUERY_VALUE)
	if err != nil {
		_ = level.Warn(logger).Log("msg", "Couldn't open registry to determine SQL instances", "err", err)
		return sqlDefaultInstance
	}
	defer func() {
		err = k.Close()
		if err != nil {
			_ = level.Warn(logger).Log("msg", "Failed to close registry key", "err", err)
		}
	}()

	instanceNames, err := k.ReadValueNames(0)
	if err != nil {
		_ = level.Warn(logger).Log("msg", "Can't ReadSubKeyNames", "err", err)
		return sqlDefaultInstance
	}

	for _, instanceName := range instanceNames {
		if instanceVersion, _, err := k.GetStringValue(instanceName); err == nil {
			sqlInstances[instanceName] = instanceVersion
		}
	}

	_ = level.Debug(logger).Log("msg", fmt.Sprintf("Detected MSSQL Instances: %#v\n", sqlInstances))

	return sqlInstances
}

type mssqlCollectorsMap map[string]mssqlCollectorFunc

func (c *Collector) getMSSQLCollectors() mssqlCollectorsMap {
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
	mssqlCollectors["sqlerrors"] = c.collectSQLErrors
	mssqlCollectors["transactions"] = c.collectTransactions
	mssqlCollectors["waitstats"] = c.collectWaitStats

	return mssqlCollectors
}

// mssqlGetPerfObjectName - Returns the name of the Windows Performance
// Counter object for the given SQL instance and Collector.
func mssqlGetPerfObjectName(sqlInstance string, collector string) string {
	prefix := "SQLServer:"
	if sqlInstance != "MSSQLSERVER" {
		prefix = "MSSQL$" + sqlInstance + ":"
	}
	suffix := ""
	switch collector {
	case "accessmethods":
		suffix = "Access Methods"
	case "availreplica":
		suffix = "Availability Replica"
	case "bufman":
		suffix = "Buffer Manager"
	case "databases":
		suffix = "Databases"
	case "dbreplica":
		suffix = "Database Replica"
	case "genstats":
		suffix = "General Statistics"
	case "locks":
		suffix = "Locks"
	case "memmgr":
		suffix = "Memory Manager"
	case "sqlerrors":
		suffix = "SQL Errors"
	case "sqlstats":
		suffix = "SQL Statistics"
	case "transactions":
		suffix = "Transactions"
	case "waitstats":
		suffix = "Wait Statistics"
	}
	return prefix + suffix
}

// A Collector is a Prometheus Collector for various WMI Win32_PerfRawData_MSSQLSERVER_* metrics
type Collector struct {
	logger log.Logger

	mssqlEnabledCollectors *string
	mssqlPrintCollectors   *bool

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
	AccessMethodsWorktablesFromCacheHits      *prometheus.Desc
	AccessMethodsWorktablesFromCacheLookups   *prometheus.Desc

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
	BufManBuffercachehits               *prometheus.Desc
	BufManBuffercachelookups            *prometheus.Desc
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
	DatabasesActiveParallelredothreads       *prometheus.Desc
	DatabasesActiveTransactions              *prometheus.Desc
	DatabasesBackupPerRestoreThroughput      *prometheus.Desc
	DatabasesBulkCopyRows                    *prometheus.Desc
	DatabasesBulkCopyThroughput              *prometheus.Desc
	DatabasesCommittableentries              *prometheus.Desc
	DatabasesDataFilesSizeKB                 *prometheus.Desc
	DatabasesDBCCLogicalScanBytes            *prometheus.Desc
	DatabasesGroupCommitTime                 *prometheus.Desc
	DatabasesLogBytesFlushed                 *prometheus.Desc
	DatabasesLogCacheHits                    *prometheus.Desc
	DatabasesLogCacheLookups                 *prometheus.Desc
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
	LocksWaitTime             *prometheus.Desc
	LocksCount                *prometheus.Desc
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

	// Win32_PerfRawData_{instance}_SQLServerSQLErrors
	SQLErrorsTotal *prometheus.Desc

	// Win32_PerfRawData_{instance}_SQLServerTransactions
	TransactionsTempDbFreeSpaceBytes             *prometheus.Desc
	TransactionsLongestTransactionRunningSeconds *prometheus.Desc
	TransactionsNonSnapshotVersionActiveTotal    *prometheus.Desc
	TransactionsSnapshotActiveTotal              *prometheus.Desc
	TransactionsActive                           *prometheus.Desc
	TransactionsUpdateConflictsTotal             *prometheus.Desc
	TransactionsUpdateSnapshotActiveTotal        *prometheus.Desc
	TransactionsVersionCleanupRateBytes          *prometheus.Desc
	TransactionsVersionGenerationRateBytes       *prometheus.Desc
	TransactionsVersionStoreSizeBytes            *prometheus.Desc
	TransactionsVersionStoreUnits                *prometheus.Desc
	TransactionsVersionStoreCreationUnits        *prometheus.Desc
	TransactionsVersionStoreTruncationUnits      *prometheus.Desc

	// Win32_PerfRawData_{instance}_SQLServerWaitStatistics
	WaitStatsLockWaits                     *prometheus.Desc
	WaitStatsMemoryGrantQueueWaits         *prometheus.Desc
	WaitStatsThreadSafeMemoryObjectsWaits  *prometheus.Desc
	WaitStatsLogWriteWaits                 *prometheus.Desc
	WaitStatsLogBufferWaits                *prometheus.Desc
	WaitStatsNetworkIOWaits                *prometheus.Desc
	WaitStatsPageIOLatchWaits              *prometheus.Desc
	WaitStatsPageLatchWaits                *prometheus.Desc
	WaitStatsNonpageLatchWaits             *prometheus.Desc
	WaitStatsWaitForTheWorkerWaits         *prometheus.Desc
	WaitStatsWorkspaceSynchronizationWaits *prometheus.Desc
	WaitStatsTransactionOwnershipWaits     *prometheus.Desc

	mssqlInstances             mssqlInstancesType
	mssqlCollectors            mssqlCollectorsMap
	mssqlChildCollectorFailure int
}

func New(logger log.Logger, config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	printCollectors := false

	c := &Collector{
		mssqlEnabledCollectors: &config.EnabledCollectors,
		mssqlPrintCollectors:   &printCollectors,
	}
	c.SetLogger(logger)

	return c
}

func NewWithFlags(app *kingpin.Application) *Collector {
	return &Collector{
		mssqlEnabledCollectors: app.Flag(
			"collectors.mssql.classes-enabled",
			"Comma-separated list of mssql WMI classes to use.").
			Default(ConfigDefaults.EnabledCollectors).String(),

		mssqlPrintCollectors: app.Flag(
			"collectors.mssql.class-print",
			"If true, print available mssql WMI classes and exit.  Only displays if the mssql collector is enabled.",
		).Bool(),
	}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *Collector) GetPerfCounter() ([]string, error) {
	enabled := utils.ExpandEnabledChildCollectors(*c.mssqlEnabledCollectors)
	c.mssqlInstances = getMSSQLInstances(c.logger)
	perfCounters := make([]string, 0, len(c.mssqlInstances)*len(enabled))
	for instance := range c.mssqlInstances {
		for _, c := range enabled {
			perfCounters = append(perfCounters, mssqlGetPerfObjectName(instance, c))
		}
	}

	return perfCounters, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build() error {
	// meta
	c.mssqlScrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "collector_duration_seconds"),
		"windows_exporter: Duration of an mssql child collection.",
		[]string{"collector", "mssql_instance"},
		nil,
	)
	c.mssqlScrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "collector_success"),
		"windows_exporter: Whether a mssql child collector was successful.",
		[]string{"collector", "mssql_instance"},
		nil,
	)

	// Win32_PerfRawData_{instance}_SQLServerAccessMethods
	c.AccessMethodsAUcleanupbatches = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_au_batch_cleanups"),
		"(AccessMethods.AUcleanupbatches)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsAUcleanups = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_au_cleanups"),
		"(AccessMethods.AUcleanups)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsByreferenceLobCreateCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_by_reference_lob_creates"),
		"(AccessMethods.ByreferenceLobCreateCount)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsByreferenceLobUseCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_by_reference_lob_uses"),
		"(AccessMethods.ByreferenceLobUseCount)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsCountLobReadahead = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_lob_read_aheads"),
		"(AccessMethods.CountLobReadahead)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsCountPullInRow = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_column_value_pulls"),
		"(AccessMethods.CountPullInRow)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsCountPushOffRow = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_column_value_pushes"),
		"(AccessMethods.CountPushOffRow)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsDeferreddroppedAUs = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_deferred_dropped_aus"),
		"(AccessMethods.DeferreddroppedAUs)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsDeferredDroppedrowsets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_deferred_dropped_rowsets"),
		"(AccessMethods.DeferredDroppedrowsets)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsDroppedrowsetcleanups = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_dropped_rowset_cleanups"),
		"(AccessMethods.Droppedrowsetcleanups)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsDroppedrowsetsskipped = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_dropped_rowset_skips"),
		"(AccessMethods.Droppedrowsetsskipped)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsExtentDeallocations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_extent_deallocations"),
		"(AccessMethods.ExtentDeallocations)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsExtentsAllocated = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_extent_allocations"),
		"(AccessMethods.ExtentsAllocated)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsFailedAUcleanupbatches = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_au_batch_cleanup_failures"),
		"(AccessMethods.FailedAUcleanupbatches)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsFailedleafpagecookie = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_leaf_page_cookie_failures"),
		"(AccessMethods.Failedleafpagecookie)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsFailedtreepagecookie = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_tree_page_cookie_failures"),
		"(AccessMethods.Failedtreepagecookie)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsForwardedRecords = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_forwarded_records"),
		"(AccessMethods.ForwardedRecords)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsFreeSpacePageFetches = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_free_space_page_fetches"),
		"(AccessMethods.FreeSpacePageFetches)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsFreeSpaceScans = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_free_space_scans"),
		"(AccessMethods.FreeSpaceScans)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsFullScans = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_full_scans"),
		"(AccessMethods.FullScans)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsIndexSearches = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_index_searches"),
		"(AccessMethods.IndexSearches)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsInSysXactwaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_insysxact_waits"),
		"(AccessMethods.InSysXactwaits)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsLobHandleCreateCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_lob_handle_creates"),
		"(AccessMethods.LobHandleCreateCount)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsLobHandleDestroyCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_lob_handle_destroys"),
		"(AccessMethods.LobHandleDestroyCount)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsLobSSProviderCreateCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_lob_ss_provider_creates"),
		"(AccessMethods.LobSSProviderCreateCount)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsLobSSProviderDestroyCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_lob_ss_provider_destroys"),
		"(AccessMethods.LobSSProviderDestroyCount)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsLobSSProviderTruncationCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_lob_ss_provider_truncations"),
		"(AccessMethods.LobSSProviderTruncationCount)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsMixedpageallocations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_mixed_page_allocations"),
		"(AccessMethods.MixedpageallocationsPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsPagecompressionattempts = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_page_compression_attempts"),
		"(AccessMethods.PagecompressionattemptsPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsPageDeallocations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_page_deallocations"),
		"(AccessMethods.PageDeallocationsPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsPagesAllocated = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_page_allocations"),
		"(AccessMethods.PagesAllocatedPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsPagescompressed = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_page_compressions"),
		"(AccessMethods.PagescompressedPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsPageSplits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_page_splits"),
		"(AccessMethods.PageSplitsPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsProbeScans = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_probe_scans"),
		"(AccessMethods.ProbeScansPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsRangeScans = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_range_scans"),
		"(AccessMethods.RangeScansPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsScanPointRevalidations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_scan_point_revalidations"),
		"(AccessMethods.ScanPointRevalidationsPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsSkippedGhostedRecords = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_ghost_record_skips"),
		"(AccessMethods.SkippedGhostedRecordsPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsTableLockEscalations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_table_lock_escalations"),
		"(AccessMethods.TableLockEscalationsPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsUsedleafpagecookie = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_leaf_page_cookie_uses"),
		"(AccessMethods.Usedleafpagecookie)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsUsedtreepagecookie = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_tree_page_cookie_uses"),
		"(AccessMethods.Usedtreepagecookie)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsWorkfilesCreated = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_workfile_creates"),
		"(AccessMethods.WorkfilesCreatedPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsWorktablesCreated = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_worktables_creates"),
		"(AccessMethods.WorktablesCreatedPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsWorktablesFromCacheHits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_worktables_from_cache_hits"),
		"(AccessMethods.WorktablesFromCacheRatio)",
		[]string{"mssql_instance"},
		nil,
	)
	c.AccessMethodsWorktablesFromCacheLookups = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_worktables_from_cache_lookups"),
		"(AccessMethods.WorktablesFromCacheRatio_Base)",
		[]string{"mssql_instance"},
		nil,
	)

	// Win32_PerfRawData_{instance}_SQLServerAvailabilityReplica
	c.AvailReplicaBytesReceivedfromReplica = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "availreplica_received_from_replica_bytes"),
		"(AvailabilityReplica.BytesReceivedfromReplica)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.AvailReplicaBytesSenttoReplica = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "availreplica_sent_to_replica_bytes"),
		"(AvailabilityReplica.BytesSenttoReplica)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.AvailReplicaBytesSenttoTransport = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "availreplica_sent_to_transport_bytes"),
		"(AvailabilityReplica.BytesSenttoTransport)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.AvailReplicaFlowControl = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "availreplica_initiated_flow_controls"),
		"(AvailabilityReplica.FlowControl)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.AvailReplicaFlowControlTimems = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "availreplica_flow_control_wait_seconds"),
		"(AvailabilityReplica.FlowControlTimems)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.AvailReplicaReceivesfromReplica = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "availreplica_receives_from_replica"),
		"(AvailabilityReplica.ReceivesfromReplica)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.AvailReplicaResentMessages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "availreplica_resent_messages"),
		"(AvailabilityReplica.ResentMessages)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.AvailReplicaSendstoReplica = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "availreplica_sends_to_replica"),
		"(AvailabilityReplica.SendstoReplica)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.AvailReplicaSendstoTransport = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "availreplica_sends_to_transport"),
		"(AvailabilityReplica.SendstoTransport)",
		[]string{"mssql_instance", "replica"},
		nil,
	)

	// Win32_PerfRawData_{instance}_SQLServerBufferManager
	c.BufManBackgroundwriterpages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_background_writer_pages"),
		"(BufferManager.Backgroundwriterpages)",
		[]string{"mssql_instance"},
		nil,
	)
	c.BufManBuffercachehits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_buffer_cache_hits"),
		"(BufferManager.Buffercachehitratio)",
		[]string{"mssql_instance"},
		nil,
	)
	c.BufManBuffercachelookups = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_buffer_cache_lookups"),
		"(BufferManager.Buffercachehitratio_Base)",
		[]string{"mssql_instance"},
		nil,
	)
	c.BufManCheckpointpages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_checkpoint_pages"),
		"(BufferManager.Checkpointpages)",
		[]string{"mssql_instance"},
		nil,
	)
	c.BufManDatabasepages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_database_pages"),
		"(BufferManager.Databasepages)",
		[]string{"mssql_instance"},
		nil,
	)
	c.BufManExtensionallocatedpages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_extension_allocated_pages"),
		"(BufferManager.Extensionallocatedpages)",
		[]string{"mssql_instance"},
		nil,
	)
	c.BufManExtensionfreepages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_extension_free_pages"),
		"(BufferManager.Extensionfreepages)",
		[]string{"mssql_instance"},
		nil,
	)
	c.BufManExtensioninuseaspercentage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_extension_in_use_as_percentage"),
		"(BufferManager.Extensioninuseaspercentage)",
		[]string{"mssql_instance"},
		nil,
	)
	c.BufManExtensionoutstandingIOcounter = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_extension_outstanding_io"),
		"(BufferManager.ExtensionoutstandingIOcounter)",
		[]string{"mssql_instance"},
		nil,
	)
	c.BufManExtensionpageevictions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_extension_page_evictions"),
		"(BufferManager.Extensionpageevictions)",
		[]string{"mssql_instance"},
		nil,
	)
	c.BufManExtensionpagereads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_extension_page_reads"),
		"(BufferManager.Extensionpagereads)",
		[]string{"mssql_instance"},
		nil,
	)
	c.BufManExtensionpageunreferencedtime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_extension_page_unreferenced_seconds"),
		"(BufferManager.Extensionpageunreferencedtime)",
		[]string{"mssql_instance"},
		nil,
	)
	c.BufManExtensionpagewrites = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_extension_page_writes"),
		"(BufferManager.Extensionpagewrites)",
		[]string{"mssql_instance"},
		nil,
	)
	c.BufManFreeliststalls = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_free_list_stalls"),
		"(BufferManager.Freeliststalls)",
		[]string{"mssql_instance"},
		nil,
	)
	c.BufManIntegralControllerSlope = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_integral_controller_slope"),
		"(BufferManager.IntegralControllerSlope)",
		[]string{"mssql_instance"},
		nil,
	)
	c.BufManLazywrites = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_lazywrites"),
		"(BufferManager.Lazywrites)",
		[]string{"mssql_instance"},
		nil,
	)
	c.BufManPagelifeexpectancy = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_page_life_expectancy_seconds"),
		"(BufferManager.Pagelifeexpectancy)",
		[]string{"mssql_instance"},
		nil,
	)
	c.BufManPagelookups = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_page_lookups"),
		"(BufferManager.Pagelookups)",
		[]string{"mssql_instance"},
		nil,
	)
	c.BufManPagereads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_page_reads"),
		"(BufferManager.Pagereads)",
		[]string{"mssql_instance"},
		nil,
	)
	c.BufManPagewrites = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_page_writes"),
		"(BufferManager.Pagewrites)",
		[]string{"mssql_instance"},
		nil,
	)
	c.BufManReadaheadpages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_read_ahead_pages"),
		"(BufferManager.Readaheadpages)",
		[]string{"mssql_instance"},
		nil,
	)
	c.BufManReadaheadtime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_read_ahead_issuing_seconds"),
		"(BufferManager.Readaheadtime)",
		[]string{"mssql_instance"},
		nil,
	)
	c.BufManTargetpages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_target_pages"),
		"(BufferManager.Targetpages)",
		[]string{"mssql_instance"},
		nil,
	)

	// Win32_PerfRawData_{instance}_SQLServerDatabaseReplica
	c.DBReplicaDatabaseFlowControlDelay = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_database_flow_control_wait_seconds"),
		"(DatabaseReplica.DatabaseFlowControlDelay)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.DBReplicaDatabaseFlowControls = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_database_initiated_flow_controls"),
		"(DatabaseReplica.DatabaseFlowControls)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.DBReplicaFileBytesReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_received_file_bytes"),
		"(DatabaseReplica.FileBytesReceived)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.DBReplicaGroupCommits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_group_commits"),
		"(DatabaseReplica.GroupCommits)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.DBReplicaGroupCommitTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_group_commit_stall_seconds"),
		"(DatabaseReplica.GroupCommitTime)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.DBReplicaLogApplyPendingQueue = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_log_apply_pending_queue"),
		"(DatabaseReplica.LogApplyPendingQueue)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.DBReplicaLogApplyReadyQueue = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_log_apply_ready_queue"),
		"(DatabaseReplica.LogApplyReadyQueue)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.DBReplicaLogBytesCompressed = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_log_compressed_bytes"),
		"(DatabaseReplica.LogBytesCompressed)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.DBReplicaLogBytesDecompressed = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_log_decompressed_bytes"),
		"(DatabaseReplica.LogBytesDecompressed)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.DBReplicaLogBytesReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_log_received_bytes"),
		"(DatabaseReplica.LogBytesReceived)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.DBReplicaLogCompressionCachehits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_log_compression_cachehits"),
		"(DatabaseReplica.LogCompressionCachehits)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.DBReplicaLogCompressionCachemisses = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_log_compression_cachemisses"),
		"(DatabaseReplica.LogCompressionCachemisses)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.DBReplicaLogCompressions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_log_compressions"),
		"(DatabaseReplica.LogCompressions)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.DBReplicaLogDecompressions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_log_decompressions"),
		"(DatabaseReplica.LogDecompressions)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.DBReplicaLogremainingforundo = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_log_remaining_for_undo"),
		"(DatabaseReplica.Logremainingforundo)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.DBReplicaLogSendQueue = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_log_send_queue"),
		"(DatabaseReplica.LogSendQueue)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.DBReplicaMirroredWriteTransactions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_mirrored_write_transactions"),
		"(DatabaseReplica.MirroredWriteTransactions)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.DBReplicaRecoveryQueue = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_recovery_queue_records"),
		"(DatabaseReplica.RecoveryQueue)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.DBReplicaRedoblocked = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_redo_blocks"),
		"(DatabaseReplica.Redoblocked)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.DBReplicaRedoBytesRemaining = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_redo_remaining_bytes"),
		"(DatabaseReplica.RedoBytesRemaining)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.DBReplicaRedoneBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_redone_bytes"),
		"(DatabaseReplica.RedoneBytes)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.DBReplicaRedones = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_redones"),
		"(DatabaseReplica.Redones)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.DBReplicaTotalLogrequiringundo = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_total_log_requiring_undo"),
		"(DatabaseReplica.TotalLogrequiringundo)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.DBReplicaTransactionDelay = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dbreplica_transaction_delay_seconds"),
		"(DatabaseReplica.TransactionDelay)",
		[]string{"mssql_instance", "replica"},
		nil,
	)

	// Win32_PerfRawData_{instance}_SQLServerDatabases
	c.DatabasesActiveParallelredothreads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_active_parallel_redo_threads"),
		"(Databases.ActiveParallelredothreads)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesActiveTransactions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_active_transactions"),
		"(Databases.ActiveTransactions)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesBackupPerRestoreThroughput = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_backup_restore_operations"),
		"(Databases.BackupPerRestoreThroughput)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesBulkCopyRows = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_bulk_copy_rows"),
		"(Databases.BulkCopyRows)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesBulkCopyThroughput = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_bulk_copy_bytes"),
		"(Databases.BulkCopyThroughput)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesCommittableentries = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_commit_table_entries"),
		"(Databases.Committableentries)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesDataFilesSizeKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_data_files_size_bytes"),
		"(Databases.DataFilesSizeKB)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesDBCCLogicalScanBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_dbcc_logical_scan_bytes"),
		"(Databases.DBCCLogicalScanBytes)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesGroupCommitTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_group_commit_stall_seconds"),
		"(Databases.GroupCommitTime)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesLogBytesFlushed = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_flushed_bytes"),
		"(Databases.LogBytesFlushed)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesLogCacheHits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_cache_hits"),
		"(Databases.LogCacheHitRatio)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesLogCacheLookups = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_cache_lookups"),
		"(Databases.LogCacheHitRatio_Base)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesLogCacheReads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_cache_reads"),
		"(Databases.LogCacheReads)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesLogFilesSizeKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_files_size_bytes"),
		"(Databases.LogFilesSizeKB)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesLogFilesUsedSizeKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_files_used_size_bytes"),
		"(Databases.LogFilesUsedSizeKB)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesLogFlushes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_flushes"),
		"(Databases.LogFlushes)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesLogFlushWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_flush_waits"),
		"(Databases.LogFlushWaits)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesLogFlushWaitTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_flush_wait_seconds"),
		"(Databases.LogFlushWaitTime)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesLogFlushWriteTimems = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_flush_write_seconds"),
		"(Databases.LogFlushWriteTimems)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesLogGrowths = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_growths"),
		"(Databases.LogGrowths)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesLogPoolCacheMisses = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_pool_cache_misses"),
		"(Databases.LogPoolCacheMisses)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesLogPoolDiskReads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_pool_disk_reads"),
		"(Databases.LogPoolDiskReads)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesLogPoolHashDeletes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_pool_hash_deletes"),
		"(Databases.LogPoolHashDeletes)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesLogPoolHashInserts = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_pool_hash_inserts"),
		"(Databases.LogPoolHashInserts)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesLogPoolInvalidHashEntry = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_pool_invalid_hash_entries"),
		"(Databases.LogPoolInvalidHashEntry)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesLogPoolLogScanPushes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_pool_log_scan_pushes"),
		"(Databases.LogPoolLogScanPushes)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesLogPoolLogWriterPushes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_pool_log_writer_pushes"),
		"(Databases.LogPoolLogWriterPushes)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesLogPoolPushEmptyFreePool = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_pool_empty_free_pool_pushes"),
		"(Databases.LogPoolPushEmptyFreePool)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesLogPoolPushLowMemory = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_pool_low_memory_pushes"),
		"(Databases.LogPoolPushLowMemory)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesLogPoolPushNoFreeBuffer = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_pool_no_free_buffer_pushes"),
		"(Databases.LogPoolPushNoFreeBuffer)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesLogPoolReqBehindTrunc = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_pool_req_behind_trunc"),
		"(Databases.LogPoolReqBehindTrunc)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesLogPoolRequestsOldVLF = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_pool_requests_old_vlf"),
		"(Databases.LogPoolRequestsOldVLF)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesLogPoolRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_pool_requests"),
		"(Databases.LogPoolRequests)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesLogPoolTotalActiveLogSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_pool_total_active_log_bytes"),
		"(Databases.LogPoolTotalActiveLogSize)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesLogPoolTotalSharedPoolSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_pool_total_shared_pool_bytes"),
		"(Databases.LogPoolTotalSharedPoolSize)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesLogShrinks = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_shrinks"),
		"(Databases.LogShrinks)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesLogTruncations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_truncations"),
		"(Databases.LogTruncations)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesPercentLogUsed = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_log_used_percent"),
		"(Databases.PercentLogUsed)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesReplPendingXacts = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_pending_repl_transactions"),
		"(Databases.ReplPendingTransactions)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesReplTransRate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_repl_transactions"),
		"(Databases.ReplTranactions)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesShrinkDataMovementBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_shrink_data_movement_bytes"),
		"(Databases.ShrinkDataMovementBytes)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesTrackedtransactions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_tracked_transactions"),
		"(Databases.Trackedtransactions)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesTransactions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_transactions"),
		"(Databases.Transactions)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesWriteTransactions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_write_transactions"),
		"(Databases.WriteTransactions)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesXTPControllerDLCLatencyPerFetch = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_xtp_controller_dlc_fetch_latency_seconds"),
		"(Databases.XTPControllerDLCLatencyPerFetch)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesXTPControllerDLCPeakLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_xtp_controller_dlc_peak_latency_seconds"),
		"(Databases.XTPControllerDLCPeakLatency)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesXTPControllerLogProcessed = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_xtp_controller_log_processed_bytes"),
		"(Databases.XTPControllerLogProcessed)",
		[]string{"mssql_instance", "database"},
		nil,
	)
	c.DatabasesXTPMemoryUsedKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "databases_xtp_memory_used_bytes"),
		"(Databases.XTPMemoryUsedKB)",
		[]string{"mssql_instance", "database"},
		nil,
	)

	// Win32_PerfRawData_{instance}_SQLServerGeneralStatistics
	c.GenStatsActiveTempTables = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_active_temp_tables"),
		"(GeneralStatistics.ActiveTempTables)",
		[]string{"mssql_instance"},
		nil,
	)
	c.GenStatsConnectionReset = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_connection_resets"),
		"(GeneralStatistics.ConnectionReset)",
		[]string{"mssql_instance"},
		nil,
	)
	c.GenStatsEventNotificationsDelayedDrop = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_event_notifications_delayed_drop"),
		"(GeneralStatistics.EventNotificationsDelayedDrop)",
		[]string{"mssql_instance"},
		nil,
	)
	c.GenStatsHTTPAuthenticatedRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_http_authenticated_requests"),
		"(GeneralStatistics.HTTPAuthenticatedRequests)",
		[]string{"mssql_instance"},
		nil,
	)
	c.GenStatsLogicalConnections = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_logical_connections"),
		"(GeneralStatistics.LogicalConnections)",
		[]string{"mssql_instance"},
		nil,
	)
	c.GenStatsLogins = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_logins"),
		"(GeneralStatistics.Logins)",
		[]string{"mssql_instance"},
		nil,
	)
	c.GenStatsLogouts = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_logouts"),
		"(GeneralStatistics.Logouts)",
		[]string{"mssql_instance"},
		nil,
	)
	c.GenStatsMarsDeadlocks = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_mars_deadlocks"),
		"(GeneralStatistics.MarsDeadlocks)",
		[]string{"mssql_instance"},
		nil,
	)
	c.GenStatsNonatomicyieldrate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_non_atomic_yields"),
		"(GeneralStatistics.Nonatomicyields)",
		[]string{"mssql_instance"},
		nil,
	)
	c.GenStatsProcessesblocked = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_blocked_processes"),
		"(GeneralStatistics.Processesblocked)",
		[]string{"mssql_instance"},
		nil,
	)
	c.GenStatsSOAPEmptyRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_soap_empty_requests"),
		"(GeneralStatistics.SOAPEmptyRequests)",
		[]string{"mssql_instance"},
		nil,
	)
	c.GenStatsSOAPMethodInvocations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_soap_method_invocations"),
		"(GeneralStatistics.SOAPMethodInvocations)",
		[]string{"mssql_instance"},
		nil,
	)
	c.GenStatsSOAPSessionInitiateRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_soap_session_initiate_requests"),
		"(GeneralStatistics.SOAPSessionInitiateRequests)",
		[]string{"mssql_instance"},
		nil,
	)
	c.GenStatsSOAPSessionTerminateRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_soap_session_terminate_requests"),
		"(GeneralStatistics.SOAPSessionTerminateRequests)",
		[]string{"mssql_instance"},
		nil,
	)
	c.GenStatsSOAPSQLRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_soapsql_requests"),
		"(GeneralStatistics.SOAPSQLRequests)",
		[]string{"mssql_instance"},
		nil,
	)
	c.GenStatsSOAPWSDLRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_soapwsdl_requests"),
		"(GeneralStatistics.SOAPWSDLRequests)",
		[]string{"mssql_instance"},
		nil,
	)
	c.GenStatsSQLTraceIOProviderLockWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_sql_trace_io_provider_lock_waits"),
		"(GeneralStatistics.SQLTraceIOProviderLockWaits)",
		[]string{"mssql_instance"},
		nil,
	)
	c.GenStatsTempdbrecoveryunitid = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_tempdb_recovery_unit_ids_generated"),
		"(GeneralStatistics.Tempdbrecoveryunitid)",
		[]string{"mssql_instance"},
		nil,
	)
	c.GenStatsTempdbrowsetid = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_tempdb_rowset_ids_generated"),
		"(GeneralStatistics.Tempdbrowsetid)",
		[]string{"mssql_instance"},
		nil,
	)
	c.GenStatsTempTablesCreationRate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_temp_tables_creations"),
		"(GeneralStatistics.TempTablesCreations)",
		[]string{"mssql_instance"},
		nil,
	)
	c.GenStatsTempTablesForDestruction = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_temp_tables_awaiting_destruction"),
		"(GeneralStatistics.TempTablesForDestruction)",
		[]string{"mssql_instance"},
		nil,
	)
	c.GenStatsTraceEventNotificationQueue = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_trace_event_notification_queue_size"),
		"(GeneralStatistics.TraceEventNotificationQueue)",
		[]string{"mssql_instance"},
		nil,
	)
	c.GenStatsTransactions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_transactions"),
		"(GeneralStatistics.Transactions)",
		[]string{"mssql_instance"},
		nil,
	)
	c.GenStatsUserConnections = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_user_connections"),
		"(GeneralStatistics.UserConnections)",
		[]string{"mssql_instance"},
		nil,
	)

	// Win32_PerfRawData_{instance}_SQLServerLocks
	c.LocksWaitTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "locks_wait_time_seconds"),
		"(Locks.AverageWaitTimems Total time in seconds which locks have been holding resources)",
		[]string{"mssql_instance", "resource"},
		nil,
	)
	c.LocksCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "locks_count"),
		"(Locks.AverageWaitTimems_Base count of how often requests have run into locks)",
		[]string{"mssql_instance", "resource"},
		nil,
	)
	c.LocksLockRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "locks_lock_requests"),
		"(Locks.LockRequests)",
		[]string{"mssql_instance", "resource"},
		nil,
	)
	c.LocksLockTimeouts = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "locks_lock_timeouts"),
		"(Locks.LockTimeouts)",
		[]string{"mssql_instance", "resource"},
		nil,
	)
	c.LocksLockTimeoutstimeout0 = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "locks_lock_timeouts_excluding_NOWAIT"),
		"(Locks.LockTimeoutstimeout0)",
		[]string{"mssql_instance", "resource"},
		nil,
	)
	c.LocksLockWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "locks_lock_waits"),
		"(Locks.LockWaits)",
		[]string{"mssql_instance", "resource"},
		nil,
	)
	c.LocksLockWaitTimems = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "locks_lock_wait_seconds"),
		"(Locks.LockWaitTimems)",
		[]string{"mssql_instance", "resource"},
		nil,
	)
	c.LocksNumberofDeadlocks = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "locks_deadlocks"),
		"(Locks.NumberofDeadlocks)",
		[]string{"mssql_instance", "resource"},
		nil,
	)

	// Win32_PerfRawData_{instance}_SQLServerMemoryManager
	c.MemMgrConnectionMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_connection_memory_bytes"),
		"(MemoryManager.ConnectionMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.MemMgrDatabaseCacheMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_database_cache_memory_bytes"),
		"(MemoryManager.DatabaseCacheMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.MemMgrExternalbenefitofmemory = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_external_benefit_of_memory"),
		"(MemoryManager.Externalbenefitofmemory)",
		[]string{"mssql_instance"},
		nil,
	)
	c.MemMgrFreeMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_free_memory_bytes"),
		"(MemoryManager.FreeMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.MemMgrGrantedWorkspaceMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_granted_workspace_memory_bytes"),
		"(MemoryManager.GrantedWorkspaceMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.MemMgrLockBlocks = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_lock_blocks"),
		"(MemoryManager.LockBlocks)",
		[]string{"mssql_instance"},
		nil,
	)
	c.MemMgrLockBlocksAllocated = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_allocated_lock_blocks"),
		"(MemoryManager.LockBlocksAllocated)",
		[]string{"mssql_instance"},
		nil,
	)
	c.MemMgrLockMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_lock_memory_bytes"),
		"(MemoryManager.LockMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.MemMgrLockOwnerBlocks = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_lock_owner_blocks"),
		"(MemoryManager.LockOwnerBlocks)",
		[]string{"mssql_instance"},
		nil,
	)
	c.MemMgrLockOwnerBlocksAllocated = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_allocated_lock_owner_blocks"),
		"(MemoryManager.LockOwnerBlocksAllocated)",
		[]string{"mssql_instance"},
		nil,
	)
	c.MemMgrLogPoolMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_log_pool_memory_bytes"),
		"(MemoryManager.LogPoolMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.MemMgrMaximumWorkspaceMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_maximum_workspace_memory_bytes"),
		"(MemoryManager.MaximumWorkspaceMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.MemMgrMemoryGrantsOutstanding = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_outstanding_memory_grants"),
		"(MemoryManager.MemoryGrantsOutstanding)",
		[]string{"mssql_instance"},
		nil,
	)
	c.MemMgrMemoryGrantsPending = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_pending_memory_grants"),
		"(MemoryManager.MemoryGrantsPending)",
		[]string{"mssql_instance"},
		nil,
	)
	c.MemMgrOptimizerMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_optimizer_memory_bytes"),
		"(MemoryManager.OptimizerMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.MemMgrReservedServerMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_reserved_server_memory_bytes"),
		"(MemoryManager.ReservedServerMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.MemMgrSQLCacheMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_sql_cache_memory_bytes"),
		"(MemoryManager.SQLCacheMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.MemMgrStolenServerMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_stolen_server_memory_bytes"),
		"(MemoryManager.StolenServerMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.MemMgrTargetServerMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_target_server_memory_bytes"),
		"(MemoryManager.TargetServerMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.MemMgrTotalServerMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_total_server_memory_bytes"),
		"(MemoryManager.TotalServerMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)

	// Win32_PerfRawData_{instance}_SQLServerSQLStatistics
	c.SQLStatsAutoParamAttempts = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_auto_parameterization_attempts"),
		"(SQLStatistics.AutoParamAttempts)",
		[]string{"mssql_instance"},
		nil,
	)
	c.SQLStatsBatchRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_batch_requests"),
		"(SQLStatistics.BatchRequests)",
		[]string{"mssql_instance"},
		nil,
	)
	c.SQLStatsFailedAutoParams = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_failed_auto_parameterization_attempts"),
		"(SQLStatistics.FailedAutoParams)",
		[]string{"mssql_instance"},
		nil,
	)
	c.SQLStatsForcedParameterizations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_forced_parameterizations"),
		"(SQLStatistics.ForcedParameterizations)",
		[]string{"mssql_instance"},
		nil,
	)
	c.SQLStatsGuidedplanexecutions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_guided_plan_executions"),
		"(SQLStatistics.Guidedplanexecutions)",
		[]string{"mssql_instance"},
		nil,
	)
	c.SQLStatsMisguidedplanexecutions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_misguided_plan_executions"),
		"(SQLStatistics.Misguidedplanexecutions)",
		[]string{"mssql_instance"},
		nil,
	)
	c.SQLStatsSafeAutoParams = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_safe_auto_parameterization_attempts"),
		"(SQLStatistics.SafeAutoParams)",
		[]string{"mssql_instance"},
		nil,
	)
	c.SQLStatsSQLAttentionrate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_sql_attentions"),
		"(SQLStatistics.SQLAttentions)",
		[]string{"mssql_instance"},
		nil,
	)
	c.SQLStatsSQLCompilations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_sql_compilations"),
		"(SQLStatistics.SQLCompilations)",
		[]string{"mssql_instance"},
		nil,
	)
	c.SQLStatsSQLReCompilations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_sql_recompilations"),
		"(SQLStatistics.SQLReCompilations)",
		[]string{"mssql_instance"},
		nil,
	)
	c.SQLStatsUnsafeAutoParams = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_unsafe_auto_parameterization_attempts"),
		"(SQLStatistics.UnsafeAutoParams)",
		[]string{"mssql_instance"},
		nil,
	)

	// Win32_PerfRawData_{instance}_SQLServerSQLErrors
	c.SQLErrorsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sql_errors_total"),
		"(SQLErrors.Total)",
		[]string{"mssql_instance", "resource"},
		nil,
	)

	// Win32_PerfRawData_{instance}_SQLServerTransactions
	c.TransactionsTempDbFreeSpaceBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_tempdb_free_space_bytes"),
		"(Transactions.FreeSpaceInTempDbKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.TransactionsLongestTransactionRunningSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_longest_transaction_running_seconds"),
		"(Transactions.LongestTransactionRunningTime)",
		[]string{"mssql_instance"},
		nil,
	)
	c.TransactionsNonSnapshotVersionActiveTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_nonsnapshot_version_active_total"),
		"(Transactions.NonSnapshotVersionTransactions)",
		[]string{"mssql_instance"},
		nil,
	)
	c.TransactionsSnapshotActiveTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_snapshot_active_total"),
		"(Transactions.SnapshotTransactions)",
		[]string{"mssql_instance"},
		nil,
	)
	c.TransactionsActive = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_active"),
		"(Transactions.Transactions)",
		[]string{"mssql_instance"},
		nil,
	)
	c.TransactionsUpdateConflictsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_update_conflicts_total"),
		"(Transactions.UpdateConflictRatio)",
		[]string{"mssql_instance"},
		nil,
	)
	c.TransactionsUpdateSnapshotActiveTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_update_snapshot_active_total"),
		"(Transactions.UpdateSnapshotTransactions)",
		[]string{"mssql_instance"},
		nil,
	)
	c.TransactionsVersionCleanupRateBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_version_cleanup_rate_bytes"),
		"(Transactions.VersionCleanupRateKBs)",
		[]string{"mssql_instance"},
		nil,
	)
	c.TransactionsVersionGenerationRateBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_version_generation_rate_bytes"),
		"(Transactions.VersionGenerationRateKBs)",
		[]string{"mssql_instance"},
		nil,
	)
	c.TransactionsVersionStoreSizeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_version_store_size_bytes"),
		"(Transactions.VersionStoreSizeKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.TransactionsVersionStoreUnits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_version_store_units"),
		"(Transactions.VersionStoreUnitCount)",
		[]string{"mssql_instance"},
		nil,
	)
	c.TransactionsVersionStoreCreationUnits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_version_store_creation_units"),
		"(Transactions.VersionStoreUnitCreation)",
		[]string{"mssql_instance"},
		nil,
	)
	c.TransactionsVersionStoreTruncationUnits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_version_store_truncation_units"),
		"(Transactions.VersionStoreUnitTruncation)",
		[]string{"mssql_instance"},
		nil,
	)

	// Win32_PerfRawData_{instance}_SQLServerWaitStatistics
	c.WaitStatsLockWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_lock_waits"),
		"(WaitStats.LockWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)

	c.WaitStatsMemoryGrantQueueWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_memory_grant_queue_waits"),
		"(WaitStats.MemoryGrantQueueWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)
	c.WaitStatsThreadSafeMemoryObjectsWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_thread_safe_memory_objects_waits"),
		"(WaitStats.ThreadSafeMemoryObjectsWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)
	c.WaitStatsLogWriteWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_log_write_waits"),
		"(WaitStats.LogWriteWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)
	c.WaitStatsLogBufferWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_log_buffer_waits"),
		"(WaitStats.LogBufferWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)
	c.WaitStatsNetworkIOWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_network_io_waits"),
		"(WaitStats.NetworkIOWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)
	c.WaitStatsPageIOLatchWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_page_io_latch_waits"),
		"(WaitStats.PageIOLatchWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)
	c.WaitStatsPageLatchWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_page_latch_waits"),
		"(WaitStats.PageLatchWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)
	c.WaitStatsNonpageLatchWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_nonpage_latch_waits"),
		"(WaitStats.NonpageLatchWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)
	c.WaitStatsWaitForTheWorkerWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_wait_for_the_worker_waits"),
		"(WaitStats.WaitForTheWorkerWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)
	c.WaitStatsWorkspaceSynchronizationWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_workspace_synchronization_waits"),
		"(WaitStats.WorkspaceSynchronizationWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)
	c.WaitStatsTransactionOwnershipWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_transaction_ownership_waits"),
		"(WaitStats.TransactionOwnershipWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)

	c.mssqlCollectors = c.getMSSQLCollectors()

	if *c.mssqlPrintCollectors {
		fmt.Printf("Available SQLServer Classes:\n") //nolint:forbidigo
		for name := range c.mssqlCollectors {
			fmt.Printf(" - %s\n", name) //nolint:forbidigo
		}

		os.Exit(0)
	}

	return nil
}

type mssqlCollectorFunc func(ctx *types.ScrapeContext, ch chan<- prometheus.Metric, sqlInstance string) error

func (c *Collector) execute(ctx *types.ScrapeContext, name string, fn mssqlCollectorFunc, ch chan<- prometheus.Metric, sqlInstance string, wg *sync.WaitGroup) {
	// Reset failure counter on each scrape
	c.mssqlChildCollectorFailure = 0
	defer wg.Done()

	begin := time.Now()
	err := fn(ctx, ch, sqlInstance)
	duration := time.Since(begin)
	var success float64

	if err != nil {
		_ = level.Error(c.logger).Log("msg", fmt.Sprintf("mssql class collector %s failed after %fs", name, duration.Seconds()), "err", err)
		success = 0
		c.mssqlChildCollectorFailure++
	} else {
		_ = level.Debug(c.logger).Log("msg", fmt.Sprintf("mssql class collector %s succeeded after %fs.", name, duration.Seconds()))
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
func (c *Collector) Collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	wg := sync.WaitGroup{}

	enabled := utils.ExpandEnabledChildCollectors(*c.mssqlEnabledCollectors)
	for sqlInstance := range c.mssqlInstances {
		for _, name := range enabled {
			function := c.mssqlCollectors[name]

			wg.Add(1)
			go c.execute(ctx, name, function, ch, sqlInstance, &wg)
		}
	}
	wg.Wait()

	// this should return an error if any? some? children errord.
	if c.mssqlChildCollectorFailure > 0 {
		return errors.New("at least one child collector failed")
	}
	return nil
}

// Win32_PerfRawData_MSSQLSERVER_SQLServerAccessMethods docs:
// - https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-access-methods-object
type mssqlAccessMethods struct {
	AUcleanupbatchesPersec        float64 `perflib:"AU cleanup batches/sec"`
	AUcleanupsPersec              float64 `perflib:"AU cleanups/sec"`
	ByreferenceLobCreateCount     float64 `perflib:"By-reference Lob Create Count"`
	ByreferenceLobUseCount        float64 `perflib:"By-reference Lob Use Count"`
	CountLobReadahead             float64 `perflib:"Count Lob Readahead"`
	CountPullInRow                float64 `perflib:"Count Pull In Row"`
	CountPushOffRow               float64 `perflib:"Count Push Off Row"`
	DeferreddroppedAUs            float64 `perflib:"Deferred dropped AUs"`
	DeferredDroppedrowsets        float64 `perflib:"Deferred Dropped rowsets"`
	DroppedrowsetcleanupsPersec   float64 `perflib:"Dropped rowset cleanups/sec"`
	DroppedrowsetsskippedPersec   float64 `perflib:"Dropped rowsets skipped/sec"`
	ExtentDeallocationsPersec     float64 `perflib:"Extent Deallocations/sec"`
	ExtentsAllocatedPersec        float64 `perflib:"Extents Allocated/sec"`
	FailedAUcleanupbatchesPersec  float64 `perflib:"Failed AU cleanup batches/sec"`
	Failedleafpagecookie          float64 `perflib:"Failed leaf page cookie"`
	Failedtreepagecookie          float64 `perflib:"Failed tree page cookie"`
	ForwardedRecordsPersec        float64 `perflib:"Forwarded Records/sec"`
	FreeSpacePageFetchesPersec    float64 `perflib:"FreeSpace Page Fetches/sec"`
	FreeSpaceScansPersec          float64 `perflib:"FreeSpace Scans/sec"`
	FullScansPersec               float64 `perflib:"Full Scans/sec"`
	IndexSearchesPersec           float64 `perflib:"Index Searches/sec"`
	InSysXactwaitsPersec          float64 `perflib:"InSysXact waits/sec"`
	LobHandleCreateCount          float64 `perflib:"LobHandle Create Count"`
	LobHandleDestroyCount         float64 `perflib:"LobHandle Destroy Count"`
	LobSSProviderCreateCount      float64 `perflib:"LobSS Provider Create Count"`
	LobSSProviderDestroyCount     float64 `perflib:"LobSS Provider Destroy Count"`
	LobSSProviderTruncationCount  float64 `perflib:"LobSS Provider Truncation Count"`
	MixedpageallocationsPersec    float64 `perflib:"Mixed page allocations/sec"`
	PagecompressionattemptsPersec float64 `perflib:"Page compression attempts/sec"`
	PageDeallocationsPersec       float64 `perflib:"Page Deallocations/sec"`
	PagesAllocatedPersec          float64 `perflib:"Pages Allocated/sec"`
	PagescompressedPersec         float64 `perflib:"Pages compressed/sec"`
	PageSplitsPersec              float64 `perflib:"Page Splits/sec"`
	ProbeScansPersec              float64 `perflib:"Probe Scans/sec"`
	RangeScansPersec              float64 `perflib:"Range Scans/sec"`
	ScanPointRevalidationsPersec  float64 `perflib:"Scan Point Revalidations/sec"`
	SkippedGhostedRecordsPersec   float64 `perflib:"Skipped Ghosted Records/sec"`
	TableLockEscalationsPersec    float64 `perflib:"Table Lock Escalations/sec"`
	Usedleafpagecookie            float64 `perflib:"Used leaf page cookie"`
	Usedtreepagecookie            float64 `perflib:"Used tree page cookie"`
	WorkfilesCreatedPersec        float64 `perflib:"Workfiles Created/sec"`
	WorktablesCreatedPersec       float64 `perflib:"Worktables Created/sec"`
	WorktablesFromCacheRatio      float64 `perflib:"Worktables From Cache Ratio"`
	WorktablesFromCacheRatio_Base float64 `perflib:"Worktables From Cache Base_Base"`
}

func (c *Collector) collectAccessMethods(ctx *types.ScrapeContext, ch chan<- prometheus.Metric, sqlInstance string) error {
	var dst []mssqlAccessMethods
	_ = level.Debug(c.logger).Log("msg", fmt.Sprintf("mssql_accessmethods collector iterating sql instance %s.", sqlInstance))

	if err := perflib.UnmarshalObject(ctx.PerfObjects[mssqlGetPerfObjectName(sqlInstance, "accessmethods")], &dst, c.logger); err != nil {
		return err
	}

	for _, v := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsAUcleanupbatches,
			prometheus.CounterValue,
			v.AUcleanupbatchesPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsAUcleanups,
			prometheus.CounterValue,
			v.AUcleanupsPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsByreferenceLobCreateCount,
			prometheus.CounterValue,
			v.ByreferenceLobCreateCount,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsByreferenceLobUseCount,
			prometheus.CounterValue,
			v.ByreferenceLobUseCount,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsCountLobReadahead,
			prometheus.CounterValue,
			v.CountLobReadahead,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsCountPullInRow,
			prometheus.CounterValue,
			v.CountPullInRow,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsCountPushOffRow,
			prometheus.CounterValue,
			v.CountPushOffRow,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsDeferreddroppedAUs,
			prometheus.GaugeValue,
			v.DeferreddroppedAUs,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsDeferredDroppedrowsets,
			prometheus.GaugeValue,
			v.DeferredDroppedrowsets,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsDroppedrowsetcleanups,
			prometheus.CounterValue,
			v.DroppedrowsetcleanupsPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsDroppedrowsetsskipped,
			prometheus.CounterValue,
			v.DroppedrowsetsskippedPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsExtentDeallocations,
			prometheus.CounterValue,
			v.ExtentDeallocationsPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsExtentsAllocated,
			prometheus.CounterValue,
			v.ExtentsAllocatedPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsFailedAUcleanupbatches,
			prometheus.CounterValue,
			v.FailedAUcleanupbatchesPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsFailedleafpagecookie,
			prometheus.CounterValue,
			v.Failedleafpagecookie,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsFailedtreepagecookie,
			prometheus.CounterValue,
			v.Failedtreepagecookie,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsForwardedRecords,
			prometheus.CounterValue,
			v.ForwardedRecordsPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsFreeSpacePageFetches,
			prometheus.CounterValue,
			v.FreeSpacePageFetchesPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsFreeSpaceScans,
			prometheus.CounterValue,
			v.FreeSpaceScansPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsFullScans,
			prometheus.CounterValue,
			v.FullScansPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsIndexSearches,
			prometheus.CounterValue,
			v.IndexSearchesPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsInSysXactwaits,
			prometheus.CounterValue,
			v.InSysXactwaitsPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsLobHandleCreateCount,
			prometheus.CounterValue,
			v.LobHandleCreateCount,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsLobHandleDestroyCount,
			prometheus.CounterValue,
			v.LobHandleDestroyCount,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsLobSSProviderCreateCount,
			prometheus.CounterValue,
			v.LobSSProviderCreateCount,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsLobSSProviderDestroyCount,
			prometheus.CounterValue,
			v.LobSSProviderDestroyCount,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsLobSSProviderTruncationCount,
			prometheus.CounterValue,
			v.LobSSProviderTruncationCount,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsMixedpageallocations,
			prometheus.CounterValue,
			v.MixedpageallocationsPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsPagecompressionattempts,
			prometheus.CounterValue,
			v.PagecompressionattemptsPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsPageDeallocations,
			prometheus.CounterValue,
			v.PageDeallocationsPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsPagesAllocated,
			prometheus.CounterValue,
			v.PagesAllocatedPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsPagescompressed,
			prometheus.CounterValue,
			v.PagescompressedPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsPageSplits,
			prometheus.CounterValue,
			v.PageSplitsPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsProbeScans,
			prometheus.CounterValue,
			v.ProbeScansPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsRangeScans,
			prometheus.CounterValue,
			v.RangeScansPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsScanPointRevalidations,
			prometheus.CounterValue,
			v.ScanPointRevalidationsPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsSkippedGhostedRecords,
			prometheus.CounterValue,
			v.SkippedGhostedRecordsPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsTableLockEscalations,
			prometheus.CounterValue,
			v.TableLockEscalationsPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsUsedleafpagecookie,
			prometheus.CounterValue,
			v.Usedleafpagecookie,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsUsedtreepagecookie,
			prometheus.CounterValue,
			v.Usedtreepagecookie,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsWorkfilesCreated,
			prometheus.CounterValue,
			v.WorkfilesCreatedPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsWorktablesCreated,
			prometheus.CounterValue,
			v.WorktablesCreatedPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsWorktablesFromCacheHits,
			prometheus.CounterValue,
			v.WorktablesFromCacheRatio,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AccessMethodsWorktablesFromCacheLookups,
			prometheus.CounterValue,
			v.WorktablesFromCacheRatio_Base,
			sqlInstance,
		)
	}
	return nil
}

// Win32_PerfRawData_MSSQLSERVER_SQLServerAvailabilityReplica docs:
// - https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-availability-replica
type mssqlAvailabilityReplica struct {
	Name                           string
	BytesReceivedfromReplicaPersec float64 `perflib:"Bytes Received from Replica/sec"`
	BytesSenttoReplicaPersec       float64 `perflib:"Bytes Sent to Replica/sec"`
	BytesSenttoTransportPersec     float64 `perflib:"Bytes Sent to Transport/sec"`
	FlowControlPersec              float64 `perflib:"Flow Control/sec"`
	FlowControlTimemsPersec        float64 `perflib:"Flow Control Time (ms/sec)"`
	ReceivesfromReplicaPersec      float64 `perflib:"Receives from Replica/sec"`
	ResentMessagesPersec           float64 `perflib:"Resent Messages/sec"`
	SendstoReplicaPersec           float64 `perflib:"Sends to Replica/sec"`
	SendstoTransportPersec         float64 `perflib:"Sends to Transport/sec"`
}

func (c *Collector) collectAvailabilityReplica(ctx *types.ScrapeContext, ch chan<- prometheus.Metric, sqlInstance string) error {
	var dst []mssqlAvailabilityReplica
	_ = level.Debug(c.logger).Log("msg", fmt.Sprintf("mssql_availreplica collector iterating sql instance %s.", sqlInstance))

	if err := perflib.UnmarshalObject(ctx.PerfObjects[mssqlGetPerfObjectName(sqlInstance, "availreplica")], &dst, c.logger); err != nil {
		return err
	}

	for _, v := range dst {
		if strings.ToLower(v.Name) == "_total" {
			continue
		}
		replicaName := v.Name

		ch <- prometheus.MustNewConstMetric(
			c.AvailReplicaBytesReceivedfromReplica,
			prometheus.CounterValue,
			v.BytesReceivedfromReplicaPersec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AvailReplicaBytesSenttoReplica,
			prometheus.CounterValue,
			v.BytesSenttoReplicaPersec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AvailReplicaBytesSenttoTransport,
			prometheus.CounterValue,
			v.BytesSenttoTransportPersec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AvailReplicaFlowControl,
			prometheus.CounterValue,
			v.FlowControlPersec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AvailReplicaFlowControlTimems,
			prometheus.CounterValue,
			v.FlowControlTimemsPersec/1000.0,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AvailReplicaReceivesfromReplica,
			prometheus.CounterValue,
			v.ReceivesfromReplicaPersec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AvailReplicaResentMessages,
			prometheus.CounterValue,
			v.ResentMessagesPersec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AvailReplicaSendstoReplica,
			prometheus.CounterValue,
			v.SendstoReplicaPersec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AvailReplicaSendstoTransport,
			prometheus.CounterValue,
			v.SendstoTransportPersec,
			sqlInstance, replicaName,
		)
	}
	return nil
}

// Win32_PerfRawData_MSSQLSERVER_SQLServerBufferManager docs:
// - https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-buffer-manager-object
type mssqlBufferManager struct {
	BackgroundwriterpagesPersec   float64 `perflib:"Background writer pages/sec"`
	Buffercachehitratio           float64 `perflib:"Buffer cache hit ratio"`
	Buffercachehitratio_Base      float64 `perflib:"Buffer cache hit ratio base_Base"`
	CheckpointpagesPersec         float64 `perflib:"Checkpoint pages/sec"`
	Databasepages                 float64 `perflib:"Database pages"`
	Extensionallocatedpages       float64 `perflib:"Extension allocated pages"`
	Extensionfreepages            float64 `perflib:"Extension free pages"`
	Extensioninuseaspercentage    float64 `perflib:"Extension in use as percentage"`
	ExtensionoutstandingIOcounter float64 `perflib:"Extension outstanding IO counter"`
	ExtensionpageevictionsPersec  float64 `perflib:"Extension page evictions/sec"`
	ExtensionpagereadsPersec      float64 `perflib:"Extension page reads/sec"`
	Extensionpageunreferencedtime float64 `perflib:"Extension page unreferenced time"`
	ExtensionpagewritesPersec     float64 `perflib:"Extension page writes/sec"`
	FreeliststallsPersec          float64 `perflib:"Free list stalls/sec"`
	IntegralControllerSlope       float64 `perflib:"Integral Controller Slope"`
	LazywritesPersec              float64 `perflib:"Lazy writes/sec"`
	Pagelifeexpectancy            float64 `perflib:"Page life expectancy"`
	PagelookupsPersec             float64 `perflib:"Page lookups/sec"`
	PagereadsPersec               float64 `perflib:"Page reads/sec"`
	PagewritesPersec              float64 `perflib:"Page writes/sec"`
	ReadaheadpagesPersec          float64 `perflib:"Readahead pages/sec"`
	ReadaheadtimePersec           float64 `perflib:"Readahead time/sec"`
	Targetpages                   float64 `perflib:"Target pages"`
}

func (c *Collector) collectBufferManager(ctx *types.ScrapeContext, ch chan<- prometheus.Metric, sqlInstance string) error {
	var dst []mssqlBufferManager
	_ = level.Debug(c.logger).Log("msg", fmt.Sprintf("mssql_bufman collector iterating sql instance %s.", sqlInstance))

	if err := perflib.UnmarshalObject(ctx.PerfObjects[mssqlGetPerfObjectName(sqlInstance, "bufman")], &dst, c.logger); err != nil {
		return err
	}

	for _, v := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.BufManBackgroundwriterpages,
			prometheus.CounterValue,
			v.BackgroundwriterpagesPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BufManBuffercachehits,
			prometheus.GaugeValue,
			v.Buffercachehitratio,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BufManBuffercachelookups,
			prometheus.GaugeValue,
			v.Buffercachehitratio_Base,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BufManCheckpointpages,
			prometheus.CounterValue,
			v.CheckpointpagesPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BufManDatabasepages,
			prometheus.GaugeValue,
			v.Databasepages,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BufManExtensionallocatedpages,
			prometheus.GaugeValue,
			v.Extensionallocatedpages,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BufManExtensionfreepages,
			prometheus.GaugeValue,
			v.Extensionfreepages,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BufManExtensioninuseaspercentage,
			prometheus.GaugeValue,
			v.Extensioninuseaspercentage,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BufManExtensionoutstandingIOcounter,
			prometheus.GaugeValue,
			v.ExtensionoutstandingIOcounter,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BufManExtensionpageevictions,
			prometheus.CounterValue,
			v.ExtensionpageevictionsPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BufManExtensionpagereads,
			prometheus.CounterValue,
			v.ExtensionpagereadsPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BufManExtensionpageunreferencedtime,
			prometheus.GaugeValue,
			v.Extensionpageunreferencedtime,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BufManExtensionpagewrites,
			prometheus.CounterValue,
			v.ExtensionpagewritesPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BufManFreeliststalls,
			prometheus.CounterValue,
			v.FreeliststallsPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BufManIntegralControllerSlope,
			prometheus.GaugeValue,
			v.IntegralControllerSlope,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BufManLazywrites,
			prometheus.CounterValue,
			v.LazywritesPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BufManPagelifeexpectancy,
			prometheus.GaugeValue,
			v.Pagelifeexpectancy,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BufManPagelookups,
			prometheus.CounterValue,
			v.PagelookupsPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BufManPagereads,
			prometheus.CounterValue,
			v.PagereadsPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BufManPagewrites,
			prometheus.CounterValue,
			v.PagewritesPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BufManReadaheadpages,
			prometheus.CounterValue,
			v.ReadaheadpagesPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BufManReadaheadtime,
			prometheus.CounterValue,
			v.ReadaheadtimePersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BufManTargetpages,
			prometheus.GaugeValue,
			v.Targetpages,
			sqlInstance,
		)
	}

	return nil
}

// Win32_PerfRawData_MSSQLSERVER_SQLServerDatabaseReplica docs:
// - https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-database-replica
type mssqlDatabaseReplica struct {
	Name                            string
	DatabaseFlowControlDelay        float64 `perflib:"Database Flow Control Delay"`
	DatabaseFlowControlsPersec      float64 `perflib:"Database Flow Controls/sec"`
	FileBytesReceivedPersec         float64 `perflib:"File Bytes Received/sec"`
	GroupCommitsPerSec              float64 `perflib:"Group Commits/Sec"`
	GroupCommitTime                 float64 `perflib:"Group Commit Time"`
	LogApplyPendingQueue            float64 `perflib:"Log Apply Pending Queue"`
	LogApplyReadyQueue              float64 `perflib:"Log Apply Ready Queue"`
	LogBytesCompressedPersec        float64 `perflib:"Log Bytes Compressed/sec"`
	LogBytesDecompressedPersec      float64 `perflib:"Log Bytes Decompressed/sec"`
	LogBytesReceivedPersec          float64 `perflib:"Log Bytes Received/sec"`
	LogCompressionCachehitsPersec   float64 `perflib:"Log Compression Cache hits/sec"`
	LogCompressionCachemissesPersec float64 `perflib:"Log Compression Cache misses/sec"`
	LogCompressionsPersec           float64 `perflib:"Log Compressions/sec"`
	LogDecompressionsPersec         float64 `perflib:"Log Decompressions/sec"`
	Logremainingforundo             float64 `perflib:"Log remaining for undo"`
	LogSendQueue                    float64 `perflib:"Log Send Queue"`
	MirroredWriteTransactionsPersec float64 `perflib:"Mirrored Write Transactions/sec"`
	RecoveryQueue                   float64 `perflib:"Recovery Queue"`
	RedoblockedPersec               float64 `perflib:"Redo blocked/sec"`
	RedoBytesRemaining              float64 `perflib:"Redo Bytes Remaining"`
	RedoneBytesPersec               float64 `perflib:"Redone Bytes/sec"`
	RedonesPersec                   float64 `perflib:"Redones/sec"`
	TotalLogrequiringundo           float64 `perflib:"Total Log requiring undo"`
	TransactionDelay                float64 `perflib:"Transaction Delay"`
}

func (c *Collector) collectDatabaseReplica(ctx *types.ScrapeContext, ch chan<- prometheus.Metric, sqlInstance string) error {
	var dst []mssqlDatabaseReplica
	_ = level.Debug(c.logger).Log("msg", fmt.Sprintf("mssql_dbreplica collector iterating sql instance %s.", sqlInstance))

	if err := perflib.UnmarshalObject(ctx.PerfObjects[mssqlGetPerfObjectName(sqlInstance, "dbreplica")], &dst, c.logger); err != nil {
		return err
	}

	for _, v := range dst {
		if strings.ToLower(v.Name) == "_total" {
			continue
		}
		replicaName := v.Name

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaDatabaseFlowControlDelay,
			prometheus.GaugeValue,
			v.DatabaseFlowControlDelay,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaDatabaseFlowControls,
			prometheus.CounterValue,
			v.DatabaseFlowControlsPersec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaFileBytesReceived,
			prometheus.CounterValue,
			v.FileBytesReceivedPersec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaGroupCommits,
			prometheus.CounterValue,
			v.GroupCommitsPerSec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaGroupCommitTime,
			prometheus.GaugeValue,
			v.GroupCommitTime,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaLogApplyPendingQueue,
			prometheus.GaugeValue,
			v.LogApplyPendingQueue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaLogApplyReadyQueue,
			prometheus.GaugeValue,
			v.LogApplyReadyQueue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaLogBytesCompressed,
			prometheus.CounterValue,
			v.LogBytesCompressedPersec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaLogBytesDecompressed,
			prometheus.CounterValue,
			v.LogBytesDecompressedPersec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaLogBytesReceived,
			prometheus.CounterValue,
			v.LogBytesReceivedPersec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaLogCompressionCachehits,
			prometheus.CounterValue,
			v.LogCompressionCachehitsPersec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaLogCompressionCachemisses,
			prometheus.CounterValue,
			v.LogCompressionCachemissesPersec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaLogCompressions,
			prometheus.CounterValue,
			v.LogCompressionsPersec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaLogDecompressions,
			prometheus.CounterValue,
			v.LogDecompressionsPersec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaLogremainingforundo,
			prometheus.GaugeValue,
			v.Logremainingforundo,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaLogSendQueue,
			prometheus.GaugeValue,
			v.LogSendQueue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaMirroredWriteTransactions,
			prometheus.CounterValue,
			v.MirroredWriteTransactionsPersec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaRecoveryQueue,
			prometheus.GaugeValue,
			v.RecoveryQueue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaRedoblocked,
			prometheus.CounterValue,
			v.RedoblockedPersec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaRedoBytesRemaining,
			prometheus.GaugeValue,
			v.RedoBytesRemaining,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaRedoneBytes,
			prometheus.CounterValue,
			v.RedoneBytesPersec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaRedones,
			prometheus.CounterValue,
			v.RedonesPersec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaTotalLogrequiringundo,
			prometheus.GaugeValue,
			v.TotalLogrequiringundo,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DBReplicaTransactionDelay,
			prometheus.GaugeValue,
			v.TransactionDelay/1000.0,
			sqlInstance, replicaName,
		)
	}
	return nil
}

// Win32_PerfRawData_MSSQLSERVER_SQLServerDatabases docs:
// - https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-databases-object?view=sql-server-2017
type mssqlDatabases struct {
	Name                             string
	Activeparallelredothreads        float64 `perflib:"Active parallel redo threads"`
	ActiveTransactions               float64 `perflib:"Active Transactions"`
	BackupPerRestoreThroughputPersec float64 `perflib:"Backup/Restore Throughput/sec"`
	BulkCopyRowsPersec               float64 `perflib:"Bulk Copy Rows/sec"`
	BulkCopyThroughputPersec         float64 `perflib:"Bulk Copy Throughput/sec"`
	Committableentries               float64 `perflib:"Commit table entries"`
	DataFilesSizeKB                  float64 `perflib:"Data File(s) Size (KB)"`
	DBCCLogicalScanBytesPersec       float64 `perflib:"DBCC Logical Scan Bytes/sec"`
	GroupCommitTimePersec            float64 `perflib:"Group Commit Time/sec"`
	LogBytesFlushedPersec            float64 `perflib:"Log Bytes Flushed/sec"`
	LogCacheHitRatio                 float64 `perflib:"Log Cache Hit Ratio"`
	LogCacheHitRatio_Base            float64 `perflib:"Log Cache Hit Ratio Base_Base"`
	LogCacheReadsPersec              float64 `perflib:"Log Cache Reads/sec"`
	LogFilesSizeKB                   float64 `perflib:"Log File(s) Size (KB)"`
	LogFilesUsedSizeKB               float64 `perflib:"Log File(s) Used Size (KB)"`
	LogFlushesPersec                 float64 `perflib:"Log Flushes/sec"`
	LogFlushWaitsPersec              float64 `perflib:"Log Flush Waits/sec"`
	LogFlushWaitTime                 float64 `perflib:"Log Flush Wait Time"`
	LogFlushWriteTimems              float64 `perflib:"Log Flush Write Time (ms)"`
	LogGrowths                       float64 `perflib:"Log Growths"`
	LogPoolCacheMissesPersec         float64 `perflib:"Log Pool Cache Misses/sec"`
	LogPoolDiskReadsPersec           float64 `perflib:"Log Pool Disk Reads/sec"`
	LogPoolHashDeletesPersec         float64 `perflib:"Log Pool Hash Deletes/sec"`
	LogPoolHashInsertsPersec         float64 `perflib:"Log Pool Hash Inserts/sec"`
	LogPoolInvalidHashEntryPersec    float64 `perflib:"Log Pool Invalid Hash Entry/sec"`
	LogPoolLogScanPushesPersec       float64 `perflib:"Log Pool Log Scan Pushes/sec"`
	LogPoolLogWriterPushesPersec     float64 `perflib:"Log Pool LogWriter Pushes/sec"`
	LogPoolPushEmptyFreePoolPersec   float64 `perflib:"Log Pool Push Empty FreePool/sec"`
	LogPoolPushLowMemoryPersec       float64 `perflib:"Log Pool Push Low Memory/sec"`
	LogPoolPushNoFreeBufferPersec    float64 `perflib:"Log Pool Push No Free Buffer/sec"`
	LogPoolReqBehindTruncPersec      float64 `perflib:"Log Pool Req. Behind Trunc/sec"`
	LogPoolRequestsOldVLFPersec      float64 `perflib:"Log Pool Requests Old VLF/sec"`
	LogPoolRequestsPersec            float64 `perflib:"Log Pool Requests/sec"`
	LogPoolTotalActiveLogSize        float64 `perflib:"Log Pool Total Active Log Size"`
	LogPoolTotalSharedPoolSize       float64 `perflib:"Log Pool Total Shared Pool Size"`
	LogShrinks                       float64 `perflib:"Log Shrinks"`
	LogTruncations                   float64 `perflib:"Log Truncations"`
	PercentLogUsed                   float64 `perflib:"Percent Log Used"`
	ReplPendingXacts                 float64 `perflib:"Repl. Pending Xacts"`
	ReplTransRate                    float64 `perflib:"Repl. Trans. Rate"`
	ShrinkDataMovementBytesPersec    float64 `perflib:"Shrink Data Movement Bytes/sec"`
	TrackedtransactionsPersec        float64 `perflib:"Tracked transactions/sec"`
	TransactionsPersec               float64 `perflib:"Transactions/sec"`
	WriteTransactionsPersec          float64 `perflib:"Write Transactions/sec"`
	XTPControllerDLCLatencyPerFetch  float64 `perflib:"XTP Controller DLC Latency/Fetch"`
	XTPControllerDLCPeakLatency      float64 `perflib:"XTP Controller DLC Peak Latency"`
	XTPControllerLogProcessedPersec  float64 `perflib:"XTP Controller Log Processed/sec"`
	XTPMemoryUsedKB                  float64 `perflib:"XTP Memory Used (KB)"`
}

func (c *Collector) collectDatabases(ctx *types.ScrapeContext, ch chan<- prometheus.Metric, sqlInstance string) error {
	var dst []mssqlDatabases
	_ = level.Debug(c.logger).Log("msg", fmt.Sprintf("mssql_databases collector iterating sql instance %s.", sqlInstance))

	if err := perflib.UnmarshalObject(ctx.PerfObjects[mssqlGetPerfObjectName(sqlInstance, "databases")], &dst, c.logger); err != nil {
		return err
	}

	for _, v := range dst {
		if strings.ToLower(v.Name) == "_total" {
			continue
		}
		dbName := v.Name

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesActiveParallelredothreads,
			prometheus.GaugeValue,
			v.Activeparallelredothreads,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesActiveTransactions,
			prometheus.GaugeValue,
			v.ActiveTransactions,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesBackupPerRestoreThroughput,
			prometheus.CounterValue,
			v.BackupPerRestoreThroughputPersec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesBulkCopyRows,
			prometheus.CounterValue,
			v.BulkCopyRowsPersec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesBulkCopyThroughput,
			prometheus.CounterValue,
			v.BulkCopyThroughputPersec*1024,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesCommittableentries,
			prometheus.GaugeValue,
			v.Committableentries,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesDataFilesSizeKB,
			prometheus.GaugeValue,
			v.DataFilesSizeKB*1024,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesDBCCLogicalScanBytes,
			prometheus.CounterValue,
			v.DBCCLogicalScanBytesPersec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesGroupCommitTime,
			prometheus.CounterValue,
			v.GroupCommitTimePersec/1000000.0,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogBytesFlushed,
			prometheus.CounterValue,
			v.LogBytesFlushedPersec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogCacheHits,
			prometheus.GaugeValue,
			v.LogCacheHitRatio,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogCacheLookups,
			prometheus.GaugeValue,
			v.LogCacheHitRatio_Base,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogCacheReads,
			prometheus.CounterValue,
			v.LogCacheReadsPersec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogFilesSizeKB,
			prometheus.GaugeValue,
			v.LogFilesSizeKB*1024,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogFilesUsedSizeKB,
			prometheus.GaugeValue,
			v.LogFilesUsedSizeKB*1024,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogFlushes,
			prometheus.CounterValue,
			v.LogFlushesPersec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogFlushWaits,
			prometheus.CounterValue,
			v.LogFlushWaitsPersec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogFlushWaitTime,
			prometheus.GaugeValue,
			v.LogFlushWaitTime/1000.0,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogFlushWriteTimems,
			prometheus.GaugeValue,
			v.LogFlushWriteTimems/1000.0,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogGrowths,
			prometheus.GaugeValue,
			v.LogGrowths,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogPoolCacheMisses,
			prometheus.CounterValue,
			v.LogPoolCacheMissesPersec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogPoolDiskReads,
			prometheus.CounterValue,
			v.LogPoolDiskReadsPersec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogPoolHashDeletes,
			prometheus.CounterValue,
			v.LogPoolHashDeletesPersec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogPoolHashInserts,
			prometheus.CounterValue,
			v.LogPoolHashInsertsPersec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogPoolInvalidHashEntry,
			prometheus.CounterValue,
			v.LogPoolInvalidHashEntryPersec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogPoolLogScanPushes,
			prometheus.CounterValue,
			v.LogPoolLogScanPushesPersec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogPoolLogWriterPushes,
			prometheus.CounterValue,
			v.LogPoolLogWriterPushesPersec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogPoolPushEmptyFreePool,
			prometheus.CounterValue,
			v.LogPoolPushEmptyFreePoolPersec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogPoolPushLowMemory,
			prometheus.CounterValue,
			v.LogPoolPushLowMemoryPersec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogPoolPushNoFreeBuffer,
			prometheus.CounterValue,
			v.LogPoolPushNoFreeBufferPersec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogPoolReqBehindTrunc,
			prometheus.CounterValue,
			v.LogPoolReqBehindTruncPersec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogPoolRequestsOldVLF,
			prometheus.CounterValue,
			v.LogPoolRequestsOldVLFPersec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogPoolRequests,
			prometheus.CounterValue,
			v.LogPoolRequestsPersec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogPoolTotalActiveLogSize,
			prometheus.GaugeValue,
			v.LogPoolTotalActiveLogSize,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogPoolTotalSharedPoolSize,
			prometheus.GaugeValue,
			v.LogPoolTotalSharedPoolSize,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogShrinks,
			prometheus.GaugeValue,
			v.LogShrinks,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesLogTruncations,
			prometheus.GaugeValue,
			v.LogTruncations,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesPercentLogUsed,
			prometheus.GaugeValue,
			v.PercentLogUsed,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesReplPendingXacts,
			prometheus.GaugeValue,
			v.ReplPendingXacts,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesReplTransRate,
			prometheus.CounterValue,
			v.ReplTransRate,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesShrinkDataMovementBytes,
			prometheus.CounterValue,
			v.ShrinkDataMovementBytesPersec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesTrackedtransactions,
			prometheus.CounterValue,
			v.TrackedtransactionsPersec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesTransactions,
			prometheus.CounterValue,
			v.TransactionsPersec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesWriteTransactions,
			prometheus.CounterValue,
			v.WriteTransactionsPersec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesXTPControllerDLCLatencyPerFetch,
			prometheus.GaugeValue,
			v.XTPControllerDLCLatencyPerFetch,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesXTPControllerDLCPeakLatency,
			prometheus.GaugeValue,
			v.XTPControllerDLCPeakLatency*1000000.0,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesXTPControllerLogProcessed,
			prometheus.CounterValue,
			v.XTPControllerLogProcessedPersec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabasesXTPMemoryUsedKB,
			prometheus.GaugeValue,
			v.XTPMemoryUsedKB*1024,
			sqlInstance, dbName,
		)
	}
	return nil
}

// Win32_PerfRawData_MSSQLSERVER_SQLServerGeneralStatistics docs:
// - https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-general-statistics-object
type mssqlGeneralStatistics struct {
	ActiveTempTables              float64 `perflib:"Active Temp Tables"`
	ConnectionResetPersec         float64 `perflib:"Connection Reset/sec"`
	EventNotificationsDelayedDrop float64 `perflib:"Event Notifications Delayed Drop"`
	HTTPAuthenticatedRequests     float64 `perflib:"HTTP Authenticated Requests"`
	LogicalConnections            float64 `perflib:"Logical Connections"`
	LoginsPersec                  float64 `perflib:"Logins/sec"`
	LogoutsPersec                 float64 `perflib:"Logouts/sec"`
	MarsDeadlocks                 float64 `perflib:"Mars Deadlocks"`
	Nonatomicyieldrate            float64 `perflib:"Non-atomic yield rate"`
	Processesblocked              float64 `perflib:"Processes blocked"`
	SOAPEmptyRequests             float64 `perflib:"SOAP Empty Requests"`
	SOAPMethodInvocations         float64 `perflib:"SOAP Method Invocations"`
	SOAPSessionInitiateRequests   float64 `perflib:"SOAP Session Initiate Requests"`
	SOAPSessionTerminateRequests  float64 `perflib:"SOAP Session Terminate Requests"`
	SOAPSQLRequests               float64 `perflib:"SOAP SQL Requests"`
	SOAPWSDLRequests              float64 `perflib:"SOAP WSDL Requests"`
	SQLTraceIOProviderLockWaits   float64 `perflib:"SQL Trace IO Provider Lock Waits"`
	Tempdbrecoveryunitid          float64 `perflib:"Tempdb recovery unit id"`
	Tempdbrowsetid                float64 `perflib:"Tempdb rowset id"`
	TempTablesCreationRate        float64 `perflib:"Temp Tables Creation Rate"`
	TempTablesForDestruction      float64 `perflib:"Temp Tables For Destruction"`
	TraceEventNotificationQueue   float64 `perflib:"Trace Event Notification Queue"`
	Transactions                  float64 `perflib:"Transactions"`
	UserConnections               float64 `perflib:"User Connections"`
}

func (c *Collector) collectGeneralStatistics(ctx *types.ScrapeContext, ch chan<- prometheus.Metric, sqlInstance string) error {
	var dst []mssqlGeneralStatistics
	_ = level.Debug(c.logger).Log("msg", fmt.Sprintf("mssql_genstats collector iterating sql instance %s.", sqlInstance))

	if err := perflib.UnmarshalObject(ctx.PerfObjects[mssqlGetPerfObjectName(sqlInstance, "genstats")], &dst, c.logger); err != nil {
		return err
	}

	for _, v := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.GenStatsActiveTempTables,
			prometheus.GaugeValue,
			v.ActiveTempTables,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GenStatsConnectionReset,
			prometheus.CounterValue,
			v.ConnectionResetPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GenStatsEventNotificationsDelayedDrop,
			prometheus.GaugeValue,
			v.EventNotificationsDelayedDrop,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GenStatsHTTPAuthenticatedRequests,
			prometheus.GaugeValue,
			v.HTTPAuthenticatedRequests,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GenStatsLogicalConnections,
			prometheus.GaugeValue,
			v.LogicalConnections,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GenStatsLogins,
			prometheus.CounterValue,
			v.LoginsPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GenStatsLogouts,
			prometheus.CounterValue,
			v.LogoutsPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GenStatsMarsDeadlocks,
			prometheus.GaugeValue,
			v.MarsDeadlocks,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GenStatsNonatomicyieldrate,
			prometheus.CounterValue,
			v.Nonatomicyieldrate,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GenStatsProcessesblocked,
			prometheus.GaugeValue,
			v.Processesblocked,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GenStatsSOAPEmptyRequests,
			prometheus.GaugeValue,
			v.SOAPEmptyRequests,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GenStatsSOAPMethodInvocations,
			prometheus.GaugeValue,
			v.SOAPMethodInvocations,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GenStatsSOAPSessionInitiateRequests,
			prometheus.GaugeValue,
			v.SOAPSessionInitiateRequests,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GenStatsSOAPSessionTerminateRequests,
			prometheus.GaugeValue,
			v.SOAPSessionTerminateRequests,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GenStatsSOAPSQLRequests,
			prometheus.GaugeValue,
			v.SOAPSQLRequests,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GenStatsSOAPWSDLRequests,
			prometheus.GaugeValue,
			v.SOAPWSDLRequests,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GenStatsSQLTraceIOProviderLockWaits,
			prometheus.GaugeValue,
			v.SQLTraceIOProviderLockWaits,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GenStatsTempdbrecoveryunitid,
			prometheus.GaugeValue,
			v.Tempdbrecoveryunitid,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GenStatsTempdbrowsetid,
			prometheus.GaugeValue,
			v.Tempdbrowsetid,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GenStatsTempTablesCreationRate,
			prometheus.CounterValue,
			v.TempTablesCreationRate,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GenStatsTempTablesForDestruction,
			prometheus.GaugeValue,
			v.TempTablesForDestruction,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GenStatsTraceEventNotificationQueue,
			prometheus.GaugeValue,
			v.TraceEventNotificationQueue,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GenStatsTransactions,
			prometheus.GaugeValue,
			v.Transactions,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GenStatsUserConnections,
			prometheus.GaugeValue,
			v.UserConnections,
			sqlInstance,
		)
	}

	return nil
}

// Win32_PerfRawData_MSSQLSERVER_SQLServerLocks docs:
// - https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-locks-object
type mssqlLocks struct {
	Name                       string
	AverageWaitTimems          float64 `perflib:"Average Wait Time (ms)"`
	AverageWaitTimems_Base     float64 `perflib:"Average Wait Time Base_Base"`
	LockRequestsPersec         float64 `perflib:"Lock Requests/sec"`
	LockTimeoutsPersec         float64 `perflib:"Lock Timeouts/sec"`
	LockTimeoutstimeout0Persec float64 `perflib:"Lock Timeouts (timeout > 0)/sec"`
	LockWaitsPersec            float64 `perflib:"Lock Waits/sec"`
	LockWaitTimems             float64 `perflib:"Lock Wait Time (ms)"`
	NumberofDeadlocksPersec    float64 `perflib:"Number of Deadlocks/sec"`
}

func (c *Collector) collectLocks(ctx *types.ScrapeContext, ch chan<- prometheus.Metric, sqlInstance string) error {
	var dst []mssqlLocks
	_ = level.Debug(c.logger).Log("msg", fmt.Sprintf("mssql_locks collector iterating sql instance %s.", sqlInstance))

	if err := perflib.UnmarshalObject(ctx.PerfObjects[mssqlGetPerfObjectName(sqlInstance, "locks")], &dst, c.logger); err != nil {
		return err
	}

	for _, v := range dst {
		if strings.ToLower(v.Name) == "_total" {
			continue
		}
		lockResourceName := v.Name

		ch <- prometheus.MustNewConstMetric(
			c.LocksWaitTime,
			prometheus.GaugeValue,
			v.AverageWaitTimems/1000.0,
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LocksCount,
			prometheus.GaugeValue,
			v.AverageWaitTimems_Base/1000.0,
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LocksLockRequests,
			prometheus.CounterValue,
			v.LockRequestsPersec,
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LocksLockTimeouts,
			prometheus.CounterValue,
			v.LockTimeoutsPersec,
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LocksLockTimeoutstimeout0,
			prometheus.CounterValue,
			v.LockTimeoutstimeout0Persec,
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LocksLockWaits,
			prometheus.CounterValue,
			v.LockWaitsPersec,
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LocksLockWaitTimems,
			prometheus.GaugeValue,
			v.LockWaitTimems/1000.0,
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LocksNumberofDeadlocks,
			prometheus.CounterValue,
			v.NumberofDeadlocksPersec,
			sqlInstance, lockResourceName,
		)
	}
	return nil
}

// Win32_PerfRawData_MSSQLSERVER_SQLServerMemoryManager docs:
// - https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-memory-manager-object
type mssqlMemoryManager struct {
	ConnectionMemoryKB       float64 `perflib:"Connection Memory (KB)"`
	DatabaseCacheMemoryKB    float64 `perflib:"Database Cache Memory (KB)"`
	Externalbenefitofmemory  float64 `perflib:"External benefit of memory"`
	FreeMemoryKB             float64 `perflib:"Free Memory (KB)"`
	GrantedWorkspaceMemoryKB float64 `perflib:"Granted Workspace Memory (KB)"`
	LockBlocks               float64 `perflib:"Lock Blocks"`
	LockBlocksAllocated      float64 `perflib:"Lock Blocks Allocated"`
	LockMemoryKB             float64 `perflib:"Lock Memory (KB)"`
	LockOwnerBlocks          float64 `perflib:"Lock Owner Blocks"`
	LockOwnerBlocksAllocated float64 `perflib:"Lock Owner Blocks Allocated"`
	LogPoolMemoryKB          float64 `perflib:"Log Pool Memory (KB)"`
	MaximumWorkspaceMemoryKB float64 `perflib:"Maximum Workspace Memory (KB)"`
	MemoryGrantsOutstanding  float64 `perflib:"Memory Grants Outstanding"`
	MemoryGrantsPending      float64 `perflib:"Memory Grants Pending"`
	OptimizerMemoryKB        float64 `perflib:"Optimizer Memory (KB)"`
	ReservedServerMemoryKB   float64 `perflib:"Reserved Server Memory (KB)"`
	SQLCacheMemoryKB         float64 `perflib:"SQL Cache Memory (KB)"`
	StolenServerMemoryKB     float64 `perflib:"Stolen Server Memory (KB)"`
	TargetServerMemoryKB     float64 `perflib:"Target Server Memory (KB)"`
	TotalServerMemoryKB      float64 `perflib:"Total Server Memory (KB)"`
}

func (c *Collector) collectMemoryManager(ctx *types.ScrapeContext, ch chan<- prometheus.Metric, sqlInstance string) error {
	var dst []mssqlMemoryManager
	_ = level.Debug(c.logger).Log("msg", fmt.Sprintf("mssql_memmgr collector iterating sql instance %s.", sqlInstance))

	if err := perflib.UnmarshalObject(ctx.PerfObjects[mssqlGetPerfObjectName(sqlInstance, "memmgr")], &dst, c.logger); err != nil {
		return err
	}

	for _, v := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.MemMgrConnectionMemoryKB,
			prometheus.GaugeValue,
			v.ConnectionMemoryKB*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MemMgrDatabaseCacheMemoryKB,
			prometheus.GaugeValue,
			v.DatabaseCacheMemoryKB*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MemMgrExternalbenefitofmemory,
			prometheus.GaugeValue,
			v.Externalbenefitofmemory,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MemMgrFreeMemoryKB,
			prometheus.GaugeValue,
			v.FreeMemoryKB*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MemMgrGrantedWorkspaceMemoryKB,
			prometheus.GaugeValue,
			v.GrantedWorkspaceMemoryKB*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MemMgrLockBlocks,
			prometheus.GaugeValue,
			v.LockBlocks,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MemMgrLockBlocksAllocated,
			prometheus.GaugeValue,
			v.LockBlocksAllocated,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MemMgrLockMemoryKB,
			prometheus.GaugeValue,
			v.LockMemoryKB*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MemMgrLockOwnerBlocks,
			prometheus.GaugeValue,
			v.LockOwnerBlocks,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MemMgrLockOwnerBlocksAllocated,
			prometheus.GaugeValue,
			v.LockOwnerBlocksAllocated,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MemMgrLogPoolMemoryKB,
			prometheus.GaugeValue,
			v.LogPoolMemoryKB*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MemMgrMaximumWorkspaceMemoryKB,
			prometheus.GaugeValue,
			v.MaximumWorkspaceMemoryKB*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MemMgrMemoryGrantsOutstanding,
			prometheus.GaugeValue,
			v.MemoryGrantsOutstanding,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MemMgrMemoryGrantsPending,
			prometheus.GaugeValue,
			v.MemoryGrantsPending,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MemMgrOptimizerMemoryKB,
			prometheus.GaugeValue,
			v.OptimizerMemoryKB*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MemMgrReservedServerMemoryKB,
			prometheus.GaugeValue,
			v.ReservedServerMemoryKB*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MemMgrSQLCacheMemoryKB,
			prometheus.GaugeValue,
			v.SQLCacheMemoryKB*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MemMgrStolenServerMemoryKB,
			prometheus.GaugeValue,
			v.StolenServerMemoryKB*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MemMgrTargetServerMemoryKB,
			prometheus.GaugeValue,
			v.TargetServerMemoryKB*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MemMgrTotalServerMemoryKB,
			prometheus.GaugeValue,
			v.TotalServerMemoryKB*1024,
			sqlInstance,
		)
	}

	return nil
}

// Win32_PerfRawData_MSSQLSERVER_SQLServerSQLStatistics docs:
// - https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-sql-statistics-object
type mssqlSQLStatistics struct {
	AutoParamAttemptsPersec       float64 `perflib:"Auto-Param Attempts/sec"`
	BatchRequestsPersec           float64 `perflib:"Batch Requests/sec"`
	FailedAutoParamsPersec        float64 `perflib:"Failed Auto-Params/sec"`
	ForcedParameterizationsPersec float64 `perflib:"Forced Parameterizations/sec"`
	GuidedplanexecutionsPersec    float64 `perflib:"Guided plan executions/sec"`
	MisguidedplanexecutionsPersec float64 `perflib:"Misguided plan executions/sec"`
	SafeAutoParamsPersec          float64 `perflib:"Safe Auto-Params/sec"`
	SQLAttentionrate              float64 `perflib:"SQL Attention rate"`
	SQLCompilationsPersec         float64 `perflib:"SQL Compilations/sec"`
	SQLReCompilationsPersec       float64 `perflib:"SQL Re-Compilations/sec"`
	UnsafeAutoParamsPersec        float64 `perflib:"Unsafe Auto-Params/sec"`
}

func (c *Collector) collectSQLStats(ctx *types.ScrapeContext, ch chan<- prometheus.Metric, sqlInstance string) error {
	var dst []mssqlSQLStatistics
	_ = level.Debug(c.logger).Log("msg", fmt.Sprintf("mssql_sqlstats collector iterating sql instance %s.", sqlInstance))

	if err := perflib.UnmarshalObject(ctx.PerfObjects[mssqlGetPerfObjectName(sqlInstance, "sqlstats")], &dst, c.logger); err != nil {
		return err
	}

	for _, v := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.SQLStatsAutoParamAttempts,
			prometheus.CounterValue,
			v.AutoParamAttemptsPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SQLStatsBatchRequests,
			prometheus.CounterValue,
			v.BatchRequestsPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SQLStatsFailedAutoParams,
			prometheus.CounterValue,
			v.FailedAutoParamsPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SQLStatsForcedParameterizations,
			prometheus.CounterValue,
			v.ForcedParameterizationsPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SQLStatsGuidedplanexecutions,
			prometheus.CounterValue,
			v.GuidedplanexecutionsPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SQLStatsMisguidedplanexecutions,
			prometheus.CounterValue,
			v.MisguidedplanexecutionsPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SQLStatsSafeAutoParams,
			prometheus.CounterValue,
			v.SafeAutoParamsPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SQLStatsSQLAttentionrate,
			prometheus.CounterValue,
			v.SQLAttentionrate,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SQLStatsSQLCompilations,
			prometheus.CounterValue,
			v.SQLCompilationsPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SQLStatsSQLReCompilations,
			prometheus.CounterValue,
			v.SQLReCompilationsPersec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SQLStatsUnsafeAutoParams,
			prometheus.CounterValue,
			v.UnsafeAutoParamsPersec,
			sqlInstance,
		)
	}

	return nil
}

// Win32_PerfRawData_MSSQLSERVER_SQLServerWaitStatistics docs:
// - https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-wait-statistics-object
type mssqlWaitStatistics struct {
	Name                                   string
	WaitStatsLockWaits                     float64 `perflib:"Lock waits"`
	WaitStatsMemoryGrantQueueWaits         float64 `perflib:"Memory grant queue waits"`
	WaitStatsThreadSafeMemoryObjectsWaits  float64 `perflib:"Thread-safe memory objects waits"`
	WaitStatsLogWriteWaits                 float64 `perflib:"Log write waits"`
	WaitStatsLogBufferWaits                float64 `perflib:"Log buffer waits"`
	WaitStatsNetworkIOWaits                float64 `perflib:"Network IO waits"`
	WaitStatsPageIOLatchWaits              float64 `perflib:"Page IO latch waits"`
	WaitStatsPageLatchWaits                float64 `perflib:"Page latch waits"`
	WaitStatsNonpageLatchWaits             float64 `perflib:"Non-Page latch waits"`
	WaitStatsWaitForTheWorkerWaits         float64 `perflib:"Wait for the worker"`
	WaitStatsWorkspaceSynchronizationWaits float64 `perflib:"Workspace synchronization waits"`
	WaitStatsTransactionOwnershipWaits     float64 `perflib:"Transaction ownership waits"`
}

func (c *Collector) collectWaitStats(ctx *types.ScrapeContext, ch chan<- prometheus.Metric, sqlInstance string) error {
	var dst []mssqlWaitStatistics
	_ = level.Debug(c.logger).Log("msg", fmt.Sprintf("mssql_waitstats collector iterating sql instance %s.", sqlInstance))

	if err := perflib.UnmarshalObject(ctx.PerfObjects[mssqlGetPerfObjectName(sqlInstance, "waitstats")], &dst, c.logger); err != nil {
		return err
	}

	for _, v := range dst {
		item := v.Name

		ch <- prometheus.MustNewConstMetric(
			c.WaitStatsLockWaits,
			prometheus.CounterValue,
			v.WaitStatsLockWaits,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WaitStatsMemoryGrantQueueWaits,
			prometheus.CounterValue,
			v.WaitStatsMemoryGrantQueueWaits,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WaitStatsThreadSafeMemoryObjectsWaits,
			prometheus.CounterValue,
			v.WaitStatsThreadSafeMemoryObjectsWaits,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WaitStatsLogWriteWaits,
			prometheus.CounterValue,
			v.WaitStatsLogWriteWaits,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WaitStatsLogBufferWaits,
			prometheus.CounterValue,
			v.WaitStatsLogBufferWaits,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WaitStatsNetworkIOWaits,
			prometheus.CounterValue,
			v.WaitStatsNetworkIOWaits,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WaitStatsPageIOLatchWaits,
			prometheus.CounterValue,
			v.WaitStatsPageIOLatchWaits,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WaitStatsPageLatchWaits,
			prometheus.CounterValue,
			v.WaitStatsPageLatchWaits,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WaitStatsNonpageLatchWaits,
			prometheus.CounterValue,
			v.WaitStatsNonpageLatchWaits,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WaitStatsWaitForTheWorkerWaits,
			prometheus.CounterValue,
			v.WaitStatsWaitForTheWorkerWaits,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WaitStatsWorkspaceSynchronizationWaits,
			prometheus.CounterValue,
			v.WaitStatsWorkspaceSynchronizationWaits,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WaitStatsTransactionOwnershipWaits,
			prometheus.CounterValue,
			v.WaitStatsTransactionOwnershipWaits,
			sqlInstance, item,
		)
	}

	return nil
}

type mssqlSQLErrors struct {
	Name         string
	ErrorsPersec float64 `perflib:"Errors/sec"`
}

// Win32_PerfRawData_MSSQLSERVER_SQLServerErrors docs:
// - https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-sql-errors-object
func (c *Collector) collectSQLErrors(ctx *types.ScrapeContext, ch chan<- prometheus.Metric, sqlInstance string) error {
	var dst []mssqlSQLErrors
	_ = level.Debug(c.logger).Log("msg", fmt.Sprintf("mssql_sqlerrors collector iterating sql instance %s.", sqlInstance))

	if err := perflib.UnmarshalObject(ctx.PerfObjects[mssqlGetPerfObjectName(sqlInstance, "sqlerrors")], &dst, c.logger); err != nil {
		return err
	}

	for _, v := range dst {
		if strings.ToLower(v.Name) == "_total" {
			continue
		}
		resource := v.Name

		ch <- prometheus.MustNewConstMetric(
			c.SQLErrorsTotal,
			prometheus.CounterValue,
			v.ErrorsPersec,
			sqlInstance, resource,
		)
	}

	return nil
}

type mssqlTransactions struct {
	FreeSpaceintempdbKB            float64 `perflib:"Free Space in tempdb (KB)"`
	LongestTransactionRunningTime  float64 `perflib:"Longest Transaction Running Time"`
	NonSnapshotVersionTransactions float64 `perflib:"NonSnapshot Version Transactions"`
	SnapshotTransactions           float64 `perflib:"Snapshot Transactions"`
	Transactions                   float64 `perflib:"Transactions"`
	Updateconflictratio            float64 `perflib:"Update conflict ratio"`
	UpdateSnapshotTransactions     float64 `perflib:"Update Snapshot Transactions"`
	VersionCleanuprateKBPers       float64 `perflib:"Version Cleanup rate (KB/s)"`
	VersionGenerationrateKBPers    float64 `perflib:"Version Generation rate (KB/s)"`
	VersionStoreSizeKB             float64 `perflib:"Version Store Size (KB)"`
	VersionStoreunitcount          float64 `perflib:"Version Store unit count"`
	VersionStoreunitcreation       float64 `perflib:"Version Store unit creation"`
	VersionStoreunittruncation     float64 `perflib:"Version Store unit truncation"`
}

// Win32_PerfRawData_MSSQLSERVER_Transactions docs:
// - https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-transactions-object
func (c *Collector) collectTransactions(ctx *types.ScrapeContext, ch chan<- prometheus.Metric, sqlInstance string) error {
	var dst []mssqlTransactions
	_ = level.Debug(c.logger).Log("msg", fmt.Sprintf("mssql_transactions collector iterating sql instance %s.", sqlInstance))

	if err := perflib.UnmarshalObject(ctx.PerfObjects[mssqlGetPerfObjectName(sqlInstance, "transactions")], &dst, c.logger); err != nil {
		return err
	}

	for _, v := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.TransactionsTempDbFreeSpaceBytes,
			prometheus.GaugeValue,
			v.FreeSpaceintempdbKB*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TransactionsLongestTransactionRunningSeconds,
			prometheus.GaugeValue,
			v.LongestTransactionRunningTime,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TransactionsNonSnapshotVersionActiveTotal,
			prometheus.CounterValue,
			v.NonSnapshotVersionTransactions,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TransactionsSnapshotActiveTotal,
			prometheus.CounterValue,
			v.SnapshotTransactions,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TransactionsActive,
			prometheus.GaugeValue,
			v.Transactions,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TransactionsUpdateConflictsTotal,
			prometheus.CounterValue,
			v.Updateconflictratio,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TransactionsUpdateSnapshotActiveTotal,
			prometheus.CounterValue,
			v.UpdateSnapshotTransactions,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TransactionsVersionCleanupRateBytes,
			prometheus.GaugeValue,
			v.VersionCleanuprateKBPers*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TransactionsVersionGenerationRateBytes,
			prometheus.GaugeValue,
			v.VersionGenerationrateKBPers*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TransactionsVersionStoreSizeBytes,
			prometheus.GaugeValue,
			v.VersionStoreSizeKB*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TransactionsVersionStoreUnits,
			prometheus.CounterValue,
			v.VersionStoreunitcount,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TransactionsVersionStoreCreationUnits,
			prometheus.CounterValue,
			v.VersionStoreunitcreation,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TransactionsVersionStoreTruncationUnits,
			prometheus.CounterValue,
			v.VersionStoreunittruncation,
			sqlInstance,
		)
	}

	return nil
}

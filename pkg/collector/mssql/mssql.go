//go:build windows

package mssql

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
	"golang.org/x/sys/windows/registry"
)

const Name = "mssql"

type Config struct {
	CollectorsEnabled []string `yaml:"collectors_enabled"`
}

var ConfigDefaults = Config{
	CollectorsEnabled: []string{
		"accessmethods",
		"availreplica",
		"bufman",
		"databases",
		"dbreplica",
		"genstats",
		"locks",
		"memmgr",
		"sqlstats",
		"sqlerrors",
		"transactions",
		"waitstats",
	},
}

type mssqlInstancesType map[string]string

func getMSSQLInstances(logger log.Logger) mssqlInstancesType {
	sqlInstances := make(mssqlInstancesType)

	// in case querying the registry fails, return the default instance
	sqlDefaultInstance := make(mssqlInstancesType)
	sqlDefaultInstance["MSSQLSERVER"] = ""

	regKey := `Software\Microsoft\Microsoft SQL Server\Instance Names\SQL`
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, regKey, registry.QUERY_VALUE)
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

// A Collector is a Prometheus Collector for various WMI Win32_PerfRawData_MSSQLSERVER_* metrics.
type Collector struct {
	config Config

	// meta
	mssqlScrapeDurationDesc *prometheus.Desc
	mssqlScrapeSuccessDesc  *prometheus.Desc

	// Win32_PerfRawData_{instance}_SQLServerAccessMethods
	accessMethodsAUcleanupbatches             *prometheus.Desc
	accessMethodsAUcleanups                   *prometheus.Desc
	accessMethodsByReferenceLobCreateCount    *prometheus.Desc
	accessMethodsByReferenceLobUseCount       *prometheus.Desc
	accessMethodsCountLobReadahead            *prometheus.Desc
	accessMethodsCountPullInRow               *prometheus.Desc
	accessMethodsCountPushOffRow              *prometheus.Desc
	accessMethodsDeferreddroppedAUs           *prometheus.Desc
	accessMethodsDeferredDroppedrowsets       *prometheus.Desc
	accessMethodsDroppedrowsetcleanups        *prometheus.Desc
	accessMethodsDroppedrowsetsskipped        *prometheus.Desc
	accessMethodsExtentDeallocations          *prometheus.Desc
	accessMethodsExtentsAllocated             *prometheus.Desc
	accessMethodsFailedAUcleanupbatches       *prometheus.Desc
	accessMethodsFailedleafpagecookie         *prometheus.Desc
	accessMethodsFailedtreepagecookie         *prometheus.Desc
	accessMethodsForwardedRecords             *prometheus.Desc
	accessMethodsFreeSpacePageFetches         *prometheus.Desc
	accessMethodsFreeSpaceScans               *prometheus.Desc
	accessMethodsFullScans                    *prometheus.Desc
	accessMethodsIndexSearches                *prometheus.Desc
	accessMethodsInSysXactwaits               *prometheus.Desc
	accessMethodsLobHandleCreateCount         *prometheus.Desc
	accessMethodsLobHandleDestroyCount        *prometheus.Desc
	accessMethodsLobSSProviderCreateCount     *prometheus.Desc
	accessMethodsLobSSProviderDestroyCount    *prometheus.Desc
	accessMethodsLobSSProviderTruncationCount *prometheus.Desc
	accessMethodsMixedPageAllocations         *prometheus.Desc
	accessMethodsPageCompressionAttempts      *prometheus.Desc
	accessMethodsPageDeallocations            *prometheus.Desc
	accessMethodsPagesAllocated               *prometheus.Desc
	accessMethodsPagesCompressed              *prometheus.Desc
	accessMethodsPageSplits                   *prometheus.Desc
	accessMethodsProbeScans                   *prometheus.Desc
	accessMethodsRangeScans                   *prometheus.Desc
	accessMethodsScanPointRevalidations       *prometheus.Desc
	accessMethodsSkippedGhostedRecords        *prometheus.Desc
	accessMethodsTableLockEscalations         *prometheus.Desc
	accessMethodsUsedleafpagecookie           *prometheus.Desc
	accessMethodsUsedtreepagecookie           *prometheus.Desc
	accessMethodsWorkfilesCreated             *prometheus.Desc
	accessMethodsWorktablesCreated            *prometheus.Desc
	accessMethodsWorktablesFromCacheHits      *prometheus.Desc
	accessMethodsWorktablesFromCacheLookups   *prometheus.Desc

	// Win32_PerfRawData_{instance}_SQLServerAvailabilityReplica
	availReplicaBytesReceivedFromReplica *prometheus.Desc
	availReplicaBytesSentToReplica       *prometheus.Desc
	availReplicaBytesSentToTransport     *prometheus.Desc
	availReplicaFlowControl              *prometheus.Desc
	availReplicaFlowControlTimeMS        *prometheus.Desc
	availReplicaReceivesFromReplica      *prometheus.Desc
	availReplicaResentMessages           *prometheus.Desc
	availReplicaSendsToReplica           *prometheus.Desc
	availReplicaSendsToTransport         *prometheus.Desc

	// Win32_PerfRawData_{instance}_SQLServerBufferManager
	bufManBackgroundwriterpages         *prometheus.Desc
	bufManBuffercachehits               *prometheus.Desc
	bufManBuffercachelookups            *prometheus.Desc
	bufManCheckpointpages               *prometheus.Desc
	bufManDatabasepages                 *prometheus.Desc
	bufManExtensionallocatedpages       *prometheus.Desc
	bufManExtensionfreepages            *prometheus.Desc
	bufManExtensioninuseaspercentage    *prometheus.Desc
	bufManExtensionoutstandingIOcounter *prometheus.Desc
	bufManExtensionpageevictions        *prometheus.Desc
	bufManExtensionpagereads            *prometheus.Desc
	bufManExtensionpageunreferencedtime *prometheus.Desc
	bufManExtensionpagewrites           *prometheus.Desc
	bufManFreeliststalls                *prometheus.Desc
	bufManIntegralControllerSlope       *prometheus.Desc
	bufManLazywrites                    *prometheus.Desc
	bufManPagelifeexpectancy            *prometheus.Desc
	bufManPagelookups                   *prometheus.Desc
	bufManPagereads                     *prometheus.Desc
	bufManPagewrites                    *prometheus.Desc
	bufManReadaheadpages                *prometheus.Desc
	bufManReadaheadtime                 *prometheus.Desc
	bufManTargetpages                   *prometheus.Desc

	// Win32_PerfRawData_{instance}_SQLServerDatabaseReplica
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

	// Win32_PerfRawData_{instance}_SQLServerDatabases
	databasesActiveParallelredothreads       *prometheus.Desc
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

	// Win32_PerfRawData_{instance}_SQLServerGeneralStatistics
	genStatsActiveTempTables              *prometheus.Desc
	genStatsConnectionReset               *prometheus.Desc
	genStatsEventNotificationsDelayedDrop *prometheus.Desc
	genStatsHTTPAuthenticatedRequests     *prometheus.Desc
	genStatsLogicalConnections            *prometheus.Desc
	genStatsLogins                        *prometheus.Desc
	genStatsLogouts                       *prometheus.Desc
	genStatsMarsDeadlocks                 *prometheus.Desc
	genStatsNonAtomicYieldRate            *prometheus.Desc
	genStatsProcessesBlocked              *prometheus.Desc
	genStatsSOAPEmptyRequests             *prometheus.Desc
	genStatsSOAPMethodInvocations         *prometheus.Desc
	genStatsSOAPSessionInitiateRequests   *prometheus.Desc
	genStatsSOAPSessionTerminateRequests  *prometheus.Desc
	genStatsSOAPSQLRequests               *prometheus.Desc
	genStatsSOAPWSDLRequests              *prometheus.Desc
	genStatsSQLTraceIOProviderLockWaits   *prometheus.Desc
	genStatsTempDBRecoveryUnitID          *prometheus.Desc
	genStatsTempDBrowSetID                *prometheus.Desc
	genStatsTempTablesCreationRate        *prometheus.Desc
	genStatsTempTablesForDestruction      *prometheus.Desc
	genStatsTraceEventNotificationQueue   *prometheus.Desc
	genStatsTransactions                  *prometheus.Desc
	genStatsUserConnections               *prometheus.Desc

	// Win32_PerfRawData_{instance}_SQLServerLocks
	locksWaitTime             *prometheus.Desc
	locksCount                *prometheus.Desc
	locksLockRequests         *prometheus.Desc
	locksLockTimeouts         *prometheus.Desc
	locksLockTimeoutstimeout0 *prometheus.Desc
	locksLockWaits            *prometheus.Desc
	locksLockWaitTimeMS       *prometheus.Desc
	locksNumberOfDeadlocks    *prometheus.Desc

	// Win32_PerfRawData_{instance}_SQLServerMemoryManager
	memMgrConnectionMemoryKB       *prometheus.Desc
	memMgrDatabaseCacheMemoryKB    *prometheus.Desc
	memMgrExternalbenefitofmemory  *prometheus.Desc
	memMgrFreeMemoryKB             *prometheus.Desc
	memMgrGrantedWorkspaceMemoryKB *prometheus.Desc
	memMgrLockBlocks               *prometheus.Desc
	memMgrLockBlocksAllocated      *prometheus.Desc
	memMgrLockMemoryKB             *prometheus.Desc
	memMgrLockOwnerBlocks          *prometheus.Desc
	memMgrLockOwnerBlocksAllocated *prometheus.Desc
	memMgrLogPoolMemoryKB          *prometheus.Desc
	memMgrMaximumWorkspaceMemoryKB *prometheus.Desc
	memMgrMemoryGrantsOutstanding  *prometheus.Desc
	memMgrMemoryGrantsPending      *prometheus.Desc
	memMgrOptimizerMemoryKB        *prometheus.Desc
	memMgrReservedServerMemoryKB   *prometheus.Desc
	memMgrSQLCacheMemoryKB         *prometheus.Desc
	memMgrStolenServerMemoryKB     *prometheus.Desc
	memMgrTargetServerMemoryKB     *prometheus.Desc
	memMgrTotalServerMemoryKB      *prometheus.Desc

	// Win32_PerfRawData_{instance}_SQLServerSQLStatistics
	sqlStatsAutoParamAttempts       *prometheus.Desc
	sqlStatsBatchRequests           *prometheus.Desc
	sqlStatsFailedAutoParams        *prometheus.Desc
	sqlStatsForcedParameterizations *prometheus.Desc
	sqlStatsGuidedplanexecutions    *prometheus.Desc
	sqlStatsMisguidedplanexecutions *prometheus.Desc
	sqlStatsSafeAutoParams          *prometheus.Desc
	sqlStatsSQLAttentionrate        *prometheus.Desc
	sqlStatsSQLCompilations         *prometheus.Desc
	sqlStatsSQLReCompilations       *prometheus.Desc
	sqlStatsUnsafeAutoParams        *prometheus.Desc

	// Win32_PerfRawData_{instance}_SQLServerSQLErrors
	sqlErrorsTotal *prometheus.Desc

	// Win32_PerfRawData_{instance}_SQLServerTransactions
	transactionsTempDbFreeSpaceBytes             *prometheus.Desc
	transactionsLongestTransactionRunningSeconds *prometheus.Desc
	transactionsNonSnapshotVersionActiveTotal    *prometheus.Desc
	transactionsSnapshotActiveTotal              *prometheus.Desc
	transactionsActive                           *prometheus.Desc
	transactionsUpdateConflictsTotal             *prometheus.Desc
	transactionsUpdateSnapshotActiveTotal        *prometheus.Desc
	transactionsVersionCleanupRateBytes          *prometheus.Desc
	transactionsVersionGenerationRateBytes       *prometheus.Desc
	transactionsVersionStoreSizeBytes            *prometheus.Desc
	transactionsVersionStoreUnits                *prometheus.Desc
	transactionsVersionStoreCreationUnits        *prometheus.Desc
	transactionsVersionStoreTruncationUnits      *prometheus.Desc

	// Win32_PerfRawData_{instance}_SQLServerWaitStatistics
	waitStatsLockWaits                     *prometheus.Desc
	waitStatsMemoryGrantQueueWaits         *prometheus.Desc
	waitStatsThreadSafeMemoryObjectsWaits  *prometheus.Desc
	waitStatsLogWriteWaits                 *prometheus.Desc
	waitStatsLogBufferWaits                *prometheus.Desc
	waitStatsNetworkIOWaits                *prometheus.Desc
	waitStatsPageIOLatchWaits              *prometheus.Desc
	waitStatsPageLatchWaits                *prometheus.Desc
	waitStatsNonPageLatchWaits             *prometheus.Desc
	waitStatsWaitForTheWorkerWaits         *prometheus.Desc
	waitStatsWorkspaceSynchronizationWaits *prometheus.Desc
	waitStatsTransactionOwnershipWaits     *prometheus.Desc

	mssqlInstances             mssqlInstancesType
	mssqlCollectors            mssqlCollectorsMap
	mssqlChildCollectorFailure int
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	if config.CollectorsEnabled == nil {
		config.CollectorsEnabled = ConfigDefaults.CollectorsEnabled
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(app *kingpin.Application) *Collector {
	c := &Collector{
		config: ConfigDefaults,
	}

	var listAllCollectors bool
	var collectorsEnabled string

	app.Flag(
		"collectors.mssql.class-print",
		"If true, print available mssql WMI classes and exit.  Only displays if the mssql collector is enabled.",
	).BoolVar(&listAllCollectors)

	app.Flag(
		"collectors.mssql.classes-enabled",
		"Comma-separated list of mssql WMI classes to use.",
	).Default(strings.Join(c.config.CollectorsEnabled, ",")).StringVar(&collectorsEnabled)

	app.PreAction(func(*kingpin.ParseContext) error {
		if listAllCollectors {
			sb := strings.Builder{}
			sb.WriteString("Available SQLServer Classes:\n - ")

			for name := range c.mssqlCollectors {
				sb.WriteString(fmt.Sprintf(" - %s\n", name))
			}

			app.UsageTemplate(sb.String()).Usage(nil)

			os.Exit(0)
		}

		return nil
	})

	app.Action(func(*kingpin.ParseContext) error {
		c.config.CollectorsEnabled = strings.Split(collectorsEnabled, ",")

		return nil
	})

	return c
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) GetPerfCounter(logger log.Logger) ([]string, error) {
	c.mssqlInstances = getMSSQLInstances(logger)
	perfCounters := make([]string, 0, len(c.mssqlInstances)*len(c.config.CollectorsEnabled))

	for instance := range c.mssqlInstances {
		for _, c := range c.config.CollectorsEnabled {
			perfCounters = append(perfCounters, mssqlGetPerfObjectName(instance, c))
		}
	}

	return perfCounters, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(_ log.Logger, _ *wmi.Client) error {
	// Result must order, to prevent test failures.
	sort.Strings(c.config.CollectorsEnabled)

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
	c.accessMethodsAUcleanupbatches = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_au_batch_cleanups"),
		"(AccessMethods.AUcleanupbatches)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsAUcleanups = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_au_cleanups"),
		"(AccessMethods.AUcleanups)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsByReferenceLobCreateCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_by_reference_lob_creates"),
		"(AccessMethods.ByreferenceLobCreateCount)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsByReferenceLobUseCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_by_reference_lob_uses"),
		"(AccessMethods.ByreferenceLobUseCount)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsCountLobReadahead = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_lob_read_aheads"),
		"(AccessMethods.CountLobReadahead)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsCountPullInRow = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_column_value_pulls"),
		"(AccessMethods.CountPullInRow)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsCountPushOffRow = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_column_value_pushes"),
		"(AccessMethods.CountPushOffRow)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsDeferreddroppedAUs = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_deferred_dropped_aus"),
		"(AccessMethods.DeferreddroppedAUs)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsDeferredDroppedrowsets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_deferred_dropped_rowsets"),
		"(AccessMethods.DeferredDroppedrowsets)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsDroppedrowsetcleanups = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_dropped_rowset_cleanups"),
		"(AccessMethods.Droppedrowsetcleanups)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsDroppedrowsetsskipped = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_dropped_rowset_skips"),
		"(AccessMethods.Droppedrowsetsskipped)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsExtentDeallocations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_extent_deallocations"),
		"(AccessMethods.ExtentDeallocations)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsExtentsAllocated = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_extent_allocations"),
		"(AccessMethods.ExtentsAllocated)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsFailedAUcleanupbatches = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_au_batch_cleanup_failures"),
		"(AccessMethods.FailedAUcleanupbatches)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsFailedleafpagecookie = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_leaf_page_cookie_failures"),
		"(AccessMethods.Failedleafpagecookie)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsFailedtreepagecookie = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_tree_page_cookie_failures"),
		"(AccessMethods.Failedtreepagecookie)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsForwardedRecords = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_forwarded_records"),
		"(AccessMethods.ForwardedRecords)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsFreeSpacePageFetches = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_free_space_page_fetches"),
		"(AccessMethods.FreeSpacePageFetches)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsFreeSpaceScans = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_free_space_scans"),
		"(AccessMethods.FreeSpaceScans)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsFullScans = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_full_scans"),
		"(AccessMethods.FullScans)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsIndexSearches = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_index_searches"),
		"(AccessMethods.IndexSearches)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsInSysXactwaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_insysxact_waits"),
		"(AccessMethods.InSysXactwaits)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsLobHandleCreateCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_lob_handle_creates"),
		"(AccessMethods.LobHandleCreateCount)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsLobHandleDestroyCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_lob_handle_destroys"),
		"(AccessMethods.LobHandleDestroyCount)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsLobSSProviderCreateCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_lob_ss_provider_creates"),
		"(AccessMethods.LobSSProviderCreateCount)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsLobSSProviderDestroyCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_lob_ss_provider_destroys"),
		"(AccessMethods.LobSSProviderDestroyCount)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsLobSSProviderTruncationCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_lob_ss_provider_truncations"),
		"(AccessMethods.LobSSProviderTruncationCount)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsMixedPageAllocations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_mixed_page_allocations"),
		"(AccessMethods.MixedpageallocationsPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsPageCompressionAttempts = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_page_compression_attempts"),
		"(AccessMethods.PagecompressionattemptsPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsPageDeallocations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_page_deallocations"),
		"(AccessMethods.PageDeallocationsPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsPagesAllocated = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_page_allocations"),
		"(AccessMethods.PagesAllocatedPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsPagesCompressed = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_page_compressions"),
		"(AccessMethods.PagescompressedPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsPageSplits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_page_splits"),
		"(AccessMethods.PageSplitsPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsProbeScans = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_probe_scans"),
		"(AccessMethods.ProbeScansPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsRangeScans = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_range_scans"),
		"(AccessMethods.RangeScansPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsScanPointRevalidations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_scan_point_revalidations"),
		"(AccessMethods.ScanPointRevalidationsPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsSkippedGhostedRecords = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_ghost_record_skips"),
		"(AccessMethods.SkippedGhostedRecordsPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsTableLockEscalations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_table_lock_escalations"),
		"(AccessMethods.TableLockEscalationsPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsUsedleafpagecookie = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_leaf_page_cookie_uses"),
		"(AccessMethods.Usedleafpagecookie)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsUsedtreepagecookie = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_tree_page_cookie_uses"),
		"(AccessMethods.Usedtreepagecookie)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsWorkfilesCreated = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_workfile_creates"),
		"(AccessMethods.WorkfilesCreatedPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsWorktablesCreated = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_worktables_creates"),
		"(AccessMethods.WorktablesCreatedPersec)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsWorktablesFromCacheHits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_worktables_from_cache_hits"),
		"(AccessMethods.WorktablesFromCacheRatio)",
		[]string{"mssql_instance"},
		nil,
	)
	c.accessMethodsWorktablesFromCacheLookups = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accessmethods_worktables_from_cache_lookups"),
		"(AccessMethods.WorktablesFromCacheRatio_Base)",
		[]string{"mssql_instance"},
		nil,
	)

	// Win32_PerfRawData_{instance}_SQLServerAvailabilityReplica
	c.availReplicaBytesReceivedFromReplica = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "availreplica_received_from_replica_bytes"),
		"(AvailabilityReplica.BytesReceivedfromReplica)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.availReplicaBytesSentToReplica = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "availreplica_sent_to_replica_bytes"),
		"(AvailabilityReplica.BytesSenttoReplica)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.availReplicaBytesSentToTransport = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "availreplica_sent_to_transport_bytes"),
		"(AvailabilityReplica.BytesSenttoTransport)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.availReplicaFlowControl = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "availreplica_initiated_flow_controls"),
		"(AvailabilityReplica.FlowControl)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.availReplicaFlowControlTimeMS = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "availreplica_flow_control_wait_seconds"),
		"(AvailabilityReplica.FlowControlTimems)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.availReplicaReceivesFromReplica = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "availreplica_receives_from_replica"),
		"(AvailabilityReplica.ReceivesfromReplica)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.availReplicaResentMessages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "availreplica_resent_messages"),
		"(AvailabilityReplica.ResentMessages)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.availReplicaSendsToReplica = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "availreplica_sends_to_replica"),
		"(AvailabilityReplica.SendstoReplica)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.availReplicaSendsToTransport = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "availreplica_sends_to_transport"),
		"(AvailabilityReplica.SendstoTransport)",
		[]string{"mssql_instance", "replica"},
		nil,
	)

	// Win32_PerfRawData_{instance}_SQLServerBufferManager
	c.bufManBackgroundwriterpages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_background_writer_pages"),
		"(BufferManager.Backgroundwriterpages)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManBuffercachehits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_buffer_cache_hits"),
		"(BufferManager.Buffercachehitratio)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManBuffercachelookups = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_buffer_cache_lookups"),
		"(BufferManager.Buffercachehitratio_Base)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManCheckpointpages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_checkpoint_pages"),
		"(BufferManager.Checkpointpages)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManDatabasepages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_database_pages"),
		"(BufferManager.Databasepages)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManExtensionallocatedpages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_extension_allocated_pages"),
		"(BufferManager.Extensionallocatedpages)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManExtensionfreepages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_extension_free_pages"),
		"(BufferManager.Extensionfreepages)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManExtensioninuseaspercentage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_extension_in_use_as_percentage"),
		"(BufferManager.Extensioninuseaspercentage)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManExtensionoutstandingIOcounter = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_extension_outstanding_io"),
		"(BufferManager.ExtensionoutstandingIOcounter)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManExtensionpageevictions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_extension_page_evictions"),
		"(BufferManager.Extensionpageevictions)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManExtensionpagereads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_extension_page_reads"),
		"(BufferManager.Extensionpagereads)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManExtensionpageunreferencedtime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_extension_page_unreferenced_seconds"),
		"(BufferManager.Extensionpageunreferencedtime)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManExtensionpagewrites = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_extension_page_writes"),
		"(BufferManager.Extensionpagewrites)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManFreeliststalls = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_free_list_stalls"),
		"(BufferManager.Freeliststalls)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManIntegralControllerSlope = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_integral_controller_slope"),
		"(BufferManager.IntegralControllerSlope)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManLazywrites = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_lazywrites"),
		"(BufferManager.Lazywrites)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManPagelifeexpectancy = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_page_life_expectancy_seconds"),
		"(BufferManager.Pagelifeexpectancy)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManPagelookups = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_page_lookups"),
		"(BufferManager.Pagelookups)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManPagereads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_page_reads"),
		"(BufferManager.Pagereads)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManPagewrites = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_page_writes"),
		"(BufferManager.Pagewrites)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManReadaheadpages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_read_ahead_pages"),
		"(BufferManager.Readaheadpages)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManReadaheadtime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_read_ahead_issuing_seconds"),
		"(BufferManager.Readaheadtime)",
		[]string{"mssql_instance"},
		nil,
	)
	c.bufManTargetpages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bufman_target_pages"),
		"(BufferManager.Targetpages)",
		[]string{"mssql_instance"},
		nil,
	)

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

	// Win32_PerfRawData_{instance}_SQLServerDatabases
	c.databasesActiveParallelredothreads = prometheus.NewDesc(
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

	// Win32_PerfRawData_{instance}_SQLServerGeneralStatistics
	c.genStatsActiveTempTables = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_active_temp_tables"),
		"(GeneralStatistics.ActiveTempTables)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsConnectionReset = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_connection_resets"),
		"(GeneralStatistics.ConnectionReset)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsEventNotificationsDelayedDrop = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_event_notifications_delayed_drop"),
		"(GeneralStatistics.EventNotificationsDelayedDrop)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsHTTPAuthenticatedRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_http_authenticated_requests"),
		"(GeneralStatistics.HTTPAuthenticatedRequests)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsLogicalConnections = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_logical_connections"),
		"(GeneralStatistics.LogicalConnections)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsLogins = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_logins"),
		"(GeneralStatistics.Logins)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsLogouts = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_logouts"),
		"(GeneralStatistics.Logouts)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsMarsDeadlocks = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_mars_deadlocks"),
		"(GeneralStatistics.MarsDeadlocks)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsNonAtomicYieldRate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_non_atomic_yields"),
		"(GeneralStatistics.Nonatomicyields)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsProcessesBlocked = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_blocked_processes"),
		"(GeneralStatistics.Processesblocked)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsSOAPEmptyRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_soap_empty_requests"),
		"(GeneralStatistics.SOAPEmptyRequests)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsSOAPMethodInvocations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_soap_method_invocations"),
		"(GeneralStatistics.SOAPMethodInvocations)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsSOAPSessionInitiateRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_soap_session_initiate_requests"),
		"(GeneralStatistics.SOAPSessionInitiateRequests)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsSOAPSessionTerminateRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_soap_session_terminate_requests"),
		"(GeneralStatistics.SOAPSessionTerminateRequests)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsSOAPSQLRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_soapsql_requests"),
		"(GeneralStatistics.SOAPSQLRequests)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsSOAPWSDLRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_soapwsdl_requests"),
		"(GeneralStatistics.SOAPWSDLRequests)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsSQLTraceIOProviderLockWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_sql_trace_io_provider_lock_waits"),
		"(GeneralStatistics.SQLTraceIOProviderLockWaits)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsTempDBRecoveryUnitID = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_tempdb_recovery_unit_ids_generated"),
		"(GeneralStatistics.Tempdbrecoveryunitid)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsTempDBrowSetID = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_tempdb_rowset_ids_generated"),
		"(GeneralStatistics.Tempdbrowsetid)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsTempTablesCreationRate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_temp_tables_creations"),
		"(GeneralStatistics.TempTablesCreations)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsTempTablesForDestruction = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_temp_tables_awaiting_destruction"),
		"(GeneralStatistics.TempTablesForDestruction)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsTraceEventNotificationQueue = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_trace_event_notification_queue_size"),
		"(GeneralStatistics.TraceEventNotificationQueue)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsTransactions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_transactions"),
		"(GeneralStatistics.Transactions)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsUserConnections = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_user_connections"),
		"(GeneralStatistics.UserConnections)",
		[]string{"mssql_instance"},
		nil,
	)

	// Win32_PerfRawData_{instance}_SQLServerLocks
	c.locksWaitTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "locks_wait_time_seconds"),
		"(Locks.AverageWaitTimems Total time in seconds which locks have been holding resources)",
		[]string{"mssql_instance", "resource"},
		nil,
	)
	c.locksCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "locks_count"),
		"(Locks.AverageWaitTimems_Base count of how often requests have run into locks)",
		[]string{"mssql_instance", "resource"},
		nil,
	)
	c.locksLockRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "locks_lock_requests"),
		"(Locks.LockRequests)",
		[]string{"mssql_instance", "resource"},
		nil,
	)
	c.locksLockTimeouts = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "locks_lock_timeouts"),
		"(Locks.LockTimeouts)",
		[]string{"mssql_instance", "resource"},
		nil,
	)
	c.locksLockTimeoutstimeout0 = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "locks_lock_timeouts_excluding_NOWAIT"),
		"(Locks.LockTimeoutstimeout0)",
		[]string{"mssql_instance", "resource"},
		nil,
	)
	c.locksLockWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "locks_lock_waits"),
		"(Locks.LockWaits)",
		[]string{"mssql_instance", "resource"},
		nil,
	)
	c.locksLockWaitTimeMS = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "locks_lock_wait_seconds"),
		"(Locks.LockWaitTimems)",
		[]string{"mssql_instance", "resource"},
		nil,
	)
	c.locksNumberOfDeadlocks = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "locks_deadlocks"),
		"(Locks.NumberOfDeadlocks)",
		[]string{"mssql_instance", "resource"},
		nil,
	)

	// Win32_PerfRawData_{instance}_SQLServerMemoryManager
	c.memMgrConnectionMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_connection_memory_bytes"),
		"(MemoryManager.ConnectionMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrDatabaseCacheMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_database_cache_memory_bytes"),
		"(MemoryManager.DatabaseCacheMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrExternalbenefitofmemory = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_external_benefit_of_memory"),
		"(MemoryManager.Externalbenefitofmemory)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrFreeMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_free_memory_bytes"),
		"(MemoryManager.FreeMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrGrantedWorkspaceMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_granted_workspace_memory_bytes"),
		"(MemoryManager.GrantedWorkspaceMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrLockBlocks = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_lock_blocks"),
		"(MemoryManager.LockBlocks)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrLockBlocksAllocated = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_allocated_lock_blocks"),
		"(MemoryManager.LockBlocksAllocated)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrLockMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_lock_memory_bytes"),
		"(MemoryManager.LockMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrLockOwnerBlocks = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_lock_owner_blocks"),
		"(MemoryManager.LockOwnerBlocks)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrLockOwnerBlocksAllocated = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_allocated_lock_owner_blocks"),
		"(MemoryManager.LockOwnerBlocksAllocated)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrLogPoolMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_log_pool_memory_bytes"),
		"(MemoryManager.LogPoolMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrMaximumWorkspaceMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_maximum_workspace_memory_bytes"),
		"(MemoryManager.MaximumWorkspaceMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrMemoryGrantsOutstanding = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_outstanding_memory_grants"),
		"(MemoryManager.MemoryGrantsOutstanding)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrMemoryGrantsPending = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_pending_memory_grants"),
		"(MemoryManager.MemoryGrantsPending)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrOptimizerMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_optimizer_memory_bytes"),
		"(MemoryManager.OptimizerMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrReservedServerMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_reserved_server_memory_bytes"),
		"(MemoryManager.ReservedServerMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrSQLCacheMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_sql_cache_memory_bytes"),
		"(MemoryManager.SQLCacheMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrStolenServerMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_stolen_server_memory_bytes"),
		"(MemoryManager.StolenServerMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrTargetServerMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_target_server_memory_bytes"),
		"(MemoryManager.TargetServerMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrTotalServerMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_total_server_memory_bytes"),
		"(MemoryManager.TotalServerMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)

	// Win32_PerfRawData_{instance}_SQLServerSQLStatistics
	c.sqlStatsAutoParamAttempts = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_auto_parameterization_attempts"),
		"(SQLStatistics.AutoParamAttempts)",
		[]string{"mssql_instance"},
		nil,
	)
	c.sqlStatsBatchRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_batch_requests"),
		"(SQLStatistics.BatchRequests)",
		[]string{"mssql_instance"},
		nil,
	)
	c.sqlStatsFailedAutoParams = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_failed_auto_parameterization_attempts"),
		"(SQLStatistics.FailedAutoParams)",
		[]string{"mssql_instance"},
		nil,
	)
	c.sqlStatsForcedParameterizations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_forced_parameterizations"),
		"(SQLStatistics.ForcedParameterizations)",
		[]string{"mssql_instance"},
		nil,
	)
	c.sqlStatsGuidedplanexecutions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_guided_plan_executions"),
		"(SQLStatistics.Guidedplanexecutions)",
		[]string{"mssql_instance"},
		nil,
	)
	c.sqlStatsMisguidedplanexecutions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_misguided_plan_executions"),
		"(SQLStatistics.Misguidedplanexecutions)",
		[]string{"mssql_instance"},
		nil,
	)
	c.sqlStatsSafeAutoParams = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_safe_auto_parameterization_attempts"),
		"(SQLStatistics.SafeAutoParams)",
		[]string{"mssql_instance"},
		nil,
	)
	c.sqlStatsSQLAttentionrate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_sql_attentions"),
		"(SQLStatistics.SQLAttentions)",
		[]string{"mssql_instance"},
		nil,
	)
	c.sqlStatsSQLCompilations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_sql_compilations"),
		"(SQLStatistics.SQLCompilations)",
		[]string{"mssql_instance"},
		nil,
	)
	c.sqlStatsSQLReCompilations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_sql_recompilations"),
		"(SQLStatistics.SQLReCompilations)",
		[]string{"mssql_instance"},
		nil,
	)
	c.sqlStatsUnsafeAutoParams = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_unsafe_auto_parameterization_attempts"),
		"(SQLStatistics.UnsafeAutoParams)",
		[]string{"mssql_instance"},
		nil,
	)

	// Win32_PerfRawData_{instance}_SQLServerSQLErrors
	c.sqlErrorsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sql_errors_total"),
		"(SQLErrors.Total)",
		[]string{"mssql_instance", "resource"},
		nil,
	)

	// Win32_PerfRawData_{instance}_SQLServerTransactions
	c.transactionsTempDbFreeSpaceBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_tempdb_free_space_bytes"),
		"(Transactions.FreeSpaceInTempDbKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.transactionsLongestTransactionRunningSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_longest_transaction_running_seconds"),
		"(Transactions.LongestTransactionRunningTime)",
		[]string{"mssql_instance"},
		nil,
	)
	c.transactionsNonSnapshotVersionActiveTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_nonsnapshot_version_active_total"),
		"(Transactions.NonSnapshotVersionTransactions)",
		[]string{"mssql_instance"},
		nil,
	)
	c.transactionsSnapshotActiveTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_snapshot_active_total"),
		"(Transactions.SnapshotTransactions)",
		[]string{"mssql_instance"},
		nil,
	)
	c.transactionsActive = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_active"),
		"(Transactions.Transactions)",
		[]string{"mssql_instance"},
		nil,
	)
	c.transactionsUpdateConflictsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_update_conflicts_total"),
		"(Transactions.UpdateConflictRatio)",
		[]string{"mssql_instance"},
		nil,
	)
	c.transactionsUpdateSnapshotActiveTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_update_snapshot_active_total"),
		"(Transactions.UpdateSnapshotTransactions)",
		[]string{"mssql_instance"},
		nil,
	)
	c.transactionsVersionCleanupRateBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_version_cleanup_rate_bytes"),
		"(Transactions.VersionCleanupRateKBs)",
		[]string{"mssql_instance"},
		nil,
	)
	c.transactionsVersionGenerationRateBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_version_generation_rate_bytes"),
		"(Transactions.VersionGenerationRateKBs)",
		[]string{"mssql_instance"},
		nil,
	)
	c.transactionsVersionStoreSizeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_version_store_size_bytes"),
		"(Transactions.VersionStoreSizeKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.transactionsVersionStoreUnits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_version_store_units"),
		"(Transactions.VersionStoreUnitCount)",
		[]string{"mssql_instance"},
		nil,
	)
	c.transactionsVersionStoreCreationUnits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_version_store_creation_units"),
		"(Transactions.VersionStoreUnitCreation)",
		[]string{"mssql_instance"},
		nil,
	)
	c.transactionsVersionStoreTruncationUnits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_version_store_truncation_units"),
		"(Transactions.VersionStoreUnitTruncation)",
		[]string{"mssql_instance"},
		nil,
	)

	// Win32_PerfRawData_{instance}_SQLServerWaitStatistics
	c.waitStatsLockWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_lock_waits"),
		"(WaitStats.LockWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)

	c.waitStatsMemoryGrantQueueWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_memory_grant_queue_waits"),
		"(WaitStats.MemoryGrantQueueWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)
	c.waitStatsThreadSafeMemoryObjectsWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_thread_safe_memory_objects_waits"),
		"(WaitStats.ThreadSafeMemoryObjectsWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)
	c.waitStatsLogWriteWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_log_write_waits"),
		"(WaitStats.LogWriteWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)
	c.waitStatsLogBufferWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_log_buffer_waits"),
		"(WaitStats.LogBufferWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)
	c.waitStatsNetworkIOWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_network_io_waits"),
		"(WaitStats.NetworkIOWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)
	c.waitStatsPageIOLatchWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_page_io_latch_waits"),
		"(WaitStats.PageIOLatchWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)
	c.waitStatsPageLatchWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_page_latch_waits"),
		"(WaitStats.PageLatchWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)
	c.waitStatsNonPageLatchWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_nonpage_latch_waits"),
		"(WaitStats.NonpageLatchWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)
	c.waitStatsWaitForTheWorkerWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_wait_for_the_worker_waits"),
		"(WaitStats.WaitForTheWorkerWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)
	c.waitStatsWorkspaceSynchronizationWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_workspace_synchronization_waits"),
		"(WaitStats.WorkspaceSynchronizationWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)
	c.waitStatsTransactionOwnershipWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_transaction_ownership_waits"),
		"(WaitStats.TransactionOwnershipWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)

	c.mssqlCollectors = c.getMSSQLCollectors()

	for _, name := range c.config.CollectorsEnabled {
		if _, ok := c.mssqlCollectors[name]; !ok {
			return errors.New("unknown mssql collector: " + name)
		}
	}

	return nil
}

type mssqlCollectorFunc func(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric, sqlInstance string) error

func (c *Collector) execute(ctx *types.ScrapeContext, logger log.Logger, name string, fn mssqlCollectorFunc, ch chan<- prometheus.Metric, sqlInstance string, wg *sync.WaitGroup) {
	// Reset failure counter on each scrape
	c.mssqlChildCollectorFailure = 0
	defer wg.Done()

	begin := time.Now()
	err := fn(ctx, logger, ch, sqlInstance)
	duration := time.Since(begin)
	var success float64

	if err != nil {
		_ = level.Error(logger).Log("msg", fmt.Sprintf("mssql class collector %s failed after %fs", name, duration.Seconds()), "err", err)
		success = 0
		c.mssqlChildCollectorFailure++
	} else {
		_ = level.Debug(logger).Log("msg", fmt.Sprintf("mssql class collector %s succeeded after %fs.", name, duration.Seconds()))
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
func (c *Collector) Collect(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	wg := sync.WaitGroup{}

	for sqlInstance := range c.mssqlInstances {
		for _, name := range c.config.CollectorsEnabled {
			function := c.mssqlCollectors[name]

			wg.Add(1)
			go c.execute(ctx, logger, name, function, ch, sqlInstance, &wg)
		}
	}

	wg.Wait()

	// this should return an error if any? some? children errored.
	if c.mssqlChildCollectorFailure > 0 {
		return errors.New("at least one child collector failed")
	}

	return nil
}

// Win32_PerfRawData_MSSQLSERVER_SQLServerAccessMethods docs:
// - https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-access-methods-object
type mssqlAccessMethods struct {
	AUcleanupbatchesPerSec        float64 `perflib:"AU cleanup batches/sec"`
	AUcleanupsPerSec              float64 `perflib:"AU cleanups/sec"`
	ByReferenceLobCreateCount     float64 `perflib:"By-reference Lob Create Count"`
	ByReferenceLobUseCount        float64 `perflib:"By-reference Lob Use Count"`
	CountLobReadahead             float64 `perflib:"Count Lob Readahead"`
	CountPullInRow                float64 `perflib:"Count Pull In Row"`
	CountPushOffRow               float64 `perflib:"Count Push Off Row"`
	DeferreddroppedAUs            float64 `perflib:"Deferred dropped AUs"`
	DeferredDroppedrowsets        float64 `perflib:"Deferred Dropped rowsets"`
	DroppedrowsetcleanupsPerSec   float64 `perflib:"Dropped rowset cleanups/sec"`
	DroppedrowsetsskippedPerSec   float64 `perflib:"Dropped rowsets skipped/sec"`
	ExtentDeallocationsPerSec     float64 `perflib:"Extent Deallocations/sec"`
	ExtentsAllocatedPerSec        float64 `perflib:"Extents Allocated/sec"`
	FailedAUcleanupbatchesPerSec  float64 `perflib:"Failed AU cleanup batches/sec"`
	FailedLeafPageCookie          float64 `perflib:"Failed leaf page cookie"`
	Failedtreepagecookie          float64 `perflib:"Failed tree page cookie"`
	ForwardedRecordsPerSec        float64 `perflib:"Forwarded Records/sec"`
	FreeSpacePageFetchesPerSec    float64 `perflib:"FreeSpace Page Fetches/sec"`
	FreeSpaceScansPerSec          float64 `perflib:"FreeSpace Scans/sec"`
	FullScansPerSec               float64 `perflib:"Full Scans/sec"`
	IndexSearchesPerSec           float64 `perflib:"Index Searches/sec"`
	InSysXactwaitsPerSec          float64 `perflib:"InSysXact waits/sec"`
	LobHandleCreateCount          float64 `perflib:"LobHandle Create Count"`
	LobHandleDestroyCount         float64 `perflib:"LobHandle Destroy Count"`
	LobSSProviderCreateCount      float64 `perflib:"LobSS Provider Create Count"`
	LobSSProviderDestroyCount     float64 `perflib:"LobSS Provider Destroy Count"`
	LobSSProviderTruncationCount  float64 `perflib:"LobSS Provider Truncation Count"`
	MixedpageallocationsPerSec    float64 `perflib:"Mixed page allocations/sec"`
	PagecompressionattemptsPerSec float64 `perflib:"Page compression attempts/sec"`
	PageDeallocationsPerSec       float64 `perflib:"Page Deallocations/sec"`
	PagesAllocatedPerSec          float64 `perflib:"Pages Allocated/sec"`
	PagesCompressedPerSec         float64 `perflib:"Pages compressed/sec"`
	PageSplitsPerSec              float64 `perflib:"Page Splits/sec"`
	ProbeScansPerSec              float64 `perflib:"Probe Scans/sec"`
	RangeScansPerSec              float64 `perflib:"Range Scans/sec"`
	ScanPointRevalidationsPerSec  float64 `perflib:"Scan Point Revalidations/sec"`
	SkippedGhostedRecordsPerSec   float64 `perflib:"Skipped Ghosted Records/sec"`
	TableLockEscalationsPerSec    float64 `perflib:"Table Lock Escalations/sec"`
	UsedLeafPageCookie            float64 `perflib:"Used leaf page cookie"`
	UsedTreePageCookie            float64 `perflib:"Used tree page cookie"`
	WorkfilesCreatedPerSec        float64 `perflib:"Workfiles Created/sec"`
	WorktablesCreatedPerSec       float64 `perflib:"Worktables Created/sec"`
	WorktablesFromCacheRatio      float64 `perflib:"Worktables From Cache Ratio"`
	WorktablesFromCacheRatioBase  float64 `perflib:"Worktables From Cache Base_Base"`
}

func (c *Collector) collectAccessMethods(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric, sqlInstance string) error {
	var dst []mssqlAccessMethods
	_ = level.Debug(logger).Log("msg", fmt.Sprintf("mssql_accessmethods collector iterating sql instance %s.", sqlInstance))

	if err := perflib.UnmarshalObject(ctx.PerfObjects[mssqlGetPerfObjectName(sqlInstance, "accessmethods")], &dst, logger); err != nil {
		return err
	}

	for _, v := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsAUcleanupbatches,
			prometheus.CounterValue,
			v.AUcleanupbatchesPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsAUcleanups,
			prometheus.CounterValue,
			v.AUcleanupsPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsByReferenceLobCreateCount,
			prometheus.CounterValue,
			v.ByReferenceLobCreateCount,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsByReferenceLobUseCount,
			prometheus.CounterValue,
			v.ByReferenceLobUseCount,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsCountLobReadahead,
			prometheus.CounterValue,
			v.CountLobReadahead,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsCountPullInRow,
			prometheus.CounterValue,
			v.CountPullInRow,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsCountPushOffRow,
			prometheus.CounterValue,
			v.CountPushOffRow,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsDeferreddroppedAUs,
			prometheus.GaugeValue,
			v.DeferreddroppedAUs,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsDeferredDroppedrowsets,
			prometheus.GaugeValue,
			v.DeferredDroppedrowsets,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsDroppedrowsetcleanups,
			prometheus.CounterValue,
			v.DroppedrowsetcleanupsPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsDroppedrowsetsskipped,
			prometheus.CounterValue,
			v.DroppedrowsetsskippedPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsExtentDeallocations,
			prometheus.CounterValue,
			v.ExtentDeallocationsPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsExtentsAllocated,
			prometheus.CounterValue,
			v.ExtentsAllocatedPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsFailedAUcleanupbatches,
			prometheus.CounterValue,
			v.FailedAUcleanupbatchesPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsFailedleafpagecookie,
			prometheus.CounterValue,
			v.FailedLeafPageCookie,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsFailedtreepagecookie,
			prometheus.CounterValue,
			v.Failedtreepagecookie,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsForwardedRecords,
			prometheus.CounterValue,
			v.ForwardedRecordsPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsFreeSpacePageFetches,
			prometheus.CounterValue,
			v.FreeSpacePageFetchesPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsFreeSpaceScans,
			prometheus.CounterValue,
			v.FreeSpaceScansPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsFullScans,
			prometheus.CounterValue,
			v.FullScansPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsIndexSearches,
			prometheus.CounterValue,
			v.IndexSearchesPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsInSysXactwaits,
			prometheus.CounterValue,
			v.InSysXactwaitsPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsLobHandleCreateCount,
			prometheus.CounterValue,
			v.LobHandleCreateCount,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsLobHandleDestroyCount,
			prometheus.CounterValue,
			v.LobHandleDestroyCount,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsLobSSProviderCreateCount,
			prometheus.CounterValue,
			v.LobSSProviderCreateCount,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsLobSSProviderDestroyCount,
			prometheus.CounterValue,
			v.LobSSProviderDestroyCount,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsLobSSProviderTruncationCount,
			prometheus.CounterValue,
			v.LobSSProviderTruncationCount,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsMixedPageAllocations,
			prometheus.CounterValue,
			v.MixedpageallocationsPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsPageCompressionAttempts,
			prometheus.CounterValue,
			v.PagecompressionattemptsPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsPageDeallocations,
			prometheus.CounterValue,
			v.PageDeallocationsPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsPagesAllocated,
			prometheus.CounterValue,
			v.PagesAllocatedPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsPagesCompressed,
			prometheus.CounterValue,
			v.PagesCompressedPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsPageSplits,
			prometheus.CounterValue,
			v.PageSplitsPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsProbeScans,
			prometheus.CounterValue,
			v.ProbeScansPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsRangeScans,
			prometheus.CounterValue,
			v.RangeScansPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsScanPointRevalidations,
			prometheus.CounterValue,
			v.ScanPointRevalidationsPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsSkippedGhostedRecords,
			prometheus.CounterValue,
			v.SkippedGhostedRecordsPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsTableLockEscalations,
			prometheus.CounterValue,
			v.TableLockEscalationsPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsUsedleafpagecookie,
			prometheus.CounterValue,
			v.UsedLeafPageCookie,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsUsedtreepagecookie,
			prometheus.CounterValue,
			v.UsedTreePageCookie,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsWorkfilesCreated,
			prometheus.CounterValue,
			v.WorkfilesCreatedPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsWorktablesCreated,
			prometheus.CounterValue,
			v.WorktablesCreatedPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsWorktablesFromCacheHits,
			prometheus.CounterValue,
			v.WorktablesFromCacheRatio,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.accessMethodsWorktablesFromCacheLookups,
			prometheus.CounterValue,
			v.WorktablesFromCacheRatioBase,
			sqlInstance,
		)
	}
	return nil
}

// Win32_PerfRawData_MSSQLSERVER_SQLServerAvailabilityReplica docs:
// - https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-availability-replica
type mssqlAvailabilityReplica struct {
	Name                           string
	BytesReceivedfromReplicaPerSec float64 `perflib:"Bytes Received from Replica/sec"`
	BytesSentToReplicaPerSec       float64 `perflib:"Bytes Sent to Replica/sec"`
	BytesSentToTransportPerSec     float64 `perflib:"Bytes Sent to Transport/sec"`
	FlowControlPerSec              float64 `perflib:"Flow Control/sec"`
	FlowControlTimeMSPerSec        float64 `perflib:"Flow Control Time (ms/sec)"`
	ReceivesfromReplicaPerSec      float64 `perflib:"Receives from Replica/sec"`
	ResentMessagesPerSec           float64 `perflib:"Resent Messages/sec"`
	SendstoReplicaPerSec           float64 `perflib:"Sends to Replica/sec"`
	SendstoTransportPerSec         float64 `perflib:"Sends to Transport/sec"`
}

func (c *Collector) collectAvailabilityReplica(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric, sqlInstance string) error {
	var dst []mssqlAvailabilityReplica
	_ = level.Debug(logger).Log("msg", fmt.Sprintf("mssql_availreplica collector iterating sql instance %s.", sqlInstance))

	if err := perflib.UnmarshalObject(ctx.PerfObjects[mssqlGetPerfObjectName(sqlInstance, "availreplica")], &dst, logger); err != nil {
		return err
	}

	for _, v := range dst {
		if strings.ToLower(v.Name) == "_total" {
			continue
		}
		replicaName := v.Name

		ch <- prometheus.MustNewConstMetric(
			c.availReplicaBytesReceivedFromReplica,
			prometheus.CounterValue,
			v.BytesReceivedfromReplicaPerSec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.availReplicaBytesSentToReplica,
			prometheus.CounterValue,
			v.BytesSentToReplicaPerSec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.availReplicaBytesSentToTransport,
			prometheus.CounterValue,
			v.BytesSentToTransportPerSec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.availReplicaFlowControl,
			prometheus.CounterValue,
			v.FlowControlPerSec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.availReplicaFlowControlTimeMS,
			prometheus.CounterValue,
			v.FlowControlTimeMSPerSec/1000.0,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.availReplicaReceivesFromReplica,
			prometheus.CounterValue,
			v.ReceivesfromReplicaPerSec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.availReplicaResentMessages,
			prometheus.CounterValue,
			v.ResentMessagesPerSec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.availReplicaSendsToReplica,
			prometheus.CounterValue,
			v.SendstoReplicaPerSec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.availReplicaSendsToTransport,
			prometheus.CounterValue,
			v.SendstoTransportPerSec,
			sqlInstance, replicaName,
		)
	}
	return nil
}

// Win32_PerfRawData_MSSQLSERVER_SQLServerBufferManager docs:
// - https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-buffer-manager-object
type mssqlBufferManager struct {
	BackgroundWriterPagesPerSec   float64 `perflib:"Background writer pages/sec"`
	BufferCacheHitRatio           float64 `perflib:"Buffer cache hit ratio"`
	BufferCacheHitRatioBase       float64 `perflib:"Buffer cache hit ratio base_Base"`
	CheckpointpagesPerSec         float64 `perflib:"Checkpoint pages/sec"`
	Databasepages                 float64 `perflib:"Database pages"`
	Extensionallocatedpages       float64 `perflib:"Extension allocated pages"`
	Extensionfreepages            float64 `perflib:"Extension free pages"`
	Extensioninuseaspercentage    float64 `perflib:"Extension in use as percentage"`
	ExtensionoutstandingIOcounter float64 `perflib:"Extension outstanding IO counter"`
	ExtensionpageevictionsPerSec  float64 `perflib:"Extension page evictions/sec"`
	ExtensionpagereadsPerSec      float64 `perflib:"Extension page reads/sec"`
	Extensionpageunreferencedtime float64 `perflib:"Extension page unreferenced time"`
	ExtensionpagewritesPerSec     float64 `perflib:"Extension page writes/sec"`
	FreeliststallsPerSec          float64 `perflib:"Free list stalls/sec"`
	IntegralControllerSlope       float64 `perflib:"Integral Controller Slope"`
	LazywritesPerSec              float64 `perflib:"Lazy writes/sec"`
	Pagelifeexpectancy            float64 `perflib:"Page life expectancy"`
	PagelookupsPerSec             float64 `perflib:"Page lookups/sec"`
	PagereadsPerSec               float64 `perflib:"Page reads/sec"`
	PagewritesPerSec              float64 `perflib:"Page writes/sec"`
	ReadaheadpagesPerSec          float64 `perflib:"Readahead pages/sec"`
	ReadaheadtimePerSec           float64 `perflib:"Readahead time/sec"`
	TargetPages                   float64 `perflib:"Target pages"`
}

func (c *Collector) collectBufferManager(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric, sqlInstance string) error {
	var dst []mssqlBufferManager
	_ = level.Debug(logger).Log("msg", fmt.Sprintf("mssql_bufman collector iterating sql instance %s.", sqlInstance))

	if err := perflib.UnmarshalObject(ctx.PerfObjects[mssqlGetPerfObjectName(sqlInstance, "bufman")], &dst, logger); err != nil {
		return err
	}

	for _, v := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.bufManBackgroundwriterpages,
			prometheus.CounterValue,
			v.BackgroundWriterPagesPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManBuffercachehits,
			prometheus.GaugeValue,
			v.BufferCacheHitRatio,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManBuffercachelookups,
			prometheus.GaugeValue,
			v.BufferCacheHitRatioBase,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManCheckpointpages,
			prometheus.CounterValue,
			v.CheckpointpagesPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManDatabasepages,
			prometheus.GaugeValue,
			v.Databasepages,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManExtensionallocatedpages,
			prometheus.GaugeValue,
			v.Extensionallocatedpages,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManExtensionfreepages,
			prometheus.GaugeValue,
			v.Extensionfreepages,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManExtensioninuseaspercentage,
			prometheus.GaugeValue,
			v.Extensioninuseaspercentage,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManExtensionoutstandingIOcounter,
			prometheus.GaugeValue,
			v.ExtensionoutstandingIOcounter,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManExtensionpageevictions,
			prometheus.CounterValue,
			v.ExtensionpageevictionsPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManExtensionpagereads,
			prometheus.CounterValue,
			v.ExtensionpagereadsPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManExtensionpageunreferencedtime,
			prometheus.GaugeValue,
			v.Extensionpageunreferencedtime,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManExtensionpagewrites,
			prometheus.CounterValue,
			v.ExtensionpagewritesPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManFreeliststalls,
			prometheus.CounterValue,
			v.FreeliststallsPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManIntegralControllerSlope,
			prometheus.GaugeValue,
			v.IntegralControllerSlope,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManLazywrites,
			prometheus.CounterValue,
			v.LazywritesPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManPagelifeexpectancy,
			prometheus.GaugeValue,
			v.Pagelifeexpectancy,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManPagelookups,
			prometheus.CounterValue,
			v.PagelookupsPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManPagereads,
			prometheus.CounterValue,
			v.PagereadsPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManPagewrites,
			prometheus.CounterValue,
			v.PagewritesPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManReadaheadpages,
			prometheus.CounterValue,
			v.ReadaheadpagesPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManReadaheadtime,
			prometheus.CounterValue,
			v.ReadaheadtimePerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bufManTargetpages,
			prometheus.GaugeValue,
			v.TargetPages,
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
	DatabaseFlowControlsPerSec      float64 `perflib:"Database Flow Controls/sec"`
	FileBytesReceivedPerSec         float64 `perflib:"File Bytes Received/sec"`
	GroupCommitsPerSec              float64 `perflib:"Group Commits/Sec"`
	GroupCommitTime                 float64 `perflib:"Group Commit Time"`
	LogApplyPendingQueue            float64 `perflib:"Log Apply Pending Queue"`
	LogApplyReadyQueue              float64 `perflib:"Log Apply Ready Queue"`
	LogBytesCompressedPerSec        float64 `perflib:"Log Bytes Compressed/sec"`
	LogBytesDecompressedPerSec      float64 `perflib:"Log Bytes Decompressed/sec"`
	LogBytesReceivedPerSec          float64 `perflib:"Log Bytes Received/sec"`
	LogCompressionCachehitsPerSec   float64 `perflib:"Log Compression Cache hits/sec"`
	LogCompressionCachemissesPerSec float64 `perflib:"Log Compression Cache misses/sec"`
	LogCompressionsPerSec           float64 `perflib:"Log Compressions/sec"`
	LogDecompressionsPerSec         float64 `perflib:"Log Decompressions/sec"`
	Logremainingforundo             float64 `perflib:"Log remaining for undo"`
	LogSendQueue                    float64 `perflib:"Log Send Queue"`
	MirroredWriteTransactionsPerSec float64 `perflib:"Mirrored Write Transactions/sec"`
	RecoveryQueue                   float64 `perflib:"Recovery Queue"`
	RedoblockedPerSec               float64 `perflib:"Redo blocked/sec"`
	RedoBytesRemaining              float64 `perflib:"Redo Bytes Remaining"`
	RedoneBytesPerSec               float64 `perflib:"Redone Bytes/sec"`
	RedonesPerSec                   float64 `perflib:"Redones/sec"`
	TotalLogrequiringundo           float64 `perflib:"Total Log requiring undo"`
	TransactionDelay                float64 `perflib:"Transaction Delay"`
}

func (c *Collector) collectDatabaseReplica(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric, sqlInstance string) error {
	var dst []mssqlDatabaseReplica
	_ = level.Debug(logger).Log("msg", fmt.Sprintf("mssql_dbreplica collector iterating sql instance %s.", sqlInstance))

	if err := perflib.UnmarshalObject(ctx.PerfObjects[mssqlGetPerfObjectName(sqlInstance, "dbreplica")], &dst, logger); err != nil {
		return err
	}

	for _, v := range dst {
		if strings.ToLower(v.Name) == "_total" {
			continue
		}
		replicaName := v.Name

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaDatabaseFlowControlDelay,
			prometheus.GaugeValue,
			v.DatabaseFlowControlDelay,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaDatabaseFlowControls,
			prometheus.CounterValue,
			v.DatabaseFlowControlsPerSec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaFileBytesReceived,
			prometheus.CounterValue,
			v.FileBytesReceivedPerSec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaGroupCommits,
			prometheus.CounterValue,
			v.GroupCommitsPerSec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaGroupCommitTime,
			prometheus.GaugeValue,
			v.GroupCommitTime,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogApplyPendingQueue,
			prometheus.GaugeValue,
			v.LogApplyPendingQueue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogApplyReadyQueue,
			prometheus.GaugeValue,
			v.LogApplyReadyQueue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogBytesCompressed,
			prometheus.CounterValue,
			v.LogBytesCompressedPerSec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogBytesDecompressed,
			prometheus.CounterValue,
			v.LogBytesDecompressedPerSec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogBytesReceived,
			prometheus.CounterValue,
			v.LogBytesReceivedPerSec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogCompressionCachehits,
			prometheus.CounterValue,
			v.LogCompressionCachehitsPerSec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogCompressionCachemisses,
			prometheus.CounterValue,
			v.LogCompressionCachemissesPerSec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogCompressions,
			prometheus.CounterValue,
			v.LogCompressionsPerSec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogDecompressions,
			prometheus.CounterValue,
			v.LogDecompressionsPerSec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogremainingforundo,
			prometheus.GaugeValue,
			v.Logremainingforundo,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaLogSendQueue,
			prometheus.GaugeValue,
			v.LogSendQueue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaMirroredWritetransactions,
			prometheus.CounterValue,
			v.MirroredWriteTransactionsPerSec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaRecoveryQueue,
			prometheus.GaugeValue,
			v.RecoveryQueue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaRedoblocked,
			prometheus.CounterValue,
			v.RedoblockedPerSec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaRedoBytesRemaining,
			prometheus.GaugeValue,
			v.RedoBytesRemaining,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaRedoneBytes,
			prometheus.CounterValue,
			v.RedoneBytesPerSec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaRedones,
			prometheus.CounterValue,
			v.RedonesPerSec,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaTotalLogrequiringundo,
			prometheus.GaugeValue,
			v.TotalLogrequiringundo,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dbReplicaTransactionDelay,
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
	ActiveParallelRedoThreads        float64 `perflib:"Active parallel redo threads"`
	ActiveTransactions               float64 `perflib:"Active Transactions"`
	BackupPerRestoreThroughputPerSec float64 `perflib:"Backup/Restore Throughput/sec"`
	BulkCopyRowsPerSec               float64 `perflib:"Bulk Copy Rows/sec"`
	BulkCopyThroughputPerSec         float64 `perflib:"Bulk Copy Throughput/sec"`
	CommitTableEntries               float64 `perflib:"Commit table entries"`
	DataFilesSizeKB                  float64 `perflib:"Data File(s) Size (KB)"`
	DBCCLogicalScanBytesPerSec       float64 `perflib:"DBCC Logical Scan Bytes/sec"`
	GroupCommitTimePerSec            float64 `perflib:"Group Commit Time/sec"`
	LogBytesFlushedPerSec            float64 `perflib:"Log Bytes Flushed/sec"`
	LogCacheHitRatio                 float64 `perflib:"Log Cache Hit Ratio"`
	LogCacheHitRatioBase             float64 `perflib:"Log Cache Hit Ratio Base_Base"`
	LogCacheReadsPerSec              float64 `perflib:"Log Cache Reads/sec"`
	LogFilesSizeKB                   float64 `perflib:"Log File(s) Size (KB)"`
	LogFilesUsedSizeKB               float64 `perflib:"Log File(s) Used Size (KB)"`
	LogFlushesPerSec                 float64 `perflib:"Log Flushes/sec"`
	LogFlushWaitsPerSec              float64 `perflib:"Log Flush Waits/sec"`
	LogFlushWaitTime                 float64 `perflib:"Log Flush Wait Time"`
	LogFlushWriteTimeMS              float64 `perflib:"Log Flush Write Time (ms)"`
	LogGrowths                       float64 `perflib:"Log Growths"`
	LogPoolCacheMissesPerSec         float64 `perflib:"Log Pool Cache Misses/sec"`
	LogPoolDiskReadsPerSec           float64 `perflib:"Log Pool Disk Reads/sec"`
	LogPoolHashDeletesPerSec         float64 `perflib:"Log Pool Hash Deletes/sec"`
	LogPoolHashInsertsPerSec         float64 `perflib:"Log Pool Hash Inserts/sec"`
	LogPoolInvalidHashEntryPerSec    float64 `perflib:"Log Pool Invalid Hash Entry/sec"`
	LogPoolLogScanPushesPerSec       float64 `perflib:"Log Pool Log Scan Pushes/sec"`
	LogPoolLogWriterPushesPerSec     float64 `perflib:"Log Pool LogWriter Pushes/sec"`
	LogPoolPushEmptyFreePoolPerSec   float64 `perflib:"Log Pool Push Empty FreePool/sec"`
	LogPoolPushLowMemoryPerSec       float64 `perflib:"Log Pool Push Low Memory/sec"`
	LogPoolPushNoFreeBufferPerSec    float64 `perflib:"Log Pool Push No Free Buffer/sec"`
	LogPoolReqBehindTruncPerSec      float64 `perflib:"Log Pool Req. Behind Trunc/sec"`
	LogPoolRequestsOldVLFPerSec      float64 `perflib:"Log Pool Requests Old VLF/sec"`
	LogPoolRequestsPerSec            float64 `perflib:"Log Pool Requests/sec"`
	LogPoolTotalActiveLogSize        float64 `perflib:"Log Pool Total Active Log Size"`
	LogPoolTotalSharedPoolSize       float64 `perflib:"Log Pool Total Shared Pool Size"`
	LogShrinks                       float64 `perflib:"Log Shrinks"`
	LogTruncations                   float64 `perflib:"Log Truncations"`
	PercentLogUsed                   float64 `perflib:"Percent Log Used"`
	ReplPendingXacts                 float64 `perflib:"Repl. Pending Xacts"`
	ReplTransRate                    float64 `perflib:"Repl. Trans. Rate"`
	ShrinkDataMovementBytesPerSec    float64 `perflib:"Shrink Data Movement Bytes/sec"`
	TrackedtransactionsPerSec        float64 `perflib:"Tracked transactions/sec"`
	TransactionsPerSec               float64 `perflib:"Transactions/sec"`
	WriteTransactionsPerSec          float64 `perflib:"Write Transactions/sec"`
	XTPControllerDLCLatencyPerFetch  float64 `perflib:"XTP Controller DLC Latency/Fetch"`
	XTPControllerDLCPeakLatency      float64 `perflib:"XTP Controller DLC Peak Latency"`
	XTPControllerLogProcessedPerSec  float64 `perflib:"XTP Controller Log Processed/sec"`
	XTPMemoryUsedKB                  float64 `perflib:"XTP Memory Used (KB)"`
}

func (c *Collector) collectDatabases(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric, sqlInstance string) error {
	var dst []mssqlDatabases
	_ = level.Debug(logger).Log("msg", fmt.Sprintf("mssql_databases collector iterating sql instance %s.", sqlInstance))

	if err := perflib.UnmarshalObject(ctx.PerfObjects[mssqlGetPerfObjectName(sqlInstance, "databases")], &dst, logger); err != nil {
		return err
	}

	for _, v := range dst {
		if strings.ToLower(v.Name) == "_total" {
			continue
		}
		dbName := v.Name

		ch <- prometheus.MustNewConstMetric(
			c.databasesActiveParallelredothreads,
			prometheus.GaugeValue,
			v.ActiveParallelRedoThreads,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesActiveTransactions,
			prometheus.GaugeValue,
			v.ActiveTransactions,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesBackupPerRestoreThroughput,
			prometheus.CounterValue,
			v.BackupPerRestoreThroughputPerSec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesBulkCopyRows,
			prometheus.CounterValue,
			v.BulkCopyRowsPerSec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesBulkCopyThroughput,
			prometheus.CounterValue,
			v.BulkCopyThroughputPerSec*1024,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesCommitTableEntries,
			prometheus.GaugeValue,
			v.CommitTableEntries,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesDataFilesSizeKB,
			prometheus.GaugeValue,
			v.DataFilesSizeKB*1024,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesDBCCLogicalScanBytes,
			prometheus.CounterValue,
			v.DBCCLogicalScanBytesPerSec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesGroupCommitTime,
			prometheus.CounterValue,
			v.GroupCommitTimePerSec/1000000.0,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogBytesFlushed,
			prometheus.CounterValue,
			v.LogBytesFlushedPerSec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogCacheHits,
			prometheus.GaugeValue,
			v.LogCacheHitRatio,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogCacheLookups,
			prometheus.GaugeValue,
			v.LogCacheHitRatioBase,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogCacheReads,
			prometheus.CounterValue,
			v.LogCacheReadsPerSec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogFilesSizeKB,
			prometheus.GaugeValue,
			v.LogFilesSizeKB*1024,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogFilesUsedSizeKB,
			prometheus.GaugeValue,
			v.LogFilesUsedSizeKB*1024,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogFlushes,
			prometheus.CounterValue,
			v.LogFlushesPerSec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogFlushWaits,
			prometheus.CounterValue,
			v.LogFlushWaitsPerSec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogFlushWaitTime,
			prometheus.GaugeValue,
			v.LogFlushWaitTime/1000.0,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogFlushWriteTimeMS,
			prometheus.GaugeValue,
			v.LogFlushWriteTimeMS/1000.0,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogGrowths,
			prometheus.GaugeValue,
			v.LogGrowths,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolCacheMisses,
			prometheus.CounterValue,
			v.LogPoolCacheMissesPerSec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolDiskReads,
			prometheus.CounterValue,
			v.LogPoolDiskReadsPerSec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolHashDeletes,
			prometheus.CounterValue,
			v.LogPoolHashDeletesPerSec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolHashInserts,
			prometheus.CounterValue,
			v.LogPoolHashInsertsPerSec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolInvalidHashEntry,
			prometheus.CounterValue,
			v.LogPoolInvalidHashEntryPerSec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolLogScanPushes,
			prometheus.CounterValue,
			v.LogPoolLogScanPushesPerSec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolLogWriterPushes,
			prometheus.CounterValue,
			v.LogPoolLogWriterPushesPerSec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolPushEmptyFreePool,
			prometheus.CounterValue,
			v.LogPoolPushEmptyFreePoolPerSec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolPushLowMemory,
			prometheus.CounterValue,
			v.LogPoolPushLowMemoryPerSec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolPushNoFreeBuffer,
			prometheus.CounterValue,
			v.LogPoolPushNoFreeBufferPerSec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolReqBehindTrunc,
			prometheus.CounterValue,
			v.LogPoolReqBehindTruncPerSec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolRequestsOldVLF,
			prometheus.CounterValue,
			v.LogPoolRequestsOldVLFPerSec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolRequests,
			prometheus.CounterValue,
			v.LogPoolRequestsPerSec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolTotalActiveLogSize,
			prometheus.GaugeValue,
			v.LogPoolTotalActiveLogSize,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogPoolTotalSharedPoolSize,
			prometheus.GaugeValue,
			v.LogPoolTotalSharedPoolSize,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogShrinks,
			prometheus.GaugeValue,
			v.LogShrinks,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesLogTruncations,
			prometheus.GaugeValue,
			v.LogTruncations,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesPercentLogUsed,
			prometheus.GaugeValue,
			v.PercentLogUsed,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesReplPendingXacts,
			prometheus.GaugeValue,
			v.ReplPendingXacts,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesReplTransRate,
			prometheus.CounterValue,
			v.ReplTransRate,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesShrinkDataMovementBytes,
			prometheus.CounterValue,
			v.ShrinkDataMovementBytesPerSec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesTrackedTransactions,
			prometheus.CounterValue,
			v.TrackedtransactionsPerSec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesTransactions,
			prometheus.CounterValue,
			v.TransactionsPerSec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesWriteTransactions,
			prometheus.CounterValue,
			v.WriteTransactionsPerSec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesXTPControllerDLCLatencyPerFetch,
			prometheus.GaugeValue,
			v.XTPControllerDLCLatencyPerFetch,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesXTPControllerDLCPeakLatency,
			prometheus.GaugeValue,
			v.XTPControllerDLCPeakLatency*1000000.0,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesXTPControllerLogProcessed,
			prometheus.CounterValue,
			v.XTPControllerLogProcessedPerSec,
			sqlInstance, dbName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databasesXTPMemoryUsedKB,
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
	ConnectionResetPerSec         float64 `perflib:"Connection Reset/sec"`
	EventNotificationsDelayedDrop float64 `perflib:"Event Notifications Delayed Drop"`
	HTTPAuthenticatedRequests     float64 `perflib:"HTTP Authenticated Requests"`
	LogicalConnections            float64 `perflib:"Logical Connections"`
	LoginsPerSec                  float64 `perflib:"Logins/sec"`
	LogoutsPerSec                 float64 `perflib:"Logouts/sec"`
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

func (c *Collector) collectGeneralStatistics(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric, sqlInstance string) error {
	var dst []mssqlGeneralStatistics
	_ = level.Debug(logger).Log("msg", fmt.Sprintf("mssql_genstats collector iterating sql instance %s.", sqlInstance))

	if err := perflib.UnmarshalObject(ctx.PerfObjects[mssqlGetPerfObjectName(sqlInstance, "genstats")], &dst, logger); err != nil {
		return err
	}

	for _, v := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.genStatsActiveTempTables,
			prometheus.GaugeValue,
			v.ActiveTempTables,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.genStatsConnectionReset,
			prometheus.CounterValue,
			v.ConnectionResetPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.genStatsEventNotificationsDelayedDrop,
			prometheus.GaugeValue,
			v.EventNotificationsDelayedDrop,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.genStatsHTTPAuthenticatedRequests,
			prometheus.GaugeValue,
			v.HTTPAuthenticatedRequests,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.genStatsLogicalConnections,
			prometheus.GaugeValue,
			v.LogicalConnections,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.genStatsLogins,
			prometheus.CounterValue,
			v.LoginsPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.genStatsLogouts,
			prometheus.CounterValue,
			v.LogoutsPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.genStatsMarsDeadlocks,
			prometheus.GaugeValue,
			v.MarsDeadlocks,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.genStatsNonAtomicYieldRate,
			prometheus.CounterValue,
			v.Nonatomicyieldrate,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.genStatsProcessesBlocked,
			prometheus.GaugeValue,
			v.Processesblocked,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.genStatsSOAPEmptyRequests,
			prometheus.GaugeValue,
			v.SOAPEmptyRequests,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.genStatsSOAPMethodInvocations,
			prometheus.GaugeValue,
			v.SOAPMethodInvocations,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.genStatsSOAPSessionInitiateRequests,
			prometheus.GaugeValue,
			v.SOAPSessionInitiateRequests,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.genStatsSOAPSessionTerminateRequests,
			prometheus.GaugeValue,
			v.SOAPSessionTerminateRequests,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.genStatsSOAPSQLRequests,
			prometheus.GaugeValue,
			v.SOAPSQLRequests,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.genStatsSOAPWSDLRequests,
			prometheus.GaugeValue,
			v.SOAPWSDLRequests,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.genStatsSQLTraceIOProviderLockWaits,
			prometheus.GaugeValue,
			v.SQLTraceIOProviderLockWaits,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.genStatsTempDBRecoveryUnitID,
			prometheus.GaugeValue,
			v.Tempdbrecoveryunitid,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.genStatsTempDBrowSetID,
			prometheus.GaugeValue,
			v.Tempdbrowsetid,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.genStatsTempTablesCreationRate,
			prometheus.CounterValue,
			v.TempTablesCreationRate,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.genStatsTempTablesForDestruction,
			prometheus.GaugeValue,
			v.TempTablesForDestruction,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.genStatsTraceEventNotificationQueue,
			prometheus.GaugeValue,
			v.TraceEventNotificationQueue,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.genStatsTransactions,
			prometheus.GaugeValue,
			v.Transactions,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.genStatsUserConnections,
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
	AverageWaitTimeMS          float64 `perflib:"Average Wait Time (ms)"`
	AverageWaitTimeMSBase      float64 `perflib:"Average Wait Time Base_Base"`
	LockRequestsPerSec         float64 `perflib:"Lock Requests/sec"`
	LockTimeoutsPerSec         float64 `perflib:"Lock Timeouts/sec"`
	LockTimeoutsTimeout0PerSec float64 `perflib:"Lock Timeouts (timeout > 0)/sec"`
	LockWaitsPerSec            float64 `perflib:"Lock Waits/sec"`
	LockWaitTimeMS             float64 `perflib:"Lock Wait Time (ms)"`
	NumberOfDeadlocksPerSec    float64 `perflib:"Number of Deadlocks/sec"`
}

func (c *Collector) collectLocks(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric, sqlInstance string) error {
	var dst []mssqlLocks
	_ = level.Debug(logger).Log("msg", fmt.Sprintf("mssql_locks collector iterating sql instance %s.", sqlInstance))

	if err := perflib.UnmarshalObject(ctx.PerfObjects[mssqlGetPerfObjectName(sqlInstance, "locks")], &dst, logger); err != nil {
		return err
	}

	for _, v := range dst {
		if strings.ToLower(v.Name) == "_total" {
			continue
		}
		lockResourceName := v.Name

		ch <- prometheus.MustNewConstMetric(
			c.locksWaitTime,
			prometheus.GaugeValue,
			v.AverageWaitTimeMS/1000.0,
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.locksCount,
			prometheus.GaugeValue,
			v.AverageWaitTimeMSBase/1000.0,
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.locksLockRequests,
			prometheus.CounterValue,
			v.LockRequestsPerSec,
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.locksLockTimeouts,
			prometheus.CounterValue,
			v.LockTimeoutsPerSec,
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.locksLockTimeoutstimeout0,
			prometheus.CounterValue,
			v.LockTimeoutsTimeout0PerSec,
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.locksLockWaits,
			prometheus.CounterValue,
			v.LockWaitsPerSec,
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.locksLockWaitTimeMS,
			prometheus.GaugeValue,
			v.LockWaitTimeMS/1000.0,
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.locksNumberOfDeadlocks,
			prometheus.CounterValue,
			v.NumberOfDeadlocksPerSec,
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

func (c *Collector) collectMemoryManager(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric, sqlInstance string) error {
	var dst []mssqlMemoryManager
	_ = level.Debug(logger).Log("msg", fmt.Sprintf("mssql_memmgr collector iterating sql instance %s.", sqlInstance))

	if err := perflib.UnmarshalObject(ctx.PerfObjects[mssqlGetPerfObjectName(sqlInstance, "memmgr")], &dst, logger); err != nil {
		return err
	}

	for _, v := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.memMgrConnectionMemoryKB,
			prometheus.GaugeValue,
			v.ConnectionMemoryKB*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.memMgrDatabaseCacheMemoryKB,
			prometheus.GaugeValue,
			v.DatabaseCacheMemoryKB*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.memMgrExternalbenefitofmemory,
			prometheus.GaugeValue,
			v.Externalbenefitofmemory,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.memMgrFreeMemoryKB,
			prometheus.GaugeValue,
			v.FreeMemoryKB*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.memMgrGrantedWorkspaceMemoryKB,
			prometheus.GaugeValue,
			v.GrantedWorkspaceMemoryKB*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.memMgrLockBlocks,
			prometheus.GaugeValue,
			v.LockBlocks,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.memMgrLockBlocksAllocated,
			prometheus.GaugeValue,
			v.LockBlocksAllocated,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.memMgrLockMemoryKB,
			prometheus.GaugeValue,
			v.LockMemoryKB*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.memMgrLockOwnerBlocks,
			prometheus.GaugeValue,
			v.LockOwnerBlocks,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.memMgrLockOwnerBlocksAllocated,
			prometheus.GaugeValue,
			v.LockOwnerBlocksAllocated,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.memMgrLogPoolMemoryKB,
			prometheus.GaugeValue,
			v.LogPoolMemoryKB*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.memMgrMaximumWorkspaceMemoryKB,
			prometheus.GaugeValue,
			v.MaximumWorkspaceMemoryKB*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.memMgrMemoryGrantsOutstanding,
			prometheus.GaugeValue,
			v.MemoryGrantsOutstanding,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.memMgrMemoryGrantsPending,
			prometheus.GaugeValue,
			v.MemoryGrantsPending,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.memMgrOptimizerMemoryKB,
			prometheus.GaugeValue,
			v.OptimizerMemoryKB*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.memMgrReservedServerMemoryKB,
			prometheus.GaugeValue,
			v.ReservedServerMemoryKB*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.memMgrSQLCacheMemoryKB,
			prometheus.GaugeValue,
			v.SQLCacheMemoryKB*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.memMgrStolenServerMemoryKB,
			prometheus.GaugeValue,
			v.StolenServerMemoryKB*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.memMgrTargetServerMemoryKB,
			prometheus.GaugeValue,
			v.TargetServerMemoryKB*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.memMgrTotalServerMemoryKB,
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
	AutoParamAttemptsPerSec       float64 `perflib:"Auto-Param Attempts/sec"`
	BatchRequestsPerSec           float64 `perflib:"Batch Requests/sec"`
	FailedAutoParamsPerSec        float64 `perflib:"Failed Auto-Params/sec"`
	ForcedParameterizationsPerSec float64 `perflib:"Forced Parameterizations/sec"`
	GuidedplanexecutionsPerSec    float64 `perflib:"Guided plan executions/sec"`
	MisguidedplanexecutionsPerSec float64 `perflib:"Misguided plan executions/sec"`
	SafeAutoParamsPerSec          float64 `perflib:"Safe Auto-Params/sec"`
	SQLAttentionrate              float64 `perflib:"SQL Attention rate"`
	SQLCompilationsPerSec         float64 `perflib:"SQL Compilations/sec"`
	SQLReCompilationsPerSec       float64 `perflib:"SQL Re-Compilations/sec"`
	UnsafeAutoParamsPerSec        float64 `perflib:"Unsafe Auto-Params/sec"`
}

func (c *Collector) collectSQLStats(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric, sqlInstance string) error {
	var dst []mssqlSQLStatistics
	_ = level.Debug(logger).Log("msg", fmt.Sprintf("mssql_sqlstats collector iterating sql instance %s.", sqlInstance))

	if err := perflib.UnmarshalObject(ctx.PerfObjects[mssqlGetPerfObjectName(sqlInstance, "sqlstats")], &dst, logger); err != nil {
		return err
	}

	for _, v := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.sqlStatsAutoParamAttempts,
			prometheus.CounterValue,
			v.AutoParamAttemptsPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.sqlStatsBatchRequests,
			prometheus.CounterValue,
			v.BatchRequestsPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.sqlStatsFailedAutoParams,
			prometheus.CounterValue,
			v.FailedAutoParamsPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.sqlStatsForcedParameterizations,
			prometheus.CounterValue,
			v.ForcedParameterizationsPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.sqlStatsGuidedplanexecutions,
			prometheus.CounterValue,
			v.GuidedplanexecutionsPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.sqlStatsMisguidedplanexecutions,
			prometheus.CounterValue,
			v.MisguidedplanexecutionsPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.sqlStatsSafeAutoParams,
			prometheus.CounterValue,
			v.SafeAutoParamsPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.sqlStatsSQLAttentionrate,
			prometheus.CounterValue,
			v.SQLAttentionrate,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.sqlStatsSQLCompilations,
			prometheus.CounterValue,
			v.SQLCompilationsPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.sqlStatsSQLReCompilations,
			prometheus.CounterValue,
			v.SQLReCompilationsPerSec,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.sqlStatsUnsafeAutoParams,
			prometheus.CounterValue,
			v.UnsafeAutoParamsPerSec,
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

func (c *Collector) collectWaitStats(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric, sqlInstance string) error {
	var dst []mssqlWaitStatistics
	_ = level.Debug(logger).Log("msg", fmt.Sprintf("mssql_waitstats collector iterating sql instance %s.", sqlInstance))

	if err := perflib.UnmarshalObject(ctx.PerfObjects[mssqlGetPerfObjectName(sqlInstance, "waitstats")], &dst, logger); err != nil {
		return err
	}

	for _, v := range dst {
		item := v.Name

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsLockWaits,
			prometheus.CounterValue,
			v.WaitStatsLockWaits,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsMemoryGrantQueueWaits,
			prometheus.CounterValue,
			v.WaitStatsMemoryGrantQueueWaits,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsThreadSafeMemoryObjectsWaits,
			prometheus.CounterValue,
			v.WaitStatsThreadSafeMemoryObjectsWaits,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsLogWriteWaits,
			prometheus.CounterValue,
			v.WaitStatsLogWriteWaits,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsLogBufferWaits,
			prometheus.CounterValue,
			v.WaitStatsLogBufferWaits,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsNetworkIOWaits,
			prometheus.CounterValue,
			v.WaitStatsNetworkIOWaits,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsPageIOLatchWaits,
			prometheus.CounterValue,
			v.WaitStatsPageIOLatchWaits,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsPageLatchWaits,
			prometheus.CounterValue,
			v.WaitStatsPageLatchWaits,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsNonPageLatchWaits,
			prometheus.CounterValue,
			v.WaitStatsNonpageLatchWaits,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsWaitForTheWorkerWaits,
			prometheus.CounterValue,
			v.WaitStatsWaitForTheWorkerWaits,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsWorkspaceSynchronizationWaits,
			prometheus.CounterValue,
			v.WaitStatsWorkspaceSynchronizationWaits,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsTransactionOwnershipWaits,
			prometheus.CounterValue,
			v.WaitStatsTransactionOwnershipWaits,
			sqlInstance, item,
		)
	}

	return nil
}

type mssqlSQLErrors struct {
	Name         string
	ErrorsPerSec float64 `perflib:"Errors/sec"`
}

// Win32_PerfRawData_MSSQLSERVER_SQLServerErrors docs:
// - https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-sql-errors-object
func (c *Collector) collectSQLErrors(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric, sqlInstance string) error {
	var dst []mssqlSQLErrors
	_ = level.Debug(logger).Log("msg", fmt.Sprintf("mssql_sqlerrors collector iterating sql instance %s.", sqlInstance))

	if err := perflib.UnmarshalObject(ctx.PerfObjects[mssqlGetPerfObjectName(sqlInstance, "sqlerrors")], &dst, logger); err != nil {
		return err
	}

	for _, v := range dst {
		if strings.ToLower(v.Name) == "_total" {
			continue
		}
		resource := v.Name

		ch <- prometheus.MustNewConstMetric(
			c.sqlErrorsTotal,
			prometheus.CounterValue,
			v.ErrorsPerSec,
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
func (c *Collector) collectTransactions(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric, sqlInstance string) error {
	var dst []mssqlTransactions
	_ = level.Debug(logger).Log("msg", fmt.Sprintf("mssql_transactions collector iterating sql instance %s.", sqlInstance))

	if err := perflib.UnmarshalObject(ctx.PerfObjects[mssqlGetPerfObjectName(sqlInstance, "transactions")], &dst, logger); err != nil {
		return err
	}

	for _, v := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.transactionsTempDbFreeSpaceBytes,
			prometheus.GaugeValue,
			v.FreeSpaceintempdbKB*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.transactionsLongestTransactionRunningSeconds,
			prometheus.GaugeValue,
			v.LongestTransactionRunningTime,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.transactionsNonSnapshotVersionActiveTotal,
			prometheus.CounterValue,
			v.NonSnapshotVersionTransactions,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.transactionsSnapshotActiveTotal,
			prometheus.CounterValue,
			v.SnapshotTransactions,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.transactionsActive,
			prometheus.GaugeValue,
			v.Transactions,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.transactionsUpdateConflictsTotal,
			prometheus.CounterValue,
			v.Updateconflictratio,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.transactionsUpdateSnapshotActiveTotal,
			prometheus.CounterValue,
			v.UpdateSnapshotTransactions,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.transactionsVersionCleanupRateBytes,
			prometheus.GaugeValue,
			v.VersionCleanuprateKBPers*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.transactionsVersionGenerationRateBytes,
			prometheus.GaugeValue,
			v.VersionGenerationrateKBPers*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.transactionsVersionStoreSizeBytes,
			prometheus.GaugeValue,
			v.VersionStoreSizeKB*1024,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.transactionsVersionStoreUnits,
			prometheus.CounterValue,
			v.VersionStoreunitcount,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.transactionsVersionStoreCreationUnits,
			prometheus.CounterValue,
			v.VersionStoreunitcreation,
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.transactionsVersionStoreTruncationUnits,
			prometheus.CounterValue,
			v.VersionStoreunittruncation,
			sqlInstance,
		)
	}

	return nil
}

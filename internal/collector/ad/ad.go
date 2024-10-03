//go:build windows

package ad

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	types2 "github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
)

const Name = "ad"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_DirectoryServices_DirectoryServices metrics.
type Collector struct {
	config    Config
	wmiClient *wmi.Client

	perfDataCollector *perfdata.Collector

	addressBookClientSessions                           *prometheus.Desc
	addressBookOperationsTotal                          *prometheus.Desc
	approximateHighestDistinguishedNameTag              *prometheus.Desc
	atqAverageRequestLatency                            *prometheus.Desc
	atqCurrentThreads                                   *prometheus.Desc
	atqEstimatedDelaySeconds                            *prometheus.Desc
	atqOutstandingRequests                              *prometheus.Desc
	bindsTotal                                          *prometheus.Desc
	changeMonitorUpdatesPending                         *prometheus.Desc
	changeMonitorsRegistered                            *prometheus.Desc
	databaseOperationsTotal                             *prometheus.Desc
	directoryOperationsTotal                            *prometheus.Desc
	directorySearchSubOperationsTotal                   *prometheus.Desc
	directoryServiceThreads                             *prometheus.Desc
	interSiteReplicationDataBytesTotal                  *prometheus.Desc
	intraSiteReplicationDataBytesTotal                  *prometheus.Desc
	ldapActiveThreads                                   *prometheus.Desc
	ldapClientSessions                                  *prometheus.Desc
	ldapClosedConnectionsTotal                          *prometheus.Desc
	ldapLastBindTimeSeconds                             *prometheus.Desc
	ldapOpenedConnectionsTotal                          *prometheus.Desc
	ldapSearchesTotal                                   *prometheus.Desc
	ldapUdpOperationsTotal                              *prometheus.Desc
	ldapWritesTotal                                     *prometheus.Desc
	linkValuesCleanedTotal                              *prometheus.Desc
	nameCacheHitsTotal                                  *prometheus.Desc
	nameCacheLookupsTotal                               *prometheus.Desc
	nameTranslationsTotal                               *prometheus.Desc
	phantomObjectsCleanedTotal                          *prometheus.Desc
	phantomObjectsVisitedTotal                          *prometheus.Desc
	replicationHighestUsn                               *prometheus.Desc
	replicationInboundLinkValueUpdatesRemaining         *prometheus.Desc
	replicationInboundObjectsFilteredTotal              *prometheus.Desc
	replicationInboundObjectsUpdatedTotal               *prometheus.Desc
	replicationInboundPropertiesFilteredTotal           *prometheus.Desc
	replicationInboundPropertiesUpdatedTotal            *prometheus.Desc
	replicationInboundSyncObjectsRemaining              *prometheus.Desc
	replicationPendingOperations                        *prometheus.Desc
	replicationPendingSynchronizations                  *prometheus.Desc
	replicationSyncRequestsSchemaMismatchFailureTotal   *prometheus.Desc
	replicationSyncRequestsSuccessTotal                 *prometheus.Desc
	replicationSyncRequestsTotal                        *prometheus.Desc
	samComputerCreationRequestsTotal                    *prometheus.Desc
	samComputerCreationSuccessfulRequestsTotal          *prometheus.Desc
	samEnumerationsTotal                                *prometheus.Desc
	samGroupEvaluationLatency                           *prometheus.Desc
	samGroupMembershipEvaluationsNonTransitiveTotal     *prometheus.Desc
	samGroupMembershipEvaluationsTotal                  *prometheus.Desc
	samGroupMembershipEvaluationsTransitiveTotal        *prometheus.Desc
	samGroupMembershipGlobalCatalogEvaluationsTotal     *prometheus.Desc
	samMembershipChangesTotal                           *prometheus.Desc
	samPasswordChangesTotal                             *prometheus.Desc
	samQueryDisplayRequestsTotal                        *prometheus.Desc
	samUserCreationRequestsTotal                        *prometheus.Desc
	samUserCreationSuccessfulRequestsTotal              *prometheus.Desc
	searchesTotal                                       *prometheus.Desc
	securityDescriptorPropagationAccessWaitTotalSeconds *prometheus.Desc
	securityDescriptorPropagationEventsQueued           *prometheus.Desc
	securityDescriptorPropagationEventsTotal            *prometheus.Desc
	securityDescriptorPropagationItemsQueuedTotal       *prometheus.Desc
	tombstonesObjectsCollectedTotal                     *prometheus.Desc
	tombstonesObjectsVisitedTotal                       *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) GetPerfCounter(_ *slog.Logger) ([]string, error) {
	return []string{}, nil
}

func (c *Collector) Close(_ *slog.Logger) error {
	return nil
}

func (c *Collector) Build(_ *slog.Logger, wmiClient *wmi.Client) error {
	if wmiClient == nil || wmiClient.SWbemServicesClient == nil {
		return errors.New("wmiClient or SWbemServicesClient is nil")
	}

	if utils.PDHEnabled() {
		counters := []string{
			abANRPerSec,
			abBrowsesPerSec,
			abClientSessions,
			abMatchesPerSec,
			abPropertyReadsPerSec,
			abProxyLookupsPerSec,
			abSearchesPerSec,
			approximateHighestDNT,
			atqEstimatedQueueDelay,
			atqOutstandingQueuedRequests,
			atqRequestLatency,
			atqThreadsLDAP,
			atqThreadsOther,
			atqThreadsTotal,
			baseSearchesPerSec,
			databaseAddsPerSec,
			databaseDeletesPerSec,
			databaseModifiesPerSec,
			databaseRecyclesPerSec,
			digestBindsPerSec,
			draHighestUSNCommittedHighPart,
			draHighestUSNCommittedLowPart,
			draHighestUSNIssuedHighPart,
			draHighestUSNIssuedLowPart,
			draInboundBytesCompressedBetweenSitesAfterCompressionSinceBoot,
			draInboundBytesCompressedBetweenSitesAfterCompressionPerSec,
			draInboundBytesCompressedBetweenSitesBeforeCompressionSinceBoot,
			draInboundBytesCompressedBetweenSitesBeforeCompressionPerSec,
			draInboundBytesNotCompressedWithinSiteSinceBoot,
			draInboundBytesNotCompressedWithinSitePerSec,
			draInboundBytesTotalSinceBoot,
			draInboundBytesTotalPerSec,
			draInboundFullSyncObjectsRemaining,
			draInboundLinkValueUpdatesRemainingInPacket,
			draInboundObjectUpdatesRemainingInPacket,
			draInboundObjectsAppliedPerSec,
			draInboundObjectsFilteredPerSec,
			draInboundObjectsPerSec,
			draInboundPropertiesAppliedPerSec,
			draInboundPropertiesFilteredPerSec,
			draInboundPropertiesTotalPerSec,
			draInboundTotalUpdatesRemainingInPacket,
			draInboundValuesDNsOnlyPerSec,
			draInboundValuesTotalPerSec,
			draOutboundBytesCompressedBetweenSitesAfterCompressionSinceBoot,
			draOutboundBytesCompressedBetweenSitesAfterCompressionPerSec,
			draOutboundBytesCompressedBetweenSitesBeforeCompressionSinceBoot,
			draOutboundBytesCompressedBetweenSitesBeforeCompressionPerSec,
			draOutboundBytesNotCompressedWithinSiteSinceBoot,
			draOutboundBytesNotCompressedWithinSitePerSec,
			draOutboundBytesTotalSinceBoot,
			draOutboundBytesTotalPerSec,
			draOutboundObjectsFilteredPerSec,
			draOutboundObjectsPerSec,
			draOutboundPropertiesPerSec,
			draOutboundValuesDNsOnlyPerSec,
			draOutboundValuesTotalPerSec,
			draPendingReplicationOperations,
			draPendingReplicationSynchronizations,
			draSyncFailuresOnSchemaMismatch,
			draSyncRequestsMade,
			draSyncRequestsSuccessful,
			draThreadsGettingNCChanges,
			draThreadsGettingNCChangesHoldingSemaphore,
			dsPercentReadsFromDRA,
			dsPercentReadsFromKCC,
			dsPercentReadsFromLSA,
			dsPercentReadsFromNSPI,
			dsPercentReadsFromNTDSAPI,
			dsPercentReadsFromSAM,
			dsPercentReadsOther,
			dsPercentSearchesFromDRA,
			dsPercentSearchesFromKCC,
			dsPercentSearchesFromLDAP,
			dsPercentSearchesFromLSA,
			dsPercentSearchesFromNSPI,
			dsPercentSearchesFromNTDSAPI,
			dsPercentSearchesFromSAM,
			dsPercentSearchesOther,
			dsPercentWritesFromDRA,
			dsPercentWritesFromKCC,
			dsPercentWritesFromLDAP,
			dsPercentWritesFromLSA,
			dsPercentWritesFromNSPI,
			dsPercentWritesFromNTDSAPI,
			dsPercentWritesFromSAM,
			dsPercentWritesOther,
			dsClientBindsPerSec,
			dsClientNameTranslationsPerSec,
			dsDirectoryReadsPerSec,
			dsDirectorySearchesPerSec,
			dsDirectoryWritesPerSec,
			dsMonitorListSize,
			dsNameCacheHitRate,
			dsNotifyQueueSize,
			dsSearchSubOperationsPerSec,
			dsSecurityDescriptorPropagationsEvents,
			dsSecurityDescriptorPropagatorAverageExclusionTime,
			dsSecurityDescriptorPropagatorRuntimeQueue,
			dsSecurityDescriptorSubOperationsPerSec,
			dsServerBindsPerSec,
			dsServerNameTranslationsPerSec,
			dsThreadsInUse,
			externalBindsPerSec,
			fastBindsPerSec,
			ldapActiveThreads,
			ldapBindTime,
			ldapClientSessions,
			ldapClosedConnectionsPerSec,
			ldapNewConnectionsPerSec,
			ldapNewSSLConnectionsPerSec,
			ldapSearchesPerSec,
			ldapSuccessfulBindsPerSec,
			ldapUDPOperationsPerSec,
			ldapWritesPerSec,
			linkValuesCleanedPerSec,
			negotiatedBindsPerSec,
			ntlmBindsPerSec,
			oneLevelSearchesPerSec,
			phantomsCleanedPerSec,
			phantomsVisitedPerSec,
			samAccountGroupEvaluationLatency,
			samDisplayInformationQueriesPerSec,
			samDomainLocalGroupMembershipEvaluationsPerSec,
			samEnumerationsPerSec,
			samGCEvaluationsPerSec,
			samGlobalGroupMembershipEvaluationsPerSec,
			samMachineCreationAttemptsPerSec,
			samMembershipChangesPerSec,
			samNonTransitiveMembershipEvaluationsPerSec,
			samPasswordChangesPerSec,
			samResourceGroupEvaluationLatency,
			samSuccessfulComputerCreationsPerSecIncludesAllRequests,
			samSuccessfulUserCreationsPerSec,
			samTransitiveMembershipEvaluationsPerSec,
			samUniversalGroupMembershipEvaluationsPerSec,
			samUserCreationAttemptsPerSec,
			simpleBindsPerSec,
			subtreeSearchesPerSec,
			tombstonesGarbageCollectedPerSec,
			tombstonesVisitedPerSec,
			transitiveOperationsMillisecondsRun,
			transitiveOperationsPerSec,
			transitiveSubOperationsPerSec,
		}

		var err error

		c.perfDataCollector, err = perfdata.NewCollector("DirectoryServices", []string{"*"}, counters)
		if err != nil {
			return fmt.Errorf("failed to create DirectoryServices collector: %w", err)
		}
	}

	c.wmiClient = wmiClient

	c.addressBookOperationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "address_book_operations_total"),
		"",
		[]string{"operation"},
		nil,
	)
	c.addressBookClientSessions = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "address_book_client_sessions"),
		"",
		nil,
		nil,
	)
	c.approximateHighestDistinguishedNameTag = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "approximate_highest_distinguished_name_tag"),
		"",
		nil,
		nil,
	)
	c.atqEstimatedDelaySeconds = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "atq_estimated_delay_seconds"),
		"",
		nil,
		nil,
	)
	c.atqOutstandingRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "atq_outstanding_requests"),
		"",
		nil,
		nil,
	)
	c.atqAverageRequestLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "atq_average_request_latency"),
		"",
		nil,
		nil,
	)
	c.atqCurrentThreads = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "atq_current_threads"),
		"",
		[]string{"service"},
		nil,
	)
	c.searchesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "searches_total"),
		"",
		[]string{"scope"},
		nil,
	)
	c.databaseOperationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "database_operations_total"),
		"",
		[]string{"operation"},
		nil,
	)
	c.bindsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "binds_total"),
		"",
		[]string{"bind_method"},
		nil,
	)
	c.replicationHighestUsn = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "replication_highest_usn"),
		"",
		[]string{"state"},
		nil,
	)
	c.intraSiteReplicationDataBytesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "replication_data_intrasite_bytes_total"),
		"",
		[]string{"direction"},
		nil,
	)
	c.interSiteReplicationDataBytesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "replication_data_intersite_bytes_total"),
		"",
		[]string{"direction"},
		nil,
	)
	c.replicationInboundSyncObjectsRemaining = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "replication_inbound_sync_objects_remaining"),
		"",
		nil,
		nil,
	)
	c.replicationInboundLinkValueUpdatesRemaining = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "replication_inbound_link_value_updates_remaining"),
		"",
		nil,
		nil,
	)
	c.replicationInboundObjectsUpdatedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "replication_inbound_objects_updated_total"),
		"",
		nil,
		nil,
	)
	c.replicationInboundObjectsFilteredTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "replication_inbound_objects_filtered_total"),
		"",
		nil,
		nil,
	)
	c.replicationInboundPropertiesUpdatedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "replication_inbound_properties_updated_total"),
		"",
		nil,
		nil,
	)
	c.replicationInboundPropertiesFilteredTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "replication_inbound_properties_filtered_total"),
		"",
		nil,
		nil,
	)
	c.replicationPendingOperations = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "replication_pending_operations"),
		"",
		nil,
		nil,
	)
	c.replicationPendingSynchronizations = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "replication_pending_synchronizations"),
		"",
		nil,
		nil,
	)
	c.replicationSyncRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "replication_sync_requests_total"),
		"",
		nil,
		nil,
	)
	c.replicationSyncRequestsSuccessTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "replication_sync_requests_success_total"),
		"",
		nil,
		nil,
	)
	c.replicationSyncRequestsSchemaMismatchFailureTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "replication_sync_requests_schema_mismatch_failure_total"),
		"",
		nil,
		nil,
	)
	c.nameTranslationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "name_translations_total"),
		"",
		[]string{"target_name"},
		nil,
	)
	c.changeMonitorsRegistered = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "change_monitors_registered"),
		"",
		nil,
		nil,
	)
	c.changeMonitorUpdatesPending = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "change_monitor_updates_pending"),
		"",
		nil,
		nil,
	)
	c.nameCacheHitsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "name_cache_hits_total"),
		"",
		nil,
		nil,
	)
	c.nameCacheLookupsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "name_cache_lookups_total"),
		"",
		nil,
		nil,
	)
	c.directoryOperationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "directory_operations_total"),
		"",
		[]string{"operation", "origin"},
		nil,
	)
	c.directorySearchSubOperationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "directory_search_suboperations_total"),
		"",
		nil,
		nil,
	)
	c.securityDescriptorPropagationEventsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "security_descriptor_propagation_events_total"),
		"",
		nil,
		nil,
	)
	c.securityDescriptorPropagationEventsQueued = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "security_descriptor_propagation_events_queued"),
		"",
		nil,
		nil,
	)
	c.securityDescriptorPropagationAccessWaitTotalSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "security_descriptor_propagation_access_wait_total_seconds"),
		"",
		nil,
		nil,
	)
	c.securityDescriptorPropagationItemsQueuedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "security_descriptor_propagation_items_queued_total"),
		"",
		nil,
		nil,
	)
	c.directoryServiceThreads = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "directory_service_threads"),
		"",
		nil,
		nil,
	)
	c.ldapClosedConnectionsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "ldap_closed_connections_total"),
		"",
		nil,
		nil,
	)
	c.ldapOpenedConnectionsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "ldap_opened_connections_total"),
		"",
		[]string{"type"},
		nil,
	)
	c.ldapActiveThreads = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "ldap_active_threads"),
		"",
		nil,
		nil,
	)
	c.ldapLastBindTimeSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "ldap_last_bind_time_seconds"),
		"",
		nil,
		nil,
	)
	c.ldapSearchesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "ldap_searches_total"),
		"",
		nil,
		nil,
	)
	c.ldapUdpOperationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "ldap_udp_operations_total"),
		"",
		nil,
		nil,
	)
	c.ldapWritesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "ldap_writes_total"),
		"",
		nil,
		nil,
	)
	c.ldapClientSessions = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "ldap_client_sessions"),
		"This is the number of sessions opened by LDAP clients at the time the data is taken. This is helpful in determining LDAP client activity and if the DC is able to handle the load. Of course, spikes during normal periods of authentication — such as first thing in the morning — are not necessarily a problem, but long sustained periods of high values indicate an overworked DC.",
		nil,
		nil,
	)
	c.linkValuesCleanedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "link_values_cleaned_total"),
		"",
		nil,
		nil,
	)
	c.phantomObjectsCleanedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "phantom_objects_cleaned_total"),
		"",
		nil,
		nil,
	)
	c.phantomObjectsVisitedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "phantom_objects_visited_total"),
		"",
		nil,
		nil,
	)
	c.samGroupMembershipEvaluationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "sam_group_membership_evaluations_total"),
		"",
		[]string{"group_type"},
		nil,
	)
	c.samGroupMembershipGlobalCatalogEvaluationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "sam_group_membership_global_catalog_evaluations_total"),
		"",
		nil,
		nil,
	)
	c.samGroupMembershipEvaluationsNonTransitiveTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "sam_group_membership_evaluations_nontransitive_total"),
		"",
		nil,
		nil,
	)
	c.samGroupMembershipEvaluationsTransitiveTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "sam_group_membership_evaluations_transitive_total"),
		"",
		nil,
		nil,
	)
	c.samGroupEvaluationLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "sam_group_evaluation_latency"),
		"The mean latency of the last 100 group evaluations performed for authentication",
		[]string{"evaluation_type"},
		nil,
	)
	c.samComputerCreationRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "sam_computer_creation_requests_total"),
		"",
		nil,
		nil,
	)
	c.samComputerCreationSuccessfulRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "sam_computer_creation_successful_requests_total"),
		"",
		nil,
		nil,
	)
	c.samUserCreationRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "sam_user_creation_requests_total"),
		"",
		nil,
		nil,
	)
	c.samUserCreationSuccessfulRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "sam_user_creation_successful_requests_total"),
		"",
		nil,
		nil,
	)
	c.samQueryDisplayRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "sam_query_display_requests_total"),
		"",
		nil,
		nil,
	)
	c.samEnumerationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "sam_enumerations_total"),
		"",
		nil,
		nil,
	)
	c.samMembershipChangesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "sam_membership_changes_total"),
		"",
		nil,
		nil,
	)
	c.samPasswordChangesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "sam_password_changes_total"),
		"",
		nil,
		nil,
	)

	c.tombstonesObjectsCollectedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "tombstoned_objects_collected_total"),
		"",
		nil,
		nil,
	)
	c.tombstonesObjectsVisitedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types2.Namespace, Name, "tombstoned_objects_visited_total"),
		"",
		nil,
		nil,
	)

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types2.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	if utils.PDHEnabled() {
		return c.collectPDH(ch)
	}

	logger = logger.With(slog.String("collector", Name))
	if err := c.collect(ch); err != nil {
		logger.Error("failed collecting ad metrics",
			slog.Any("err", err),
		)

		return err
	}

	return nil
}

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_DirectoryServices_DirectoryServices
	if err := c.wmiClient.Query("SELECT * FROM Win32_PerfRawData_DirectoryServices_DirectoryServices", &dst); err != nil {
		return err
	}

	if len(dst) == 0 {
		return errors.New("WMI query returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.addressBookOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].ABANRPersec),
		"ambiguous_name_resolution",
	)
	ch <- prometheus.MustNewConstMetric(
		c.addressBookOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].ABBrowsesPersec),
		"browse",
	)
	ch <- prometheus.MustNewConstMetric(
		c.addressBookOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].ABMatchesPersec),
		"find",
	)
	ch <- prometheus.MustNewConstMetric(
		c.addressBookOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].ABPropertyReadsPersec),
		"property_read",
	)
	ch <- prometheus.MustNewConstMetric(
		c.addressBookOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].ABSearchesPersec),
		"search",
	)
	ch <- prometheus.MustNewConstMetric(
		c.addressBookOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].ABProxyLookupsPersec),
		"proxy_search",
	)

	ch <- prometheus.MustNewConstMetric(
		c.addressBookClientSessions,
		prometheus.GaugeValue,
		float64(dst[0].ABClientSessions),
	)

	ch <- prometheus.MustNewConstMetric(
		c.approximateHighestDistinguishedNameTag,
		prometheus.GaugeValue,
		float64(dst[0].ApproximatehighestDNT),
	)

	ch <- prometheus.MustNewConstMetric(
		c.atqEstimatedDelaySeconds,
		prometheus.GaugeValue,
		float64(dst[0].ATQEstimatedQueueDelay)/1000,
	)
	ch <- prometheus.MustNewConstMetric(
		c.atqOutstandingRequests,
		prometheus.GaugeValue,
		float64(dst[0].ATQOutstandingQueuedRequests),
	)
	ch <- prometheus.MustNewConstMetric(
		c.atqAverageRequestLatency,
		prometheus.GaugeValue,
		float64(dst[0].ATQRequestLatency),
	)
	ch <- prometheus.MustNewConstMetric(
		c.atqCurrentThreads,
		prometheus.GaugeValue,
		float64(dst[0].ATQThreadsLDAP),
		"ldap",
	)
	ch <- prometheus.MustNewConstMetric(
		c.atqCurrentThreads,
		prometheus.GaugeValue,
		float64(dst[0].ATQThreadsOther),
		"other",
	)

	ch <- prometheus.MustNewConstMetric(
		c.searchesTotal,
		prometheus.CounterValue,
		float64(dst[0].BasesearchesPersec),
		"base",
	)
	ch <- prometheus.MustNewConstMetric(
		c.searchesTotal,
		prometheus.CounterValue,
		float64(dst[0].SubtreesearchesPersec),
		"subtree",
	)
	ch <- prometheus.MustNewConstMetric(
		c.searchesTotal,
		prometheus.CounterValue,
		float64(dst[0].OnelevelsearchesPersec),
		"one_level",
	)

	ch <- prometheus.MustNewConstMetric(
		c.databaseOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DatabaseaddsPersec),
		"add",
	)
	ch <- prometheus.MustNewConstMetric(
		c.databaseOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DatabasedeletesPersec),
		"delete",
	)
	ch <- prometheus.MustNewConstMetric(
		c.databaseOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DatabasemodifysPersec),
		"modify",
	)
	ch <- prometheus.MustNewConstMetric(
		c.databaseOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DatabaserecyclesPersec),
		"recycle",
	)

	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		float64(dst[0].DigestBindsPersec),
		"digest",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSClientBindsPersec),
		"ds_client",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSServerBindsPersec),
		"ds_server",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		float64(dst[0].ExternalBindsPersec),
		"external",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		float64(dst[0].FastBindsPersec),
		"fast",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		float64(dst[0].NegotiatedBindsPersec),
		"negotiate",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		float64(dst[0].NTLMBindsPersec),
		"ntlm",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		float64(dst[0].SimpleBindsPersec),
		"simple",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		float64(dst[0].LDAPSuccessfulBindsPersec),
		"ldap",
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationHighestUsn,
		prometheus.CounterValue,
		float64(dst[0].DRAHighestUSNCommittedHighpart<<32)+float64(dst[0].DRAHighestUSNCommittedLowpart),
		"committed",
	)
	ch <- prometheus.MustNewConstMetric(
		c.replicationHighestUsn,
		prometheus.CounterValue,
		float64(dst[0].DRAHighestUSNIssuedHighpart<<32)+float64(dst[0].DRAHighestUSNIssuedLowpart),
		"issued",
	)

	ch <- prometheus.MustNewConstMetric(
		c.interSiteReplicationDataBytesTotal,
		prometheus.CounterValue,
		float64(dst[0].DRAInboundBytesCompressedBetweenSitesAfterCompressionPersec),
		"inbound",
	)
	// The pre-compression data size seems to have little value? Skipping for now
	// ch <- prometheus.MustNewConstMetric(
	// 	c.interSiteReplicationDataBytesTotal,
	// 	prometheus.CounterValue,
	// 	float64(dst[0].DRAInboundBytesCompressedBetweenSitesBeforeCompressionPersec),
	// 	"inbound",
	// )
	ch <- prometheus.MustNewConstMetric(
		c.interSiteReplicationDataBytesTotal,
		prometheus.CounterValue,
		float64(dst[0].DRAOutboundBytesCompressedBetweenSitesAfterCompressionPersec),
		"outbound",
	)
	// ch <- prometheus.MustNewConstMetric(
	// 	c.interSiteReplicationDataBytesTotal,
	// 	prometheus.CounterValue,
	// 	float64(dst[0].DRAOutboundBytesCompressedBetweenSitesBeforeCompressionPersec),
	// 	"outbound",
	// )
	ch <- prometheus.MustNewConstMetric(
		c.intraSiteReplicationDataBytesTotal,
		prometheus.CounterValue,
		float64(dst[0].DRAInboundBytesNotCompressedWithinSitePersec),
		"inbound",
	)
	ch <- prometheus.MustNewConstMetric(
		c.intraSiteReplicationDataBytesTotal,
		prometheus.CounterValue,
		float64(dst[0].DRAOutboundBytesNotCompressedWithinSitePersec),
		"outbound",
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationInboundSyncObjectsRemaining,
		prometheus.GaugeValue,
		float64(dst[0].DRAInboundFullSyncObjectsRemaining),
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationInboundLinkValueUpdatesRemaining,
		prometheus.GaugeValue,
		float64(dst[0].DRAInboundLinkValueUpdatesRemaininginPacket),
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationInboundObjectsUpdatedTotal,
		prometheus.CounterValue,
		float64(dst[0].DRAInboundObjectsAppliedPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.replicationInboundObjectsFilteredTotal,
		prometheus.CounterValue,
		float64(dst[0].DRAInboundObjectsFilteredPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationInboundPropertiesUpdatedTotal,
		prometheus.CounterValue,
		float64(dst[0].DRAInboundPropertiesAppliedPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.replicationInboundPropertiesFilteredTotal,
		prometheus.CounterValue,
		float64(dst[0].DRAInboundPropertiesFilteredPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationPendingOperations,
		prometheus.GaugeValue,
		float64(dst[0].DRAPendingReplicationOperations),
	)
	ch <- prometheus.MustNewConstMetric(
		c.replicationPendingSynchronizations,
		prometheus.GaugeValue,
		float64(dst[0].DRAPendingReplicationSynchronizations),
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationSyncRequestsTotal,
		prometheus.CounterValue,
		float64(dst[0].DRASyncRequestsMade),
	)
	ch <- prometheus.MustNewConstMetric(
		c.replicationSyncRequestsSuccessTotal,
		prometheus.CounterValue,
		float64(dst[0].DRASyncRequestsSuccessful),
	)
	ch <- prometheus.MustNewConstMetric(
		c.replicationSyncRequestsSchemaMismatchFailureTotal,
		prometheus.CounterValue,
		float64(dst[0].DRASyncFailuresonSchemaMismatch),
	)

	ch <- prometheus.MustNewConstMetric(
		c.nameTranslationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSClientNameTranslationsPersec),
		"client",
	)
	ch <- prometheus.MustNewConstMetric(
		c.nameTranslationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSServerNameTranslationsPersec),
		"server",
	)

	ch <- prometheus.MustNewConstMetric(
		c.changeMonitorsRegistered,
		prometheus.GaugeValue,
		float64(dst[0].DSMonitorListSize),
	)
	ch <- prometheus.MustNewConstMetric(
		c.changeMonitorUpdatesPending,
		prometheus.GaugeValue,
		float64(dst[0].DSNotifyQueueSize),
	)

	ch <- prometheus.MustNewConstMetric(
		c.nameCacheHitsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSNameCachehitrate),
	)
	ch <- prometheus.MustNewConstMetric(
		c.nameCacheLookupsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSNameCachehitrate_Base),
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentReadsfromDRA),
		"read",
		"replication_agent",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentReadsfromKCC),
		"read",
		"knowledge_consistency_checker",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentReadsfromLSA),
		"read",
		"local_security_authority",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentReadsfromNSPI),
		"read",
		"name_service_provider_interface",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentReadsfromNTDSAPI),
		"read",
		"directory_service_api",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentReadsfromSAM),
		"read",
		"security_account_manager",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentReadsOther),
		"read",
		"other",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentSearchesfromDRA),
		"search",
		"replication_agent",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentSearchesfromKCC),
		"search",
		"knowledge_consistency_checker",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentSearchesfromLDAP),
		"search",
		"ldap",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentSearchesfromLSA),
		"search",
		"local_security_authority",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentSearchesfromNSPI),
		"search",
		"name_service_provider_interface",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentSearchesfromNTDSAPI),
		"search",
		"directory_service_api",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentSearchesfromSAM),
		"search",
		"security_account_manager",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentSearchesOther),
		"search",
		"other",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentWritesfromDRA),
		"write",
		"replication_agent",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentWritesfromKCC),
		"write",
		"knowledge_consistency_checker",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentWritesfromLDAP),
		"write",
		"ldap",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentWritesfromLSA),
		"write",
		"local_security_authority",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentWritesfromNSPI),
		"write",
		"name_service_provider_interface",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentWritesfromNTDSAPI),
		"write",
		"directory_service_api",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentWritesfromSAM),
		"write",
		"security_account_manager",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentWritesOther),
		"write",
		"other",
	)

	ch <- prometheus.MustNewConstMetric(
		c.directorySearchSubOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSSearchsuboperationsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.securityDescriptorPropagationEventsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSSecurityDescriptorsuboperationsPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.securityDescriptorPropagationEventsQueued,
		prometheus.GaugeValue,
		float64(dst[0].DSSecurityDescriptorPropagationsEvents),
	)
	ch <- prometheus.MustNewConstMetric(
		c.securityDescriptorPropagationAccessWaitTotalSeconds,
		prometheus.GaugeValue,
		float64(dst[0].DSSecurityDescriptorPropagatorAverageExclusionTime),
	)
	ch <- prometheus.MustNewConstMetric(
		c.securityDescriptorPropagationItemsQueuedTotal,
		prometheus.CounterValue,
		float64(dst[0].DSSecurityDescriptorPropagatorRuntimeQueue),
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryServiceThreads,
		prometheus.GaugeValue,
		float64(dst[0].DSThreadsinUse),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ldapClosedConnectionsTotal,
		prometheus.CounterValue,
		float64(dst[0].LDAPClosedConnectionsPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.ldapOpenedConnectionsTotal,
		prometheus.CounterValue,
		float64(dst[0].LDAPNewConnectionsPersec),
		"ldap",
	)
	ch <- prometheus.MustNewConstMetric(
		c.ldapOpenedConnectionsTotal,
		prometheus.CounterValue,
		float64(dst[0].LDAPNewSSLConnectionsPersec),
		"ldaps",
	)

	ch <- prometheus.MustNewConstMetric(
		c.ldapActiveThreads,
		prometheus.GaugeValue,
		float64(dst[0].LDAPActiveThreads),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ldapLastBindTimeSeconds,
		prometheus.GaugeValue,
		float64(dst[0].LDAPBindTime)/1000,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ldapSearchesTotal,
		prometheus.CounterValue,
		float64(dst[0].LDAPSearchesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ldapUdpOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].LDAPUDPoperationsPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.ldapWritesTotal,
		prometheus.CounterValue,
		float64(dst[0].LDAPWritesPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.ldapClientSessions,
		prometheus.GaugeValue,
		float64(dst[0].LDAPClientSessions),
	)

	ch <- prometheus.MustNewConstMetric(
		c.linkValuesCleanedTotal,
		prometheus.CounterValue,
		float64(dst[0].LinkValuesCleanedPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.phantomObjectsCleanedTotal,
		prometheus.CounterValue,
		float64(dst[0].PhantomsCleanedPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.phantomObjectsVisitedTotal,
		prometheus.CounterValue,
		float64(dst[0].PhantomsVisitedPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.samGroupMembershipEvaluationsTotal,
		prometheus.CounterValue,
		float64(dst[0].SAMGlobalGroupMembershipEvaluationsPersec),
		"global",
	)
	ch <- prometheus.MustNewConstMetric(
		c.samGroupMembershipEvaluationsTotal,
		prometheus.CounterValue,
		float64(dst[0].SAMDomainLocalGroupMembershipEvaluationsPersec),
		"domain_local",
	)
	ch <- prometheus.MustNewConstMetric(
		c.samGroupMembershipEvaluationsTotal,
		prometheus.CounterValue,
		float64(dst[0].SAMUniversalGroupMembershipEvaluationsPersec),
		"universal",
	)
	ch <- prometheus.MustNewConstMetric(
		c.samGroupMembershipGlobalCatalogEvaluationsTotal,
		prometheus.CounterValue,
		float64(dst[0].SAMGCEvaluationsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.samGroupMembershipEvaluationsNonTransitiveTotal,
		prometheus.CounterValue,
		float64(dst[0].SAMNonTransitiveMembershipEvaluationsPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.samGroupMembershipEvaluationsTransitiveTotal,
		prometheus.CounterValue,
		float64(dst[0].SAMTransitiveMembershipEvaluationsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.samGroupEvaluationLatency,
		prometheus.GaugeValue,
		float64(dst[0].SAMAccountGroupEvaluationLatency),
		"account_group",
	)
	ch <- prometheus.MustNewConstMetric(
		c.samGroupEvaluationLatency,
		prometheus.GaugeValue,
		float64(dst[0].SAMResourceGroupEvaluationLatency),
		"resource_group",
	)

	ch <- prometheus.MustNewConstMetric(
		c.samComputerCreationRequestsTotal,
		prometheus.CounterValue,
		float64(dst[0].SAMSuccessfulComputerCreationsPersecIncludesallrequests),
	)
	ch <- prometheus.MustNewConstMetric(
		c.samComputerCreationSuccessfulRequestsTotal,
		prometheus.CounterValue,
		float64(dst[0].SAMMachineCreationAttemptsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.samUserCreationRequestsTotal,
		prometheus.CounterValue,
		float64(dst[0].SAMUserCreationAttemptsPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.samUserCreationSuccessfulRequestsTotal,
		prometheus.CounterValue,
		float64(dst[0].SAMSuccessfulUserCreationsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.samQueryDisplayRequestsTotal,
		prometheus.CounterValue,
		float64(dst[0].SAMDisplayInformationQueriesPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.samEnumerationsTotal,
		prometheus.CounterValue,
		float64(dst[0].SAMEnumerationsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.samMembershipChangesTotal,
		prometheus.CounterValue,
		float64(dst[0].SAMMembershipChangesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.samPasswordChangesTotal,
		prometheus.CounterValue,
		float64(dst[0].SAMPasswordChangesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.tombstonesObjectsCollectedTotal,
		prometheus.CounterValue,
		float64(dst[0].TombstonesGarbageCollectedPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.tombstonesObjectsVisitedTotal,
		prometheus.CounterValue,
		float64(dst[0].TombstonesVisitedPersec),
	)

	return nil
}

func (c *Collector) collectPDH(ch chan<- prometheus.Metric) error {
	data, err := c.perfDataCollector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect DirectoryServices (AD) metrics: %w", err)
	}

	adData, ok := data["NTDS"]

	if !ok {
		return errors.New("perflib query for DirectoryServices (AD) returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.addressBookOperationsTotal,
		prometheus.CounterValue,
		adData[abANRPerSec].FirstValue,
		"ambiguous_name_resolution",
	)
	ch <- prometheus.MustNewConstMetric(
		c.addressBookOperationsTotal,
		prometheus.CounterValue,
		adData[abBrowsesPerSec].FirstValue,
		"browse",
	)
	ch <- prometheus.MustNewConstMetric(
		c.addressBookOperationsTotal,
		prometheus.CounterValue,
		adData[abMatchesPerSec].FirstValue,
		"find",
	)
	ch <- prometheus.MustNewConstMetric(
		c.addressBookOperationsTotal,
		prometheus.CounterValue,
		adData[abPropertyReadsPerSec].FirstValue,
		"property_read",
	)
	ch <- prometheus.MustNewConstMetric(
		c.addressBookOperationsTotal,
		prometheus.CounterValue,
		adData[abSearchesPerSec].FirstValue,
		"search",
	)
	ch <- prometheus.MustNewConstMetric(
		c.addressBookOperationsTotal,
		prometheus.CounterValue,
		adData[abProxyLookupsPerSec].FirstValue,
		"proxy_search",
	)

	ch <- prometheus.MustNewConstMetric(
		c.addressBookClientSessions,
		prometheus.GaugeValue,
		adData[abClientSessions].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.approximateHighestDistinguishedNameTag,
		prometheus.GaugeValue,
		adData[approximateHighestDNT].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.atqEstimatedDelaySeconds,
		prometheus.GaugeValue,
		adData[atqEstimatedQueueDelay].FirstValue/1000,
	)
	ch <- prometheus.MustNewConstMetric(
		c.atqOutstandingRequests,
		prometheus.GaugeValue,
		adData[atqOutstandingQueuedRequests].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.atqAverageRequestLatency,
		prometheus.GaugeValue,
		adData[atqRequestLatency].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.atqCurrentThreads,
		prometheus.GaugeValue,
		adData[atqThreadsLDAP].FirstValue,
		"ldap",
	)
	ch <- prometheus.MustNewConstMetric(
		c.atqCurrentThreads,
		prometheus.GaugeValue,
		adData[atqThreadsOther].FirstValue,
		"other",
	)

	ch <- prometheus.MustNewConstMetric(
		c.searchesTotal,
		prometheus.CounterValue,
		adData[baseSearchesPerSec].FirstValue,
		"base",
	)
	ch <- prometheus.MustNewConstMetric(
		c.searchesTotal,
		prometheus.CounterValue,
		adData[subtreeSearchesPerSec].FirstValue,
		"subtree",
	)
	ch <- prometheus.MustNewConstMetric(
		c.searchesTotal,
		prometheus.CounterValue,
		adData[oneLevelSearchesPerSec].FirstValue,
		"one_level",
	)

	ch <- prometheus.MustNewConstMetric(
		c.databaseOperationsTotal,
		prometheus.CounterValue,
		adData[databaseAddsPerSec].FirstValue,
		"add",
	)
	ch <- prometheus.MustNewConstMetric(
		c.databaseOperationsTotal,
		prometheus.CounterValue,
		adData[databaseDeletesPerSec].FirstValue,
		"delete",
	)
	ch <- prometheus.MustNewConstMetric(
		c.databaseOperationsTotal,
		prometheus.CounterValue,
		adData[databaseModifiesPerSec].FirstValue,
		"modify",
	)
	ch <- prometheus.MustNewConstMetric(
		c.databaseOperationsTotal,
		prometheus.CounterValue,
		adData[databaseRecyclesPerSec].FirstValue,
		"recycle",
	)

	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		adData[digestBindsPerSec].FirstValue,
		"digest",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		adData[dsClientBindsPerSec].FirstValue,
		"ds_client",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		adData[dsServerBindsPerSec].FirstValue,
		"ds_server",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		adData[externalBindsPerSec].FirstValue,
		"external",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		adData[fastBindsPerSec].FirstValue,
		"fast",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		adData[negotiatedBindsPerSec].FirstValue,
		"negotiate",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		adData[ntlmBindsPerSec].FirstValue,
		"ntlm",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		adData[simpleBindsPerSec].FirstValue,
		"simple",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		adData[ldapSuccessfulBindsPerSec].FirstValue,
		"ldap",
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationHighestUsn,
		prometheus.CounterValue,
		float64(uint64(adData[draHighestUSNCommittedHighPart].FirstValue)<<32)+adData[draHighestUSNCommittedLowPart].FirstValue,
		"committed",
	)
	ch <- prometheus.MustNewConstMetric(
		c.replicationHighestUsn,
		prometheus.CounterValue,
		float64(uint64(adData[draHighestUSNIssuedHighPart].FirstValue)<<32)+adData[draHighestUSNIssuedLowPart].FirstValue,
		"issued",
	)

	ch <- prometheus.MustNewConstMetric(
		c.interSiteReplicationDataBytesTotal,
		prometheus.CounterValue,
		adData[draInboundBytesCompressedBetweenSitesAfterCompressionPerSec].FirstValue,
		"inbound",
	)
	// The pre-compression data size seems to have little value? Skipping for now
	// ch <- prometheus.MustNewConstMetric(
	// 	c.interSiteReplicationDataBytesTotal,
	// 	prometheus.CounterValue,
	// 	float64(dst[0].DRAInboundBytesCompressedBetweenSitesBeforeCompressionPersec),
	// 	"inbound",
	// )
	ch <- prometheus.MustNewConstMetric(
		c.interSiteReplicationDataBytesTotal,
		prometheus.CounterValue,
		adData[draOutboundBytesCompressedBetweenSitesAfterCompressionPerSec].FirstValue,
		"outbound",
	)
	// ch <- prometheus.MustNewConstMetric(
	// 	c.interSiteReplicationDataBytesTotal,
	// 	prometheus.CounterValue,
	// 	float64(dst[0].DRAOutboundBytesCompressedBetweenSitesBeforeCompressionPersec),
	// 	"outbound",
	// )
	ch <- prometheus.MustNewConstMetric(
		c.intraSiteReplicationDataBytesTotal,
		prometheus.CounterValue,
		adData[draInboundBytesNotCompressedWithinSitePerSec].FirstValue,
		"inbound",
	)
	ch <- prometheus.MustNewConstMetric(
		c.intraSiteReplicationDataBytesTotal,
		prometheus.CounterValue,
		adData[draOutboundBytesNotCompressedWithinSitePerSec].FirstValue,
		"outbound",
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationInboundSyncObjectsRemaining,
		prometheus.GaugeValue,
		adData[draInboundFullSyncObjectsRemaining].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationInboundLinkValueUpdatesRemaining,
		prometheus.GaugeValue,
		adData[draInboundLinkValueUpdatesRemainingInPacket].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationInboundObjectsUpdatedTotal,
		prometheus.CounterValue,
		adData[draInboundObjectsAppliedPerSec].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.replicationInboundObjectsFilteredTotal,
		prometheus.CounterValue,
		adData[draInboundObjectsFilteredPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationInboundPropertiesUpdatedTotal,
		prometheus.CounterValue,
		adData[draInboundPropertiesAppliedPerSec].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.replicationInboundPropertiesFilteredTotal,
		prometheus.CounterValue,
		adData[draInboundPropertiesFilteredPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationPendingOperations,
		prometheus.GaugeValue,
		adData[draPendingReplicationOperations].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.replicationPendingSynchronizations,
		prometheus.GaugeValue,
		adData[draPendingReplicationSynchronizations].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationSyncRequestsTotal,
		prometheus.CounterValue,
		adData[draSyncRequestsMade].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.replicationSyncRequestsSuccessTotal,
		prometheus.CounterValue,
		adData[draSyncRequestsSuccessful].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.replicationSyncRequestsSchemaMismatchFailureTotal,
		prometheus.CounterValue,
		adData[draSyncFailuresOnSchemaMismatch].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.nameTranslationsTotal,
		prometheus.CounterValue,
		adData[dsClientNameTranslationsPerSec].FirstValue,
		"client",
	)
	ch <- prometheus.MustNewConstMetric(
		c.nameTranslationsTotal,
		prometheus.CounterValue,
		adData[dsServerNameTranslationsPerSec].FirstValue,
		"server",
	)

	ch <- prometheus.MustNewConstMetric(
		c.changeMonitorsRegistered,
		prometheus.GaugeValue,
		adData[dsMonitorListSize].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.changeMonitorUpdatesPending,
		prometheus.GaugeValue,
		adData[dsNotifyQueueSize].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.nameCacheHitsTotal,
		prometheus.CounterValue,
		adData[dsNameCacheHitRate].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.nameCacheLookupsTotal,
		prometheus.CounterValue,
		adData[dsNameCacheHitRate].SecondValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		adData[dsPercentReadsFromDRA].FirstValue,
		"read",
		"replication_agent",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		adData[dsPercentReadsFromKCC].FirstValue,
		"read",
		"knowledge_consistency_checker",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		adData[dsPercentReadsFromLSA].FirstValue,
		"read",
		"local_security_authority",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		adData[dsPercentReadsFromNSPI].FirstValue,
		"read",
		"name_service_provider_interface",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		adData[dsPercentReadsFromNTDSAPI].FirstValue,
		"read",
		"directory_service_api",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		adData[dsPercentReadsFromSAM].FirstValue,
		"read",
		"security_account_manager",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		adData[dsPercentReadsOther].FirstValue,
		"read",
		"other",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		adData[dsPercentSearchesFromDRA].FirstValue,
		"search",
		"replication_agent",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		adData[dsPercentSearchesFromKCC].FirstValue,
		"search",
		"knowledge_consistency_checker",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		adData[dsPercentSearchesFromLDAP].FirstValue,
		"search",
		"ldap",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		adData[dsPercentSearchesFromLSA].FirstValue,
		"search",
		"local_security_authority",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		adData[dsPercentSearchesFromNSPI].FirstValue,
		"search",
		"name_service_provider_interface",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		adData[dsPercentSearchesFromNTDSAPI].FirstValue,
		"search",
		"directory_service_api",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		adData[dsPercentSearchesFromSAM].FirstValue,
		"search",
		"security_account_manager",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		adData[dsPercentSearchesOther].FirstValue,
		"search",
		"other",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		adData[dsPercentWritesFromDRA].FirstValue,
		"write",
		"replication_agent",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		adData[dsPercentWritesFromKCC].FirstValue,
		"write",
		"knowledge_consistency_checker",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		adData[dsPercentWritesFromLDAP].FirstValue,
		"write",
		"ldap",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		adData[dsPercentSearchesFromLSA].FirstValue,
		"write",
		"local_security_authority",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		adData[dsPercentWritesFromNSPI].FirstValue,
		"write",
		"name_service_provider_interface",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		adData[dsPercentWritesFromNTDSAPI].FirstValue,
		"write",
		"directory_service_api",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		adData[dsPercentWritesFromSAM].FirstValue,
		"write",
		"security_account_manager",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		adData[dsPercentWritesOther].FirstValue,
		"write",
		"other",
	)

	ch <- prometheus.MustNewConstMetric(
		c.directorySearchSubOperationsTotal,
		prometheus.CounterValue,
		adData[dsSearchSubOperationsPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.securityDescriptorPropagationEventsTotal,
		prometheus.CounterValue,
		adData[dsSecurityDescriptorSubOperationsPerSec].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.securityDescriptorPropagationEventsQueued,
		prometheus.GaugeValue,
		adData[dsSecurityDescriptorPropagationsEvents].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.securityDescriptorPropagationAccessWaitTotalSeconds,
		prometheus.GaugeValue,
		adData[dsSecurityDescriptorPropagatorAverageExclusionTime].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.securityDescriptorPropagationItemsQueuedTotal,
		prometheus.CounterValue,
		adData[dsSecurityDescriptorPropagatorRuntimeQueue].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryServiceThreads,
		prometheus.GaugeValue,
		adData[dsThreadsInUse].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ldapClosedConnectionsTotal,
		prometheus.CounterValue,
		adData[ldapClosedConnectionsPerSec].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.ldapOpenedConnectionsTotal,
		prometheus.CounterValue,
		adData[ldapNewConnectionsPerSec].FirstValue,
		"ldap",
	)
	ch <- prometheus.MustNewConstMetric(
		c.ldapOpenedConnectionsTotal,
		prometheus.CounterValue,
		adData[ldapNewSSLConnectionsPerSec].FirstValue,
		"ldaps",
	)

	ch <- prometheus.MustNewConstMetric(
		c.ldapActiveThreads,
		prometheus.GaugeValue,
		adData[ldapActiveThreads].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ldapLastBindTimeSeconds,
		prometheus.GaugeValue,
		adData[ldapBindTime].FirstValue/1000,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ldapSearchesTotal,
		prometheus.CounterValue,
		adData[ldapSearchesPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ldapUdpOperationsTotal,
		prometheus.CounterValue,
		adData[ldapUDPOperationsPerSec].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.ldapWritesTotal,
		prometheus.CounterValue,
		adData[ldapWritesPerSec].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.ldapClientSessions,
		prometheus.GaugeValue,
		adData[ldapClientSessions].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.linkValuesCleanedTotal,
		prometheus.CounterValue,
		adData[linkValuesCleanedPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.phantomObjectsCleanedTotal,
		prometheus.CounterValue,
		adData[phantomsCleanedPerSec].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.phantomObjectsVisitedTotal,
		prometheus.CounterValue,
		adData[phantomsVisitedPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samGroupMembershipEvaluationsTotal,
		prometheus.CounterValue,
		adData[samGlobalGroupMembershipEvaluationsPerSec].FirstValue,
		"global",
	)
	ch <- prometheus.MustNewConstMetric(
		c.samGroupMembershipEvaluationsTotal,
		prometheus.CounterValue,
		adData[samDomainLocalGroupMembershipEvaluationsPerSec].FirstValue,
		"domain_local",
	)
	ch <- prometheus.MustNewConstMetric(
		c.samGroupMembershipEvaluationsTotal,
		prometheus.CounterValue,
		adData[samUniversalGroupMembershipEvaluationsPerSec].FirstValue,
		"universal",
	)
	ch <- prometheus.MustNewConstMetric(
		c.samGroupMembershipGlobalCatalogEvaluationsTotal,
		prometheus.CounterValue,
		adData[samGCEvaluationsPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samGroupMembershipEvaluationsNonTransitiveTotal,
		prometheus.CounterValue,
		adData[samNonTransitiveMembershipEvaluationsPerSec].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.samGroupMembershipEvaluationsTransitiveTotal,
		prometheus.CounterValue,
		adData[samTransitiveMembershipEvaluationsPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samGroupEvaluationLatency,
		prometheus.GaugeValue,
		adData[samAccountGroupEvaluationLatency].FirstValue,
		"account_group",
	)
	ch <- prometheus.MustNewConstMetric(
		c.samGroupEvaluationLatency,
		prometheus.GaugeValue,
		adData[samResourceGroupEvaluationLatency].FirstValue,
		"resource_group",
	)

	ch <- prometheus.MustNewConstMetric(
		c.samComputerCreationRequestsTotal,
		prometheus.CounterValue,
		adData[samSuccessfulComputerCreationsPerSecIncludesAllRequests].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.samComputerCreationSuccessfulRequestsTotal,
		prometheus.CounterValue,
		adData[samMachineCreationAttemptsPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samUserCreationRequestsTotal,
		prometheus.CounterValue,
		adData[samUserCreationAttemptsPerSec].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.samUserCreationSuccessfulRequestsTotal,
		prometheus.CounterValue,
		adData[samSuccessfulUserCreationsPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samQueryDisplayRequestsTotal,
		prometheus.CounterValue,
		adData[samDisplayInformationQueriesPerSec].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.samEnumerationsTotal,
		prometheus.CounterValue,
		adData[samEnumerationsPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samMembershipChangesTotal,
		prometheus.CounterValue,
		adData[samMembershipChangesPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samPasswordChangesTotal,
		prometheus.CounterValue,
		adData[samPasswordChangesPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.tombstonesObjectsCollectedTotal,
		prometheus.CounterValue,
		adData[tombstonesGarbageCollectedPerSec].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.tombstonesObjectsVisitedTotal,
		prometheus.CounterValue,
		adData[tombstonesVisitedPerSec].FirstValue,
	)

	return nil
}

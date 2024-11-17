//go:build windows

package ad

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "ad"

type Config struct{}

var ConfigDefaults = Config{}

type Collector struct {
	config Config

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

func (c *Collector) Close() error {
	c.perfDataCollector.Close()

	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *mi.Session) error {
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

	c.perfDataCollector, err = perfdata.NewCollector("DirectoryServices", perfdata.InstanceAll, counters)
	if err != nil {
		return fmt.Errorf("failed to create DirectoryServices collector: %w", err)
	}

	c.addressBookOperationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "address_book_operations_total"),
		"",
		[]string{"operation"},
		nil,
	)
	c.addressBookClientSessions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "address_book_client_sessions"),
		"",
		nil,
		nil,
	)
	c.approximateHighestDistinguishedNameTag = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "approximate_highest_distinguished_name_tag"),
		"",
		nil,
		nil,
	)
	c.atqEstimatedDelaySeconds = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "atq_estimated_delay_seconds"),
		"",
		nil,
		nil,
	)
	c.atqOutstandingRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "atq_outstanding_requests"),
		"",
		nil,
		nil,
	)
	c.atqAverageRequestLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "atq_average_request_latency"),
		"",
		nil,
		nil,
	)
	c.atqCurrentThreads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "atq_current_threads"),
		"",
		[]string{"service"},
		nil,
	)
	c.searchesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "searches_total"),
		"",
		[]string{"scope"},
		nil,
	)
	c.databaseOperationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "database_operations_total"),
		"",
		[]string{"operation"},
		nil,
	)
	c.bindsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "binds_total"),
		"",
		[]string{"bind_method"},
		nil,
	)
	c.replicationHighestUsn = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_highest_usn"),
		"",
		[]string{"state"},
		nil,
	)
	c.intraSiteReplicationDataBytesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_data_intrasite_bytes_total"),
		"",
		[]string{"direction"},
		nil,
	)
	c.interSiteReplicationDataBytesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_data_intersite_bytes_total"),
		"",
		[]string{"direction"},
		nil,
	)
	c.replicationInboundSyncObjectsRemaining = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_inbound_sync_objects_remaining"),
		"",
		nil,
		nil,
	)
	c.replicationInboundLinkValueUpdatesRemaining = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_inbound_link_value_updates_remaining"),
		"",
		nil,
		nil,
	)
	c.replicationInboundObjectsUpdatedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_inbound_objects_updated_total"),
		"",
		nil,
		nil,
	)
	c.replicationInboundObjectsFilteredTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_inbound_objects_filtered_total"),
		"",
		nil,
		nil,
	)
	c.replicationInboundPropertiesUpdatedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_inbound_properties_updated_total"),
		"",
		nil,
		nil,
	)
	c.replicationInboundPropertiesFilteredTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_inbound_properties_filtered_total"),
		"",
		nil,
		nil,
	)
	c.replicationPendingOperations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_pending_operations"),
		"",
		nil,
		nil,
	)
	c.replicationPendingSynchronizations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_pending_synchronizations"),
		"",
		nil,
		nil,
	)
	c.replicationSyncRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_sync_requests_total"),
		"",
		nil,
		nil,
	)
	c.replicationSyncRequestsSuccessTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_sync_requests_success_total"),
		"",
		nil,
		nil,
	)
	c.replicationSyncRequestsSchemaMismatchFailureTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_sync_requests_schema_mismatch_failure_total"),
		"",
		nil,
		nil,
	)
	c.nameTranslationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "name_translations_total"),
		"",
		[]string{"target_name"},
		nil,
	)
	c.changeMonitorsRegistered = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "change_monitors_registered"),
		"",
		nil,
		nil,
	)
	c.changeMonitorUpdatesPending = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "change_monitor_updates_pending"),
		"",
		nil,
		nil,
	)
	c.nameCacheHitsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "name_cache_hits_total"),
		"",
		nil,
		nil,
	)
	c.nameCacheLookupsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "name_cache_lookups_total"),
		"",
		nil,
		nil,
	)
	c.directoryOperationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "directory_operations_total"),
		"",
		[]string{"operation", "origin"},
		nil,
	)
	c.directorySearchSubOperationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "directory_search_suboperations_total"),
		"",
		nil,
		nil,
	)
	c.securityDescriptorPropagationEventsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "security_descriptor_propagation_events_total"),
		"",
		nil,
		nil,
	)
	c.securityDescriptorPropagationEventsQueued = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "security_descriptor_propagation_events_queued"),
		"",
		nil,
		nil,
	)
	c.securityDescriptorPropagationAccessWaitTotalSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "security_descriptor_propagation_access_wait_total_seconds"),
		"",
		nil,
		nil,
	)
	c.securityDescriptorPropagationItemsQueuedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "security_descriptor_propagation_items_queued_total"),
		"",
		nil,
		nil,
	)
	c.directoryServiceThreads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "directory_service_threads"),
		"",
		nil,
		nil,
	)
	c.ldapClosedConnectionsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ldap_closed_connections_total"),
		"",
		nil,
		nil,
	)
	c.ldapOpenedConnectionsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ldap_opened_connections_total"),
		"",
		[]string{"type"},
		nil,
	)
	c.ldapActiveThreads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ldap_active_threads"),
		"",
		nil,
		nil,
	)
	c.ldapLastBindTimeSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ldap_last_bind_time_seconds"),
		"",
		nil,
		nil,
	)
	c.ldapSearchesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ldap_searches_total"),
		"",
		nil,
		nil,
	)
	c.ldapUdpOperationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ldap_udp_operations_total"),
		"",
		nil,
		nil,
	)
	c.ldapWritesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ldap_writes_total"),
		"",
		nil,
		nil,
	)
	c.ldapClientSessions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ldap_client_sessions"),
		"This is the number of sessions opened by LDAP clients at the time the data is taken. This is helpful in determining LDAP client activity and if the DC is able to handle the load. Of course, spikes during normal periods of authentication — such as first thing in the morning — are not necessarily a problem, but long sustained periods of high values indicate an overworked DC.",
		nil,
		nil,
	)
	c.linkValuesCleanedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "link_values_cleaned_total"),
		"",
		nil,
		nil,
	)
	c.phantomObjectsCleanedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "phantom_objects_cleaned_total"),
		"",
		nil,
		nil,
	)
	c.phantomObjectsVisitedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "phantom_objects_visited_total"),
		"",
		nil,
		nil,
	)
	c.samGroupMembershipEvaluationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_group_membership_evaluations_total"),
		"",
		[]string{"group_type"},
		nil,
	)
	c.samGroupMembershipGlobalCatalogEvaluationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_group_membership_global_catalog_evaluations_total"),
		"",
		nil,
		nil,
	)
	c.samGroupMembershipEvaluationsNonTransitiveTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_group_membership_evaluations_nontransitive_total"),
		"",
		nil,
		nil,
	)
	c.samGroupMembershipEvaluationsTransitiveTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_group_membership_evaluations_transitive_total"),
		"",
		nil,
		nil,
	)
	c.samGroupEvaluationLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_group_evaluation_latency"),
		"The mean latency of the last 100 group evaluations performed for authentication",
		[]string{"evaluation_type"},
		nil,
	)
	c.samComputerCreationRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_computer_creation_requests_total"),
		"",
		nil,
		nil,
	)
	c.samComputerCreationSuccessfulRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_computer_creation_successful_requests_total"),
		"",
		nil,
		nil,
	)
	c.samUserCreationRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_user_creation_requests_total"),
		"",
		nil,
		nil,
	)
	c.samUserCreationSuccessfulRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_user_creation_successful_requests_total"),
		"",
		nil,
		nil,
	)
	c.samQueryDisplayRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_query_display_requests_total"),
		"",
		nil,
		nil,
	)
	c.samEnumerationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_enumerations_total"),
		"",
		nil,
		nil,
	)
	c.samMembershipChangesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_membership_changes_total"),
		"",
		nil,
		nil,
	)
	c.samPasswordChangesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_password_changes_total"),
		"",
		nil,
		nil,
	)

	c.tombstonesObjectsCollectedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "tombstoned_objects_collected_total"),
		"",
		nil,
		nil,
	)
	c.tombstonesObjectsVisitedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "tombstoned_objects_visited_total"),
		"",
		nil,
		nil,
	)

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect DirectoryServices (AD) metrics: %w", err)
	}

	data, ok := perfData["NTDS"]

	if !ok {
		return errors.New("perflib query for DirectoryServices (AD) returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.addressBookOperationsTotal,
		prometheus.CounterValue,
		data[abANRPerSec].FirstValue,
		"ambiguous_name_resolution",
	)
	ch <- prometheus.MustNewConstMetric(
		c.addressBookOperationsTotal,
		prometheus.CounterValue,
		data[abBrowsesPerSec].FirstValue,
		"browse",
	)
	ch <- prometheus.MustNewConstMetric(
		c.addressBookOperationsTotal,
		prometheus.CounterValue,
		data[abMatchesPerSec].FirstValue,
		"find",
	)
	ch <- prometheus.MustNewConstMetric(
		c.addressBookOperationsTotal,
		prometheus.CounterValue,
		data[abPropertyReadsPerSec].FirstValue,
		"property_read",
	)
	ch <- prometheus.MustNewConstMetric(
		c.addressBookOperationsTotal,
		prometheus.CounterValue,
		data[abSearchesPerSec].FirstValue,
		"search",
	)
	ch <- prometheus.MustNewConstMetric(
		c.addressBookOperationsTotal,
		prometheus.CounterValue,
		data[abProxyLookupsPerSec].FirstValue,
		"proxy_search",
	)

	ch <- prometheus.MustNewConstMetric(
		c.addressBookClientSessions,
		prometheus.GaugeValue,
		data[abClientSessions].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.approximateHighestDistinguishedNameTag,
		prometheus.GaugeValue,
		data[approximateHighestDNT].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.atqEstimatedDelaySeconds,
		prometheus.GaugeValue,
		data[atqEstimatedQueueDelay].FirstValue/1000,
	)
	ch <- prometheus.MustNewConstMetric(
		c.atqOutstandingRequests,
		prometheus.GaugeValue,
		data[atqOutstandingQueuedRequests].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.atqAverageRequestLatency,
		prometheus.GaugeValue,
		data[atqRequestLatency].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.atqCurrentThreads,
		prometheus.GaugeValue,
		data[atqThreadsLDAP].FirstValue,
		"ldap",
	)
	ch <- prometheus.MustNewConstMetric(
		c.atqCurrentThreads,
		prometheus.GaugeValue,
		data[atqThreadsOther].FirstValue,
		"other",
	)

	ch <- prometheus.MustNewConstMetric(
		c.searchesTotal,
		prometheus.CounterValue,
		data[baseSearchesPerSec].FirstValue,
		"base",
	)
	ch <- prometheus.MustNewConstMetric(
		c.searchesTotal,
		prometheus.CounterValue,
		data[subtreeSearchesPerSec].FirstValue,
		"subtree",
	)
	ch <- prometheus.MustNewConstMetric(
		c.searchesTotal,
		prometheus.CounterValue,
		data[oneLevelSearchesPerSec].FirstValue,
		"one_level",
	)

	ch <- prometheus.MustNewConstMetric(
		c.databaseOperationsTotal,
		prometheus.CounterValue,
		data[databaseAddsPerSec].FirstValue,
		"add",
	)
	ch <- prometheus.MustNewConstMetric(
		c.databaseOperationsTotal,
		prometheus.CounterValue,
		data[databaseDeletesPerSec].FirstValue,
		"delete",
	)
	ch <- prometheus.MustNewConstMetric(
		c.databaseOperationsTotal,
		prometheus.CounterValue,
		data[databaseModifiesPerSec].FirstValue,
		"modify",
	)
	ch <- prometheus.MustNewConstMetric(
		c.databaseOperationsTotal,
		prometheus.CounterValue,
		data[databaseRecyclesPerSec].FirstValue,
		"recycle",
	)

	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		data[digestBindsPerSec].FirstValue,
		"digest",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		data[dsClientBindsPerSec].FirstValue,
		"ds_client",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		data[dsServerBindsPerSec].FirstValue,
		"ds_server",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		data[externalBindsPerSec].FirstValue,
		"external",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		data[fastBindsPerSec].FirstValue,
		"fast",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		data[negotiatedBindsPerSec].FirstValue,
		"negotiate",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		data[ntlmBindsPerSec].FirstValue,
		"ntlm",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		data[simpleBindsPerSec].FirstValue,
		"simple",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		data[ldapSuccessfulBindsPerSec].FirstValue,
		"ldap",
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationHighestUsn,
		prometheus.CounterValue,
		float64(uint64(data[draHighestUSNCommittedHighPart].FirstValue)<<32)+data[draHighestUSNCommittedLowPart].FirstValue,
		"committed",
	)
	ch <- prometheus.MustNewConstMetric(
		c.replicationHighestUsn,
		prometheus.CounterValue,
		float64(uint64(data[draHighestUSNIssuedHighPart].FirstValue)<<32)+data[draHighestUSNIssuedLowPart].FirstValue,
		"issued",
	)

	ch <- prometheus.MustNewConstMetric(
		c.interSiteReplicationDataBytesTotal,
		prometheus.CounterValue,
		data[draInboundBytesCompressedBetweenSitesAfterCompressionPerSec].FirstValue,
		"inbound",
	)
	// The pre-compression perfData size seems to have little value? Skipping for now
	// ch <- prometheus.MustNewConstMetric(
	// 	c.interSiteReplicationDataBytesTotal,
	// 	prometheus.CounterValue,
	// 	float64(dst[0].DRAInboundBytesCompressedBetweenSitesBeforeCompressionPersec),
	// 	"inbound",
	// )
	ch <- prometheus.MustNewConstMetric(
		c.interSiteReplicationDataBytesTotal,
		prometheus.CounterValue,
		data[draOutboundBytesCompressedBetweenSitesAfterCompressionPerSec].FirstValue,
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
		data[draInboundBytesNotCompressedWithinSitePerSec].FirstValue,
		"inbound",
	)
	ch <- prometheus.MustNewConstMetric(
		c.intraSiteReplicationDataBytesTotal,
		prometheus.CounterValue,
		data[draOutboundBytesNotCompressedWithinSitePerSec].FirstValue,
		"outbound",
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationInboundSyncObjectsRemaining,
		prometheus.GaugeValue,
		data[draInboundFullSyncObjectsRemaining].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationInboundLinkValueUpdatesRemaining,
		prometheus.GaugeValue,
		data[draInboundLinkValueUpdatesRemainingInPacket].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationInboundObjectsUpdatedTotal,
		prometheus.CounterValue,
		data[draInboundObjectsAppliedPerSec].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.replicationInboundObjectsFilteredTotal,
		prometheus.CounterValue,
		data[draInboundObjectsFilteredPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationInboundPropertiesUpdatedTotal,
		prometheus.CounterValue,
		data[draInboundPropertiesAppliedPerSec].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.replicationInboundPropertiesFilteredTotal,
		prometheus.CounterValue,
		data[draInboundPropertiesFilteredPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationPendingOperations,
		prometheus.GaugeValue,
		data[draPendingReplicationOperations].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.replicationPendingSynchronizations,
		prometheus.GaugeValue,
		data[draPendingReplicationSynchronizations].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationSyncRequestsTotal,
		prometheus.CounterValue,
		data[draSyncRequestsMade].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.replicationSyncRequestsSuccessTotal,
		prometheus.CounterValue,
		data[draSyncRequestsSuccessful].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.replicationSyncRequestsSchemaMismatchFailureTotal,
		prometheus.CounterValue,
		data[draSyncFailuresOnSchemaMismatch].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.nameTranslationsTotal,
		prometheus.CounterValue,
		data[dsClientNameTranslationsPerSec].FirstValue,
		"client",
	)
	ch <- prometheus.MustNewConstMetric(
		c.nameTranslationsTotal,
		prometheus.CounterValue,
		data[dsServerNameTranslationsPerSec].FirstValue,
		"server",
	)

	ch <- prometheus.MustNewConstMetric(
		c.changeMonitorsRegistered,
		prometheus.GaugeValue,
		data[dsMonitorListSize].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.changeMonitorUpdatesPending,
		prometheus.GaugeValue,
		data[dsNotifyQueueSize].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.nameCacheHitsTotal,
		prometheus.CounterValue,
		data[dsNameCacheHitRate].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.nameCacheLookupsTotal,
		prometheus.CounterValue,
		data[dsNameCacheHitRate].SecondValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data[dsPercentReadsFromDRA].FirstValue,
		"read",
		"replication_agent",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data[dsPercentReadsFromKCC].FirstValue,
		"read",
		"knowledge_consistency_checker",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data[dsPercentReadsFromLSA].FirstValue,
		"read",
		"local_security_authority",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data[dsPercentReadsFromNSPI].FirstValue,
		"read",
		"name_service_provider_interface",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data[dsPercentReadsFromNTDSAPI].FirstValue,
		"read",
		"directory_service_api",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data[dsPercentReadsFromSAM].FirstValue,
		"read",
		"security_account_manager",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data[dsPercentReadsOther].FirstValue,
		"read",
		"other",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data[dsPercentSearchesFromDRA].FirstValue,
		"search",
		"replication_agent",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data[dsPercentSearchesFromKCC].FirstValue,
		"search",
		"knowledge_consistency_checker",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data[dsPercentSearchesFromLDAP].FirstValue,
		"search",
		"ldap",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data[dsPercentSearchesFromLSA].FirstValue,
		"search",
		"local_security_authority",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data[dsPercentSearchesFromNSPI].FirstValue,
		"search",
		"name_service_provider_interface",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data[dsPercentSearchesFromNTDSAPI].FirstValue,
		"search",
		"directory_service_api",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data[dsPercentSearchesFromSAM].FirstValue,
		"search",
		"security_account_manager",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data[dsPercentSearchesOther].FirstValue,
		"search",
		"other",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data[dsPercentWritesFromDRA].FirstValue,
		"write",
		"replication_agent",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data[dsPercentWritesFromKCC].FirstValue,
		"write",
		"knowledge_consistency_checker",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data[dsPercentWritesFromLDAP].FirstValue,
		"write",
		"ldap",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data[dsPercentSearchesFromLSA].FirstValue,
		"write",
		"local_security_authority",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data[dsPercentWritesFromNSPI].FirstValue,
		"write",
		"name_service_provider_interface",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data[dsPercentWritesFromNTDSAPI].FirstValue,
		"write",
		"directory_service_api",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data[dsPercentWritesFromSAM].FirstValue,
		"write",
		"security_account_manager",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data[dsPercentWritesOther].FirstValue,
		"write",
		"other",
	)

	ch <- prometheus.MustNewConstMetric(
		c.directorySearchSubOperationsTotal,
		prometheus.CounterValue,
		data[dsSearchSubOperationsPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.securityDescriptorPropagationEventsTotal,
		prometheus.CounterValue,
		data[dsSecurityDescriptorSubOperationsPerSec].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.securityDescriptorPropagationEventsQueued,
		prometheus.GaugeValue,
		data[dsSecurityDescriptorPropagationsEvents].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.securityDescriptorPropagationAccessWaitTotalSeconds,
		prometheus.GaugeValue,
		data[dsSecurityDescriptorPropagatorAverageExclusionTime].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.securityDescriptorPropagationItemsQueuedTotal,
		prometheus.CounterValue,
		data[dsSecurityDescriptorPropagatorRuntimeQueue].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryServiceThreads,
		prometheus.GaugeValue,
		data[dsThreadsInUse].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ldapClosedConnectionsTotal,
		prometheus.CounterValue,
		data[ldapClosedConnectionsPerSec].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.ldapOpenedConnectionsTotal,
		prometheus.CounterValue,
		data[ldapNewConnectionsPerSec].FirstValue,
		"ldap",
	)
	ch <- prometheus.MustNewConstMetric(
		c.ldapOpenedConnectionsTotal,
		prometheus.CounterValue,
		data[ldapNewSSLConnectionsPerSec].FirstValue,
		"ldaps",
	)

	ch <- prometheus.MustNewConstMetric(
		c.ldapActiveThreads,
		prometheus.GaugeValue,
		data[ldapActiveThreads].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ldapLastBindTimeSeconds,
		prometheus.GaugeValue,
		data[ldapBindTime].FirstValue/1000,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ldapSearchesTotal,
		prometheus.CounterValue,
		data[ldapSearchesPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ldapUdpOperationsTotal,
		prometheus.CounterValue,
		data[ldapUDPOperationsPerSec].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.ldapWritesTotal,
		prometheus.CounterValue,
		data[ldapWritesPerSec].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.ldapClientSessions,
		prometheus.GaugeValue,
		data[ldapClientSessions].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.linkValuesCleanedTotal,
		prometheus.CounterValue,
		data[linkValuesCleanedPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.phantomObjectsCleanedTotal,
		prometheus.CounterValue,
		data[phantomsCleanedPerSec].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.phantomObjectsVisitedTotal,
		prometheus.CounterValue,
		data[phantomsVisitedPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samGroupMembershipEvaluationsTotal,
		prometheus.CounterValue,
		data[samGlobalGroupMembershipEvaluationsPerSec].FirstValue,
		"global",
	)
	ch <- prometheus.MustNewConstMetric(
		c.samGroupMembershipEvaluationsTotal,
		prometheus.CounterValue,
		data[samDomainLocalGroupMembershipEvaluationsPerSec].FirstValue,
		"domain_local",
	)
	ch <- prometheus.MustNewConstMetric(
		c.samGroupMembershipEvaluationsTotal,
		prometheus.CounterValue,
		data[samUniversalGroupMembershipEvaluationsPerSec].FirstValue,
		"universal",
	)
	ch <- prometheus.MustNewConstMetric(
		c.samGroupMembershipGlobalCatalogEvaluationsTotal,
		prometheus.CounterValue,
		data[samGCEvaluationsPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samGroupMembershipEvaluationsNonTransitiveTotal,
		prometheus.CounterValue,
		data[samNonTransitiveMembershipEvaluationsPerSec].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.samGroupMembershipEvaluationsTransitiveTotal,
		prometheus.CounterValue,
		data[samTransitiveMembershipEvaluationsPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samGroupEvaluationLatency,
		prometheus.GaugeValue,
		data[samAccountGroupEvaluationLatency].FirstValue,
		"account_group",
	)
	ch <- prometheus.MustNewConstMetric(
		c.samGroupEvaluationLatency,
		prometheus.GaugeValue,
		data[samResourceGroupEvaluationLatency].FirstValue,
		"resource_group",
	)

	ch <- prometheus.MustNewConstMetric(
		c.samComputerCreationRequestsTotal,
		prometheus.CounterValue,
		data[samSuccessfulComputerCreationsPerSecIncludesAllRequests].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.samComputerCreationSuccessfulRequestsTotal,
		prometheus.CounterValue,
		data[samMachineCreationAttemptsPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samUserCreationRequestsTotal,
		prometheus.CounterValue,
		data[samUserCreationAttemptsPerSec].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.samUserCreationSuccessfulRequestsTotal,
		prometheus.CounterValue,
		data[samSuccessfulUserCreationsPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samQueryDisplayRequestsTotal,
		prometheus.CounterValue,
		data[samDisplayInformationQueriesPerSec].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.samEnumerationsTotal,
		prometheus.CounterValue,
		data[samEnumerationsPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samMembershipChangesTotal,
		prometheus.CounterValue,
		data[samMembershipChangesPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samPasswordChangesTotal,
		prometheus.CounterValue,
		data[samPasswordChangesPerSec].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.tombstonesObjectsCollectedTotal,
		prometheus.CounterValue,
		data[tombstonesGarbageCollectedPerSec].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.tombstonesObjectsVisitedTotal,
		prometheus.CounterValue,
		data[tombstonesVisitedPerSec].FirstValue,
	)

	return nil
}

//go:build windows

package ad

import (
	"errors"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "ad"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_DirectoryServices_DirectoryServices metrics.
type Collector struct {
	config Config

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

func (c *Collector) GetPerfCounter(_ log.Logger) ([]string, error) {
	return []string{}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(_ log.Logger) error {
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
func (c *Collector) Collect(_ *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	if err := c.collect(logger, ch); err != nil {
		_ = level.Error(logger).Log("msg", "failed collecting ad metrics", "err", err)
		return err
	}
	return nil
}

// Win32_PerfRawData_DirectoryServices_DirectoryServices docs:
// - https://msdn.microsoft.com/en-us/library/ms803980.aspx
type Win32_PerfRawData_DirectoryServices_DirectoryServices struct {
	Name string

	ABANRPersec                                                      uint32
	ABBrowsesPersec                                                  uint32
	ABClientSessions                                                 uint32
	ABMatchesPersec                                                  uint32
	ABPropertyReadsPersec                                            uint32
	ABProxyLookupsPersec                                             uint32
	ABSearchesPersec                                                 uint32
	ApproximatehighestDNT                                            uint32
	ATQEstimatedQueueDelay                                           uint32
	ATQOutstandingQueuedRequests                                     uint32
	ATQRequestLatency                                                uint32
	ATQThreadsLDAP                                                   uint32
	ATQThreadsOther                                                  uint32
	ATQThreadsTotal                                                  uint32
	BasesearchesPersec                                               uint32
	DatabaseaddsPersec                                               uint32
	DatabasedeletesPersec                                            uint32
	DatabasemodifysPersec                                            uint32
	DatabaserecyclesPersec                                           uint32
	DigestBindsPersec                                                uint32
	DRAHighestUSNCommittedHighpart                                   uint64
	DRAHighestUSNCommittedLowpart                                    uint64
	DRAHighestUSNIssuedHighpart                                      uint64
	DRAHighestUSNIssuedLowpart                                       uint64
	DRAInboundBytesCompressedBetweenSitesAfterCompressionPersec      uint32
	DRAInboundBytesCompressedBetweenSitesAfterCompressionSinceBoot   uint32
	DRAInboundBytesCompressedBetweenSitesBeforeCompressionPersec     uint32
	DRAInboundBytesCompressedBetweenSitesBeforeCompressionSinceBoot  uint32
	DRAInboundBytesNotCompressedWithinSitePersec                     uint32
	DRAInboundBytesNotCompressedWithinSiteSinceBoot                  uint32
	DRAInboundBytesTotalPersec                                       uint32
	DRAInboundBytesTotalSinceBoot                                    uint32
	DRAInboundFullSyncObjectsRemaining                               uint32
	DRAInboundLinkValueUpdatesRemaininginPacket                      uint32
	DRAInboundObjectsAppliedPersec                                   uint32
	DRAInboundObjectsFilteredPersec                                  uint32
	DRAInboundObjectsPersec                                          uint32
	DRAInboundObjectUpdatesRemaininginPacket                         uint32
	DRAInboundPropertiesAppliedPersec                                uint32
	DRAInboundPropertiesFilteredPersec                               uint32
	DRAInboundPropertiesTotalPersec                                  uint32
	DRAInboundTotalUpdatesRemaininginPacket                          uint32
	DRAInboundValuesDNsonlyPersec                                    uint32
	DRAInboundValuesTotalPersec                                      uint32
	DRAOutboundBytesCompressedBetweenSitesAfterCompressionPersec     uint32
	DRAOutboundBytesCompressedBetweenSitesAfterCompressionSinceBoot  uint32
	DRAOutboundBytesCompressedBetweenSitesBeforeCompressionPersec    uint32
	DRAOutboundBytesCompressedBetweenSitesBeforeCompressionSinceBoot uint32
	DRAOutboundBytesNotCompressedWithinSitePersec                    uint32
	DRAOutboundBytesNotCompressedWithinSiteSinceBoot                 uint32
	DRAOutboundBytesTotalPersec                                      uint32
	DRAOutboundBytesTotalSinceBoot                                   uint32
	DRAOutboundObjectsFilteredPersec                                 uint32
	DRAOutboundObjectsPersec                                         uint32
	DRAOutboundPropertiesPersec                                      uint32
	DRAOutboundValuesDNsonlyPersec                                   uint32
	DRAOutboundValuesTotalPersec                                     uint32
	DRAPendingReplicationOperations                                  uint32
	DRAPendingReplicationSynchronizations                            uint32
	DRASyncFailuresonSchemaMismatch                                  uint32
	DRASyncRequestsMade                                              uint32
	DRASyncRequestsSuccessful                                        uint32
	DRAThreadsGettingNCChanges                                       uint32
	DRAThreadsGettingNCChangesHoldingSemaphore                       uint32
	DSClientBindsPersec                                              uint32
	DSClientNameTranslationsPersec                                   uint32
	DSDirectoryReadsPersec                                           uint32
	DSDirectorySearchesPersec                                        uint32
	DSDirectoryWritesPersec                                          uint32
	DSMonitorListSize                                                uint32
	DSNameCachehitrate                                               uint32
	DSNameCachehitrate_Base                                          uint32
	DSNotifyQueueSize                                                uint32
	DSPercentReadsfromDRA                                            uint32
	DSPercentReadsfromKCC                                            uint32
	DSPercentReadsfromLSA                                            uint32
	DSPercentReadsfromNSPI                                           uint32
	DSPercentReadsfromNTDSAPI                                        uint32
	DSPercentReadsfromSAM                                            uint32
	DSPercentReadsOther                                              uint32
	DSPercentSearchesfromDRA                                         uint32
	DSPercentSearchesfromKCC                                         uint32
	DSPercentSearchesfromLDAP                                        uint32
	DSPercentSearchesfromLSA                                         uint32
	DSPercentSearchesfromNSPI                                        uint32
	DSPercentSearchesfromNTDSAPI                                     uint32
	DSPercentSearchesfromSAM                                         uint32
	DSPercentSearchesOther                                           uint32
	DSPercentWritesfromDRA                                           uint32
	DSPercentWritesfromKCC                                           uint32
	DSPercentWritesfromLDAP                                          uint32
	DSPercentWritesfromLSA                                           uint32
	DSPercentWritesfromNSPI                                          uint32
	DSPercentWritesfromNTDSAPI                                       uint32
	DSPercentWritesfromSAM                                           uint32
	DSPercentWritesOther                                             uint32
	DSSearchsuboperationsPersec                                      uint32
	DSSecurityDescriptorPropagationsEvents                           uint32
	DSSecurityDescriptorPropagatorAverageExclusionTime               uint32
	DSSecurityDescriptorPropagatorRuntimeQueue                       uint32
	DSSecurityDescriptorsuboperationsPersec                          uint32
	DSServerBindsPersec                                              uint32
	DSServerNameTranslationsPersec                                   uint32
	DSThreadsinUse                                                   uint32
	ExternalBindsPersec                                              uint32
	FastBindsPersec                                                  uint32
	LDAPActiveThreads                                                uint32
	LDAPBindTime                                                     uint32
	LDAPClientSessions                                               uint32
	LDAPClosedConnectionsPersec                                      uint32
	LDAPNewConnectionsPersec                                         uint32
	LDAPNewSSLConnectionsPersec                                      uint32
	LDAPSearchesPersec                                               uint32
	LDAPSuccessfulBindsPersec                                        uint32
	LDAPUDPoperationsPersec                                          uint32
	LDAPWritesPersec                                                 uint32
	LinkValuesCleanedPersec                                          uint32
	NegotiatedBindsPersec                                            uint32
	NTLMBindsPersec                                                  uint32
	OnelevelsearchesPersec                                           uint32
	PhantomsCleanedPersec                                            uint32
	PhantomsVisitedPersec                                            uint32
	SAMAccountGroupEvaluationLatency                                 uint32
	SAMDisplayInformationQueriesPersec                               uint32
	SAMDomainLocalGroupMembershipEvaluationsPersec                   uint32
	SAMEnumerationsPersec                                            uint32
	SAMGCEvaluationsPersec                                           uint32
	SAMGlobalGroupMembershipEvaluationsPersec                        uint32
	SAMMachineCreationAttemptsPersec                                 uint32
	SAMMembershipChangesPersec                                       uint32
	SAMNonTransitiveMembershipEvaluationsPersec                      uint32
	SAMPasswordChangesPersec                                         uint32
	SAMResourceGroupEvaluationLatency                                uint32
	SAMSuccessfulComputerCreationsPersecIncludesallrequests          uint32
	SAMSuccessfulUserCreationsPersec                                 uint32
	SAMTransitiveMembershipEvaluationsPersec                         uint32
	SAMUniversalGroupMembershipEvaluationsPersec                     uint32
	SAMUserCreationAttemptsPersec                                    uint32
	SimpleBindsPersec                                                uint32
	SubtreesearchesPersec                                            uint32
	TombstonesGarbageCollectedPersec                                 uint32
	TombstonesVisitedPersec                                          uint32
	Transitiveoperationsmillisecondsrun                              uint32
	TransitiveoperationsPersec                                       uint32
	TransitivesuboperationsPersec                                    uint32
}

func (c *Collector) collect(logger log.Logger, ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_DirectoryServices_DirectoryServices
	q := wmi.QueryAll(&dst, logger)
	if err := wmi.Query(q, &dst); err != nil {
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

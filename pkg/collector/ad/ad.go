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

// A collector is a Prometheus collector for WMI Win32_PerfRawData_DirectoryServices_DirectoryServices metrics
type collector struct {
	logger log.Logger

	AddressBookOperationsTotal                          *prometheus.Desc
	AddressBookClientSessions                           *prometheus.Desc
	ApproximateHighestDistinguishedNameTag              *prometheus.Desc
	AtqEstimatedDelaySeconds                            *prometheus.Desc
	AtqOutstandingRequests                              *prometheus.Desc
	AtqAverageRequestLatency                            *prometheus.Desc
	AtqCurrentThreads                                   *prometheus.Desc
	SearchesTotal                                       *prometheus.Desc
	DatabaseOperationsTotal                             *prometheus.Desc
	BindsTotal                                          *prometheus.Desc
	ReplicationHighestUsn                               *prometheus.Desc
	IntersiteReplicationDataBytesTotal                  *prometheus.Desc
	IntrasiteReplicationDataBytesTotal                  *prometheus.Desc
	ReplicationInboundSyncObjectsRemaining              *prometheus.Desc
	ReplicationInboundLinkValueUpdatesRemaining         *prometheus.Desc
	ReplicationInboundObjectsUpdatedTotal               *prometheus.Desc
	ReplicationInboundObjectsFilteredTotal              *prometheus.Desc
	ReplicationInboundPropertiesUpdatedTotal            *prometheus.Desc
	ReplicationInboundPropertiesFilteredTotal           *prometheus.Desc
	ReplicationPendingOperations                        *prometheus.Desc
	ReplicationPendingSynchronizations                  *prometheus.Desc
	ReplicationSyncRequestsTotal                        *prometheus.Desc
	ReplicationSyncRequestsSuccessTotal                 *prometheus.Desc
	ReplicationSyncRequestsSchemaMismatchFailureTotal   *prometheus.Desc
	DirectoryOperationsTotal                            *prometheus.Desc
	NameTranslationsTotal                               *prometheus.Desc
	ChangeMonitorsRegistered                            *prometheus.Desc
	ChangeMonitorUpdatesPending                         *prometheus.Desc
	NameCacheHitsTotal                                  *prometheus.Desc
	NameCacheLookupsTotal                               *prometheus.Desc
	DirectorySearchSuboperationsTotal                   *prometheus.Desc
	SecurityDescriptorPropagationEventsTotal            *prometheus.Desc
	SecurityDescriptorPropagationEventsQueued           *prometheus.Desc
	SecurityDescriptorPropagationAccessWaitTotalSeconds *prometheus.Desc
	SecurityDescriptorPropagationItemsQueuedTotal       *prometheus.Desc
	DirectoryServiceThreads                             *prometheus.Desc
	LdapClosedConnectionsTotal                          *prometheus.Desc
	LdapOpenedConnectionsTotal                          *prometheus.Desc
	LdapActiveThreads                                   *prometheus.Desc
	LdapLastBindTimeSeconds                             *prometheus.Desc
	LdapSearchesTotal                                   *prometheus.Desc
	LdapUdpOperationsTotal                              *prometheus.Desc
	LdapWritesTotal                                     *prometheus.Desc
	LdapClientSessions                                  *prometheus.Desc
	LinkValuesCleanedTotal                              *prometheus.Desc
	PhantomObjectsCleanedTotal                          *prometheus.Desc
	PhantomObjectsVisitedTotal                          *prometheus.Desc
	SamGroupMembershipEvaluationsTotal                  *prometheus.Desc
	SamGroupMembershipGlobalCatalogEvaluationsTotal     *prometheus.Desc
	SamGroupMembershipEvaluationsNontransitiveTotal     *prometheus.Desc
	SamGroupMembershipEvaluationsTransitiveTotal        *prometheus.Desc
	SamGroupEvaluationLatency                           *prometheus.Desc
	SamComputerCreationRequestsTotal                    *prometheus.Desc
	SamComputerCreationSuccessfulRequestsTotal          *prometheus.Desc
	SamUserCreationRequestsTotal                        *prometheus.Desc
	SamUserCreationSuccessfulRequestsTotal              *prometheus.Desc
	SamQueryDisplayRequestsTotal                        *prometheus.Desc
	SamEnumerationsTotal                                *prometheus.Desc
	SamMembershipChangesTotal                           *prometheus.Desc
	SamPasswordChangesTotal                             *prometheus.Desc
	TombstonedObjectsCollectedTotal                     *prometheus.Desc
	TombstonedObjectsVisitedTotal                       *prometheus.Desc
}

func New(logger log.Logger, _ *Config) types.Collector {
	c := &collector{}
	c.SetLogger(logger)
	return c
}

func NewWithFlags(_ *kingpin.Application) types.Collector {
	return &collector{}
}

func (c *collector) GetName() string {
	return Name
}

func (c *collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *collector) GetPerfCounter() ([]string, error) {
	return []string{}, nil
}

func (c *collector) Build() error {
	c.AddressBookOperationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "address_book_operations_total"),
		"",
		[]string{"operation"},
		nil,
	)
	c.AddressBookClientSessions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "address_book_client_sessions"),
		"",
		nil,
		nil,
	)
	c.ApproximateHighestDistinguishedNameTag = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "approximate_highest_distinguished_name_tag"),
		"",
		nil,
		nil,
	)
	c.AtqEstimatedDelaySeconds = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "atq_estimated_delay_seconds"),
		"",
		nil,
		nil,
	)
	c.AtqOutstandingRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "atq_outstanding_requests"),
		"",
		nil,
		nil,
	)
	c.AtqAverageRequestLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "atq_average_request_latency"),
		"",
		nil,
		nil,
	)
	c.AtqCurrentThreads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "atq_current_threads"),
		"",
		[]string{"service"},
		nil,
	)
	c.SearchesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "searches_total"),
		"",
		[]string{"scope"},
		nil,
	)
	c.DatabaseOperationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "database_operations_total"),
		"",
		[]string{"operation"},
		nil,
	)
	c.BindsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "binds_total"),
		"",
		[]string{"bind_method"},
		nil,
	)
	c.ReplicationHighestUsn = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_highest_usn"),
		"",
		[]string{"state"},
		nil,
	)
	c.IntrasiteReplicationDataBytesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_data_intrasite_bytes_total"),
		"",
		[]string{"direction"},
		nil,
	)
	c.IntersiteReplicationDataBytesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_data_intersite_bytes_total"),
		"",
		[]string{"direction"},
		nil,
	)
	c.ReplicationInboundSyncObjectsRemaining = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_inbound_sync_objects_remaining"),
		"",
		nil,
		nil,
	)
	c.ReplicationInboundLinkValueUpdatesRemaining = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_inbound_link_value_updates_remaining"),
		"",
		nil,
		nil,
	)
	c.ReplicationInboundObjectsUpdatedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_inbound_objects_updated_total"),
		"",
		nil,
		nil,
	)
	c.ReplicationInboundObjectsFilteredTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_inbound_objects_filtered_total"),
		"",
		nil,
		nil,
	)
	c.ReplicationInboundPropertiesUpdatedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_inbound_properties_updated_total"),
		"",
		nil,
		nil,
	)
	c.ReplicationInboundPropertiesFilteredTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_inbound_properties_filtered_total"),
		"",
		nil,
		nil,
	)
	c.ReplicationPendingOperations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_pending_operations"),
		"",
		nil,
		nil,
	)
	c.ReplicationPendingSynchronizations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_pending_synchronizations"),
		"",
		nil,
		nil,
	)
	c.ReplicationSyncRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_sync_requests_total"),
		"",
		nil,
		nil,
	)
	c.ReplicationSyncRequestsSuccessTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_sync_requests_success_total"),
		"",
		nil,
		nil,
	)
	c.ReplicationSyncRequestsSchemaMismatchFailureTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_sync_requests_schema_mismatch_failure_total"),
		"",
		nil,
		nil,
	)
	c.NameTranslationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "name_translations_total"),
		"",
		[]string{"target_name"},
		nil,
	)
	c.ChangeMonitorsRegistered = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "change_monitors_registered"),
		"",
		nil,
		nil,
	)
	c.ChangeMonitorUpdatesPending = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "change_monitor_updates_pending"),
		"",
		nil,
		nil,
	)
	c.NameCacheHitsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "name_cache_hits_total"),
		"",
		nil,
		nil,
	)
	c.NameCacheLookupsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "name_cache_lookups_total"),
		"",
		nil,
		nil,
	)
	c.DirectoryOperationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "directory_operations_total"),
		"",
		[]string{"operation", "origin"},
		nil,
	)
	c.DirectorySearchSuboperationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "directory_search_suboperations_total"),
		"",
		nil,
		nil,
	)
	c.SecurityDescriptorPropagationEventsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "security_descriptor_propagation_events_total"),
		"",
		nil,
		nil,
	)
	c.SecurityDescriptorPropagationEventsQueued = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "security_descriptor_propagation_events_queued"),
		"",
		nil,
		nil,
	)
	c.SecurityDescriptorPropagationAccessWaitTotalSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "security_descriptor_propagation_access_wait_total_seconds"),
		"",
		nil,
		nil,
	)
	c.SecurityDescriptorPropagationItemsQueuedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "security_descriptor_propagation_items_queued_total"),
		"",
		nil,
		nil,
	)
	c.DirectoryServiceThreads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "directory_service_threads"),
		"",
		nil,
		nil,
	)
	c.LdapClosedConnectionsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ldap_closed_connections_total"),
		"",
		nil,
		nil,
	)
	c.LdapOpenedConnectionsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ldap_opened_connections_total"),
		"",
		[]string{"type"},
		nil,
	)
	c.LdapActiveThreads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ldap_active_threads"),
		"",
		nil,
		nil,
	)
	c.LdapLastBindTimeSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ldap_last_bind_time_seconds"),
		"",
		nil,
		nil,
	)
	c.LdapSearchesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ldap_searches_total"),
		"",
		nil,
		nil,
	)
	c.LdapUdpOperationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ldap_udp_operations_total"),
		"",
		nil,
		nil,
	)
	c.LdapWritesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ldap_writes_total"),
		"",
		nil,
		nil,
	)
	c.LdapClientSessions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ldap_client_sessions"),
		"This is the number of sessions opened by LDAP clients at the time the data is taken. This is helpful in determining LDAP client activity and if the DC is able to handle the load. Of course, spikes during normal periods of authentication — such as first thing in the morning — are not necessarily a problem, but long sustained periods of high values indicate an overworked DC.",
		nil,
		nil,
	)
	c.LinkValuesCleanedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "link_values_cleaned_total"),
		"",
		nil,
		nil,
	)
	c.PhantomObjectsCleanedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "phantom_objects_cleaned_total"),
		"",
		nil,
		nil,
	)
	c.PhantomObjectsVisitedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "phantom_objects_visited_total"),
		"",
		nil,
		nil,
	)
	c.SamGroupMembershipEvaluationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_group_membership_evaluations_total"),
		"",
		[]string{"group_type"},
		nil,
	)
	c.SamGroupMembershipGlobalCatalogEvaluationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_group_membership_global_catalog_evaluations_total"),
		"",
		nil,
		nil,
	)
	c.SamGroupMembershipEvaluationsNontransitiveTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_group_membership_evaluations_nontransitive_total"),
		"",
		nil,
		nil,
	)
	c.SamGroupMembershipEvaluationsTransitiveTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_group_membership_evaluations_transitive_total"),
		"",
		nil,
		nil,
	)
	c.SamGroupEvaluationLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_group_evaluation_latency"),
		"The mean latency of the last 100 group evaluations performed for authentication",
		[]string{"evaluation_type"},
		nil,
	)
	c.SamComputerCreationRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_computer_creation_requests_total"),
		"",
		nil,
		nil,
	)
	c.SamComputerCreationSuccessfulRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_computer_creation_successful_requests_total"),
		"",
		nil,
		nil,
	)
	c.SamUserCreationRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_user_creation_requests_total"),
		"",
		nil,
		nil,
	)
	c.SamUserCreationSuccessfulRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_user_creation_successful_requests_total"),
		"",
		nil,
		nil,
	)
	c.SamQueryDisplayRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_query_display_requests_total"),
		"",
		nil,
		nil,
	)
	c.SamEnumerationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_enumerations_total"),
		"",
		nil,
		nil,
	)
	c.SamMembershipChangesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_membership_changes_total"),
		"",
		nil,
		nil,
	)
	c.SamPasswordChangesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_password_changes_total"),
		"",
		nil,
		nil,
	)

	c.TombstonedObjectsCollectedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "tombstoned_objects_collected_total"),
		"",
		nil,
		nil,
	)
	c.TombstonedObjectsVisitedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "tombstoned_objects_visited_total"),
		"",
		nil,
		nil,
	)

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.collect(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting ad metrics", "err", err)
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

func (c *collector) collect(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_DirectoryServices_DirectoryServices
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}
	if len(dst) == 0 {
		return errors.New("WMI query returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.AddressBookOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].ABANRPersec),
		"ambiguous_name_resolution",
	)
	ch <- prometheus.MustNewConstMetric(
		c.AddressBookOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].ABBrowsesPersec),
		"browse",
	)
	ch <- prometheus.MustNewConstMetric(
		c.AddressBookOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].ABMatchesPersec),
		"find",
	)
	ch <- prometheus.MustNewConstMetric(
		c.AddressBookOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].ABPropertyReadsPersec),
		"property_read",
	)
	ch <- prometheus.MustNewConstMetric(
		c.AddressBookOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].ABSearchesPersec),
		"search",
	)
	ch <- prometheus.MustNewConstMetric(
		c.AddressBookOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].ABProxyLookupsPersec),
		"proxy_search",
	)

	ch <- prometheus.MustNewConstMetric(
		c.AddressBookClientSessions,
		prometheus.GaugeValue,
		float64(dst[0].ABClientSessions),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ApproximateHighestDistinguishedNameTag,
		prometheus.GaugeValue,
		float64(dst[0].ApproximatehighestDNT),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AtqEstimatedDelaySeconds,
		prometheus.GaugeValue,
		float64(dst[0].ATQEstimatedQueueDelay)/1000,
	)
	ch <- prometheus.MustNewConstMetric(
		c.AtqOutstandingRequests,
		prometheus.GaugeValue,
		float64(dst[0].ATQOutstandingQueuedRequests),
	)
	ch <- prometheus.MustNewConstMetric(
		c.AtqAverageRequestLatency,
		prometheus.GaugeValue,
		float64(dst[0].ATQRequestLatency),
	)
	ch <- prometheus.MustNewConstMetric(
		c.AtqCurrentThreads,
		prometheus.GaugeValue,
		float64(dst[0].ATQThreadsLDAP),
		"ldap",
	)
	ch <- prometheus.MustNewConstMetric(
		c.AtqCurrentThreads,
		prometheus.GaugeValue,
		float64(dst[0].ATQThreadsOther),
		"other",
	)

	ch <- prometheus.MustNewConstMetric(
		c.SearchesTotal,
		prometheus.CounterValue,
		float64(dst[0].BasesearchesPersec),
		"base",
	)
	ch <- prometheus.MustNewConstMetric(
		c.SearchesTotal,
		prometheus.CounterValue,
		float64(dst[0].SubtreesearchesPersec),
		"subtree",
	)
	ch <- prometheus.MustNewConstMetric(
		c.SearchesTotal,
		prometheus.CounterValue,
		float64(dst[0].OnelevelsearchesPersec),
		"one_level",
	)

	ch <- prometheus.MustNewConstMetric(
		c.DatabaseOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DatabaseaddsPersec),
		"add",
	)
	ch <- prometheus.MustNewConstMetric(
		c.DatabaseOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DatabasedeletesPersec),
		"delete",
	)
	ch <- prometheus.MustNewConstMetric(
		c.DatabaseOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DatabasemodifysPersec),
		"modify",
	)
	ch <- prometheus.MustNewConstMetric(
		c.DatabaseOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DatabaserecyclesPersec),
		"recycle",
	)

	ch <- prometheus.MustNewConstMetric(
		c.BindsTotal,
		prometheus.CounterValue,
		float64(dst[0].DigestBindsPersec),
		"digest",
	)
	ch <- prometheus.MustNewConstMetric(
		c.BindsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSClientBindsPersec),
		"ds_client",
	)
	ch <- prometheus.MustNewConstMetric(
		c.BindsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSServerBindsPersec),
		"ds_server",
	)
	ch <- prometheus.MustNewConstMetric(
		c.BindsTotal,
		prometheus.CounterValue,
		float64(dst[0].ExternalBindsPersec),
		"external",
	)
	ch <- prometheus.MustNewConstMetric(
		c.BindsTotal,
		prometheus.CounterValue,
		float64(dst[0].FastBindsPersec),
		"fast",
	)
	ch <- prometheus.MustNewConstMetric(
		c.BindsTotal,
		prometheus.CounterValue,
		float64(dst[0].NegotiatedBindsPersec),
		"negotiate",
	)
	ch <- prometheus.MustNewConstMetric(
		c.BindsTotal,
		prometheus.CounterValue,
		float64(dst[0].NTLMBindsPersec),
		"ntlm",
	)
	ch <- prometheus.MustNewConstMetric(
		c.BindsTotal,
		prometheus.CounterValue,
		float64(dst[0].SimpleBindsPersec),
		"simple",
	)
	ch <- prometheus.MustNewConstMetric(
		c.BindsTotal,
		prometheus.CounterValue,
		float64(dst[0].LDAPSuccessfulBindsPersec),
		"ldap",
	)

	ch <- prometheus.MustNewConstMetric(
		c.ReplicationHighestUsn,
		prometheus.CounterValue,
		float64(dst[0].DRAHighestUSNCommittedHighpart<<32)+float64(dst[0].DRAHighestUSNCommittedLowpart),
		"committed",
	)
	ch <- prometheus.MustNewConstMetric(
		c.ReplicationHighestUsn,
		prometheus.CounterValue,
		float64(dst[0].DRAHighestUSNIssuedHighpart<<32)+float64(dst[0].DRAHighestUSNIssuedLowpart),
		"issued",
	)

	ch <- prometheus.MustNewConstMetric(
		c.IntersiteReplicationDataBytesTotal,
		prometheus.CounterValue,
		float64(dst[0].DRAInboundBytesCompressedBetweenSitesAfterCompressionPersec),
		"inbound",
	)
	// The pre-compression data size seems to have little value? Skipping for now
	// ch <- prometheus.MustNewConstMetric(
	// 	c.IntersiteReplicationDataBytesTotal,
	// 	prometheus.CounterValue,
	// 	float64(dst[0].DRAInboundBytesCompressedBetweenSitesBeforeCompressionPersec),
	// 	"inbound",
	// )
	ch <- prometheus.MustNewConstMetric(
		c.IntersiteReplicationDataBytesTotal,
		prometheus.CounterValue,
		float64(dst[0].DRAOutboundBytesCompressedBetweenSitesAfterCompressionPersec),
		"outbound",
	)
	// ch <- prometheus.MustNewConstMetric(
	// 	c.IntersiteReplicationDataBytesTotal,
	// 	prometheus.CounterValue,
	// 	float64(dst[0].DRAOutboundBytesCompressedBetweenSitesBeforeCompressionPersec),
	// 	"outbound",
	// )
	ch <- prometheus.MustNewConstMetric(
		c.IntrasiteReplicationDataBytesTotal,
		prometheus.CounterValue,
		float64(dst[0].DRAInboundBytesNotCompressedWithinSitePersec),
		"inbound",
	)
	ch <- prometheus.MustNewConstMetric(
		c.IntrasiteReplicationDataBytesTotal,
		prometheus.CounterValue,
		float64(dst[0].DRAOutboundBytesNotCompressedWithinSitePersec),
		"outbound",
	)

	ch <- prometheus.MustNewConstMetric(
		c.ReplicationInboundSyncObjectsRemaining,
		prometheus.GaugeValue,
		float64(dst[0].DRAInboundFullSyncObjectsRemaining),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ReplicationInboundLinkValueUpdatesRemaining,
		prometheus.GaugeValue,
		float64(dst[0].DRAInboundLinkValueUpdatesRemaininginPacket),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ReplicationInboundObjectsUpdatedTotal,
		prometheus.CounterValue,
		float64(dst[0].DRAInboundObjectsAppliedPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.ReplicationInboundObjectsFilteredTotal,
		prometheus.CounterValue,
		float64(dst[0].DRAInboundObjectsFilteredPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ReplicationInboundPropertiesUpdatedTotal,
		prometheus.CounterValue,
		float64(dst[0].DRAInboundPropertiesAppliedPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.ReplicationInboundPropertiesFilteredTotal,
		prometheus.CounterValue,
		float64(dst[0].DRAInboundPropertiesFilteredPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ReplicationPendingOperations,
		prometheus.GaugeValue,
		float64(dst[0].DRAPendingReplicationOperations),
	)
	ch <- prometheus.MustNewConstMetric(
		c.ReplicationPendingSynchronizations,
		prometheus.GaugeValue,
		float64(dst[0].DRAPendingReplicationSynchronizations),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ReplicationSyncRequestsTotal,
		prometheus.CounterValue,
		float64(dst[0].DRASyncRequestsMade),
	)
	ch <- prometheus.MustNewConstMetric(
		c.ReplicationSyncRequestsSuccessTotal,
		prometheus.CounterValue,
		float64(dst[0].DRASyncRequestsSuccessful),
	)
	ch <- prometheus.MustNewConstMetric(
		c.ReplicationSyncRequestsSchemaMismatchFailureTotal,
		prometheus.CounterValue,
		float64(dst[0].DRASyncFailuresonSchemaMismatch),
	)

	ch <- prometheus.MustNewConstMetric(
		c.NameTranslationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSClientNameTranslationsPersec),
		"client",
	)
	ch <- prometheus.MustNewConstMetric(
		c.NameTranslationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSServerNameTranslationsPersec),
		"server",
	)

	ch <- prometheus.MustNewConstMetric(
		c.ChangeMonitorsRegistered,
		prometheus.GaugeValue,
		float64(dst[0].DSMonitorListSize),
	)
	ch <- prometheus.MustNewConstMetric(
		c.ChangeMonitorUpdatesPending,
		prometheus.GaugeValue,
		float64(dst[0].DSNotifyQueueSize),
	)

	ch <- prometheus.MustNewConstMetric(
		c.NameCacheHitsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSNameCachehitrate),
	)
	ch <- prometheus.MustNewConstMetric(
		c.NameCacheLookupsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSNameCachehitrate_Base),
	)

	ch <- prometheus.MustNewConstMetric(
		c.DirectoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentReadsfromDRA),
		"read",
		"replication_agent",
	)
	ch <- prometheus.MustNewConstMetric(
		c.DirectoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentReadsfromKCC),
		"read",
		"knowledge_consistency_checker",
	)
	ch <- prometheus.MustNewConstMetric(
		c.DirectoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentReadsfromLSA),
		"read",
		"local_security_authority",
	)
	ch <- prometheus.MustNewConstMetric(
		c.DirectoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentReadsfromNSPI),
		"read",
		"name_service_provider_interface",
	)
	ch <- prometheus.MustNewConstMetric(
		c.DirectoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentReadsfromNTDSAPI),
		"read",
		"directory_service_api",
	)
	ch <- prometheus.MustNewConstMetric(
		c.DirectoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentReadsfromSAM),
		"read",
		"security_account_manager",
	)
	ch <- prometheus.MustNewConstMetric(
		c.DirectoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentReadsOther),
		"read",
		"other",
	)
	ch <- prometheus.MustNewConstMetric(
		c.DirectoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentSearchesfromDRA),
		"search",
		"replication_agent",
	)
	ch <- prometheus.MustNewConstMetric(
		c.DirectoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentSearchesfromKCC),
		"search",
		"knowledge_consistency_checker",
	)
	ch <- prometheus.MustNewConstMetric(
		c.DirectoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentSearchesfromLDAP),
		"search",
		"ldap",
	)
	ch <- prometheus.MustNewConstMetric(
		c.DirectoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentSearchesfromLSA),
		"search",
		"local_security_authority",
	)
	ch <- prometheus.MustNewConstMetric(
		c.DirectoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentSearchesfromNSPI),
		"search",
		"name_service_provider_interface",
	)
	ch <- prometheus.MustNewConstMetric(
		c.DirectoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentSearchesfromNTDSAPI),
		"search",
		"directory_service_api",
	)
	ch <- prometheus.MustNewConstMetric(
		c.DirectoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentSearchesfromSAM),
		"search",
		"security_account_manager",
	)
	ch <- prometheus.MustNewConstMetric(
		c.DirectoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentSearchesOther),
		"search",
		"other",
	)
	ch <- prometheus.MustNewConstMetric(
		c.DirectoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentWritesfromDRA),
		"write",
		"replication_agent",
	)
	ch <- prometheus.MustNewConstMetric(
		c.DirectoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentWritesfromKCC),
		"write",
		"knowledge_consistency_checker",
	)
	ch <- prometheus.MustNewConstMetric(
		c.DirectoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentWritesfromLDAP),
		"write",
		"ldap",
	)
	ch <- prometheus.MustNewConstMetric(
		c.DirectoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentWritesfromLSA),
		"write",
		"local_security_authority",
	)
	ch <- prometheus.MustNewConstMetric(
		c.DirectoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentWritesfromNSPI),
		"write",
		"name_service_provider_interface",
	)
	ch <- prometheus.MustNewConstMetric(
		c.DirectoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentWritesfromNTDSAPI),
		"write",
		"directory_service_api",
	)
	ch <- prometheus.MustNewConstMetric(
		c.DirectoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentWritesfromSAM),
		"write",
		"security_account_manager",
	)
	ch <- prometheus.MustNewConstMetric(
		c.DirectoryOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSPercentWritesOther),
		"write",
		"other",
	)

	ch <- prometheus.MustNewConstMetric(
		c.DirectorySearchSuboperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSSearchsuboperationsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SecurityDescriptorPropagationEventsTotal,
		prometheus.CounterValue,
		float64(dst[0].DSSecurityDescriptorsuboperationsPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.SecurityDescriptorPropagationEventsQueued,
		prometheus.GaugeValue,
		float64(dst[0].DSSecurityDescriptorPropagationsEvents),
	)
	ch <- prometheus.MustNewConstMetric(
		c.SecurityDescriptorPropagationAccessWaitTotalSeconds,
		prometheus.GaugeValue,
		float64(dst[0].DSSecurityDescriptorPropagatorAverageExclusionTime),
	)
	ch <- prometheus.MustNewConstMetric(
		c.SecurityDescriptorPropagationItemsQueuedTotal,
		prometheus.CounterValue,
		float64(dst[0].DSSecurityDescriptorPropagatorRuntimeQueue),
	)

	ch <- prometheus.MustNewConstMetric(
		c.DirectoryServiceThreads,
		prometheus.GaugeValue,
		float64(dst[0].DSThreadsinUse),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LdapClosedConnectionsTotal,
		prometheus.CounterValue,
		float64(dst[0].LDAPClosedConnectionsPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.LdapOpenedConnectionsTotal,
		prometheus.CounterValue,
		float64(dst[0].LDAPNewConnectionsPersec),
		"ldap",
	)
	ch <- prometheus.MustNewConstMetric(
		c.LdapOpenedConnectionsTotal,
		prometheus.CounterValue,
		float64(dst[0].LDAPNewSSLConnectionsPersec),
		"ldaps",
	)

	ch <- prometheus.MustNewConstMetric(
		c.LdapActiveThreads,
		prometheus.GaugeValue,
		float64(dst[0].LDAPActiveThreads),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LdapLastBindTimeSeconds,
		prometheus.GaugeValue,
		float64(dst[0].LDAPBindTime)/1000,
	)

	ch <- prometheus.MustNewConstMetric(
		c.LdapSearchesTotal,
		prometheus.CounterValue,
		float64(dst[0].LDAPSearchesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LdapUdpOperationsTotal,
		prometheus.CounterValue,
		float64(dst[0].LDAPUDPoperationsPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.LdapWritesTotal,
		prometheus.CounterValue,
		float64(dst[0].LDAPWritesPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.LdapClientSessions,
		prometheus.GaugeValue,
		float64(dst[0].LDAPClientSessions),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LinkValuesCleanedTotal,
		prometheus.CounterValue,
		float64(dst[0].LinkValuesCleanedPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PhantomObjectsCleanedTotal,
		prometheus.CounterValue,
		float64(dst[0].PhantomsCleanedPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.PhantomObjectsVisitedTotal,
		prometheus.CounterValue,
		float64(dst[0].PhantomsVisitedPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SamGroupMembershipEvaluationsTotal,
		prometheus.CounterValue,
		float64(dst[0].SAMGlobalGroupMembershipEvaluationsPersec),
		"global",
	)
	ch <- prometheus.MustNewConstMetric(
		c.SamGroupMembershipEvaluationsTotal,
		prometheus.CounterValue,
		float64(dst[0].SAMDomainLocalGroupMembershipEvaluationsPersec),
		"domain_local",
	)
	ch <- prometheus.MustNewConstMetric(
		c.SamGroupMembershipEvaluationsTotal,
		prometheus.CounterValue,
		float64(dst[0].SAMUniversalGroupMembershipEvaluationsPersec),
		"universal",
	)
	ch <- prometheus.MustNewConstMetric(
		c.SamGroupMembershipGlobalCatalogEvaluationsTotal,
		prometheus.CounterValue,
		float64(dst[0].SAMGCEvaluationsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SamGroupMembershipEvaluationsNontransitiveTotal,
		prometheus.CounterValue,
		float64(dst[0].SAMNonTransitiveMembershipEvaluationsPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.SamGroupMembershipEvaluationsTransitiveTotal,
		prometheus.CounterValue,
		float64(dst[0].SAMTransitiveMembershipEvaluationsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SamGroupEvaluationLatency,
		prometheus.GaugeValue,
		float64(dst[0].SAMAccountGroupEvaluationLatency),
		"account_group",
	)
	ch <- prometheus.MustNewConstMetric(
		c.SamGroupEvaluationLatency,
		prometheus.GaugeValue,
		float64(dst[0].SAMResourceGroupEvaluationLatency),
		"resource_group",
	)

	ch <- prometheus.MustNewConstMetric(
		c.SamComputerCreationRequestsTotal,
		prometheus.CounterValue,
		float64(dst[0].SAMSuccessfulComputerCreationsPersecIncludesallrequests),
	)
	ch <- prometheus.MustNewConstMetric(
		c.SamComputerCreationSuccessfulRequestsTotal,
		prometheus.CounterValue,
		float64(dst[0].SAMMachineCreationAttemptsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SamUserCreationRequestsTotal,
		prometheus.CounterValue,
		float64(dst[0].SAMUserCreationAttemptsPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.SamUserCreationSuccessfulRequestsTotal,
		prometheus.CounterValue,
		float64(dst[0].SAMSuccessfulUserCreationsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SamQueryDisplayRequestsTotal,
		prometheus.CounterValue,
		float64(dst[0].SAMDisplayInformationQueriesPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.SamEnumerationsTotal,
		prometheus.CounterValue,
		float64(dst[0].SAMEnumerationsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SamMembershipChangesTotal,
		prometheus.CounterValue,
		float64(dst[0].SAMMembershipChangesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SamPasswordChangesTotal,
		prometheus.CounterValue,
		float64(dst[0].SAMPasswordChangesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.TombstonedObjectsCollectedTotal,
		prometheus.CounterValue,
		float64(dst[0].TombstonesGarbageCollectedPersec),
	)
	ch <- prometheus.MustNewConstMetric(
		c.TombstonedObjectsVisitedTotal,
		prometheus.CounterValue,
		float64(dst[0].TombstonesVisitedPersec),
	)

	return nil
}

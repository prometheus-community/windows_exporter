// Copyright 2024 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows

package ad

import (
	"fmt"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "ad"

type Config struct{}

//nolint:gochecknoglobals
var ConfigDefaults = Config{}

type Collector struct {
	config Config

	perfDataCollector *pdh.Collector
	perfDataObject    []perfDataCounterValues

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

	var err error

	c.perfDataCollector, err = pdh.NewCollector[perfDataCounterValues](pdh.CounterTypeRaw, "DirectoryServices", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create DirectoryServices collector: %w", err)
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollector.Collect(&c.perfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect DirectoryServices (AD) metrics: %w", err)
	}

	ch <- prometheus.MustNewConstMetric(
		c.addressBookOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].AbANRPerSec,
		"ambiguous_name_resolution",
	)
	ch <- prometheus.MustNewConstMetric(
		c.addressBookOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].AbBrowsesPerSec,
		"browse",
	)
	ch <- prometheus.MustNewConstMetric(
		c.addressBookOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].AbMatchesPerSec,
		"find",
	)
	ch <- prometheus.MustNewConstMetric(
		c.addressBookOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].AbPropertyReadsPerSec,
		"property_read",
	)
	ch <- prometheus.MustNewConstMetric(
		c.addressBookOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].AbSearchesPerSec,
		"search",
	)
	ch <- prometheus.MustNewConstMetric(
		c.addressBookOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].AbProxyLookupsPerSec,
		"proxy_search",
	)

	ch <- prometheus.MustNewConstMetric(
		c.addressBookClientSessions,
		prometheus.GaugeValue,
		c.perfDataObject[0].AbClientSessions,
	)

	ch <- prometheus.MustNewConstMetric(
		c.approximateHighestDistinguishedNameTag,
		prometheus.GaugeValue,
		c.perfDataObject[0].ApproximateHighestDNT,
	)

	ch <- prometheus.MustNewConstMetric(
		c.atqEstimatedDelaySeconds,
		prometheus.GaugeValue,
		c.perfDataObject[0].AtqEstimatedQueueDelay/1000,
	)
	ch <- prometheus.MustNewConstMetric(
		c.atqOutstandingRequests,
		prometheus.GaugeValue,
		c.perfDataObject[0].AtqOutstandingQueuedRequests,
	)
	ch <- prometheus.MustNewConstMetric(
		c.atqAverageRequestLatency,
		prometheus.GaugeValue,
		c.perfDataObject[0].AtqRequestLatency,
	)
	ch <- prometheus.MustNewConstMetric(
		c.atqCurrentThreads,
		prometheus.GaugeValue,
		c.perfDataObject[0].AtqThreadsLDAP,
		"ldap",
	)
	ch <- prometheus.MustNewConstMetric(
		c.atqCurrentThreads,
		prometheus.GaugeValue,
		c.perfDataObject[0].AtqThreadsOther,
		"other",
	)

	ch <- prometheus.MustNewConstMetric(
		c.searchesTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].BaseSearchesPerSec,
		"base",
	)
	ch <- prometheus.MustNewConstMetric(
		c.searchesTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].SubtreeSearchesPerSec,
		"subtree",
	)
	ch <- prometheus.MustNewConstMetric(
		c.searchesTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].OneLevelSearchesPerSec,
		"one_level",
	)

	ch <- prometheus.MustNewConstMetric(
		c.databaseOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DatabaseAddsPerSec,
		"add",
	)
	ch <- prometheus.MustNewConstMetric(
		c.databaseOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DatabaseDeletesPerSec,
		"delete",
	)
	ch <- prometheus.MustNewConstMetric(
		c.databaseOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DatabaseModifiesPerSec,
		"modify",
	)
	ch <- prometheus.MustNewConstMetric(
		c.databaseOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DatabaseRecyclesPerSec,
		"recycle",
	)

	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DigestBindsPerSec,
		"digest",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsClientBindsPerSec,
		"ds_client",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsServerBindsPerSec,
		"ds_server",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].ExternalBindsPerSec,
		"external",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].FastBindsPerSec,
		"fast",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].NegotiatedBindsPerSec,
		"negotiate",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].NTLMBindsPerSec,
		"ntlm",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].SimpleBindsPerSec,
		"simple",
	)
	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].LdapSuccessfulBindsPerSec,
		"ldap",
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationHighestUsn,
		prometheus.CounterValue,
		float64(uint64(c.perfDataObject[0].DRAHighestUSNCommittedHighPart)<<32)+c.perfDataObject[0].DRAHighestUSNCommittedLowPart,
		"committed",
	)
	ch <- prometheus.MustNewConstMetric(
		c.replicationHighestUsn,
		prometheus.CounterValue,
		float64(uint64(c.perfDataObject[0].DRAHighestUSNIssuedHighPart)<<32)+c.perfDataObject[0].DRAHighestUSNIssuedLowPart,
		"issued",
	)

	ch <- prometheus.MustNewConstMetric(
		c.interSiteReplicationDataBytesTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DRAInboundBytesCompressedBetweenSitesAfterCompressionPerSec,
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
		c.perfDataObject[0].DRAOutboundBytesCompressedBetweenSitesAfterCompressionPerSec,
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
		c.perfDataObject[0].DRAInboundBytesNotCompressedWithinSitePerSec,
		"inbound",
	)
	ch <- prometheus.MustNewConstMetric(
		c.intraSiteReplicationDataBytesTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DRAOutboundBytesNotCompressedWithinSitePerSec,
		"outbound",
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationInboundSyncObjectsRemaining,
		prometheus.GaugeValue,
		c.perfDataObject[0].DRAInboundFullSyncObjectsRemaining,
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationInboundLinkValueUpdatesRemaining,
		prometheus.GaugeValue,
		c.perfDataObject[0].DRAInboundLinkValueUpdatesRemainingInPacket,
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationInboundObjectsUpdatedTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DRAInboundObjectsAppliedPerSec,
	)
	ch <- prometheus.MustNewConstMetric(
		c.replicationInboundObjectsFilteredTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DRAInboundObjectsFilteredPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationInboundPropertiesUpdatedTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DRAInboundPropertiesAppliedPerSec,
	)
	ch <- prometheus.MustNewConstMetric(
		c.replicationInboundPropertiesFilteredTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DRAInboundPropertiesFilteredPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationPendingOperations,
		prometheus.GaugeValue,
		c.perfDataObject[0].DRAPendingReplicationOperations,
	)
	ch <- prometheus.MustNewConstMetric(
		c.replicationPendingSynchronizations,
		prometheus.GaugeValue,
		c.perfDataObject[0].DRAPendingReplicationSynchronizations,
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationSyncRequestsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DRASyncRequestsMade,
	)
	ch <- prometheus.MustNewConstMetric(
		c.replicationSyncRequestsSuccessTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DRASyncRequestsSuccessful,
	)
	ch <- prometheus.MustNewConstMetric(
		c.replicationSyncRequestsSchemaMismatchFailureTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DRASyncFailuresOnSchemaMismatch,
	)

	ch <- prometheus.MustNewConstMetric(
		c.nameTranslationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsClientNameTranslationsPerSec,
		"client",
	)
	ch <- prometheus.MustNewConstMetric(
		c.nameTranslationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsServerNameTranslationsPerSec,
		"server",
	)

	ch <- prometheus.MustNewConstMetric(
		c.changeMonitorsRegistered,
		prometheus.GaugeValue,
		c.perfDataObject[0].DsMonitorListSize,
	)
	ch <- prometheus.MustNewConstMetric(
		c.changeMonitorUpdatesPending,
		prometheus.GaugeValue,
		c.perfDataObject[0].DsNotifyQueueSize,
	)

	ch <- prometheus.MustNewConstMetric(
		c.nameCacheHitsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsNameCacheHitRate,
	)
	ch <- prometheus.MustNewConstMetric(
		c.nameCacheLookupsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsNameCacheHitRateSecondValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsPercentReadsFromDRA,
		"read",
		"replication_agent",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsPercentReadsFromKCC,
		"read",
		"knowledge_consistency_checker",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsPercentReadsFromLSA,
		"read",
		"local_security_authority",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsPercentReadsFromNSPI,
		"read",
		"name_service_provider_interface",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsPercentReadsFromNTDSAPI,
		"read",
		"directory_service_api",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsPercentReadsFromSAM,
		"read",
		"security_account_manager",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsPercentReadsOther,
		"read",
		"other",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsPercentSearchesFromDRA,
		"search",
		"replication_agent",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsPercentSearchesFromKCC,
		"search",
		"knowledge_consistency_checker",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsPercentSearchesFromLDAP,
		"search",
		"ldap",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsPercentSearchesFromLSA,
		"search",
		"local_security_authority",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsPercentSearchesFromNSPI,
		"search",
		"name_service_provider_interface",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsPercentSearchesFromNTDSAPI,
		"search",
		"directory_service_api",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsPercentSearchesFromSAM,
		"search",
		"security_account_manager",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsPercentSearchesOther,
		"search",
		"other",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsPercentWritesFromDRA,
		"write",
		"replication_agent",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsPercentWritesFromKCC,
		"write",
		"knowledge_consistency_checker",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsPercentWritesFromLDAP,
		"write",
		"ldap",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsPercentSearchesFromLSA,
		"write",
		"local_security_authority",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsPercentWritesFromNSPI,
		"write",
		"name_service_provider_interface",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsPercentWritesFromNTDSAPI,
		"write",
		"directory_service_api",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsPercentWritesFromSAM,
		"write",
		"security_account_manager",
	)
	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsPercentWritesOther,
		"write",
		"other",
	)

	ch <- prometheus.MustNewConstMetric(
		c.directorySearchSubOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsSearchSubOperationsPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.securityDescriptorPropagationEventsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsSecurityDescriptorSubOperationsPerSec,
	)
	ch <- prometheus.MustNewConstMetric(
		c.securityDescriptorPropagationEventsQueued,
		prometheus.GaugeValue,
		c.perfDataObject[0].DsSecurityDescriptorPropagationsEvents,
	)
	ch <- prometheus.MustNewConstMetric(
		c.securityDescriptorPropagationAccessWaitTotalSeconds,
		prometheus.GaugeValue,
		c.perfDataObject[0].DsSecurityDescriptorPropagatorAverageExclusionTime,
	)
	ch <- prometheus.MustNewConstMetric(
		c.securityDescriptorPropagationItemsQueuedTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DsSecurityDescriptorPropagatorRuntimeQueue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryServiceThreads,
		prometheus.GaugeValue,
		c.perfDataObject[0].DsThreadsInUse,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ldapClosedConnectionsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].LdapClosedConnectionsPerSec,
	)
	ch <- prometheus.MustNewConstMetric(
		c.ldapOpenedConnectionsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].LdapNewConnectionsPerSec,
		"ldap",
	)
	ch <- prometheus.MustNewConstMetric(
		c.ldapOpenedConnectionsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].LdapNewSSLConnectionsPerSec,
		"ldaps",
	)

	ch <- prometheus.MustNewConstMetric(
		c.ldapActiveThreads,
		prometheus.GaugeValue,
		c.perfDataObject[0].LdapActiveThreads,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ldapLastBindTimeSeconds,
		prometheus.GaugeValue,
		c.perfDataObject[0].LdapBindTime/1000,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ldapSearchesTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].LdapSearchesPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ldapUdpOperationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].LdapUDPOperationsPerSec,
	)
	ch <- prometheus.MustNewConstMetric(
		c.ldapWritesTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].LdapWritesPerSec,
	)
	ch <- prometheus.MustNewConstMetric(
		c.ldapClientSessions,
		prometheus.GaugeValue,
		c.perfDataObject[0].LdapClientSessions,
	)

	ch <- prometheus.MustNewConstMetric(
		c.linkValuesCleanedTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].LinkValuesCleanedPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.phantomObjectsCleanedTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].PhantomsCleanedPerSec,
	)
	ch <- prometheus.MustNewConstMetric(
		c.phantomObjectsVisitedTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].PhantomsVisitedPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samGroupMembershipEvaluationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].SamGlobalGroupMembershipEvaluationsPerSec,
		"global",
	)
	ch <- prometheus.MustNewConstMetric(
		c.samGroupMembershipEvaluationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].SamDomainLocalGroupMembershipEvaluationsPerSec,
		"domain_local",
	)
	ch <- prometheus.MustNewConstMetric(
		c.samGroupMembershipEvaluationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].SamUniversalGroupMembershipEvaluationsPerSec,
		"universal",
	)
	ch <- prometheus.MustNewConstMetric(
		c.samGroupMembershipGlobalCatalogEvaluationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].SamGCEvaluationsPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samGroupMembershipEvaluationsNonTransitiveTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].SamNonTransitiveMembershipEvaluationsPerSec,
	)
	ch <- prometheus.MustNewConstMetric(
		c.samGroupMembershipEvaluationsTransitiveTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].SamTransitiveMembershipEvaluationsPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samGroupEvaluationLatency,
		prometheus.GaugeValue,
		c.perfDataObject[0].SamAccountGroupEvaluationLatency,
		"account_group",
	)
	ch <- prometheus.MustNewConstMetric(
		c.samGroupEvaluationLatency,
		prometheus.GaugeValue,
		c.perfDataObject[0].SamResourceGroupEvaluationLatency,
		"resource_group",
	)

	ch <- prometheus.MustNewConstMetric(
		c.samComputerCreationRequestsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].SamSuccessfulComputerCreationsPerSecIncludesAllRequests,
	)
	ch <- prometheus.MustNewConstMetric(
		c.samComputerCreationSuccessfulRequestsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].SamMachineCreationAttemptsPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samUserCreationRequestsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].SamUserCreationAttemptsPerSec,
	)
	ch <- prometheus.MustNewConstMetric(
		c.samUserCreationSuccessfulRequestsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].SamSuccessfulUserCreationsPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samQueryDisplayRequestsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].SamDisplayInformationQueriesPerSec,
	)
	ch <- prometheus.MustNewConstMetric(
		c.samEnumerationsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].SamEnumerationsPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samMembershipChangesTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].SamMembershipChangesPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samPasswordChangesTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].SamPasswordChangesPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.tombstonesObjectsCollectedTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].TombstonesGarbageCollectedPerSec,
	)
	ch <- prometheus.MustNewConstMetric(
		c.tombstonesObjectsVisitedTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].TombstonesVisitedPerSec,
	)

	return nil
}

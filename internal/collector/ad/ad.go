// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
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
	"time"

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

	perfDataCollector *pdh.Collector[perfDataCounterValues]
	perfDataObject    []perfDataCounterValues

	addressBookClientSessions                                *prometheus.Desc
	addressBookOperationsTotal                               *prometheus.Desc
	approximateHighestDistinguishedNameTag                   *prometheus.Desc
	atqAverageRequestLatencySeconds                          *prometheus.Desc
	atqCurrentThreads                                        *prometheus.Desc
	atqEstimatedDelaySeconds                                 *prometheus.Desc
	atqOutstandingRequests                                   *prometheus.Desc
	bindsTotal                                               *prometheus.Desc
	changeMonitorUpdatesPending                              *prometheus.Desc
	changeMonitorsRegistered                                 *prometheus.Desc
	databaseOperationsTotal                                  *prometheus.Desc
	directoryOperationsTotal                                 *prometheus.Desc
	directorySearchSubOperationsTotal                        *prometheus.Desc
	directoryServiceThreads                                  *prometheus.Desc
	interSiteReplicationDataBytesTotal                       *prometheus.Desc
	intraSiteReplicationDataBytesTotal                       *prometheus.Desc
	ldapActiveThreads                                        *prometheus.Desc
	ldapClientSessions                                       *prometheus.Desc
	ldapClosedConnectionsTotal                               *prometheus.Desc
	ldapLastBindTimeSeconds                                  *prometheus.Desc
	ldapOpenedConnectionsTotal                               *prometheus.Desc
	ldapSearchesTotal                                        *prometheus.Desc
	ldapUdpOperationsTotal                                   *prometheus.Desc
	ldapWritesTotal                                          *prometheus.Desc
	linkValuesCleanedTotal                                   *prometheus.Desc
	nameCacheHitsTotal                                       *prometheus.Desc
	nameCacheLookupsTotal                                    *prometheus.Desc
	nameTranslationsTotal                                    *prometheus.Desc
	phantomObjectsCleanedTotal                               *prometheus.Desc
	phantomObjectsVisitedTotal                               *prometheus.Desc
	replicationHighestUsn                                    *prometheus.Desc
	replicationInboundLinkValueUpdatesRemaining              *prometheus.Desc
	replicationInboundObjectsFilteredTotal                   *prometheus.Desc
	replicationInboundObjectsUpdatedTotal                    *prometheus.Desc
	replicationInboundPropertiesFilteredTotal                *prometheus.Desc
	replicationInboundPropertiesUpdatedTotal                 *prometheus.Desc
	replicationInboundSyncObjectsRemaining                   *prometheus.Desc
	replicationPendingOperations                             *prometheus.Desc
	replicationPendingSynchronizations                       *prometheus.Desc
	replicationSyncRequestsSchemaMismatchFailureTotal        *prometheus.Desc
	replicationSyncRequestsSuccessTotal                      *prometheus.Desc
	replicationSyncRequestsTotal                             *prometheus.Desc
	samComputerCreationRequestsTotal                         *prometheus.Desc
	samComputerCreationSuccessfulRequestsTotal               *prometheus.Desc
	samEnumerationsTotal                                     *prometheus.Desc
	samGroupEvaluationLatencySeconds                         *prometheus.Desc
	samGroupMembershipEvaluationsNonTransitiveTotal          *prometheus.Desc
	samGroupMembershipEvaluationsTotal                       *prometheus.Desc
	samGroupMembershipEvaluationsTransitiveTotal             *prometheus.Desc
	samGroupMembershipGlobalCatalogEvaluationsTotal          *prometheus.Desc
	samMembershipChangesTotal                                *prometheus.Desc
	samPasswordChangesTotal                                  *prometheus.Desc
	samQueryDisplayRequestsTotal                             *prometheus.Desc
	samUserCreationRequestsTotal                             *prometheus.Desc
	samUserCreationSuccessfulRequestsTotal                   *prometheus.Desc
	searchesTotal                                            *prometheus.Desc
	securityDescriptorPropagationAverageExclusionTimeSeconds *prometheus.Desc
	securityDescriptorPropagationEventsQueued                *prometheus.Desc
	securityDescriptorPropagationSubOperationsTotal          *prometheus.Desc
	securityDescriptorPropagationItemsQueued                 *prometheus.Desc
	tombstonesObjectsCollectedTotal                          *prometheus.Desc
	tombstonesObjectsVisitedTotal                            *prometheus.Desc
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

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	c.buildDescriptors()

	var err error

	c.perfDataCollector, err = pdh.NewCollector[perfDataCounterValues](logger.With(slog.String("collector", Name)), pdh.CounterTypeRaw, "DirectoryServices", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create DirectoryServices collector: %w", err)
	}

	return nil
}

func (c *Collector) buildDescriptors() {
	c.addressBookOperationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "address_book_operations_total"),
		"Total number of address book (NSPI) operations, by operation (AB ANR/Browses/Matches/Property Reads/Searches/Proxy Lookups).",
		[]string{"operation"},
		nil,
	)
	c.addressBookClientSessions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "address_book_client_sessions"),
		"Number of connected address book (NSPI) client sessions (AB Client Sessions).",
		nil,
		nil,
	)
	c.approximateHighestDistinguishedNameTag = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "approximate_highest_distinguished_name_tag"),
		"Approximate highest distinguished name tag (DNT) allocated in the directory database (Approximate highest DNT).",
		nil,
		nil,
	)
	c.atqEstimatedDelaySeconds = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "atq_estimated_delay_seconds"),
		"Estimated time in seconds the next request will spend in the ATQ queue before it is serviced (ATQ Estimated Queue Delay).",
		nil,
		nil,
	)
	c.atqOutstandingRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "atq_outstanding_requests"),
		"Number of requests queued at the ATQ thread pool (ATQ Outstanding Queued Requests).",
		nil,
		nil,
	)
	c.atqAverageRequestLatencySeconds = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "atq_average_request_latency_seconds"),
		"Average time in seconds to process a request, not including time spent on the queue (ATQ Request Latency).",
		nil,
		nil,
	)
	c.atqCurrentThreads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "atq_current_threads"),
		"Number of ATQ threads currently servicing requests, by service (ATQ Threads LDAP/Other).",
		[]string{"service"},
		nil,
	)
	c.searchesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "searches_total"),
		"Total number of directory searches, by search scope (Base/Subtree/Onelevel searches).",
		[]string{"scope"},
		nil,
	)
	c.databaseOperationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "database_operations_total"),
		"Total number of directory database operations, by operation (Database adds/deletes/modifys/recycles).",
		[]string{"operation"},
		nil,
	)
	c.bindsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "binds_total"),
		"Total number of binds, by bind method.",
		[]string{"bind_method"},
		nil,
	)
	c.replicationHighestUsn = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_highest_usn_total"),
		"Highest update sequence number (USN) committed or issued by this domain controller (DRA Highest USN Committed/Issued).",
		[]string{"state"},
		nil,
	)
	c.intraSiteReplicationDataBytesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_data_intrasite_bytes_total"),
		"Total number of uncompressed bytes replicated within the site, by direction (DRA Inbound/Outbound Bytes Not Compressed (Within Site)).",
		[]string{"direction"},
		nil,
	)
	c.interSiteReplicationDataBytesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_data_intersite_bytes_total"),
		"Total number of compressed bytes, after compression, replicated between sites, by direction (DRA Inbound/Outbound Bytes Compressed (Between Sites, After Compression)).",
		[]string{"direction"},
		nil,
	)
	c.replicationInboundSyncObjectsRemaining = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_inbound_sync_objects_remaining"),
		"Number of objects remaining until a full synchronization completes (DRA Inbound Full Sync Objects Remaining).",
		nil,
		nil,
	)
	c.replicationInboundLinkValueUpdatesRemaining = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_inbound_link_value_updates_remaining"),
		"Number of link value updates received in the current replication packet that are not yet applied (DRA Inbound Link Value Updates Remaining in Packet).",
		nil,
		nil,
	)
	c.replicationInboundObjectsUpdatedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_inbound_objects_updated_total"),
		"Total number of inbound replicated objects applied to the local directory (DRA Inbound Objects Applied).",
		nil,
		nil,
	)
	c.replicationInboundObjectsFilteredTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_inbound_objects_filtered_total"),
		"Total number of inbound replicated objects that had no updates to apply (DRA Inbound Objects Filtered).",
		nil,
		nil,
	)
	c.replicationInboundPropertiesUpdatedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_inbound_properties_updated_total"),
		"Total number of inbound replicated properties applied to the local directory (DRA Inbound Properties Applied).",
		nil,
		nil,
	)
	c.replicationInboundPropertiesFilteredTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_inbound_properties_filtered_total"),
		"Total number of inbound replicated property changes that were already applied (DRA Inbound Properties Filtered).",
		nil,
		nil,
	)
	c.replicationPendingOperations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_pending_operations"),
		"Number of queued replication operations not yet performed (DRA Pending Replication Operations).",
		nil,
		nil,
	)
	c.replicationPendingSynchronizations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_pending_synchronizations"),
		"Number of directory synchronizations queued for this server but not yet processed (DRA Pending Replication Synchronizations).",
		nil,
		nil,
	)
	c.replicationSyncRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_sync_requests_total"),
		"Total number of synchronization requests made to replication partners (DRA Sync Requests Made).",
		nil,
		nil,
	)
	c.replicationSyncRequestsSuccessTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_sync_requests_success_total"),
		"Total number of successful synchronization requests made to replication partners (DRA Sync Requests Successful).",
		nil,
		nil,
	)
	c.replicationSyncRequestsSchemaMismatchFailureTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_sync_requests_schema_mismatch_failure_total"),
		"Total number of synchronization requests that failed because of a schema mismatch (DRA Sync Failures on Schema Mismatch).",
		nil,
		nil,
	)
	c.nameTranslationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "name_translations_total"),
		"Total number of name translation requests, by origin of the request (DS Client/Server Name Translations).",
		[]string{"target_name"},
		nil,
	)
	c.changeMonitorsRegistered = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "change_monitors_registered"),
		"Number of change notification requests registered with this server (DS Monitor List Size).",
		nil,
		nil,
	)
	c.changeMonitorUpdatesPending = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "change_monitor_updates_pending"),
		"Number of pending change notifications not yet sent to clients (DS Notify Queue Size).",
		nil,
		nil,
	)
	c.nameCacheHitsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "name_cache_hits_total"),
		"Total number of name translations resolved from the directory name cache (DS Name Cache hit rate numerator).",
		nil,
		nil,
	)
	c.nameCacheLookupsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "name_cache_lookups_total"),
		"Total number of lookups in the directory name cache (DS Name Cache hit rate denominator).",
		nil,
		nil,
	)
	c.directoryOperationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "directory_operations_total"),
		"Total number of directory reads, searches and writes, by operation and origin (DS % Reads/Searches/Writes from ... numerators).",
		[]string{"operation", "origin"},
		nil,
	)
	c.directorySearchSubOperationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "directory_search_suboperations_total"),
		"Total number of directory search sub-operations; one sub-operation roughly corresponds to one object examined (DS Search sub-operations).",
		nil,
		nil,
	)
	c.securityDescriptorPropagationSubOperationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "security_descriptor_propagation_suboperations_total"),
		"Total number of security descriptor propagation sub-operations; one sub-operation roughly corresponds to one object examined (DS Security Descriptor sub-operations).",
		nil,
		nil,
	)
	c.securityDescriptorPropagationEventsQueued = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "security_descriptor_propagation_events_queued"),
		"Number of security descriptor propagation events that are queued but not yet processed (DS Security Descriptor Propagations Events).",
		nil,
		nil,
	)
	c.securityDescriptorPropagationAverageExclusionTimeSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "security_descriptor_propagation_average_exclusion_time_seconds"),
		"Average time in seconds the security descriptor propagator waits for exclusive access to database elements during a sub-operation (DS Security Descriptor Propagator Average Exclusion Time, reported in milliseconds like the other latency counters of the object).",
		nil,
		nil,
	)
	c.securityDescriptorPropagationItemsQueued = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "security_descriptor_propagation_items_queued"),
		"Number of objects remaining to examine while processing the current security descriptor propagation event (DS Security Descriptor Propagator Runtime Queue).",
		nil,
		nil,
	)
	c.directoryServiceThreads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "directory_service_threads"),
		"Number of threads in use by the directory service (DS Threads in Use).",
		nil,
		nil,
	)
	c.ldapClosedConnectionsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ldap_closed_connections_total"),
		"Total number of closed LDAP connections (LDAP Closed Connections).",
		nil,
		nil,
	)
	c.ldapOpenedConnectionsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ldap_opened_connections_total"),
		"Total number of opened LDAP connections, by type (LDAP New Connections, LDAP New SSL Connections).",
		[]string{"type"},
		nil,
	)
	c.ldapActiveThreads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ldap_active_threads"),
		"Number of threads in use by the LDAP subsystem (LDAP Active Threads).",
		nil,
		nil,
	)
	c.ldapLastBindTimeSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ldap_last_bind_time_seconds"),
		"Time in seconds taken by the last successful LDAP bind (LDAP Bind Time).",
		nil,
		nil,
	)
	c.ldapSearchesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ldap_searches_total"),
		"Total number of LDAP search operations (LDAP Searches).",
		nil,
		nil,
	)
	c.ldapUdpOperationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ldap_udp_operations_total"),
		"Total number of LDAP UDP operations (LDAP UDP operations).",
		nil,
		nil,
	)
	c.ldapWritesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ldap_writes_total"),
		"Total number of LDAP write operations (LDAP Writes).",
		nil,
		nil,
	)
	c.ldapClientSessions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ldap_client_sessions"),
		"Number of sessions opened by LDAP clients (LDAP Client Sessions). Long sustained high values can indicate an overworked domain controller.",
		nil,
		nil,
	)
	c.linkValuesCleanedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "link_values_cleaned_total"),
		"Total number of link values cleaned (Link Values Cleaned).",
		nil,
		nil,
	)
	c.phantomObjectsCleanedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "phantom_objects_cleaned_total"),
		"Total number of phantom objects cleaned (Phantoms Cleaned).",
		nil,
		nil,
	)
	c.phantomObjectsVisitedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "phantom_objects_visited_total"),
		"Total number of phantom objects visited (Phantoms Visited).",
		nil,
		nil,
	)
	c.samGroupMembershipEvaluationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_group_membership_evaluations_total"),
		"Total number of SAM group membership evaluations, by group type (SAM Global/Domain Local/Universal Group Membership Evaluations).",
		[]string{"group_type"},
		nil,
	)
	c.samGroupMembershipGlobalCatalogEvaluationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_group_membership_global_catalog_evaluations_total"),
		"Total number of SAM group membership evaluations that required a global catalog (SAM GC Evaluations).",
		nil,
		nil,
	)
	c.samGroupMembershipEvaluationsNonTransitiveTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_group_membership_evaluations_nontransitive_total"),
		"Total number of non-transitive SAM group membership evaluations (SAM Non-Transitive Membership Evaluations).",
		nil,
		nil,
	)
	c.samGroupMembershipEvaluationsTransitiveTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_group_membership_evaluations_transitive_total"),
		"Total number of transitive SAM group membership evaluations (SAM Transitive Membership Evaluations).",
		nil,
		nil,
	)
	c.samGroupEvaluationLatencySeconds = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_group_evaluation_latency_seconds"),
		"Mean latency in seconds of the last 100 group evaluations performed for authentication, by evaluation type (SAM Account/Resource Group Evaluation Latency, reported in milliseconds like the other latency counters of the object).",
		[]string{"evaluation_type"},
		nil,
	)
	c.samComputerCreationRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_computer_creation_requests_total"),
		"Total number of attempts to create computer accounts (SAM Machine Creation Attempts).",
		nil,
		nil,
	)
	c.samComputerCreationSuccessfulRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_computer_creation_successful_requests_total"),
		"Total number of computer accounts created successfully (SAM Successful Computer Creations).",
		nil,
		nil,
	)
	c.samUserCreationRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_user_creation_requests_total"),
		"Total number of attempts to create user accounts (SAM User Creation Attempts).",
		nil,
		nil,
	)
	c.samUserCreationSuccessfulRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_user_creation_successful_requests_total"),
		"Total number of user accounts created successfully (SAM Successful User Creations).",
		nil,
		nil,
	)
	c.samQueryDisplayRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_query_display_requests_total"),
		"Total number of SAM display information queries (SAM Display Information Queries).",
		nil,
		nil,
	)
	c.samEnumerationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_enumerations_total"),
		"Total number of SAM enumerations (SAM Enumerations).",
		nil,
		nil,
	)
	c.samMembershipChangesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_membership_changes_total"),
		"Total number of SAM group membership changes (SAM Membership Changes).",
		nil,
		nil,
	)
	c.samPasswordChangesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sam_password_changes_total"),
		"Total number of SAM password changes (SAM Password Changes).",
		nil,
		nil,
	)

	c.tombstonesObjectsCollectedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "tombstoned_objects_collected_total"),
		"Total number of tombstoned objects garbage collected (Tombstones Garbage Collected).",
		nil,
		nil,
	)
	c.tombstonesObjectsVisitedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "tombstoned_objects_visited_total"),
		"Total number of tombstoned objects visited by garbage collection (Tombstones Visited).",
		nil,
		nil,
	)
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric, _ time.Duration) error {
	err := c.perfDataCollector.Collect(&c.perfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect DirectoryServices (AD) metrics: %w", err)
	} else if len(c.perfDataObject) == 0 {
		return fmt.Errorf("failed to collect DirectoryServices (AD) metrics: %w", types.ErrNoDataUnexpected)
	}

	c.collectMetrics(ch, c.perfDataObject[0])

	return nil
}

// collectMetrics publishes the metrics of one DirectoryServices sample.
// Values the counter reports in milliseconds are converted to seconds.
func (c *Collector) collectMetrics(ch chan<- prometheus.Metric, data perfDataCounterValues) {
	ch <- prometheus.MustNewConstMetric(
		c.addressBookOperationsTotal,
		prometheus.CounterValue,
		data.AbANRPerSec,
		"ambiguous_name_resolution",
	)

	ch <- prometheus.MustNewConstMetric(
		c.addressBookOperationsTotal,
		prometheus.CounterValue,
		data.AbBrowsesPerSec,
		"browse",
	)

	ch <- prometheus.MustNewConstMetric(
		c.addressBookOperationsTotal,
		prometheus.CounterValue,
		data.AbMatchesPerSec,
		"find",
	)

	ch <- prometheus.MustNewConstMetric(
		c.addressBookOperationsTotal,
		prometheus.CounterValue,
		data.AbPropertyReadsPerSec,
		"property_read",
	)

	ch <- prometheus.MustNewConstMetric(
		c.addressBookOperationsTotal,
		prometheus.CounterValue,
		data.AbSearchesPerSec,
		"search",
	)

	ch <- prometheus.MustNewConstMetric(
		c.addressBookOperationsTotal,
		prometheus.CounterValue,
		data.AbProxyLookupsPerSec,
		"proxy_search",
	)

	ch <- prometheus.MustNewConstMetric(
		c.addressBookClientSessions,
		prometheus.GaugeValue,
		data.AbClientSessions,
	)

	ch <- prometheus.MustNewConstMetric(
		c.approximateHighestDistinguishedNameTag,
		prometheus.GaugeValue,
		data.ApproximateHighestDNT,
	)

	ch <- prometheus.MustNewConstMetric(
		c.atqEstimatedDelaySeconds,
		prometheus.GaugeValue,
		data.AtqEstimatedQueueDelay/1000,
	)

	ch <- prometheus.MustNewConstMetric(
		c.atqOutstandingRequests,
		prometheus.GaugeValue,
		data.AtqOutstandingQueuedRequests,
	)

	ch <- prometheus.MustNewConstMetric(
		c.atqAverageRequestLatencySeconds,
		prometheus.GaugeValue,
		data.AtqRequestLatency/1000,
	)

	ch <- prometheus.MustNewConstMetric(
		c.atqCurrentThreads,
		prometheus.GaugeValue,
		data.AtqThreadsLDAP,
		"ldap",
	)

	ch <- prometheus.MustNewConstMetric(
		c.atqCurrentThreads,
		prometheus.GaugeValue,
		data.AtqThreadsOther,
		"other",
	)

	ch <- prometheus.MustNewConstMetric(
		c.searchesTotal,
		prometheus.CounterValue,
		data.BaseSearchesPerSec,
		"base",
	)

	ch <- prometheus.MustNewConstMetric(
		c.searchesTotal,
		prometheus.CounterValue,
		data.SubtreeSearchesPerSec,
		"subtree",
	)

	ch <- prometheus.MustNewConstMetric(
		c.searchesTotal,
		prometheus.CounterValue,
		data.OneLevelSearchesPerSec,
		"one_level",
	)

	ch <- prometheus.MustNewConstMetric(
		c.databaseOperationsTotal,
		prometheus.CounterValue,
		data.DatabaseAddsPerSec,
		"add",
	)

	ch <- prometheus.MustNewConstMetric(
		c.databaseOperationsTotal,
		prometheus.CounterValue,
		data.DatabaseDeletesPerSec,
		"delete",
	)

	ch <- prometheus.MustNewConstMetric(
		c.databaseOperationsTotal,
		prometheus.CounterValue,
		data.DatabaseModifiesPerSec,
		"modify",
	)

	ch <- prometheus.MustNewConstMetric(
		c.databaseOperationsTotal,
		prometheus.CounterValue,
		data.DatabaseRecyclesPerSec,
		"recycle",
	)

	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		data.DigestBindsPerSec,
		"digest",
	)

	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		data.DsClientBindsPerSec,
		"ds_client",
	)

	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		data.DsServerBindsPerSec,
		"ds_server",
	)

	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		data.ExternalBindsPerSec,
		"external",
	)

	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		data.FastBindsPerSec,
		"fast",
	)

	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		data.NegotiatedBindsPerSec,
		"negotiate",
	)

	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		data.NTLMBindsPerSec,
		"ntlm",
	)

	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		data.SimpleBindsPerSec,
		"simple",
	)

	ch <- prometheus.MustNewConstMetric(
		c.bindsTotal,
		prometheus.CounterValue,
		data.LdapSuccessfulBindsPerSec,
		"ldap",
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationHighestUsn,
		prometheus.CounterValue,
		float64(uint64(data.DRAHighestUSNCommittedHighPart)<<32)+data.DRAHighestUSNCommittedLowPart,
		"committed",
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationHighestUsn,
		prometheus.CounterValue,
		float64(uint64(data.DRAHighestUSNIssuedHighPart)<<32)+data.DRAHighestUSNIssuedLowPart,
		"issued",
	)

	ch <- prometheus.MustNewConstMetric(
		c.interSiteReplicationDataBytesTotal,
		prometheus.CounterValue,
		data.DRAInboundBytesCompressedBetweenSitesAfterCompressionPerSec,
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
		data.DRAOutboundBytesCompressedBetweenSitesAfterCompressionPerSec,
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
		data.DRAInboundBytesNotCompressedWithinSitePerSec,
		"inbound",
	)

	ch <- prometheus.MustNewConstMetric(
		c.intraSiteReplicationDataBytesTotal,
		prometheus.CounterValue,
		data.DRAOutboundBytesNotCompressedWithinSitePerSec,
		"outbound",
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationInboundSyncObjectsRemaining,
		prometheus.GaugeValue,
		data.DRAInboundFullSyncObjectsRemaining,
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationInboundLinkValueUpdatesRemaining,
		prometheus.GaugeValue,
		data.DRAInboundLinkValueUpdatesRemainingInPacket,
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationInboundObjectsUpdatedTotal,
		prometheus.CounterValue,
		data.DRAInboundObjectsAppliedPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationInboundObjectsFilteredTotal,
		prometheus.CounterValue,
		data.DRAInboundObjectsFilteredPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationInboundPropertiesUpdatedTotal,
		prometheus.CounterValue,
		data.DRAInboundPropertiesAppliedPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationInboundPropertiesFilteredTotal,
		prometheus.CounterValue,
		data.DRAInboundPropertiesFilteredPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationPendingOperations,
		prometheus.GaugeValue,
		data.DRAPendingReplicationOperations,
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationPendingSynchronizations,
		prometheus.GaugeValue,
		data.DRAPendingReplicationSynchronizations,
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationSyncRequestsTotal,
		prometheus.CounterValue,
		data.DRASyncRequestsMade,
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationSyncRequestsSuccessTotal,
		prometheus.CounterValue,
		data.DRASyncRequestsSuccessful,
	)

	ch <- prometheus.MustNewConstMetric(
		c.replicationSyncRequestsSchemaMismatchFailureTotal,
		prometheus.CounterValue,
		data.DRASyncFailuresOnSchemaMismatch,
	)

	ch <- prometheus.MustNewConstMetric(
		c.nameTranslationsTotal,
		prometheus.CounterValue,
		data.DsClientNameTranslationsPerSec,
		"client",
	)

	ch <- prometheus.MustNewConstMetric(
		c.nameTranslationsTotal,
		prometheus.CounterValue,
		data.DsServerNameTranslationsPerSec,
		"server",
	)

	ch <- prometheus.MustNewConstMetric(
		c.changeMonitorsRegistered,
		prometheus.GaugeValue,
		data.DsMonitorListSize,
	)

	ch <- prometheus.MustNewConstMetric(
		c.changeMonitorUpdatesPending,
		prometheus.GaugeValue,
		data.DsNotifyQueueSize,
	)

	ch <- prometheus.MustNewConstMetric(
		c.nameCacheHitsTotal,
		prometheus.CounterValue,
		data.DsNameCacheHitRate,
	)

	ch <- prometheus.MustNewConstMetric(
		c.nameCacheLookupsTotal,
		prometheus.CounterValue,
		data.DsNameCacheHitRateSecondValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data.DsPercentReadsFromDRA,
		"read",
		"replication_agent",
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data.DsPercentReadsFromKCC,
		"read",
		"knowledge_consistency_checker",
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data.DsPercentReadsFromLSA,
		"read",
		"local_security_authority",
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data.DsPercentReadsFromNSPI,
		"read",
		"name_service_provider_interface",
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data.DsPercentReadsFromNTDSAPI,
		"read",
		"directory_service_api",
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data.DsPercentReadsFromSAM,
		"read",
		"security_account_manager",
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data.DsPercentReadsOther,
		"read",
		"other",
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data.DsPercentSearchesFromDRA,
		"search",
		"replication_agent",
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data.DsPercentSearchesFromKCC,
		"search",
		"knowledge_consistency_checker",
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data.DsPercentSearchesFromLDAP,
		"search",
		"ldap",
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data.DsPercentSearchesFromLSA,
		"search",
		"local_security_authority",
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data.DsPercentSearchesFromNSPI,
		"search",
		"name_service_provider_interface",
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data.DsPercentSearchesFromNTDSAPI,
		"search",
		"directory_service_api",
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data.DsPercentSearchesFromSAM,
		"search",
		"security_account_manager",
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data.DsPercentSearchesOther,
		"search",
		"other",
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data.DsPercentWritesFromDRA,
		"write",
		"replication_agent",
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data.DsPercentWritesFromKCC,
		"write",
		"knowledge_consistency_checker",
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data.DsPercentWritesFromLDAP,
		"write",
		"ldap",
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data.DsPercentWritesFromLSA,
		"write",
		"local_security_authority",
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data.DsPercentWritesFromNSPI,
		"write",
		"name_service_provider_interface",
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data.DsPercentWritesFromNTDSAPI,
		"write",
		"directory_service_api",
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data.DsPercentWritesFromSAM,
		"write",
		"security_account_manager",
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryOperationsTotal,
		prometheus.CounterValue,
		data.DsPercentWritesOther,
		"write",
		"other",
	)

	ch <- prometheus.MustNewConstMetric(
		c.directorySearchSubOperationsTotal,
		prometheus.CounterValue,
		data.DsSearchSubOperationsPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.securityDescriptorPropagationSubOperationsTotal,
		prometheus.CounterValue,
		data.DsSecurityDescriptorSubOperationsPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.securityDescriptorPropagationEventsQueued,
		prometheus.GaugeValue,
		data.DsSecurityDescriptorPropagationsEvents,
	)

	ch <- prometheus.MustNewConstMetric(
		c.securityDescriptorPropagationAverageExclusionTimeSeconds,
		prometheus.GaugeValue,
		data.DsSecurityDescriptorPropagatorAverageExclusionTime/1000,
	)

	ch <- prometheus.MustNewConstMetric(
		c.securityDescriptorPropagationItemsQueued,
		prometheus.GaugeValue,
		data.DsSecurityDescriptorPropagatorRuntimeQueue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.directoryServiceThreads,
		prometheus.GaugeValue,
		data.DsThreadsInUse,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ldapClosedConnectionsTotal,
		prometheus.CounterValue,
		data.LdapClosedConnectionsPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ldapOpenedConnectionsTotal,
		prometheus.CounterValue,
		data.LdapNewConnectionsPerSec,
		"ldap",
	)

	ch <- prometheus.MustNewConstMetric(
		c.ldapOpenedConnectionsTotal,
		prometheus.CounterValue,
		data.LdapNewSSLConnectionsPerSec,
		"ldaps",
	)

	ch <- prometheus.MustNewConstMetric(
		c.ldapActiveThreads,
		prometheus.GaugeValue,
		data.LdapActiveThreads,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ldapLastBindTimeSeconds,
		prometheus.GaugeValue,
		data.LdapBindTime/1000,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ldapSearchesTotal,
		prometheus.CounterValue,
		data.LdapSearchesPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ldapUdpOperationsTotal,
		prometheus.CounterValue,
		data.LdapUDPOperationsPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ldapWritesTotal,
		prometheus.CounterValue,
		data.LdapWritesPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ldapClientSessions,
		prometheus.GaugeValue,
		data.LdapClientSessions,
	)

	ch <- prometheus.MustNewConstMetric(
		c.linkValuesCleanedTotal,
		prometheus.CounterValue,
		data.LinkValuesCleanedPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.phantomObjectsCleanedTotal,
		prometheus.CounterValue,
		data.PhantomsCleanedPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.phantomObjectsVisitedTotal,
		prometheus.CounterValue,
		data.PhantomsVisitedPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samGroupMembershipEvaluationsTotal,
		prometheus.CounterValue,
		data.SamGlobalGroupMembershipEvaluationsPerSec,
		"global",
	)

	ch <- prometheus.MustNewConstMetric(
		c.samGroupMembershipEvaluationsTotal,
		prometheus.CounterValue,
		data.SamDomainLocalGroupMembershipEvaluationsPerSec,
		"domain_local",
	)

	ch <- prometheus.MustNewConstMetric(
		c.samGroupMembershipEvaluationsTotal,
		prometheus.CounterValue,
		data.SamUniversalGroupMembershipEvaluationsPerSec,
		"universal",
	)

	ch <- prometheus.MustNewConstMetric(
		c.samGroupMembershipGlobalCatalogEvaluationsTotal,
		prometheus.CounterValue,
		data.SamGCEvaluationsPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samGroupMembershipEvaluationsNonTransitiveTotal,
		prometheus.CounterValue,
		data.SamNonTransitiveMembershipEvaluationsPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samGroupMembershipEvaluationsTransitiveTotal,
		prometheus.CounterValue,
		data.SamTransitiveMembershipEvaluationsPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samGroupEvaluationLatencySeconds,
		prometheus.GaugeValue,
		data.SamAccountGroupEvaluationLatency/1000,
		"account_group",
	)

	ch <- prometheus.MustNewConstMetric(
		c.samGroupEvaluationLatencySeconds,
		prometheus.GaugeValue,
		data.SamResourceGroupEvaluationLatency/1000,
		"resource_group",
	)

	ch <- prometheus.MustNewConstMetric(
		c.samComputerCreationRequestsTotal,
		prometheus.CounterValue,
		data.SamMachineCreationAttemptsPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samComputerCreationSuccessfulRequestsTotal,
		prometheus.CounterValue,
		data.SamSuccessfulComputerCreationsPerSecIncludesAllRequests,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samUserCreationRequestsTotal,
		prometheus.CounterValue,
		data.SamUserCreationAttemptsPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samUserCreationSuccessfulRequestsTotal,
		prometheus.CounterValue,
		data.SamSuccessfulUserCreationsPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samQueryDisplayRequestsTotal,
		prometheus.CounterValue,
		data.SamDisplayInformationQueriesPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samEnumerationsTotal,
		prometheus.CounterValue,
		data.SamEnumerationsPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samMembershipChangesTotal,
		prometheus.CounterValue,
		data.SamMembershipChangesPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samPasswordChangesTotal,
		prometheus.CounterValue,
		data.SamPasswordChangesPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.tombstonesObjectsCollectedTotal,
		prometheus.CounterValue,
		data.TombstonesGarbageCollectedPerSec,
	)

	ch <- prometheus.MustNewConstMetric(
		c.tombstonesObjectsVisitedTotal,
		prometheus.CounterValue,
		data.TombstonesVisitedPerSec,
	)
}

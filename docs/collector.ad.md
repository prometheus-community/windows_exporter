# ad collector

The ad collector exposes metrics about a Active Directory Domain Services domain controller

|||
-|-
Metric name prefix  | `ad`
Classes             | [`Win32_PerfRawData_DirectoryServices_DirectoryServices`](https://msdn.microsoft.com/en-us/library/ms803980.aspx)
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_ad_address_book_operations_total` | Total number of address book (NSPI) operations, by operation (AB ANR/Browses/Matches/Property Reads/Searches/Proxy Lookups) | counter | `operation`
`windows_ad_address_book_client_sessions` | Number of connected address book (NSPI) client sessions (AB Client Sessions) | gauge | None
`windows_ad_approximate_highest_distinguished_name_tag` | Approximate highest distinguished name tag (DNT) allocated in the directory database (Approximate highest DNT) | gauge | None
`windows_ad_atq_estimated_delay_seconds` | Estimated time in seconds the next request will spend in the ATQ queue before it is serviced (ATQ Estimated Queue Delay) | gauge | None
`windows_ad_atq_outstanding_requests` | Number of requests queued at the ATQ thread pool (ATQ Outstanding Queued Requests) | gauge | None
`windows_ad_atq_average_request_latency_seconds` | Average time in seconds to process a request, not including time spent on the queue (ATQ Request Latency) | gauge | None
`windows_ad_atq_current_threads` | Number of ATQ threads currently servicing requests, by service (ATQ Threads LDAP/Other) | gauge | `service`
`windows_ad_searches_total` | Total number of directory searches, by search scope (Base/Subtree/Onelevel searches) | counter | `scope`
`windows_ad_database_operations_total` | Total number of directory database operations, by operation (Database adds/deletes/modifys/recycles) | counter | `operation`
`windows_ad_binds_total` | Total number of binds, by bind method | counter | `bind_method`
`windows_ad_replication_highest_usn` | Highest update sequence number (USN) committed or issued by this domain controller (DRA Highest USN Committed/Issued) | counter | `state`
`windows_ad_replication_data_intrasite_bytes_total` | Total number of uncompressed bytes replicated within the site, by direction (DRA Inbound/Outbound Bytes Not Compressed (Within Site)) | counter | `direction`
`windows_ad_replication_data_intersite_bytes_total` | Total number of compressed bytes, after compression, replicated between sites, by direction (DRA Inbound/Outbound Bytes Compressed (Between Sites, After Compression)) | counter | `direction`
`windows_ad_replication_inbound_sync_objects_remaining` | Number of objects remaining until a full synchronization completes (DRA Inbound Full Sync Objects Remaining) | gauge | None
`windows_ad_replication_inbound_link_value_updates_remaining` | Number of link value updates received in the current replication packet that are not yet applied (DRA Inbound Link Value Updates Remaining in Packet) | gauge | None
`windows_ad_replication_inbound_objects_updated_total` | Total number of inbound replicated objects applied to the local directory (DRA Inbound Objects Applied) | counter | None
`windows_ad_replication_inbound_objects_filtered_total` | Total number of inbound replicated objects that had no updates to apply (DRA Inbound Objects Filtered) | counter | None
`windows_ad_replication_inbound_properties_updated_total` | Total number of inbound replicated properties applied to the local directory (DRA Inbound Properties Applied) | counter | None
`windows_ad_replication_inbound_properties_filtered_total` | Total number of inbound replicated property changes that were already applied (DRA Inbound Properties Filtered) | counter | None
`windows_ad_replication_pending_operations` | Number of queued replication operations not yet performed (DRA Pending Replication Operations) | gauge | None
`windows_ad_replication_pending_synchronizations` | Number of directory synchronizations queued for this server but not yet processed (DRA Pending Replication Synchronizations) | gauge | None
`windows_ad_replication_sync_requests_total` | Total number of synchronization requests made to replication partners (DRA Sync Requests Made) | counter | None
`windows_ad_replication_sync_requests_success_total` | Total number of successful synchronization requests made to replication partners (DRA Sync Requests Successful) | counter | None
`windows_ad_replication_sync_requests_schema_mismatch_failure_total` | Total number of synchronization requests that failed because of a schema mismatch (DRA Sync Failures on Schema Mismatch) | counter | None
`windows_ad_name_translations_total` | Total number of name translation requests, by origin of the request (DS Client/Server Name Translations) | counter | `target_name`
`windows_ad_change_monitors_registered` | Number of change notification requests registered with this server (DS Monitor List Size) | gauge | None
`windows_ad_change_monitor_updates_pending` | Number of pending change notifications not yet sent to clients (DS Notify Queue Size) | gauge | None
`windows_ad_name_cache_hits_total` | Total number of name translations resolved from the directory name cache (DS Name Cache hit rate numerator) | counter | None
`windows_ad_name_cache_lookups_total` | Total number of lookups in the directory name cache (DS Name Cache hit rate denominator) | counter | None
`windows_ad_directory_operations_total` | Total number of directory reads, searches and writes, by operation and origin (DS % Reads/Searches/Writes from ... numerators) | counter | `operation`, `origin`
`windows_ad_directory_search_suboperations_total` | Total number of directory search sub-operations; one sub-operation roughly corresponds to one object examined (DS Search sub-operations) | counter | None
`windows_ad_security_descriptor_propagation_suboperations_total` | Total number of security descriptor propagation sub-operations; one sub-operation roughly corresponds to one object examined (DS Security Descriptor sub-operations) | counter | None
`windows_ad_security_descriptor_propagation_events_queued` | Number of security descriptor propagation events that are queued but not yet processed (DS Security Descriptor Propagations Events) | gauge | None
`windows_ad_security_descriptor_propagation_average_exclusion_time_seconds` | Average time in seconds the security descriptor propagator waits for exclusive access to database elements during a sub-operation (DS Security Descriptor Propagator Average Exclusion Time, reported in milliseconds like the other latency counters of the object) | gauge | None
`windows_ad_security_descriptor_propagation_items_queued` | Number of objects remaining to examine while processing the current security descriptor propagation event (DS Security Descriptor Propagator Runtime Queue) | gauge | None
`windows_ad_directory_service_threads` | Number of threads in use by the directory service (DS Threads in Use) | gauge | None
`windows_ad_ldap_closed_connections_total` | Total number of closed LDAP connections (LDAP Closed Connections) | counter | None
`windows_ad_ldap_opened_connections_total` | Total number of opened LDAP connections, by type (LDAP New Connections, LDAP New SSL Connections) | counter | `type`
`windows_ad_ldap_active_threads` | Number of threads in use by the LDAP subsystem (LDAP Active Threads) | gauge | None
`windows_ad_ldap_last_bind_time_seconds` | Time in seconds taken by the last successful LDAP bind (LDAP Bind Time) | gauge | None
`windows_ad_ldap_searches_total` | Total number of LDAP search operations (LDAP Searches) | counter | None
`windows_ad_ldap_udp_operations_total` | Total number of LDAP UDP operations (LDAP UDP operations) | counter | None
`windows_ad_ldap_writes_total` | Total number of LDAP write operations (LDAP Writes) | counter | None
`windows_ad_ldap_client_sessions` | Number of sessions opened by LDAP clients (LDAP Client Sessions). Long sustained high values can indicate an overworked domain controller | gauge | None
`windows_ad_link_values_cleaned_total` | Total number of link values cleaned (Link Values Cleaned) | counter | None
`windows_ad_phantom_objects_cleaned_total` | Total number of phantom objects cleaned (Phantoms Cleaned) | counter | None
`windows_ad_phantom_objects_visited_total` | Total number of phantom objects visited (Phantoms Visited) | counter | None
`windows_ad_sam_group_membership_evaluations_total` | Total number of SAM group membership evaluations, by group type (SAM Global/Domain Local/Universal Group Membership Evaluations) | counter | `group_type`
`windows_ad_sam_group_membership_global_catalog_evaluations_total` | Total number of SAM group membership evaluations that required a global catalog (SAM GC Evaluations) | counter | None
`windows_ad_sam_group_membership_evaluations_nontransitive_total` | Total number of non-transitive SAM group membership evaluations (SAM Non-Transitive Membership Evaluations) | counter | None
`windows_ad_sam_group_membership_evaluations_transitive_total` | Total number of transitive SAM group membership evaluations (SAM Transitive Membership Evaluations) | counter | None
`windows_ad_sam_group_evaluation_latency_seconds` | Mean latency in seconds of the last 100 group evaluations performed for authentication, by evaluation type (SAM Account/Resource Group Evaluation Latency, reported in milliseconds like the other latency counters of the object) | gauge | `evaluation_type`
`windows_ad_sam_computer_creation_requests_total` | Total number of attempts to create computer accounts (SAM Machine Creation Attempts) | counter | None
`windows_ad_sam_computer_creation_successful_requests_total` | Total number of computer accounts created successfully (SAM Successful Computer Creations) | counter | None
`windows_ad_sam_user_creation_requests_total` | Total number of attempts to create user accounts (SAM User Creation Attempts) | counter | None
`windows_ad_sam_user_creation_successful_requests_total` | Total number of user accounts created successfully (SAM Successful User Creations) | counter | None
`windows_ad_sam_query_display_requests_total` | Total number of SAM display information queries (SAM Display Information Queries) | counter | None
`windows_ad_sam_enumerations_total` | Total number of SAM enumerations (SAM Enumerations) | counter | None
`windows_ad_sam_membership_changes_total` | Total number of SAM group membership changes (SAM Membership Changes) | counter | None
`windows_ad_sam_password_changes_total` | Total number of SAM password changes (SAM Password Changes) | counter | None
`windows_ad_tombstoned_objects_collected_total` | Total number of tombstoned objects garbage collected (Tombstones Garbage Collected) | counter | None
`windows_ad_tombstoned_objects_visited_total` | Total number of tombstoned objects visited by garbage collection (Tombstones Visited) | counter | None

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

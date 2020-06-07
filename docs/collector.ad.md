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
`windows_ad_address_book_operations_total` | _Not yet documented_ | counter | `operation`
`windows_ad_address_book_client_sessions` | _Not yet documented_ | gauge | None
`windows_ad_approximate_highest_distinguished_name_tag` | _Not yet documented_ | gauge | None
`windows_ad_atq_estimated_delay_seconds` | _Not yet documented_ | gauge | None
`windows_ad_atq_outstanding_requests` | _Not yet documented_ | gauge | None
`windows_ad_atq_average_request_latency` | _Not yet documented_ | gauge | None
`windows_ad_atq_current_threads` | _Not yet documented_ | gauge | `service`
`windows_ad_searches_total` | _Not yet documented_ | counter | `scope`
`windows_ad_database_operations_total` | _Not yet documented_ | counter | `operation`
`windows_ad_binds_total` | _Not yet documented_ | counter | `bind_method`
`windows_ad_replication_highest_usn` | _Not yet documented_ | counter | `state`
`windows_ad_replication_data_intrasite_bytes_total` | _Not yet documented_ | counter | `direction`
`windows_ad_replication_data_intersite_bytes_total` | _Not yet documented_ | counter | `direction`
`windows_ad_replication_inbound_sync_objects_remaining` | _Not yet documented_ | gauge | None
`windows_ad_replication_inbound_link_value_updates_remaining` | _Not yet documented_ | gauge | None
`windows_ad_replication_inbound_objects_updated_total` | _Not yet documented_ | counter | None
`windows_ad_replication_inbound_objects_filtered_total` | _Not yet documented_ | counter | None
`windows_ad_replication_inbound_properties_updated_total` | _Not yet documented_ | counter | None
`windows_ad_replication_inbound_properties_filtered_total` | _Not yet documented_ | counter | None
`windows_ad_replication_pending_operations` | _Not yet documented_ | gauge | None
`windows_ad_replication_pending_synchronizations` | _Not yet documented_ | gauge | None
`windows_ad_replication_sync_requests_total` | _Not yet documented_ | counter | None
`windows_ad_replication_sync_requests_success_total` | _Not yet documented_ | counter | None
`windows_ad_replication_sync_requests_schema_mismatch_failure_total` | _Not yet documented_ | counter | None
`windows_ad_name_translations_total` | _Not yet documented_ | counter | `target_name`
`windows_ad_change_monitors_registered` | _Not yet documented_ | gauge | None
`windows_ad_change_monitor_updates_pending` | _Not yet documented_ | gauge | None
`windows_ad_name_cache_hits_total` | _Not yet documented_ | counter | None
`windows_ad_name_cache_lookups_total` | _Not yet documented_ | counter | None
`windows_ad_directory_operations_total` | _Not yet documented_ | counter | `operation`, `origin`
`windows_ad_directory_search_suboperations_total` | _Not yet documented_ | counter | None
`windows_ad_security_descriptor_propagation_events_total` | _Not yet documented_ | counter | None
`windows_ad_security_descriptor_propagation_events_queued` | _Not yet documented_ | gauge | None
`windows_ad_security_descriptor_propagation_access_wait_total_seconds` | _Not yet documented_ | gauge | None
`windows_ad_security_descriptor_propagation_items_queued_total` | _Not yet documented_ | counter | None
`windows_ad_directory_service_threads` | _Not yet documented_ | gauge | None
`windows_ad_ldap_closed_connections_total` | _Not yet documented_ | counter | None
`windows_ad_ldap_opened_connections_total` | _Not yet documented_ | counter | `type`
`windows_ad_ldap_active_threads` | _Not yet documented_ | gauge | None
`windows_ad_ldap_last_bind_time_seconds` | _Not yet documented_ | gauge | None
`windows_ad_ldap_searches_total` | _Not yet documented_ | counter | None
`windows_ad_ldap_udp_operations_total` | _Not yet documented_ | counter | None
`windows_ad_ldap_writes_total` | _Not yet documented_ | counter | None
`windows_ad_link_values_cleaned_total` | _Not yet documented_ | counter | None
`windows_ad_phantom_objects_cleaned_total` | _Not yet documented_ | counter | None
`windows_ad_phantom_objects_visited_total` | _Not yet documented_ | counter | None
`windows_ad_sam_group_membership_evaluations_total` | _Not yet documented_ | counter | `group_type`
`windows_ad_sam_group_membership_global_catalog_evaluations_total` | _Not yet documented_ | counter | None
`windows_ad_sam_group_membership_evaluations_nontransitive_total` | _Not yet documented_ | counter | None
`windows_ad_sam_group_membership_evaluations_transitive_total` | _Not yet documented_ | counter | None
`windows_ad_sam_group_evaluation_latency` | _Not yet documented_ | gauge | `evaluation_type`
`windows_ad_sam_computer_creation_requests_total` | _Not yet documented_ | counter | None
`windows_ad_sam_computer_creation_successful_requests_total` | _Not yet documented_ | counter | None
`windows_ad_sam_user_creation_requests_total` | _Not yet documented_ | counter | None
`windows_ad_sam_user_creation_successful_requests_total` | _Not yet documented_ | counter | None
`windows_ad_sam_query_display_requests_total` | _Not yet documented_ | counter | None
`windows_ad_sam_enumerations_total` | _Not yet documented_ | counter | None
`windows_ad_sam_membership_changes_total` | _Not yet documented_ | counter | None
`windows_ad_sam_password_changes_total` | _Not yet documented_ | counter | None
`windows_ad_tombstoned_objects_collected_total` | _Not yet documented_ | counter | None
`windows_ad_tombstoned_objects_visited_total` | _Not yet documented_ | counter | None

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

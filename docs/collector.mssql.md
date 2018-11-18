# mssql collector

The mssql collector exposes metrics about the MSSQL server

|||
-|-
Metric name prefix  | `mssql`
Classes             | [`Win32_PerfRawData_MSSQLSERVER_SQLServerAccessMethods`](https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-access-methods-object)<br/>[`Win32_PerfRawData_MSSQLSERVER_SQLServerAvailabilityReplica`](https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-availability-replica)<br/>[`Win32_PerfRawData_MSSQLSERVER_SQLServerBufferManager`](https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-buffer-manager-object)<br/>[`Win32_PerfRawData_MSSQLSERVER_SQLServerDatabaseReplica`](https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-database-replica)<br/>[`Win32_PerfRawData_MSSQLSERVER_SQLServerDatabases`](https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-databases-object?view=sql-server-2017)<br/>[`Win32_PerfRawData_MSSQLSERVER_SQLServerGeneralStatistics`](https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-general-statistics-object)<br/>[`Win32_PerfRawData_MSSQLSERVER_SQLServerLocks`](https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-locks-object)<br/>[`Win32_PerfRawData_MSSQLSERVER_SQLServerMemoryManager`](https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-memory-manager-object)<br/>[`Win32_PerfRawData_MSSQLSERVER_SQLServerSQLStatistics`](https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-sql-statistics-object)
Enabled by default? | No

## Flags

### `--collectors.mssql.classes-enabled`

Comma-separated list of MSSQL WMI classes to use. Supported values are `accessmethods`, `availreplica`, `bufman`, `databases`, `dbreplica`, `genstats`, `locks`, `memmgr` and `sqlstats`.

### `--collectors.mssql.class-print`

If true, print available mssql WMI classes and exit.  Only displays if the mssql collector is enabled.

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`wmi_mssql_collector_duration_seconds` | The time taken for each sub-collector to return | counter | `collector`, `instance`
`wmi_mssql_collector_success` | 1 if sub-collector succeeded, 0 otherwise | counter | `collector`, `instance`
`wmi_mssql_accessmethods_au_batch_cleanups` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_au_cleanups` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_by_reference_lob_creates` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_by_reference_lob_uses` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_lob_read_aheads` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_column_value_pulls` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_column_value_pushes` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_deferred_dropped_aus` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_deferred_dropped_rowsets` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_dropped_rowset_cleanups` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_dropped_rowset_skips` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_extent_deallocations` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_extent_allocations` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_au_batch_cleanup_failures` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_leaf_page_cookie_failures` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_tree_page_cookie_failures` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_forwarded_records` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_free_space_page_fetches` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_free_space_scans` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_full_scans` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_index_searches` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_insysxact_waits` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_lob_handle_creates` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_lob_handle_destroys` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_lob_ss_provider_creates` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_lob_ss_provider_destroys` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_lob_ss_provider_truncations` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_mixed_page_allocations` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_page_compression_attempts` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_page_deallocations` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_page_allocations` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_page_compressions` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_page_splits` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_probe_scans` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_range_scans` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_scan_point_revalidations` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_ghost_record_skips` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_table_lock_escalations` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_leaf_page_cookie_uses` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_tree_page_cookie_uses` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_workfile_creates` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_worktables_creates` | _Not yet documented_ | counter | `instance`
`wmi_mssql_accessmethods_worktables_from_cache_ratio` | _Not yet documented_ | counter | `instance`
`wmi_mssql_availreplica_received_from_replica_bytes` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_availreplica_sent_to_replica_bytes` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_availreplica_sent_to_transport_bytes` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_availreplica_initiated_flow_controls` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_availreplica_flow_control_wait_seconds` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_availreplica_receives_from_replica` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_availreplica_resent_messages` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_availreplica_sends_to_replica` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_availreplica_sends_to_transport` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_bufman_background_writer_pages` | _Not yet documented_ | counter | `instance`
`wmi_mssql_bufman_buffer_cache_hit_ratio` | _Not yet documented_ | counter | `instance`
`wmi_mssql_bufman_checkpoint_pages` | _Not yet documented_ | counter | `instance`
`wmi_mssql_bufman_database_pages` | _Not yet documented_ | counter | `instance`
`wmi_mssql_bufman_extension_allocated_pages` | _Not yet documented_ | counter | `instance`
`wmi_mssql_bufman_extension_free_pages` | _Not yet documented_ | counter | `instance`
`wmi_mssql_bufman_extension_in_use_as_percentage` | _Not yet documented_ | counter | `instance`
`wmi_mssql_bufman_extension_outstanding_io` | _Not yet documented_ | counter | `instance`
`wmi_mssql_bufman_extension_page_evictions` | _Not yet documented_ | counter | `instance`
`wmi_mssql_bufman_extension_page_reads` | _Not yet documented_ | counter | `instance`
`wmi_mssql_bufman_extension_page_unreferenced_seconds` | _Not yet documented_ | counter | `instance`
`wmi_mssql_bufman_extension_page_writes` | _Not yet documented_ | counter | `instance`
`wmi_mssql_bufman_free_list_stalls` | _Not yet documented_ | counter | `instance`
`wmi_mssql_bufman_integral_controller_slope` | _Not yet documented_ | counter | `instance`
`wmi_mssql_bufman_lazywrites` | _Not yet documented_ | counter | `instance`
`wmi_mssql_bufman_page_life_expectancy_seconds` | _Not yet documented_ | counter | `instance`
`wmi_mssql_bufman_page_lookups` | _Not yet documented_ | counter | `instance`
`wmi_mssql_bufman_page_reads` | _Not yet documented_ | counter | `instance`
`wmi_mssql_bufman_page_writes` | _Not yet documented_ | counter | `instance`
`wmi_mssql_bufman_read_ahead_pages` | _Not yet documented_ | counter | `instance`
`wmi_mssql_bufman_read_ahead_issuing_seconds` | _Not yet documented_ | counter | `instance`
`wmi_mssql_bufman_target_pages` | _Not yet documented_ | counter | `instance`
`wmi_mssql_dbreplica_database_flow_control_wait_seconds` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_dbreplica_database_initiated_flow_controls` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_dbreplica_received_file_bytes` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_dbreplica_group_commits` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_dbreplica_group_commit_stall_seconds` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_dbreplica_log_apply_pending_queue` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_dbreplica_log_apply_ready_queue` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_dbreplica_log_compressed_bytes` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_dbreplica_log_decompressed_bytes` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_dbreplica_log_received_bytes` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_dbreplica_log_compression_cachehits` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_dbreplica_log_compression_cachemisses` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_dbreplica_log_compressions` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_dbreplica_log_decompressions` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_dbreplica_log_remaining_for_undo` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_dbreplica_log_send_queue` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_dbreplica_mirrored_write_transactions` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_dbreplica_recovery_queue_records` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_dbreplica_redo_blocks` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_dbreplica_redo_remaining_bytes` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_dbreplica_redone_bytes` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_dbreplica_redones` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_dbreplica_total_log_requiring_undo` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_dbreplica_transaction_delay_seconds` | _Not yet documented_ | counter | `instance`, `replica`
`wmi_mssql_databases_active_transactions` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_backup_restore_operations` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_bulk_copy_rows` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_bulk_copy_bytes` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_commit_table_entries` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_data_files_size_bytes` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_dbcc_logical_scan_bytes` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_group_commit_stall_seconds` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_log_flushed_bytes` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_log_cache_hit_ratio` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_log_cache_reads` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_log_files_size_bytes` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_log_files_used_size_bytes` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_log_flushes` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_log_flush_waits` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_log_flush_wait_seconds` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_log_flush_write_seconds` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_log_growths` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_log_pool_cache_misses` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_log_pool_disk_reads` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_log_pool_hash_deletes` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_log_pool_hash_inserts` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_log_pool_invalid_hash_entries` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_log_pool_log_scan_pushes` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_log_pool_log_writer_pushes` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_log_pool_empty_free_pool_pushes` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_log_pool_low_memory_pushes` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_log_pool_no_free_buffer_pushes` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_log_pool_req_behind_trunc` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_log_pool_requests_old_vlf` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_log_pool_requests` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_log_pool_total_active_log_bytes` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_log_pool_total_shared_pool_bytes` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_log_shrinks` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_log_truncations` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_log_used_percent` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_pending_repl_transactions` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_repl_transactions` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_shrink_data_movement_bytes` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_tracked_transactions` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_transactions` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_write_transactions` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_xtp_controller_dlc_fetch_latency_seconds` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_xtp_controller_dlc_peak_latency_seconds` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_xtp_controller_log_processed_bytes` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_databases_xtp_memory_used_bytes` | _Not yet documented_ | counter | `instance`, `database`
`wmi_mssql_genstats_active_temp_tables` | _Not yet documented_ | counter | `instance`
`wmi_mssql_genstats_connection_resets` | _Not yet documented_ | counter | `instance`
`wmi_mssql_genstats_event_notifications_delayed_drop` | _Not yet documented_ | counter | `instance`
`wmi_mssql_genstats_http_authenticated_requests` | _Not yet documented_ | counter | `instance`
`wmi_mssql_genstats_logical_connections` | _Not yet documented_ | counter | `instance`
`wmi_mssql_genstats_logins` | _Not yet documented_ | counter | `instance`
`wmi_mssql_genstats_logouts` | _Not yet documented_ | counter | `instance`
`wmi_mssql_genstats_mars_deadlocks` | _Not yet documented_ | counter | `instance`
`wmi_mssql_genstats_non_atomic_yields` | _Not yet documented_ | counter | `instance`
`wmi_mssql_genstats_blocked_processes` | _Not yet documented_ | counter | `instance`
`wmi_mssql_genstats_soap_empty_requests` | _Not yet documented_ | counter | `instance`
`wmi_mssql_genstats_soap_method_invocations` | _Not yet documented_ | counter | `instance`
`wmi_mssql_genstats_soap_session_initiate_requests` | _Not yet documented_ | counter | `instance`
`wmi_mssql_genstats_soap_session_terminate_requests` | _Not yet documented_ | counter | `instance`
`wmi_mssql_genstats_soapsql_requests` | _Not yet documented_ | counter | `instance`
`wmi_mssql_genstats_soapwsdl_requests` | _Not yet documented_ | counter | `instance`
`wmi_mssql_genstats_sql_trace_io_provider_lock_waits` | _Not yet documented_ | counter | `instance`
`wmi_mssql_genstats_tempdb_recovery_unit_ids_generated` | _Not yet documented_ | counter | `instance`
`wmi_mssql_genstats_tempdb_rowset_ids_generated` | _Not yet documented_ | counter | `instance`
`wmi_mssql_genstats_temp_tables_creations` | _Not yet documented_ | counter | `instance`
`wmi_mssql_genstats_temp_tables_awaiting_destruction` | _Not yet documented_ | counter | `instance`
`wmi_mssql_genstats_trace_event_notification_queue_size` | _Not yet documented_ | counter | `instance`
`wmi_mssql_genstats_transactions` | _Not yet documented_ | counter | `instance`
`wmi_mssql_genstats_user_connections` | _Not yet documented_ | counter | `instance`
`wmi_mssql_locks_average_wait_seconds` | _Not yet documented_ | counter | `instance`, `resource`
`wmi_mssql_locks_lock_requests` | _Not yet documented_ | counter | `instance`, `resource`
`wmi_mssql_locks_lock_timeouts` | _Not yet documented_ | counter | `instance`, `resource`
`wmi_mssql_locks_lock_timeouts_excluding_NOWAIT` | _Not yet documented_ | counter | `instance`, `resource`
`wmi_mssql_locks_lock_waits` | _Not yet documented_ | counter | `instance`, `resource`
`wmi_mssql_locks_lock_wait_seconds` | _Not yet documented_ | counter | `instance`, `resource`
`wmi_mssql_locks_deadlocks` | _Not yet documented_ | counter | `instance`, `resource`
`wmi_mssql_memmgr_connection_memory_bytes` | _Not yet documented_ | counter | `instance`
`wmi_mssql_memmgr_database_cache_memory_bytes` | _Not yet documented_ | counter | `instance`
`wmi_mssql_memmgr_external_benefit_of_memory` | _Not yet documented_ | counter | `instance`
`wmi_mssql_memmgr_free_memory_bytes` | _Not yet documented_ | counter | `instance`
`wmi_mssql_memmgr_granted_workspace_memory_bytes` | _Not yet documented_ | counter | `instance`
`wmi_mssql_memmgr_lock_blocks` | _Not yet documented_ | counter | `instance`
`wmi_mssql_memmgr_allocated_lock_blocks` | _Not yet documented_ | counter | `instance`
`wmi_mssql_memmgr_lock_memory_bytes` | _Not yet documented_ | counter | `instance`
`wmi_mssql_memmgr_lock_owner_blocks` | _Not yet documented_ | counter | `instance`
`wmi_mssql_memmgr_allocated_lock_owner_blocks` | _Not yet documented_ | counter | `instance`
`wmi_mssql_memmgr_log_pool_memory_bytes` | _Not yet documented_ | counter | `instance`
`wmi_mssql_memmgr_maximum_workspace_memory_bytes` | _Not yet documented_ | counter | `instance`
`wmi_mssql_memmgr_outstanding_memory_grants` | _Not yet documented_ | counter | `instance`
`wmi_mssql_memmgr_pending_memory_grants` | _Not yet documented_ | counter | `instance`
`wmi_mssql_memmgr_optimizer_memory_bytes` | _Not yet documented_ | counter | `instance`
`wmi_mssql_memmgr_reserved_server_memory_bytes` | _Not yet documented_ | counter | `instance`
`wmi_mssql_memmgr_sql_cache_memory_bytes` | _Not yet documented_ | counter | `instance`
`wmi_mssql_memmgr_stolen_server_memory_bytes` | _Not yet documented_ | counter | `instance`
`wmi_mssql_memmgr_target_server_memory_bytes` | _Not yet documented_ | counter | `instance`
`wmi_mssql_memmgr_total_server_memory_bytes` | _Not yet documented_ | counter | `instance`
`wmi_mssql_sqlstats_auto_parameterization_attempts` | _Not yet documented_ | counter | `instance`
`wmi_mssql_sqlstats_batch_requests` | _Not yet documented_ | counter | `instance`
`wmi_mssql_sqlstats_failed_auto_parameterization_attempts` | _Not yet documented_ | counter | `instance`
`wmi_mssql_sqlstats_forced_parameterizations` | _Not yet documented_ | counter | `instance`
`wmi_mssql_sqlstats_guided_plan_executions` | _Not yet documented_ | counter | `instance`
`wmi_mssql_sqlstats_misguided_plan_executions` | _Not yet documented_ | counter | `instance`
`wmi_mssql_sqlstats_safe_auto_parameterization_attempts` | _Not yet documented_ | counter | `instance`
`wmi_mssql_sqlstats_sql_attentions` | _Not yet documented_ | counter | `instance`
`wmi_mssql_sqlstats_sql_compilations` | _Not yet documented_ | counter | `instance`
`wmi_mssql_sqlstats_sql_recompilations` | _Not yet documented_ | counter | `instance`
`wmi_mssql_sqlstats_unsafe_auto_parameterization_attempts` | _Not yet documented_ | counter | `instance`

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

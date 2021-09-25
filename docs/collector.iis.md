# iis collector

The iis collector exposes metrics about the IIS server

|||
-|-
Metric name prefix  | `iis`
Data source         | Perflib
Enabled by default? | No

## Flags

### `--collector.iis.site-whitelist`

If given, a site needs to match the whitelist regexp in order for the corresponding metrics to be reported.

### `--collector.iis.site-blacklist`

If given, a site needs to *not* match the blacklist regexp in order for the corresponding metrics to be reported.

### `--collector.iis.app-whitelist`

If given, an application needs to match the whitelist regexp in order for the corresponding metrics to be reported.

### `--collector.iis.app-blacklist`

If given, an application needs to *not* match the blacklist regexp in order for the corresponding metrics to be reported.

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_iis_current_anonymous_users` | _Not yet documented_ | counter | `site`
`windows_iis_current_blocked_async_io_requests` | _Not yet documented_ | counter | `site`
`windows_iis_current_cgi_requests` | _Not yet documented_ | counter | `site`
`windows_iis_current_connections` | _Not yet documented_ | counter | `site`
`windows_iis_current_isapi_extension_requests` | _Not yet documented_ | counter | `site`
`windows_iis_current_non_anonymous_users` | _Not yet documented_ | counter | `site`
`windows_iis_received_bytes_total` | _Not yet documented_ | counter | `site`
`windows_iis_sent_bytes_total` | _Not yet documented_ | counter | `site`
`windows_iis_anonymous_users_total` | _Not yet documented_ | counter | `site`
`windows_iis_blocked_async_io_requests_total` | _Not yet documented_ | counter | `site`
`windows_iis_cgi_requests_total` | _Not yet documented_ | counter | `site`
`windows_iis_connection_attempts_all_instances_total` | _Not yet documented_ | counter | `site`
`windows_iis_requests_total` | _Not yet documented_ | counter | `site`, `method`
`windows_iis_files_received_total` | _Not yet documented_ | counter | `site`
`windows_iis_files_sent_total` | _Not yet documented_ | counter | `site`
`windows_iis_ipapi_extension_requests_total` | _Not yet documented_ | counter | `site`
`windows_iis_locked_errors_total` | _Not yet documented_ | counter | `site`
`windows_iis_logon_attempts_total` | _Not yet documented_ | counter | `site`
`windows_iis_non_anonymous_users_total` | _Not yet documented_ | counter | `site`
`windows_iis_not_found_errors_total` | _Not yet documented_ | counter | `site`
`windows_iis_rejected_async_io_requests_total` | _Not yet documented_ | counter | `site`
`windows_iis_current_application_pool_state` | _Not yet documented_ | counter | `app`, `state`
`windows_iis_current_application_pool_start_time` | _Not yet documented_ | counter | `app`
`windows_iis_current_worker_processes` | _Not yet documented_ | counter | `app`
`windows_iis_maximum_worker_processes` | _Not yet documented_ | counter | `app`
`windows_iis_recent_worker_process_failures` | _Not yet documented_ | counter | `app`
`windows_iis_time_since_last_worker_process_failure` | _Not yet documented_ | counter | `app`
`windows_iis_total_application_pool_recycles` | _Not yet documented_ | counter | `app`
`windows_iis_total_application_pool_start_time` | _Not yet documented_ | counter | `app`
`windows_iis_total_worker_processes_created` | _Not yet documented_ | counter | `app`
`windows_iis_total_worker_process_failures` | _Not yet documented_ | counter | `app`
`windows_iis_total_worker_process_ping_failures` | _Not yet documented_ | counter | `app`
`windows_iis_total_worker_process_shutdown_failures` | _Not yet documented_ | counter | `app`
`windows_iis_total_worker_process_startup_failures` | _Not yet documented_ | counter | `app`
`windows_iis_worker_cache_active_flushed_entries` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_file_cache_memory_bytes` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_file_cache_max_memory_bytes` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_file_cache_flushes_total` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_file_cache_queries_total` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_file_cache_hits_total` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_file_cache_items` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_file_cache_items_total` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_file_cache_items_flushed_total` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_uri_cache_flushes_total` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_uri_cache_queries_total` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_uri_cache_hits_total` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_uri_cache_items` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_uri_cache_items_total` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_uri_cache_items_flushed_total` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_metadata_cache_items` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_metadata_cache_flushes_total` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_metadata_cache_queries_total` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_metadata_cache_hits_total` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_metadata_cache_items_cached_total` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_metadata_cache_items_flushed_total` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_output_cache_active_flushed_items` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_output_cache_items` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_output_cache_memory_bytes` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_output_queries_total` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_output_cache_hits_total` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_output_cache_items_flushed_total` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_output_cache_flushes_total` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_threads` | _Not yet documented_ | counter | `app`, `pid`, `state`
`windows_iis_worker_max_threads` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_requests_total` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_current_requests` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_request_errors_total` | _Not yet documented_ | counter | `app`, `pid`, `status_code`
`windows_iis_worker_current_websocket_requests` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_websocket_connection_attempts_total` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_websocket_connection_accepted_total` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_websocket_connection_rejected_total` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_server_cache_active_flushed_entries` | _Not yet documented_ | counter | None
`windows_iis_server_file_cache_memory_bytes` | _Not yet documented_ | counter | None
`windows_iis_server_file_cache_max_memory_bytes` | _Not yet documented_ | counter | None
`windows_iis_server_file_cache_flushes_total` | _Not yet documented_ | counter | None
`windows_iis_server_file_cache_queries_total` | _Not yet documented_ | counter | None
`windows_iis_server_file_cache_hits_total` | _Not yet documented_ | counter | None
`windows_iis_server_file_cache_items` | _Not yet documented_ | counter | None
`windows_iis_server_file_cache_items_total` | _Not yet documented_ | counter | None
`windows_iis_server_file_cache_items_flushed_total` | _Not yet documented_ | counter | None
`windows_iis_server_uri_cache_flushes_total` | _Not yet documented_ | counter | `mode`
`windows_iis_server_uri_cache_queries_total` | _Not yet documented_ | counter | `mode`
`windows_iis_server_uri_cache_hits_total` | _Not yet documented_ | counter | `mode`
`windows_iis_server_uri_cache_items` | _Not yet documented_ | counter | `mode`
`windows_iis_server_uri_cache_items_total` | _Not yet documented_ | counter | `mode`
`windows_iis_server_uri_cache_items_flushed_total` | _Not yet documented_ | counter | `mode`
`windows_iis_server_metadata_cache_items` | _Not yet documented_ | counter | None
`windows_iis_server_metadata_cache_flushes_total` | _Not yet documented_ | counter | None
`windows_iis_server_metadata_cache_queries_total` | _Not yet documented_ | counter | None
`windows_iis_server_metadata_cache_hits_total` | _Not yet documented_ | counter | None
`windows_iis_server_metadata_cache_items_cached_total` | _Not yet documented_ | counter | None
`windows_iis_server_metadata_cache_items_flushed_total` | _Not yet documented_ | counter | None
`windows_iis_server_output_cache_active_flushed_items` | _Not yet documented_ | counter | None
`windows_iis_server_output_cache_items` | _Not yet documented_ | counter | None
`windows_iis_server_output_cache_memory_bytes` | _Not yet documented_ | counter | None
`windows_iis_server_output_cache_queries_total` | _Not yet documented_ | counter | None
`windows_iis_server_output_cache_hits_total` | _Not yet documented_ | counter | None
`windows_iis_server_output_cache_items_flushed_total` | _Not yet documented_ | counter | None
`windows_iis_server_output_cache_flushes_total` | _Not yet documented_ | counter | None

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

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
`windows_iis_current_anonymous_users` | The number of users who currently have an anonymous request pending with the web service | counter | `site`
`windows_iis_current_blocked_async_io_requests` | _Not yet documented_ | counter | `site`
`windows_iis_current_cgi_requests` | The number of CGI requests that are being processed simultaneously by the web service | counter | `site`
`windows_iis_current_connections` |  The number of active connections to the web service | counter | `site`
`windows_iis_current_isapi_extension_requests` | The number of [ISAPI extension](https://docs.microsoft.com/en-us/previous-versions/iis/6.0-sdk/ms525172(v=vs.90)) requests that are being processed simultaneously by the web service | counter | `site`
`windows_iis_current_non_anonymous_users` | The number of users who currently have a nonanonymous request pending with the web service | counter | `site`
`windows_iis_service_uptime` | The uptime for the web service or a Web site (seconds) | gauge | `site`
`windows_iis_received_bytes_total` | The total bytes of data that have been received by the web service since the service started | counter | `site`
`windows_iis_sent_bytes_total` | The number of data bytes that have been sent by the web service since the service started | counter | `site`
`windows_iis_anonymous_users_total` | The number of users who have established an anonymous request since the web service started | counter | `site`
`windows_iis_blocked_async_io_requests_total` | _Not yet documented_ | counter | `site`
`windows_iis_cgi_requests_total` | The number of all CGI requests that have been made since the web service started | counter | `site`
`windows_iis_connection_attempts_all_instances_total` | The number of connections to the web service that have been attempted since the service started | counter | `site`
`windows_iis_requests_total` | The number of requests that have been made since the web service was started | counter | `site`, `method`
`windows_iis_files_received_total` | The total number of files that have been received by the FTP service since the service started | counter | `site`
`windows_iis_files_sent_total` | The total number of files that have been sent by the FTP service since the service started | counter | `site`
`windows_iis_ipapi_extension_requests_total` | The number of ISAPI extension requests that have been made since the web service started | counter | `site`
`windows_iis_locked_errors_total` | The number of requests that have been made since the service started that could not be satisfied by the server because the requested document was locked. Usually reported as HTTP error 423 | counter | `site`
`windows_iis_logon_attempts_total` | The number of attempts to log on to the web service that have occurred since the service started | counter | `site`
`windows_iis_non_anonymous_users_total` | The number of users who have made nonanonymous requests to the web service since the service started | counter | `site`
`windows_iis_not_found_errors_total` | The number of requests that have been made since the service started that were not satisfied by the server because the requested document was not found. Usually reported as HTTP error 404 | counter | `site`
`windows_iis_rejected_async_io_requests_total` | _Not yet documented_ | counter | `site`
`windows_iis_current_application_pool_state` | The current status of the application pool (1 - Uninitialized, 2 - Initialized, 3 - Running, 4 - Disabling, 5 - Disabled, 6 - Shutdown Pending, 7 - Delete Pending) | counter | `app`, `state`
`windows_iis_current_application_pool_start_time` | The unix timestamp for the application pool start time | counter | `app`
`windows_iis_current_worker_processes` | The current number of worker processes that are running in the application pool | counter | `app`
`windows_iis_maximum_worker_processes` | The maximum number of worker processes that have been created for the application pool since Windows Process Activation Service (WAS) started | counter | `app`
`windows_iis_recent_worker_process_failures` | The number of times that worker processes for the application pool failed during the rapid-fail protection interval | counter | `app`
`windows_iis_time_since_last_worker_process_failure` | The length of time, in seconds, since the last worker process failure occurred for the application pool | counter | `app`
`windows_iis_total_application_pool_recycles` | The number of times that the application pool has been recycled since Windows Process Activation Service (WAS) started | counter | `app`
`windows_iis_total_application_pool_start_time` | The unix timestamp for the application pool of when the Windows Process Activation Service (WAS) started | counter | `app`
`windows_iis_total_worker_processes_created` | The number of worker processes created for the application pool since Windows Process Activation Service (WAS) started | counter | `app`
`windows_iis_total_worker_process_failures` | The number of times that worker processes have crashed since the application pool was started | counter | `app`
`windows_iis_total_worker_process_ping_failures` | The number of times that Windows Process Activation Service (WAS) did not receive a response to ping messages sent to a worker process | counter | `app`
`windows_iis_total_worker_process_shutdown_failures` | The number of times that Windows Process Activation Service (WAS) failed to shut down a worker process | counter | `app`
`windows_iis_total_worker_process_startup_failures` | The number of times that Windows Process Activation Service (WAS) failed to start a worker process | counter | `app`
`windows_iis_worker_cache_active_flushed_entries` | Number of file handles cached that will be closed when all current transfers complete | counter | `app`, `pid`
`windows_iis_worker_file_cache_memory_bytes` | Current number of bytes used by file cache | counter | `app`, `pid`
`windows_iis_worker_file_cache_max_memory_bytes` | Maximum number of bytes used by file cache | counter | `app`, `pid`
`windows_iis_worker_file_cache_flushes_total` | Total number of file cache flushes (since service startup) | counter | `app`, `pid`
`windows_iis_worker_file_cache_queries_total` | Total file cache queries (hits + misses) | counter | `app`, `pid`
`windows_iis_worker_file_cache_hits_total` | Total number of successful lookups in the user-mode file cache | counter | `app`, `pid`
`windows_iis_worker_file_cache_items` | Current number of files whose contents are present in user-mode cache | counter | `app`, `pid`
`windows_iis_worker_file_cache_items_total` | Total number of files whose contents were ever added to the user-mode cache (since service startup) | counter | `app`, `pid`
`windows_iis_worker_file_cache_items_flushed_total` | Total number of file handles that have been removed from the user-mode cache (since service startup) | counter | `app`, `pid`
`windows_iis_worker_uri_cache_flushes_total` | Total number of URI cache flushes (since service startup) | counter | `app`, `pid`
`windows_iis_worker_uri_cache_queries_total` | Total number of uri cache queries (hits + misses) | counter | `app`, `pid`
`windows_iis_worker_uri_cache_hits_total` | Total number of successful lookups in the user-mode URI cache (since service startup) | counter | `app`, `pid`
`windows_iis_worker_uri_cache_items` | Number of URI information blocks currently in the user-mode cache | counter | `app`, `pid`
`windows_iis_worker_uri_cache_items_total` | Total number of URI information blocks added to the user-mode cache (since service startup) | counter | `app`, `pid`
`windows_iis_worker_uri_cache_items_flushed_total` | The number of URI information blocks that have been removed from the user-mode cache (since service startup) | counter | `app`, `pid`
`windows_iis_worker_metadata_cache_items` | Number of metadata information blocks currently present in user-mode cache | counter | `app`, `pid`
`windows_iis_worker_metadata_cache_flushes_total` | Total number of user-mode metadata cache flushes (since service startup) | counter | `app`, `pid`
`windows_iis_worker_metadata_cache_queries_total` | Total metadata cache queries (hits + misses) | counter | `app`, `pid`
`windows_iis_worker_metadata_cache_hits_total` | Total number of successful lookups in the user-mode metadata cache (since service startup) | counter | `app`, `pid`
`windows_iis_worker_metadata_cache_items_cached_total` | Total number of metadata information blocks added to the user-mode cache (since service startup) | counter | `app`, `pid`
`windows_iis_worker_metadata_cache_items_flushed_total` | Total number of metadata information blocks removed from the user-mode cache (since service startup) | counter | `app`, `pid`
`windows_iis_worker_output_cache_active_flushed_items` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_output_cache_items` | Number of items current present in output cache | counter | `app`, `pid`
`windows_iis_worker_output_cache_memory_bytes` | Current number of bytes used by output cache | counter | `app`, `pid`
`windows_iis_worker_output_queries_total` | Total number of output cache queries (hits + misses) | counter | `app`, `pid`
`windows_iis_worker_output_cache_hits_total` | Total number of successful lookups in output cache (since service startup) | counter | `app`, `pid`
`windows_iis_worker_output_cache_items_flushed_total` | Total number of items flushed from output cache (since service startup) | counter | `app`, `pid`
`windows_iis_worker_output_cache_flushes_total` | Total number of flushes of output cache (since service startup) | counter | `app`, `pid`
`windows_iis_worker_threads` | Number of threads actively processing requests in the worker process | counter | `app`, `pid`, `state`
`windows_iis_worker_max_threads` | Maximum number of threads to which the thread pool can grow as needed | counter | `app`, `pid`
`windows_iis_worker_requests_total` | Total number of HTTP requests served by the worker process | counter | `app`, `pid`
`windows_iis_worker_current_requests` | Current number of requests being processed by the worker process | counter | `app`, `pid`
`windows_iis_worker_request_errors_total` | Total number of requests that returned an error | counter | `app`, `pid`, `status_code`
`windows_iis_worker_current_websocket_requests` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_websocket_connection_attempts_total` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_websocket_connection_accepted_total` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_worker_websocket_connection_rejected_total` | _Not yet documented_ | counter | `app`, `pid`
`windows_iis_server_cache_active_flushed_entries` | Number of file handles cached that will be closed when all current transfers complete | counter | None
`windows_iis_server_file_cache_memory_bytes` | Current number of bytes used by file cache | counter | None
`windows_iis_server_file_cache_max_memory_bytes` | Maximum number of bytes used by file cache | counter | None
`windows_iis_server_file_cache_flushes_total` | Total number of file cache flushes (since service startup) | counter | None
`windows_iis_server_file_cache_queries_total` | Total number of file cache queries (hits + misses) | counter | None
`windows_iis_server_file_cache_hits_total` | Total number of successful lookups in the file cache | counter | None
`windows_iis_server_file_cache_items` | Current number of files whose contents are present in cache | counter | None
`windows_iis_server_file_cache_items_total` | Total number of files whose contents were ever added to the cache (since service startup) | counter | None
`windows_iis_server_file_cache_items_flushed_total` | Total number of file handles that have been removed from the cache (since service startup) | counter | None
`windows_iis_server_uri_cache_flushes_total` | Total number of URI cache flushes (since service startup) | counter | `mode`
`windows_iis_server_uri_cache_queries_total` | Total number of uri cache queries (hits + misses) | counter | `mode`
`windows_iis_server_uri_cache_hits_total` | Total number of successful lookups in the URI cache (since service startup) | counter | `mode`
`windows_iis_server_uri_cache_items` | Number of URI information blocks currently in the cache | counter | `mode`
`windows_iis_server_uri_cache_items_total` | Total number of URI information blocks added to the cache (since service startup) | counter | `mode`
`windows_iis_server_uri_cache_items_flushed_total` | The number of URI information blocks that have been removed from the cache (since service startup) | counter | `mode`
`windows_iis_server_metadata_cache_items` | Number of metadata information blocks currently present in cache | counter | None
`windows_iis_server_metadata_cache_flushes_total` | Total number of metadata cache flushes (since service startup) | counter | None
`windows_iis_server_metadata_cache_queries_total` | Total metadata cache queries (hits + misses) | counter | None
`windows_iis_server_metadata_cache_hits_total` | Total number of successful lookups in the metadata cache (since service startup) | counter | None
`windows_iis_server_metadata_cache_items_cached_total` | Total number of metadata information blocks added to the cache (since service startup) | counter | None
`windows_iis_server_metadata_cache_items_flushed_total` | Total number of metadata information blocks removed from the cache (since service startup) | counter | None
`windows_iis_server_output_cache_active_flushed_items` | _Not yet documented_ | counter | None
`windows_iis_server_output_cache_items` | Number of items current present in output cache | counter | None
`windows_iis_server_output_cache_memory_bytes` | Current number of bytes used by output cache | counter | None
`windows_iis_server_output_cache_queries_total` | Total output cache queries (hits + misses) | counter | None
`windows_iis_server_output_cache_hits_total` | Total number of successful lookups in output cache (since service startup) | counter | None
`windows_iis_server_output_cache_items_flushed_total` | Total number of items flushed from output cache (since service startup) | counter | None
`windows_iis_server_output_cache_flushes_total` | Total number of flushes of output cache (since service startup) | counter | None

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

# process collector

The process collector exposes metrics about processes

|||
-|-
Metric name prefix  | `process`
Data source         | Perflib
Counters            | `Process`
Enabled by default? | No

## Flags

### `--collector.process.whitelist`

Regexp of processes to include. Process name must both match whitelist and not
match blacklist to be included. Recommended to keep down number of returned
metrics.

### `--collector.process.blacklist`

Regexp of processes to exclude. Process name must both match whitelist and not
match blacklist to be included. Recommended to keep down number of returned
metrics.

### Example
To match all firefox processes: `--collector.process.whitelist="firefox.+"`.
Note that multiple processes with the same name will be disambiguated by
Windows by adding a number suffix, such as `firefox#2`. Your regexp must take
these suffixes into consideration.

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`wmi_process_start_time` | _Not yet documented_ | gauge | `process`, `process_id`, `creating_process_id`
`wmi_process_cpu_time_total` | _Not yet documented_ | counter | `process`, `process_id`, `creating_process_id`
`wmi_process_handle_count` | _Not yet documented_ | gauge | `process`, `process_id`, `creating_process_id`
`wmi_process_io_bytes_total` | _Not yet documented_ | counter | `process`, `process_id`, `creating_process_id`
`wmi_process_io_operations_total` | _Not yet documented_ | counter | `process`, `process_id`, `creating_process_id`
`wmi_process_page_faults_total` | _Not yet documented_ | counter | `process`, `process_id`, `creating_process_id`
`wmi_process_page_file_bytes` | _Not yet documented_ | gauge | `process`, `process_id`, `creating_process_id`
`wmi_process_pool_bytes` | _Not yet documented_ | gauge | `process`, `process_id`, `creating_process_id`
`wmi_process_priority_base` | _Not yet documented_ | gauge | `process`, `process_id`, `creating_process_id`
`wmi_process_private_bytes` | _Not yet documented_ | gauge | `process`, `process_id`, `creating_process_id`
`wmi_process_thread_count` | _Not yet documented_ | gauge | `process`, `process_id`, `creating_process_id`
`wmi_process_virtual_bytes` | _Not yet documented_ | gauge | `process`, `process_id`, `creating_process_id`
`wmi_process_working_set` | _Not yet documented_ | gauge | `process`, `process_id`, `creating_process_id`

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

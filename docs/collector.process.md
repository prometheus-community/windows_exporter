# process collector

The process collector exposes metrics about processes

|||
-|-
Metric name prefix  | `process`
Classes             | [`Win32_PerfRawData_PerfProc_Process`](https://msdn.microsoft.com/en-us/library/aa394323(v=vs.85).aspx)
Enabled by default? | No

## Flags

### `--collector.process.processes-where`

A WMI filter on which processes to include. Recommended to keep down number of returned metrics.

`%` is a wildcard, and can be used to match on substrings.

Example: `--collector.process.processes-where="Name LIKE 'firefox%'`

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

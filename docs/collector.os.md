# os collector

The os collector exposes metrics about the operating system

|||
-|-
Metric name prefix  | `os`
Classes             | [`Win32_OperatingSystem`](https://msdn.microsoft.com/en-us/library/aa394239)
Enabled by default? | Yes

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`wmi_os_paging_limit_bytes` | _Not yet documented_ | gauge | None
`wmi_os_paging_free_bytes` | _Not yet documented_ | gauge | None
`wmi_os_physical_memory_free_bytes` | _Not yet documented_ | gauge | None
`wmi_os_time` | _Not yet documented_ | gauge | None
`wmi_os_timezone` | _Not yet documented_ | gauge | `timezone`
`wmi_os_processes` | _Not yet documented_ | gauge | None
`wmi_os_processes_limit` | _Not yet documented_ | gauge | None
`wmi_os_process_memory_limix_bytes` | _Not yet documented_ | gauge | None
`wmi_os_users` | _Not yet documented_ | gauge | None
`wmi_os_virtual_memory_bytes` | _Not yet documented_ | gauge | None
`wmi_os_visible_memory_bytes` | _Not yet documented_ | gauge | None
`wmi_os_virtual_memory_free_bytes` | _Not yet documented_ | gauge | None

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

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
`wmi_os_info` | Contains full product name & version in labels | gauge | `product`, `version`
`wmi_os_paging_limit_bytes` | Total number of bytes that can be sotred in the operating system paging files. 0 (zero) indicates that there are no paging files | gauge | None
`wmi_os_paging_free_bytes` | Number of bytes that can be mapped into the operating system paging files without causing any other pages to be swapped out | gauge | None
`wmi_os_physical_memory_free_bytes` | Bytes of physical memory currently unused and available | gauge | None
`wmi_os_time` | Current time as reported by the operating system, in [Unix time](https://en.wikipedia.org/wiki/Unix_time). See [time.Unix()](https://golang.org/pkg/time/#Unix) for details | gauge | None
`wmi_os_timezone` | Current timezone as reported by the operating system. See [time.Zone()](https://golang.org/pkg/time/#Time.Zone) for details | gauge | `timezone`
`wmi_os_processes` | Number of process contexts currently loaded or running on the operating system | gauge | None
`wmi_os_processes_limit` | Maximum number of process contexts the operating system can support. The default value set by the provider is 4294967295 (0xFFFFFFFF) | gauge | None
`wmi_os_process_memory_limit_bytes` | Maximum number of bytes of memory that can be allocated to a process | gauge | None
`wmi_os_users` | Number of user sessions for which the operating system is storing state information currently. For a list of current active logon sessions, see [`logon`](collector.logon.md) | gauge | None
`wmi_os_virtual_memory_bytes` | Bytes of virtual memory | gauge | None
`wmi_os_visible_memory_bytes` | Total bytes of physical memory available to the operating system. This value does not necessarily indicate the true amount of physical memory, but what is reported to the operating system as available to it | gauge | None
`wmi_os_virtual_memory_free_bytes` | Bytes of virtual memory currently unused and available | gauge | None

### Example metric
Show current number of processes
```
wmi_os_processes{instance="localhost"}
```

## Useful queries
Find all devices not set to UTC timezone
```
wmi_os_timezone{timezone != "UTC"}
```

## Alerting examples
**prometheus.rules**
```
# Alert on hosts that have exhausted all available physical memory
- alert: MemoryExhausted
  expr: wmi_os_physical_memory_free_bytes == 0
  for: 10m
  labels:
    severity: high
  annotations:
    summary: "Host {{ $labels.instance }} is out of memory"
    description: "{{ $labels.instance }} has exhausted all available physical memory"

# Alert on hosts with greater than 90% memory usage
- alert: MemoryLow
  expr: 100 - 100 * wmi_os_physical_memory_free_bytes / wmi_cs_physical_memory_bytes > 90
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "Memory usage for host {{ $labels.instance }} is greater than 90%"
```

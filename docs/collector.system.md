# system collector

The system collector exposes metrics about ...

|||
-|-
Metric name prefix  | `system`
Data source         | Perflib
Classes             | [`Win32_PerfRawData_PerfOS_System`](https://web.archive.org/web/20050830140516/http://msdn.microsoft.com/library/en-us/wmisdk/wmi/win32_perfrawdata_perfos_system.asp)
Enabled by default? | Yes

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_system_context_switches_total` | Total number of [context switches](https://en.wikipedia.org/wiki/Context_switch) | counter | None
`windows_system_exception_dispatches_total` | Total exceptions dispatched by the system | counter | None
`windows_system_processor_queue_length` | Number of threads in the processor queue. There is a single queue for processor time even on computers with multiple processors. | gauge | None
`windows_system_system_calls_total` | Total combined calls to Windows NT system service routines by all processes running on the computer | counter | None
`windows_system_system_up_time` | Time of last boot of system | gauge | None
`windows_system_threads` | Number of Windows system [threads](https://en.wikipedia.org/wiki/Thread_(computing)) | gauge | None

### Example metric
Show current number of system threads
```
windows_system_threads{instance="localhost"}
```

## Useful queries
Find hosts that have rebooted in the last 24 hours
```
time() - windows_system_system_up_time < 86400
```

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

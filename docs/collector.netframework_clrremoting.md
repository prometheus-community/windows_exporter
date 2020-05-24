# netframework_clrremoting collector

The netframework_clrremoting collector exposes metrics about dotnet remoting.

|||
-|-
Metric name prefix  | `netframework_clrremoting`
Classes             | `Win32_PerfRawData_NETFramework_NETCLRRemoting`
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_netframework_clrremoting_channels_total` | Displays the total number of remoting channels registered across all application domains since application started. | counter | `process`
`windows_netframework_clrremoting_context_bound_classes_loaded` | Displays the current number of context-bound classes that are loaded. | gauge | `process`
`windows_netframework_clrremoting_context_bound_objects_total` | Displays the total number of context-bound objects allocated. | counter | `process`
`windows_netframework_clrremoting_context_proxies_total` | Displays the total number of remoting proxy objects in this process since it started. | counter | `process`
`windows_netframework_clrremoting_contexts` | Displays the current number of remoting contexts in the application. | gauge | `process`
`windows_netframework_clrremoting_remote_calls_total` | Displays the total number of remote procedure calls invoked since the application started. | counter | `process`

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

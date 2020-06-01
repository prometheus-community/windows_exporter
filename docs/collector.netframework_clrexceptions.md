# netframework_clrexceptions collector

The netframework_clrexceptions collector exposes metrics about CLR exceptions in the dotnet framework.

|||
-|-
Metric name prefix  | `netframework_clrexceptions`
Classes             | `Win32_PerfRawData_NETFramework_NETCLRExceptions`
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_netframework_clrexceptions_exceptions_thrown_total` | Displays the total number of exceptions thrown since the application started. This includes both .NET exceptions and unmanaged exceptions that are converted into .NET exceptions. | counter | `process`
`windows_netframework_clrexceptions_exceptions_filters_total` | Displays the total number of .NET exception filters executed. An exception filter evaluates regardless of whether an exception is handled. | counter | `process`
`windows_netframework_clrexceptions_exceptions_finallys_total` | Displays the total number of finally blocks executed. Only the finally blocks executed for an exception are counted; finally blocks on normal code paths are not counted by this counter. | counter | `process`
`windows_netframework_clrexceptions_throw_to_catch_depth_total` | Displays the total number of stack frames traversed, from the frame that threw the exception to the frame that handled the exception. | counter | `process`

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

# netframework_clrinterop collector

The netframework_clrinterop collector exposes metrics about interop between the dotnet framework and outside components.

|||
-|-
Metric name prefix  | `netframework_clrinterop`
Classes             | `Win32_PerfRawData_NETFramework_NETCLRInterop`
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_netframework_clrinterop_com_callable_wrappers_total` | Displays the current number of COM callable wrappers (CCWs). A CCW is a proxy for a managed object being referenced from an unmanaged COM client. | counter | `process`
`windows_netframework_clrinterop_interop_marshalling_total` | Displays the total number of times arguments and return values have been marshaled from managed to unmanaged code, and vice versa, since the application started. | counter | `process`
`windows_netframework_clrinterop_interop_stubs_created_total` | Displays the current number of stubs created by the common language runtime. Stubs are responsible for marshaling arguments and return values from managed to unmanaged code, and vice versa, during a COM interop call or a platform invoke call. | counter | `process`

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

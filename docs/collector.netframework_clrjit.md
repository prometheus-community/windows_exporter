# netframework_clrjit collector

The netframework_clrjit collector exposes metrics about the dotnet Just-in-Time compiler.

|||
-|-
Metric name prefix  | `netframework_clrjit`
Classes             | `Win32_PerfRawData_NETFramework_NETCLRJit`
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_netframework_clrjit_jit_methods_total` | Displays the total number of methods JIT-compiled since the application started. This counter does not include pre-JIT-compiled methods. | counter | `process`
`windows_netframework_clrjit_jit_time_percent` | Displays the percentage of time spent in JIT compilation. This counter is updated at the end of every JIT compilation phase. A JIT compilation phase occurs when a method and its dependencies are compiled. | gauge | `process`
`windows_netframework_clrjit_jit_standard_failures_total` | Displays the peak number of methods the JIT compiler has failed to compile since the application started. This failure can occur if the MSIL cannot be verified or if there is an internal error in the JIT compiler. | counter | `process`
`windows_netframework_clrjit_jit_il_bytes_total` | Displays the total number of Microsoft intermediate language (MSIL) bytes compiled by the just-in-time (JIT) compiler since the application started | counter | `process`

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

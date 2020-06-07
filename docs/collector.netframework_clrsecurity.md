# netframework_clrsecurity collector

The netframework_clrsecurity collector exposes metrics about security checks in dotnet applications

|||
-|-
Metric name prefix  | `netframework_clrsecurity`
Classes             | `Win32_PerfRawData_NETFramework_NETCLRSecurity`
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_netframework_clrsecurity_link_time_checks_total` | Displays the total number of link-time code access security checks since the application started. | counter | `process`
`windows_netframework_clrsecurity_rt_checks_time_percent` | Displays the percentage of time spent performing runtime code access security checks in the last sample. | gauge | `process`
`windows_netframework_clrsecurity_stack_walk_depth` | Displays the depth of the stack during that last runtime code access security check. | gauge | `process`
`windows_netframework_clrsecurity_runtime_checks_total` | Displays the total number of runtime code access security checks performed since the application started. | counter | `process`

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

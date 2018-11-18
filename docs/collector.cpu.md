# cpu collector

The cpu collector exposes metrics about CPU usage

|||
-|-
Metric name prefix  | `cpu`
Classes             | [`Win32_PerfRawData_PerfOS_Processor`](https://msdn.microsoft.com/en-us/library/aa394317(v=vs.90).aspx)
Enabled by default? | Yes

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`wmi_cpu_cstate_seconds_total` | Time spent in low-power idle states | counter | `core`, `state`
`wmi_cpu_time_total` | Time that processor spent in different modes (idle, user, system, ...) | counter | `core`, `mode`
`wmi_cpu_interrupts_total` | Total number of received and serviced hardware interrupts | counter | `core`
`wmi_cpu_dpcs_total` | Total number of received and serviced deferred procedure calls (DPCs) | counter | `core`

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

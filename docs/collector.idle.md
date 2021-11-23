# idle collector

The idle collector exposes metrics about idle time.

|||
-|-
Metric name prefix  | `idle`
Data source         | wmi   
Classes             | [`Win32_Process`](https://docs.microsoft.com/en-us/windows/win32/cimwin32prov/win32-process)
Enabled by default? | No

## Flags

### `--collector.idle.period`

The period of time in seconds of inactivity for a system to transition to idle, default to 480 (6 minutes)

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_idle` | Will have the value 1 if the system is idle or 0 if is not | gauge | none

### Example metric

`windows_idle 1`

## Useful queries

_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples

_This collector does not yet have alerting examples, we would appreciate your help adding them!_

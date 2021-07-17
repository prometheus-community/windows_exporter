# scheduled_task collector

The scheduled_task collector exposes metrics about Windows Task Scheduler

|||
-|-
Metric name prefix  | `scheduled_task`
Data source         | OLE
Enabled by default? | No

## Flags

### `--collector.scheduled_task.whitelist`

If given, the path of the task needs to match the whitelist regexp in order for the corresponding metrics to be reported.

### `--collector.scheduled_task.blacklist`

If given, the path of the task needs to *not* match the blacklist regexp in order for the corresponding metrics to be reported.

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_scheduled_task_last_result` | The result that was returned the last time the registered task was run | gauge | task
`windows_scheduled_task_missed_runs` | The number of times the registered task missed a scheduled run | gauge | task
`windows_scheduled_task_state` | The current state of a scheduled task | gauge | task, state

For the values of the `state` label, see below.

### State

A task can be in the following states:
- `disabled`
- `queued`
- `ready`
- `running`
- `unknown`


### Example metric

```
windows_scheduled_task_last_result{task="/Microsoft/Windows/Chkdsk/SyspartRepair"} 1
windows_scheduled_task_missed_runs{task="/Microsoft/Windows/Chkdsk/SyspartRepair"} 0
windows_scheduled_task_state{state="disabled",task="/Microsoft/Windows/Chkdsk/SyspartRepair"} 1
windows_scheduled_task_state{state="queued",task="/Microsoft/Windows/Chkdsk/SyspartRepair"} 0
windows_scheduled_task_state{state="ready",task="/Microsoft/Windows/Chkdsk/SyspartRepair"} 0
windows_scheduled_task_state{state="running",task="/Microsoft/Windows/Chkdsk/SyspartRepair"} 0
windows_scheduled_task_state{state="unknown",task="/Microsoft/Windows/Chkdsk/SyspartRepair"} 0
```

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
**prometheus.rules**
```yaml
  - alert: "WindowsScheduledTaskFailure"
    expr: "windows_scheduled_task_last_result == 0"
    for: "1d"
    labels:
      severity: "high"
    annotations:
      summary: "Scheduled Task Failed"
      description: "Scheduled task '{{ $labels.task }}' failed for 1 day"
```

# scheduled_task collector

The scheduled_task collector exposes metrics about Windows Task Scheduler

|||
-|-
Metric name prefix  | `scheduled_task`
Data source         | OLE
Enabled by default? | No

## Flags

### `--collector.scheduled_task.include`

If given, the path of the task needs to match the include regexp in order for the corresponding metrics to be reported.

E.G. `--collector.scheduled_task.include="Firefox.*"`

### `--collector.scheduled_task.exclude`

If given, the path of the task needs to *not* match the exclude regexp in order for the corresponding metrics to be reported.

E.G. `--collector.scheduled_task.exclude="/Microsoft/.+"`

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_scheduled_task_last_result` | 1 if the last result code is zero, 0 otherwise; omitted for tasks that have never run | gauge | task
`windows_scheduled_task_last_result_code` | The raw Task Scheduler LastTaskResult code as an unsigned 32-bit value | gauge | task
`windows_scheduled_task_last_result_status` | The last result status, 1 if the current status, 0 otherwise | gauge | task, status
`windows_scheduled_task_missed_runs` | The number of times the registered task missed a scheduled run | gauge | task
`windows_scheduled_task_state` | The current state of a scheduled task | gauge | task, state

For the values of the `state` and `status` labels, see below.

### Last result

`windows_scheduled_task_last_result_status` uses the same enum convention as
`windows_service_state`: every status is exported for each task, with exactly one
status set to 1 and all others set to 0. Both new result metrics are exported even
when a task has never run.

Status | Result code (decimal) | Result code (hexadecimal) | Meaning
-------|-----------------------|---------------------------|--------
`success` | 0 | `0x00000000` | The task completed successfully
`ready` | 267008 | `0x00041300` | The task is ready to run
`running` | 267009 | `0x00041301` | The task is currently running
`disabled` | 267010 | `0x00041302` | The task is disabled
`has_not_run` | 267011 | `0x00041303` | The task has never run
`no_more_runs` | 267012 | `0x00041304` | No more runs are scheduled
`not_scheduled` | 267013 | `0x00041305` | Required scheduling properties are missing
`terminated` | 267014 | `0x00041306` | The last run was terminated by the user
`no_valid_triggers` | 267015 | `0x00041307` | No triggers exist or all triggers are disabled
`event_trigger` | 267016 | `0x00041308` | Event triggers do not have set run times
`queued` | 267045 | `0x00041325` | The task has been queued to run
`error` | Any other code | Any other code | An application exit code, scheduler error, or unrecognized result

The raw code remains available in `windows_scheduled_task_last_result_code` to
distinguish results grouped under `error`. Codes are exported as decimal values;
Task Scheduler may display them in hexadecimal. For example, `0x8004130B` is
exported as `2147750667`, even when Windows returns it as a signed value.
See [Microsoft's Task Scheduler result codes](https://learn.microsoft.com/en-us/windows/win32/taskschd/task-scheduler-error-and-success-constants)
for details.

The existing `windows_scheduled_task_last_result` retains its behavior: a running
result code produces 0, which does not necessarily mean the task failed.
`windows_scheduled_task_last_result` and `windows_scheduled_task_missed_runs` are
omitted when the result is `has_not_run`.

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
windows_scheduled_task_last_result_code{task="/Microsoft/Windows/Chkdsk/SyspartRepair"} 0
windows_scheduled_task_last_result_status{status="success",task="/Microsoft/Windows/Chkdsk/SyspartRepair"} 1
windows_scheduled_task_last_result_status{status="ready",task="/Microsoft/Windows/Chkdsk/SyspartRepair"} 0
windows_scheduled_task_last_result_status{status="running",task="/Microsoft/Windows/Chkdsk/SyspartRepair"} 0
windows_scheduled_task_last_result_status{status="disabled",task="/Microsoft/Windows/Chkdsk/SyspartRepair"} 0
windows_scheduled_task_last_result_status{status="has_not_run",task="/Microsoft/Windows/Chkdsk/SyspartRepair"} 0
windows_scheduled_task_last_result_status{status="no_more_runs",task="/Microsoft/Windows/Chkdsk/SyspartRepair"} 0
windows_scheduled_task_last_result_status{status="not_scheduled",task="/Microsoft/Windows/Chkdsk/SyspartRepair"} 0
windows_scheduled_task_last_result_status{status="terminated",task="/Microsoft/Windows/Chkdsk/SyspartRepair"} 0
windows_scheduled_task_last_result_status{status="no_valid_triggers",task="/Microsoft/Windows/Chkdsk/SyspartRepair"} 0
windows_scheduled_task_last_result_status{status="event_trigger",task="/Microsoft/Windows/Chkdsk/SyspartRepair"} 0
windows_scheduled_task_last_result_status{status="queued",task="/Microsoft/Windows/Chkdsk/SyspartRepair"} 0
windows_scheduled_task_last_result_status{status="error",task="/Microsoft/Windows/Chkdsk/SyspartRepair"} 0
windows_scheduled_task_missed_runs{task="/Microsoft/Windows/Chkdsk/SyspartRepair"} 0
windows_scheduled_task_state{state="disabled",task="/Microsoft/Windows/Chkdsk/SyspartRepair"} 1
windows_scheduled_task_state{state="queued",task="/Microsoft/Windows/Chkdsk/SyspartRepair"} 0
windows_scheduled_task_state{state="ready",task="/Microsoft/Windows/Chkdsk/SyspartRepair"} 0
windows_scheduled_task_state{state="running",task="/Microsoft/Windows/Chkdsk/SyspartRepair"} 0
windows_scheduled_task_state{state="unknown",task="/Microsoft/Windows/Chkdsk/SyspartRepair"} 0
```

## Useful queries

Current result status for each task:

```promql
windows_scheduled_task_last_result_status == 1
```

Tasks that have never run:

```promql
windows_scheduled_task_last_result_status{status="has_not_run"} == 1
```

When using the existing success/failure metric, exclude running tasks:

```promql
(windows_scheduled_task_last_result == 0)
unless on (job, instance, task)
(windows_scheduled_task_state{state="running"} == 1)
```

## Alerting examples
**prometheus.rules**
```yaml
  - alert: "WindowsScheduledTaskFailure"
    expr: |
      (windows_scheduled_task_last_result_status{status=~"error|terminated"} == 1)
      unless on (job, instance, task)
      (windows_scheduled_task_state{state="running"} == 1)
    for: "1d"
    labels:
      severity: "high"
    annotations:
      summary: "Scheduled Task Failed"
      description: "Scheduled task '{{ $labels.task }}' failed for 1 day"
```

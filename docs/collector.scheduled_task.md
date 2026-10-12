# scheduled_task collector

The scheduled_task collector exposes metrics about Windows Task Scheduler

|||
-|-
Metric name prefix  | `scheduled_task`
Data source         | OLE
Enabled by default? | No

## Flags

### `--collector.scheduled_task.include`

If given, the path of the task needs to match the include regular expression in order for the corresponding metrics to be reported. Default: `.+`

E.G. `--collector.scheduled_task.include="Firefox.*"`

### `--collector.scheduled_task.exclude`

If given, the path of the task needs to *not* match the exclude regular expression in order for the corresponding metrics to be reported. Default: empty

E.G. `--collector.scheduled_task.exclude="/Microsoft/.+"`

Each task folder costs the Task Scheduler service two requests per scrape. The
collector does not read folders whose tasks the include and exclude expressions
reject whatever their names are, such as everything below `/Microsoft/` in the
example above. Expressions of the form `/Folder/.+` or `/Folder/.*` (also as
alternatives) are recognized for exclude, and a literal folder prefix for include.
Errors reading folders that are not walked are not reported.

Most folders of a Windows installation are below `/Microsoft/`, so excluding
them saves most of a scrape. To keep some of them, either list the folders to
exclude:

E.G. `--collector.scheduled_task.exclude="/Microsoft/Windows/(?:Defrag|Diagnosis|WindowsUpdate)/.+"`

or exclude all of them and add the tasks to keep with
`--collector.scheduled_task.include-override`.

### `--collector.scheduled_task.include-override`

If given, tasks whose path matches the include-override regular expression are
reported whether or not they match include and exclude. The collector reads the
folders that may hold such tasks. Default: empty

E.G. `--collector.scheduled_task.exclude="/Microsoft/.+" --collector.scheduled_task.include-override="/Microsoft/Windows/Defrag/ScheduledDefrag"`

### `--collector.scheduled_task.skip-missed-runs`

If set, the collector does not read or report
`windows_scheduled_task_missed_runs`. Reading the missed runs of a task costs
the Task Scheduler service two requests per scrape, about a quarter of a
scrape. Default: `false`

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_scheduled_task_last_result` | **Deprecated:** use `windows_scheduled_task_last_result_status`. 1 if the last result code is zero, 0 otherwise; omitted for tasks that have never run | gauge | task
`windows_scheduled_task_last_result_status` | The last result status of a scheduled task, 1 if the current status, 0 otherwise | gauge | task, status
`windows_scheduled_task_missed_runs` | The number of times the registered task missed a scheduled run; omitted with `--collector.scheduled_task.skip-missed-runs` | gauge | task
`windows_scheduled_task_state` | The current state of a scheduled task | gauge | task, state

For the values of the `state` and `status` labels, see below.

### Last result

`windows_scheduled_task_last_result_status` uses the same enum convention as
`windows_service_state`: every status is exported for each task, with exactly one
status set to 1 and all others set to 0. The status metric is exported even when a
task has never run.

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
`unknown` | Any other code | Any other code | The result code has no named status mapping

Application exit codes, scheduler errors, and unrecognized results are grouped
under `unknown`. This status means the result code has no named mapping; it does
not establish whether the task succeeded or failed.
See [Microsoft's Task Scheduler result codes](https://learn.microsoft.com/en-us/windows/win32/taskschd/task-scheduler-error-and-success-constants)
for details.

### Deprecated metric

`windows_scheduled_task_last_result` is deprecated. Use
`windows_scheduled_task_last_result_status` to distinguish success, running,
tasks that have never run, and other scheduler statuses. For example, use
`windows_scheduled_task_last_result_status{status="success"} == 1` to select
successful results.

The deprecated metric remains available with its existing labels and values for
compatibility. A running result code still produces 0, which does not necessarily
mean the task failed. `windows_scheduled_task_last_result` and
`windows_scheduled_task_missed_runs` are omitted when the result is `has_not_run`.

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
windows_scheduled_task_last_result_status{status="unknown",task="/Microsoft/Windows/Chkdsk/SyspartRepair"} 0
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

For dashboards still using the deprecated success/failure metric, exclude running tasks:

```promql
(windows_scheduled_task_last_result == 0)
unless on (job, instance, task)
(windows_scheduled_task_state{state="running"} == 1)
```

## Alerting examples

This example treats unknown or terminated results as failures. An `unknown`
status only means that the result code has no named mapping; adjust the alert
selection for your task's exit-code conventions. Running tasks report the
`running` status, so they do not match this alert.

**prometheus.rules**
```yaml
  - alert: "WindowsScheduledTaskFailure"
    expr: 'windows_scheduled_task_last_result_status{status=~"unknown|terminated"} == 1'
    for: "1d"
    labels:
      severity: "high"
    annotations:
      summary: "Scheduled Task Failed"
      description: "Scheduled task '{{ $labels.task }}' failed for 1 day"
```

---
title: WindowsScheduledTaskFailed
---

## Meaning

The last run of the scheduled task `task` ended with the result `unknown` or
`terminated`, and the task has not completed successfully since for one hour.
The alert needs the `scheduled_task` collector. `terminated` means the run was
stopped by a user or by the scheduler's time limit. `unknown` is every result
code without a named status, which includes the exit code of the task's
program and the error codes of the Task Scheduler; the documentation of the
collector lists the named codes. A program that exits with a code other than
zero on success raises this alert on every run; exclude such tasks with
`taskSelector` or with the include and exclude flags of the collector. A
disabled task keeps its last result but does not raise the alert.

## Impact

The work of the task did not happen: backups, cleanups, reports, certificate
renewals or synchronizations that are scheduled on this host. The impact
depends on the task.

## Diagnosis

1. Inspect `windows_scheduled_task_last_result_status` for the task and
   `windows_scheduled_task_state` to see whether the task is still enabled
   and queued.
2. On the host, read the last result and the history:

   ```powershell
   Get-ScheduledTaskInfo -TaskName '<name>' -TaskPath '<folder>'
   Get-WinEvent -LogName Microsoft-Windows-TaskScheduler/Operational -MaxEvents 50 | Where-Object Message -like '*<name>*'
   ```

   The history log is off by default; enable it with
   `wevtutil set-log Microsoft-Windows-TaskScheduler/Operational /enabled:true`.
3. Translate the result code: `0x1` is a generic failure of the program, `0x2`
   a missing file, `0x80070005` an access denied, and codes of the form
   `0x8xxxxxxx` are HRESULTs of the scheduler or of the program. The other
   codes between `0x41300` and `0x41325` are named statuses, such as `ready`
   or `running`, and do not raise this alert.
4. Run the program by hand under the task's account to reproduce the failure.

## Mitigation

Fix the program, its arguments, its working directory or the permissions of
the account the task runs as. Start the task once with `Start-ScheduledTask`
and confirm the result turns to `success`. Disable or delete tasks that are
not needed anymore instead of leaving them failing; disabling a task clears
the alert.

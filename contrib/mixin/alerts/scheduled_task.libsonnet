{
  local c = self._config,
  local a = (import '../lib/alerts.libsonnet')(c),
  prometheusAlerts+:: {
    groups+: a.group('windows-scheduled-task', collectors=['scheduled_task'], rules=[
      {
        alert: 'WindowsScheduledTaskFailed',
        // unknown is any result code without a named status, including the
        // exit codes of the task's program.
        // A disabled task keeps its last result, so disabling a broken task
        // clears the alert.
        expr: |||
          windows_scheduled_task_last_result_status{%(windowsExporterSelector)s, %(taskSelector)s, status=~"unknown|terminated"} == 1
          unless ignoring (status, state) (windows_scheduled_task_state{%(windowsExporterSelector)s, %(taskSelector)s, state="disabled"} == 1)
        ||| % c,
        'for': '1h',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Windows scheduled task did not succeed',
          description: 'The last run of task {{ $labels.task }} on {{ $labels.instance }} ended with result {{ $labels.status }}, and the task has not succeeded in the hour since.',
        },
        state:: true,
      },
    ]),
  },
}

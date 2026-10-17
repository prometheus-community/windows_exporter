{
  local c = self._config,
  local a = (import '../lib/alerts.libsonnet')(c),
  prometheusAlerts+:: {
    groups+: a.group('windows-service', collectors=['service'], rules=[
      {
        alert: 'WindowsServiceNotRunning',
        // The matches ignore only the labels of the right-hand metric, so
        // every target label, including ones a selector adds, takes part.
        expr: |||
          max without (state) (windows_service_state{%(windowsExporterSelector)s, %(serviceSelector)s, state="running"} == 0)
          and ignoring (start_mode) (windows_service_start_mode{%(windowsExporterSelector)s, %(serviceSelector)s, start_mode="auto"} == 1)
          unless ignoring (display_name, run_as, path_name, delayed_start, trigger_start) (windows_service_info{%(windowsExporterSelector)s, %(serviceSelector)s, trigger_start="true"} == 1)
        ||| % c,
        'for': '15m',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Windows automatic service is not running',
          description: 'Service {{ $labels.service }} on {{ $labels.instance }} starts automatically but has not been running for 15 minutes.',
        },
        state:: true,
      },
      {
        alert: 'WindowsServiceCrashLooping',
        // Every service process has its own series, so the series seen in the
        // window count the processes the service ran as.
        expr: |||
          count without (process_id) (count_over_time(windows_service_process{%(windowsExporterSelector)s, %(serviceSelector)s}[15m])) - 1 > %(serviceRestartThreshold)s
        ||| % c,
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Windows service is restarting repeatedly',
          description: 'Service {{ $labels.service }} on {{ $labels.instance }} has restarted more than %(serviceRestartThreshold)s times in 15 minutes.' % c,
        },
      },
    ]),
  },
}

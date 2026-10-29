{
  local c = self._config,
  local a = (import '../lib/alerts.libsonnet')(c),
  prometheusAlerts+:: {
    groups+: a.group('windows-time', collectors=['time'], rules=[
      {
        alert: 'WindowsClockNotSynchronising',
        expr: |||
          windows_time_ntp_client_time_sources{%(windowsExporterSelector)s} == 0
        ||| % c,
        'for': '10m',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Windows clock has no NTP time source',
          description: 'Windows Time on {{ $labels.instance }} has had no active NTP time source for 10 minutes.',
        },
        state:: true,
      },
      {
        alert: 'WindowsClockSkewDetected',
        expr: |||
          abs(windows_time_computed_time_offset_seconds{%(windowsExporterSelector)s}) > %(clockOffsetThresholdSeconds)g
        ||| % c,
        'for': '10m',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Windows clock offset is high',
          description: 'Windows Time on {{ $labels.instance }} reports a clock offset above %(clockOffsetThresholdSeconds)g seconds for 10 minutes.' % c,
        },
      },
      {
        alert: 'WindowsClockSourceLocal',
        // Covers domain members (NT5DS) too, which the NTP source count misses.
        expr: |||
          windows_time_clock_sync_source{%(windowsExporterSelector)s, type=~"Local CMOS Clock|NoSync"} == 1
        ||| % c,
        'for': '10m',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Windows clock has no external time source',
          description: 'Windows Time on {{ $labels.instance }} has used {{ $labels.type }} as its clock source for 10 minutes.',
        },
        state:: true,
      },
      {
        alert: 'WindowsClockSyncFailing',
        expr: |||
          windows_time_sync_last_result{%(windowsExporterSelector)s, result!="success"} == 1
        ||| % c,
        'for': '1h',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Windows clock synchronization fails',
          description: 'The last synchronization attempt of Windows Time on {{ $labels.instance }} has ended with {{ $labels.result }} for 1 hour.',
        },
        state:: true,
      },
    ]),
  },
}

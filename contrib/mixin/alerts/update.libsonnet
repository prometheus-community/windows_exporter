{
  local c = self._config,
  local a = (import '../lib/alerts.libsonnet')(c),
  prometheusAlerts+:: {
    groups+: a.group('windows-update', collectors=['update'], rules=[
      {
        alert: 'WindowsUpdatesPending',
        expr: |||
          count without (id, revision, category, severity, title) (windows_update_pending_info{%(windowsExporterSelector)s, deployment_action="installation"}) > 0
        ||| % c,
        'for': c.updatePendingDuration,
        labels: { severity: c.infoSeverity },
        annotations: {
          summary: 'Windows updates are waiting for installation',
          description: '{{ $labels.instance }} has had {{ $value }} updates waiting for installation for %(updatePendingDuration)s.' % c,
        },
        state:: true,
      },
    ]),
  },
}

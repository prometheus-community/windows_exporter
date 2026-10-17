{
  local c = self._config,
  local a = (import '../lib/alerts.libsonnet')(c),
  prometheusAlerts+:: {
    groups+: a.group('windows-license', collectors=['license'], rules=[
      {
        alert: 'WindowsLicenseNotGenuine',
        expr: |||
          windows_license_status{%(windowsExporterSelector)s, state!="genuine"} == 1
        ||| % c,
        'for': '1h',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Windows license is not genuine',
          description: 'Windows on {{ $labels.instance }} has reported license status {{ $labels.state }} for 1 hour.',
        },
        state:: true,
      },
    ]),
  },
}

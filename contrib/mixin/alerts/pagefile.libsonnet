{
  local c = self._config,
  local a = (import '../lib/alerts.libsonnet')(c),
  prometheusAlerts+:: {
    groups+: a.group('windows-pagefile', collectors=['pagefile'], rules=[
      {
        alert: 'WindowsPagefileAlmostFull',
        expr: |||
          (1 - (windows_pagefile_free_bytes{%(windowsExporterSelector)s} / (windows_pagefile_limit_bytes{%(windowsExporterSelector)s} > 0))) * 100 > %(pagefileUsageThreshold)s
        ||| % c,
        'for': '15m',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Windows page file is almost full',
          description: 'Page file {{ $labels.file }} on {{ $labels.instance }} has been more than %(pagefileUsageThreshold)s%% in use for 15 minutes.' % c,
        },
      },
    ]),
  },
}

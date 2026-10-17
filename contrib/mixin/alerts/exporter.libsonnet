{
  local c = self._config,
  local a = (import '../lib/alerts.libsonnet')(c),
  prometheusAlerts+:: {
    groups+: a.group('windows-exporter', collectors=[], rules=[
      {
        alert: 'WindowsExporterDown',
        expr: |||
          up{%(windowsExporterSelector)s} == 0
        ||| % c,
        'for': '5m',
        labels: { severity: c.criticalSeverity },
        annotations: {
          summary: 'Windows exporter is unreachable',
          description: 'Prometheus cannot scrape Windows exporter on {{ $labels.instance }} (job {{ $labels.job }}).',
        },
        state:: true,
      },
      {
        alert: 'WindowsCollectorFailed',
        expr: |||
          windows_exporter_collector_success{%(windowsExporterSelector)s} == 0
        ||| % c,
        'for': '5m',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Windows collector failed',
          description: 'Collector {{ $labels.collector }} on {{ $labels.instance }} has failed for at least 5 minutes.',
        },
        state:: true,
      },
      {
        alert: 'WindowsCollectorTimeout',
        expr: |||
          windows_exporter_collector_timeout{%(windowsExporterSelector)s} == 1
        ||| % c,
        'for': '10m',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Windows collector exceeds the scrape timeout',
          description: 'Collector {{ $labels.collector }} on {{ $labels.instance }} has not finished within the scrape timeout for 10 minutes.',
        },
        state:: true,
      },
      {
        alert: 'WindowsScrapeQueueTimeout',
        expr: |||
          windows_exporter_scrape_queue_timeout{%(windowsExporterSelector)s} == 1
        ||| % c,
        'for': '10m',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Windows exporter scrapes time out in the queue',
          description: 'Scrapes of {{ $labels.instance }} have timed out waiting for an earlier scrape to finish for 10 minutes.',
        },
        state:: true,
      },
    ]),
  },
}

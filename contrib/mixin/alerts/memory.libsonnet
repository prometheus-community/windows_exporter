{
  local c = self._config,
  local a = (import '../lib/alerts.libsonnet')(c),
  prometheusAlerts+:: {
    groups+: a.group('windows-memory', collectors=['memory'], rules=[
      {
        alert: 'WindowsMemoryHighUsage',
        expr: |||
          windows:memory_usage:ratio{%(windowsExporterSelector)s} * 100 > %(memoryHighUsageThreshold)s
        ||| % c,
        'for': '15m',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Windows memory usage is high',
          description: 'Physical memory usage on {{ $labels.instance }} has exceeded %(memoryHighUsageThreshold)s%% for 15 minutes.' % c,
        },
      },
      {
        alert: 'WindowsMemoryCommitHighUsage',
        expr: |||
          windows:memory_committed:ratio{%(windowsExporterSelector)s} * 100 > %(memoryCommitHighUsageThreshold)s
        ||| % c,
        'for': '15m',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Windows memory commit usage is high',
          description: 'Committed memory on {{ $labels.instance }} has exceeded %(memoryCommitHighUsageThreshold)s%% of the commit limit for 15 minutes.' % c,
        },
      },
      {
        alert: 'WindowsMemoryPagingHigh',
        expr: |||
          windows:memory_swap_page_operations:rate{%(windowsExporterSelector)s} > %(memoryPagingThreshold)s
        ||| % c,
        'for': '15m',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Windows memory paging is high',
          description: 'Paging on {{ $labels.instance }} has exceeded %(memoryPagingThreshold)s pages per second for 15 minutes.' % c,
        },
      },
    ]),
  },
}

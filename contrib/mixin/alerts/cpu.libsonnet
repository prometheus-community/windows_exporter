{
  local c = self._config,
  local a = (import '../lib/alerts.libsonnet')(c),
  prometheusAlerts+:: {
    groups+: a.group('windows-cpu', collectors=['cpu'], rules=[
      {
        alert: 'WindowsCPUHighUsage',
        expr: |||
          windows:cpu_usage:ratio{%(windowsExporterSelector)s} * 100 > %(cpuHighUsageThreshold)s
        ||| % c,
        'for': '15m',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Windows CPU usage is high',
          description: 'CPU usage on {{ $labels.instance }} has exceeded %(cpuHighUsageThreshold)s%% for 15 minutes.' % c,
        },
      },
      {
        alert: 'WindowsCPUSaturation',
        expr: |||
          windows:cpu_queue_per_core:ratio{%(windowsExporterSelector)s} > %(cpuQueuePerCoreThreshold)s
        ||| % c,
        'for': '15m',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Windows processor queue is long',
          description: 'More than %(cpuQueuePerCoreThreshold)s threads per logical processor have been waiting for CPU time on {{ $labels.instance }} for 15 minutes.' % c,
        },
        collectors:: ['cpu', 'system'],
      },
    ]),
  },
}

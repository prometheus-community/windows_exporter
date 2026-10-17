{
  local c = self._config,
  local a = (import '../lib/alerts.libsonnet')(c),
  prometheusAlerts+:: {
    groups+: a.group('windows-disk', collectors=['logical_disk'], rules=[
      {
        alert: 'WindowsDiskReadLatencyHigh',
        expr: |||
          windows:logical_disk_read_latency:seconds{%(windowsExporterSelector)s, %(volumeSelector)s} > %(diskLatencyThresholdSeconds)g
        ||| % c,
        'for': '15m',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Windows disk read latency is high',
          description: 'Volume {{ $labels.volume }} on {{ $labels.instance }} has averaged more than %(diskLatencyThresholdSeconds)g seconds per read operation for 15 minutes.' % c,
        },
      },
      {
        alert: 'WindowsDiskWriteLatencyHigh',
        expr: |||
          windows:logical_disk_write_latency:seconds{%(windowsExporterSelector)s, %(volumeSelector)s} > %(diskLatencyThresholdSeconds)g
        ||| % c,
        'for': '15m',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Windows disk write latency is high',
          description: 'Volume {{ $labels.volume }} on {{ $labels.instance }} has averaged more than %(diskLatencyThresholdSeconds)g seconds per write operation for 15 minutes.' % c,
        },
      },
      {
        alert: 'WindowsDiskAlmostFull',
        expr: |||
          windows:logical_disk_free:ratio{%(windowsExporterSelector)s, %(volumeSelector)s} * 100 < %(diskFreeWarningThreshold)s
        ||| % c,
        'for': '15m',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Windows disk space is low',
          description: 'Volume {{ $labels.volume }} on {{ $labels.instance }} has less than %(diskFreeWarningThreshold)s%% free space. Windows may delay space metrics by 10–15 minutes.' % c,
        },
      },
      {
        alert: 'WindowsDiskAlmostFull',
        expr: |||
          windows:logical_disk_free:ratio{%(windowsExporterSelector)s, %(volumeSelector)s} * 100 < %(diskFreeCriticalThreshold)s
        ||| % c,
        'for': '15m',
        labels: { severity: c.criticalSeverity },
        annotations: {
          summary: 'Windows disk space is low',
          description: 'Volume {{ $labels.volume }} on {{ $labels.instance }} has less than %(diskFreeCriticalThreshold)s%% free space. Windows may delay space metrics by 10–15 minutes.' % c,
        },
      },
      {
        alert: 'WindowsDiskFillingUp',
        expr: |||
          windows:logical_disk_free:ratio{%(windowsExporterSelector)s, %(volumeSelector)s} * 100 < %(diskFreeWarningThreshold)s
          and predict_linear(windows_logical_disk_free_bytes{%(windowsExporterSelector)s, %(volumeSelector)s}[%(diskPredictionWindow)s], %(diskPredictionSeconds)s) < 0
        ||| % c { diskPredictionSeconds: c.diskPredictionHours * 3600 },
        'for': '1h',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Windows disk is predicted to fill up',
          description: 'Volume {{ $labels.volume }} on {{ $labels.instance }} is predicted to fill within %(diskPredictionHours)s hours.' % c,
        },
      },
    ]),
  },
}

{
  local c = self._config,
  local q = (import '../lib/queries.libsonnet')(c),
  local alert(name, expr, duration, severity, summary, description) = {
    alert: name,
    expr: expr,
    'for': duration,
    labels: { severity: severity },
    annotations: { summary: summary, description: description },
  },
  prometheusAlerts+:: {
    groups: [
      {
        name: 'windows-exporter-alerts',
        rules: [
          alert(
            'WindowsExporterDown',
            q.metric('up') + ' == 0',
            '5m',
            c.criticalSeverity,
            'Windows exporter is unreachable',
            'Prometheus cannot scrape Windows exporter on {{ $labels.instance }} (job {{ $labels.job }}).'
          ),
          alert(
            'WindowsCollectorFailed',
            q.metric('windows_exporter_collector_success') + ' == 0',
            '5m',
            'warning',
            'Windows collector failed',
            'Collector {{ $labels.collector }} on {{ $labels.instance }} has failed for at least 5 minutes.'
          ),
          alert(
            'WindowsCPUHighUsage',
            q.metric('windows:cpu_usage:ratio') + ' * 100 > %s' % c.cpuHighUsageThreshold,
            '15m',
            'warning',
            'Windows CPU usage is high',
            'CPU usage on {{ $labels.instance }} has exceeded %s%% for 15 minutes.' % c.cpuHighUsageThreshold
          ),
          alert(
            'WindowsMemoryHighUsage',
            q.metric('windows:memory_usage:ratio') + ' * 100 > %s' % c.memoryHighUsageThreshold,
            '15m',
            'warning',
            'Windows memory usage is high',
            'Physical memory usage on {{ $labels.instance }} has exceeded %s%% for 15 minutes.' % c.memoryHighUsageThreshold
          ),
        ] + [
          alert(
            'WindowsDiskAlmostFull',
            q.metric('windows:logical_disk_free:ratio', [c.volumeSelector]) + ' * 100 < %s' % entry[1],
            '15m',
            entry[0],
            'Windows disk space is low',
            'Volume {{ $labels.volume }} on {{ $labels.instance }} has less than %s%% free space. Windows may delay space metrics by 10–15 minutes.' % entry[1]
          )
          for entry in [
            ['warning', c.diskFreeWarningThreshold],
            [c.criticalSeverity, c.diskFreeCriticalThreshold],
          ]
        ] + [
          alert(
            'WindowsDiskFillingUp',
            q.metric('windows:logical_disk_free:ratio', [c.volumeSelector]) + ' * 100 < %s and predict_linear(%s[%s], %s) < 0' % [
              c.diskFreeWarningThreshold,
              q.metric('windows_logical_disk_free_bytes', [c.volumeSelector]),
              c.diskPredictionWindow,
              c.diskPredictionHours * 3600,
            ],
            '1h',
            'warning',
            'Windows disk is predicted to fill up',
            'Volume {{ $labels.volume }} on {{ $labels.instance }} is predicted to fill within %s hours.' % c.diskPredictionHours
          ),
        ],
      },
    ] + if c.enableActiveDirectory then [
      {
        name: 'windows-active-directory',
        rules: [
          alert(
            'WindowsADPendingReplication',
            q.metric('windows_ad_replication_pending_operations') + ' > %s' % c.adPendingReplicationThreshold,
            '15m',
            'warning',
            'Active Directory replication queue is high',
            'Pending replication operations on {{ $labels.instance }} have exceeded %s for 15 minutes.' % c.adPendingReplicationThreshold
          ),
          alert(
            'WindowsADReplicationFailures',
            '(' + q.rate('windows_ad_replication_sync_requests_total') + ' - ' + q.rate('windows_ad_replication_sync_requests_success_total') + ') > %s' % c.adReplicationFailureRateThreshold,
            '15m',
            'warning',
            'Active Directory replication requests are failing',
            'Replication synchronization requests on {{ $labels.instance }} have been failing for 15 minutes.'
          ),
        ],
      },
    ] else [],
  },
}

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
          alert(
            'WindowsMemoryCommitHighUsage',
            q.metric('windows:memory_committed:ratio') + ' * 100 > %s' % c.memoryCommitHighUsageThreshold,
            '15m',
            'warning',
            'Windows memory commit usage is high',
            'Committed memory on {{ $labels.instance }} has exceeded %s%% of the commit limit for 15 minutes.' % c.memoryCommitHighUsageThreshold
          ),
        ] + [
          alert(
            'WindowsDisk%sLatencyHigh' % direction[0],
            q.metric('windows:logical_disk_%s_latency:seconds' % direction[1], [c.volumeSelector]) + ' > %g' % c.diskLatencyThresholdSeconds,
            '15m',
            'warning',
            'Windows disk %s latency is high' % direction[1],
            'Volume {{ $labels.volume }} on {{ $labels.instance }} has averaged more than %g seconds per %s operation for 15 minutes.' % [c.diskLatencyThresholdSeconds, direction[1]]
          )
          for direction in [['Read', 'read'], ['Write', 'write']]
        ] + [
          alert(
            'WindowsNetwork%s' % outcome[0],
            '(%s + %s) > %g' % [
              q.metric('windows:net_received_%s:rate' % outcome[1], [c.nicSelector]),
              q.metric('windows:net_outbound_%s:rate' % outcome[1], [c.nicSelector]),
              outcome[2],
            ],
            '15m',
            'warning',
            'Windows network packet %s are high' % outcome[3],
            'Packet %s on interface {{ $labels.nic }} on {{ $labels.instance }} have exceeded %g per second for 15 minutes.' % [outcome[3], outcome[2]]
          )
          for outcome in [
            ['Errors', 'errors', c.networkErrorRateThreshold, 'errors'],
            ['Discards', 'discarded', c.networkDiscardRateThreshold, 'discards'],
          ]
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
    ] + (if c.enableActiveDirectory then [
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
         ] else []) + if c.enableTime then [
      {
        name: 'windows-time',
        rules: [
          alert(
            'WindowsClockNotSynchronising',
            q.metric('windows_time_ntp_client_time_sources') + ' == 0',
            '10m',
            'warning',
            'Windows clock has no NTP time source',
            'Windows Time on {{ $labels.instance }} has had no active NTP time source for 10 minutes.'
          ),
          alert(
            'WindowsClockSkewDetected',
            'abs(%s) > %g' % [q.metric('windows_time_computed_time_offset_seconds'), c.clockOffsetThresholdSeconds],
            '10m',
            'warning',
            'Windows clock offset is high',
            'Windows Time on {{ $labels.instance }} reports a clock offset above %g seconds for 10 minutes.' % c.clockOffsetThresholdSeconds
          ),
        ],
      },
    ] else [],
  },
}

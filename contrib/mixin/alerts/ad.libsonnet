{
  local c = self._config,
  local a = (import '../lib/alerts.libsonnet')(c),
  prometheusAlerts+:: {
    groups+: a.group('windows-active-directory', collectors=['ad'], rules=[
      {
        alert: 'WindowsADPendingReplication',
        expr: |||
          windows_ad_replication_pending_operations{%(windowsExporterSelector)s} > %(adPendingReplicationThreshold)s
        ||| % c,
        'for': '15m',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Active Directory replication queue is high',
          description: 'Pending replication operations on {{ $labels.instance }} have exceeded %(adPendingReplicationThreshold)s for 15 minutes.' % c,
        },
      },
      {
        alert: 'WindowsADReplicationFailures',
        expr: |||
          (rate(windows_ad_replication_sync_requests_total{%(windowsExporterSelector)s}[%(rateInterval)s]) - rate(windows_ad_replication_sync_requests_success_total{%(windowsExporterSelector)s}[%(rateInterval)s])) > %(adReplicationFailureRateThreshold)s
        ||| % c,
        'for': '15m',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Active Directory replication requests are failing',
          description: 'Replication synchronization requests on {{ $labels.instance }} have been failing for 15 minutes.',
        },
      },
    ]),
  },
}

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
      {
        alert: 'WindowsADReplicationPartnerFailing',
        expr: |||
          windows_ad_replication_partner_consecutive_failures{%(windowsExporterSelector)s} > 0
        ||| % c,
        'for': '1h',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Active Directory replication from a partner is failing',
          description: '{{ $labels.instance }} has failed to replicate {{ $labels.naming_context }} from {{ $labels.partner_site }}/{{ $labels.partner }} {{ $value }} times in a row, for at least an hour.',
        },
      },
    ]),
  },
}

{
  local c = self._config,
  local a = (import '../lib/alerts.libsonnet')(c),
  prometheusAlerts+:: {
    groups+: a.group('windows-dns', collectors=['dns'], rules=[
      {
        alert: 'WindowsDNSZoneTransferFailures',
        expr: |||
          rate(windows_dns_zone_transfer_failures_total{%(windowsExporterSelector)s}[%(rateInterval)s]) > %(dnsZoneTransferFailureRateThreshold)g
        ||| % c,
        'for': '15m',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'DNS zone transfers are failing',
          description: 'Zone transfers from DNS server {{ $labels.instance }} have been failing for 15 minutes.',
        },
      },
    ]),
  },
}

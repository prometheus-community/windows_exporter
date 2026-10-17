{
  local c = self._config,
  local a = (import '../lib/alerts.libsonnet')(c),
  prometheusAlerts+:: {
    groups+: a.group('windows-network', collectors=['net'], rules=[
      {
        alert: 'WindowsNetworkErrors',
        expr: |||
          (
            (windows:net_received_errors:rate{%(windowsExporterSelector)s, %(nicSelector)s} + windows:net_outbound_errors:rate{%(windowsExporterSelector)s, %(nicSelector)s})
            / ((windows:net_received_packets:rate{%(windowsExporterSelector)s, %(nicSelector)s} + windows:net_sent_packets:rate{%(windowsExporterSelector)s, %(nicSelector)s}) > 0)
          ) * 100 > %(networkErrorRatioThreshold)g
        ||| % c,
        'for': '15m',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Windows network packet errors are high',
          description: 'Packet errors on interface {{ $labels.nic }} on {{ $labels.instance }} have exceeded %(networkErrorRatioThreshold)g%% of all packets for 15 minutes.' % c,
        },
      },
      {
        alert: 'WindowsNetworkDiscards',
        expr: |||
          (
            (windows:net_received_discarded:rate{%(windowsExporterSelector)s, %(nicSelector)s} + windows:net_outbound_discarded:rate{%(windowsExporterSelector)s, %(nicSelector)s})
            / ((windows:net_received_packets:rate{%(windowsExporterSelector)s, %(nicSelector)s} + windows:net_sent_packets:rate{%(windowsExporterSelector)s, %(nicSelector)s}) > 0)
          ) * 100 > %(networkDiscardRatioThreshold)g
        ||| % c,
        'for': '15m',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Windows network packet discards are high',
          description: 'Packet discards on interface {{ $labels.nic }} on {{ $labels.instance }} have exceeded %(networkDiscardRatioThreshold)g%% of all packets for 15 minutes.' % c,
        },
      },
      {
        alert: 'WindowsNetworkUtilizationHigh',
        expr: |||
          (windows:net_received_utilization:ratio{%(windowsExporterSelector)s, %(nicSelector)s} * 100 > %(networkUtilizationThreshold)s)
          or (windows:net_sent_utilization:ratio{%(windowsExporterSelector)s, %(nicSelector)s} * 100 > %(networkUtilizationThreshold)s)
        ||| % c,
        'for': '15m',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Windows network interface utilization is high',
          description: 'Traffic on interface {{ $labels.nic }} on {{ $labels.instance }} has exceeded %(networkUtilizationThreshold)s%% of its bandwidth in at least one direction for 15 minutes.' % c,
        },
      },
      {
        alert: 'WindowsNetworkInterfaceFlapping',
        expr: |||
          max without (status) (changes(windows_net_nic_operation_status{%(windowsExporterSelector)s, %(nicSelector)s, status="up"}[%(rateInterval)s])) > 2
        ||| % c,
        'for': '2m',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Windows network interface is flapping',
          description: 'Interface {{ $labels.nic }} on {{ $labels.instance }} has changed its operational status more than twice in %(rateInterval)s.' % c,
        },
      },
    ]),
  },
}

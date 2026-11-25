{
  local c = self._config,
  local a = (import '../lib/alerts.libsonnet')(c),
  prometheusAlerts+:: {
    groups+: a.group('windows-tcp', collectors=['tcp'], rules=[
      // Percent of the dynamic port range held by connections in any state,
      // including TIME_WAIT. Needs the dynamic_ports sub-collector.
      {
        alert: 'WindowsTCPDynamicPortsAlmostExhausted',
        expr: |||
          (windows_tcp_dynamic_ports_in_use{%(windowsExporterSelector)s} / (windows_tcp_dynamic_port_range_size{%(windowsExporterSelector)s} > 0)) * 100 > %(tcpDynamicPortUsageThreshold)s
        ||| % c,
        'for': '10m',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'TCP dynamic ports are almost exhausted',
          description: 'Connections on {{ $labels.instance }} have held more than %(tcpDynamicPortUsageThreshold)s%% of the TCP dynamic port range for 10 minutes.' % c,
        },
      },
    ]),
  },
}

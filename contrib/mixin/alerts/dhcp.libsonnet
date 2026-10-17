{
  local c = self._config,
  local a = (import '../lib/alerts.libsonnet')(c),
  prometheusAlerts+:: {
    groups+: a.group('windows-dhcp', collectors=['dhcp'], rules=[
      // Percent of free addresses in a scope; a scope without addresses has
      // no sample, and only scopes in an enabled state count.
      {
        alert: 'WindowsDHCPScopeAlmostExhausted',
        expr: |||
          ((windows_dhcp_scope_addresses_free{%(windowsExporterSelector)s} / ((windows_dhcp_scope_addresses_free{%(windowsExporterSelector)s} + windows_dhcp_scope_addresses_in_use{%(windowsExporterSelector)s}) > 0)) * 100) < %(dhcpScopeFreeWarningThreshold)s
          and ignoring (state) (windows_dhcp_scope_state{%(windowsExporterSelector)s, state=~"Enabled|EnabledSwitched"} == 1)
        ||| % c,
        'for': '15m',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'DHCP scope is almost exhausted',
          description: 'DHCP scope {{ $labels.scope }} on {{ $labels.instance }} has had less than %(dhcpScopeFreeWarningThreshold)s%% free addresses for 15 minutes.' % c,
        },
      },
      {
        alert: 'WindowsDHCPScopeAlmostExhausted',
        expr: |||
          ((windows_dhcp_scope_addresses_free{%(windowsExporterSelector)s} / ((windows_dhcp_scope_addresses_free{%(windowsExporterSelector)s} + windows_dhcp_scope_addresses_in_use{%(windowsExporterSelector)s}) > 0)) * 100) < %(dhcpScopeFreeCriticalThreshold)s
          and ignoring (state) (windows_dhcp_scope_state{%(windowsExporterSelector)s, state=~"Enabled|EnabledSwitched"} == 1)
        ||| % c,
        'for': '15m',
        labels: { severity: c.criticalSeverity },
        annotations: {
          summary: 'DHCP scope is almost exhausted',
          description: 'DHCP scope {{ $labels.scope }} on {{ $labels.instance }} has had less than %(dhcpScopeFreeCriticalThreshold)s%% free addresses for 15 minutes.' % c,
        },
      },
    ]),
  },
}

{
  local c = self._config,
  local a = (import '../lib/alerts.libsonnet')(c),
  prometheusAlerts+:: {
    groups+: a.group('windows-system', collectors=['system'], rules=[
      {
        alert: 'WindowsHostRebooted',
        // Fires once per reboot: for the first ten minutes after the boot,
        // and only on a host that was monitored before it. The exporter is
        // down while the host reboots, so the second operand takes the last
        // boot time of the hour before, not the sample of ten minutes ago.
        expr: |||
          (time() - windows_system_boot_time_timestamp{%(windowsExporterSelector)s}) < 600
          and (time() - last_over_time(windows_system_boot_time_timestamp{%(windowsExporterSelector)s}[1h] offset 10m)) > 600
        ||| % c,
        labels: { severity: c.infoSeverity },
        annotations: {
          summary: 'Windows host has rebooted',
          description: '{{ $labels.instance }} rebooted {{ $value | humanizeDuration }} ago.',
        },
        state:: true,
      },
    ]),
  },
}

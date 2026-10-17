{
  local c = self._config,
  local a = (import '../lib/alerts.libsonnet')(c),
  prometheusAlerts+:: {
    groups+: a.group('windows-hyperv', collectors=['hyperv'], rules=[
      {
        alert: 'WindowsHyperVWMIUnhealthy',
        expr: |||
          windows_hyperv_wmi_health{%(windowsExporterSelector)s} == 0
        ||| % c,
        'for': '10m',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Hyper-V WMI namespace is not responding',
          description: 'The Hyper-V WMI namespace on {{ $labels.instance }} has not responded for 10 minutes; Hyper-V metrics are missing.',
        },
        state:: true,
      },
      {
        alert: 'WindowsHyperVReplicationCritical',
        expr: |||
          windows_hyperv_replica_vm_health{%(windowsExporterSelector)s} == 3
        ||| % c,
        'for': '5m',
        labels: { severity: c.criticalSeverity },
        annotations: {
          summary: 'Hyper-V replication is critical',
          description: 'Replication of virtual machine {{ $labels.vm }} ({{ $labels.relationship }}) on {{ $labels.instance }} has been in critical health for 5 minutes.',
        },
        state:: true,
      },
      {
        alert: 'WindowsHyperVReplicationWarning',
        expr: |||
          windows_hyperv_replica_vm_health{%(windowsExporterSelector)s} == 2
        ||| % c,
        'for': '15m',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Hyper-V replication has a warning',
          description: 'Replication of virtual machine {{ $labels.vm }} ({{ $labels.relationship }}) on {{ $labels.instance }} has been in warning health for 15 minutes.',
        },
        state:: true,
      },
    ]),
  },
}

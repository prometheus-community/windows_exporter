// Every cluster node reports the state of the whole cluster, so these
// alerts fire once per node; group them by cluster labels in Alertmanager.
{
  local c = self._config,
  local a = (import '../lib/alerts.libsonnet')(c),
  prometheusAlerts+:: {
    groups+: a.group('windows-cluster', collectors=['mscluster'], rules=[
      {
        alert: 'WindowsClusterNodeDown',
        expr: |||
          windows_mscluster_node_state{%(windowsExporterSelector)s} == 1
        ||| % c,
        'for': '5m',
        labels: { severity: c.criticalSeverity },
        annotations: {
          summary: 'Failover cluster node is down',
          description: 'Cluster node {{ $labels.node }} has been down for 5 minutes, as reported by {{ $labels.instance }}.',
        },
        state:: true,
      },
      {
        alert: 'WindowsClusterResourceFailed',
        expr: |||
          windows_mscluster_resource_state{%(windowsExporterSelector)s} == 4
        ||| % c,
        'for': '5m',
        labels: { severity: c.criticalSeverity },
        annotations: {
          summary: 'Failover cluster resource has failed',
          description: 'Cluster resource {{ $labels.resource }} ({{ $labels.type }}) in group {{ $labels.resource_group }} has been in the failed state for 5 minutes, as reported by {{ $labels.instance }}.',
        },
        state:: true,
      },
      {
        alert: 'WindowsClusterResourceGroupFailed',
        expr: |||
          windows_mscluster_resourcegroup_state{%(windowsExporterSelector)s} == 2
        ||| % c,
        'for': '5m',
        labels: { severity: c.criticalSeverity },
        annotations: {
          summary: 'Failover cluster resource group has failed',
          description: 'Cluster resource group {{ $labels.resource_group }} has been in the failed state for 5 minutes, as reported by {{ $labels.instance }}.',
        },
        state:: true,
      },
      {
        alert: 'WindowsClusterSharedVolumeAlmostFull',
        expr: |||
          (windows_mscluster_shared_volumes_free_bytes{%(windowsExporterSelector)s, %(sharedVolumeSelector)s} / (windows_mscluster_shared_volumes_total_bytes{%(windowsExporterSelector)s, %(sharedVolumeSelector)s} > 0)) * 100 < %(diskFreeWarningThreshold)s
        ||| % c,
        'for': '15m',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Cluster shared volume is almost full',
          description: 'Cluster shared volume {{ $labels.shared_volume }} has had less than %(diskFreeWarningThreshold)s%% free space for 15 minutes, as reported by {{ $labels.instance }}.' % c,
        },
      },
      {
        alert: 'WindowsClusterSharedVolumeAlmostFull',
        expr: |||
          (windows_mscluster_shared_volumes_free_bytes{%(windowsExporterSelector)s, %(sharedVolumeSelector)s} / (windows_mscluster_shared_volumes_total_bytes{%(windowsExporterSelector)s, %(sharedVolumeSelector)s} > 0)) * 100 < %(diskFreeCriticalThreshold)s
        ||| % c,
        'for': '15m',
        labels: { severity: c.criticalSeverity },
        annotations: {
          summary: 'Cluster shared volume is almost full',
          description: 'Cluster shared volume {{ $labels.shared_volume }} has had less than %(diskFreeCriticalThreshold)s%% free space for 15 minutes, as reported by {{ $labels.instance }}.' % c,
        },
      },
    ]),
  },
}

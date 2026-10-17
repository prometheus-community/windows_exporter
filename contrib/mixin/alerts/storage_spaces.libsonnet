// On Storage Spaces Direct every node reports the clustered pools and
// virtual disks, so these alerts fire once per node as well.
{
  local c = self._config,
  local a = (import '../lib/alerts.libsonnet')(c),
  prometheusAlerts+:: {
    groups+: a.group('windows-storage-spaces', collectors=['storage_spaces'], rules=[
      {
        alert: 'WindowsStoragePoolUnhealthy',
        // The primordial pool holds the disks that are not in a concrete pool
        // yet; only the info metric tells it apart.
        expr: |||
          (windows_storage_spaces_pool_health_status{%(windowsExporterSelector)s} == 2)
          and ignoring (primordial) windows_storage_spaces_pool_info{%(windowsExporterSelector)s, primordial="false"}
        ||| % c,
        'for': '10m',
        labels: { severity: c.criticalSeverity },
        annotations: {
          summary: 'Storage pool is unhealthy',
          description: 'Storage pool {{ $labels.pool }} on {{ $labels.instance }} has reported unhealthy status for 10 minutes.',
        },
        state:: true,
      },
      {
        alert: 'WindowsStoragePoolUnhealthy',
        expr: |||
          (windows_storage_spaces_pool_health_status{%(windowsExporterSelector)s} == 1)
          and ignoring (primordial) windows_storage_spaces_pool_info{%(windowsExporterSelector)s, primordial="false"}
        ||| % c,
        'for': '10m',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Storage pool has a health warning',
          description: 'Storage pool {{ $labels.pool }} on {{ $labels.instance }} has reported warning status for 10 minutes.',
        },
        state:: true,
      },
      {
        alert: 'WindowsStorageVirtualDiskUnhealthy',
        expr: |||
          windows_storage_spaces_virtual_disk_health_status{%(windowsExporterSelector)s} == 2
        ||| % c,
        'for': '10m',
        labels: { severity: c.criticalSeverity },
        annotations: {
          summary: 'Storage Spaces virtual disk is unhealthy',
          description: 'Virtual disk {{ $labels.virtual_disk }} on {{ $labels.instance }} has reported unhealthy status for 10 minutes.',
        },
        state:: true,
      },
      {
        alert: 'WindowsStorageVirtualDiskUnhealthy',
        expr: |||
          windows_storage_spaces_virtual_disk_health_status{%(windowsExporterSelector)s} == 1
        ||| % c,
        'for': '10m',
        labels: { severity: 'warning' },
        annotations: {
          summary: 'Storage Spaces virtual disk has a health warning',
          description: 'Virtual disk {{ $labels.virtual_disk }} on {{ $labels.instance }} has reported warning status for 10 minutes.',
        },
        state:: true,
      },
    ]),
  },
}

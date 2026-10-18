{
  local c = self._config,
  local a = (import '../lib/alerts.libsonnet')(c),
  prometheusAlerts+:: {
    groups+: a.group('windows-diskdrive', collectors=['diskdrive'], rules=[
      {
        alert: 'WindowsDiskDriveNotOK',
        // One series per status with the current status set to 1, so the
        // alert carries the status the drive reports. Some USB and virtual
        // drives report Unknown for good.
        expr: |||
          windows_diskdrive_status{%(windowsExporterSelector)s, status!~"OK|Unknown"} == 1
        ||| % c,
        'for': '5m',
        labels: { severity: c.criticalSeverity },
        annotations: {
          summary: 'Windows disk drive is not healthy',
          description: 'Disk drive {{ $labels.device_id }} on {{ $labels.instance }} has reported status {{ $labels.status }} for 5 minutes.',
        },
        state:: true,
      },
    ]),
  },
}

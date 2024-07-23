# logical_disk collector

The logical_disk collector exposes metrics about logical disks (in contrast to physical disks)

|||
-|-
Metric name prefix  | `logical_disk`
Data source         | Perflib
Counters             | `LogicalDisk` ([`Win32_PerfRawData_PerfDisk_LogicalDisk`](https://msdn.microsoft.com/en-us/windows/hardware/aa394307(v=vs.71)))
Enabled by default? | Yes

## Flags

### `--collector.logical_disk.volume-include`

If given, a disk needs to match the include regexp in order for the corresponding disk metrics to be reported

### `--collector.logical_disk.volume-exclude`

If given, a disk needs to *not* match the exclude regexp in order for the corresponding disk metrics to be reported

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_logical_disk_info` | A metric with a constant '1' value labeled with logical disk information | gauge | `disk`,`filesystem`,`serial_number`,`volume`,`volume_name`,`type`
`windows_logical_disk_requests_queued` | Number of requests outstanding on the disk at the time the performance data is collected | gauge | `volume`
`windows_logical_disk_avg_read_requests_queued` | Average number of read requests that were queued for the selected disk during the sample interval | gauge | `volume`
`windows_logical_disk_avg_write_requests_queued` | Average number of write requests that were queued for the selected disk during the sample interval | gauge | `volume`
`windows_logical_disk_read_bytes_total` | Rate at which bytes are transferred from the disk during read operations | counter | `volume`
`windows_logical_disk_reads_total` | Rate of read operations on the disk | counter | `volume`
`windows_logical_disk_write_bytes_total` | Rate at which bytes are transferred to the disk during write operations  | counter | `volume`
`windows_logical_disk_writes_total` | Rate of write operations on the disk  | counter | `volume`
`windows_logical_disk_read_seconds_total` | Seconds the disk was busy servicing read requests | counter | `volume`
`windows_logical_disk_write_seconds_total` | Seconds the disk was busy servicing write requests | counter | `volume`
`windows_logical_disk_free_bytes` | Unused space of the disk in bytes (not real time, updates every 10-15 min) | gauge | `volume`
`windows_logical_disk_size_bytes` | Total size of the disk in bytes (not real time, updates every 10-15 min) | gauge | `volume`
`windows_logical_disk_idle_seconds_total` | Seconds the disk was idle (not servicing read/write requests) | counter | `volume`
`windows_logical_disk_split_ios_total` | Number of I/Os to the disk split into multiple I/Os | counter | `volume`
`windows_logical_disk_readonly` | Whether the logical disk is read-only | gauge | `volume`

### Warning about size metrics
The `free_bytes` and `size_bytes` metrics are not updated in real time and might have a delay of 10-15min.
This is the same behavior as the windows performance counters.

### Example metric
Query the rate of write operations to a disk
```
rate(windows_logical_disk_read_bytes_total{instance="localhost", volume=~"C:"}[2m])
```

Logical Volume information
```
windows_logical_disk_info{disk_id="0",filesystem="",serial_number="",type="",volume="HarddiskVolume2",volume_name=""} 1
windows_logical_disk_info{disk_id="0",filesystem="",serial_number="",type="",volume="HarddiskVolume3",volume_name=""} 1
windows_logical_disk_info{disk_id="0",filesystem="NTFS",serial_number="668EEC37",type="fixed",volume="C:",volume_name="Windows"} 1
windows_logical_disk_info{disk_id="1",filesystem="NTFS",serial_number="50AE953B",type="fixed",volume="D:",volume_name="Temporary Storage"} 1
windows_logical_disk_info{disk_id="1",filesystem="ReFS",serial_number="C69B59AD",type="fixed",volume="G:",volume_name="Volume"} 1
```

## Useful queries
Calculate rate of total IOPS for disk
```
rate(windows_logical_disk_reads_total{instance="localhost", volume="C:"}[2m]) + rate(windows_logical_disk_writes_total{instance="localhost", volume="C:"}[2m])
```

Show volume usage (%)
```
100.0 - 100 * (windows_logical_disk_free_bytes{instance="localhost", volume="C:"} / windows_logical_disk_size_bytes{instance="localhost", volume="C:"})
```

## Alerting examples
**prometheus.rules**
```yaml
groups:
- name: Windows Disk Alerts
  rules:

  # Sends an alert when disk space usage is above 95%
  - alert: DiskSpaceUsage
    expr: 100.0 - 100 * (windows_logical_disk_free_bytes / windows_logical_disk_size_bytes) > 95
    for: 10m
    labels:
      severity: high
    annotations:
      summary: "Disk Space Usage (instance {{ $labels.instance }})"
      description: "Disk Space on Drive is used more than 95%\n  VALUE = {{ $value }}\n  LABELS: {{ $labels }}"

  # Alerts on disks with over 85% space usage predicted to fill within the next four days
  - alert: DiskFilling
    expr: 100 * (windows_logical_disk_free_bytes / windows_logical_disk_size_bytes) < 15 and predict_linear(windows_logical_disk_free_bytes[6h], 4 * 24 * 3600) < 0
    for: 10m
    labels:
      severity: warning
    annotations:
      summary: "Disk full in four days (instance {{ $labels.instance }})"
      description: "{{ $labels.volume }} is expected to fill up within four days. Currently {{ $value | humanize }}% is available.\n VALUE = {{ $value }}\n LABELS: {{ $labels }}"
```

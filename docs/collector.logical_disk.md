# logical_disk collector

The logical_disk collector exposes metrics about logical disks (in contrast to physical disks)

|||
-|-
Metric name prefix  | `logical_disk`
Data source         | Perflib
Counters             | `LogicalDisk` ([`Win32_PerfRawData_PerfDisk_LogicalDisk`](https://msdn.microsoft.com/en-us/windows/hardware/aa394307(v=vs.71)))
Enabled by default? | Yes

## Flags

### `--collector.logical_disk.volume-whitelist`

If given, a disk needs to match the whitelist regexp in order for the corresponding disk metrics to be reported

### `--collector.logical_disk.volume-blacklist`

If given, a disk needs to *not* match the blacklist regexp in order for the corresponding disk metrics to be reported

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`requests_queued` | Number of requests outstanding on the disk at the time the performance data is collected | gauge | `volume`
`read_bytes_total` | Rate at which bytes are transferred from the disk during read operations | counter | `volume`
`reads_total` | Rate of read operations on the disk | counter | `volume`
`write_bytes_total` | Rate at which bytes are transferred to the disk during write operations  | counter | `volume`
`writes_total` | Rate of write operations on the disk  | counter | `volume`
`read_seconds_total` | Seconds the disk was busy servicing read requests | counter | `volume`
`write_seconds_total` | Seconds the disk was busy servicing write requests | counter | `volume`
`free_bytes` | Unused space of the disk in bytes (not real time, updates every 10-15 min) | gauge | `volume`
`size_bytes` | Total size of the disk in bytes (not real time, updates every 10-15 min) | gauge | `volume`
`idle_seconds_total` | Seconds the disk was idle (not servicing read/write requests) | counter | `volume`
`split_ios_total` | Number of I/Os to the disk split into multiple I/Os | counter | `volume`

### Warning about size metrics
The `free_bytes` and `size_bytes` metrics are not updated in real time and might have a delay of 10-15min.
This is the same behavior as the windows performance counters.

### Example metric
Query the rate of write operations to a disk
```
rate(windows_logical_disk_read_bytes_total{instance="localhost", volume=~"C:"}[2m])
```

## Useful queries
Calculate rate of total IOPS for disk
```
rate(windows_logical_disk_reads_total{instance="localhost", volume="C:"}[2m]) + rate(windows_logical_disk_writes_total{instance="localhost", volume="C:"}[2m])
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

# logical_disk collector

The logical_disk collector exposes metrics about logical disks (in contrast to physical disks)

|                     |                  |
|---------------------|------------------|
| Metric name prefix  | `logical_disk`   |
| Data source         | Performance Data |
| Counters            | `LogicalDisk`    |
| Enabled by default? | Yes              |

## Flags

### `--collector.logical_disk.volume-include`

If given, a disk needs to match the include regular expression in order for the corresponding disk metrics to be reported

### `--collector.logical_disk.volume-exclude`

If given, a disk needs to *not* match the exclude regular expression in order for the corresponding disk metrics to be reported

### `--collector.logical_disk.enabled`

Comma-separated list of collectors to use. Available collectors: metrics, bitlocker_status. Defaults to metrics, if not specified.

## Metrics

| Name                                             | Description                                                                                                               | Type    | Labels                                                            |
|--------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------|---------|-------------------------------------------------------------------|
| `windows_logical_disk_info`                      | A metric with a constant '1' value labeled with logical disk information                                                  | gauge   | `disk`,`type`,`volume`,`volume_name`,`filesystem`,`serial_number` |
| `windows_logical_disk_requests_queued`           | Number of requests outstanding on the disk at the time the performance data is collected                                  | gauge   | `volume`                                                          |
| `windows_logical_disk_read_bytes_total`          | Rate at which bytes are transferred from the disk during read operations                                                  | counter | `volume`                                                          |
| `windows_logical_disk_reads_total`               | Rate of read operations on the disk                                                                                       | counter | `volume`                                                          |
| `windows_logical_disk_write_bytes_total`         | Rate at which bytes are transferred to the disk during write operations                                                   | counter | `volume`                                                          |
| `windows_logical_disk_writes_total`              | Rate of write operations on the disk                                                                                      | counter | `volume`                                                          |
| `windows_logical_disk_read_seconds_total`        | Seconds the disk was busy servicing read requests                                                                         | counter | `volume`                                                          |
| `windows_logical_disk_write_seconds_total`       | Seconds the disk was busy servicing write requests                                                                        | counter | `volume`                                                          |
| `windows_logical_disk_free_bytes`                | Unused space of the disk in bytes                                                                                         | gauge   | `volume`                                                          |
| `windows_logical_disk_size_bytes`                | Total size of the disk in bytes                                                                                           | gauge   | `volume`                                                          |
| `windows_logical_disk_idle_seconds_total`        | Seconds the disk was idle (not servicing read/write requests)                                                             | counter | `volume`                                                          |
| `windows_logical_disk_split_ios_total`           | Number of I/Os to the disk split into multiple I/Os                                                                       | counter | `volume`                                                          |
| `windows_logical_disk_read_latency_seconds_total` | Shows the average time, in seconds, of a read operation from the disk                                                     | counter | `volume`                                                          |
| `windows_logical_disk_write_latency_seconds_total` | Shows the average time, in seconds, of a write operation to the disk                                                      | counter | `volume`                                                          |
| `windows_logical_disk_read_write_latency_seconds_total` | Shows the time, in seconds, of the average disk transfer                                                                  | counter | `volume`                                                          |
| `windows_logical_disk_readonly`                  | Whether the logical disk is read-only                                                                                     | gauge   | `volume`                                                          |
| `windows_logical_disk_bitlocker_status`          | BitLocker status for the logical disk                                                                                     | gauge   | `volume`,`status`                                                 |

### Average queue length
`Avg. Disk Read Queue Length` and `Avg. Disk Write Queue Length` are averages that Windows computes from two samples.
Their raw values only grow: the queue length summed over time, in seconds.
These raw values are the same as the ones of `% Disk Read Time` and `% Disk Write Time`,
which are exposed as `windows_logical_disk_read_seconds_total` and `windows_logical_disk_write_seconds_total`.

The former `windows_logical_disk_avg_read_requests_queued` and `windows_logical_disk_avg_write_requests_queued`
gauges exposed these raw values and were removed in 0.32.
Use the rate of the `*_seconds_total` counters for the average queue length instead:
```
rate(windows_logical_disk_read_seconds_total[5m])
rate(windows_logical_disk_write_seconds_total[5m])
```

### Size metrics
The `free_bytes` and `size_bytes` metrics are read from the volume on every scrape. Both are the free and total bytes
of the whole volume, regardless of NTFS disk quotas of the account running windows_exporter.

If the volume can't be queried, they fall back to the `Free Megabytes` and `% Free Space` performance counters,
which Windows updates only every 10-15 minutes and in whole megabytes. On Windows Server 2016, which lacks
`GetDiskSpaceInformation`, `size_bytes` also falls back to the performance counter if a disk quota applies to the account.

### Volume information
The `disk` and `type` labels of `windows_logical_disk_info` are cached for up to 10 minutes.
The other labels and `windows_logical_disk_readonly` are read on every scrape. `windows_logical_disk_readonly` is not
reported for removable and CD-ROM drives without media.

### BitLocker status
The `bitlocker_status` sub-collector reads the BitLocker status through `fveapi.dll`, the BitLocker API of Windows.
It works when windows_exporter runs as a service (`LocalSystem`) and as a non-elevated user.
The API is undocumented, so the status is derived from reverse-engineered flags.

The `status` label has one of these values:

| Status                   | Meaning                                                                   |
|--------------------------|---------------------------------------------------------------------------|
| `on`                     | The volume is fully encrypted and protection is active                    |
| `off`                    | The volume is not encrypted                                               |
| `encrypting`             | Encryption is in progress or paused                                       |
| `decrypting`             | Decryption is in progress or paused                                       |
| `suspended`              | The volume is fully encrypted, but protection is suspended                |
| `locked`                 | The volume is locked                                                      |
| `waiting_for_activation` | The volume is encrypted with a clear key, but no key protector exists yet |
| `unknown`                | The status flags do not match a known state                               |
| `disabled`               | Not reported anymore, always 0. Kept for compatibility                    |

A fully encrypted volume is only reported as `on` if protection is active.

### Example metric
Query the rate of write operations to a disk
```
rate(windows_logical_disk_read_bytes_total{instance="localhost", volume=~"C:"}[2m])
```

Logical Volume information
```
windows_logical_disk_info{disk="0",filesystem="FAT32",serial_number="5C1A2B3D",type="fixed",volume="HarddiskVolume2",volume_name=""} 1
windows_logical_disk_info{disk="0",filesystem="NTFS",serial_number="BC9D5FA8",type="fixed",volume="HarddiskVolume3",volume_name=""} 1
windows_logical_disk_info{disk="0",filesystem="NTFS",serial_number="668EEC37",type="fixed",volume="C:",volume_name="Windows"} 1
windows_logical_disk_info{disk="1",filesystem="NTFS",serial_number="50AE953B",type="fixed",volume="D:",volume_name="Temporary Storage"} 1
windows_logical_disk_info{disk="1",filesystem="ReFS",serial_number="C69B59AD",type="fixed",volume="G:",volume_name="Volume"} 1
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

Disk Activity
```promql
  (
      rate(windows_logical_disk_read_seconds_total[2m])
    +
      rate(windows_logical_disk_write_seconds_total[2m])
  )
/
  (
      rate(windows_logical_disk_read_seconds_total[2m])
    +
      rate(windows_logical_disk_write_seconds_total[2m])
    +
      rate(windows_logical_disk_idle_seconds_total[2m])
  )
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

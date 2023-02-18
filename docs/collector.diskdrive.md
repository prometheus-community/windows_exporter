# diskdrive collector

The diskdrive collector exposes metrics about physical disks

|                     |                                                                                                                                                              |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Metric name prefix  | `diskdrive`                                                                                                                                                  |
| Classes             | [`Win32_PerfRawData_DNS_DNS`](https://learn.microsoft.com/en-us/windows/win32/cimwin32prov/win32-diskdrive)                                                  |
| Enabled by default? | No                                                                                                                                                           |

## Flags

None

## Metrics

| Name                      | Description                                                                                                                                                      | Type    | Labels |
| ------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- | ------ |
| `disk_drive_info`         | General identifiable information about the disk drive                                                                                                            | gauge   | name,caption,device_id,model |
| `disk_drive_availability` | The disk drive's current availability                                                                                                                            | gauge   | name,availability            |
| `disk_drive_partitions`   | Number of partitions on the drive                                                                                                                                | gauge   | name                         |
| `disk_drive_size`         | Size of the disk drive. It is calculated by multiplying the total number of cylinders, tracks in each cylinder, sectors in each track, and bytes in each sector. | gauge   | name                         |
| `disk_drive_status`       | Operational status of the drive                                                                                                                                  | gauge   | name,status                  |

## Alerting examples
**prometheus.rules**
```yaml
groups:
- name: Windows Disk Alerts
  rules:

  - alert: Drive_Status
    expr: windows_disk_drive_status{status="OK"} != 1
    for: 10m
    labels:
      severity: high
    annotations:
      summary: "Instance: {{ $labels.instance }} has drive status: {{ $labels.status }} on disk {{ $labels.name }}"
      description: "Drive Status Unhealthy"
```

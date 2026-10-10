# diskdrive collector

The diskdrive collector exposes metrics about physical disks

|                     |                                                                                                                                                              |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Metric name prefix  | `diskdrive`                                                                                                                                                  |
| Classes             | [`Win32_DiskDrive`](https://learn.microsoft.com/en-us/windows/win32/cimwin32prov/win32-diskdrive)                                                            |
| Enabled by default? | No                                                                                                                                                           |

At startup, the collector queries `Win32_DiskDrive` and reads the same
properties through SetupAPI, the configuration manager and disk IOCTLs. If both
results are identical, every scrape uses the native APIs. Otherwise the
collector keeps using WMI and logs both results at info level. Startup
therefore still requires WMI.

The native values follow the CIMWin32 provider:

- `name` and `device_id` are `PHYSICALDRIVE<n>` from `IOCTL_STORAGE_GET_DEVICE_NUMBER`.
- `model` is the device's friendly name. `caption` is the friendly name, the
  device description or the drive path, in that order.
- `diskdrive_size` is cylinders × tracks per cylinder × sectors per track × bytes
  per sector from `IOCTL_DISK_GET_DRIVE_GEOMETRY`. It can be smaller than the disk's
  byte size.
- `diskdrive_partitions` counts recognized MBR partitions, or GPT partitions
  other than the Microsoft reserved partition.
- `diskdrive_status` is derived from the device node status flags. A successful
  `IOCTL_STORAGE_PREDICT_FAILURE` call replaces it with `Pred Fail` or `OK`.
- `Win32_DiskDrive` never sets `Availability`, so all `diskdrive_availability`
  series are 0.

## Flags

None

## Metrics

| Name                      | Description                                                                                                                                                      | Type    | Labels |
| ------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- | ------ |
| `diskdrive_info`         | General identifiable information about the disk drive                                                                                                            | gauge   | name,caption,device_id,model |
| `diskdrive_availability` | The disk drive's current availability                                                                                                                            | gauge   | name,availability            |
| `diskdrive_partitions`   | Number of partitions on the drive                                                                                                                                | gauge   | name                         |
| `diskdrive_size`         | Size of the disk drive. It is calculated by multiplying the total number of cylinders, tracks in each cylinder, sectors in each track, and bytes in each sector. | gauge   | name                         |
| `diskdrive_status`       | Operational status of the drive                                                                                                                                  | gauge   | name,status                  |

## Alerting examples
**prometheus.rules**
```yaml
groups:
- name: Windows Disk Alerts
  rules:

  - alert: Drive_Status
    expr: windows_diskdrive_status{status="OK"} != 1
    for: 10m
    labels:
      severity: high
    annotations:
      summary: "Instance: {{ $labels.instance }} has drive status: {{ $labels.status }} on disk {{ $labels.name }}"
      description: "Drive Status Unhealthy"
```

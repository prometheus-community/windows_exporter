# physical_disk collector

The physical_disk collector exposes metrics about physical disks

|||
-|-
Metric name prefix  | `physical_disk`
Data source         | Perflib
Counters             | `physicalDisk` ([`Win32_PerfRawData_PerfDisk_physicalDisk`](https://msdn.microsoft.com/en-us/windows/hardware/aa394307(v=vs.71)))
Enabled by default? | Yes

## Flags

### `--collector.physical_disk.disk-include`

If given, a disk needs to match the include regexp in order for the corresponding disk metrics to be reported

### `--collector.physical_disk.disk-exclude`

If given, a disk needs to *not* match the exclude regexp in order for the corresponding disk metrics to be reported

## Metrics

| Name                                                   | Description                                                                                             | Type    | Labels |
|--------------------------------------------------------|---------------------------------------------------------------------------------------------------------|---------|--------|
| windows_physical_disk_requests_queued                  | The number of requests queued to the disk (PhysicalDisk.CurrentDiskQueueLength)                         | Gauge   | disk   |
| windows_physical_disk_read_bytes_total                 | The number of bytes transferred from the disk during read operations (PhysicalDisk.DiskReadBytesPerSec) | Counter | disk   |
| windows_physical_disk_reads_total                      | The number of read operations on the disk (PhysicalDisk.DiskReadsPerSec)                                | Counter | disk   |
| windows_physical_disk_write_bytes_total                | The number of bytes transferred to the disk during write operations (PhysicalDisk.DiskWriteBytesPerSec) | Counter | disk   |
| windows_physical_disk_writes_total                     | The number of write operations on the disk (PhysicalDisk.DiskWritesPerSec)                              | Counter | disk   |
| windows_physical_disk_read_seconds_total               | Seconds that the disk was busy servicing read requests (PhysicalDisk.PercentDiskReadTime)               | Counter | disk   |
| windows_physical_disk_write_seconds_total              | Seconds that the disk was busy servicing write requests (PhysicalDisk.PercentDiskWriteTime)             | Counter | disk   |
| windows_physical_disk_idle_seconds_total               | Seconds that the disk was idle (PhysicalDisk.PercentIdleTime)                                           | Counter | disk   |
| windows_physical_disk_split_ios_total                  | The number of I/Os to the disk that were split into multiple I/Os (PhysicalDisk.SplitIOPerSec)          | Counter | disk   |
| windows_physical_disk_read_latency_seconds_total       | The average time, in seconds, of a read operation from the disk (PhysicalDisk.AvgDiskSecPerRead)        | Counter | disk   |
| windows_physical_disk_write_latency_seconds_total      | The average time, in seconds, of a write operation to the disk (PhysicalDisk.AvgDiskSecPerWrite)        | Counter | disk   |
| windows_physical_disk_read_write_latency_seconds_total | The time, in seconds, of the average disk transfer (PhysicalDisk.AvgDiskSecPerTransfer)                 | Counter | disk   |


### Warning about size metrics
The `free_bytes` and `size_bytes` metrics are not updated in real time and might have a delay of 10-15min.
This is the same behavior as the windows performance counters.

### Example metric
Query the rate of write operations to a disk
```
rate(windows_physical_disk_read_bytes_total{instance="localhost", disk=~"0"}[2m])
```

## Useful queries
Calculate rate of total IOPS for disk
```
rate(windows_physical_disk_reads_total{instance="localhost", disk=~"0"}[2m]) + rate(windows_physical_disk_writes_total{instance="localhost", disk=~"0"}[2m])
```

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

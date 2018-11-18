# logical_disk collector

The logical_disk collector exposes metrics about logical disks (in contrast to physical disks)

|||
-|-
Metric name prefix  | `logical_disk`
Classes             | [`Win32_PerfRawData_PerfDisk_LogicalDisk`](https://msdn.microsoft.com/en-us/windows/hardware/aa394307(v=vs.71))
Enabled by default? | Yes

## Flags

### `--collector.logical_disk.volume-whitelist`

If given, a disk needs to match the whitelist regexp in order for the corresponding disk metrics to be reported

### `--collector.logical_disk.volume-blacklist`

If given, a disk needs to *not* match the blacklist regexp in order for the corresponding disk metrics to be reported

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`requests_queued` | _Not yet documented_ | gauge | `volume`
`read_bytes_total` | _Not yet documented_ | counter | `volume`
`reads_total` | _Not yet documented_ | counter | `volume`
`write_bytes_total` | _Not yet documented_ | counter | `volume`
`writes_total` | _Not yet documented_ | counter | `volume`
`read_seconds_total` | _Not yet documented_ | counter | `volume`
`write_seconds_total` | _Not yet documented_ | counter | `volume`
`free_bytes` | _Not yet documented_ | gauge | `volume`
`size_bytes` | _Not yet documented_ | gauge | `volume`
`idle_seconds_total` | _Not yet documented_ | counter | `volume`
`split_ios_total` | _Not yet documented_ | counter | `volume`

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

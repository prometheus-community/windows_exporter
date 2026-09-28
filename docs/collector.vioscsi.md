# vioscsi collector

The vioscsi collector exposes metrics from the VirtIO SCSI adapter driver via the `VioScsiExtendedInfoGuid` WMI class under `root\wmi`.

|||
-|-
Metric name prefix  | `vioscsi`
Classes             | `VioScsiExtendedInfoGuid`
Enabled by default? | No

## Flags

None.

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_vioscsi_info` | VirtIO SCSI adapter information | gauge | `adapter`, `indirect`, `event_index`, `dpc_redirection`, `concurrent_channels`, `interrupt_msg_ranges`, `completion_during_start_io`, `ring_packed`
`windows_vioscsi_queue_depth` | VirtIO SCSI queue depth | gauge | `adapter`
`windows_vioscsi_queues_count` | Number of VirtIO SCSI queues | gauge | `adapter`
`windows_vioscsi_physical_breaks` | Maximum number of scatter-gather segments | gauge | `adapter`
`windows_vioscsi_response_time_threshold_seconds` | VirtIO SCSI adapter response time tracing threshold in seconds | gauge | `adapter`

## Useful queries

### List VirtIO SCSI adapter capabilities
```promql
windows_vioscsi_info
```

## Alerting examples

```yaml
  - alert: "VioSCSIHighResponseTime"
    expr: "windows_vioscsi_response_time_threshold_seconds > 1"
    for: "5m"
    labels:
      urgency: "warning"
    annotations:
      summary: "VirtIO SCSI adapter response time threshold is high"
```

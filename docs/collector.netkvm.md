# netkvm collector

The netkvm collector exposes metrics from the VirtIO network adapter (NetKVM) driver via WMI classes under `root\wmi`.

|||
-|-
Metric name prefix  | `netkvm`
Classes             | `NetKvm_Config`, `NetKvm_Diag`
Enabled by default? | No

## Flags

None.

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_netkvm_info` | NetKVM adapter information | gauge | `adapter`, `standby`, `rsc_v4`, `rsc_v6`, `uso_v4`, `uso_v6`
`windows_netkvm_queues` | Number of virtio queues | gauge | `adapter`
`windows_netkvm_rx_free_buffers` | Number of currently free receive buffers in the first receive queue | gauge | `adapter`
`windows_netkvm_tx_queue_size` | Transmit queue size | gauge | `adapter`
`windows_netkvm_memory_allocated_bytes` | Allocated shared memory in bytes | gauge | `adapter`
`windows_netkvm_init_duration_seconds` | Adapter initialization time in seconds | gauge | `adapter`
`windows_netkvm_lazy_alloc_duration_seconds` | Lazy memory allocation time in seconds (-1 if not completed) | gauge | `adapter`
`windows_netkvm_tx_large_offload_total` | Number of LSO offloaded transmit frames | counter | `adapter`
`windows_netkvm_tx_udp_offload_total` | Number of USO offloaded transmit frames | counter | `adapter`
`windows_netkvm_tx_checksum_offload_total` | Number of checksum offloaded transmit frames | counter | `adapter`
`windows_netkvm_tx_copied_total` | Number of copied transmit packets | counter | `adapter`
`windows_netkvm_tx_dropped_total` | Number of dropped transmit packets | counter | `adapter`
`windows_netkvm_tx_min_free_buffers` | Transmit minimum free buffer watermark | gauge | `adapter`
`windows_netkvm_rx_coalesced_windows_total` | Number of receive frames coalesced by Windows RSC | counter | `adapter`
`windows_netkvm_rx_coalesced_host_total` | Number of receive frames coalesced by the host | counter | `adapter`
`windows_netkvm_rx_checksum_ok_total` | Number of receive frames with hardware-verified checksum | counter | `adapter`
`windows_netkvm_rx_priority_total` | Number of priority-tagged receive frames | counter | `adapter`
`windows_netkvm_rx_low_resources_total` | Number of receive indications with low-resource flag | counter | `adapter`
`windows_netkvm_rx_min_free_buffers` | Receive minimum free buffer watermark | gauge | `adapter`
`windows_netkvm_rss_device_supported` | Whether the device supports RSS (0 or 1) | gauge | `adapter`
`windows_netkvm_rss_device_hash_supported` | Whether the device reports hash (0 or 1) | gauge | `adapter`
`windows_netkvm_rss_active` | Whether RSS is currently active (0 or 1) | gauge | `adapter`
`windows_netkvm_rss_hits_total` | Number of RSS hash hits | counter | `adapter`
`windows_netkvm_rss_misses_total` | Number of RSS hash misses | counter | `adapter`
`windows_netkvm_rss_unclassified_total` | Number of RSS unclassified packets | counter | `adapter`
`windows_netkvm_rss_errors_total` | Number of RSS errors | counter | `adapter`
`windows_netkvm_ctrl_commands_total` | Number of virtio control commands sent | counter | `adapter`
`windows_netkvm_ctrl_commands_timed_out_total` | Number of virtio control commands that timed out | counter | `adapter`
`windows_netkvm_ctrl_commands_failed_total` | Number of virtio control commands that failed | counter | `adapter`

## Useful queries

### Detect dropped transmit packets
```promql
rate(windows_netkvm_tx_dropped_total[5m]) > 0
```

### Detect control command failures
```promql
rate(windows_netkvm_ctrl_commands_failed_total[5m]) > 0
```

## Alerting examples

```yaml
  - alert: "NetKVMTxDropped"
    expr: "rate(windows_netkvm_tx_dropped_total[5m]) > 0"
    for: "5m"
    labels:
      urgency: "warning"
    annotations:
      summary: "VirtIO network adapter is dropping transmit packets"
```

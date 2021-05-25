# container collector

The container collector exposes metrics about containers running on a Hyper-V system

|||
-|-
Metric name prefix  | `container`
Data source         | [hcsshim](https://github.com/Microsoft/hcsshim)
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_container_available` | Available | counter | `container_id`
`windows_container_count` | Number of containers | gauge | `container_id`
`windows_container_cpu_usage_seconds_kernelmode` | Run time in Kernel mode in Seconds | counter | `container_id`
`windows_container_cpu_usage_seconds_usermode` | Run Time in User mode in Seconds | counter | `container_id`
`windows_container_cpu_usage_seconds_total` | Total Run time in Seconds | counter | `container_id`
`windows_container_memory_usage_commit_bytes` | Memory Usage Commit Bytes | gauge | `container_id`
`windows_container_memory_usage_commit_peak_bytes` | Memory Usage Commit Peak Bytes | gauge | `container_id`
`windows_container_memory_usage_private_working_set_bytes` | Memory Usage Private Working Set Bytes | gauge | `container_id`
`windows_container_network_receive_bytes_total` | Bytes Received on Interface | counter | `container_id`, `interface`
`windows_container_network_receive_packets_total` | Packets Received on Interface | counter | `container_id`, `interface`
`windows_container_network_receive_packets_dropped_total` | Dropped Incoming Packets on Interface | counter | `container_id`, `interface`
`windows_container_network_transmit_bytes_total` | Bytes Sent on Interface | counter | `container_id`, `interface`
`windows_container_network_transmit_packets_total` | Packets Sent on Interface | counter | `container_id`, `interface`
`windows_container_network_transmit_packets_dropped_total` | Dropped Outgoing Packets on Interface | counter | `container_id`, `interface`

### Example metric
_windows_container_network_receive_bytes_total{container_id="docker://1bd30e8b8ac28cbd76a9b697b4d7bb9d760267b0733d1bc55c60024e98d1e43e",interface="822179E7-002C-4280-ABBA-28BCFE401826"} 9.3305343e+07_

This metric means that total _9.3305343e+07_ bytes received on interface _822179E7-002C-4280-ABBA-28BCFE401826_ for container _docker://1bd30e8b8ac28cbd76a9b697b4d7bb9d760267b0733d1bc55c60024e98d1e43e_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

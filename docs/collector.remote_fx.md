# remote_fx collector

The remote_fx collector exposes Performance Counters regarding the RemoteFX protocol (RDP). It exposes both network and graphics related performance counters.

|||
-|-
Metric name prefix  | `remote_fx`
Data source         | Performance Data
Classes             | [`Win32_PerfRawData_Counters_RemoteFXNetwork`](https://wutils.com/wmi/root/cimv2/win32_perfrawdata_counters_remotefxnetwork/), [`Win32_PerfRawData_Counters_RemoteFXGraphics`](https://wutils.com/wmi/root/cimv2/win32_perfrawdata_counters_remotefxgraphics), [more info...](https://docs.microsoft.com/en-us/azure/virtual-desktop/remotefx-graphics-performance-counters)
Enabled by default? | No


## Flags

None

## Metrics (Network)

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_remote_fx_net_base_udp_rtt_seconds` | Base UDP round-trip time (RTT) detected in seconds. | gauge | `session_name`
`windows_remote_fx_net_base_tcp_rtt_seconds` | Base TCP round-trip time (RTT) detected in seconds. | gauge | `session_name`
`windows_remote_fx_net_current_tcp_bandwidth_bytes_per_second` | TCP bandwidth detected in bytes per second. | gauge | `session_name`
`windows_remote_fx_net_current_tcp_rtt_seconds` | Average TCP round-trip time (RTT) detected in seconds. | gauge | `session_name`
`windows_remote_fx_net_current_udp_bandwidth_bytes_per_second` | UDP bandwidth detected in bytes per second. | gauge | `session_name`
`windows_remote_fx_net_current_udp_rtt_seconds` | Average UDP round-trip time (RTT) detected in seconds. | gauge | `session_name`
`windows_remote_fx_net_received_bytes_total` | Total number of bytes received since the connection started. | counter | `session_name`
`windows_remote_fx_net_sent_bytes_total` | Total number of bytes sent since the connection started. | counter | `session_name`
`windows_remote_fx_net_udp_packets_received_total` | Total number of packets received over UDP. | counter | `session_name`
`windows_remote_fx_net_udp_packets_sent_total` | Total number of packets sent over UDP. | counter | `session_name`
`windows_remote_fx_net_loss_total` | Numerator of the loss rate. | counter | `session_name`
`windows_remote_fx_net_loss_base_total` | Base (denominator) of the loss rate. | counter | `session_name`
`windows_remote_fx_net_fec_total` | Numerator of the Forward Error Correction (FEC) rate. | counter | `session_name`
`windows_remote_fx_net_fec_base_total` | Base (denominator) of the Forward Error Correction (FEC) rate. | counter | `session_name`
`windows_remote_fx_net_retransmission_total` | Numerator of the packet retransmission rate. | counter | `session_name`
`windows_remote_fx_net_retransmission_base_total` | Base (denominator) of the packet retransmission rate. | counter | `session_name`

Windows reports the loss, FEC and retransmission rates as sample fractions: two cumulative values whose
deltas give the ratio. Divide the rate of the numerator by the rate of its base, see [Useful queries](#useful-queries).

The RTT counters appear to be measured only with UDP transport; they stay `0` for sessions over TCP.
Windows keeps `Total Sent Bytes` and `Total Received Bytes` in 32 bits, so the byte counters wrap after 4 GiB,
which looks like a counter reset.

## Metrics (Graphics)

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_remote_fx_gfx_average_encoding_time_seconds` | Average frame encoding time in seconds. | gauge | `session_name`
`windows_remote_fx_gfx_frame_quality_ratio` | Quality of the output frame relative to the quality of the source frame (0-1). | gauge | `session_name`
`windows_remote_fx_gfx_frames_skipped_insufficient_resource_total` | Total number of frames skipped due to insufficient resources. `resource` is `client`, `network` or `server`. | counter | `session_name`, `resource`
`windows_remote_fx_gfx_graphics_compression_ratio` | Ratio of the number of bytes encoded to the number of bytes input. | gauge | `session_name`
`windows_remote_fx_gfx_input_frames_total` | Total number of source frames provided as input to RemoteFX graphics. | counter | `session_name`
`windows_remote_fx_gfx_output_frames_total` | Total number of frames sent to the client. | counter | `session_name`
`windows_remote_fx_gfx_source_frames_total` | Total number of frames composed by the source (DWM). | counter | `session_name`

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries

Packet loss ratio of each RDP session over the last 5 minutes:

```
rate(windows_remote_fx_net_loss_total[5m]) / rate(windows_remote_fx_net_loss_base_total[5m])
```

Frames skipped per second, by cause:

```
sum by (session_name, resource) (rate(windows_remote_fx_gfx_frames_skipped_insufficient_resource_total[5m]))
```

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

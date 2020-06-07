# remote_fx collector

The remote_fx collector exposes Performance Counters regarding the RemoteFX protocol (RDP). It exposes both network and graphics related performance counters.

|||
-|-
Metric name prefix  | `remote_fx`
Data source         | Perflib
Classes             | [`Win32_PerfRawData_Counters_RemoteFXNetwork`](https://wutils.com/wmi/root/cimv2/win32_perfrawdata_counters_remotefxnetwork/), [`Win32_PerfRawData_Counters_RemoteFXGraphics`](https://wutils.com/wmi/root/cimv2/win32_perfrawdata_counters_remotefxgraphics), [more info...](https://docs.microsoft.com/en-us/azure/virtual-desktop/remotefx-graphics-performance-counters)
Enabled by default? | No


## Flags

None

## Metrics (Network)

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_remote_fx_net_base_udp_rtt_seconds` | Base UDP round-trip time (RTT) detected in seconds. | gauge | `session_name`
`windows_remote_fx_net_base_tcp_rtt_seconds` | Base TCP round-trip time (RTT) detected in seconds. | gauge | `session_name`
`windows_remote_fx_net_current_tcp_bandwidth` | TCP Bandwidth detected in bytes per second. | gauge | `session_name`
`windows_remote_fx_net_current_tcp_rtt_seconds` | Average TCP round-trip time (RTT) detected in seconds. | gauge | `session_name`
`windows_remote_fx_net_current_udp_bandwidth` | UDP Bandwidth detected in bytes per second. | gauge | `session_name`
`windows_remote_fx_net_current_udp_rtt_seconds` | Average UDP round-trip time (RTT) detected in seconds. | gauge | `session_name`
`windows_remote_fx_net_received_bytes_total` | _Not yet documented_ | counter | `session_name`
`windows_remote_fx_net_sent_bytes_total` | _Not yet documented_ | counter | `session_name`
`windows_remote_fx_net_udp_packets_received_total` | Rate in packets per second at which packets are received over UDP. | counter | `session_name`
`windows_remote_fx_net_udp_packets_sent_total` | Rate in packets per second at which packets are sent over UDP. | counter | `session_name`

## Metrics (Graphics)

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_remote_fx_gfx_average_encoding_time_seconds` | Average frame encoding time. | gauge | `session_name`
`windows_remote_fx_gfx_frame_quality` | Quality of the output frame expressed as a percentage of the quality of the source frame. | gauge | `session_name`
`windows_remote_fx_gfx_frames_skipped_insufficient_resource_total` | Number of frames skipped per second due to insufficient resources. resources are client, server or network. | counter | `session_name`, `resource`
`windows_remote_fx_gfx_graphics_compression_ratio` | Ratio of the number of bytes encoded to the number of bytes input. | gauge | `session_name`
`windows_remote_fx_gfx_input_frames_total` | Number of sources frames provided as input to RemoteFX graphics per second. | counter | `session_name`
`windows_remote_fx_gfx_output_frames_total` | Number of frames sent to the client per second. | counter | `session_name`
`windows_remote_fx_gfx_source_frames_total` | Number of frames composed by the source (DWM) per second. | counter | `session_name`

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

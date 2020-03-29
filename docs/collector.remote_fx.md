# remote_fx collector

The remote_fx collector exposes Performance Counters regarding the RemoteFX protocol (RDP). It exposes both network and graphics related performance counters.

|||
-|-
Metric name prefix  | `remote_fx`
Classes             | [`Win32_PerfRawData_Counters_RemoteFXNetwork`](https://wutils.com/wmi/root/cimv2/win32_perfrawdata_counters_remotefxnetwork/), [`Win32_PerfRawData_Counters_RemoteFXGraphics`](https://wutils.com/wmi/root/cimv2/win32_perfrawdata_counters_remotefxgraphics), [more info...](https://docs.microsoft.com/en-us/azure/virtual-desktop/remotefx-graphics-performance-counters)
Enabled by default? | No


## Flags

None

## Metrics (Network)

Name | Description | Type | Labels
-----|-------------|------|-------
`wmi_remote_fx_net_base_tcp_rrt` | Base TCP round-trip time (RTT) detected in milliseconds. | gauge | `session`
`wmi_remote_fx_net_base_udp_rrt` | Base UDP round-trip time (RTT) detected in milliseconds. | gauge | `session`
`wmi_remote_fx_net_current_tcp_bandwidth` | TCP Bandwidth detected in thousands of bits per second (1000 bps). | gauge | `session`
`wmi_remote_fx_net_current_tcp_rtt` | Average TCP round-trip time (RTT) detected in milliseconds. | gauge | `session`
`wmi_remote_fx_net_current_udp_bandwidth` | UDP Bandwidth detected in thousands of bits per second (1000 bps). | gauge | `session`
`wmi_remote_fx_net_current_udp_rtt` | Average UDP round-trip time (RTT) detected in milliseconds. | gauge | `session`
`wmi_remote_fx_net_fec_rate` | Forward Error Correction (FEC) percentage | gauge | `session`
`wmi_remote_fx_net_fec_rate_base` | Forward Error Correction (FEC) percentage _Base value | gauge | `session`
`wmi_remote_fx_net_loss_rate` | Loss percentage | gauge | `session`
`wmi_remote_fx_net_loss_rate_base` | Loss percentage _Base value. | gauge | `session`
`wmi_remote_fx_net_retransmission_rate` | Percentage of packets that have been retransmitted | gauge | `session`
`wmi_remote_fx_net_retransmission_rate_Base` | Percentage of packets that have been retransmitted _base value | gauge | `session`
`wmi_remote_fx_net_tcp_received_rate` | Rate in bits per second (bps) at which data is received over TCP. | gauge | `session`
`wmi_remote_fx_net_tcp_sent_rate` | Rate in bits per second (bps) at which data is sent over TCP. | gauge | `session`
`wmi_remote_fx_net_total_received_rate` | Rate in bits per second (bps) at which data is received. | gauge | `session`
`wmi_remote_fx_net_total_sent_rate` | Rate in bits per second (bps) at which data is sent. | gauge | `session`
`wmi_remote_fx_net_udp_packets_received_persec` | Rate in packets per second at which packets are received over UDP. | gauge | `session`
`wmi_remote_fx_net_udp_packets_sent_persec` | Rate in packets per second at which packets are sent over UDP. | gauge | `session`
`wmi_remote_fx_net_udp_received_rate` | Rate in bits per second (bps) at which data is received over UDP. | gauge | `session`
`wmi_remote_fx_net_udp_sent_rate` | Rate in bits per second (bps) at which data is sent over UDP. | gauge | `session`

## Metrics (Graphics)

Name | Description | Type | Labels
-----|-------------|------|-------
`wmi_remote_fx_gfx_average_encoding_time` | Average frame encoding time. | gauge | `session`
`wmi_remote_fx_gfx_frame_quality` | Quality of the output frame expressed as a percentage of the quality of the source frame. | gauge | `session`
`wmi_remote_fx_gfx_frames_skipped_persec_insufficient_clt_res` | Number of frames skipped per second due to insufficient client resources. | gauge | `session`
`wmi_remote_fx_gfx_frames_skipped_persec_insufficient_net_res` | Number of frames skipped per second due to insufficient network resources. | gauge | `session`
`wmi_remote_fx_gfx_frames_skipped_persec_insufficient_srv_res` | Number of frames skipped per second due to insufficient server resources. | gauge | `session`
`wmi_remote_fx_gfx_graphics_compression_ratio` | Ratio of the number of bytes encoded to the number of bytes input. | gauge | `session`
`wmi_remote_fx_gfx_input_frames_persec` | Number of sources frames provided as input to RemoteFX graphics per second. | gauge | `session`
`wmi_remote_fx_gfx_output_frames_persec` | Number of frames sent to the client per second. | gauge | `session`
`wmi_remote_fx_gfx_source_frames_persec` | Number of frames composed by the source (DWM) per second. | gauge | `session`

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

# horizon_blast collector

The horizon_blast collector exposes metrics about Omnissa Horizon Blast protocol sessions

|                     |                      |
|---------------------|----------------------|
| Metric name prefix  | `horizon_blast`      |
| Source              | Performance counters |
| Enabled by default? | No                   |

## Flags

None

## Metrics

### Session Counters

| Name | Description | Type | Labels |
|------|-------------|------|--------|
| `windows_horizon_blast_session_automatic_reconnect_count_total` | The number of automatic reconnects for the session. | counter | None |
| `windows_horizon_blast_session_cumulative_received_bytes_udp_total` | Cumulative bytes received over UDP. | counter | None |
| `windows_horizon_blast_session_cumulative_transmitted_bytes_udp_total` | Cumulative bytes transmitted over UDP. | counter | None |
| `windows_horizon_blast_session_cumulative_received_bytes_tcp_total` | Cumulative bytes received over TCP. | counter | None |
| `windows_horizon_blast_session_cumulative_transmitted_bytes_tcp_total` | Cumulative bytes transmitted over TCP. | counter | None |
| `windows_horizon_blast_session_instantaneous_received_bytes_udp` | Instantaneous bytes received over UDP. | gauge | None |
| `windows_horizon_blast_session_instantaneous_transmitted_bytes_udp` | Instantaneous bytes transmitted over UDP. | gauge | None |
| `windows_horizon_blast_session_instantaneous_received_bytes_tcp` | Instantaneous bytes received over TCP. | gauge | None |
| `windows_horizon_blast_session_instantaneous_transmitted_bytes_tcp` | Instantaneous bytes transmitted over TCP. | gauge | None |
| `windows_horizon_blast_session_received_packets_total` | Total packets received. | counter | None |
| `windows_horizon_blast_session_transmitted_packets_total` | Total packets transmitted. | counter | None |
| `windows_horizon_blast_session_received_bytes_total` | Total bytes received. | counter | None |
| `windows_horizon_blast_session_transmitted_bytes_total` | Total bytes transmitted. | counter | None |
| `windows_horizon_blast_session_jitter_uplink_milliseconds` | Uplink jitter in milliseconds. | gauge | None |
| `windows_horizon_blast_session_rtt_milliseconds` | Round-trip time in milliseconds. | gauge | None |
| `windows_horizon_blast_session_packet_loss_uplink_percent` | Uplink packet loss percentage. | gauge | None |
| `windows_horizon_blast_session_estimated_bandwidth_uplink_kbps` | Estimated uplink bandwidth in Kbps. | gauge | None |

### Imaging Counters

| Name | Description | Type | Labels |
|------|-------------|------|--------|
| `windows_horizon_blast_imaging_encoder_type` | Encoder type used for imaging. | gauge | None |
| `windows_horizon_blast_imaging_total_dirty_frames_total` | Total number of dirty frames. | counter | None |
| `windows_horizon_blast_imaging_total_poll_total` | Total number of polls. | counter | None |
| `windows_horizon_blast_imaging_total_fbc_total` | Total number of FBC operations. | counter | None |
| `windows_horizon_blast_imaging_total_frames_total` | Total number of frames. | counter | None |
| `windows_horizon_blast_imaging_dirty_frames_per_second` | Dirty frames per second. | gauge | None |
| `windows_horizon_blast_imaging_poll_rate` | Poll rate. | gauge | None |
| `windows_horizon_blast_imaging_fbc_rate` | FBC rate. | gauge | None |
| `windows_horizon_blast_imaging_frames_per_second` | Frames per second. | gauge | None |
| `windows_horizon_blast_imaging_out_queueing_time_seconds` | Out queueing time in seconds. | gauge | None |
| `windows_horizon_blast_imaging_inbound_bandwidth_kbps` | Inbound bandwidth in Kbps. | gauge | None |
| `windows_horizon_blast_imaging_outbound_bandwidth_kbps` | Outbound bandwidth in Kbps. | gauge | None |
| `windows_horizon_blast_imaging_received_packets_total` | Total packets received. | counter | None |
| `windows_horizon_blast_imaging_transmitted_packets_total` | Total packets transmitted. | counter | None |
| `windows_horizon_blast_imaging_received_bytes_total` | Total bytes received. | counter | None |
| `windows_horizon_blast_imaging_transmitted_bytes_total` | Total bytes transmitted. | counter | None |

### Audio Counters

| Name | Description | Type | Labels |
|------|-------------|------|--------|
| `windows_horizon_blast_audio_out_queueing_time_seconds` | Out queueing time in seconds. | gauge | None |
| `windows_horizon_blast_audio_inbound_bandwidth_kbps` | Inbound bandwidth in Kbps. | gauge | None |
| `windows_horizon_blast_audio_outbound_bandwidth_kbps` | Outbound bandwidth in Kbps. | gauge | None |
| `windows_horizon_blast_audio_received_packets_total` | Total packets received. | counter | None |
| `windows_horizon_blast_audio_transmitted_packets_total` | Total packets transmitted. | counter | None |
| `windows_horizon_blast_audio_received_bytes_total` | Total bytes received. | counter | None |
| `windows_horizon_blast_audio_transmitted_bytes_total` | Total bytes transmitted. | counter | None |

### CDR Counters (Client Drive Redirection)

| Name | Description | Type | Labels |
|------|-------------|------|--------|
| `windows_horizon_blast_cdr_out_queueing_time_seconds` | Out queueing time in seconds. | gauge | None |
| `windows_horizon_blast_cdr_inbound_bandwidth_kbps` | Inbound bandwidth in Kbps. | gauge | None |
| `windows_horizon_blast_cdr_outbound_bandwidth_kbps` | Outbound bandwidth in Kbps. | gauge | None |
| `windows_horizon_blast_cdr_received_packets_total` | Total packets received. | counter | None |
| `windows_horizon_blast_cdr_transmitted_packets_total` | Total packets transmitted. | counter | None |
| `windows_horizon_blast_cdr_received_bytes_total` | Total bytes received. | counter | None |
| `windows_horizon_blast_cdr_transmitted_bytes_total` | Total bytes transmitted. | counter | None |

### Clipboard Counters

| Name | Description | Type | Labels |
|------|-------------|------|--------|
| `windows_horizon_blast_clipboard_out_queueing_time_seconds` | Out queueing time in seconds. | gauge | None |
| `windows_horizon_blast_clipboard_inbound_bandwidth_kbps` | Inbound bandwidth in Kbps. | gauge | None |
| `windows_horizon_blast_clipboard_outbound_bandwidth_kbps` | Outbound bandwidth in Kbps. | gauge | None |
| `windows_horizon_blast_clipboard_received_packets_total` | Total packets received. | counter | None |
| `windows_horizon_blast_clipboard_transmitted_packets_total` | Total packets transmitted. | counter | None |
| `windows_horizon_blast_clipboard_received_bytes_total` | Total bytes received. | counter | None |
| `windows_horizon_blast_clipboard_transmitted_bytes_total` | Total bytes transmitted. | counter | None |

### HTML5 MMR Counters

| Name | Description | Type | Labels |
|------|-------------|------|--------|
| `windows_horizon_blast_html5_mmr_out_queueing_time_seconds` | Out queueing time in seconds. | gauge | None |
| `windows_horizon_blast_html5_mmr_inbound_bandwidth_kbps` | Inbound bandwidth in Kbps. | gauge | None |
| `windows_horizon_blast_html5_mmr_outbound_bandwidth_kbps` | Outbound bandwidth in Kbps. | gauge | None |
| `windows_horizon_blast_html5_mmr_received_packets_total` | Total packets received. | counter | None |
| `windows_horizon_blast_html5_mmr_transmitted_packets_total` | Total packets transmitted. | counter | None |
| `windows_horizon_blast_html5_mmr_received_bytes_total` | Total bytes received. | counter | None |
| `windows_horizon_blast_html5_mmr_transmitted_bytes_total` | Total bytes transmitted. | counter | None |

### Other Feature Counters

| Name | Description | Type | Labels |
|------|-------------|------|--------|
| `windows_horizon_blast_other_feature_out_queueing_time_seconds` | Out queueing time in seconds. | gauge | `feature` |
| `windows_horizon_blast_other_feature_inbound_bandwidth_kbps` | Inbound bandwidth in Kbps. | gauge | `feature` |
| `windows_horizon_blast_other_feature_outbound_bandwidth_kbps` | Outbound bandwidth in Kbps. | gauge | `feature` |
| `windows_horizon_blast_other_feature_received_packets_total` | Total packets received. | counter | `feature` |
| `windows_horizon_blast_other_feature_transmitted_packets_total` | Total packets transmitted. | counter | `feature` |
| `windows_horizon_blast_other_feature_received_bytes_total` | Total bytes received. | counter | `feature` |
| `windows_horizon_blast_other_feature_transmitted_bytes_total` | Total bytes transmitted. | counter | `feature` |

### Printing Counters

| Name | Description | Type | Labels |
|------|-------------|------|--------|
| `windows_horizon_blast_printing_out_queueing_time_seconds` | Out queueing time in seconds. | gauge | None |
| `windows_horizon_blast_printing_inbound_bandwidth_kbps` | Inbound bandwidth in Kbps. | gauge | None |
| `windows_horizon_blast_printing_outbound_bandwidth_kbps` | Outbound bandwidth in Kbps. | gauge | None |
| `windows_horizon_blast_printing_received_packets_total` | Total packets received. | counter | None |
| `windows_horizon_blast_printing_transmitted_packets_total` | Total packets transmitted. | counter | None |
| `windows_horizon_blast_printing_received_bytes_total` | Total bytes received. | counter | None |
| `windows_horizon_blast_printing_transmitted_bytes_total` | Total bytes transmitted. | counter | None |

### RDE Server Counters

| Name | Description | Type | Labels |
|------|-------------|------|--------|
| `windows_horizon_blast_rde_server_out_queueing_time_seconds` | Out queueing time in seconds. | gauge | None |
| `windows_horizon_blast_rde_server_inbound_bandwidth_kbps` | Inbound bandwidth in Kbps. | gauge | None |
| `windows_horizon_blast_rde_server_outbound_bandwidth_kbps` | Outbound bandwidth in Kbps. | gauge | None |
| `windows_horizon_blast_rde_server_received_packets_total` | Total packets received. | counter | None |
| `windows_horizon_blast_rde_server_transmitted_packets_total` | Total packets transmitted. | counter | None |
| `windows_horizon_blast_rde_server_received_bytes_total` | Total bytes received. | counter | None |
| `windows_horizon_blast_rde_server_transmitted_bytes_total` | Total bytes transmitted. | counter | None |

### RTAV Counters (Real-Time Audio-Video)

| Name | Description | Type | Labels |
|------|-------------|------|--------|
| `windows_horizon_blast_rtav_out_queueing_time_seconds` | Out queueing time in seconds. | gauge | None |
| `windows_horizon_blast_rtav_inbound_bandwidth_kbps` | Inbound bandwidth in Kbps. | gauge | None |
| `windows_horizon_blast_rtav_outbound_bandwidth_kbps` | Outbound bandwidth in Kbps. | gauge | None |
| `windows_horizon_blast_rtav_received_packets_total` | Total packets received. | counter | None |
| `windows_horizon_blast_rtav_transmitted_packets_total` | Total packets transmitted. | counter | None |
| `windows_horizon_blast_rtav_received_bytes_total` | Total bytes received. | counter | None |
| `windows_horizon_blast_rtav_transmitted_bytes_total` | Total bytes transmitted. | counter | None |

### SDR Counters (Session Data Redirection)

| Name | Description | Type | Labels |
|------|-------------|------|--------|
| `windows_horizon_blast_sdr_out_queueing_time_seconds` | Out queueing time in seconds. | gauge | None |
| `windows_horizon_blast_sdr_inbound_bandwidth_kbps` | Inbound bandwidth in Kbps. | gauge | None |
| `windows_horizon_blast_sdr_outbound_bandwidth_kbps` | Outbound bandwidth in Kbps. | gauge | None |
| `windows_horizon_blast_sdr_received_packets_total` | Total packets received. | counter | None |
| `windows_horizon_blast_sdr_transmitted_packets_total` | Total packets transmitted. | counter | None |
| `windows_horizon_blast_sdr_received_bytes_total` | Total bytes received. | counter | None |
| `windows_horizon_blast_sdr_transmitted_bytes_total` | Total bytes transmitted. | counter | None |

### Serial Port and Scanner Counters

| Name | Description | Type | Labels |
|------|-------------|------|--------|
| `windows_horizon_blast_serial_port_scanner_out_queueing_time_seconds` | Out queueing time in seconds. | gauge | None |
| `windows_horizon_blast_serial_port_scanner_inbound_bandwidth_kbps` | Inbound bandwidth in Kbps. | gauge | None |
| `windows_horizon_blast_serial_port_scanner_outbound_bandwidth_kbps` | Outbound bandwidth in Kbps. | gauge | None |
| `windows_horizon_blast_serial_port_scanner_received_packets_total` | Total packets received. | counter | None |
| `windows_horizon_blast_serial_port_scanner_transmitted_packets_total` | Total packets transmitted. | counter | None |
| `windows_horizon_blast_serial_port_scanner_received_bytes_total` | Total bytes received. | counter | None |
| `windows_horizon_blast_serial_port_scanner_transmitted_bytes_total` | Total bytes transmitted. | counter | None |

### Smart Card Counters

| Name | Description | Type | Labels |
|------|-------------|------|--------|
| `windows_horizon_blast_smart_card_out_queueing_time_seconds` | Out queueing time in seconds. | gauge | None |
| `windows_horizon_blast_smart_card_inbound_bandwidth_kbps` | Inbound bandwidth in Kbps. | gauge | None |
| `windows_horizon_blast_smart_card_outbound_bandwidth_kbps` | Outbound bandwidth in Kbps. | gauge | None |
| `windows_horizon_blast_smart_card_received_packets_total` | Total packets received. | counter | None |
| `windows_horizon_blast_smart_card_transmitted_packets_total` | Total packets transmitted. | counter | None |
| `windows_horizon_blast_smart_card_received_bytes_total` | Total bytes received. | counter | None |
| `windows_horizon_blast_smart_card_transmitted_bytes_total` | Total bytes transmitted. | counter | None |

### USB Counters

| Name | Description | Type | Labels |
|------|-------------|------|--------|
| `windows_horizon_blast_usb_out_queueing_time_seconds` | Out queueing time in seconds. | gauge | None |
| `windows_horizon_blast_usb_inbound_bandwidth_kbps` | Inbound bandwidth in Kbps. | gauge | None |
| `windows_horizon_blast_usb_outbound_bandwidth_kbps` | Outbound bandwidth in Kbps. | gauge | None |
| `windows_horizon_blast_usb_received_packets_total` | Total packets received. | counter | None |
| `windows_horizon_blast_usb_transmitted_packets_total` | Total packets transmitted. | counter | None |
| `windows_horizon_blast_usb_received_bytes_total` | Total bytes received. | counter | None |
| `windows_horizon_blast_usb_transmitted_bytes_total` | Total bytes transmitted. | counter | None |

### View Scanner Counters

| Name | Description | Type | Labels |
|------|-------------|------|--------|
| `windows_horizon_blast_view_scanner_out_queueing_time_seconds` | Out queueing time in seconds. | gauge | None |
| `windows_horizon_blast_view_scanner_inbound_bandwidth_kbps` | Inbound bandwidth in Kbps. | gauge | None |
| `windows_horizon_blast_view_scanner_outbound_bandwidth_kbps` | Outbound bandwidth in Kbps. | gauge | None |
| `windows_horizon_blast_view_scanner_received_packets_total` | Total packets received. | counter | None |
| `windows_horizon_blast_view_scanner_transmitted_packets_total` | Total packets transmitted. | counter | None |
| `windows_horizon_blast_view_scanner_received_bytes_total` | Total bytes received. | counter | None |
| `windows_horizon_blast_view_scanner_transmitted_bytes_total` | Total bytes transmitted. | counter | None |

### Windows Media MMR Counters

| Name | Description | Type | Labels |
|------|-------------|------|--------|
| `windows_horizon_blast_windows_media_mmr_out_queueing_time_seconds` | Out queueing time in seconds. | gauge | None |
| `windows_horizon_blast_windows_media_mmr_inbound_bandwidth_kbps` | Inbound bandwidth in Kbps. | gauge | None |
| `windows_horizon_blast_windows_media_mmr_outbound_bandwidth_kbps` | Outbound bandwidth in Kbps. | gauge | None |
| `windows_horizon_blast_windows_media_mmr_received_packets_total` | Total packets received. | counter | None |
| `windows_horizon_blast_windows_media_mmr_transmitted_packets_total` | Total packets transmitted. | counter | None |
| `windows_horizon_blast_windows_media_mmr_received_bytes_total` | Total bytes received. | counter | None |
| `windows_horizon_blast_windows_media_mmr_transmitted_bytes_total` | Total bytes transmitted. | counter | None |

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

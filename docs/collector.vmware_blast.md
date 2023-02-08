# vmware_blast collector

The vmware_blast collector exposes metrics relating to VMware Blast sessions

|||
-|-
Metric name prefix  | `vmware_blast`
Classes             | `Win32_PerfRawData_Counters_VMwareBlastAudioCounters`,`Win32_PerfRawData_Counters_VMwareBlastCDRCounters`,`Win32_PerfRawData_Counters_VMwareBlastClipboardCounters`,`Win32_PerfRawData_Counters_VMwareBlastHTML5MMRCounters`,`Win32_PerfRawData_Counters_VMwareBlastImagingCounters`,`Win32_PerfRawData_Counters_VMwareBlastRTAVCounters`,`Win32_PerfRawData_Counters_VMwareBlastSerialPortandScannerCounters`,`Win32_PerfRawData_Counters_VMwareBlastSessionCounters`,`Win32_PerfRawData_Counters_VMwareBlastSkypeforBusinessControlCounters`,`Win32_PerfRawData_Counters_VMwareBlastThinPrintCounters`,`Win32_PerfRawData_Counters_VMwareBlastUSBCounters`,`Win32_PerfRawData_Counters_VMwareBlastWindowsMediaMMRCounters`
Enabled by default? | No

## Flags

None

## Metrics

Some of these metrics may not be collected, depending on the installation options chosen when installing the Horizon agent

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_vmware_blast_audio_received_bytes_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_audio_received_packets_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_audio_transmitted_bytes_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_audio_transmitted_packets_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_cdr_received_bytes_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_cdr_received_packets_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_cdr_transmitted_bytes_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_cdr_transmitted_packets_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_clipboard_received_bytes_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_clipboard_received_packets_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_clipboard_transmitted_bytes_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_clipboard_transmitted_packets_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_html5_mmr_received_bytes_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_html5_mmr_received_packets_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_html5_mmr_transmitted_bytes_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_html5_mmr_transmitted_packets_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_imaging_dirty_frames_per_second` | _Not yet documented_ | gauge | None
`windows_vmware_blast_imaging_fbc_rate` | _Not yet documented_ | gauge | None
`windows_vmware_blast_imaging_frames_per_second` | _Not yet documented_ | gauge | None
`windows_vmware_blast_imaging_poll_rate` | _Not yet documented_ | gauge | None
`windows_vmware_blast_imaging_received_bytes_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_imaging_received_packets_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_imaging_dirty_frames_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_imaging_fbc_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_imaging_frames_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_imaging_poll_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_imaging_transmitted_bytes_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_imaging_transmitted_packets_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_rtav_received_bytes_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_rtav_received_packets_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_rtav_transmitted_bytes_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_rtav_transmitted_packets_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_serial_port_and_scanner_received_bytes_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_serial_port_and_scanner_received_packets_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_serial_port_and_scanner_transmitted_bytes_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_serial_port_and_scanner_transmitted_packets_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_session_automatic_reconnect_count_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_session_cumlative_received_bytes_over_tcp_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_session_cumlative_received_bytes_over_udp_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_session_cumlative_transmitted_bytes_over_tcp_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_session_cumlative_transmitted_bytes_over_udp_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_session_estimated_bandwidth_uplink` | _Not yet documented_ | gauge | None
`windows_vmware_blast_session_instantaneous_received_bytes_over_tcp_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_session_instantaneous_received_bytes_over_udp_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_session_instantaneous_transmitted_bytes_over_tcp_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_session_instantaneous_transmitted_bytes_over_udp_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_session_jitter_uplink` | _Not yet documented_ | gauge | None
`windows_vmware_blast_session_packet_loss_uplink` | _Not yet documented_ | gauge | None
`windows_vmware_blast_session_received_bytes_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_session_received_packets_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_session_rtt` | _Not yet documented_ | gauge | None
`windows_vmware_blast_session_transmitted_bytes_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_session_transmitted_packets_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_skype_for_business_control_received_bytes_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_skype_for_business_control_received_packets_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_skype_for_business_control_transmitted_bytes_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_skype_for_business_control_transmitted_packets_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_thinprint_received_bytes_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_thinprint_received_packets_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_thinprint_transmitted_bytes_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_thinprint_transmitted_packets_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_usb_received_bytes_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_usb_received_packets_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_usb_transmitted_bytes_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_usb_transmitted_packets_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_windows_media_mmr_received_bytes_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_windows_media_mmr_received_packets_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_windows_media_mmr_transmitted_bytes_total` | _Not yet documented_ | counter | None
`windows_vmware_blast_windows_media_mmr_transmitted_packets_total` | _Not yet documented_ | counter | None

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

# hyperv collector

The hyperv collector exposes metrics about the Hyper-V hypervisor

|||
-|-
Metric name prefix  | `hyperv`
Classes             | `Win32_PerfRawData_VmmsVirtualMachineStats_HyperVVirtualMachineHealthSummary`<br/>`Win32_PerfRawData_VidPerfProvider_HyperVVMVidPartition`<br/>`Win32_PerfRawData_HvStats_HyperVHypervisorRootPartition`<br/>`Win32_PerfRawData_HvStats_HyperVHypervisor`<br/>`Win32_PerfRawData_HvStats_HyperVHypervisorRootVirtualProcessor`<br/>`Win32_PerfRawData_HvStats_HyperVHypervisorVirtualProcessor`<br/>`Win32_PerfRawData_NvspSwitchStats_HyperVVirtualSwitch`<br/>`Win32_PerfRawData_EthernetPerfProvider_HyperVLegacyNetworkAdapter`<br/>`Win32_PerfRawData_Counters_HyperVVirtualStorageDevice`<br/>`Win32_PerfRawData_NvspNicStats_HyperVVirtualNetworkAdapter`
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`wmi_hyper_health_critical` | _Not yet documented_ | counter | None
`wmi_hyper_health_ok` | _Not yet documented_ | counter | None
`wmi_hyper_vid_physical_pages_allocated` | _Not yet documented_ | counter | `vm`
`wmi_hyper_vid_preferred_numa_node_index` | _Not yet documented_ | counter | `vm`
`wmi_hyper_vid_remote_physical_pages` | _Not yet documented_ | counter | `vm`
`wmi_hyper_root_partition_address_spaces` | _Not yet documented_ | counter | None
`wmi_hyper_root_partition_attached_devices` | _Not yet documented_ | counter | None
`wmi_hyper_root_partition_deposited_pages` | _Not yet documented_ | counter | None
`wmi_hyper_root_partition_device_dma_errors` | _Not yet documented_ | counter | None
`wmi_hyper_root_partition_device_interrupt_errors` | _Not yet documented_ | counter | None
`wmi_hyper_root_partition_device_interrupt_mappings` | _Not yet documented_ | counter | None
`wmi_hyper_root_partition_device_interrupt_throttle_events` | _Not yet documented_ | counter | None
`wmi_hyper_root_partition_preferred_numa_node_index` | _Not yet documented_ | counter | None
`wmi_hyper_root_partition_gpa_space_modifications` | _Not yet documented_ | counter | None
`wmi_hyper_root_partition_io_tlb_flush_cost` | _Not yet documented_ | counter | None
`wmi_hyper_root_partition_io_tlb_flush` | _Not yet documented_ | counter | None
`wmi_hyper_root_partition_recommended_virtual_tlb_size` | _Not yet documented_ | counter | None
`wmi_hyper_root_partition_physical_pages_allocated` | _Not yet documented_ | counter | None
`wmi_hyper_root_partition_1G_device_pages` | _Not yet documented_ | counter | None
`wmi_hyper_root_partition_1G_gpa_pages` | _Not yet documented_ | counter | None
`wmi_hyper_root_partition_2M_device_pages` | _Not yet documented_ | counter | None
`wmi_hyper_root_partition_2M_gpa_pages` | _Not yet documented_ | counter | None
`wmi_hyper_root_partition_4K_device_pages` | _Not yet documented_ | counter | None
`wmi_hyper_root_partition_4K_gpa_pages` | _Not yet documented_ | counter | None
`wmi_hyper_root_partition_virtual_tlb_flush_entires` | _Not yet documented_ | counter | None
`wmi_hyper_root_partition_virtual_tlb_pages` | _Not yet documented_ | counter | None
`wmi_hyper_hypervisor_virtual_processors` | _Not yet documented_ | counter | None
`wmi_hyper_hypervisor_logical_processors` | _Not yet documented_ | counter | None
`wmi_hyper_host_cpu_guest_run_time` | _Not yet documented_ | counter | `core`
`wmi_hyper_host_cpu_hypervisor_run_time` | _Not yet documented_ | counter | `core`
`wmi_hyper_host_cpu_remote_run_time` | _Not yet documented_ | counter | `core`
`wmi_hyper_host_cpu_total_run_time` | _Not yet documented_ | counter | `core`
`wmi_hyper_vm_cpu_guest_run_time` | _Not yet documented_ | counter | `vm`, `core`
`wmi_hyper_vm_cpu_hypervisor_run_time` | _Not yet documented_ | counter | `vm`, `core`
`wmi_hyper_vm_cpu_remote_run_time` | _Not yet documented_ | counter | `vm`, `core`
`wmi_hyper_vm_cpu_total_run_time` | _Not yet documented_ | counter | `vm`, `core`
`wmi_hyper_vswitch_broadcast_packets_received_total` | _Not yet documented_ | counter | `vswitch`
`wmi_hyper_vswitch_broadcast_packets_sent_total` | _Not yet documented_ | counter | `vswitch`
`wmi_hyper_vswitch_bytes_total` | _Not yet documented_ | counter | `vswitch`
`wmi_hyper_vswitch_bytes_received_total` | _Not yet documented_ | counter | `vswitch`
`wmi_hyper_vswitch_bytes_sent_total` | _Not yet documented_ | counter | `vswitch`
`wmi_hyper_vswitch_directed_packets_received_total` | _Not yet documented_ | counter | `vswitch`
`wmi_hyper_vswitch_directed_packets_send_total` | _Not yet documented_ | counter | `vswitch`
`wmi_hyper_vswitch_dropped_packets_incoming_total` | _Not yet documented_ | counter | `vswitch`
`wmi_hyper_vswitch_dropped_packets_outcoming_total` | _Not yet documented_ | counter | `vswitch`
`wmi_hyper_vswitch_extensions_dropped_packets_incoming_total` | _Not yet documented_ | counter | `vswitch`
`wmi_hyper_vswitch_extensions_dropped_packets_outcoming_total` | _Not yet documented_ | counter | `vswitch`
`wmi_hyper_vswitch_learned_mac_addresses_total` | _Not yet documented_ | counter | `vswitch`
`wmi_hyper_vswitch_multicast_packets_received_total` | _Not yet documented_ | counter | `vswitch`
`wmi_hyper_vswitch_multicast_packets_sent_total` | _Not yet documented_ | counter | `vswitch`
`wmi_hyper_vswitch_number_of_send_channel_moves_total` | _Not yet documented_ | counter | `vswitch`
`wmi_hyper_vswitch_number_of_vmq_moves_total` | _Not yet documented_ | counter | `vswitch`
`wmi_hyper_vswitch_packets_flooded_total` | _Not yet documented_ | counter | `vswitch`
`wmi_hyper_vswitch_packets_total` | _Not yet documented_ | counter | `vswitch`
`wmi_hyper_vswitch_packets_received_total` | _Not yet documented_ | counter | `vswitch`
`wmi_hyper_vswitch_packets_sent_total` | _Not yet documented_ | counter | `vswitch`
`wmi_hyper_vswitch_purged_mac_addresses_total` | _Not yet documented_ | counter | `vswitch`
`wmi_hyper_ethernet_bytes_dropped` | _Not yet documented_ | counter | `adapter`
`wmi_hyper_ethernet_bytes_received` | _Not yet documented_ | counter | `adapter`
`wmi_hyper_ethernet_bytes_sent` | _Not yet documented_ | counter | `adapter`
`wmi_hyper_ethernet_frames_dropped` | _Not yet documented_ | counter | `adapter`
`wmi_hyper_ethernet_frames_received` | _Not yet documented_ | counter | `adapter`
`wmi_hyper_ethernet_frames_sent` | _Not yet documented_ | counter | `adapter`
`wmi_hyper_vm_device_error_count` | _Not yet documented_ | counter | `vm_device`
`wmi_hyper_vm_device_queue_length` | _Not yet documented_ | counter | `vm_device`
`wmi_hyper_vm_device_bytes_read` | _Not yet documented_ | counter | `vm_device`
`wmi_hyper_vm_device_operations_read` | _Not yet documented_ | counter | `vm_device`
`wmi_hyper_vm_device_bytes_written` | _Not yet documented_ | counter | `vm_device`
`wmi_hyper_vm_device_operations_written` | _Not yet documented_ | counter | `vm_device`
`wmi_hyper_vm_interface_bytes_received` | _Not yet documented_ | counter | `vm_interface`
`wmi_hyper_vm_interface_bytes_sent` | _Not yet documented_ | counter | `vm_interface`
`wmi_hyper_vm_interface_packets_incoming_dropped` | _Not yet documented_ | counter | `vm_interface`
`wmi_hyper_vm_interface_packets_outgoing_dropped` | _Not yet documented_ | counter | `vm_interface`
`wmi_hyper_vm_interface_packets_received` | _Not yet documented_ | counter | `vm_interface`
`wmi_hyper_vm_interface_packets_sent` | _Not yet documented_ | counter | `vm_interface`

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

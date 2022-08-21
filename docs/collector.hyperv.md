 # hyperv collector

The hyperv collector exposes metrics about the Hyper-V hypervisor

|||
-|-
Metric name prefix  | `hyperv`
Classes             | `Win32_PerfRawData_VmmsVirtualMachineStats_HyperVVirtualMachineHealthSummary`<br/>`Win32_PerfRawData_VidPerfProvider_HyperVVMVidPartition`<br/>`Win32_PerfRawData_HvStats_HyperVHypervisorRootPartition`<br/>`Win32_PerfRawData_HvStats_HyperVHypervisor`<br/>`Win32_PerfRawData_HvStats_HyperVHypervisorLogicalProcessor`<br/>`Win32_PerfRawData_HvStats_HyperVHypervisorRootVirtualProcessor`<br/>`Win32_PerfRawData_HvStats_HyperVHypervisorVirtualProcessor`<br/>`Win32_PerfRawData_NvspSwitchStats_HyperVVirtualSwitch`<br/>`Win32_PerfRawData_EthernetPerfProvider_HyperVLegacyNetworkAdapter`<br/>`Win32_PerfRawData_Counters_HyperVVirtualStorageDevice`<br/>`Win32_PerfRawData_NvspNicStats_HyperVVirtualNetworkAdapter`
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_hyperv_health_critical` | _Not yet documented_ | counter | None
`windows_hyperv_health_ok` | _Not yet documented_ | counter | None
`windows_hyperv_vid_physical_pages_allocated` | _Not yet documented_ | counter | `vm`
`windows_hyperv_vid_preferred_numa_node_index` | _Not yet documented_ | counter | `vm`
`windows_hyperv_vid_remote_physical_pages` | _Not yet documented_ | counter | `vm`
`windows_hyperv_root_partition_address_spaces` | _Not yet documented_ | counter | None
`windows_hyperv_root_partition_attached_devices` | _Not yet documented_ | counter | None
`windows_hyperv_root_partition_deposited_pages` | _Not yet documented_ | counter | None
`windows_hyperv_root_partition_device_dma_errors` | _Not yet documented_ | counter | None
`windows_hyperv_root_partition_device_interrupt_errors` | _Not yet documented_ | counter | None
`windows_hyperv_root_partition_device_interrupt_mappings` | _Not yet documented_ | counter | None
`windows_hyperv_root_partition_device_interrupt_throttle_events` | _Not yet documented_ | counter | None
`windows_hyperv_root_partition_preferred_numa_node_index` | _Not yet documented_ | counter | None
`windows_hyperv_root_partition_gpa_space_modifications` | _Not yet documented_ | counter | None
`windows_hyperv_root_partition_io_tlb_flush_cost` | _Not yet documented_ | counter | None
`windows_hyperv_root_partition_io_tlb_flush` | _Not yet documented_ | counter | None
`windows_hyperv_root_partition_recommended_virtual_tlb_size` | _Not yet documented_ | counter | None
`windows_hyperv_root_partition_physical_pages_allocated` | _Not yet documented_ | counter | None
`windows_hyperv_root_partition_1G_device_pages` | _Not yet documented_ | counter | None
`windows_hyperv_root_partition_1G_gpa_pages` | _Not yet documented_ | counter | None
`windows_hyperv_root_partition_2M_device_pages` | _Not yet documented_ | counter | None
`windows_hyperv_root_partition_2M_gpa_pages` | _Not yet documented_ | counter | None
`windows_hyperv_root_partition_4K_device_pages` | _Not yet documented_ | counter | None
`windows_hyperv_root_partition_4K_gpa_pages` | _Not yet documented_ | counter | None
`windows_hyperv_root_partition_virtual_tlb_flush_entires` | _Not yet documented_ | counter | None
`windows_hyperv_root_partition_virtual_tlb_pages` | _Not yet documented_ | counter | None
`windows_hyperv_hypervisor_virtual_processors` | _Not yet documented_ | counter | None
`windows_hyperv_hypervisor_logical_processors` | _Not yet documented_ | counter | None
`windows_hyperv_host_lp_guest_run_time_percent` | _Not yet documented_ | counter | `core`
`windows_hyperv_host_lp_hypervisor_run_time_percent` | _Not yet documented_ | counter | `core`
`windows_hyperv_host_lp_total_run_time_percent` | _Not yet documented_ | counter | `core`
`windows_hyperv_host_cpu_guest_run_time` | _Not yet documented_ | counter | `core`
`windows_hyperv_host_cpu_hypervisor_run_time` | _Not yet documented_ | counter | `core`
`windows_hyperv_host_cpu_remote_run_time` | _Not yet documented_ | counter | `core`
`windows_hyperv_host_cpu_total_run_time` | _Not yet documented_ | counter | `core`
`windows_hyperv_vm_cpu_guest_run_time` | _Not yet documented_ | counter | `vm`, `core`
`windows_hyperv_vm_cpu_hypervisor_run_time` | _Not yet documented_ | counter | `vm`, `core`
`windows_hyperv_vm_cpu_remote_run_time` | _Not yet documented_ | counter | `vm`, `core`
`windows_hyperv_vm_memory_added_total` | _Not yet documented_ | counter | `vm`
`windows_hyperv_vm_memory_pressure_average` | _Not yet documented_ | gauge | `vm`
`windows_hyperv_vm_memory_pressure_current` | _Not yet documented_ | counter | `vm`
`windows_hyperv_vm_memory_physical_guest_visible` | _Not yet documented_ | gauge | `vm`
`windows_hyperv_vm_memory_pressure_maximum` | _Not yet documented_ | gauge | `vm`
`windows_hyperv_vm_memory_add_operations_total` | _Not yet documented_ | counter | `vm`
`windows_hyperv_vm_memory_remove_operations_total` | _Not yet documented_ | counter | `vm`
`windows_hyperv_vm_memory_pressure_minumim` | _Not yet documented_ | gauge | `vm`
`windows_hyperv_vm_memory_physical` | _Not yet documented_ | gauge | `vm`
`windows_hyperv_vm_memory_removed_total` | _Not yet documented_ | counter | `vm`
`windows_hyperv_vm_cpu_total_run_time` | _Not yet documented_ | counter | `vm`, `core`
`windows_hyperv_vswitch_broadcast_packets_received_total` | _Not yet documented_ | counter | `vswitch`
`windows_hyperv_vswitch_broadcast_packets_sent_total` | _Not yet documented_ | counter | `vswitch`
`windows_hyperv_vswitch_bytes_total` | _Not yet documented_ | counter | `vswitch`
`windows_hyperv_vswitch_bytes_received_total` | _Not yet documented_ | counter | `vswitch`
`windows_hyperv_vswitch_bytes_sent_total` | _Not yet documented_ | counter | `vswitch`
`windows_hyperv_vswitch_directed_packets_received_total` | _Not yet documented_ | counter | `vswitch`
`windows_hyperv_vswitch_directed_packets_send_total` | _Not yet documented_ | counter | `vswitch`
`windows_hyperv_vswitch_dropped_packets_incoming_total` | _Not yet documented_ | counter | `vswitch`
`windows_hyperv_vswitch_dropped_packets_outcoming_total` | _Not yet documented_ | counter | `vswitch`
`windows_hyperv_vswitch_extensions_dropped_packets_incoming_total` | _Not yet documented_ | counter | `vswitch`
`windows_hyperv_vswitch_extensions_dropped_packets_outcoming_total` | _Not yet documented_ | counter | `vswitch`
`windows_hyperv_vswitch_learned_mac_addresses_total` | _Not yet documented_ | counter | `vswitch`
`windows_hyperv_vswitch_multicast_packets_received_total` | _Not yet documented_ | counter | `vswitch`
`windows_hyperv_vswitch_multicast_packets_sent_total` | _Not yet documented_ | counter | `vswitch`
`windows_hyperv_vswitch_number_of_send_channel_moves_total` | _Not yet documented_ | counter | `vswitch`
`windows_hyperv_vswitch_number_of_vmq_moves_total` | _Not yet documented_ | counter | `vswitch`
`windows_hyperv_vswitch_packets_flooded_total` | _Not yet documented_ | counter | `vswitch`
`windows_hyperv_vswitch_packets_total` | _Not yet documented_ | counter | `vswitch`
`windows_hyperv_vswitch_packets_received_total` | _Not yet documented_ | counter | `vswitch`
`windows_hyperv_vswitch_packets_sent_total` | _Not yet documented_ | counter | `vswitch`
`windows_hyperv_vswitch_purged_mac_addresses_total` | _Not yet documented_ | counter | `vswitch`
`windows_hyperv_ethernet_bytes_dropped` | _Not yet documented_ | counter | `adapter`
`windows_hyperv_ethernet_bytes_received` | _Not yet documented_ | counter | `adapter`
`windows_hyperv_ethernet_bytes_sent` | _Not yet documented_ | counter | `adapter`
`windows_hyperv_ethernet_frames_dropped` | _Not yet documented_ | counter | `adapter`
`windows_hyperv_ethernet_frames_received` | _Not yet documented_ | counter | `adapter`
`windows_hyperv_ethernet_frames_sent` | _Not yet documented_ | counter | `adapter`
`windows_hyperv_vm_device_error_count` | _Not yet documented_ | counter | `vm_device`
`windows_hyperv_vm_device_queue_length` | _Not yet documented_ | counter | `vm_device`
`windows_hyperv_vm_device_bytes_read` | _Not yet documented_ | counter | `vm_device`
`windows_hyperv_vm_device_operations_read` | _Not yet documented_ | counter | `vm_device`
`windows_hyperv_vm_device_bytes_written` | _Not yet documented_ | counter | `vm_device`
`windows_hyperv_vm_device_operations_written` | _Not yet documented_ | counter | `vm_device`
`windows_hyperv_vm_interface_bytes_received` | _Not yet documented_ | counter | `vm_interface`
`windows_hyperv_vm_interface_bytes_sent` | _Not yet documented_ | counter | `vm_interface`
`windows_hyperv_vm_interface_packets_incoming_dropped` | _Not yet documented_ | counter | `vm_interface`
`windows_hyperv_vm_interface_packets_outgoing_dropped` | _Not yet documented_ | counter | `vm_interface`
`windows_hyperv_vm_interface_packets_received` | _Not yet documented_ | counter | `vm_interface`
`windows_hyperv_vm_interface_packets_sent` | _Not yet documented_ | counter | `vm_interface`

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
Percent of physical CPU resources used per VM (on instance "localhost")
```
(sum (rate(windows_hyperv_vm_cpu_hypervisor_run_time{instance="localhost"}[1m]))) / ignoring(vm) group_left max (windows_cs_logical_processors{instance="localhost"}) / 100000
```
Percent of physical CPU resources used by all VMs (on all monitored hosts)
```
(sum by (instance)(rate(windows_hyperv_vm_cpu_total_run_time{}[1m]))) / max by (instance)(windows_cs_logical_processors{}) / 100000
```
Percent of physical CPU resources by the hosts themselves (on all monitored hosts)
```
(sum by (instance)(rate(windows_hyperv_host_cpu_total_run_time{}[1m]))) / sum by (instance)(windows_cs_logical_processors{}) / 100000
```
Percent of physical CPU resources by the hypervisor (on all monitored hosts)
```
(sum by (instance)(rate(windows_hyperv_host_lp_total_run_time_percent{}[1m]))) / sum by (instance)(windows_hyperv_hypervisor_logical_processors{}) / 100000
```

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

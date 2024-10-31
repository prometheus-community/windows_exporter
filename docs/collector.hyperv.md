 # hyperv collector

The hyperv collector exposes metrics about the Hyper-V hypervisor

|                     |                      |
|---------------------|----------------------|
| Metric name prefix  | `hyperv`             |
| Source              | Performance counters |
| Enabled by default? | No                   |

## Flags

None

## Metrics

### All

| Name                                                                | Description          | Type    | Labels         |
|---------------------------------------------------------------------|----------------------|---------|----------------|
| `windows_hyperv_health_critical`                                    | _Not yet documented_ | counter | None           |
| `windows_hyperv_health_ok`                                          | _Not yet documented_ | counter | None           |
| `windows_hyperv_vid_physical_pages_allocated`                       | _Not yet documented_ | counter | `vm`           |
| `windows_hyperv_vid_preferred_numa_node_index`                      | _Not yet documented_ | counter | `vm`           |
| `windows_hyperv_vid_remote_physical_pages`                          | _Not yet documented_ | counter | `vm`           |
| `windows_hyperv_root_partition_address_spaces`                      | _Not yet documented_ | counter | None           |
| `windows_hyperv_root_partition_attached_devices`                    | _Not yet documented_ | counter | None           |
| `windows_hyperv_root_partition_deposited_pages`                     | _Not yet documented_ | counter | None           |
| `windows_hyperv_root_partition_device_dma_errors`                   | _Not yet documented_ | counter | None           |
| `windows_hyperv_root_partition_device_interrupt_errors`             | _Not yet documented_ | counter | None           |
| `windows_hyperv_root_partition_device_interrupt_mappings`           | _Not yet documented_ | counter | None           |
| `windows_hyperv_root_partition_device_interrupt_throttle_events`    | _Not yet documented_ | counter | None           |
| `windows_hyperv_root_partition_preferred_numa_node_index`           | _Not yet documented_ | counter | None           |
| `windows_hyperv_root_partition_gpa_space_modifications`             | _Not yet documented_ | counter | None           |
| `windows_hyperv_root_partition_io_tlb_flush_cost`                   | _Not yet documented_ | counter | None           |
| `windows_hyperv_root_partition_io_tlb_flush`                        | _Not yet documented_ | counter | None           |
| `windows_hyperv_root_partition_recommended_virtual_tlb_size`        | _Not yet documented_ | counter | None           |
| `windows_hyperv_root_partition_physical_pages_allocated`            | _Not yet documented_ | counter | None           |
| `windows_hyperv_root_partition_1G_device_pages`                     | _Not yet documented_ | counter | None           |
| `windows_hyperv_root_partition_1G_gpa_pages`                        | _Not yet documented_ | counter | None           |
| `windows_hyperv_root_partition_2M_device_pages`                     | _Not yet documented_ | counter | None           |
| `windows_hyperv_root_partition_2M_gpa_pages`                        | _Not yet documented_ | counter | None           |
| `windows_hyperv_root_partition_4K_device_pages`                     | _Not yet documented_ | counter | None           |
| `windows_hyperv_root_partition_4K_gpa_pages`                        | _Not yet documented_ | counter | None           |
| `windows_hyperv_root_partition_virtual_tlb_flush_entires`           | _Not yet documented_ | counter | None           |
| `windows_hyperv_root_partition_virtual_tlb_pages`                   | _Not yet documented_ | counter | None           |
| `windows_hyperv_hypervisor_virtual_processors`                      | _Not yet documented_ | counter | None           |
| `windows_hyperv_hypervisor_logical_processors`                      | _Not yet documented_ | counter | None           |
| `windows_hyperv_host_lp_guest_run_time_percent`                     | _Not yet documented_ | counter | `core`         |
| `windows_hyperv_host_lp_hypervisor_run_time_percent`                | _Not yet documented_ | counter | `core`         |
| `windows_hyperv_host_lp_total_run_time_percent`                     | _Not yet documented_ | counter | `core`         |
| `windows_hyperv_host_cpu_guest_run_time`                            | _Not yet documented_ | counter | `core`         |
| `windows_hyperv_host_cpu_hypervisor_run_time`                       | _Not yet documented_ | counter | `core`         |
| `windows_hyperv_host_cpu_remote_run_time`                           | _Not yet documented_ | counter | `core`         |
| `windows_hyperv_host_cpu_total_run_time`                            | _Not yet documented_ | counter | `core`         |
| `windows_hyperv_host_cpu_wait_time_per_dispatch_total`              | _Not yet documented_ | counter | `core`         |
| `windows_hyperv_vm_cpu_guest_run_time`                              | _Not yet documented_ | counter | `vm`, `core`   |
| `windows_hyperv_vm_cpu_hypervisor_run_time`                         | _Not yet documented_ | counter | `vm`, `core`   |
| `windows_hyperv_vm_cpu_remote_run_time`                             | _Not yet documented_ | counter | `vm`, `core`   |
| `windows_hyperv_vm_cpu_wait_time_per_dispatch_total`                | _Not yet documented_ | counter | `vm`, `core`   |
| `windows_hyperv_vm_cpu_total_run_time`                              | _Not yet documented_ | counter | `vm`, `core`   |
| `windows_hyperv_vswitch_broadcast_packets_received_total`           | _Not yet documented_ | counter | `vswitch`      |
| `windows_hyperv_vswitch_broadcast_packets_sent_total`               | _Not yet documented_ | counter | `vswitch`      |
| `windows_hyperv_vswitch_bytes_total`                                | _Not yet documented_ | counter | `vswitch`      |
| `windows_hyperv_vswitch_bytes_received_total`                       | _Not yet documented_ | counter | `vswitch`      |
| `windows_hyperv_vswitch_bytes_sent_total`                           | _Not yet documented_ | counter | `vswitch`      |
| `windows_hyperv_vswitch_directed_packets_received_total`            | _Not yet documented_ | counter | `vswitch`      |
| `windows_hyperv_vswitch_directed_packets_send_total`                | _Not yet documented_ | counter | `vswitch`      |
| `windows_hyperv_vswitch_dropped_packets_incoming_total`             | _Not yet documented_ | counter | `vswitch`      |
| `windows_hyperv_vswitch_dropped_packets_outcoming_total`            | _Not yet documented_ | counter | `vswitch`      |
| `windows_hyperv_vswitch_extensions_dropped_packets_incoming_total`  | _Not yet documented_ | counter | `vswitch`      |
| `windows_hyperv_vswitch_extensions_dropped_packets_outcoming_total` | _Not yet documented_ | counter | `vswitch`      |
| `windows_hyperv_vswitch_learned_mac_addresses_total`                | _Not yet documented_ | counter | `vswitch`      |
| `windows_hyperv_vswitch_multicast_packets_received_total`           | _Not yet documented_ | counter | `vswitch`      |
| `windows_hyperv_vswitch_multicast_packets_sent_total`               | _Not yet documented_ | counter | `vswitch`      |
| `windows_hyperv_vswitch_number_of_send_channel_moves_total`         | _Not yet documented_ | counter | `vswitch`      |
| `windows_hyperv_vswitch_number_of_vmq_moves_total`                  | _Not yet documented_ | counter | `vswitch`      |
| `windows_hyperv_vswitch_packets_flooded_total`                      | _Not yet documented_ | counter | `vswitch`      |
| `windows_hyperv_vswitch_packets_total`                              | _Not yet documented_ | counter | `vswitch`      |
| `windows_hyperv_vswitch_packets_received_total`                     | _Not yet documented_ | counter | `vswitch`      |
| `windows_hyperv_vswitch_packets_sent_total`                         | _Not yet documented_ | counter | `vswitch`      |
| `windows_hyperv_vswitch_purged_mac_addresses_total`                 | _Not yet documented_ | counter | `vswitch`      |
| `windows_hyperv_ethernet_bytes_dropped`                             | _Not yet documented_ | counter | `adapter`      |
| `windows_hyperv_ethernet_bytes_received`                            | _Not yet documented_ | counter | `adapter`      |
| `windows_hyperv_ethernet_bytes_sent`                                | _Not yet documented_ | counter | `adapter`      |
| `windows_hyperv_ethernet_frames_dropped`                            | _Not yet documented_ | counter | `adapter`      |
| `windows_hyperv_ethernet_frames_received`                           | _Not yet documented_ | counter | `adapter`      |
| `windows_hyperv_ethernet_frames_sent`                               | _Not yet documented_ | counter | `adapter`      |
| `windows_hyperv_vm_device_error_count`                              | _Not yet documented_ | counter | `vm_device`    |
| `windows_hyperv_vm_device_queue_length`                             | _Not yet documented_ | counter | `vm_device`    |
| `windows_hyperv_vm_device_bytes_read`                               | _Not yet documented_ | counter | `vm_device`    |
| `windows_hyperv_vm_device_operations_read`                          | _Not yet documented_ | counter | `vm_device`    |
| `windows_hyperv_vm_device_bytes_written`                            | _Not yet documented_ | counter | `vm_device`    |
| `windows_hyperv_vm_device_operations_written`                       | _Not yet documented_ | counter | `vm_device`    |
| `windows_hyperv_vm_interface_bytes_received`                        | _Not yet documented_ | counter | `vm_interface` |
| `windows_hyperv_vm_interface_bytes_sent`                            | _Not yet documented_ | counter | `vm_interface` |
| `windows_hyperv_vm_interface_packets_incoming_dropped`              | _Not yet documented_ | counter | `vm_interface` |
| `windows_hyperv_vm_interface_packets_outgoing_dropped`              | _Not yet documented_ | counter | `vm_interface` |
| `windows_hyperv_vm_interface_packets_received`                      | _Not yet documented_ | counter | `vm_interface` |
| `windows_hyperv_vm_interface_packets_sent`                          | _Not yet documented_ | counter | `vm_interface` |





### Hyper-V Dynamic Memory Balancer

Some metrics explained: https://learn.microsoft.com/en-us/archive/blogs/chrisavis/monitoring-dynamic-memory-in-windows-server-hyper-v-2012

| Name                                                                          | Description                                                                                          | Type  | Labels     |
|-------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------|-------|------------|
| `windows_hyperv_dynamic_memory_balancer_available_memory_bytes`               | This counter represents the amount of memory left on the node.                                       | gauge | `balancer` |
| `windows_hyperv_dynamic_memory_balancer_available_memory_for_balancing_bytes` | This counter represents the available memory for balancing purposes.                                 | gauge | `balancer` |
| `windows_hyperv_dynamic_memory_balancer_average_pressure_ratio`               | This counter represents the average system pressure on the balancer node among all balanced objects. | gauge | `balancer` |
| `windows_hyperv_dynamic_memory_balancer_system_current_pressure_ratio`        | This counter represents the current pressure in the system.                                          | gauge | `balancer` |


### Hyper-V Dynamic Memory VM


| Name                                                                   | Description                                                                                    | Type    | Labels |
|------------------------------------------------------------------------|------------------------------------------------------------------------------------------------|---------|--------|
| `windows_hyperv_dynamic_memory_vm_added_bytes_total`                   | This counter represents the cummulative amount of memory added to the VM.                      | counter | `vm`   |
| `windows_hyperv_dynamic_memory_vm_pressure_current_ratio`              | This counter represents the current pressure in the VM.                                        | gauge   | `vm`   |
| `windows_hyperv_dynamic_memory_vm_guest_available_bytes`               | This counter represents the current amount of available memory in the VM (reported by the VM). | gauge   | `vm`   |
| `windows_hyperv_dynamic_memory_vm_guest_visible_physical_memory_bytes` | This counter represents the amount of memory visible in the VM                                 | gauge   | `vm`   |
| `windows_hyperv_dynamic_memory_vm_pressure_maximum_ratio`              | This counter represents the maximum pressure band in the VM.                                   | gauge   | `vm`   |
| `windows_hyperv_dynamic_memory_vm_add_operations_total`                | This counter represents the total number of add operations for the VM.                         | counter | `vm`   |
| `windows_hyperv_dynamic_memory_vm_remove_operations_total`             | This counter represents the total number of remove operations for the VM.                      | counter | `vm`   |
| `windows_hyperv_dynamic_memory_vm_pressure_minimum_ratio`              | This counter represents the minimum pressure band in the VM.                                   | gauge   | `vm`   |
| `windows_hyperv_dynamic_memory_vm_physical`                            | This counter represents the current amount of memory in the VM.                                | gauge   | `vm`   |
| `windows_hyperv_dynamic_memory_vm_removed_bytes_total`                 | This counter represents the cummulative amount of memory removed from the VM.                  | counter | `vm`   |

### Hyper-V VM Vid Partition

| Name                                           | Description                                                             | Type  | Labels |
|------------------------------------------------|-------------------------------------------------------------------------|-------|--------|
| `windows_hyperv_vid_physical_pages_allocated`  | The number of physical pages allocated                                  | gauge | `vm`   |
| `windows_hyperv_vid_preferred_numa_node_index` | The preferred NUMA node index associated with this partition            | gauge | `vm`   |
| `windows_hyperv_vid_remote_physical_pages`     | The number of physical pages not allocated from the preferred NUMA node | gauge | `vm`   |


### HyperHyper-V Virtual Machine Health Summary

| Name                             | Description                                                                 | Type  | Labels |
|----------------------------------|-----------------------------------------------------------------------------|-------|--------|
| `windows_hyperv_health_critical` | This counter represents the number of virtual machines with critical health | gauge | None   |
| `windows_hyperv_health_ok`       | This counter represents the number of virtual machines with ok health       | gauge | None   |


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

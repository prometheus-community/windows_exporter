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
| `windows_hyperv_ethernet_bytes_dropped`                             | _Not yet documented_ | counter | `adapter`      |
| `windows_hyperv_ethernet_bytes_received`                            | _Not yet documented_ | counter | `adapter`      |
| `windows_hyperv_ethernet_bytes_sent`                                | _Not yet documented_ | counter | `adapter`      |
| `windows_hyperv_ethernet_frames_dropped`                            | _Not yet documented_ | counter | `adapter`      |
| `windows_hyperv_ethernet_frames_received`                           | _Not yet documented_ | counter | `adapter`      |
| `windows_hyperv_ethernet_frames_sent`                               | _Not yet documented_ | counter | `adapter`      |

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
| `windows_hyperv_dynamic_memory_vm_added_bytes_total`                   | This counter represents the cumulative amount of memory added to the VM.                       | counter | `vm`   |
| `windows_hyperv_dynamic_memory_vm_pressure_current_ratio`              | This counter represents the current pressure in the VM.                                        | gauge   | `vm`   |
| `windows_hyperv_dynamic_memory_vm_guest_available_bytes`               | This counter represents the current amount of available memory in the VM (reported by the VM). | gauge   | `vm`   |
| `windows_hyperv_dynamic_memory_vm_guest_visible_physical_memory_bytes` | This counter represents the amount of memory visible in the VM                                 | gauge   | `vm`   |
| `windows_hyperv_dynamic_memory_vm_pressure_maximum_ratio`              | This counter represents the maximum pressure band in the VM.                                   | gauge   | `vm`   |
| `windows_hyperv_dynamic_memory_vm_add_operations_total`                | This counter represents the total number of add operations for the VM.                         | counter | `vm`   |
| `windows_hyperv_dynamic_memory_vm_remove_operations_total`             | This counter represents the total number of remove operations for the VM.                      | counter | `vm`   |
| `windows_hyperv_dynamic_memory_vm_pressure_minimum_ratio`              | This counter represents the minimum pressure band in the VM.                                   | gauge   | `vm`   |
| `windows_hyperv_dynamic_memory_vm_physical`                            | This counter represents the current amount of memory in the VM.                                | gauge   | `vm`   |
| `windows_hyperv_dynamic_memory_vm_removed_bytes_total`                 | This counter represents the cumulative amount of memory removed from the VM.                   | counter | `vm`   |

### Hyper-V Hypervisor Logical Processor

| Name                                                                 | Description                                                            | Type    | Labels         |
|----------------------------------------------------------------------|------------------------------------------------------------------------|---------|----------------|
| `windows_hyperv_hypervisor_logical_processor_time_total`             | Time that processor spent in different modes (hypervisor, guest, idle) | counter | `core`.`state` |
| `windows_hyperv_hypervisor_logical_processor_context_switches_total` | The rate of virtual processor context switches on the processor.       | counter | `core`         |


### Hyper-V Hypervisor Root Partition

| Name                                                             | Description                                                                                                                                              | Type    | Labels |
|------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------|---------|--------|
| `windows_hyperv_root_partition_address_spaces`                   | The number of address spaces in the virtual TLB of the partition                                                                                         | gauge   | None   |
| `windows_hyperv_root_partition_attached_devices`                 | The number of devices attached to the partition                                                                                                          | gauge   | None   |
| `windows_hyperv_root_partition_deposited_pages`                  | The number of pages deposited into the partition                                                                                                         | gauge   | None   |
| `windows_hyperv_root_partition_device_dma_errors`                | An indicator of illegal DMA requests generated by all devices assigned to the partition                                                                  | gauge   | None   |
| `windows_hyperv_root_partition_device_interrupt_errors`          | An indicator of illegal interrupt requests generated by all devices assigned to the partition                                                            | gauge   | None   |
| `windows_hyperv_root_partition_device_interrupt_mappings`        | The number of device interrupt mappings used by the partition                                                                                            | gauge   | None   |
| `windows_hyperv_root_partition_device_interrupt_throttle_events` | The number of times an interrupt from a device assigned to the partition was temporarily throttled because the device was generating too many interrupts | gauge   | None   |
| `windows_hyperv_root_partition_preferred_numa_node_index`        | The number of pages present in the GPA space of the partition (zero for root partition)                                                                  | gauge   | None   |
| `windows_hyperv_root_partition_gpa_space_modifications`          | The rate of modifications to the GPA space of the partition                                                                                              | counter | None   |
| `windows_hyperv_root_partition_io_tlb_flush_cost`                | The average time (in nanoseconds) spent processing an I/O TLB flush                                                                                      | gauge   | None   |
| `windows_hyperv_root_partition_io_tlb_flush`                     | The rate of flushes of I/O TLBs of the partition                                                                                                         | counter | None   |
| `windows_hyperv_root_partition_recommended_virtual_tlb_size`     | The recommended number of pages to be deposited for the virtual TLB                                                                                      | gauge   | None   |
| `windows_hyperv_root_partition_physical_pages_allocated`         | The number of timer interrupts skipped for the partition                                                                                                 | gauge   | None   |
| `windows_hyperv_root_partition_1G_device_pages`                  | The number of 1G pages present in the device space of the partition                                                                                      | gauge   | None   |
| `windows_hyperv_root_partition_1G_gpa_pages`                     | The number of 1G pages present in the GPA space of the partition                                                                                         | gauge   | None   |
| `windows_hyperv_root_partition_2M_device_pages`                  | The number of 2M pages present in the device space of the partition                                                                                      | gauge   | None   |
| `windows_hyperv_root_partition_2M_gpa_pages`                     | The number of 2M pages present in the GPA space of the partition                                                                                         | gauge   | None   |
| `windows_hyperv_root_partition_4K_device_pages`                  | The number of 4K pages present in the device space of the partition                                                                                      | gauge   | None   |
| `windows_hyperv_root_partition_4K_gpa_pages`                     | The number of 4K pages present in the GPA space of the partition                                                                                         | gauge   | None   |
| `windows_hyperv_root_partition_virtual_tlb_flush_entries`        | The rate of flushes of the entire virtual TLB                                                                                                            | counter | None   |
| `windows_hyperv_root_partition_virtual_tlb_pages`                | The number of pages used by the virtual TLB of the partition                                                                                             | gauge   | None   |

### Hyper-V Virtual Network Adapter

| Name                                                                    | Description                                                                                                             | Type    | Labels    |
|-------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------|---------|-----------|
| `windows_hyperv_virtual_network_adapter_received_bytes_total`           | This counter represents the total number of bytes received per second by the network adapter                            | counter | `adapter` |
| `windows_hyperv_virtual_network_adapter_sent_bytes_total`               | This counter represents the total number of bytes sent per second by the network adapter                                | counter | `adapter` |
| `windows_hyperv_virtual_network_adapter_incoming_dropped_packets_total` | This counter represents the total number of dropped packets per second in the incoming direction of the network adapter | counter | `adapter` |
| `windows_hyperv_virtual_network_adapter_outgoing_dropped_packets_total` | This counter represents the total number of dropped packets per second in the outgoing direction of the network adapter | counter | `adapter` |
| `windows_hyperv_virtual_network_adapter_received_packets_total`         | This counter represents the total number of packets received per second by the network adapter                          | counter | `adapter` |
| `windows_hyperv_virtual_network_adapter_sent_packets_total`             | This counter represents the total number of packets sent per second by the network adapter                              | counter | `adapter` |

### Hyper-V Virtual Switch

| Name                                                                | Description                                                                                                              | Type    | Labels    |
|---------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------|---------|-----------|
| `windows_hyperv_vswitch_broadcast_packets_received_total`           | This represents the total number of broadcast packets received per second by the virtual switch                          | counter | `vswitch` |
| `windows_hyperv_vswitch_broadcast_packets_sent_total`               | This represents the total number of broadcast packets sent per second by the virtual switch                              | counter | `vswitch` |
| `windows_hyperv_vswitch_bytes_total`                                | This represents the total number of bytes per second traversing the virtual switch                                       | counter | `vswitch` |
| `windows_hyperv_vswitch_bytes_received_total`                       | This represents the total number of bytes received per second by the virtual switch                                      | counter | `vswitch` |
| `windows_hyperv_vswitch_bytes_sent_total`                           | This represents the total number of bytes sent per second by the virtual switch                                          | counter | `vswitch` |
| `windows_hyperv_vswitch_directed_packets_received_total`            | This represents the total number of directed packets received per second by the virtual switch                           | counter | `vswitch` |
| `windows_hyperv_vswitch_directed_packets_send_total`                | This represents the total number of directed packets sent per second by the virtual switch                               | counter | `vswitch` |
| `windows_hyperv_vswitch_dropped_packets_incoming_total`             | This represents the total number of packet dropped per second by the virtual switch in the incoming direction            | counter | `vswitch` |
| `windows_hyperv_vswitch_dropped_packets_outcoming_total`            | This represents the total number of packet dropped per second by the virtual switch in the outgoing direction            | counter | `vswitch` |
| `windows_hyperv_vswitch_extensions_dropped_packets_incoming_total`  | This represents the total number of packet dropped per second by the virtual switch extensions in the incoming direction | counter | `vswitch` |
| `windows_hyperv_vswitch_extensions_dropped_packets_outcoming_total` | This represents the total number of packet dropped per second by the virtual switch extensions in the outgoing direction | counter | `vswitch` |
| `windows_hyperv_vswitch_learned_mac_addresses_total`                | This counter represents the total number of learned MAC addresses of the virtual switch                                  | counter | `vswitch` |
| `windows_hyperv_vswitch_multicast_packets_received_total`           | This represents the total number of multicast packets received per second by the virtual switch                          | counter | `vswitch` |
| `windows_hyperv_vswitch_multicast_packets_sent_total`               | This represents the total number of multicast packets sent per second by the virtual switch                              | counter | `vswitch` |
| `windows_hyperv_vswitch_number_of_send_channel_moves_total`         | This represents the total number of send channel moves per second on this virtual switch                                 | counter | `vswitch` |
| `windows_hyperv_vswitch_number_of_vmq_moves_total`                  | This represents the total number of VMQ moves per second on this virtual switch                                          | counter | `vswitch` |
| `windows_hyperv_vswitch_packets_flooded_total`                      | This counter represents the total number of packets flooded by the virtual switch                                        | counter | `vswitch` |
| `windows_hyperv_vswitch_packets_total`                              | This represents the total number of packets per second traversing the virtual switch                                     | counter | `vswitch` |
| `windows_hyperv_vswitch_packets_received_total`                     | This represents the total number of packets received per second by the virtual switch                                    | counter | `vswitch` |
| `windows_hyperv_vswitch_packets_sent_total`                         | This represents the total number of packets send per second by the virtual switch                                        | counter | `vswitch` |
| `windows_hyperv_vswitch_purged_mac_addresses_total`                 | This counter represents the total number of purged MAC addresses of the virtual switch                                   | counter | `vswitch` |

### Hyper-V Virtual Storage Device

| Name                                                                | Description                                                                                                          | Type    | Labels   |
|---------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------|---------|----------|
| `windows_hyperv_virtual_storage_device_error_count_total`           | This counter represents the total number of errors that have occurred on this virtual device.                        | counter | `device` |
| `windows_hyperv_virtual_storage_device_queue_length`                | This counter represents the average queue length on this virtual device.                                             | gauge   | `device` |
| `windows_hyperv_virtual_storage_device_bytes_read`                  | This counter represents the total number of bytes that have been read on this virtual device.                        | counter | `device` |
| `windows_hyperv_virtual_storage_device_operations_read_total`       | This counter represents the total number of read operations that have occurred on this virtual device.                | counter | `device` |
| `windows_hyperv_virtual_storage_device_bytes_written`               | This counter represents the total number of bytes that have been written on this virtual device.                     | counter | `device` |
| `windows_hyperv_virtual_storage_device_operations_written_total`    | This counter represents the total number of write operations that have occurred on this virtual device.               | counter | `device` |
| `windows_hyperv_virtual_storage_device_latency_seconds`             | This counter represents the average IO transfer latency for this virtual device.                                     | gauge   | `device` |
| `windows_hyperv_virtual_storage_device_throughput`                  | This counter represents the average number of 8KB IO transfers completed by this virtual device.                     | gauge   | `device` |
| `windows_hyperv_virtual_storage_device_normalized_throughput`       | This counter represents the average number of IO transfers completed by this virtual device.                         | gauge   | `device` |
| `windows_hyperv_virtual_storage_device_lower_queue_length`          | This counter represents the average queue length on the underlying storage subsystem for this device.                | gauge   | `device` |
| `windows_hyperv_virtual_storage_device_lower_latency_seconds`       | This counter represents the average IO transfer latency on the underlying storage subsystem for this virtual device. | gauge   | `device` |
| `windows_hyperv_virtual_storage_device_io_quota_replenishment_rate` | This counter represents the IO quota replenishment rate for this virtual device.                                     | gauge   | `device` |

### Hyper-V VM Vid Partition

| Name                                           | Description                                                             | Type  | Labels |
|------------------------------------------------|-------------------------------------------------------------------------|-------|--------|
| `windows_hyperv_vid_physical_pages_allocated`  | The number of physical pages allocated                                  | gauge | `vm`   |
| `windows_hyperv_vid_preferred_numa_node_index` | The preferred NUMA node index associated with this partition            | gauge | `vm`   |
| `windows_hyperv_vid_remote_physical_pages`     | The number of physical pages not allocated from the preferred NUMA node | gauge | `vm`   |


### Hyper-V Virtual Machine Health Summary

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

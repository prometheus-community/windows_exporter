 # hyperv collector

The hyperv collector exposes metrics about the Hyper-V hypervisor

|                     |                      |
|---------------------|----------------------|
| Metric name prefix  | `hyperv`             |
| Source              | Performance counters |
| Enabled by default? | No                   |

## Flags

### `--collectors.hyperv.enabled`
Comma-separated list of collectors to use, for example:
`--collectors.hyperv.enabled=dynamic_memory_balancer,dynamic_memory_vm,hypervisor_logical_processor,hypervisor_root_partition,hypervisor_root_virtual_processor,hypervisor_virtual_processor,legacy_network_adapter,virtual_machine_health_summary,virtual_machine_vid_partition,virtual_network_adapter,virtual_storage_device,virtual_switch`.
Matching is case-sensitive.

## Metrics

### Hyper-V Datastore
### Hyper-V Datastore Metrics Documentation

This documentation outlines the available metrics for monitoring Hyper-V Datastore performance and resource usage using Prometheus. All metrics are prefixed with `windows_hyperv_datastore`.

| Metric Name                                                            | Description                                                                     | Type    | Labels    |
|------------------------------------------------------------------------|---------------------------------------------------------------------------------|---------|-----------|
| `windows_hyperv_datastore_fragmentation_ratio`                         | Represents the fragmentation ratio of the DataStore.                            | gauge   | datastore |
| `windows_hyperv_datastore_sector_size_bytes`                           | Represents the sector size of the DataStore in bytes.                           | gauge   | datastore |
| `windows_hyperv_datastore_data_alignment_bytes`                        | Represents the data alignment of the DataStore in bytes.                        | gauge   | datastore |
| `windows_hyperv_datastore_current_replay_log_size_bytes`               | Represents the current replay log size of the DataStore in bytes.               | gauge   | datastore |
| `windows_hyperv_datastore_available_entries`                           | Represents the number of available entries inside object tables.                | gauge   | datastore |
| `windows_hyperv_datastore_empty_entries`                               | Represents the number of empty entries inside object tables.                    | gauge   | datastore |
| `windows_hyperv_datastore_free_bytes`                                  | Represents the number of free bytes inside key tables.                          | gauge   | datastore |
| `windows_hyperv_datastore_data_end_bytes`                              | Represents the data end of the DataStore in bytes.                              | gauge   | datastore |
| `windows_hyperv_datastore_file_objects`                                | Represents the number of file objects in the DataStore.                         | gauge   | datastore |
| `windows_hyperv_datastore_object_tables`                               | Represents the number of object tables in the DataStore.                        | gauge   | datastore |
| `windows_hyperv_datastore_key_tables`                                  | Represents the number of key tables in the DataStore.                           | gauge   | datastore |
| `windows_hyperv_datastore_file_data_size_bytes`                        | Represents the file data size in bytes of the DataStore.                        | gauge   | datastore |
| `windows_hyperv_datastore_table_data_size_bytes`                       | Represents the table data size in bytes of the DataStore.                       | gauge   | datastore |
| `windows_hyperv_datastore_names_size_bytes`                            | Represents the names size in bytes of the DataStore.                            | gauge   | datastore |
| `windows_hyperv_datastore_number_of_keys`                              | Represents the number of keys in the DataStore.                                 | gauge   | datastore |
| `windows_hyperv_datastore_reconnect_latency_microseconds`              | Represents the reconnect latency in microseconds of the DataStore.              | gauge   | datastore |
| `windows_hyperv_datastore_disconnect_count`                            | Represents the disconnect count of the DataStore.                               | counter | datastore |
| `windows_hyperv_datastore_write_to_file_byte_latency_microseconds`     | Represents the write-to-file byte latency in microseconds of the DataStore.     | gauge   | datastore |
| `windows_hyperv_datastore_write_to_file_byte_count`                    | Represents the write-to-file byte count of the DataStore.                       | counter | datastore |
| `windows_hyperv_datastore_write_to_file_count`                         | Represents the write-to-file count of the DataStore.                            | counter | datastore |
| `windows_hyperv_datastore_read_from_file_byte_latency_microseconds`    | Represents the read-from-file byte latency in microseconds of the DataStore.    | gauge   | datastore |
| `windows_hyperv_datastore_read_from_file_byte_count`                   | Represents the read-from-file byte count of the DataStore.                      | counter | datastore |
| `windows_hyperv_datastore_read_from_file_count`                        | Represents the read-from-file count of the DataStore.                           | counter | datastore |
| `windows_hyperv_datastore_write_to_storage_byte_latency_microseconds`  | Represents the write-to-storage byte latency in microseconds of the DataStore.  | gauge   | datastore |
| `windows_hyperv_datastore_write_to_storage_byte_count`                 | Represents the write-to-storage byte count of the DataStore.                    | counter | datastore |
| `windows_hyperv_datastore_write_to_storage_count`                      | Represents the write-to-storage count of the DataStore.                         | counter | datastore |
| `windows_hyperv_datastore_read_from_storage_byte_latency_microseconds` | Represents the read-from-storage byte latency in microseconds of the DataStore. | gauge   | datastore |
| `windows_hyperv_datastore_read_from_storage_byte_count`                | Represents the read-from-storage byte count of the DataStore.                   | counter | datastore |
| `windows_hyperv_datastore_read_from_storage_count`                     | Represents the read-from-storage count of the DataStore.                        | counter | datastore |
| `windows_hyperv_datastore_commit_byte_latency_microseconds`            | Represents the commit byte latency in microseconds of the DataStore.            | gauge   | datastore |
| `windows_hyperv_datastore_commit_byte_count`                           | Represents the commit byte count of the DataStore.                              | counter | datastore |
| `windows_hyperv_datastore_commit_count`                                | Represents the commit count of the DataStore.                                   | counter | datastore |
| `windows_hyperv_datastore_cache_update_operation_latency_microseconds` | Represents the cache update operation latency in microseconds of the DataStore. | gauge   | datastore |
| `windows_hyperv_datastore_cache_update_operation_count`                | Represents the cache update operation count of the DataStore.                   | counter | datastore |
| `windows_hyperv_datastore_commit_operation_latency_microseconds`       | Represents the commit operation latency in microseconds of the DataStore.       | gauge   | datastore |
| `windows_hyperv_datastore_commit_operation_count`                      | Represents the commit operation count of the DataStore.                         | counter | datastore |
| `windows_hyperv_datastore_compact_operation_latency_microseconds`      | Represents the compact operation latency in microseconds of the DataStore.      | gauge   | datastore |
| `windows_hyperv_datastore_compact_operation_count`                     | Represents the compact operation count of the DataStore.                        | counter | datastore |
| `windows_hyperv_datastore_load_file_operation_latency_microseconds`    | Represents the load file operation latency in microseconds of the DataStore.    | gauge   | datastore |
| `windows_hyperv_datastore_load_file_operation_count`                   | Represents the load file operation count of the DataStore.                      | counter | datastore |
| `windows_hyperv_datastore_remove_operation_latency_microseconds`       | Represents the remove operation latency in microseconds of the DataStore.       | gauge   | datastore |
| `windows_hyperv_datastore_remove_operation_count`                      | Represents the remove operation count of the DataStore.                         | counter | datastore |
| `windows_hyperv_datastore_query_size_operation_latency_microseconds`   | Represents the query size operation latency in microseconds of the DataStore.   | gauge   | datastore |
| `windows_hyperv_datastore_query_size_operation_count`                  | Represents the query size operation count of the DataStore.                     | counter | datastore |
| `windows_hyperv_datastore_set_operation_latency_microseconds`          | Represents the set operation latency in microseconds of the DataStore.          | gauge   | datastore |
| `windows_hyperv_datastore_set_operation_count`                         | Represents the set operation count of the DataStore.                            | counter | datastore |

### Hyper-V Dynamic Memory Balancer

Some metrics explained: https://learn.microsoft.com/en-us/archive/blogs/chrisavis/monitoring-dynamic-memory-in-windows-server-hyper-v-2012

| Name                                                                          | Description                                                                             | Type  | Labels     |
|-------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------|-------|------------|
| `windows_hyperv_dynamic_memory_balancer_available_memory_bytes`               | Represents the amount of memory left on the node.                                       | gauge | `balancer` |
| `windows_hyperv_dynamic_memory_balancer_available_memory_for_balancing_bytes` | Represents the available memory for balancing purposes.                                 | gauge | `balancer` |
| `windows_hyperv_dynamic_memory_balancer_average_pressure_ratio`               | Represents the average system pressure on the balancer node among all balanced objects. | gauge | `balancer` |
| `windows_hyperv_dynamic_memory_balancer_system_current_pressure_ratio`        | Represents the current pressure in the system.                                          | gauge | `balancer` |


### Hyper-V Dynamic Memory VM

| Name                                                                   | Description                                                                       | Type    | Labels |
|------------------------------------------------------------------------|-----------------------------------------------------------------------------------|---------|--------|
| `windows_hyperv_dynamic_memory_vm_added_bytes_total`                   | Represents the cumulative amount of memory added to the VM.                       | counter | `vm`   |
| `windows_hyperv_dynamic_memory_vm_pressure_current_ratio`              | Represents the current pressure in the VM.                                        | gauge   | `vm`   |
| `windows_hyperv_dynamic_memory_vm_guest_available_bytes`               | Represents the current amount of available memory in the VM (reported by the VM). | gauge   | `vm`   |
| `windows_hyperv_dynamic_memory_vm_guest_visible_physical_memory_bytes` | Represents the amount of memory visible in the VM                                 | gauge   | `vm`   |
| `windows_hyperv_dynamic_memory_vm_pressure_maximum_ratio`              | Represents the maximum pressure band in the VM.                                   | gauge   | `vm`   |
| `windows_hyperv_dynamic_memory_vm_add_operations_total`                | Represents the total number of add operations for the VM.                         | counter | `vm`   |
| `windows_hyperv_dynamic_memory_vm_remove_operations_total`             | Represents the total number of remove operations for the VM.                      | counter | `vm`   |
| `windows_hyperv_dynamic_memory_vm_pressure_minimum_ratio`              | Represents the minimum pressure band in the VM.                                   | gauge   | `vm`   |
| `windows_hyperv_dynamic_memory_vm_physical`                            | Represents the current amount of memory in the VM.                                | gauge   | `vm`   |
| `windows_hyperv_dynamic_memory_vm_removed_bytes_total`                 | Represents the cumulative amount of memory removed from the VM.                   | counter | `vm`   |

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


### Hyper-V Hypervisor Root Virtual Processor

| Name                                                                      | Description                                                                                                       | Type    | Labels         |
|---------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------|---------|----------------|
| `windows_hyperv_hypervisor_root_virtual_processor_time_total`             | Time that processor spent in different modes (hypervisor, guest_run, guest_idle, remote, total)                   | counter | `core`.`state` |
| `windows_hyperv_hypervisor_root_virtual_cpu_wait_time_per_dispatch_total` | The average time (in nanoseconds) spent waiting for a virtual processor to be dispatched onto a logical processor | counter | `core`         |


### Hyper-V Legacy Network Adapter

| Name                                                          | Description                                                             | Type    | Labels    |
|---------------------------------------------------------------|-------------------------------------------------------------------------|---------|-----------|
| `windows_hyperv_legacy_network_adapter_bytes_dropped_total`   | Bytes Dropped is the number of bytes dropped on the network adapter     | counter | `adapter` |
| `windows_hyperv_legacy_network_adapter_bytes_received_total`  | Bytes received is the number of bytes received on the network adapter   | counter | `adapter` |
| `windows_hyperv_legacy_network_adapter_bytes_sent_total`      | Bytes sent is the number of bytes sent over the network adapter         | counter | `adapter` |
| `windows_hyperv_legacy_network_adapter_frames_dropped_total`  | Frames Dropped is the number of frames dropped on the network adapter   | counter | `adapter` |
| `windows_hyperv_legacy_network_adapter_frames_received_total` | Frames received is the number of frames received on the network adapter | counter | `adapter` |
| `windows_hyperv_legacy_network_adapter_frames_sent_total`     | Frames sent is the number of frames sent over the network adapter       | counter | `adapter` |


### Hyper-V Hypervisor Virtual Processor

| Name                                                                           | Description                                                                                                        | Type    | Labels       |
|--------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------|---------|--------------|
| `windows_hyperv_hypervisor_virtual_processor_time_total`                       | Time that processor spent in different modes (hypervisor, guest_run, guest_idle, remote)                           | counter | `vm`, `core` |
| `windows_hyperv_hypervisor_virtual_processor_total_run_time_total`             | Time that processor spent                                                                                          | counter | `vm`, `core` |
| `windows_hyperv_hypervisor_virtual_processor_cpu_wait_time_per_dispatch_total` | The average time (in nanoseconds) spent waiting for a virtual processor to be dispatched onto a logical processor. | counter | `vm`, `core` |

### Hyper-V Virtual Network Adapter

| Name                                                                    | Description                                                                                                | Type    | Labels    |
|-------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------|---------|-----------|
| `windows_hyperv_virtual_network_adapter_received_bytes_total`           | Represents the total number of bytes received per second by the network adapter                            | counter | `adapter` |
| `windows_hyperv_virtual_network_adapter_sent_bytes_total`               | Represents the total number of bytes sent per second by the network adapter                                | counter | `adapter` |
| `windows_hyperv_virtual_network_adapter_incoming_dropped_packets_total` | Represents the total number of dropped packets per second in the incoming direction of the network adapter | counter | `adapter` |
| `windows_hyperv_virtual_network_adapter_outgoing_dropped_packets_total` | Represents the total number of dropped packets per second in the outgoing direction of the network adapter | counter | `adapter` |
| `windows_hyperv_virtual_network_adapter_received_packets_total`         | Represents the total number of packets received per second by the network adapter                          | counter | `adapter` |
| `windows_hyperv_virtual_network_adapter_sent_packets_total`             | Represents the total number of packets sent per second by the network adapter                              | counter | `adapter` |

### Hyper-V Virtual Network Adapter Drop Reasons

| Name                                                  | Description                                  | Type    | Labels                         |
|-------------------------------------------------------|----------------------------------------------|---------|--------------------------------|
| `windows_hyperv_virtual_network_adapter_drop_reasons` | Hyper-V Virtual Network Adapter Drop Reasons | counter | `adapter`,`direction`,`reason` |

### Hyper-V Virtual SMB

| Name                                                  | Description                                                                       | Type    | Labels     |
|-------------------------------------------------------|-----------------------------------------------------------------------------------|---------|------------|
| `windows_hyperv_virtual_smb_direct_mapped_sections`   | Represents the number of direct-mapped sections in the virtual SMB`               | gauge   | `instance` |
| `windows_hyperv_virtual_smb_direct_mapped_pages`      | Represents the number of direct-mapped pages in the virtual SMB`                  | gauge   | `instance` |
| `windows_hyperv_virtual_smb_write_bytes_rdma`         | Represents the number of bytes written per second using RDMA in the virtual SMB`  | counter | `instance` |
| `windows_hyperv_virtual_smb_write_bytes`              | Represents the number of bytes written per second in the virtual SMB`             | counter | `instance` |
| `windows_hyperv_virtual_smb_read_bytes_rdma`          | Represents the number of bytes read per second using RDMA in the virtual SMB`     | counter | `instance` |
| `windows_hyperv_virtual_smb_read_bytes`               | Represents the number of bytes read per second in the virtual SMB`                | counter | `instance` |
| `windows_hyperv_virtual_smb_flush_requests`           | Represents the number of flush requests per second in the virtual SMB`            | counter | `instance` |
| `windows_hyperv_virtual_smb_write_requests_rdma`      | Represents the number of write requests per second using RDMA in the virtual SMB` | counter | `instance` |
| `windows_hyperv_virtual_smb_write_requests`           | Represents the number of write requests per second in the virtual SMB`            | counter | `instance` |
| `windows_hyperv_virtual_smb_read_requests_rdma`       | Represents the number of read requests per second using RDMA in the virtual SMB`  | counter | `instance` |
| `windows_hyperv_virtual_smb_read_requests`            | Represents the number of read requests per second in the virtual SMB`             | counter | `instance` |
| `windows_hyperv_virtual_smb_current_pending_requests` | Represents the current number of pending requests in the virtual SMB`             | gauge   | `instance` |
| `windows_hyperv_virtual_smb_current_open_file_count`  | Represents the current number of open files in the virtual SMB`                   | gauge   | `instance` |
| `windows_hyperv_virtual_smb_tree_connect_count`       | Represents the number of tree connects in the virtual SMB`                        | gauge   | `instance` |
| `windows_hyperv_virtual_smb_requests`                 | Represents the number of requests per second in the virtual SMB`                  | counter | `instance` |
| `windows_hyperv_virtual_smb_sent_bytes`               | Represents the number of bytes sent per second in the virtual SMB`                | counter | `instance` |
| `windows_hyperv_virtual_smb_received_bytes`           | Represents the number of bytes received per second in the virtual SMB`            | counter | `instance` |


### Hyper-V Virtual Switch

| Name                                                                | Description                                                                                                         | Type    | Labels    |
|---------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------|---------|-----------|
| `windows_hyperv_vswitch_broadcast_packets_received_total`           | Represents the total number of broadcast packets received per second by the virtual switch                          | counter | `vswitch` |
| `windows_hyperv_vswitch_broadcast_packets_sent_total`               | Represents the total number of broadcast packets sent per second by the virtual switch                              | counter | `vswitch` |
| `windows_hyperv_vswitch_bytes_total`                                | Represents the total number of bytes per second traversing the virtual switch                                       | counter | `vswitch` |
| `windows_hyperv_vswitch_bytes_received_total`                       | Represents the total number of bytes received per second by the virtual switch                                      | counter | `vswitch` |
| `windows_hyperv_vswitch_bytes_sent_total`                           | Represents the total number of bytes sent per second by the virtual switch                                          | counter | `vswitch` |
| `windows_hyperv_vswitch_directed_packets_received_total`            | Represents the total number of directed packets received per second by the virtual switch                           | counter | `vswitch` |
| `windows_hyperv_vswitch_directed_packets_send_total`                | Represents the total number of directed packets sent per second by the virtual switch                               | counter | `vswitch` |
| `windows_hyperv_vswitch_dropped_packets_incoming_total`             | Represents the total number of packet dropped per second by the virtual switch in the incoming direction            | counter | `vswitch` |
| `windows_hyperv_vswitch_dropped_packets_outcoming_total`            | Represents the total number of packet dropped per second by the virtual switch in the outgoing direction            | counter | `vswitch` |
| `windows_hyperv_vswitch_extensions_dropped_packets_incoming_total`  | Represents the total number of packet dropped per second by the virtual switch extensions in the incoming direction | counter | `vswitch` |
| `windows_hyperv_vswitch_extensions_dropped_packets_outcoming_total` | Represents the total number of packet dropped per second by the virtual switch extensions in the outgoing direction | counter | `vswitch` |
| `windows_hyperv_vswitch_learned_mac_addresses_total`                | Represents the total number of learned MAC addresses of the virtual switch                                          | counter | `vswitch` |
| `windows_hyperv_vswitch_multicast_packets_received_total`           | Represents the total number of multicast packets received per second by the virtual switch                          | counter | `vswitch` |
| `windows_hyperv_vswitch_multicast_packets_sent_total`               | Represents the total number of multicast packets sent per second by the virtual switch                              | counter | `vswitch` |
| `windows_hyperv_vswitch_number_of_send_channel_moves_total`         | Represents the total number of send channel moves per second on this virtual switch                                 | counter | `vswitch` |
| `windows_hyperv_vswitch_number_of_vmq_moves_total`                  | Represents the total number of VMQ moves per second on this virtual switch                                          | counter | `vswitch` |
| `windows_hyperv_vswitch_packets_flooded_total`                      | Represents the total number of packets flooded by the virtual switch                                                | counter | `vswitch` |
| `windows_hyperv_vswitch_packets_total`                              | Represents the total number of packets per second traversing the virtual switch                                     | counter | `vswitch` |
| `windows_hyperv_vswitch_packets_received_total`                     | Represents the total number of packets received per second by the virtual switch                                    | counter | `vswitch` |
| `windows_hyperv_vswitch_packets_sent_total`                         | Represents the total number of packets send per second by the virtual switch                                        | counter | `vswitch` |
| `windows_hyperv_vswitch_purged_mac_addresses_total`                 | Represents the total number of purged MAC addresses of the virtual switch                                           | counter | `vswitch` |

### Hyper-V Virtual Storage Device

| Name                                                                | Description                                                                                             | Type    | Labels   |
|---------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------|---------|----------|
| `windows_hyperv_virtual_storage_device_error_count_total`           | Represents the total number of errors that have occurred on this virtual device.                        | counter | `device` |
| `windows_hyperv_virtual_storage_device_queue_length`                | Represents the average queue length on this virtual device.                                             | gauge   | `device` |
| `windows_hyperv_virtual_storage_device_bytes_read`                  | Represents the total number of bytes that have been read on this virtual device.                        | counter | `device` |
| `windows_hyperv_virtual_storage_device_operations_read_total`       | Represents the total number of read operations that have occurred on this virtual device.               | counter | `device` |
| `windows_hyperv_virtual_storage_device_bytes_written`               | Represents the total number of bytes that have been written on this virtual device.                     | counter | `device` |
| `windows_hyperv_virtual_storage_device_operations_written_total`    | Represents the total number of write operations that have occurred on this virtual device.              | counter | `device` |
| `windows_hyperv_virtual_storage_device_latency_seconds`             | Represents the average IO transfer latency for this virtual device.                                     | gauge   | `device` |
| `windows_hyperv_virtual_storage_device_throughput`                  | Represents the average number of 8KB IO transfers completed by this virtual device.                     | gauge   | `device` |
| `windows_hyperv_virtual_storage_device_normalized_throughput`       | Represents the average number of IO transfers completed by this virtual device.                         | gauge   | `device` |
| `windows_hyperv_virtual_storage_device_lower_queue_length`          | Represents the average queue length on the underlying storage subsystem for this device.                | gauge   | `device` |
| `windows_hyperv_virtual_storage_device_lower_latency_seconds`       | Represents the average IO transfer latency on the underlying storage subsystem for this virtual device. | gauge   | `device` |
| `windows_hyperv_virtual_storage_device_io_quota_replenishment_rate` | Represents the IO quota replenishment rate for this virtual device.                                     | gauge   | `device` |

### Hyper-V VM Vid Partition

| Name                                           | Description                                                             | Type  | Labels |
|------------------------------------------------|-------------------------------------------------------------------------|-------|--------|
| `windows_hyperv_vid_physical_pages_allocated`  | The number of physical pages allocated                                  | gauge | `vm`   |
| `windows_hyperv_vid_preferred_numa_node_index` | The preferred NUMA node index associated with this partition            | gauge | `vm`   |
| `windows_hyperv_vid_remote_physical_pages`     | The number of physical pages not allocated from the preferred NUMA node | gauge | `vm`   |


### Hyper-V Virtual Machine Health Summary

| Name                                                 | Description                                           | Type  | Labels |
|------------------------------------------------------|-------------------------------------------------------|-------|--------|
| `windows_hyperv_virtual_machine_health_total_count` | Represents the number of virtual machines with health | gauge | None   |


### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
Percent of physical CPU resources used per VM (on instance "localhost")
```
(sum (rate(windows_hyperv_hypervisor_virtual_processor_time_total{state="hypervisor",instance="localhost"}[1m]))) / ignoring(state,vm) group_left max (windows_cpu_logical_processor{instance="localhost"}) / 100000
```
Percent of physical CPU resources used by all VMs (on all monitored hosts)
```
(sum by (instance)(rate(windows_hyperv_hypervisor_virtual_processor_total_run_time_total{}[1m]))) / max by (instance)(windows_cpu_logical_processor{}) / 100000
```
Percent of physical CPU resources by the hosts themselves (on all monitored hosts)
```
(sum by (instance)(rate(windows_hyperv_hypervisor_root_virtual_processor_total_run_time_total{state="total"}[1m]))) / sum by (instance)(windows_cpu_logical_processor{}) / 100000
```
Percent of physical CPU resources by the hypervisor (on all monitored hosts)
```
(sum by (instance)(rate(windows_hyperv_hypervisor_logical_processor_total_run_time_total{}[1m]))) / sum by (instance)(windows_cpu_logical_processor{}) / 100000
```

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

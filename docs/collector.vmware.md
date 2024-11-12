# vmware collector

The vmware collector exposes metrics about a VMware guest VM

|                     |                      |
|---------------------|----------------------|
| Metric name prefix  | `vmware`             |
| Source              | Performance counters |
| Enabled by default? | No                   |

## Flags

None

## Metrics

| Name                                        | Description                                                                                                                                                                                                                                                                                                                                 | Type    | Labels |
|---------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|--------|
| `windows_vmware_mem_active_bytes`           | The estimated amount of memory the virtual machine is actively using.                                                                                                                                                                                                                                                                       | gauge   | None   |
| `windows_vmware_mem_ballooned_bytes`        | The amount of memory that has been reclaimed from this virtual machine via the VMware Memory Balloon mechanism.                                                                                                                                                                                                                             | gauge   | None   |
| `windows_vmware_mem_limit_bytes`            | The maximum amount of memory that is allowed to the virtual machine. Assigning a Memory Limit ensures that this virtual machine never consumes more than a certain amount of the allowed memory. By limiting the amount of memory consumed, a portion of this shared resource is allowed to other virtual machines.                         | gauge   | None   |
| `windows_vmware_mem_mapped_bytes`           | The mapped memory size of this virtual machine. This is the current total amount of guest memory that is backed by physical memory. Note that this number may include pages of memory shared between multiple virtual machines and thus may be an overestimate of the amount of physical host memory consumed by this virtual machine.      | gauge   | None   |
| `windows_vmware_mem_overhead_bytes`         | The amount of overhead memory associated with this virtual machine consumed on the host system.                                                                                                                                                                                                                                             | gauge   | None   |
| `windows_vmware_mem_reservation_bytes`      | The minimum amount of memory that is guaranteed to the virtual machine. Assigning a Memory Reservation ensures that even as other virtual machines on the same host consume memory, there is still a certain minimum amount for this virtual machine.                                                                                       | gauge   | None   |
| `windows_vmware_mem_shared_bytes`           | The amount of physical memory associated with this virtual machine that is copy-on-write (COW) shared on the host.                                                                                                                                                                                                                          | gauge   | None   |
| `windows_vmware_mem_shared_saved_bytes`     | The estimated amount of physical memory on the host saved from copy-on-write (COW) shared guest physical memory.                                                                                                                                                                                                                            | gauge   | None   |
| `windows_vmware_mem_shares`                 | The number of memory shares allocated to the virtual machine.                                                                                                                                                                                                                                                                               | gauge   | None   |
| `windows_vmware_mem_swapped_bytes`          | The amount of memory associated with this virtual machine that has been swapped by ESX.                                                                                                                                                                                                                                                     | gauge   | None   |
| `windows_vmware_mem_target_size_bytes`      | Memory Target Size                                                                                                                                                                                                                                                                                                                          | gauge   | None   |
| `windows_vmware_mem_used_bytes`             | The estimated amount of physical host memory currently consumed for this virtual machine’s physical memory.                                                                                                                                                                                                                                 | gauge   | None   |
| `windows_vmware_cpu_limit_mhz`              | The maximum processing power in MHz allowed to the virtual machine. Assigning a CPU Limit ensures that this virtual machine never consumes more than a certain amount of the available processor power. By limiting the amount of processing power consumed, a portion of the processing power becomes available to other virtual machines. | gauge   | None   |
| `windows_vmware_cpu_reservation_mhz`        | The minimum processing power in MHz available to the virtual machine. Assigning a CPU Reservation ensures that even as other virtual machines on the same host consume shared processing power, there is still a certain minimum amount for this virtual machine.                                                                           | gauge   | None   |
| `windows_vmware_cpu_shares`                 | The number of CPU shares allocated to the virtual machine.                                                                                                                                                                                                                                                                                  | gauge   | None   |
| `windows_vmware_cpu_stolen_seconds_total`   | The time that the VM was runnable but not scheduled to run                                                                                                                                                                                                                                                                                  | counter | None   |
| `windows_vmware_cpu_time_seconds_total`     | Current load of the VM’s virtual processor                                                                                                                                                                                                                                                                                                  | counter | None   |
| `windows_vmware_cpu_effective_vm_speed_mhz` | The effective speed of the VM’s virtual CPU                                                                                                                                                                                                                                                                                                 | gauge   | None   |
| `windows_vmware_host_processor_speed_mhz`   | Host Processor speed                                                                                                                                                                                                                                                                                                                        | gauge   | None   |

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

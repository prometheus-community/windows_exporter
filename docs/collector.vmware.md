# vmware collector

The vmware collector exposes metrics about a VMware guest VM

|||
-|-
Metric name prefix  | `vmware`
Classes             | `Win32_PerfRawData_vmGuestLib_VMem`, `Win32_PerfRawData_vmGuestLib_VCPU`
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_vmware_mem_active_bytes` | _Not yet documented_ | gauge | None
`windows_vmware_mem_ballooned_bytes` | _Not yet documented_ | gauge | None
`windows_vmware_mem_limit_bytes` | _Not yet documented_ | gauge | None
`windows_vmware_mem_mapped_bytes` | _Not yet documented_ | gauge | None
`windows_vmware_mem_overhead_bytes` | _Not yet documented_ | gauge | None
`windows_vmware_mem_reservation_bytes` | _Not yet documented_ | gauge | None
`windows_vmware_mem_shared_bytes` | _Not yet documented_ | gauge | None
`windows_vmware_mem_shared_saved_bytes` | _Not yet documented_ | gauge | None
`windows_vmware_mem_shares` | _Not yet documented_ | gauge | None
`windows_vmware_mem_swapped_bytes` | _Not yet documented_ | gauge | None
`windows_vmware_mem_target_size_bytes` | _Not yet documented_ | gauge | None
`windows_vmware_mem_used_bytes` | _Not yet documented_ | gauge | None
`windows_vmware_cpu_limit_mhz` | _Not yet documented_ | gauge | None
`windows_vmware_cpu_reservation_mhz` | _Not yet documented_ | gauge | None
`windows_vmware_cpu_shares` | _Not yet documented_ | gauge | None
`windows_vmware_cpu_stolen_seconds_total` | _Not yet documented_ | counter | None
`windows_vmware_cpu_time_seconds_total` | _Not yet documented_ | counter | None
`windows_vmware_effective_vm_speed_mhz` | _Not yet documented_ | gauge | None
`windows_vmware_host_processor_speed_mhz` | _Not yet documented_ | gauge | None

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

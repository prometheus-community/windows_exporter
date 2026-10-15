# `hyperv` collector

The collector reads 16 `Hyper-V ...` PDH objects with `pdh.CounterTypeRaw`
from its sub-collectors, plus HCS, WMI and `virtdisk.dll` data. The objects
exist only where the Hyper-V role is installed; CI does not provision it.
Almost all objects are multi-instance; `Hyper-V Virtual Machine Health Summary`
is single-instance. These notes were recorded on Windows Server 2025 build
26100 with one running Generation 2 VM (`ubuntu`, 6 vCPUs, 2 GiB static
memory) on a host with 12 logical processors.

- `pdh.InstancesAll` drops the `_Total` instance, so the processor objects
  return `Hv LP 0`..`Hv LP 11`, `Root VP 0`..`Root VP 11` and
  `ubuntu:Hv VP 0`..`ubuntu:Hv VP 5`. The host logical processor count is the
  number of `Hv LP` instances (12 here).
- Hypervisor processor objects: every `%` counter and `CPU Wait Time Per
  Dispatch` is `PERF_PRECISION_100NS_TIMER`. The raw value is cumulative time
  in 100 ns ticks. On a logical processor `% Idle Time` + `% Total Run Time`
  equals the time since boot (453 s after a 452.7 s uptime), and
  `% Total Run Time` is about `% Guest Run Time` + `% Hypervisor Run Time`.
  The help text of `CPU Wait Time Per Dispatch` and `I/O TLB Flush Cost` says
  "average time (in nanoseconds)", but the raw value is cumulative 100 ns
  ticks.
- `Hyper-V Hypervisor Root Partition` is read for the instance `Root` only.
  All its non-`/sec` counters are `PERF_COUNTER_LARGE_RAWCOUNT`. Page counts
  and sizes are gauges; by the help text `Skipped Timer Ticks` and
  `Device Interrupt Throttle Events` count events.
- Dynamic Memory: memory counters are in MB (`Physical Memory` and
  `Guest Visible Physical Memory` are 2048 for the 2 GiB VM). `Added Memory`
  and `Removed Memory` are cumulative MB, `Memory Add/Remove Operations` are
  cumulative counts. Pressure counters are percentages and exceed 100
  (145 here). The VM instance exists even with Dynamic Memory disabled.
  Instances are VM names; VMs of the Host Compute Service (WSL 2) have an
  empty instance name.
- `Hyper-V DataStore`: instances are `<VM GUID> <file kind>` with the kinds
  `configuration file`, `runtime file`, `worker process configuration` and
  `worker process runtime state`. Every counter is
  `PERF_COUNTER_LARGE_RAWCOUNT` and the help text only repeats the name. The
  `... count` counters are cumulative. The `... latency microseconds`
  counters are not cumulative: `Commit operation latency microseconds` of the
  configuration file dropped from 6659 to 5213 while `Commit operation count`
  stayed 1.
- `Hyper-V VM Vid Partition`: pages are 4 KiB. The 2 GiB VM had 526336
  `Physical Pages Allocated`.
- `Hyper-V Virtual Storage Device`: instances are the disk path with `\`
  replaced by `-` and `(...)` by `[...]`, for example
  `C:-ProgramData-Microsoft-Windows-Virtual Hard Disks-ubuntu.vhdx`; a `\\?\`
  prefix becomes `--?-` (the `.vmgs` guest state file). A DVD drive shows up
  as `ISO`. Only running VMs have instances.
  - `Latency`, `Lower Latency`, `Queue Length` and `Lower Queue Length` all
    had the same raw value (56686): the sum of IO latencies in 100 ns ticks.
    The base of `Latency` (10) equals the raw `Throughput` and
    `Read Count` + `Write Count` (6 + 4).
  - `Throughput` (`PERF_COUNTER_COUNTER`) counts IO transfers.
    `Normalized Throughput` (`PERF_COUNTER_BULK_COUNT`) counts 8 KB IO
    transfers by its help text. Both raw values are cumulative.
  - `Read Count`, `Write Count` and `Error Count` are
    `PERF_COUNTER_RAWCOUNT` but cumulative ("total number of ...").
- `Hyper-V Virtual Switch` and `Hyper-V Virtual Network Adapter`: all
  `/sec` counters are `PERF_COUNTER_BULK_COUNT`, so the raw value is a
  cumulative count. `Bytes/sec` is `Bytes Received/sec` + `Bytes Sent/sec`
  and `Packets/sec` is `Packets Received/sec` + `Packets Sent/sec` in the
  samples. Switch instances are the switch names (`External`, `Internal`).
  VM adapter instances are `<VM>_<adapter>_<VM GUID>--<GUID>`. Host side
  instances look like `Internal_<GUID>`, `Ethernet_<GUID>`,
  `DrPort <GUID>_<GUID>` and
  `Microsoft Hyper-V Network Adapter__DEVICE_{<GUID>}`. The collector uses
  them unchanged as `adapter`.
- `Hyper-V Virtual Network Adapter Drop Reasons` has the same instances as
  `Hyper-V Virtual Network Adapter`. Every counter is a cumulative
  `PERF_COUNTER_LARGE_RAWCOUNT`; only `Outgoing NotReady` was non-zero (9).
- `Hyper-V Replica VM` exists but had no instance (no replication).
  By the help text `Replication Latency` and `Average Replication Latency`
  are in seconds, the size counters in bytes, and `Replication Count`,
  `Network Bytes Sent/Recv` and `Resynchronized Bytes` are cumulative since
  VMMS start.
- `Hyper-V Virtual SMB` and `Hyper-V Legacy Network Adapter` had no
  instances (no SMB-backed VM storage, no legacy adapter). In Virtual SMB,
  `Tree Connect Count` is `PERF_COUNTER_COUNTER` but its help text says
  "the current number of Tree Connects".
- `Hyper-V Virtual Machine Health Summary` counts VMs: `Health Ok` was 1 for
  the one VM.
- The `wmi_health`, `virtual_hard_disk` and host VM processor metrics do not
  come from PDH, so the table below does not cover them.

## Counter reference

Recorded on Windows Server 2025 (build 26100). `Raw` and `Base` are the first and second
values that `pdh.CounterTypeRaw` returns for the sample instance.

### `Hyper-V DataStore`

Multi-instance. Instances: `2F713ADB-4A60-4959-8986-949D88D9D6AC configuration file`, `2F713ADB-4A60-4959-8986-949D88D9D6AC runtime file`, `2F713ADB-4A60-4959-8986-949D88D9D6AC worker process configuration`, `2F713ADB-4A60-4959-8986-949D88D9D6AC worker process runtime state`. Sample instance: `2F713ADB-4A60-4959-8986-949D88D9D6AC configuration file`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Cache update operation count` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Cache update operation latency microseconds` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Commit byte count` | `PERF_COUNTER_LARGE_RAWCOUNT` | 12210 |  |
| `Commit byte latency microseconds` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Commit count` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Commit operation count` | `PERF_COUNTER_LARGE_RAWCOUNT` | 1 |  |
| `Commit operation latency microseconds` | `PERF_COUNTER_LARGE_RAWCOUNT` | 6659 |  |
| `Compact operation count` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Compact operation latency microseconds` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Current replay logSize` | `PERF_COUNTER_LARGE_RAWCOUNT` | 32768 |  |
| `Data alignment` | `PERF_COUNTER_LARGE_RAWCOUNT` | 4096 |  |
| `Data end` | `PERF_COUNTER_LARGE_RAWCOUNT` | 77824 |  |
| `Disconnect count` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `File data size in bytes` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Fragmentation ratio` | `PERF_COUNTER_LARGE_RAWCOUNT` | 12 |  |
| `Load file operation count` | `PERF_COUNTER_LARGE_RAWCOUNT` | 1 |  |
| `Load file operation latency microseconds` | `PERF_COUNTER_LARGE_RAWCOUNT` | 4350 |  |
| `Names size in bytes` | `PERF_COUNTER_LARGE_RAWCOUNT` | 4500 |  |
| `Number of available entries inside object tables` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Number of empty entries inside object tables` | `PERF_COUNTER_LARGE_RAWCOUNT` | 218 |  |
| `Number of file objects` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Number of free bytes inside key tables` | `PERF_COUNTER_LARGE_RAWCOUNT` | 6974 |  |
| `Number of key tables` | `PERF_COUNTER_LARGE_RAWCOUNT` | 8 |  |
| `Number of keys` | `PERF_COUNTER_LARGE_RAWCOUNT` | 361 |  |
| `Number of object tables` | `PERF_COUNTER_LARGE_RAWCOUNT` | 1 |  |
| `Query size operation count` | `PERF_COUNTER_LARGE_RAWCOUNT` | 208 |  |
| `Query size operation latency microseconds` | `PERF_COUNTER_LARGE_RAWCOUNT` | 3 |  |
| `Read from file byte count` | `PERF_COUNTER_LARGE_RAWCOUNT` | 51356 |  |
| `Read from file byte latency microseconds` | `PERF_COUNTER_LARGE_RAWCOUNT` | 9 |  |
| `Read from file count` | `PERF_COUNTER_LARGE_RAWCOUNT` | 236 |  |
| `Read from storage byte count` | `PERF_COUNTER_LARGE_RAWCOUNT` | 18470 |  |
| `Read from storage byte latency microseconds` | `PERF_COUNTER_LARGE_RAWCOUNT` | 1 |  |
| `Read from storage count` | `PERF_COUNTER_LARGE_RAWCOUNT` | 490 |  |
| `Reconnect latency microseconds` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Remove operation count` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Remove operation latency microseconds` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Sector size` | `PERF_COUNTER_LARGE_RAWCOUNT` | 512 |  |
| `Set operation count` | `PERF_COUNTER_LARGE_RAWCOUNT` | 5 |  |
| `Set operation latency microseconds` | `PERF_COUNTER_LARGE_RAWCOUNT` | 6 |  |
| `Table data size in bytes` | `PERF_COUNTER_LARGE_RAWCOUNT` | 21247 |  |
| `Write to file byte count` | `PERF_COUNTER_LARGE_RAWCOUNT` | 4234 |  |
| `Write to file byte latency microseconds` | `PERF_COUNTER_LARGE_RAWCOUNT` | 12 |  |
| `Write to file count` | `PERF_COUNTER_LARGE_RAWCOUNT` | 8 |  |
| `Write to storage byte count` | `PERF_COUNTER_LARGE_RAWCOUNT` | 58 |  |
| `Write to storage byte latency microseconds` | `PERF_COUNTER_LARGE_RAWCOUNT` | 6 |  |
| `Write to storage count` | `PERF_COUNTER_LARGE_RAWCOUNT` | 5 |  |

Not read by the collector:

- `PERF_COUNTER_LARGE_RAWCOUNT`: `File lock acquire latency microseconds`, `File lock count`, `File lock release latency microseconds`, `Get operation count`, `Get operation latency microseconds`, `Storage lock acquire latency microseconds`, `Storage lock count`, `Storage lock release latency microseconds`

### `Hyper-V Dynamic Memory Balancer`

Multi-instance. Instances: `System Balancer`. Sample instance: `System Balancer`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Available Memory` | `PERF_COUNTER_RAWCOUNT` | 4240 |  |
| `Available Memory For Balancing` | `PERF_COUNTER_RAWCOUNT` | 4240 |  |
| `Average Pressure` | `PERF_COUNTER_RAWCOUNT` | 145 |  |
| `System Current Pressure` | `PERF_COUNTER_RAWCOUNT` | 100 |  |

### `Hyper-V Dynamic Memory VM`

Multi-instance. Instances: `ubuntu`. Sample instance: `ubuntu`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Added Memory` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Average Pressure` | `PERF_COUNTER_RAWCOUNT` | 146 |  |
| `Current Pressure` | `PERF_COUNTER_RAWCOUNT` | 145 |  |
| `Guest Available Memory` | `PERF_COUNTER_RAWCOUNT` | 1025 |  |
| `Guest Visible Physical Memory` | `PERF_COUNTER_RAWCOUNT` | 2048 |  |
| `Maximum Pressure` | `PERF_COUNTER_RAWCOUNT` | 149 |  |
| `Memory Add Operations` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Memory Remove Operations` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Minimum Pressure` | `PERF_COUNTER_RAWCOUNT` | 143 |  |
| `Physical Memory` | `PERF_COUNTER_RAWCOUNT` | 2048 |  |
| `Removed Memory` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |

Not read by the collector:

- `PERF_COUNTER_RAWCOUNT`: `Smart Paging Working Set Size`

### `Hyper-V Hypervisor Logical Processor`

Multi-instance. Instances: `_Total`, `Hv LP 0`, `Hv LP 1`, `Hv LP 10`, `Hv LP 11`, `Hv LP 2`, `Hv LP 3`, `Hv LP 4`, ... (13 in total). Sample instance: `Hv LP 0`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `% Guest Run Time` | `PERF_PRECISION_100NS_TIMER` | 9270346767 |  |
| `% Hypervisor Run Time` | `PERF_PRECISION_100NS_TIMER` | 1438355934 |  |
| `% Idle Time` | `PERF_PRECISION_100NS_TIMER` | 23500462022 |  |
| `% Total Run Time` | `PERF_PRECISION_100NS_TIMER` | 10708658352 |  |
| `Context Switches/sec` | `PERF_COUNTER_BULK_COUNT` | 9299560 |  |

Not read by the collector:

- `PERF_COUNTER_BULK_COUNT`: `C1 Transitions/sec`, `C2 Transitions/sec`, `C3 Transitions/sec`, `CPPC Request Context Switches/sec`, `Hardware Interrupts/sec`, `Hypervisor Branch Predictor Flushes/sec`, `Hypervisor Immediate L1 Data Cache Flushes/sec`, `Hypervisor L1 Data Cache Flushes/sec`, `Hypervisor Microarchitectural Buffer Flushes/sec`, `Inter-Processor Interrupts Sent/sec`, `Inter-Processor Interrupts/sec`, `Performance Monitoring Interrupts/sec`, `Posted Interrupt Notifications/sec`, `Scheduler Interrupts/sec`, `Timer Interrupts/sec`, `Total Interrupts/sec`
- `PERF_COUNTER_LARGE_RAWCOUNT`: `Monitor Transition Cost`, `Reserve Group Id`, `Root Vp Index`, `Running Priority`, `Scheduler Local Run List Size`
- `PERF_PRECISION_100NS_TIMER`: `% C1 Time`, `% C2 Time`, `% C3 Time`

### `Hyper-V Hypervisor Root Partition`

Multi-instance. Instances: `_Total`, `Root`. Sample instance: `Root`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `1G device pages` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `1G GPA pages` | `PERF_COUNTER_LARGE_RAWCOUNT` | 11 |  |
| `2M device pages` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `2M GPA pages` | `PERF_COUNTER_LARGE_RAWCOUNT` | 2461 |  |
| `4K device pages` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `4K GPA pages` | `PERF_COUNTER_LARGE_RAWCOUNT` | 51200 |  |
| `Address Spaces` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Attached Devices` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Deposited Pages` | `PERF_COUNTER_LARGE_RAWCOUNT` | 29667 |  |
| `Device DMA Errors` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Device Interrupt Errors` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Device Interrupt Mappings` | `PERF_COUNTER_LARGE_RAWCOUNT` | 2 |  |
| `Device Interrupt Throttle Events` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `GPA Pages` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `GPA Space Modifications/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `I/O TLB Flush Cost` | `PERF_PRECISION_100NS_TIMER` | 0 |  |
| `I/O TLB Flushes/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Recommended Virtual TLB Size` | `PERF_COUNTER_LARGE_RAWCOUNT` | 64 |  |
| `Skipped Timer Ticks` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Virtual TLB Flush Entires/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Virtual TLB Pages` | `PERF_COUNTER_LARGE_RAWCOUNT` | 64 |  |

Not read by the collector:

- `PERF_COUNTER_BULK_COUNT`: `Nested TLB Trimmed Pages/sec`, `Pages Recombined/sec`, `Pages Shattered/sec`
- `PERF_COUNTER_LARGE_RAWCOUNT`: `Nested TLB Free List Size`, `Nested TLB Size`, `Recommended Nested TLB Size`, `Virtual Processors`
- `PERF_COUNTER_LARGE_RAWCOUNT_HEX`: `CPPC Requested Performance Level`

### `Hyper-V Hypervisor Root Virtual Processor`

Multi-instance. Instances: `_Total`, `Root VP 0`, `Root VP 1`, `Root VP 10`, `Root VP 11`, `Root VP 2`, `Root VP 3`, `Root VP 4`, ... (13 in total). Sample instance: `Root VP 0`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `% Guest Run Time` | `PERF_PRECISION_100NS_TIMER` | 9150616357 |  |
| `% Hypervisor Run Time` | `PERF_PRECISION_100NS_TIMER` | 1260894784 |  |
| `% Remote Run Time` | `PERF_PRECISION_100NS_TIMER` | 0 |  |
| `% Total Run Time` | `PERF_PRECISION_100NS_TIMER` | 10411471300 |  |
| `CPU Wait Time Per Dispatch` | `PERF_PRECISION_100NS_TIMER` | 290865977 |  |

Not read by the collector:

- `PERF_COUNTER_BULK_COUNT`: `Address Domain Flushes/sec`, `Address Space Evictions/sec`, `Address Space Flushes/sec`, `Address Space Switches/sec`, `APIC EOI Accesses/sec`, `APIC IPIs Sent/sec`, `APIC MMIO Accesses/sec`, `APIC Self IPIs Sent/sec`, `APIC TPR Accesses/sec`, `Bus Lock Acquisitions/sec`, `Control Register Accesses Forwarded/sec`, `Control Register Accesses/sec`, `CPPC Request Context Switches/sec`, `CPU Group Hypercalls/sec`, `CPUID Instructions Forwarded/sec`, `CPUID Instructions/sec`, `Debug Register Accesses Forwarded/sec`, `Debug Register Accesses/sec`, `Deposit Hypercalls/sec`, `Device Domain Hypercalls/sec`, `Emulated Instructions Forwarded/sec`, `Emulated Instructions/sec`, `Event Log Hypercalls/sec`, `Extended Hypercall Intercept Messages/sec`, `Extended Hypercalls/sec`, `External Interrupts Forwarded/sec`, `External Interrupts/sec`, `Flush Physical Address List Hypercalls/sec`, `Flush Physical Address Space Hypercalls/sec`, `GIF Instruction Emulation Intercepts/sec`, `Global GVA Range Flushes/sec`, `Global I/O TLB Flushes/sec`, `GPA Space Hypercalls/sec`, `Guest Page Table Maps/sec`, `Hardware Interrupts/sec`, `HLT Instructions Forwarded/sec`, `HLT Instructions/sec`, `Hypercalls Forwarded/sec`, `Hypercalls/sec`, `InvEpt All Context Emulation Intercepts/sec`, `InvEpt Single Context Emulation Intercepts/sec`, `InvVpid All Context Emulation Intercepts/sec`, `InvVpid Single Address Emulation Intercepts/sec`, `InvVpid Single Context Emulation Intercepts/sec`, `IO Instructions Forwarded/sec`, `IO Instructions/sec`, `IO Intercept Messages/sec`, `IOMMU Hypercalls/sec`, `Large Page TLB Fills/sec`, `Local Flushed GVA Ranges/sec`, `Local I/O TLB Flushes/sec`, `Logical Processor Dispatches/sec`, `Logical Processor Hypercalls/sec`, `Logical Processor Migrations/sec`, `Long Spin Wait Hypercalls/sec`, `MBEC Nested Page Table Switches/sec`, `Memory Intercept Messages/sec`, `MSR Accesses Forwarded/sec`, `MSR Accesses/sec`, `MWAIT Instructions Forwarded/sec`, `MWAIT Instructions/sec`, `Nested Page Fault Intercepts/sec`, `Nested SLAT Hard Page Faults/sec`, `Nested SLAT Soft Page Faults/sec`, `Nested TLB Page Table Evictions/sec`, `Nested TLB Page Table Reclamations/sec`, `Nested VM Entries/sec`, `Other Hypercalls/sec`, `Other Intercepts Forwarded/sec`, `Other Intercepts/sec`, `Other Messages/sec`, `Other Reflected Guest Exceptions/sec`, `Page Fault Intercepts Forwarded/sec`, `Page Fault Intercepts/sec`, `Page Invalidations Forwarded/sec`, `Page Invalidations/sec`, `Page Scans/sec`, `Page Table Allocations/sec`, `Page Table Evictions/sec`, `Page Table Reclamations/sec`, `Page Table Resets/sec`, `Page Table Validations/sec`, `Page Table Write Intercepts/sec`, `Pending Interrupts Forwarded/sec`, `Pending Interrupts/sec`, `Performance Monitoring Interrupts/sec`, `Performance Monitoring IPT MSR Accesses/sec`, `Performance Monitoring LBR MSR Accesses/sec`, `Performance Monitoring vPMU MSR Accesses/sec`, `Posted Interrupt Notifications/sec`, `Posted Interrupt Scans/sec`, `RDPMC Instructions/sec`, `Reflected Guest Page Faults/sec`, `Small Page TLB Fills/sec`, `SVM Hypercalls/sec`, `Synthetic Interrupt Hypercalls/sec`, `Synthetic Interrupts/sec`, `Total Intercepts/sec`, `Total Messages/sec`, `Total Virtualization Instructions Emulated/sec`, `Virtual Interrupt Hypercalls/sec`, `Virtual Interrupts/sec`, `Virtual MMU Hypercalls/sec`, `Virtual Processor Hypercalls/sec`, `VMCLEAR Emulation Intercepts/sec`, `VMLOAD Emulation Intercepts/sec`, `VMPTRLD Emulation Intercepts/sec`, `VMPTRST Emulation Intercepts/sec`, `VMREAD Emulation Intercepts/sec`, `VMSAVE Emulation Intercepts/sec`, `VMWRITE Emulation Intercepts/sec`, `VMXOFF Emulation Intercepts/sec`, `VMXON Emulation Intercepts/sec`, `VSM Hypercalls/sec`, `VTL1 Dispatches/sec`, `VTL2 Dispatches/sec`
- `PERF_COUNTER_LARGE_RAWCOUNT`: `Scheduling Priority`
- `PERF_PRECISION_100NS_TIMER`: `% Guest-Relative Utilization`, `% Total Core Run Time`, `% VTL1 Run Time`, `% VTL2 Run Time`, `Control Register Accesses Cost`, `Control Register Accesses Forwarding Cost`, `CPU Contention Time Per Dispatch`, `CPU Wake Up Time Per Dispatch`, `CPUID Instructions Cost`, `CPUID Instructions Forwarding Cost`, `Debug Register Accesses Cost`, `Debug Register Accesses Forwarding Cost`, `Emulated Instructions Cost`, `Emulated Instructions Forwarding Cost`, `External Interrupts Cost`, `GIF Instruction Emulation Cost`, `Global I/O TLB Flush Cost`, `HLT Instructions Cost`, `HLT Instructions Forwarding Cost`, `Hypercalls Cost`, `Hypercalls Forwarding Cost`, `InvEpt All Context Instruction Emulation Cost`, `InvEpt Single Context Instruction Emulation Cost`, `InvVpid All Context Instruction Emulation Cost`, `InvVpid Single Address Instruction Emulation Cost`, `InvVpid Single Context Instruction Emulation Cost`, `IO Instructions Cost`, `IO Instructions Forwarding Cost`, `Local I/O TLB Flush Cost`, `MSR Accesses Cost`, `MSR Accesses Forwarding Cost`, `MWAIT Instructions Cost`, `MWAIT Instructions Forwarding Cost`, `Nested Page Fault Intercepts Cost`, `Nested SLAT Hard Page Faults Cost`, `Nested SLAT Soft Page Faults Cost`, `Nested VM Entries Cost`, `Other Intercepts Cost`, `Other Intercepts Forwarding Cost`, `Page Fault Intercepts Cost`, `Page Fault Intercepts Forwarding Cost`, `Page Invalidations Cost`, `Page Invalidations Forwarding Cost`, `Pending Interrupts Cost`, `Pending Interrupts Forwarding Cost`, `RDPMC Instructions Cost`, `Total Intercepts Cost`, `Total Virtualization Instructions Emulation Cost`, `VMCLEAR Instruction Emulation Cost`, `VMLOAD Instruction Emulation Cost`, `VMPTRLD Instruction Emulation Cost`, `VMPTRST Instruction Emulation Cost`, `VMREAD Instruction Emulation Cost`, `VMSAVE Instruction Emulation Cost`, `VMWRITE Instruction Emulation Cost`, `VMXOFF Instruction Emulation Cost`, `VMXON Instruction Emulation Cost`, `VTL1 Average Run Time`, `VTL2 Average Run Time`

### `Hyper-V Hypervisor Virtual Processor`

Multi-instance. Instances: `_Total`, `ubuntu:Hv VP 0`, `ubuntu:Hv VP 1`, `ubuntu:Hv VP 2`, `ubuntu:Hv VP 3`, `ubuntu:Hv VP 4`, `ubuntu:Hv VP 5`. Sample instance: `ubuntu:Hv VP 0`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `% Guest Run Time` | `PERF_PRECISION_100NS_TIMER` | 2198066669 |  |
| `% Hypervisor Run Time` | `PERF_PRECISION_100NS_TIMER` | 982721208 |  |
| `% Remote Run Time` | `PERF_PRECISION_100NS_TIMER` | 0 |  |
| `% Total Run Time` | `PERF_PRECISION_100NS_TIMER` | 3180787877 |  |
| `CPU Wait Time Per Dispatch` | `PERF_PRECISION_100NS_TIMER` | 1217187193 |  |

Not read by the collector:

- `PERF_COUNTER_BULK_COUNT`: `Address Domain Flushes/sec`, `Address Space Evictions/sec`, `Address Space Flushes/sec`, `Address Space Switches/sec`, `APIC EOI Accesses/sec`, `APIC IPIs Sent/sec`, `APIC MMIO Accesses/sec`, `APIC Self IPIs Sent/sec`, `APIC TPR Accesses/sec`, `Bus Lock Acquisitions/sec`, `Control Register Accesses Forwarded/sec`, `Control Register Accesses/sec`, `CPPC Request Context Switches/sec`, `CPU Group Hypercalls/sec`, `CPUID Instructions Forwarded/sec`, `CPUID Instructions/sec`, `Debug Register Accesses Forwarded/sec`, `Debug Register Accesses/sec`, `Deposit Hypercalls/sec`, `Device Domain Hypercalls/sec`, `Emulated Instructions Forwarded/sec`, `Emulated Instructions/sec`, `Event Log Hypercalls/sec`, `Extended Hypercall Intercept Messages/sec`, `Extended Hypercalls/sec`, `External Interrupts Forwarded/sec`, `External Interrupts/sec`, `Flush Physical Address List Hypercalls/sec`, `Flush Physical Address Space Hypercalls/sec`, `GIF Instruction Emulation Intercepts/sec`, `Global GVA Range Flushes/sec`, `Global I/O TLB Flushes/sec`, `GPA Space Hypercalls/sec`, `Guest Page Table Maps/sec`, `Hardware Interrupts/sec`, `HLT Instructions Forwarded/sec`, `HLT Instructions/sec`, `Hypercalls Forwarded/sec`, `Hypercalls/sec`, `InvEpt All Context Emulation Intercepts/sec`, `InvEpt Single Context Emulation Intercepts/sec`, `InvVpid All Context Emulation Intercepts/sec`, `InvVpid Single Address Emulation Intercepts/sec`, `InvVpid Single Context Emulation Intercepts/sec`, `IO Instructions Forwarded/sec`, `IO Instructions/sec`, `IO Intercept Messages/sec`, `IOMMU Hypercalls/sec`, `Large Page TLB Fills/sec`, `Local Flushed GVA Ranges/sec`, `Local I/O TLB Flushes/sec`, `Logical Processor Dispatches/sec`, `Logical Processor Hypercalls/sec`, `Logical Processor Migrations/sec`, `Long Spin Wait Hypercalls/sec`, `MBEC Nested Page Table Switches/sec`, `Memory Intercept Messages/sec`, `MSR Accesses Forwarded/sec`, `MSR Accesses/sec`, `MWAIT Instructions Forwarded/sec`, `MWAIT Instructions/sec`, `Nested Page Fault Intercepts/sec`, `Nested SLAT Hard Page Faults/sec`, `Nested SLAT Soft Page Faults/sec`, `Nested TLB Page Table Evictions/sec`, `Nested TLB Page Table Reclamations/sec`, `Nested VM Entries/sec`, `Other Hypercalls/sec`, `Other Intercepts Forwarded/sec`, `Other Intercepts/sec`, `Other Messages/sec`, `Other Reflected Guest Exceptions/sec`, `Page Fault Intercepts Forwarded/sec`, `Page Fault Intercepts/sec`, `Page Invalidations Forwarded/sec`, `Page Invalidations/sec`, `Page Scans/sec`, `Page Table Allocations/sec`, `Page Table Evictions/sec`, `Page Table Reclamations/sec`, `Page Table Resets/sec`, `Page Table Validations/sec`, `Page Table Write Intercepts/sec`, `Pending Interrupts Forwarded/sec`, `Pending Interrupts/sec`, `Performance Monitoring Interrupts/sec`, `Performance Monitoring IPT MSR Accesses/sec`, `Performance Monitoring LBR MSR Accesses/sec`, `Performance Monitoring vPMU MSR Accesses/sec`, `Posted Interrupt Notifications/sec`, `Posted Interrupt Scans/sec`, `RDPMC Instructions/sec`, `Reflected Guest Page Faults/sec`, `Small Page TLB Fills/sec`, `SVM Hypercalls/sec`, `Synthetic Interrupt Hypercalls/sec`, `Synthetic Interrupts/sec`, `Total Intercepts/sec`, `Total Messages/sec`, `Total Virtualization Instructions Emulated/sec`, `Virtual Interrupt Hypercalls/sec`, `Virtual Interrupts/sec`, `Virtual MMU Hypercalls/sec`, `Virtual Processor Hypercalls/sec`, `VMCLEAR Emulation Intercepts/sec`, `VMLOAD Emulation Intercepts/sec`, `VMPTRLD Emulation Intercepts/sec`, `VMPTRST Emulation Intercepts/sec`, `VMREAD Emulation Intercepts/sec`, `VMSAVE Emulation Intercepts/sec`, `VMWRITE Emulation Intercepts/sec`, `VMXOFF Emulation Intercepts/sec`, `VMXON Emulation Intercepts/sec`, `VSM Hypercalls/sec`, `VTL1 Dispatches/sec`, `VTL2 Dispatches/sec`
- `PERF_COUNTER_LARGE_RAWCOUNT`: `Scheduling Priority`
- `PERF_PRECISION_100NS_TIMER`: `% Guest-Relative Utilization`, `% Total Core Run Time`, `% VTL1 Run Time`, `% VTL2 Run Time`, `Control Register Accesses Cost`, `Control Register Accesses Forwarding Cost`, `CPU Contention Time Per Dispatch`, `CPU Wake Up Time Per Dispatch`, `CPUID Instructions Cost`, `CPUID Instructions Forwarding Cost`, `Debug Register Accesses Cost`, `Debug Register Accesses Forwarding Cost`, `Emulated Instructions Cost`, `Emulated Instructions Forwarding Cost`, `External Interrupts Cost`, `GIF Instruction Emulation Cost`, `Global I/O TLB Flush Cost`, `HLT Instructions Cost`, `HLT Instructions Forwarding Cost`, `Hypercalls Cost`, `Hypercalls Forwarding Cost`, `InvEpt All Context Instruction Emulation Cost`, `InvEpt Single Context Instruction Emulation Cost`, `InvVpid All Context Instruction Emulation Cost`, `InvVpid Single Address Instruction Emulation Cost`, `InvVpid Single Context Instruction Emulation Cost`, `IO Instructions Cost`, `IO Instructions Forwarding Cost`, `Local I/O TLB Flush Cost`, `MSR Accesses Cost`, `MSR Accesses Forwarding Cost`, `MWAIT Instructions Cost`, `MWAIT Instructions Forwarding Cost`, `Nested Page Fault Intercepts Cost`, `Nested SLAT Hard Page Faults Cost`, `Nested SLAT Soft Page Faults Cost`, `Nested VM Entries Cost`, `Other Intercepts Cost`, `Other Intercepts Forwarding Cost`, `Page Fault Intercepts Cost`, `Page Fault Intercepts Forwarding Cost`, `Page Invalidations Cost`, `Page Invalidations Forwarding Cost`, `Pending Interrupts Cost`, `Pending Interrupts Forwarding Cost`, `RDPMC Instructions Cost`, `Total Intercepts Cost`, `Total Virtualization Instructions Emulation Cost`, `VMCLEAR Instruction Emulation Cost`, `VMLOAD Instruction Emulation Cost`, `VMPTRLD Instruction Emulation Cost`, `VMPTRST Instruction Emulation Cost`, `VMREAD Instruction Emulation Cost`, `VMSAVE Instruction Emulation Cost`, `VMWRITE Instruction Emulation Cost`, `VMXOFF Instruction Emulation Cost`, `VMXON Instruction Emulation Cost`, `VTL1 Average Run Time`, `VTL2 Average Run Time`

### `Hyper-V Legacy Network Adapter`

Multi-instance. No instances on the recording host, so the table has types only.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Bytes Dropped` | `PERF_COUNTER_LARGE_RAWCOUNT` |  |  |
| `Bytes Received/sec` | `PERF_COUNTER_BULK_COUNT` |  |  |
| `Bytes Sent/sec` | `PERF_COUNTER_BULK_COUNT` |  |  |
| `Frames Dropped` | `PERF_COUNTER_LARGE_RAWCOUNT` |  |  |
| `Frames Received/sec` | `PERF_COUNTER_BULK_COUNT` |  |  |
| `Frames Sent/sec` | `PERF_COUNTER_BULK_COUNT` |  |  |

### `Hyper-V Replica VM`

Multi-instance. No instances on the recording host, so the table has types only.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Average Replication Latency` | `PERF_COUNTER_LARGE_RAWCOUNT` |  |  |
| `Average Replication Size` | `PERF_COUNTER_LARGE_RAWCOUNT` |  |  |
| `Compression Efficiency` | `PERF_COUNTER_LARGE_RAWCOUNT` |  |  |
| `Last Replication Size` | `PERF_COUNTER_LARGE_RAWCOUNT` |  |  |
| `Network Bytes Recv` | `PERF_COUNTER_LARGE_RAWCOUNT` |  |  |
| `Network Bytes Sent` | `PERF_COUNTER_LARGE_RAWCOUNT` |  |  |
| `Replication Count` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `Replication Latency` | `PERF_COUNTER_LARGE_RAWCOUNT` |  |  |
| `Resynchronized Bytes` | `PERF_COUNTER_LARGE_RAWCOUNT` |  |  |

### `Hyper-V Virtual Machine Health Summary`

Single-instance.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Health Critical` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Health Ok` | `PERF_COUNTER_RAWCOUNT` | 1 |  |

### `Hyper-V Virtual Network Adapter`

Multi-instance. Instances: `DrPort 700EF58F-56E7-4982-9DC4-4BCD5FC2C56F_B3548C76-D68D-4199-B45F-5C9DE21BD03C`, `Ethernet_700EF58F-56E7-4982-9DC4-4BCD5FC2C56E`, `Internal_FDECB3D5-AD64-4B6A-B96C-6AFD1616F33A`, `Microsoft Hyper-V Network Adapter__DEVICE_{9FE7CB58-7B03-4A3F-814F-9CD52BF15705}`, `ubuntu_Network Adapter_2F713ADB-4A60-4959-8986-949D88D9D6AC--F28DA7DA-095C-4544-B094-97085584E17D`. Sample instance: `DrPort 700EF58F-56E7-4982-9DC4-4BCD5FC2C56F_B3548C76-D68D-4199-B45F-5C9DE21BD03C`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Broadcast Packets Received/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Broadcast Packets Sent/sec` | `PERF_COUNTER_BULK_COUNT` | 18 |  |
| `Bytes Received/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Bytes Sent/sec` | `PERF_COUNTER_BULK_COUNT` | 1708 |  |
| `Bytes/sec` | `PERF_COUNTER_BULK_COUNT` | 1708 |  |
| `Directed Packets Received/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Directed Packets Sent/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Dropped Packets Incoming/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Dropped Packets Outgoing/sec` | `PERF_COUNTER_BULK_COUNT` | 21 |  |
| `Extensions Dropped Packets Incoming/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Extensions Dropped Packets Outgoing/sec` | `PERF_COUNTER_BULK_COUNT` | 30 |  |
| `Multicast Packets Received/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Multicast Packets Sent/sec` | `PERF_COUNTER_BULK_COUNT` | 12 |  |
| `Packets Received/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Packets Sent/sec` | `PERF_COUNTER_BULK_COUNT` | 30 |  |
| `Packets/sec` | `PERF_COUNTER_BULK_COUNT` | 30 |  |

Not read by the collector:

- `PERF_COUNTER_BULK_COUNT`: `IPsec offload Bytes Receive/sec`, `IPsec offload Bytes Sent/sec`, `Packets Failed Software IP Rx CSO/sec`, `Packets Failed Software Rx CSO Parsing/sec`, `Packets Failed Software TCP Rx CSO/sec`, `Packets Failed Software UDP Rx CSO/sec`, `Packets Passed Software IP Rx CSO/sec`, `Packets Passed Software TCP Rx CSO/sec`, `Packets Passed Software UDP Rx CSO/sec`, `Packets With Software IP Tx CSO/sec`, `Packets With Software TCP Tx CSO/sec`, `Packets With Software UDP Tx CSO/sec`
- `PERF_COUNTER_LARGE_RAWCOUNT`: `Packets Failed Software IP Rx CSO`, `Packets Failed Software TCP Rx CSO`, `Packets Failed Software UDP Rx CSO`, `RSC Coalesce Events`, `RSC Coalesced Bytes`, `RSC Coalesced Event Bucket 1 (0To1)`, `RSC Coalesced Event Bucket 2 (2To3)`, `RSC Coalesced Event Bucket 3 (4To7)`, `RSC Coalesced Event Bucket 4 (8To15)`, `RSC Coalesced Event Bucket 5 (16To31)`, `RSC Coalesced Event Bucket 6 (32To63)`, `RSC Coalesced Packet Bucket 1 (0To1)`, `RSC Coalesced Packet Bucket 2 (2To3)`, `RSC Coalesced Packet Bucket 3 (4To7)`, `RSC Coalesced Packet Bucket 4 (8To15)`, `RSC Coalesced Packet Bucket 5 (16To31)`, `RSC Coalesced Packet Bucket 6 (32To63)`, `RSC Coalesced Packets`, `RSC Packets Processed`

### `Hyper-V Virtual Network Adapter Drop Reasons`

Multi-instance. Instances: `DrPort 700EF58F-56E7-4982-9DC4-4BCD5FC2C56F_B3548C76-D68D-4199-B45F-5C9DE21BD03C`, `Ethernet_700EF58F-56E7-4982-9DC4-4BCD5FC2C56E`, `Internal_FDECB3D5-AD64-4B6A-B96C-6AFD1616F33A`, `Microsoft Hyper-V Network Adapter__DEVICE_{9FE7CB58-7B03-4A3F-814F-9CD52BF15705}`, `ubuntu_Network Adapter_2F713ADB-4A60-4959-8986-949D88D9D6AC--F28DA7DA-095C-4544-B094-97085584E17D`. Sample instance: `DrPort 700EF58F-56E7-4982-9DC4-4BCD5FC2C56F_B3548C76-D68D-4199-B45F-5C9DE21BD03C`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Incoming BridgeReserved` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Incoming Busy` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Incoming DhcpGuard` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Incoming Disconnected` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Incoming FailedPvlanSetting` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Incoming FailedSecurityPolicy` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Incoming Filtered` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Incoming FilteredVLAN` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Incoming InvalidConfig` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Incoming InvalidData` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Incoming InvalidPacket` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Incoming Ipsec` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Incoming MacSpoofing` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Incoming MTUMismatch` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Incoming NativeFwdingReq` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Incoming NotAccepted` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Incoming NotReady` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Incoming Qos` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Incoming RequiredExtensionMissing` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Incoming Resources` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Incoming RouterGuard` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Incoming UnauthorizedMAC` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Incoming UnauthorizedVLAN` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Incoming Unknown` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Incoming VirtualSubnetId` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Outgoing BridgeReserved` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Outgoing Busy` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Outgoing DhcpGuard` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Outgoing Disconnected` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Outgoing FailedPvlanSetting` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Outgoing FailedSecurityPolicy` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Outgoing Filtered` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Outgoing FilteredVLAN` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Outgoing InvalidConfig` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Outgoing InvalidData` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Outgoing InvalidPacket` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Outgoing Ipsec` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Outgoing MacSpoofing` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Outgoing MTUMismatch` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Outgoing NativeFwdingReq` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Outgoing NotAccepted` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Outgoing NotReady` | `PERF_COUNTER_LARGE_RAWCOUNT` | 9 |  |
| `Outgoing Qos` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Outgoing RequiredExtensionMissing` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Outgoing Resources` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Outgoing RouterGuard` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Outgoing UnauthorizedMAC` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Outgoing UnauthorizedVLAN` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Outgoing Unknown` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Outgoing VirtualSubnetId` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |

Not read by the collector:

- `PERF_COUNTER_LARGE_RAWCOUNT`: `Incoming FailedDestinationListUpdate`, `Incoming FailedPacketFilter`, `Incoming FilteredIsolationUntagged`, `Incoming InjectedIcmp`, `Incoming InvalidDestMac`, `Incoming InvalidFirstNBTooSmall`, `Incoming InvalidPDQueue`, `Incoming InvalidSourceMac`, `Incoming InvalidVlanFormat`, `Incoming LowPowerPacketFilter`, `Incoming NicDisabled`, `Incoming StormLimit`, `Incoming SwitchDataFlowDisabled`, `Outgoing FailedDestinationListUpdate`, `Outgoing FailedPacketFilter`, `Outgoing FilteredIsolationUntagged`, `Outgoing InjectedIcmp`, `Outgoing InvalidDestMac`, `Outgoing InvalidFirstNBTooSmall`, `Outgoing InvalidPDQueue`, `Outgoing InvalidSourceMac`, `Outgoing InvalidVlanFormat`, `Outgoing LowPowerPacketFilter`, `Outgoing NicDisabled`, `Outgoing StormLimit`, `Outgoing SwitchDataFlowDisabled`

### `Hyper-V Virtual SMB`

Multi-instance. No instances on the recording host, so the table has types only.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Current Open File Count` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `Current Pending Requests` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `Direct-Mapped Pages` | `PERF_COUNTER_LARGE_RAWCOUNT` |  |  |
| `Direct-Mapped Sections` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `Flush Requests/sec` | `PERF_COUNTER_COUNTER` |  |  |
| `Read Bytes/sec` | `PERF_COUNTER_BULK_COUNT` |  |  |
| `Read Bytes/sec (RDMA)` | `PERF_COUNTER_BULK_COUNT` |  |  |
| `Read Requests/sec` | `PERF_COUNTER_COUNTER` |  |  |
| `Read Requests/sec (RDMA)` | `PERF_COUNTER_COUNTER` |  |  |
| `Received Bytes/sec` | `PERF_COUNTER_BULK_COUNT` |  |  |
| `Requests/sec` | `PERF_COUNTER_COUNTER` |  |  |
| `Sent Bytes/sec` | `PERF_COUNTER_BULK_COUNT` |  |  |
| `Tree Connect Count` | `PERF_COUNTER_COUNTER` |  |  |
| `Write Bytes/sec` | `PERF_COUNTER_BULK_COUNT` |  |  |
| `Write Bytes/sec (RDMA)` | `PERF_COUNTER_BULK_COUNT` |  |  |
| `Write Requests/sec` | `PERF_COUNTER_COUNTER` |  |  |
| `Write Requests/sec (RDMA)` | `PERF_COUNTER_COUNTER` |  |  |

Not read by the collector:

- `PERF_AVERAGE_TIMER`: `Avg. sec/Request`

### `Hyper-V Virtual Storage Device`

Multi-instance. Instances: `--?-C:-ProgramData-Microsoft-Windows-Hyper-V-Virtual Machines-2F713ADB-4A60-4959-8986-949D88D9D6AC.vmgs`, `C:-ProgramData-Microsoft-Windows-Virtual Hard Disks-ubuntu.vhdx`, `ISO`. Sample instance: `--?-C:-ProgramData-Microsoft-Windows-Hyper-V-Virtual Machines-2F713ADB-4A60-4959-8986-949D88D9D6AC.vmgs`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Error Count` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Io Quota Replenishment Rate` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Latency` | `PERF_AVERAGE_TIMER` | 56686 | 10 |
| `Lower Latency` | `PERF_AVERAGE_TIMER` | 56686 | 10 |
| `Lower Queue Length` | `PERF_COUNTER_LARGE_QUEUELEN_TYPE` | 56686 |  |
| `Normalized Throughput` | `PERF_COUNTER_BULK_COUNT` | 95 |  |
| `Queue Length` | `PERF_COUNTER_LARGE_QUEUELEN_TYPE` | 56686 |  |
| `Read Bytes/sec` | `PERF_COUNTER_BULK_COUNT` | 252416 |  |
| `Read Count` | `PERF_COUNTER_RAWCOUNT` | 6 |  |
| `Throughput` | `PERF_COUNTER_COUNTER` | 10 |  |
| `Write Bytes/sec` | `PERF_COUNTER_BULK_COUNT` | 468992 |  |
| `Write Count` | `PERF_COUNTER_RAWCOUNT` | 4 |  |

Not read by the collector:

- `PERF_COUNTER_BULK_COUNT`: `Byte Quota Replenishment Rate`
- `PERF_COUNTER_COUNTER`: `Read Operations/Sec`, `Write Operations/Sec`
- `PERF_COUNTER_LARGE_RAWCOUNT`: `Maximum Bandwidth`, `Maximum IO Rate`, `Minimum IO Rate`
- `PERF_COUNTER_RAWCOUNT`: `Adapter Open Channel Count`, `Flush Count`, `Maximum Adapter Worker Count`

### `Hyper-V Virtual Switch`

Multi-instance. Instances: `External`, `Internal`. Sample instance: `External`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Broadcast Packets Received/sec` | `PERF_COUNTER_BULK_COUNT` | 208 |  |
| `Broadcast Packets Sent/sec` | `PERF_COUNTER_BULK_COUNT` | 190 |  |
| `Bytes Received/sec` | `PERF_COUNTER_BULK_COUNT` | 6209983540 |  |
| `Bytes Sent/sec` | `PERF_COUNTER_BULK_COUNT` | 6209976573 |  |
| `Bytes/sec` | `PERF_COUNTER_BULK_COUNT` | 12419960113 |  |
| `Directed Packets Received/sec` | `PERF_COUNTER_BULK_COUNT` | 1035024 |  |
| `Directed Packets Sent/sec` | `PERF_COUNTER_BULK_COUNT` | 1035024 |  |
| `Dropped Packets Incoming/sec` | `PERF_COUNTER_BULK_COUNT` | 23 |  |
| `Dropped Packets Outgoing/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Extensions Dropped Packets Incoming/sec` | `PERF_COUNTER_BULK_COUNT` | 96 |  |
| `Extensions Dropped Packets Outgoing/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Learned Mac Addresses` | `PERF_COUNTER_LARGE_RAWCOUNT` | 3 |  |
| `Multicast Packets Received/sec` | `PERF_COUNTER_BULK_COUNT` | 438 |  |
| `Multicast Packets Sent/sec` | `PERF_COUNTER_BULK_COUNT` | 360 |  |
| `Number of Send Channel Moves/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Number of VMQ Moves/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Packets Flooded` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Packets Received/sec` | `PERF_COUNTER_BULK_COUNT` | 1035670 |  |
| `Packets Sent/sec` | `PERF_COUNTER_BULK_COUNT` | 1035574 |  |
| `Packets/sec` | `PERF_COUNTER_BULK_COUNT` | 2071244 |  |
| `Purged Mac Addresses` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |

Not read by the collector:

- `PERF_COUNTER_BULK_COUNT`: `Learned Mac Addresses/sec`, `Packets Flooded/sec`, `Purged Mac Addresses/sec`
- `PERF_COUNTER_LARGE_RAWCOUNT`: `RSC Coalesce Events`, `RSC Coalesced Bytes`, `RSC Coalesced Event Bucket 1 (0To1)`, `RSC Coalesced Event Bucket 2 (2To3)`, `RSC Coalesced Event Bucket 3 (4To7)`, `RSC Coalesced Event Bucket 4 (8To15)`, `RSC Coalesced Event Bucket 5 (16To31)`, `RSC Coalesced Event Bucket 6 (32To63)`, `RSC Coalesced Packet Bucket 1 (0To1)`, `RSC Coalesced Packet Bucket 2 (2To3)`, `RSC Coalesced Packet Bucket 3 (4To7)`, `RSC Coalesced Packet Bucket 4 (8To15)`, `RSC Coalesced Packet Bucket 5 (16To31)`, `RSC Coalesced Packet Bucket 6 (32To63)`, `RSC Coalesced Packets`, `RSC Packets Processed`

### `Hyper-V VM Vid Partition`

Multi-instance. Instances: `_Total`, `ubuntu`. Sample instance: `ubuntu`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Physical Pages Allocated` | `PERF_COUNTER_LARGE_RAWCOUNT` | 526336 |  |
| `Preferred NUMA Node Index` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Remote Physical Pages` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |

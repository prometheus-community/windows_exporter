---
title: WindowsCPUSaturation
---

## Meaning

The number of threads waiting for a processor, averaged over the rate window
and divided by the number of logical processors, exceeds the configured
threshold for 15 minutes. The default is two waiting threads per logical
processor. The metric is `windows:cpu_queue_per_core:ratio`, built from
`windows_system_processor_queue_length` and `windows:cpu_cores:count`.

## Impact

Runnable threads wait for CPU time, so applications respond slowly even when
their own CPU usage looks moderate. A long queue with CPU usage below 100%
points at a few busy cores, a single-threaded bottleneck, or processor
scheduling problems such as parked cores or hypervisor contention.

## Diagnosis

1. Inspect `windows:cpu_queue_per_core:ratio` next to `windows:cpu_usage:ratio`
   for the host. Saturation with high usage is a capacity problem; saturation
   with low usage points at scheduling.
2. Compare the per-core rates of `windows_cpu_time_seconds_total`. One or two
   cores at 100% while others idle indicate a thread that cannot spread.
3. Check `windows_cpu_parking_status`, `windows_cpu_core_frequency_hertz` and
   the power plan. On virtual machines check the hypervisor for CPU ready time
   or steal.
4. Use Task Manager or Performance Monitor (`System\Processor Queue Length`,
   `Process\Thread Count`) to find the processes with many runnable threads.

## Mitigation

Reduce or redistribute the workload, add logical processors, or fix the power
plan and parking settings. Give the virtual machine more CPU share or move it to
a less loaded host. Tune the threshold for hosts whose workload keeps a long
queue by design, such as build servers. Confirm the queue length per core and
application latency return to normal.

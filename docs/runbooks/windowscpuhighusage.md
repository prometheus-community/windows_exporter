---
title: WindowsCPUHighUsage
---

## Meaning

Average non-idle CPU usage across logical processors exceeds the configured
threshold for 15 minutes. The default is 90%, measured over a five-minute rate
window. This measures CPU time; Task Manager's frequency-adjusted utility can
show a different value.

## Impact

The host has little CPU capacity for additional work. Applications may show
increased latency, timeouts, or reduced throughput.

## Diagnosis

1. Inspect `windows:cpu_usage:ratio` for the affected host and compare it with
   its workload and normal baseline.
2. Inspect `windows:cpu_time:ratio_rate` by `mode`. High user time points toward
   application work; high privileged, interrupt, or DPC time can indicate kernel
   or driver activity. DPC and interrupt modes overlap privileged time; do not
   sum all modes to estimate usage.
3. Use Task Manager, Resource Monitor, or Performance Monitor to identify the
   busy processes. Compare individual core usage to detect a single-threaded
   bottleneck.
4. Check scheduled tasks, antivirus scans, recent deployments, and hypervisor
   CPU contention.

## Mitigation

Reduce or redistribute the workload, reschedule expensive background tasks, or
add CPU capacity. Investigate a process or driver consuming unexpected CPU
before restarting it. Confirm CPU usage and application latency return to normal.
For expected sustained workloads, tune the threshold against service objectives
and available capacity.

---
title: WindowsMemoryHighUsage
---

## Meaning

Physical memory usage exceeds the configured threshold for 15 minutes. The
default is 90%. Usage is total physical memory minus available memory; available
memory includes reclaimable pages and differs from strictly free memory.

## Impact

Memory pressure may cause paging, application latency, and allocation failures.
A high ratio alone does not prove the system is paging heavily.

## Diagnosis

1. Inspect `windows:memory_usage:ratio`, `windows:memory_used:bytes`, and
   `windows_memory_available_bytes` on the affected host.
2. Check `windows:memory_swap_page_operations:rate`, disk latency, and
   `windows:memory_committed:ratio` for related pressure.
3. In Resource Monitor, inspect processes' working sets, commit usage, and hard
   faults. Look for growth after a deployment or an unexpected workload increase.
4. Inspect `windows_memory_pool_nonpaged_bytes` and
   `windows_memory_pool_paged_bytes` if application working sets do not explain
   the usage. Driver or kernel pools may be growing.

## Mitigation

Reduce concurrent work, move workloads, or add physical memory. Investigate leaks
in the responsible application or driver. Restart a service only after assessing
its availability impact and preserving useful diagnostics. Pagefile changes
affect the commit limit and do not add physical memory. Confirm available memory
and service latency recover.

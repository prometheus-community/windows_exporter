---
title: WindowsMemoryPagingHigh
---

## Meaning

The rate of pages read from or written to disk to resolve hard page faults,
`windows:memory_swap_page_operations:rate`, exceeds the configured threshold
for 15 minutes. The default is 1000 pages per second. The counter counts
pagefile traffic and memory-mapped file I/O alike.

## Impact

The host is short of physical memory and uses the disk as backing store, so
every affected process slows down to disk speed. Sustained paging also
competes with application I/O on the system disk.

## Diagnosis

1. Inspect the paging rate together with `windows:memory_usage:ratio`,
   `windows_memory_available_bytes` and `windows:memory_committed:ratio`. Low
   available memory confirms memory pressure; a high rate with ample memory
   points at a process reading a large memory-mapped file, such as a backup or
   an index rebuild.
2. Split the rate into reads and writes with
   `windows_memory_swap_page_reads_total` and
   `windows_memory_swap_page_writes_total`. Reads dominate when the working
   sets of processes do not fit; writes dominate when the system trims and
   writes modified pages.
3. Find the processes with the largest working sets and page fault rates in
   Resource Monitor or, with the `process` collector enabled, with
   `windows_process_working_set_private_bytes` and
   `windows_process_page_faults_total`.
4. Check `windows_memory_standby_cache_lifetime_seconds`; a low lifetime means
   cached pages are repurposed quickly, another sign of memory pressure.

## Mitigation

Stop or move the process that uses the memory, fix a memory leak, or add
physical memory. Move large sequential file jobs off peak hours when they cause
the paging. Confirm the paging rate drops and available memory recovers. Raise
the threshold on hosts whose workload maps large files by design.

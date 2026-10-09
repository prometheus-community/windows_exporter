---
title: WindowsDiskReadLatencyHigh
---

## Meaning

Average disk read service time exceeds the configured threshold for 15 minutes.
The default is 0.05 seconds (50 ms) per operation, computed from read-time and
read-operation counter rates. Volumes with no reads produce no latency sample.

## Impact

Slow storage reads may delay application requests, database queries, and paging.
The recording is an average, so it does not expose tail latency.

## Diagnosis

1. Inspect `windows:logical_disk_read_latency:seconds` for the affected `volume`
   alongside read IOPS, read throughput, and `windows:logical_disk_busy:ratio`.
2. Inspect `windows_logical_disk_requests_queued` and use Resource Monitor to
   identify processes generating disk reads. Look for backup, antivirus, and
   indexing jobs.
3. Check System event logs and storage-controller logs for disk, controller,
   multipath, or network-storage errors.
4. For virtual machines, check datastore latency and competing workloads on the
   hypervisor or storage array. Low IOPS can make averages noisy.

## Mitigation

Reduce competing reads or reschedule background jobs. Address storage-path errors
and provision enough IOPS and throughput for the workload. Investigate excessive
paging if memory pressure is contributing. Tune the latency threshold to the
storage tier and application's service objectives, then confirm both service
latency and disk read latency improve.

---
title: WindowsDiskWriteLatencyHigh
---

## Meaning

Average disk write service time exceeds the configured threshold for 15 minutes.
The default is 0.05 seconds (50 ms) per operation, computed from write-time and
write-operation counter rates. Volumes with no writes produce no latency sample.

## Impact

Slow writes may delay transactions, log flushing, file uploads, or paging. An
average can conceal short periods of very high latency.

## Diagnosis

1. Inspect `windows:logical_disk_write_latency:seconds`, write IOPS, write
   throughput, `windows:logical_disk_busy:ratio`, and queued requests for the
   affected `volume`.
2. Use Resource Monitor to find processes generating writes. Review backup jobs,
   database checkpoints, log growth, and scheduled maintenance.
3. Check free space and storage-controller health. Inspect System event logs for
   disk, controller, multipath, or network-storage errors.
4. For virtual machines or remote storage, inspect the backing datastore and
   array for throughput limits and contention.

## Mitigation

Reduce competing writes or move maintenance jobs. Repair storage-path problems
and expand IOPS or throughput capacity. Coordinate database or logging changes
with the application owner. Avoid disabling durability guarantees to silence an
alert. Confirm application write latency recovers and the disk metrics improve.

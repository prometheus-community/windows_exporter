---
title: WindowsHyperVReplicationWarning
---

## Meaning

Hyper-V Replica reports warning health for the virtual machine `vm` for 15
minutes. Warning health means that some replication cycles missed their
interval, that more than a fifth of the cycles in the last hour failed, or
that the replication is waiting for an initial copy or a resynchronization.
The alert carries the `relationship` label and needs the `hyperv` collector
with its `replica_vm` subcollector. [WindowsHyperVReplicationCritical](../windowshypervreplicationcritical/)
follows when replication stops working.

## Impact

The replica lags behind the primary more than planned. The recovery point
objective is at risk, and a failover would lose more changes than expected.

## Diagnosis

1. Inspect `windows_hyperv_replica_vm_health`, `windows_hyperv_replica_vm_state`
   and `windows_hyperv_replica_vm_replication_latency_seconds` for the
   virtual machine, and `windows_hyperv_replica_vm_average_replication_size_bytes`
   to see whether the change rate of the virtual machine has grown.
2. On the primary host, read the statistics of the last hour:

   ```powershell
   Measure-VMReplication -VMName '<vm>' | Format-List *
   Get-VMReplication -VMName '<vm>' | Format-List State, Health, LastReplicationTime, FrequencySec
   ```

3. Check the bandwidth between the hosts, the load on the replica storage,
   and whether the replication interval is too short for the change rate of
   the virtual machine.

## Mitigation

Relieve the bottleneck: more bandwidth, less load on the replica storage, or a
longer replication interval through `Set-VMReplication -ReplicationFrequencySec`.
Resume a waiting replication with `Resume-VMReplication`. Reset the statistics
with `Reset-VMReplicationStatistics` after the fix; the health is computed
over the last hour and takes up to an hour to return to normal.

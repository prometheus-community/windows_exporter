---
title: WindowsHyperVReplicationCritical
---

## Meaning

Hyper-V Replica reports critical health for the virtual machine `vm` for five
minutes. The alert carries the `relationship` label, `primary` for the normal
replication and `extended` for an extended replica, and needs the `hyperv`
collector with its `replica_vm` subcollector. Critical health means that
replication has not succeeded for long enough that the replica misses the
recovery point objective, or that replication is paused because of an error.

## Impact

The replica of the virtual machine is stale. A failover to it loses the
changes since the last successful replication, and the recovery point
objective of the service is not met. If the primary host fails now, data is
lost.

## Diagnosis

1. Inspect `windows_hyperv_replica_vm_health` and
   `windows_hyperv_replica_vm_state` for the virtual machine; the state tells
   whether replication is suspended, resynchronizing or in error, and
   `windows_hyperv_replica_vm_replication_latency_seconds` how long the last
   cycle took.
2. On the primary host, read the replication details:

   ```powershell
   Measure-VMReplication -VMName '<vm>' | Format-List *
   Get-VMReplication -VMName '<vm>' | Format-List State, Health, LastReplicationTime, ReplicationMode
   Get-WinEvent -LogName Microsoft-Windows-Hyper-V-VMMS-Admin -MaxEvents 50 | Where-Object Message -like '*<vm>*'
   ```

3. Check the replica server: free space on the replica storage, the
   Hyper-V Replica HTTP or HTTPS listener, the certificate for authentication,
   and the network path between the hosts. Check whether a snapshot merge or a
   storage migration on the primary blocks replication.

## Mitigation

Fix the cause on the replica server or the network, then resume replication
with `Resume-VMReplication -VMName '<vm>'`; add `-Resynchronize` when the
state asks for a resynchronization. Confirm the health returns to normal and
the last replication time advances. If the replica cannot be recovered, remove
and re-enable replication, which sends a full initial copy.

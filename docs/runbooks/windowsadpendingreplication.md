---
title: WindowsADPendingReplication
---

## Meaning

Pending Active Directory replication operations exceed the configured threshold
for 15 minutes. The default is 50 operations. This alert is generated only with
`enableActiveDirectory: true` and requires the `ad` collector on domain controllers.

## Impact

Directory changes may be delayed between domain controllers, causing inconsistent
authentication, group membership, DNS, or policy behavior.

## Diagnosis

1. Inspect `windows_ad_replication_pending_operations` and determine whether the
   queue is growing or draining.
2. On a domain controller, inspect replication health:

   ```powershell
   repadmin /replsummary
   repadmin /showrepl
   repadmin /queue
   dcdiag /test:Replications
   ```

3. Check the Directory Service event log, partner reachability, DNS resolution,
   site links and schedules, and CPU/disk/network pressure.
4. Correlate the queue with [WindowsADReplicationFailures](../windowsadreplicationfailures/)
   and planned large directory changes.

## Mitigation

Restore connectivity and DNS between replication partners. Correct replication
errors and review site-link capacity or schedules with the directory owner.
Relieve resource pressure on domain controllers. Allow expected bulk changes to
drain while monitoring progress. Avoid forcing widespread replication or removing
partners before understanding the failure. Confirm the queue decreases and
replication health checks succeed.

---
title: WindowsClusterResourceGroupFailed
---

## Meaning

The failover cluster reports the resource group `resource_group` as failed for
five minutes. The alert needs the `mscluster` collector and fires once per
reporting node, named by the `instance` label. A group fails when a resource in
it fails and the group cannot fail over anymore: it reached its failover
threshold within the failover period, or no other node can host it.

## Impact

The whole clustered role is down and stays down until an operator intervenes.

## Diagnosis

1. Inspect `windows_mscluster_resource_state` for the resources of the group
   to find the failed resource, and `windows_mscluster_resourcegroup_owner_node`
   to see on which node the group stopped.
2. On a cluster node, read the group and its history:

   ```powershell
   Get-ClusterGroup -Name '<group>' | Format-List *
   Get-ClusterGroup -Name '<group>' | Get-ClusterResource
   Get-ClusterLog -TimeSpan 60 -Destination .
   Get-WinEvent -LogName System -FilterXPath "*[System[Provider[@Name='Microsoft-Windows-FailoverClustering'] and (EventID=1069 or EventID=1205 or EventID=1254)]]" -MaxEvents 20
   ```

   Event 1254 says the group exceeded its failover threshold.
3. Follow [WindowsClusterResourceFailed](../windowsclusterresourcefailed/) for
   the failed resource, and check the possible owners of the group.

## Mitigation

Fix the failed resource, then bring the group online with `Start-ClusterGroup`
on a node that can host it. A group that failed over repeatedly may need a
higher failover threshold or a longer failover period, but a resource that
fails on every node needs its cause fixed first. Confirm the group state
returns to online.

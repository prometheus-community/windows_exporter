---
title: WindowsClusterResourceFailed
---

## Meaning

The failover cluster reports the resource `resource` of type `type` in the
group `resource_group` as failed for five minutes. The alert needs the
`mscluster` collector and fires once per reporting node, named by the
`instance` label. A failed resource has exhausted its restart attempts on the
current node, or its group has reached its failover threshold.

## Impact

The role that depends on the resource is down: a clustered disk, IP address,
network name, file share, SQL Server instance or virtual machine. Other
resources in the same group that depend on it are offline too.

## Diagnosis

1. Inspect `windows_mscluster_resource_state` for the group and
   `windows_mscluster_resourcegroup_state` and `windows_mscluster_resource_owner_node`
   to see where the group is and whether the whole group failed.
2. On a cluster node, read the resource and the reason:

   ```powershell
   Get-ClusterResource -Name '<resource>' | Format-List *
   Get-ClusterResource -Name '<resource>' | Get-ClusterParameter
   Get-ClusterLog -TimeSpan 30 -Destination .
   Get-WinEvent -LogName System -FilterXPath "*[System[Provider[@Name='Microsoft-Windows-FailoverClustering'] and (EventID=1069 or EventID=1205)]]" -MaxEvents 20
   ```

   Event 1069 names the failed resource and its error code.
3. Check the dependency of the resource: a disk that is not reachable from the
   node, an IP address whose network is down, a network name whose computer
   object in Active Directory is missing or disabled, or an application whose
   service fails to start.

## Mitigation

Fix the dependency, then bring the resource online with
`Start-ClusterResource`, or move the group to another node with
`Move-ClusterGroup`. Review the restart and failover thresholds of the
resource if it failed because of a transient problem. Confirm the resource
state returns to online and the group is online.

---
title: WindowsClusterSharedVolumeAlmostFull
---

## Meaning

The Cluster Shared Volume `shared_volume` has less free space than the
configured threshold for 15 minutes. The warning fires below
`diskFreeWarningThreshold` (10 percent by default) and the critical alert
below `diskFreeCriticalThreshold` (5 percent); both fire below the critical
threshold, so inhibit the warning in Alertmanager. The alert needs the
`mscluster` collector and fires once per reporting node, named by the
`instance` label. `sharedVolumeSelector` excludes the
`ClusterPerformanceHistory` volume of the cluster's own health history.

## Impact

Virtual machines and databases on the volume cannot grow their files. Dynamic
virtual hard disks pause the virtual machine when the volume is full, Hyper-V
checkpoints and backups fail, and SQL Server stops writing.

## Diagnosis

1. Compare `windows_mscluster_shared_volumes_free_bytes` and
   `windows_mscluster_shared_volumes_total_bytes` over the last days to see
   whether the volume fills steadily or filled suddenly.
2. On a cluster node, find the large files on the volume:

   ```powershell
   Get-ClusterSharedVolume | Select-Object Name, @{n='FreeGB';e={[int]($_.SharedVolumeInfo.Partition.FreeSpace/1GB)}}, @{n='SizeGB';e={[int]($_.SharedVolumeInfo.Partition.Size/1GB)}}
   Get-ChildItem 'C:\ClusterStorage\<volume>' -Recurse -File | Sort-Object Length -Descending | Select-Object -First 20 FullName, @{n='GB';e={[int]($_.Length/1GB)}}
   ```

3. Look for leftover checkpoints (`.avhdx` files), orphaned virtual hard
   disks, backup snapshots that were not cleaned up, and dynamic disks that
   grew.

## Mitigation

Delete or merge checkpoints, remove orphaned files, move virtual machines to
another volume with storage migration, or extend the volume on the storage
array and then in the cluster with `Resize-Partition`. Confirm free space is
above the threshold. Raise the thresholds only for volumes whose growth is
known and planned.

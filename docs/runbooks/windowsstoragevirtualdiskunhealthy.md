---
title: WindowsStorageVirtualDiskUnhealthy
---

## Meaning

The Storage Spaces virtual disk `virtual_disk` reports a health status of
warning or unhealthy for ten minutes. The critical alert fires for unhealthy,
the warning for warning. The alert needs the `storage_spaces` collector. A
virtual disk is in warning state when it is degraded or incomplete and still
accessible, and unhealthy when it is detached or has lost data. On Storage
Spaces Direct every node reports the clustered virtual disks, so the alert
fires once per node; the `instance` label names the reporter.

## Impact

A degraded virtual disk still serves data but has no redundancy left; a
further disk failure loses the data. A detached virtual disk takes its volumes
offline, and everything stored on them is unavailable.

## Diagnosis

1. Inspect `windows_storage_spaces_virtual_disk_health_status` and
   `windows_storage_spaces_pool_health_status` to see whether the pool is
   affected too; [WindowsStoragePoolUnhealthy](../windowsstoragepoolunhealthy/)
   then names the pool.
2. On the host, read the virtual disk, its footprint and the repair jobs:

   ```powershell
   Get-VirtualDisk -FriendlyName '<disk>' | Select-Object FriendlyName, HealthStatus, OperationalStatus, ResiliencySettingName, DetachedReason
   Get-VirtualDisk -FriendlyName '<disk>' | Get-PhysicalDisk | Select-Object FriendlyName, HealthStatus, OperationalStatus
   Get-StorageJob
   ```

3. Check for a failed physical disk, a pool that ran out of space for
   allocation, or a disk that was detached by a policy after an error.

## Mitigation

Replace the failed physical disk as in [WindowsStoragePoolUnhealthy](../windowsstoragepoolunhealthy/)
and repair the virtual disk with `Repair-VirtualDisk`. Attach a detached disk
with `Connect-VirtualDisk` once the cause is fixed. Confirm the health returns
to healthy once the repair jobs finish.

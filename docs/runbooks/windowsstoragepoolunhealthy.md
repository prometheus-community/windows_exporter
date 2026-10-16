---
title: WindowsStoragePoolUnhealthy
---

## Meaning

The Storage Spaces pool `pool` reports a health status of warning or unhealthy
for ten minutes. The critical alert fires for unhealthy, the warning for
warning. The alert needs the `storage_spaces` collector and joins
`windows_storage_spaces_pool_info` so the primordial pool, which only groups
the disks that are not in a pool yet, does not raise it. The pool's
operational status, exposed as `windows_storage_spaces_pool_operational_status`,
tells why: degraded, read-only, or incomplete. On Storage Spaces Direct every
node reports the clustered pool, so the alert fires once per node; the
`instance` label names the reporter.

## Impact

A pool in warning state has lost a physical disk or redundancy; the virtual
disks on it still work but have no tolerance left for another failure, or
cannot allocate more space. An unhealthy pool has lost more disks than its
resiliency tolerates, so virtual disks on it are detached or read-only.

## Diagnosis

1. Inspect `windows_storage_spaces_pool_health_status`,
   `windows_storage_spaces_pool_operational_status` and
   `windows_storage_spaces_virtual_disk_health_status` for the pool and its
   virtual disks.
2. On the host, find the failed disk:

   ```powershell
   Get-StoragePool -FriendlyName '<pool>' | Select-Object FriendlyName, HealthStatus, OperationalStatus, IsReadOnly
   Get-StoragePool -FriendlyName '<pool>' | Get-PhysicalDisk | Select-Object FriendlyName, SerialNumber, HealthStatus, OperationalStatus, Usage
   Get-StoragePool -FriendlyName '<pool>' | Get-VirtualDisk | Select-Object FriendlyName, HealthStatus, OperationalStatus
   Get-StorageJob
   ```

3. Check the System log for disk and storage events, and whether a repair job
   is already running.

## Mitigation

Replace the failed physical disk: retire it with
`Set-PhysicalDisk -Usage Retired`, add the new disk with `Add-PhysicalDisk`,
repair the virtual disks with `Repair-VirtualDisk`, and remove the old disk
with `Remove-PhysicalDisk`. Add capacity when the pool is full. Confirm the
health returns to healthy once the repair jobs finish.

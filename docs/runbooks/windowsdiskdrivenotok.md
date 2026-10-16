---
title: WindowsDiskDriveNotOK
---

## Meaning

The physical drive `device_id` reports an operational status other than `OK`
for five minutes. The alert carries the status, such as `Pred Fail`, `Degraded`
or `Error`, in its `status` label. It needs the `diskdrive` collector, which
reads the status that the drive and its driver report through WMI. The
`Unknown` status does not raise the alert: some USB and virtual drives report
it permanently. Transient statuses such as `Starting` or `Stopping` raise it
only when they last longer than five minutes.

## Impact

`Pred Fail` is the drive's own prediction that it is about to fail. Any other
status means the drive is not fully working. Data on a single drive is at risk;
a drive in a mirror or parity set leaves the set without redundancy until it is
replaced.

## Diagnosis

1. Inspect `windows_diskdrive_status` for the host and check whether the
   volumes on the drive still report normal latency with
   `windows:logical_disk_read_latency:seconds` and
   `windows:logical_disk_write_latency:seconds`.
2. On the host, read the drive and its reliability counters:

   ```powershell
   Get-CimInstance Win32_DiskDrive | Select-Object DeviceID, Model, Status
   Get-PhysicalDisk | Select-Object FriendlyName, HealthStatus, OperationalStatus
   Get-PhysicalDisk | Get-StorageReliabilityCounter | Select-Object DeviceId, ReadErrorsTotal, WriteErrorsTotal, Temperature, Wear
   ```

3. Check the System event log for disk, storahci, stornvme or RAID controller
   events. On a virtual machine, check the hypervisor and the datastore of the
   virtual disk.

## Mitigation

Back up the data on the drive first. Replace a drive that predicts its failure
or reports errors through the hardware or virtualization team. For a drive in
a storage pool or RAID set, follow the vendor's replacement and rebuild
procedure and confirm the set is redundant again. The alert resolves when the
status returns to `OK`.

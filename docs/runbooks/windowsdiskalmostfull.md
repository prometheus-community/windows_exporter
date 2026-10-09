---
title: WindowsDiskAlmostFull
---

## Meaning

The volume's free space remains below a configured threshold for 15 minutes.
Defaults are 10% for warning and 5% for critical. Both severity levels fire below
the critical threshold unless Alertmanager inhibition suppresses the warning.
Windows can delay the underlying space counters by 10–15 minutes.

## Impact

Applications may fail to write data or logs. Low free space can prevent updates,
pagefile growth, database operations, and normal operating system behavior.

## Diagnosis

1. Use the `volume` label to identify the affected filesystem and inspect
   `windows:logical_disk_free:ratio` and `windows_logical_disk_free_bytes`.
2. Verify current capacity locally; the exporter counters may lag:

   ```powershell
   Get-Volume | Select-Object DriveLetter, FileSystemLabel, SizeRemaining, Size
   ```

3. Identify growing logs, temporary files, backups, database files, and application
   data. Check retention policies and recent jobs.
4. Consult [WindowsDiskFillingUp](../windowsdiskfillingup/) when free space is
   declining rapidly.

## Mitigation

Expand the volume or remove files through the owning application's supported
retention and cleanup process. Preserve required data and backups; avoid manually
deleting database files or Windows system files. Reduce incoming writes if the
volume may fill before capacity is restored. Verify space locally and allow the
Windows counters to refresh before expecting the alert to resolve.

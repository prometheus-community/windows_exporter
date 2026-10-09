---
title: WindowsMemoryCommitHighUsage
---

## Meaning

Committed memory exceeds the configured percentage of the system commit limit
for 15 minutes. The default threshold is 90%. The commit limit depends on
physical memory and pagefile capacity.

## Impact

The system may reject new memory commitments, causing application failures even
when some physical memory remains available.

## Diagnosis

1. Inspect `windows:memory_committed:ratio`, `windows_memory_committed_bytes`,
   and `windows_memory_commit_limit`. Determine whether committed bytes increased
   or the limit decreased.
2. In Task Manager or Resource Monitor, inspect per-process commit size and look
   for persistent growth. Check recent workload and deployment changes.
3. Inspect pagefile configuration and the free space on its volumes:

   ```powershell
   Get-CimInstance Win32_PageFileUsage
   Get-CimInstance Win32_PageFileSetting
   Get-Volume
   ```

4. Check physical memory pressure and paging alongside this alert. The recording
   rule omits invalid zero commit limits rather than reporting infinity.

## Mitigation

Reduce memory commitments or address the leaking process. Restore pagefile
capacity and disk free space when the commit limit is unexpectedly low. Review
pagefile sizing with the workload owner; some changes require a reboot. Add
physical memory for sustained demand. Confirm committed bytes remain safely
below the commit limit and applications can allocate memory normally.

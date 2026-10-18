---
title: WindowsPagefileAlmostFull
---

## Meaning

The page file `file` has been more than the configured share in use for 15
minutes. The default `pagefileUsageThreshold` is 80 percent of
`windows_pagefile_limit_bytes`, the current size of the file. The alert needs
the `pagefile` collector. A system-managed page file grows up to its maximum
when it fills, so a high usage of the current size on such a system shows that
the file has already grown.

## Impact

When the page file is full, Windows cannot commit more memory. Allocations
fail, processes crash with out-of-memory errors, and the system can become
unresponsive. The commit limit alert [WindowsMemoryCommitHighUsage](../windowsmemorycommithighusage/)
usually fires at the same time.

## Diagnosis

1. Compare `windows_pagefile_free_bytes` and `windows_pagefile_limit_bytes`
   with `windows:memory_committed:ratio` and `windows_memory_available_bytes`
   to see whether physical memory is exhausted too.
2. Find the processes that hold the commit charge:

   ```powershell
   Get-Process | Sort-Object PagedMemorySize64 -Descending | Select-Object -First 10 Name, Id, @{n='CommitMB';e={[int]($_.PagedMemorySize64/1MB)}}
   Get-CimInstance Win32_PageFileUsage | Select-Object Name, AllocatedBaseSize, CurrentUsage, PeakUsage
   ```

3. Check for a recent change: a new workload, a memory leak that grows over
   days, or a page file that was made smaller or fixed in size.

## Mitigation

Stop or restart the process that leaks memory, or move workload off the host.
Add physical memory or enlarge the page file through the normal change
process; a fixed page file needs enough room for the peak commit charge of the
host. Confirm usage falls below the threshold and stays there.

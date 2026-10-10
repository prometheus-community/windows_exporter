# process collector

IIS application pool names (`--collector.process.iis`) come from two sources.

- Primary: the `-ap "<pool>"` argument of each `w3wp` command line, read with
  `NtQueryInformationProcess(ProcessCommandLineInformation)`. It needs only
  `PROCESS_QUERY_LIMITED_INFORMATION` and Windows 8.1 or Server 2012 R2.
- Fallback: `WorkerProcess` in `root\WebAdministration`, used only when it
  answered during `Build` and a command line can't be read or has no `-ap`.
  It needs the optional IIS Management Scripts and Tools feature
  (`Web-Scripting-Tools`), and asks WAS through RSCA on every query.
- `splitCommandLine` follows `CommandLineToArgvW`. `FuzzSplitCommandLine`
  compares it with `windows.DecomposeCommandLine`; run it after parser changes
  with `go test -run '^$' -fuzz FuzzSplitCommandLine ./internal/collector/process/`.
- Without IIS, `TestCollectorWorkerProcessCommandLine` starts a suspended copy
  of the test binary named `w3wp.exe` to check the published label. The IIS
  tests compare against WMI when available and against the
  `IIS APPPOOL\<pool>` process owner.
- Don't create or change IIS application pools on a development machine to
  test edge cases; use suspended processes with crafted command lines.

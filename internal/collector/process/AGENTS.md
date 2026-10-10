# Process collector

The collector reads all processes with one
`NtQuerySystemInformation(SystemProcessInformation)` call, not with the perflib
`Process` or PDH `Process V2` counter sets. The perflib provider (perfproc) reads
the same kernel data, so the metric values must stay identical to the `Process`
counter set. `CounterVersion` and `--collector.process.counter-version` are
deprecated and ignored; keep the field for library users such as Grafana Alloy.

- Process names follow the perflib instance names: the image name without a
  case-insensitive `.exe` suffix, other extensions such as `.scr` are kept, and
  PID 0 is `Idle`. Users' include and exclude expressions depend on these names.
- The `windows_process_info` labels (owner, command line, process group ID) need
  `OpenProcess` and are cached by PID and creation time, because Windows reuses
  PIDs. The creation time of the opened handle is compared with the snapshot,
  so a PID reused between the snapshot and the lookup is not attributed to the
  old process. Failed lookups are not cached and retry on the next scrape;
  access-denied results are cached as empty labels.
- Read the snapshot buffer with bounds checks against the returned length, and
  keep it 8-byte aligned (`[]uint64` backing array).

To check value parity after changes, compare the snapshot against the `Process`
counter set by PID in a temporary test, for example with
`pdh.NewCollector(..., pdh.CounterTypeRaw, "Process", pdh.InstancesAll)`, including
processes with names like `a.b.exe` and `tool.scr`. Run it as a standard user
too: neither data source needs administrator rights.

## IIS application pools

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

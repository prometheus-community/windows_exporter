# iis collector

- The `Web Service` and `Web Service Cache` objects are served by the perflib
  V1 provider `w3ctrs.dll`, which reads a counter snapshot that W3SVC writes to
  shared memory. When the snapshot is older than `FreshTimeForCounters`
  (default 15 s), the provider signals W3SVC, calls `Sleep` for
  `CheckCountersEveryNMiliseconds` (default 250 ms) and checks the snapshot
  again, up to `NumberOfTimesToCheckCounters` (default 4) times. The defaults
  are built into the DLL; nonzero `REG_DWORD` values below
  `HKLM\SYSTEM\CurrentControlSet\Services\W3SVC\Performance` override them.
  Microsoft does not document these values.
- So a scrape that hits a stale snapshot waits 250 ms, up to 4 × 250 ms,
  inside `PdhCollectQueryData` of the `Web Service` query. Go execution traces
  from a production server showed 259 to 261 ms there, and PDH reads of
  `system`, `logical_disk` and `net` that started at the same moment blocked
  for the same time. `cpu`, `memory` and the other IIS objects, including
  `Web Service Cache`, were never slow, and there were no IIS event log
  entries. The exact lock inside Windows that makes the other reads wait is
  not known
  ([#2698](https://github.com/prometheus-community/windows_exporter/issues/2698)).
- This is Windows behavior, not an exporter lock or timeout. Don't look for an
  exporter-side cause of flat ~250 ms `iis` durations, and don't add
  workarounds that hide them. A repro on a GitHub runner with IIS did not show
  it in 40 minutes, so don't expect to reproduce it in CI.
- To find which Windows call a collector blocks in, capture a Go execution
  trace with `--debug.enabled` and `/debug/pprof/trace?seconds=600`. Collectors
  waiting in a Windows call use no CPU, so a CPU profile does not show it.

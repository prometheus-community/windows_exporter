# scheduled_task collector

Scrape time is spent in the Task Scheduler service. The exporter's own CPU
time is small, but the Schedule service (`svchost`) uses about as much CPU time
as the scrape takes. Measure that with the raw `PercentProcessorTime` of
`Win32_PerfRawData_PerfProc_Process` for the service's PID (100 ns units),
which an unprivileged user can read.

- `IRegisteredTask` `Name` and `Path` are answered locally from the enumerated
  collection. `State`, `Enabled`, `LastTaskResult` and `NumberOfMissedRuns` each
  make a separate RPC to the Schedule service (roughly 0.1–0.3 ms per call).
  `ITaskFolder::GetTasks` and `GetFolders` add one RPC per folder each.
  `IRegisteredTaskCollection::get_Item` is local, so enumerating with
  `_NewEnum` and batched `IEnumVARIANT::Next` saves nothing.
- Keep the serial walk. Reading folders with concurrent workers, each with its
  own connection, halved the scrape time but cost the Schedule service about
  40% more CPU time in total with four workers. That is a bad trade on small
  VMs (#2643). Re-opening folders by path instead of walking the folder
  objects also adds a round trip per folder.
- Read only the properties that are published, apply the include/exclude
  filter to the path before any service read, and keep property reads
  conditional on whether the metric that uses them is emitted.
- `folderFilter` skips subfolders whose tasks the include/exclude expressions
  reject whatever their names are, by analyzing the parsed expressions. It must
  stay conservative: when in doubt, walk the folder. A Windows 11 desktop has
  about 150 folders for about 200 tasks, and listing the tasks and subfolders of
  each costs ~0.15 ms per call, so `exclude=/Microsoft/.+` saves most of a scrape.
  A folder that the include override may match is always walked, and an
  override the filter cannot parse disables folder skipping.
- `taskReader` carries the task and folder filters and which properties to
  read. With `skip-missed-runs`, `NumberOfMissedRuns` is neither read nor
  published, which saves its two RPCs per task.
- The fake native task in `scheduled_task_native_test.go` counts reads per
  vtable slot; extend those assertions when adding a property read.
- To find slow calls, time each call type over a full enumeration on a live
  system. Run it as `SYSTEM` as well as a user: unprivileged accounts see fewer
  tasks.
- Each task publishes 19 const metrics, and building them is most of the
  allocations of a scrape. `collectMetrics` reuses the metrics of the previous
  scrape through `internal/metriccache` while a task's path, state, last result
  and missed runs are unchanged. Add new inputs of a task's metrics to
  `taskValues`, or the cache returns stale series.

## Measured costs and rejected approaches

Measured in October 2026 on a Windows 11 24H2 desktop (taskschd.dll
10.0.26100.5074), unprivileged: 207 tasks in 150 folders (134 leaf folders, 36
of them empty), about 150 ms per scrape without filters.

- Per call: `GetTasks` ~170 µs and `GetFolders` ~145 µs per folder; `State` and
  `LastTaskResult` ~130 µs per task; `NumberOfMissedRuns` ~310 µs per task
  (~560 µs for disabled tasks). `Path` and folder `Path` take 1–2 µs (local).
  `NumberOfMissedRuns` alone is about 28% of a scrape.
- `NumberOfMissedRuns` is computed with two RPCs: `SchRpcGetLastRunInfo`
  (opnum 16), then `SchRpcScheduledRuntimes` (opnum 15) from the start of the last run
  to the current local time with an unlimited `cRequested`. The service
  computes every scheduled run in that window from the task definition and
  returns all of them, so tasks with many missed runs (13,146 seen) send large
  responses each scrape. Tasks without time triggers return
  `SCHED_S_TASK_NOT_SCHEDULED`. `LastTaskResult` also calls
  `SchRpcGetLastRunInfo`, so reading both costs one duplicate round trip.
  `SchRpcGetNumberOfMissedRuns` (opnum 18) always returns 0 and is not what COM
  uses.
- Missed runs are rare but real: 6 of 207 tasks were non-zero, 2 of them
  disabled. Skipping the read for disabled tasks would change published values.
- A persistent COM connection (dedicated thread, service kept connected) saves
  ~1.7 ms per scrape. `CoCreateInstance` and `Connect` are cheap; the service's
  per-call time dominates.
- Direct MS-TSCH RPC over `ncalrpc` (hand-marshalled NDR, `I_RpcSendReceive`)
  is only 5–8 µs faster per call. It would only save the duplicate
  `SchRpcGetLastRunInfo` (~18 ms per scrape, ~12%) plus ~9 ms of client CPU, at
  the cost of a hand-written NDR client. Not worth it. Note that
  `SchRpcGetTaskInfo` only returns the state with `SCH_FLAG_STATE`
  (`0x10000000`).
- One `SchRpcEnumInstances` for all running tasks took ~21 ms, slower than the
  per-task state reads.
- There is no flat or recursive task enumeration: `GetTasks`/`GetFolders` and
  `SchRpcEnumTasks`/`SchRpcEnumFolders` list one folder each, with only the
  hidden-task flag. The registry `TaskCache` is undocumented and not readable
  without admin rights.
- `MSFT_ScheduledTask` (CIM, `root/Microsoft/Windows/TaskScheduler`) returns a
  flat list with path and state, but is about 3× slower: ~440 ms through
  `internal/mi` in-process for the same 207 tasks, whatever properties are
  selected, plus ~120 ms for last result and missed runs. The work moves into
  `WmiPrvSE` (up to ~30% of one core while querying).
- `missed_runs` can be dropped with `skip-missed-runs`. Not taken yet: a cache
  duration that serves the last task list.
- Profiling the exporter with pprof on Windows attributes the time blocked in
  these COM calls to `runtime.cgocall`; that is waiting, not exporter CPU. The
  Schedule service's own CPU needs an elevated reader of its process times.

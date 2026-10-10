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
- The Schedule service answers requests from several connections in parallel,
  but concurrent requests cost it more CPU time in total (about 40% more with
  four workers). `taskFolderWorkers` therefore reads serially, with main's
  folder-object walk, on hosts with fewer than eight logical CPUs.
  The serial walk never uses `walkTaskFolders`, whose `Folder(path)` per folder
  adds service CPU on its own.
- `walkTaskFolders` reads folders with several goroutines, each with its own
  locked OS thread, MTA initialization and connected `ITaskService`; no
  interface pointer crosses workers, only folder paths do. Tasks are returned
  in depth-first order regardless of which worker read a folder. Test the walk
  with the pure Go readers in `scheduled_task_walk_test.go`. The workers
  inherit the `collector` pprof label from the collect goroutine.
- Read only the properties that are published, apply the include/exclude
  filter to the path before any service read, and keep property reads
  conditional on whether the metric that uses them is emitted.
- The fake native task in `scheduled_task_native_test.go` counts reads per
  vtable slot; extend those assertions when adding a property read.
- To find slow calls, time each call type over a full enumeration on a live
  system. Run it as `SYSTEM` as well as a user: unprivileged accounts see fewer
  tasks.

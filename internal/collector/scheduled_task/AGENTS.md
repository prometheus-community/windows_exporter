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

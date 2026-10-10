# scheduled_task collector

Scrape time is spent waiting on the Task Scheduler service, not on CPU.

- `IRegisteredTask` `Name` and `Path` are answered locally from the enumerated
  collection. `State`, `Enabled`, `LastTaskResult` and `NumberOfMissedRuns` each
  make a separate RPC to the Schedule service (roughly 0.1–0.3 ms per call).
  `ITaskFolder::GetTasks` and `GetFolders` add one RPC per folder each.
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

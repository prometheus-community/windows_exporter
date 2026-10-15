# MI package guidance

`Session.Query`, `QueryUnmarshal` and `QueryFunc` receive results through
`MI_OperationCallbacks` ([`callbacks.go`](callbacks.go)). Do not move them back
to the synchronous `MI_Operation_GetInstance` loop: when its result hand-over
waits, e.g. for parallel queries or slow providers such as `MSFT_StoragePool`,
the WMI client leaks Event handles (`miutils!RtlInterlockedCompareWait`).
`Test_MI_ParallelQuery_HandleGrowth` guards this. `QueryInstances` and
`Operation.GetInstance` remain as the synchronous low-level API for tests.

Rules for the callback path:

- Create native callbacks once (`sync.OnceValue`) and identify the query by
  the callback context. `windows.NewCallback` never frees a callback and a
  process can create only about 2000.
- Keep the callback context and the `MI_OperationCallbacks` pinned until
  `MI_Operation_Close` has returned.
- Never call `Cancel` or `Close` from a callback. The goroutine that started
  the query cancels on request and closes after the final callback
  (`moreResults == MI_FALSE`).
- Recover panics in callbacks; they must not unwind into MI.
- Never block a callback on a goroutine. `QueryFunc` copies each instance
  (`MI_Instance_Clone`) in the callback and runs `fn` for the copies on the
  calling goroutine. Handing the original over and waiting for `fn` costs two
  thread wake-ups per instance, about 25% more process CPU than the copy
  (`Benchmark_MI_ProcessCPU`).
- MI reports parameter errors from within `MI_Session_QueryInstances`; the
  runtime runs that callback on the calling goroutine. Never run caller code
  such as the `QueryFunc` handler inside a callback: a panic cannot reach the
  caller there and `runtime.Goexit` is fatal.
- While waiting for callbacks, keep a timer pending. Otherwise the runtime can
  declare a false deadlock (golang/go#55015, still open in Go 1.27);
  `Test_MI_Query_DeadlockDetector` reproduces it. Do not use a goroutine that
  sleeps forever.
- Guard shared callback state with a mutex: results of one operation may
  arrive on different MI threads. `-race` cannot reveal a missing lock there,
  because the runtime treats every callback as synchronized with every native
  call.

Query timeouts and options:

- Every query runs with an MI timeout. MI treats a zero timeout as unset, so
  a non-positive timeout selects `DefaultQueryTimeout` (the session default
  options) and a positive one below 1ms is raised to 1ms.
- Operation options must outlive every operation that uses them. Options for
  whole-second timeouts are cached per session and shared by concurrent
  operations; `Session.Close` deletes them after `MI_Session_Close` has
  closed every operation. Other timeouts, e.g. the remaining scrape budget,
  get per-query options that are deleted after the operation is closed.

Cancellation: `QueryContext` and `QueryFuncContext` do not start a query
for a done context. When it is done during the query, `wait` fails the query
through `fail`/`cancelRequested`, so `Cancel` runs on the waiting goroutine,
and still waits for the final callback before `close` frees anything. The
handler is not called after the context is done; a copy it skips fails the
query with the context's cause, even after the final result, so partial
results never look complete. A context that is done after the final result
(`complete`, set under `mu`) does not fail a query that has nothing left to
hand over.

Run `CGO_ENABLED=1 go test -race` (gcc from w64devkit) and a
`-gcflags=all=-d=checkptr` run for changes in this package.

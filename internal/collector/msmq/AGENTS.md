# `msmq` collector

The collector reads the PDH object `MSMQ Queue` with `pdh.CounterTypeRaw`. The
object exists only where the Message Queuing feature (MSMQ service) is
installed. It is multi-instance. These notes were recorded on Windows Server
2025 (build 26100) with no application queues.

- All four counters are `PERF_COUNTER_LARGE_RAWCOUNT` and are gauges: the help
  text says they count messages or bytes that "currently reside" in the queue.
  `Bytes ...` counters are bytes.
- Instances: `Computer Queues` plus one instance per queue, named
  `<computer>\private$\<queue>` in lower case. A fresh install already has the
  internal queues `server2025\private$\admin_queue$` and
  `server2025\private$\order_queue$`. The collector uses the instance name
  unchanged as the `queue` label.
- For `Computer Queues`, `Messages/Bytes in Queue` is the dead-letter queue
  and `Messages/Bytes in Journal Queue` is the computer journal.
- All values were 0 on the recording host.

## Counter reference

Recorded on Windows Server 2025 (build 26100). `Raw` and `Base` are the first and second
values that `pdh.CounterTypeRaw` returns for the sample instance.

### `MSMQ Queue`

Multi-instance. Instances: `Computer Queues`, `server2025\private$\admin_queue$`, `server2025\private$\order_queue$`. Sample instance: `Computer Queues`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Bytes in Journal Queue` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Bytes in Queue` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Messages in Journal Queue` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Messages in Queue` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |

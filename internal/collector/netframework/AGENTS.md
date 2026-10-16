# `netframework` collector

The collector reads the eight .NET Framework PDH objects `.NET CLR Exceptions`,
`.NET CLR Interop`, `.NET CLR Jit`, `.NET CLR Loading`,
`.NET CLR LocksAndThreads`, `.NET CLR Memory`, `.NET CLR Remoting` and
`.NET CLR Security` with `pdh.CounterTypeRaw`, all in one shared query. The
objects come from the .NET Framework performance counter provider
(`mscoree.dll`), not from .NET (Core). These notes were recorded on the
`windows-2022` GitHub Actions runner with a CI probe, because the objects are
missing on the German development machine although the provider is
registered.

- Every object has one instance per process that loaded the .NET Framework
  CLR, named after the process image without `.exe`, plus `_Global_`. Two
  processes with the same image report the same instance name, e.g.
  `powershell` twice. The collector skips `_Global_` and publishes later
  duplicates with an occurrence suffix (`powershell#1`).
- `% Time in Jit`, `% Time in RT checks` and `% Time in GC` are
  `PERF_RAW_FRACTION`. Their value is raw / base (`,secondvalue`), not a
  percentage and not a cumulative value. Per the .NET documentation, they
  cover the time since the end of the previous JIT, check or GC phase and are
  updated when the next one ends. A base of 0 means
  no sample yet; skip the ratio then. On `_Global_` the base is `0xFFFFFFFF`
  and the ratio is meaningless.
- `% Time Loading` and `% Time Sig. Authenticating` are
  `PERF_COUNTER_TIMER`, not fractions. The collector does not read them.
- Every `/ sec` counter is `PERF_COUNTER_COUNTER` or `PERF_COUNTER_BULK_COUNT`,
  whose raw value is a cumulative count since the process started, except the
  Interop `# of TLB exports / sec` and `# of TLB imports / sec`, which are
  `PERF_COUNTER_RAWCOUNT`.
- The `.NET CLR Memory` sizes and counts are `PERF_COUNTER_LARGE_RAWCOUNT`
  gauges in bytes or counts; `# Gen N Collections` and `# Induced GC` are
  cumulative. The object also has a `Process ID` counter.

## Counter reference

Recorded on the `windows-2022` GitHub Actions runner. `Raw` and `Base` are the
first and second values that `pdh.CounterTypeRaw` returns for a `powershell`
instance; `Base` is shown for fraction counters only.

### `.NET CLR Exceptions`

One instance per process that loaded the CLR, plus `_Global_`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `# of Exceps Thrown` | `PERF_COUNTER_RAWCOUNT` | 6 |  |
| `# of Filters / sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `# of Finallys / sec` | `PERF_COUNTER_COUNTER` | 14 |  |
| `Throw To Catch Depth / sec` | `PERF_COUNTER_COUNTER` | 30 |  |

Not read by the collector:

- `PERF_COUNTER_COUNTER`: `# of Exceps Thrown / sec`

### `.NET CLR Interop`

One instance per process that loaded the CLR, plus `_Global_`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `# of CCWs` | `PERF_COUNTER_RAWCOUNT` | 5 |  |
| `# of marshalling` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `# of Stubs` | `PERF_COUNTER_RAWCOUNT` | 226 |  |

Not read by the collector:

- `PERF_COUNTER_RAWCOUNT`: `# of TLB exports / sec`, `# of TLB imports / sec`

### `.NET CLR Jit`

One instance per process that loaded the CLR, plus `_Global_`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `# of Methods Jitted` | `PERF_COUNTER_RAWCOUNT` | 49 |  |
| `% Time in Jit` | `PERF_RAW_FRACTION` | 0 | 10000 |
| `Standard Jit Failures` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total # of IL Bytes Jitted` | `PERF_COUNTER_RAWCOUNT` | 6261 |  |

Not read by the collector:

- `PERF_COUNTER_RAWCOUNT`: `# of IL Bytes Jitted`
- `PERF_COUNTER_COUNTER`: `IL Bytes Jitted / sec`

### `.NET CLR Loading`

One instance per process that loaded the CLR, plus `_Global_`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Bytes in Loader Heap` | `PERF_COUNTER_LARGE_RAWCOUNT` | 1957888 |  |
| `Current appdomains` | `PERF_COUNTER_RAWCOUNT` | 1 |  |
| `Current Assemblies` | `PERF_COUNTER_RAWCOUNT` | 17 |  |
| `Current Classes Loaded` | `PERF_COUNTER_RAWCOUNT` | 352 |  |
| `Total # of Load Failures` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Appdomains` | `PERF_COUNTER_RAWCOUNT` | 1 |  |
| `Total appdomains unloaded` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Assemblies` | `PERF_COUNTER_RAWCOUNT` | 17 |  |
| `Total Classes Loaded` | `PERF_COUNTER_RAWCOUNT` | 352 |  |

Not read by the collector:

- `PERF_COUNTER_COUNTER`: `Rate of appdomains`, `Rate of appdomains unloaded`, `Rate of Assemblies`, `Rate of Classes Loaded`, `Rate of Load Failures`
- `PERF_COUNTER_TIMER`: `% Time Loading`
- `PERF_COUNTER_RAWCOUNT`: `Assembly Search Length`

### `.NET CLR LocksAndThreads`

One instance per process that loaded the CLR, plus `_Global_`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `# of current logical Threads` | `PERF_COUNTER_RAWCOUNT` | 10 |  |
| `# of current physical Threads` | `PERF_COUNTER_RAWCOUNT` | 8 |  |
| `# of current recognized threads` | `PERF_COUNTER_RAWCOUNT` | 2 |  |
| `# of total recognized threads` | `PERF_COUNTER_RAWCOUNT` | 2 |  |
| `Current Queue Length` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Queue Length Peak` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total # of Contentions` | `PERF_COUNTER_RAWCOUNT` | 0 |  |

Not read by the collector:

- `PERF_COUNTER_COUNTER`: `Contention Rate / sec`, `Queue Length / sec`, `rate of recognized threads / sec`

### `.NET CLR Memory`

One instance per process that loaded the CLR, plus `_Global_`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `# GC Handles` | `PERF_COUNTER_LARGE_RAWCOUNT` | 328 |  |
| `# Gen 0 Collections` | `PERF_COUNTER_LARGE_RAWCOUNT` | 1 |  |
| `# Gen 1 Collections` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `# Gen 2 Collections` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `# Induced GC` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `# of Pinned Objects` | `PERF_COUNTER_LARGE_RAWCOUNT` | 2 |  |
| `# of Sink Blocks in use` | `PERF_COUNTER_LARGE_RAWCOUNT` | 20 |  |
| `# Total committed Bytes` | `PERF_COUNTER_LARGE_RAWCOUNT` | 17883136 |  |
| `# Total reserved Bytes` | `PERF_COUNTER_LARGE_RAWCOUNT` | 402644992 |  |
| `% Time in GC` | `PERF_RAW_FRACTION` | 530853 | 3964773039 |
| `Allocated Bytes/sec` | `PERF_COUNTER_BULK_COUNT` | 17877616 |  |
| `Finalization Survivors` | `PERF_COUNTER_LARGE_RAWCOUNT` | 126 |  |
| `Gen 0 heap size` | `PERF_COUNTER_LARGE_RAWCOUNT` | 134217728 |  |
| `Gen 0 Promoted Bytes/Sec` | `PERF_COUNTER_BULK_COUNT` | 10683176 |  |
| `Gen 1 heap size` | `PERF_COUNTER_LARGE_RAWCOUNT` | 16777480 |  |
| `Gen 1 Promoted Bytes/Sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Gen 2 heap size` | `PERF_COUNTER_LARGE_RAWCOUNT` | 24 |  |
| `Large Object Heap size` | `PERF_COUNTER_LARGE_RAWCOUNT` | 1100176 |  |
| `Process ID` | `PERF_COUNTER_LARGE_RAWCOUNT` | 13232 |  |

Not read by the collector:

- `PERF_COUNTER_LARGE_RAWCOUNT`: `# Bytes in all Heaps`, `Promoted Finalization-Memory from Gen 0`, `Promoted Memory from Gen 0`, `Promoted Memory from Gen 1`

### `.NET CLR Remoting`

One instance per process that loaded the CLR, plus `_Global_`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Channels` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Context Proxies` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Context-Bound Classes Loaded` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Context-Bound Objects Alloc / sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Contexts` | `PERF_COUNTER_RAWCOUNT` | 1 |  |
| `Total Remote Calls` | `PERF_COUNTER_RAWCOUNT` | 0 |  |

Not read by the collector:

- `PERF_COUNTER_COUNTER`: `Remote Calls/sec`

### `.NET CLR Security`

One instance per process that loaded the CLR, plus `_Global_`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `# Link Time Checks` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `% Time in RT checks` | `PERF_RAW_FRACTION` | 128 | 325 |
| `Stack Walk Depth` | `PERF_COUNTER_RAWCOUNT` | 1 |  |
| `Total Runtime Checks` | `PERF_COUNTER_RAWCOUNT` | 4288 |  |

Not read by the collector:

- `PERF_COUNTER_TIMER`: `% Time Sig. Authenticating`

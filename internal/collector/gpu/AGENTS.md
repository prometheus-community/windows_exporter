# gpu collector

The collector reads five PDH objects with `pdh.CounterTypeRaw`:
`GPU Engine`, `GPU Adapter Memory`, `GPU Local Adapter Memory`,
`GPU Non Local Adapter Memory` and `GPU Process Memory`. WDDM (dxgkrnl)
provides them, so they exist on any host with a WDDM display driver, including
the Microsoft Basic Render Driver on VMs without a GPU. These notes were
recorded on a Windows 11 workstation with an NVIDIA GeForce GTX 1070
(driver 32.0.15.8266).

- All counters the collector reads are `PERF_COUNTER_LARGE_RAWCOUNT`
  (`NumberOfItems64`), so the raw value is the value:
  - `Running Time` is cumulative engine busy time in 100 ns units. The
    collector divides it by 10,000,000 for `engine_time_seconds`.
  - The memory counters are gauges in bytes.
- `Utilization Percentage` is `PERF_100NSEC_TIMER` and has the same raw value
  as `Running Time`; Task Manager derives the percentage from two samples.
  The collector exports the counter instead, so `rate()` of
  `engine_time_seconds` gives the utilization.
- Instance names encode the labels, see `parseGPUCounterInstanceString`:
  - Engine: `pid_<pid>_luid_0x<high>_0x<low>_phys_<n>_eng_<n>_engtype_<type>`.
    Types seen: `3D`, `Copy`, `LegacyOverlay`, `Security`, `VideoDecode`,
    `VideoEncode`, `VR`. Other vendors append an index (`Compute_0`), which is
    why `engtype` keeps everything after the key.
  - Adapter memory: `luid_..._phys_<n>`. Local and non local adapter memory
    may append `_part_<n>`, but not consistently: on the recording host the
    NVIDIA adapter had a local instance without `part` and a non local
    instance with `part_0`, while the second adapter had only a local
    instance with `part_0`.
  - Process memory: `pid_<pid>_luid_..._phys_<n>`.
- The recording host had two LUIDs: `0x00000000_0x0001120B` (the GTX 1070,
  9 engines per process) and `0x00000000_0x0001241E`, which has no PnP display
  device and 25 engines, all `3D`. The second one is presumably the Microsoft
  Basic Render Driver (inferred, not verified).
- Engine instances are per process and engine: 510 instances across 40
  processes on an ordinary desktop, 48 process memory instances. Every
  process that opens a D3D device gets a full set of engine instances, so the
  engine metric has high cardinality.

## Counter reference

Recorded on Windows 11 Pro (build 26300). `Raw` and `Base` are the first and
second values that `pdh.CounterTypeRaw` returns for the sample instance.

### `GPU Engine`

Multi-instance. 510 instances. Sample instance:
`pid_1064_luid_0x00000000_0x0001120B_phys_0_eng_0_engtype_3D` (`dwm`).

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Running Time` | `PERF_COUNTER_LARGE_RAWCOUNT` | 1660393259 |  |

Not read by the collector:

- `PERF_100NSEC_TIMER`: `Utilization Percentage`

### `GPU Adapter Memory`

Multi-instance. Instances: `luid_0x00000000_0x0001120B_phys_0`,
`luid_0x00000000_0x0001241E_phys_0`. Sample instance:
`luid_0x00000000_0x0001120B_phys_0`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Dedicated Usage` | `PERF_COUNTER_LARGE_RAWCOUNT` | 1641607168 |  |
| `Shared Usage` | `PERF_COUNTER_LARGE_RAWCOUNT` | 87773184 |  |
| `Total Committed` | `PERF_COUNTER_LARGE_RAWCOUNT` | 1962553344 |  |

### `GPU Local Adapter Memory`

Multi-instance. Instances: `luid_0x00000000_0x0001120B_phys_0`,
`luid_0x00000000_0x0001241E_phys_0_part_0`. Sample instance:
`luid_0x00000000_0x0001120B_phys_0`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Local Usage` | `PERF_COUNTER_LARGE_RAWCOUNT` | 903745536 |  |

### `GPU Non Local Adapter Memory`

Multi-instance. Instances: `luid_0x00000000_0x0001120B_phys_0_part_0`.
Sample instance: `luid_0x00000000_0x0001120B_phys_0_part_0`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Non Local Usage` | `PERF_COUNTER_LARGE_RAWCOUNT` | 87773184 |  |

### `GPU Process Memory`

Multi-instance. 48 instances. Sample instance:
`pid_1064_luid_0x00000000_0x0001120B_phys_0` (`dwm`).

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Dedicated Usage` | `PERF_COUNTER_LARGE_RAWCOUNT` | 1126207488 |  |
| `Local Usage` | `PERF_COUNTER_LARGE_RAWCOUNT` | 314675200 |  |
| `Non Local Usage` | `PERF_COUNTER_LARGE_RAWCOUNT` | 4931584 |  |
| `Shared Usage` | `PERF_COUNTER_LARGE_RAWCOUNT` | 4931584 |  |
| `Total Committed` | `PERF_COUNTER_LARGE_RAWCOUNT` | 717426688 |  |

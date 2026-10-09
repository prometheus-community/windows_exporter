# GPU collector

The GPU collector exposes metrics about GPU usage and memory consumption, both at the adapter (physical GPU) and
per-process level.

|                     |                                      |
|---------------------|--------------------------------------|
| Metric name prefix  | `gpu`                                |
| Data source         | Perflib, D3DKMT (dxgkrnl)            |
| Counters            | GPU Engine, GPU Adapter, GPU Process |
| Enabled by default? | No                                   |

## Flags

None

## Metrics

These metrics are available on supported versions of Windows with compatible GPUs and drivers:

### Adapter-level Metrics

| Name                                             | Description                                                                        | Type  | Labels                                                                                                         |
|--------------------------------------------------|------------------------------------------------------------------------------------|-------|----------------------------------------------------------------------------------------------------------------|
| `windows_gpu_info`                               | A metric with a constant '1' value labeled with GPU device information.            | gauge | `architecture`,`bus_number`,`device_id`,`driver_version`,`function_number`,`luid`,`name`,`phys`,`wddm_version` |
| `windows_gpu_dedicated_system_memory_size_bytes` | The size, in bytes, of memory that is dedicated from system memory.                | gauge | `device_id`,`luid`                                                                                             |
| `windows_gpu_dedicated_video_memory_size_bytes`  | The size, in bytes, of memory that is dedicated from video memory.                 | gauge | `device_id`,`luid`                                                                                             |
| `windows_gpu_shared_system_memory_size_bytes`    | The size, in bytes, of memory from system memory that can be shared by many users. | gauge | `device_id`,`luid`                                                                                             |
| `windows_gpu_adapter_memory_committed_bytes`     | Total committed GPU memory in bytes per physical GPU                               | gauge | `device_id`,`luid`,`phys`                                                                                      |
| `windows_gpu_adapter_memory_dedicated_bytes`     | Dedicated GPU memory usage in bytes per physical GPU                               | gauge | `device_id`,`luid`,`phys`                                                                                      |
| `windows_gpu_adapter_memory_shared_bytes`        | Shared GPU memory usage in bytes per physical GPU                                  | gauge | `device_id`,`luid`,`phys`                                                                                      |
| `windows_gpu_local_adapter_memory_bytes`         | Local adapter memory usage in bytes per physical GPU                               | gauge | `device_id`,`luid`,`phys`,`part`                                                                               |
| `windows_gpu_non_local_adapter_memory_bytes`     | Non-local adapter memory usage in bytes per physical GPU                           | gauge | `device_id`,`luid`,`phys`,`part`                                                                               |

The `driver_version`, `wddm_version` and `architecture` labels of `windows_gpu_info` are empty if the driver does not report them.

### Sensor Metrics

The sensor metrics are read from the graphics kernel (dxgkrnl) with `D3DKMTQueryAdapterInfo`, the same source Task Manager
uses. They are vendor-neutral, but depend on WDDM 2.4 or newer and on driver support. Drivers report 0 for values they do
not support, so a metric is only exposed if the driver reported a non-zero value or capability for it at startup. Microsoft
software adapters, like the Microsoft Basic Render Driver, are skipped and never queried. Query failures are logged at debug
level and never fail the scrape.

| Name                                       | Description                                                                           | Type    | Labels                          |
|--------------------------------------------|---------------------------------------------------------------------------------------|---------|---------------------------------|
| `windows_gpu_temperature_celsius`          | Main temperature sensor reading of the physical GPU in degrees Celsius                | gauge   | `device_id`,`luid`,`phys`       |
| `windows_gpu_temperature_warning_celsius`  | Temperature in degrees Celsius at which the physical GPU starts throttling            | gauge   | `device_id`,`luid`,`phys`       |
| `windows_gpu_temperature_max_celsius`      | Maximum temperature in degrees Celsius before the physical GPU takes damage           | gauge   | `device_id`,`luid`,`phys`       |
| `windows_gpu_fan_speed_rpm`                | Current speed of the main fan in revolutions per minute (0 while the fan is stopped)  | gauge   | `device_id`,`luid`,`phys`       |
| `windows_gpu_fan_speed_max_rpm`            | Maximum speed of the main fan in revolutions per minute                               | gauge   | `device_id`,`luid`,`phys`       |
| `windows_gpu_power_usage_ratio`            | Current power draw as a ratio (0–1) of the maximum power (TDP) of the physical GPU    | gauge   | `device_id`,`luid`,`phys`       |
| `windows_gpu_memory_frequency_hertz`       | Current clock frequency of the GPU memory in hertz                                    | gauge   | `device_id`,`luid`,`phys`       |
| `windows_gpu_memory_frequency_max_hertz`   | Maximum clock frequency of the GPU memory in hertz, while not overclocked             | gauge   | `device_id`,`luid`,`phys`       |
| `windows_gpu_engine_frequency_hertz`       | Current clock frequency of the GPU engine in hertz                                    | gauge   | `device_id`,`luid`,`phys`,`eng` |
| `windows_gpu_engine_frequency_max_hertz`   | Maximum clock frequency of the GPU engine in hertz, while not overclocked             | gauge   | `device_id`,`luid`,`phys`,`eng` |

The engine frequency metrics are only exposed for engines that report a maximum frequency, typically the 3D engine. The
`eng` label matches the `eng` label of `windows_gpu_engine_time_seconds`.

### Per-process Metrics

| Name                                         | Description                                     | Type    | Labels                                                    |
|----------------------------------------------|-------------------------------------------------|---------|-----------------------------------------------------------|
| `windows_gpu_engine_time_seconds`            | Total running time of the GPU engine in seconds | counter | `device_id`,`luid`,`phys`, `eng`, `engtype`, `process_id` |
| `windows_gpu_process_memory_committed_bytes` | Total committed GPU memory in bytes per process | gauge   | `device_id`,`luid`,`phys`,`process_id`                    |
| `windows_gpu_process_memory_dedicated_bytes` | Dedicated GPU memory usage in bytes per process | gauge   | `device_id`,`luid`,`phys`,`process_id`                    |
| `windows_gpu_process_memory_local_bytes`     | Local GPU memory usage in bytes per process     | gauge   | `device_id`,`luid`,`phys`,`process_id`                    |
| `windows_gpu_process_memory_non_local_bytes` | Non-local GPU memory usage in bytes per process | gauge   | `device_id`,`luid`,`phys`,`process_id`                    |
| `windows_gpu_process_memory_shared_bytes`    | Shared GPU memory usage in bytes per process    | gauge   | `device_id`,`luid`,`phys`,`process_id`                    |

## Metric Labels

* `luid`,`phys`: Physical GPU index (e.g., "0")
* `eng`: GPU engine index (e.g., "0", "1", ...)
* `engtype`: GPU engine type (e.g., "3D", "Copy", "VideoDecode", etc.)
* `process_id`: Process ID
* `driver_version`: Kernel mode driver version (e.g., "32.0.15.8266")
* `wddm_version`: WDDM version of the driver (e.g., "3.2")
* `architecture`: GPU architecture as reported by the driver (e.g., "Pascal")

## Example Metric

These are basic queries to help you get started with GPU monitoring on Windows using Prometheus.

**Show GPU information for a specific physical GPU (0):**

```promql
windows_gpu_info{bus_number="8",device_id="PCI\\VEN_10DE&DEV_1B81&SUBSYS_61733842&REV_A1",function_number="0",luid="0x00000000_0x00010F8A",name="NVIDIA GeForce GTX 1070",phys="0"} 1
```

**Show total dedicated GPU memory (in bytes) usage on GPU 0:**

```promql
windows_gpu_adapter_memory_dedicated_bytes{phys="0"}
```

**Aggregate GPU utilization across all processes for a physical GPU (3D engine):**

```promql
sum by (phys) (
  rate(windows_gpu_engine_time_seconds{phys="0", engtype="3D"}[1m])
) * 100
```

**Show GPU utilization for a specific process (3D engine):**

```promql
sum by (phys, process_id) (
  rate(windows_gpu_engine_time_seconds{process_id="1234", engtype="3D"}[1m])
) * 100
```

**Show dedicated GPU memory per process:**

```promql
windows_gpu_adapter_memory_dedicated_bytes
```

## Useful Queries

**Show top 5 processes by GPU utilization (all engines):**

```promql
topk(5, sum by (process_id) (
  rate(windows_gpu_engine_time_seconds[1m])
) * 100)
```

**Show GPU memory usage per physical GPU:**

```promql
sum by (phys) (
  windows_gpu_adapter_memory_dedicated_bytes
)
```

Show GPU engine time with process owner and command line:

```promql
windows_gpu_engine_time_seconds * on(process_id) group_left(owner, cmdline) windows_process_info
```

**Show the engine clock as a ratio of its maximum clock:**

```promql
windows_gpu_engine_frequency_hertz / windows_gpu_engine_frequency_max_hertz
```

## Alerting Examples

**prometheus.rules**

```yaml
# Alert on processes using more than 80% of a GPU's capacity over 10 minutes
- alert: HighGpuUtilization
  expr: |
    sum by (process_id) (
      rate(windows_gpu_engine_time_seconds[1m])
    ) * 100 > 80
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "High GPU Utilization (process {{ $labels.process_id }})"
    description: "Process is using more than 80% of GPU resources\n  VALUE = {{ $value }}\n  LABELS: {{ $labels }}"
# Alert on GPUs running within 5 °C of their throttling temperature
- alert: GpuTemperatureHigh
  expr: |
    windows_gpu_temperature_celsius > windows_gpu_temperature_warning_celsius - 5
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "GPU temperature high (luid {{ $labels.luid }})"
    description: "GPU is close to its throttling temperature\n  VALUE = {{ $value }}\n  LABELS: {{ $labels }}"
```

## Notes

* Per-process metrics allow you to identify which processes are consuming GPU resources.
* Adapter-level metrics provide an overview of total GPU memory usage.
* For overall GPU utilization, aggregate per-process metrics in Prometheus using queries such as `sum()`.
* The collector relies on Windows performance counters; ensure your system and drivers support these counters.
* Sensor metrics (temperature, fan, power, clocks) require WDDM 2.4 or newer and a driver that reports these values.

## Enabling the Collector

To enable the GPU collector, add `gpu` to the list of enabled collectors in your windows_exporter configuration.

Example (command line):

```shell
windows_exporter.exe --collectors.enabled=gpu
```

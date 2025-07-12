# gpu collector

The gpu collector exposes metrics about GPU usage and memory consumption, both at the adapter (physical GPU) and
per-process level.

|                     |                                      |
|---------------------|--------------------------------------|
| Metric name prefix  | `gpu`                                |
| Data source         | Perflib                              |
| Counters            | GPU Engine, GPU Adapter, GPU Process |
| Enabled by default? | No                                   |

## Flags

None

## Metrics

These metrics are available on supported versions of Windows with compatible GPUs and drivers:

### Adapter-level Metrics

| Name                                             | Description                                                                        | Type  | Labels        |
|--------------------------------------------------|------------------------------------------------------------------------------------|-------|---------------|
| `windows_gpu_info`                               | A metric with a constant '1' value labeled with gpu device information.            | gauge | `luid`,`name`,`bus_number`,`phys`,`function_number` |
| `windows_gpu_dedicated_system_memory_size_bytes` | The size, in bytes, of memory that is dedicated from system memory.                | gauge | `luid`        |
| `windows_gpu_dedicated_video_memory_size_bytes`  | The size, in bytes, of memory that is dedicated from video memory.                 | gauge | `luid`        |
| `windows_gpu_shared_system_memory_size_bytes`    | The size, in bytes, of memory from system memory that can be shared by many users. | gauge | `luid`        |
| `windows_gpu_adapter_memory_committed_bytes`     | Total committed GPU memory in bytes per physical GPU                               | gauge | `luid`,`phys` |
| `windows_gpu_adapter_memory_dedicated_bytes`     | Dedicated GPU memory usage in bytes per physical GPU                               | gauge | `luid`,`phys` |
| `windows_gpu_adapter_memory_shared_bytes`        | Shared GPU memory usage in bytes per physical GPU                                  | gauge | `luid`,`phys` |
| `windows_gpu_local_adapter_memory_bytes`         | Local adapter memory usage in bytes per physical GPU                               | gauge | `luid`,`phys` |
| `windows_gpu_non_local_adapter_memory_bytes`     | Non-local adapter memory usage in bytes per physical GPU                           | gauge | `luid`,`phys` |

### Per-process Metrics

| Name                                         | Description                                     | Type    | Labels                                        |
|----------------------------------------------|-------------------------------------------------|---------|-----------------------------------------------|
| `windows_gpu_engine_time_seconds`            | Total running time of the GPU engine in seconds | counter | `luid`,`phys`, `eng`, `engtype`, `process_id` |
| `windows_gpu_process_memory_committed_bytes` | Total committed GPU memory in bytes per process | gauge   | `luid`,`phys`,`process_id`                    |
| `windows_gpu_process_memory_dedicated_bytes` | Dedicated GPU memory usage in bytes per process | gauge   | `luid`,`phys`,`process_id`                    |
| `windows_gpu_process_memory_local_bytes`     | Local GPU memory usage in bytes per process     | gauge   | `luid`,`phys`,`process_id`                    |
| `windows_gpu_process_memory_non_local_bytes` | Non-local GPU memory usage in bytes per process | gauge   | `luid`,`phys`,`process_id`                    |
| `windows_gpu_process_memory_shared_bytes`    | Shared GPU memory usage in bytes per process    | gauge   | `luid`,`phys`,`process_id`                    |

## Metric Labels

* `luid`,`phys`: Physical GPU index (e.g., "0")
* `eng`: GPU engine index (e.g., "0", "1", ...)
* `engtype`: GPU engine type (e.g., "3D", "Copy", "VideoDecode", etc.)
* `process_id`: Process ID

## Example Metric

These are basic queries to help you get started with GPU monitoring on Windows using Prometheus.

**Show GPU information for a specific physical GPU (0):**

```promql
windows_gpu_info{description="NVIDIA GeForce GTX 1070",friendly_name="",hardware_id="PCI\\VEN_10DE&DEV_1B81&SUBSYS_61733842&REV_A1",phys="0",physical_device_object_name="\\Device\\NTPNP_PCI0027"} 1
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
```

## Notes

* Per-process metrics allow you to identify which processes are consuming GPU resources.
* Adapter-level metrics provide an overview of total GPU memory usage.
* For overall GPU utilization, aggregate per-process metrics in Prometheus using queries such as `sum()`.
* The collector relies on Windows performance counters; ensure your system and drivers support these counters.

## Enabling the Collector

To enable the GPU collector, add `gpu` to the list of enabled collectors in your windows_exporter configuration.

Example (command line):

```shell
windows_exporter.exe --collectors.enabled=gpu
```

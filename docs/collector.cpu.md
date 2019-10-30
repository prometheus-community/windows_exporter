# cpu collector

The cpu collector exposes metrics about CPU usage

|||
-|-
Metric name prefix  | `cpu`
Data source         | Perflib
Counters            | `ProcessorInformation` (Windows Server 2008R2 and later) `Processor` (older versions)
Enabled by default? | Yes

## Flags

None

## Metrics
These metrics are available on all versions of Windows:

Name | Description | Type | Labels
-----|-------------|------|-------
`wmi_cpu_cstate_seconds_total` | Time spent in low-power idle states | counter | `core`, `state`
`wmi_cpu_time_total` | Time that processor spent in different modes (idle, user, system, ...) | counter | `core`, `mode`
`wmi_cpu_interrupts_total` | Total number of received and serviced hardware interrupts | counter | `core`
`wmi_cpu_dpcs_total` | Total number of received and serviced deferred procedure calls (DPCs) | counter | `core`

These metrics are only exposed on Windows Server 2008R2 and later:

Name | Description | Type | Labels
-----|-------------|------|-------
`wmi_cpu_clock_interrupts_total` | Total number of received and serviced clock tick interrupts | `core`
`wmi_cpu_idle_break_events_total` | Total number of time processor was woken from idle | `core`
`wmi_cpu_parking_status` | Parking Status represents whether a processor is parked or not | `gauge`
`wmi_cpu_core_frequency_mhz` | Core frequency in megahertz | `gauge`
`wmi_cpu_processor_performance` | Processor Performance is the average performance of the processor while it is executing instructions, as a percentage of the nominal performance of the processor. On some processors, Processor Performance may exceed 100% | `gauge`

### Example metric
Show frequency of host CPU cores
```
wmi_cpu_core_frequency_mhz{instance="localhost"}
```

## Useful queries
Show cpu usage by mode.
```
sum by (mode) (irate(wmi_cpu_time_total{instance="localhost"}[5m]))
```

## Alerting examples
**prometheus.rules**
```yaml
# Alert on hosts with more than 80% CPU usage over a 10 minute period
- alert: CpuUsage
  expr: 100 - (avg by (instance) (irate(wmi_cpu_time_total{mode="idle"}[2m])) * 100) > 80
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "CPU Usage (instance {{ $labels.instance }})"
    description: "CPU Usage is more than 80%\n  VALUE = {{ $value }}\n  LABELS: {{ $labels }}"
```

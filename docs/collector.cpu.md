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
`windows_cpu_cstate_seconds_total` | Time spent in low-power idle states | counter | `core`, `state`
`windows_cpu_time_total` | Time that processor spent in different modes (dpc, idle, interrupt, privileged, user) | counter | `core`, `mode`
`windows_cpu_interrupts_total` | Total number of received and serviced hardware interrupts | counter | `core`
`windows_cpu_dpcs_total` | Total number of received and serviced deferred procedure calls (DPCs) | counter | `core`

These metrics are only exposed on Windows Server 2008R2 and later:

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_cpu_clock_interrupts_total` | Total number of received and serviced clock tick interrupts | counter | `core`
`windows_cpu_idle_break_events_total` | Total number of time processor was woken from idle | counter | `core`
`windows_cpu_parking_status` | Parking Status represents whether a processor is parked or not | gauge | `core`
`windows_cpu_core_frequency_mhz` | Core frequency in megahertz | gauge | `core`
`windows_cpu_processor_performance_total` | Processor Performance is the number of CPU cycles executing instructions by each core; it is believed to be similar to the value that the APERF MSR would show, were it exposed | counter | `core`
`windows_cpu_processor_mperf_total` | Processor MPerf Total is proportioanl to the number of TSC ticks each core has accumulated while executing instructions. Due to the manner in which it is presented, it should be scaled by 1e2 to properly line up with Processor Performance Total. As above, it is believed to be closely related to the MPERF MSR. | counter | `core`
`windows_cpu_processor_rtc_total` | RTC total is assumed to represent the 64Hz tick rate in Windows. It is not by itself useful, but can be used with `windows_cpu_processor_utility_total` to more accurately measure CPU utilisation than with `windows_cpu_time_total` | counter | `core`
`windows_cpu_processor_utility_total` | Processor Utility Total is a newer, more accurate measure of CPU utilization, in particular handling modern CPUs with variant CPU frequencies. The rate of this counter divided by the rate of `windows_cpu_processor_rtc_total` should provide an accurate view of CPU utilisation on modern systems, as observed in Task Manager. | counter | `core`
`windows_cpu_processor_privileged_utility_total` | Processor Privilged Utility Total, when used in a similar fashion to `windows_cpu_processor_utility_total` will show the portion of CPU utilization which is happening in privileged mode. | counter | `core`

### Example metric
Show frequency of host CPU cores
```
windows_cpu_core_frequency_mhz{instance="localhost"}
```

## Useful queries
Show cpu usage by mode.
```
sum by (mode) (irate(windows_cpu_time_total{instance="localhost"}[5m]))
```
Show per-cpu utilisation using the processor utility metrics
```
rate(windows_cpu_processor_utility_total{instance="localhost"}[5m]) / rate(windows_cpu_processor_rtc_total{instance="localhost"}[5m])
```
Show actual average CPU frequency in Hz
```
avg by(instance) (
    1e4 * windows_cpu_core_frequency_mhz{}
    * rate(windows_cpu_processor_performance_total{}[5m])
    / rate(windows_cpu_processor_mperf_total{}[5m])
)
```


## Alerting examples
**prometheus.rules**
```yaml
# Alert on hosts with more than 80% CPU usage over a 10 minute period
- alert: CpuUsage
  expr: 100 - (avg by (instance) (irate(windows_cpu_time_total{mode="idle"}[2m])) * 100) > 80
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "CPU Usage (instance {{ $labels.instance }})"
    description: "CPU Usage is more than 80%\n  VALUE = {{ $value }}\n  LABELS: {{ $labels }}"
# Alert on hosts which are not boosting their CPU frequencies
- alert: NoCpuTurbo
  expr: |
    avg by(instance) (
        1e4 * windows_cpu_core_frequency_mhz{}
        * rate(windows_cpu_processor_performance_total{}[5m])
        / rate(windows_cpu_processor_mperf_total{}[5m])
    )
    /
    (1e6 * avg by (instance) (windows_cpu_core_frequency_mhz))
    < 1.1
  for: 1h
  annotations:
    summary: "CPU Frequency on {{ $labels.instance }} is less than 110% of base frequency, suggesting it is not able to boost.
```

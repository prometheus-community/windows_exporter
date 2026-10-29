# time collector

The time collector exposes the Windows Time Service and other time related metrics.
If the Windows Time Service is stopped after collection has started, collector metric values will reset to 0.

Please note the Time Service perflib counters are only available on [Windows Server 2016 or newer](https://docs.microsoft.com/en-us/windows-server/networking/windows-time-service/windows-server-2016-improvements).

|                     |        |
|---------------------|--------|
| Metric name prefix  | `time` |
| Data source         | PDH    |
| Enabled by default? | No     |

## Flags

### `--collector.time.enabled`
Comma-separated list of collectors to use. Supported values are `system_time`, `clock_source`, `ntp` and `sync_status`. Defaults to all, if not specified.
For example: `--collector.time.enabled=ntp,system_time`.
Matching is case-sensitive.



## Metrics

| Name                                               | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | Type    | Labels     |
|----------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|------------|
| `windows_time_clock_frequency_adjustment`          | Adjustment made to the local system clock frequency by W32Time in nominal clock units. This counter helps visualize the finer adjustments being made by W32Time to synchronize the local clock.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       | gauge   | None       |
| `windows_time_clock_frequency_adjustment_ratio`    | Adjustment made to the local system clock frequency by W32Time, as a ratio. Windows reports it in parts per billion (PPB); 1 PPB (a ratio of 1e-9) implies the system clock was adjusted at a rate of 1 nanosecond per second (1 ns/s). The smallest possible adjustment can vary and is expected to be in the order of 100's of PPB.                                                                                                                                                                                                                                                                                                                                                                                                                  | gauge   | None       |
| `windows_time_computed_time_offset_seconds`        | The absolute time offset between the system clock and the chosen time source, as computed by the W32Time service, in seconds. When a new valid sample is available, the computed time is updated with the time offset indicated by the sample. This time is the actual time offset of the local clock. W32Time initiates clock correction by using this offset and updates the computed time in between samples with the remaining time offset that needs to be applied to the local clock. Clock accuracy can be tracked by using this performance counter with a low polling interval (for example, 256 seconds or less) and looking for the counter value to be smaller than the desired clock accuracy limit.     | gauge   | None       |
| `windows_time_ntp_client_time_sources`             | Active number of NTP Time sources being used by the client. This is a count of active, distinct IP addresses of time servers that are responding to this client's requests.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | gauge   | None       |
| `windows_time_ntp_round_trip_delay_seconds`        | Total roundtrip delay experienced by the NTP client in receiving a response from the server for the most recent request, in seconds. This is the time elapsed on the NTP client between transmitting a request to the NTP server and receiving a valid response from the server.                                                                                                                                                                                                                                                                                                                                                                                                                                      | gauge   | None       |
| `windows_time_ntp_server_outgoing_responses_total` | Total number of requests responded to by the NTP server.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              | counter | None       |
| `windows_time_ntp_server_incoming_requests_total`  | Total number of requests received by the NTP server.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  | counter | None       |
| `windows_time_current_timestamp_seconds`           | Current time as reported by the operating system, in [Unix time](https://en.wikipedia.org/wiki/Unix_time). See [time.UnixMicro()](https://golang.org/pkg/time/#UnixMicro) for details                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | gauge   | None       |
| `windows_time_timezone`                            | Current timezone as reported by the operating system.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 | gauge   | `timezone` |
| `windows_time_clock_sync_source`                   | This value reflects the sync source of the system clock.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              | gauge   | `type`     |
| `windows_time_sync_phase_offset_seconds` | Signed time difference between the local clock and the time source of W32Time, in seconds. Positive values mean the local clock is ahead. Unlike `windows_time_computed_time_offset_seconds`, it keeps the sign. Collector `sync_status`. | gauge | None |
| `windows_time_sync_stratum` | Stratum of the local clock of W32Time. 0 means unspecified, 16 unsynchronized. Collector `sync_status`. | gauge | None |
| `windows_time_sync_leap_indicator` | NTP leap indicator of W32Time: 0 no warning, 1 last minute has 61 seconds, 2 last minute has 59 seconds, 3 not synchronized. Collector `sync_status`. | gauge | None |
| `windows_time_sync_poll_interval_seconds` | Current poll interval of W32Time, in seconds. Collector `sync_status`. | gauge | None |
| `windows_time_sync_root_delay_seconds` | Total round-trip delay of W32Time to the primary time source, in seconds. Collector `sync_status`. | gauge | None |
| `windows_time_sync_root_dispersion_seconds` | Root dispersion of W32Time: the maximum error relative to the primary time source, in seconds. Collector `sync_status`. | gauge | None |
| `windows_time_sync_state` | State of the W32Time clock discipline (`unset`, `hold`, `sync`, `spike`), 1 if the current state, 0 otherwise. Collector `sync_status`. | gauge | `state` |
| `windows_time_sync_last_result` | Result of the last synchronization attempt of W32Time (`success`, `no_data`, `stale_data`, `change_too_big`, `shutdown`), 1 if the current result, 0 otherwise. Collector `sync_status`. | gauge | `result` |
| `windows_time_sync_last_success_timestamp_seconds` | Time of the last successful synchronization of W32Time, in unix time. Omitted until the first synchronization. Collector `sync_status`. | gauge | None |
| `windows_time_sync_source_info` | Time source that W32Time synchronizes with, as `w32tm /query /status` shows it. The value is always 1. Collector `sync_status`. | gauge | `source` |

### Example metric
```
# HELP windows_time_clock_sync_source This value reflects the sync source of the system clock.
# TYPE windows_time_clock_sync_source gauge
windows_time_clock_sync_source{type="AllSync"} 0
windows_time_clock_sync_source{type="Local CMOS Clock"} 0
windows_time_clock_sync_source{type="NT5DS"} 0
windows_time_clock_sync_source{type="NTP"} 1
windows_time_clock_sync_source{type="NoSync"} 0
# HELP windows_time_current_timestamp_seconds OperatingSystem.LocalDateTime
# TYPE windows_time_current_timestamp_seconds gauge
windows_time_current_timestamp_seconds 1.74862554e+09
```

The `sync_status` metrics come from the same status as `w32tm /query /status`.
They need the Windows Time service, but no administrative rights.

## Useful queries
Seconds since the last successful synchronization:
```
time() - windows_time_sync_last_success_timestamp_seconds
```

## Alerting examples
**prometheus.rules**
```yaml
# Alert on hosts with an NTP client delay of more than 1 second, for a 5 minute period or longer.
- alert: NTPClientDelay
  expr: windows_time_ntp_round_trip_delay_seconds > 1
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "NTP client delay: (instance {{ $labels.instance }})"
    description: "RTT for NTP client is greater than 1 second!\nVALUE = {{ $value }}sec\n  LABELS: {{ $labels }}"

# Alert on hosts whose clock was not synchronized for more than 6 hours.
- alert: TimeNotSynchronized
  expr: time() - windows_time_sync_last_success_timestamp_seconds > 6 * 3600
  for: 15m
  labels:
    severity: warning
  annotations:
    summary: "Clock not synchronized: (instance {{ $labels.instance }})"
    description: "W32Time last synchronized {{ $value | humanizeDuration }} ago.\n  LABELS: {{ $labels }}"
```

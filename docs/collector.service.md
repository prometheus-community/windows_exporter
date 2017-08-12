# Service collector

The service collector exposes metrics about Windows service startup behaviour and state.

Collector namespace | `service`
Classes             | [`Win32_Service`](https://msdn.microsoft.com/en-us/library/aa394418(v=vs.85).aspx)
Enabled by default? | Yes

## Flags

### `--collector.service.services-where`
A WQL `WHERE` clause for filtering the services returned.

Example: `--collector.service.services-where "StartMode != 'Disabled'"

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`wmi_service_state` | State of the named service, 1 if in the given state | gauge | name, state |
`wmi_service_start_mode` | Start mode of the named service, 1 if in the given mode | gauge | name, start_mode |

### Example

The IIS service starts automatically:

`wmi_service_start_mode{name="w3svc", start_mode="auto"} 1`

## Useful queries

## Alerting examples
### Ensure IIS is running
```
ALERT IISIsNotRunning
  IF wmi_service_state{name="w3svc", state="running"} < 1
  FOR 5m
  LABELS {
    urgency = "immediate"
  }
  ANNOTATIONS {
    description = "The IIS service on {{ $labels.instance }} is not running."
  }
```

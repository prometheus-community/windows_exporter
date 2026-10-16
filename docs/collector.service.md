# service collector

The service collector exposes metrics about Windows Services

|                     |           |
|---------------------|-----------|
| Metric name prefix  | `service` |
| Classes             | none      |
| Enabled by default? | Yes       |

## Flags

### `--collector.service.include`

Regular expression of service to include. Service name (not the display name!) must both
match `include` and not match `exclude` to be included.
Recommended to keep down number of returned metrics. Default: `.+`

### `--collector.service.exclude`

Regular expression of service to exclude. Service name (not the display name!) must both
match `include` and not match `exclude` to be included.
Recommended to keep down number of returned metrics. Default: empty

### `--collector.service.start-mode-include`

Comma separated list of service start modes to include.
Possible values: auto, boot, manual, disabled, system.
Recommended to keep down number of returned metrics. Default: `auto,boot,manual,disabled,system`

## Configuration cache

The service configuration (display name, start mode, delayed start, triggers, path and
account, used by `windows_service_info` and `windows_service_start_mode`) is cached between scrapes.
The collector watches the registry key `HKLM\SYSTEM\CurrentControlSet\Services`. When a
service's own key changes, for example after `sc config`, a change in the Services console
or a reinstall, its configuration is queried again on the next scrape. All configurations
are queried again at the latest every 10 minutes. The service state and process are queried
on every scrape. Reading the delayed start flag and the triggers takes extra calls per service,
which run when a configuration is queried: on a cache miss, or on every scrape if the registry
key can't be watched and the cache is disabled.

## Metrics

| Name                         | Description                                                                                   | Type  | Labels                                                                   |
|------------------------------|-----------------------------------------------------------------------------------------------|-------|--------------------------------------------------------------------------|
| `windows_service_info`       | Contains service information run as user in labels, constant 1                                | gauge | service, display_name, run_as, path_name, delayed_start, trigger_start |
| `windows_service_start_mode` | The start mode of the service, 1 if the current start mode, 0 otherwise                       | gauge | service, start_mode                                                      |
| `windows_service_state`      | The state of the service, 1 if the current state, 0 otherwise                                 | gauge | service, state                                                           |
| `windows_service_process`    | Process of started service. The value is the creation time of the process as a Unix timestamp | gauge | service, process_id                                                      |

### States

A service can be in the following states:
- `stopped`
- `start pending`
- `stop pending`
- `running`
- `continue pending`
- `pause pending`
- `paused`

### Start modes

A service can have the following start modes:
- `boot`
- `system`
- `auto`
- `manual`
- `disabled`

Note that there is some overlap with service state.

### Delayed and trigger start

The Services console shows "Automatic (Delayed Start)" and "(Trigger Start)" in addition to
the start mode. `windows_service_start_mode` reports these services with their start mode
(`auto` or `manual`); the labels of `windows_service_info` tell them apart:

- `delayed_start` is `true` for an `auto` service that starts after the other automatic
  services, `false` otherwise. Windows ignores the delayed start setting of services that
  don't start automatically, so it is always `false` for them.
- `trigger_start` is `true` if the service has triggers, which start or stop it when an event
  occurs, for example when a device arrives, the computer joins a domain or a network becomes
  available. Such a service can stop by design while its start mode is `auto`.

Both labels are empty if the service configuration, or the delayed start flag and triggers,
can't be read. The other labels are still published in the second case.

### Run As

Account name under which a service runs. Depending on the service type, the account name may be in the form of "DomainName\Username" or UPN format ("Username@DomainName").

### Example metric

```
# HELP windows_service_info A metric with a constant '1' value labeled with service information
# TYPE windows_service_info gauge
windows_service_info{delayed_start="false",display_name="Declared Configuration(DC) service",service="dcsvc",path_name="C:\\WINDOWS\\system32\\svchost.exe -k netsvcs -p",run_as="LocalSystem",trigger_start="true"} 1
windows_service_info{delayed_start="false",display_name="Designs",service="Themes",path_name="C:\\WINDOWS\\System32\\svchost.exe -k netsvcs -p",run_as="LocalSystem",trigger_start="false"} 1
# HELP windows_service_process Process of started service. The value is the creation time of the process as a unix timestamp.
# TYPE windows_service_process gauge
windows_service_process{service="Themes",process_id="2856"} 1.7244891e+09
# HELP windows_service_start_mode The start mode of the service (StartMode)
# TYPE windows_service_start_mode gauge
windows_service_start_mode{service="Themes",start_mode="auto"} 1
windows_service_start_mode{service="Themes",start_mode="boot"} 0
windows_service_start_mode{service="Themes",start_mode="disabled"} 0
windows_service_start_mode{service="Themes",start_mode="manual"} 0
windows_service_start_mode{service="Themes",start_mode="system"} 0
windows_service_start_mode{service="dcsvc",start_mode="auto"} 0
windows_service_start_mode{service="dcsvc",start_mode="boot"} 0
windows_service_start_mode{service="dcsvc",start_mode="disabled"} 0
windows_service_start_mode{service="dcsvc",start_mode="manual"} 1
windows_service_start_mode{service="dcsvc",start_mode="system"} 0
# HELP windows_service_state The state of the service (State)
# TYPE windows_service_state gauge
windows_service_state{service="Themes",state="continue pending"} 0
windows_service_state{service="Themes",state="pause pending"} 0
windows_service_state{service="Themes",state="paused"} 0
windows_service_state{service="Themes",state="running"} 1
windows_service_state{service="Themes",state="start pending"} 0
windows_service_state{service="Themes",state="stop pending"} 0
windows_service_state{service="Themes",state="stopped"} 0
windows_service_state{service="dcsvc",state="continue pending"} 0
windows_service_state{service="dcsvc",state="pause pending"} 0
windows_service_state{service="dcsvc",state="paused"} 0
windows_service_state{service="dcsvc",state="running"} 0
windows_service_state{service="dcsvc",state="start pending"} 0
windows_service_state{service="dcsvc",state="stop pending"} 0
windows_service_state{service="dcsvc",state="stopped"} 1
```

## Useful queries
Counts the number of Microsoft SQL Server/Agent Processes

```
count(windows_service_state{service=~"(sqlserveragent|mssqlserver)",state="running"})
```

## Alerting examples
**prometheus.rules**
```yaml
groups:
- name: Microsoft SQL Server Alerts
  rules:

  # Sends an alert when the 'sqlserveragent' service is not in the running state for 3 minutes.
  - alert: SQL Server Agent DOWN
    expr: windows_service_state{instance="SQL",service="sqlserveragent",state="running"} == 0
    for: 3m
    labels:
      severity: high
    annotations:
      summary: "Service {{ $labels.service }} down"
      description: "Service {{ $labels.service }} on instance {{ $labels.instance }} has been down for more than 3 minutes."

  # Sends an alert when the 'mssqlserver' service is not in the running state for 3 minutes.
  - alert: SQL Server DOWN
    expr: windows_service_state{instance="SQL",service="mssqlserver",state="running"} == 0
    for: 3m
    labels:
      severity: high
    annotations:
      summary: "Service {{ $labels.service }} down"
      description: "Service {{ $labels.service }} on instance {{ $labels.instance }} has been down for more than 3 minutes."
```
In this example, `instance` is the target label of the host. So each alert will be processed per host, which is then used in the alert description.

The next rule alerts on every automatic service that is not running, except trigger-start
services, which stop by design:

```yaml
groups:
- name: Windows services
  rules:

  # Sends an alert when a service with start mode auto and without triggers is not running for 15 minutes.
  - alert: WindowsAutoServiceNotRunning
    expr: |
      windows_service_state{state="running"} == 0
      and on (instance, service) windows_service_start_mode{start_mode="auto"} == 1
      unless on (instance, service) windows_service_info{trigger_start="true"}
    for: 15m
    labels:
      severity: warning
    annotations:
      summary: "Service {{ $labels.service }} not running"
      description: "Automatic service {{ $labels.service }} on instance {{ $labels.instance }} has not been running for more than 15 minutes."
```

Delayed-start services start about two minutes after the other automatic services, so the
`for` duration covers them after a reboot. To leave them out as well, add
`unless on (instance, service) windows_service_info{delayed_start="true"}`.

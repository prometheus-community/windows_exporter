# service collector

The service collector exposes metrics about Windows Services

|||
-|-
Metric name prefix  | `service`
Classes             | none
Enabled by default? | Yes

## Flags

None

## Metrics

| Name                         | Description                                                                                   | Type  | Labels                                |
|------------------------------|-----------------------------------------------------------------------------------------------|-------|---------------------------------------|
| `windows_service_info`       | Contains service information run as user in labels, constant 1                                | gauge | name, display_name, path_name, run_as |
| `windows_service_start_mode` | The start mode of the service, 1 if the current start mode, 0 otherwise                       | gauge | name, start_mode                      |
| `windows_service_state`      | The state of the service, 1 if the current state, 0 otherwise                                 | gauge | name, state                           |
| `windows_service_process`    | Process of started service. The value is the creation time of the process as a unix timestamp | gauge | name, process_id                      |

### States

A service can be in the following states:
- `stopped`
- `start pending`
- `stop pending`
- `running`
- `continue pending`
- `pause pending`
- `paused`
- `unknown`

### Start modes

A service can have the following start modes:
- `boot`
- `system`
- `auto`
- `manual`
- `disabled`

Note that there is some overlap with service state.

### Run As

Account name under which a service runs. Depending on the service type, the account name may be in the form of "DomainName\Username" or UPN format ("Username@DomainName").

### Example metric

```
# HELP windows_service_info A metric with a constant '1' value labeled with service information
# TYPE windows_service_info gauge
windows_service_info{display_name="Declared Configuration(DC) service",name="dcsvc",path_name="C:\\WINDOWS\\system32\\svchost.exe -k netsvcs -p",run_as="LocalSystem"} 1
windows_service_info{display_name="Designs",name="Themes",path_name="C:\\WINDOWS\\System32\\svchost.exe -k netsvcs -p",run_as="LocalSystem"} 1
# HELP windows_service_process Process of started service. The value is the creation time of the process as a unix timestamp.
# TYPE windows_service_process gauge
windows_service_process{name="Themes",process_id="2856"} 1.7244891e+09
# HELP windows_service_start_mode The start mode of the service (StartMode)
# TYPE windows_service_start_mode gauge
windows_service_start_mode{name="Themes",start_mode="auto"} 1
windows_service_start_mode{name="Themes",start_mode="boot"} 0
windows_service_start_mode{name="Themes",start_mode="disabled"} 0
windows_service_start_mode{name="Themes",start_mode="manual"} 0
windows_service_start_mode{name="Themes",start_mode="system"} 0
windows_service_start_mode{name="dcsvc",start_mode="auto"} 0
windows_service_start_mode{name="dcsvc",start_mode="boot"} 0
windows_service_start_mode{name="dcsvc",start_mode="disabled"} 0
windows_service_start_mode{name="dcsvc",start_mode="manual"} 1
windows_service_start_mode{name="dcsvc",start_mode="system"} 0
# HELP windows_service_state The state of the service (State)
# TYPE windows_service_state gauge
windows_service_state{name="Themes",state="continue pending"} 0
windows_service_state{name="Themes",state="pause pending"} 0
windows_service_state{name="Themes",state="paused"} 0
windows_service_state{name="Themes",state="running"} 1
windows_service_state{name="Themes",state="start pending"} 0
windows_service_state{name="Themes",state="stop pending"} 0
windows_service_state{name="Themes",state="stopped"} 0
windows_service_state{name="dcsvc",state="continue pending"} 0
windows_service_state{name="dcsvc",state="pause pending"} 0
windows_service_state{name="dcsvc",state="paused"} 0
windows_service_state{name="dcsvc",state="running"} 0
windows_service_state{name="dcsvc",state="start pending"} 0
windows_service_state{name="dcsvc",state="stop pending"} 0
windows_service_state{name="dcsvc",state="stopped"} 1
```

## Useful queries
Counts the number of Microsoft SQL Server/Agent Processes

```
count(windows_service_state{name=~"(sqlserveragent|mssqlserver)",state="running"})
```

## Alerting examples
**prometheus.rules**
```yaml
groups:
- name: Microsoft SQL Server Alerts
  rules:

  # Sends an alert when the 'sqlserveragent' service is not in the running state for 3 minutes.
  - alert: SQL Server Agent DOWN
    expr: windows_service_state{instance="SQL",name="sqlserveragent",state="running"} == 0
    for: 3m
    labels:
      severity: high
    annotations:
      summary: "Service {{ $labels.name }} down"
      description: "Service {{ $labels.name }} on instance {{ $labels.instance }} has been down for more than 3 minutes."

  # Sends an alert when the 'mssqlserver' service is not in the running state for 3 minutes.
  - alert: SQL Server DOWN
    expr: windows_service_state{instance="SQL",name="mssqlserver",state="running"} == 0
    for: 3m
    labels:
      severity: high
    annotations:
      summary: "Service {{ $labels.name }} down"
      description: "Service {{ $labels.name }} on instance {{ $labels.instance }} has been down for more than 3 minutes."
```
In this example, `instance` is the target label of the host. So each alert will be processed per host, which is then used in the alert description.

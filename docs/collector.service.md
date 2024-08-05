# service collector

The service collector exposes metrics about Windows Services

|||
-|-
Metric name prefix  | `service`
Classes             | none
Enabled by default? | No

## Flags

None

## Metrics

| Name                         | Description                                                          | Type  | Labels                    |
|------------------------------|----------------------------------------------------------------------|-------|---------------------------|
| `windows_service_state`      | The state of the service, 1 if the current state, 0 otherwise        | gauge | name, display_name, state |
| `windows_service_process_id` | Contains the process_id of the service process in labels, constant 1 | gauge | name, process_id          |

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

### Example metric

```
windows_service_state{display_name="Declared Configuration(DC) service",name="dcsvc",status="continue pending"} 0
windows_service_state{display_name="Declared Configuration(DC) service",name="dcsvc",status="pause pending"} 0
windows_service_state{display_name="Declared Configuration(DC) service",name="dcsvc",status="paused"} 0
windows_service_state{display_name="Declared Configuration(DC) service",name="dcsvc",status="running"} 0
windows_service_state{display_name="Declared Configuration(DC) service",name="dcsvc",status="start pending"} 0
windows_service_state{display_name="Declared Configuration(DC) service",name="dcsvc",status="stop pending"} 0
windows_service_state{display_name="Declared Configuration(DC) service",name="dcsvc",status="stopped"} 1
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
      summary: "Service {{ $labels.exported_name }} down"
      description: "Service {{ $labels.exported_name }} on instance {{ $labels.instance }} has been down for more than 3 minutes."

  # Sends an alert when the 'mssqlserver' service is not in the running state for 3 minutes.
  - alert: SQL Server DOWN
    expr: windows_service_state{instance="SQL",name="mssqlserver",state="running"} == 0
    for: 3m
    labels:
      severity: high
    annotations:
      summary: "Service {{ $labels.exported_name }} down"
      description: "Service {{ $labels.exported_name }} on instance {{ $labels.instance }} has been down for more than 3 minutes."
```
In this example, `instance` is the target label of the host. So each alert will be processed per host, which is then used in the alert description.

# service collector

The service collector exposes metrics about Windows Services

|||
-|-
Metric name prefix  | `service`
Classes             | [`Win32_Service`](https://msdn.microsoft.com/en-us/library/aa394418(v=vs.85).aspx)
Enabled by default? | Yes

## Flags

### `--collector.service.services-where`

A WMI filter on which services to include. Recommended to keep down number of returned metrics.

Example: `--collector.service.services-where="Name='wmi_exporter'"`

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`wmi_service_state` | The state of the service, 1 if the current state, 0 otherwise | gauge | name, state
`wmi_service_start_mode` | The start mode of the service, 1 if the current start mode, 0 otherwise | gauge | name, start_mode
`wmi_service_status` | The status of the service, 1 if the current status, 0 otherwise | gauge | name, status

For the values of the `state`, `start_mode` and `status` labels, see below.

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

### Status

A service can have any of the following statuses:
- `ok`
- `error`
- `degraded`
- `unknown`
- `pred fail`
- `starting`
- `stopping`
- `service`
- `stressed`
- `nonrecover`
- `no contact`
- `lost comm`

Note that there is some overlap with service state.

### Example metric
Lists the services that have a 'disabled' start mode.
```
wmi_service_start_mode{exported_name=~"(mssqlserver|sqlserveragent)",start_mode="disabled"}
```

## Useful queries
Counts the number of Microsoft SQL Server/Agent Processes
```
count(wmi_service_state{exported_name=~"(sqlserveragent|mssqlserver)",state="running"})
```

## Alerting examples
**prometheus.rules**
```yaml
groups:
- name: Microsoft SQL Server Alerts
  rules:

  # Sends an alert when the 'sqlserveragent' service is not in the running state for 3 minutes. 
  - alert: SQL Server Agent DOWN
    expr: wmi_service_state{instance="SQL",exported_name="sqlserveragent",state="running"} == 0
    for: 3m
    labels:
      severity: high
    annotations:
      summary: "Service {{ $labels.exported_name }} down"
      description: "Service {{ $labels.exported_name }} on instance {{ $labels.instance }} has been down for more than 3 minutes."
      
  # Sends an alert when the 'mssqlserver' service is not in the running state for 3 minutes. 
  - alert: SQL Server DOWN
    expr: wmi_service_state{instance="SQL",exported_name="mssqlserver",state="running"} == 0
    for: 3m
    labels:
      severity: high
    annotations:
      summary: "Service {{ $labels.exported_name }} down"
      description: "Service {{ $labels.exported_name }} on instance {{ $labels.instance }} has been down for more than 3 minutes."
```
In this example, `instance` is the target label of the host. So each alert will be processed per host, which is then used in the alert description.

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

Example: `--collector.service.services-where="Name='windows_exporter'"`

Example config win_exporter.yml for multiple services: `services-where: Name='SQLServer' OR Name='Couchbase' OR Name='Spooler' OR Name='ActiveMQ'`

### `--collector.service.use-api`

Uses API calls instead of WMI for performance optimization. **Note** the previous flag (`--collector.service.services-where`) won't have any effect on this mode.

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_service_info` | Contains service information in labels, constant 1 | gauge | name, display_name, process_id, run_as
`windows_service_state` | The state of the service, 1 if the current state, 0 otherwise | gauge | name, state
`windows_service_start_mode` | The start mode of the service, 1 if the current start mode, 0 otherwise | gauge | name, start_mode
`windows_service_status` | The status of the service, 1 if the current status, 0 otherwise | gauge | name, status

For the values of the `state`, `start_mode`, `status` and `run_as` labels, see below.

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

### Status (not available in API mode)

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

### Run As

Account name under which a service runs. Depending on the service type, the account name may be in the form of "DomainName\Username" or UPN format ("Username@DomainName").

It corresponds to the `StartName` attribute of the `Win32_Service` class.
`StartName` attribute can be NULL and in such case the label is reported as an empty string. Notice that if the attribute is NULL the service is logged on as the `LocalSystem` account or, for kernel or system-level drive, it runs with a default object name created by the I/O system based on the service name, for example, DWDOM\Admin.

### Example metric
Lists the services that have a 'disabled' start mode.
```
windows_service_start_mode{exported_name=~"(mssqlserver|sqlserveragent)",start_mode="disabled"}
```

## Useful queries
Counts the number of Microsoft SQL Server/Agent Processes
```
count(windows_service_state{exported_name=~"(sqlserveragent|mssqlserver)",state="running"})
```

## Alerting examples
**prometheus.rules**
```yaml
groups:
- name: Microsoft SQL Server Alerts
  rules:

  # Sends an alert when the 'sqlserveragent' service is not in the running state for 3 minutes.
  - alert: SQL Server Agent DOWN
    expr: windows_service_state{instance="SQL",exported_name="sqlserveragent",state="running"} == 0
    for: 3m
    labels:
      severity: high
    annotations:
      summary: "Service {{ $labels.exported_name }} down"
      description: "Service {{ $labels.exported_name }} on instance {{ $labels.instance }} has been down for more than 3 minutes."

  # Sends an alert when the 'mssqlserver' service is not in the running state for 3 minutes.
  - alert: SQL Server DOWN
    expr: windows_service_state{instance="SQL",exported_name="mssqlserver",state="running"} == 0
    for: 3m
    labels:
      severity: high
    annotations:
      summary: "Service {{ $labels.exported_name }} down"
      description: "Service {{ $labels.exported_name }} on instance {{ $labels.instance }} has been down for more than 3 minutes."
```
In this example, `instance` is the target label of the host. So each alert will be processed per host, which is then used in the alert description.

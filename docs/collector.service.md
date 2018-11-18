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
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

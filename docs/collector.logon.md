# logon collector

The logon collector exposes metrics detailing the active user logon sessions.

|||
-|-
Metric name prefix  | `logon`
Classes             | [`Win32_LogonSession`](https://docs.microsoft.com/en-us/windows/win32/cimwin32prov/win32-logonsession)
Enabled by default? | No

> :warning: **On some deployments, this collector seems to have some memory/timeout issues**: See [#583](https://github.com/prometheus-community/windows_exporter/issues/583)

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_logon_logon_type` | Number of active user logon sessions | gauge | status

### Example metric
Query the total number of interactive logon sessions
```
windows_logon_logon_type{status="interactive"}
```

## Useful queries
Query the total number of local and remote (I.E. Terminal Services) interactive sessions.
```
windows_logon_logon_type{status=~"interactive|remote_interactive"}
```

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

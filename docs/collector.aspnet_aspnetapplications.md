# aspnet_aspnetapplications collector

The aspnet_aspnetapplications collector exposes metrics about ASP.NET Applications.

|||
-|-
Metric name prefix  | `aspnet_aspnetapplications`
Classes             | `Win32_PerfRawData_ASPNET_ASPNETApplications"`
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`wmi_...` | ... | counter/gauge/histogram/summary | ...
`wmi_aspnet_aspnetapplications_percent_managed_processor_timeestimated ` | This allows you to get the estimated CPU time for your specific application, and not the entire IIS application pool which could be multiple applications. | gauge | `process`

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

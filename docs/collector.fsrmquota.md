# Microsoft File Server Resource Manager (FSRM) Quotas collector

The fsrmquota collector exposes metrics about File Server Resource Manager Quotas. Note that this collector has only been tested against Windows server 2012R2.
Other FSRM versions may work but are not tested.

|||
-|-
Metric name prefix  | `fsrmquota`
Data source         | wmi
Counters            | `FSRMQUOTA`
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_fsrmquota_count` | Number of Quotas | counter |None
`windows_fsrmquota_description` | A string up to 1KB in size. Optional. The default value is an empty string. (Description) | counter |`path`, `template`,`description`
`windows_fsrmquota_disabled` | If 1, the quota is disabled. The default value is 0. (Disabled) | counter |`path`, `template`
`windows_fsrmquota_matchestemplate` | If 1, the property values of this quota match those values of the template from which it was derived. (MatchesTemplate) | counter |`path`, `template`
`windows_fsrmquota_peak_usage_bytes ` | The highest amount of disk space usage charged to this quota. (PeakUsage) | counter |`path`, `template`
`windows_fsrmquota_size_bytes` | The size of the quota. If the Template property is not provided then the Size property must be provided (Size) | counter |`path`, `template`
`windows_fsrmquota_softlimit` | If 1, the quota is a soft limit. If 0, the quota is a hard limit. The default value is 0. Optional (SoftLimit) | counter |`path`, `template`
`windows_fsrmquota_template` | A valid quota template name. Up to 1KB in size. Optional (Template) | counter |`path`, `template`
`windows_fsrmquota_usage_bytes` | The current amount of disk space usage charged to this quota. (Usage) | counter |`path`, `template`


### Example metric
Show rate of Quotas usage:
```
rate(windows_fsrmquota_usage_bytes)[1d]
```

## Useful queries

## Alerting examples
**prometheus.rules**
```yaml
  - alert: "HighQuotasUsage"
    expr: "windows_fsrmquota_usage_bytes{instance="SERVER1.COM:9182"} / windows_fsrmquota_size{instance="SERVER1.COM:9182"} >0.85"
    for: "10m"
    labels:
      severity: "high"
    annotations:
      summary: "High Quotas Usage"
      description: "High use of File Resource.\n Quotas: {{ $labels.path }}\n Current use : {{ $value }}"
```

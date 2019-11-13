# Microsoft File Server Resource Manager (FSRM) Quotas collector

The msft_fsrmquota collector exposes metrics about File Server Ressource Manager Quotas. Note that this collector has only been tested against Windows server 2012R2.
Other FSRM versions may work but are not tested.

|||
-|-
Metric name prefix  | `msft_fsrmquota`
Data source         | ??
Counters            | `MSFT_FSRMQUOTA`
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------

`wmi_msft_fsrmquota_count` | Number of Quotas | counter |None
`wmi_msft_fsrmquota_description` | A string up to 1KB in size. Optional. The default value is an empty string. (Description) | counter |None
`wmi_msft_fsrmquota_disabled` | If True, the quota is disabled. The default value is False. (Disabled) | counter |None
`wmi_msft_fsrmquota_matchestemplate` | If True, the property values of this quota match those values of the template from which it was derived. (MatchesTemplate) | counter |None
`wmi_msft_fsrmquota_peak_usage` | The highest amount of disk space usage charged to this quota. (PeakUsage) | counter |None
`wmi_msft_fsrmquota_size` | The size of the quota. If the Template property is not provided then the Size property must be provided (Size) | counter |None
`wmi_msft_fsrmquota_softlimit` | If True, the quota is a soft limit. If False, the quota is a hard limit. The default value is False. Optional (SoftLimit) | counter |None
`wmi_msft_fsrmquota_template` | A valid quota template name. Up to 1KB in size. Optional (Template) | counter |None
`wmi_msft_fsrmquota_usage` | The current amount of disk space usage charged to this quota. (Usage) | counter |None


### Example metric
Show rate of Quotas usage:
```
rate(wmi_msft_fsrmquota_usage)[1d]
```

## Useful queries

## Alerting examples
**prometheus.rules**
```yaml
  - alert: "HighQuotasUsage"
    expr: "wmi_msft_fsrmquota_usage{instance="UFRTRP01LMRP050.COMAUGROUP.COM:9182"} / wmi_msft_fsrmquota_size{instance="UFRTRP01LMRP050.COMAUGROUP.COM:9182"} >0.85"
    for: "10m"
    labels:
      severity: "high"
    annotations:
      summary: "High Quotas Usage"
      description: "High use of File Ressource.\n Quotas: {{ $labels.quotaPath }}\n Current use : {{ $value }}"
```

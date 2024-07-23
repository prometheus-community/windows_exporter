# license collector

The license collector exposes metrics about the Windows license status.

|||
-|-
Metric name prefix  | `license`
Data source         | Win32
Enabled by default? | No

## Flags

None

## Metrics

| Name                     | Description    | Type  | Labels  |
|--------------------------|----------------|-------|---------|
| `windows_license_status` | license status | gauge | `state` |

### Example metric

```
# HELP windows_license_status Status of windows license
# TYPE windows_license_status gauge
windows_license_status{state="genuine"} 1
windows_license_status{state="invalid_license"} 0
windows_license_status{state="last"} 0
windows_license_status{state="offline"} 0
windows_license_status{state="tampered"} 0
```


## Useful queries

Show if the license is genuine

```
windows_license_status{state="genuine"}
```

## Alerting examples
**prometheus.rules**
```yaml
  - alert: "WindowsLicense"
    expr: 'windows_license_status{state="genuine"} == 0'
    for: "10m"
    labels:
      severity: "high"
    annotations:
      summary: "Windows system license is not genuine"
      description: "The Windows system license is not genuine. Please check the license status."
```

# adfs collector

The adfs collector exposes metrics about Active Directory Federation Services. Note that this collector has only been tested against ADFS 4.0 (2016).
Other ADFS versions may work but are not tested.

|||
-|-
Metric name prefix  | `adfs`
Data source         | Perflib
Counters            | `AD FS`
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_adfs_ad_login_connection_failures_total` | Total number of connection failures between the ADFS server and the Active Directory domain controller(s) | counter | None
`windows_adfs_certificate_authentications_total` | Total number of [User Certificate](https://docs.microsoft.com/en-us/windows-server/identity/ad-fs/operations/configure-user-certificate-authentication) authentications. I.E. smart cards or mobile devices with provisioned client certificates | counter | None
`windows_adfs_device_authentications_total` | Total number of [device authentications](https://docs.microsoft.com/en-us/windows-server/identity/ad-fs/operations/device-authentication-controls-in-ad-fs) (SignedToken, clientTLS, PkeyAuth). Device authentication is only available on ADFS 2016 or later | counter | None
`windows_adfs_extranet_account_lockouts_total` | Total number of [extranet lockouts](https://docs.microsoft.com/en-us/windows-server/identity/ad-fs/operations/configure-ad-fs-extranet-smart-lockout-protection). Requires the Extranet Lockout feature to be enabled | counter | None
`windows_adfs_federated_authentications_total` | Total number of authentications from federated sources. E.G. Office365 | counter | None
`windows_adfs_passport_authentications_total` | Total number of authentications from [Microsoft Passport](https://en.wikipedia.org/wiki/Microsoft_account) (now named Microsoft Account) | counter | None
`windows_adfs_password_change_failed_total` | Total number of failed password changes. The Password Change Portal must be enabled in the AD FS Management tool in order to allow user password changes | counter | None
`windows_adfs_password_change_succeeded_total` | Total number of succeeded password changes. The Password Change Portal must be enabled in the AD FS Management tool in order to allow user password changes | counter | None
`windows_adfs_token_requests_total` | Total number of requested access tokens | counter | None
`windows_adfs_windows_integrated_authentications_total` | Total number of Windows integrated authentications using Kerberos or NTLM | counter | None

### Example metric
Show rate of device authentications in AD FS:
```
rate(windows_adfs_device_authentications)[2m]
```

## Useful queries

## Alerting examples
**prometheus.rules**
```yaml
  - alert: "HighExtranetLockouts"
    expr: "rate(windows_adfs_extranet_account_lockouts)[2m] > 100"
    for: "10m"
    labels:
      severity: "high"
    annotations:
      summary: "High number of AD FS extranet lockouts"
      description: "High number of AD FS extranet lockouts may indicate a password spray attack.\n Server: {{ $labels.instance }}\n Number of lockouts: {{ $value }}"
```

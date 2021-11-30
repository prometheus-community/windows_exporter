# adfs collector

The ADFS collector exposes metrics about Active Directory Federation Services. Note that this collector has only been tested against ADFS 4.0/ [Farm Behavior (FLB) 3](https://docs.microsoft.com/en-us/windows-server/identity/ad-fs/deployment/upgrading-to-ad-fs-in-windows-server#ad-fs-farm-behavior-levels-fbl) (Server 2016).
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
`ad_login_connection_failures_total` | Total number of connection failures to an Active Directory domain controller | counter | None
`certificate_authentications_total` | Total number of User Certificate authentications | counter | None
`device_authentications_total` | Total number of Device authentications | counter | None
`extranet_account_lockouts_total` | Total number of Extranet Account Lockouts | counter | None
`federated_authentications_total` | Total number of authentications from a federated source | counter | None
`passport_authentications_total` | Total number of Microsoft Passport SSO authentications | counter | None
`passive_requests_total` | Total number of passive (browser-based) requests | counter | None
`password_change_failed_total` | Total number of failed password changes | counter | None
`password_change_succeeded_total` | Total number of successful password changes | counter | None
`token_requests_total` | Total number of token requests | counter | None
`windows_integrated_authentications_total` | Total number of Windows integrated authentications (Kerberos/NTLM) | counter | None
`oauth_authorization_requests_total` | Total number of incoming requests to the OAuth Authorization endpoint | counter | None
`oauth_client_authentication_success_total` | Total number of successful OAuth client Authentications | counter | None
`oauth_client_authentication_failure_total` | Total number of failed OAuth client Authentications | counter | None
`oauth_client_credentials_failure_total` | Total number of failed OAuth Client Credentials Requests | counter | None
`oauth_client_credentials_success_total` | Total number of successful RP tokens issued for OAuth Client Credentials Requests | counter | None
`oauth_client_privkey_jtw_authentication_failure_total` | Total number of failed OAuth Client Private Key Jwt Authentications | counter | None
`oauth_client_privkey_jwt_authentications_success_total` | Total number of successful OAuth Client Private Key Jwt Authentications | counter | None
`oauth_client_secret_basic_authentications_failure_total` | Total number of failed OAuth Client Secret Basic Authentications | counter | None
`oauth_client_secret_basic_authentications_success_total` | Total number of successful OAuth Client Secret Basic Authentications | counter | None
`oauth_client_secret_post_authentications_failure_total` | Total number of failed OAuth Client Secret Post Authentications | counter | None
`oauth_client_secret_post_authentications_success_total` | Total number of successful OAuth Client Secret Post Authentications | counter | None
`oauth_client_windows_authentications_failure_total` | Total number of failed OAuth Client Windows Integrated Authentications | counter | None
`oauth_client_windows_authentications_success_total` | Total number of successful OAuth Client Windows Integrated Authentications | counter | None
`oauth_logon_certificate_requests_failure_total` | Total number of failed OAuth Logon Certificate Requests | counter | None
`oauth_logon_certificate_token_requests_success_total` | Total number of successful RP tokens issued for OAuth Logon Certificate Requests | counter | None
`oauth_password_grant_requests_failure_total` | Total number of failed OAuth Password Grant Requests | counter | None
`oauth_password_grant_requests_success_total` | Total number of successful OAuth Password Grant Requests | counter | None
`oauth_token_requests_success_total` | Total number of successful RP tokens issued over OAuth protocol | counter | None
`samlp_token_requests_success_total` | Total number of successful RP tokens issued over SAML-P protocol | counter | None
`sso_authentications_failure_total` | Total number of failed SSO authentications | counter | None
`sso_authentications_success_total` | Total number of successful SSO authentications | counter | None
`wsfed_token_requests_success_total` | Total number of successful RP tokens issued over WS-Fed protocol | counter | None
`wstrust_token_requests_success_total` | Total number of successful RP tokens issued over WS-Trust protocol | counter | None
`userpassword_authentications_failure_total` | Total number of failed AD U/P authentications | counter | None
`userpassword_authentications_success_total` | Total number of successful AD U/P authentications | counter | None
`external_authentications_failure_total` | Total number of failed authentications from external MFA providers | counter | None
`external_authentications_success_total` | Total number of successful authentications from external MFA providers | counter | None
`db_artifact_failure_total` | Total number of failures connecting to the artifact database | counter | None
`db_artifact_query_time_seconds_total` | Accumulator of time taken for an artifact database query | counter | None
`db_config_failure_total` | Total number of failures connecting to the configuration database | counter | None
`db_config_query_time_seconds_total` | Accumulator of time taken for a configuration database query | counter | None
`federation_metadata_requests_total` | Total number of Federation Metadata requests | counter | None

### Example metric
Show rate of device authentications in AD FS:
```
rate(windows_adfs_device_authentications)[2m]
```

## Useful queries

|Query|Description|
|---|----|
|`rate(windows_adfs_oauth_password_grant_requests_failure_total[5m])`| Rate of OAuth requests failing due to bad client/resource values|
|`rate(windows_adfs_userpassword_authentications_failures_total[5m])`| Rate of `/adfs/oauth2/token/` requests failing due to bad username/password values (possible credential spraying)|

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

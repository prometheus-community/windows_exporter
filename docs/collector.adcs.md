# adcs collector

The adcs collector exposes metrics about Active Directory Certificate Services, Note that this collector has only been tested against Windows Server 2019.
Other Windows Server versions may work but are not tested.

|||
-|-
Metric name prefix  | `adcs`
Data source         | Perflib
Counters            | `Certification Authority`
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_adcs_requests_total` | Total certificate requests processed | counter | `cert_template`
`windows_adcs_request_processing_time_seconds` | Last time elapsed for certificate requests | gauge | `cert_template`
`windows_adcs_retrievals_total` | Total certificate retrieval requests processed | counter | `cert_template`
`windows_adcs_retrievals_processing_time_seconds` | Last time elapsed for certificate retrieval request | gauge | `cert_template`
`windows_adcs_failed_requests_total` | Total failed certificate requests processed | counter | `cert_template`
`windows_adcs_issued_requests_total` | Total issued certificate requests processed | counter | `cert_template`
`windows_adcs_pending_requests_total` | Total pending certificate requests processed | counter | `cert_template`
`windows_adcs_request_cryptographic_signing_time_seconds` | Last time elapsed for signing operation request | gauge | `cert_template`
`windows_adcs_request_policy_module_processing_time_seconds` | Last time elapsed for policy module processing request | gauge | `cert_template`
`windows_adcs_challenge_responses_total` | Total certificate challenge responses processed | counter | `cert_template`
`windows_adcs_challenge_response_processing_time_seconds` | Last time elapsed for challenge response | gauge | `cert_template`
`windows_adcs_signed_certificate_timestamp_lists_total` | Total Signed Certificate Timestamp Lists processed | counter | `cert_template`
`windows_adcs_signed_certificate_timestamp_list_processing_time_seconds` | Last time elapsed for Signed Certificate Timestamp List | gauge | `cert_template`

### Example metric
```
windows_adcs_issued_requests_total{cert_template="Administrator"} 0
windows_adcs_issued_requests_total{cert_template="DirectoryEmailReplication"} 0
windows_adcs_issued_requests_total{cert_template="DomainController"} 1
windows_adcs_issued_requests_total{cert_template="DomainControllerAuthentication"} 0
windows_adcs_issued_requests_total{cert_template="EFS"} 0
windows_adcs_issued_requests_total{cert_template="EFSRecovery"} 0
windows_adcs_issued_requests_total{cert_template="KerberosAuthentication"} 0
windows_adcs_issued_requests_total{cert_template="Machine"} 0
windows_adcs_issued_requests_total{cert_template="SubCA"} 0
windows_adcs_issued_requests_total{cert_template="User"} 0
windows_adcs_issued_requests_total{cert_template="WebServer"} 0
windows_adcs_issued_requests_total{cert_template="_Total"} 1
```

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

# terminal_service_gateway collector

The terminal_service_gateway collector exposes terminal service gateway performance metrics.

|||
-|-
Metric name prefix  | `terminal_service_gateway`
Data source         | Perflib/WMI
Classes             | [`Win32_PerfRawData_TSGateway_TerminalServiceGateway`](https://wutils.com/wmi/root/cimv2/win32_perfrawdata_tsgateway_terminalservicegateway/)
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_terminal_service_gateway_connection_request_authorization_time_seconds` | gauge | Shows the average connection request authentication and authorization times in seconds
`windows_terminal_service_gateway_current_connections_count` | gauge | Shows the total number of active/inactive connections to the RDG server at any given moment
`windows_terminal_service_gateway_failed_connection_authorization_total` | gauge | Shows the total number of requests that failed due to insufficient connection authorization privilege
`windows_terminal_service_gateway_failed_connections_total` | gauge | Shows the number of connection requests that are all failed due to errors and authorization failure
`windows_terminal_service_gateway_failed_resource_authorization_total` | gauge | Shows the total number of requests that failed due to insufficient resource authorization privilege
`windows_terminal_service_gateway_successful_connections_total` | gauge | Shows the number of requests that were successfully processed and connected

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

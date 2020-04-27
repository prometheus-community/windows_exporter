# exchange collector

The exchange collector exposes metrics about the MS Exchange server. All metrics are picked based on this document : [https://docs.microsoft.com/en-us/exchange/exchange-2013-performance-counters-exchange-2013-help]|(https://docs.microsoft.com/en-us/exchange/exchange-2013-performance-counters-exchange-2013-help)

|||
-|-
Metric name prefix  | `exchange`
Classes 			| [Win32_PerfRawData_MSExchangeADAccess_MSExchangeADAccessProcesses](https://docs.microsoft.com/en-us/exchange/)<br/> [Win32_PerfRawData_MSExchangeTransportQueues_MSExchangeTransportueues](https://docs.microsoft.com/en-us/exchange/)<br/> [Win32_PerfRawData_ESE_MSExchangeDatabaseInstances](https://docs.microsoft.com/en-us/exchange/)<br/> [Win32_PerfRawData_MSExchangeHttpProxy_MSExchangeHttpProxy](https://docs.microsoft.com/en-us/exchange/)<br/> [Win32_PerfRawData_MSExchangeActiveSync_MSExchangeActiveSync](https://docs.microsoft.com/en-us/exchange/)<br/> [Win32_PerfRawData_MSExchangeAvailabilityService_MSExchangeAvailabilityService](https://docs.microsoft.com/en-us/exchange/)<br/> [Win32_PerfRawData_MSExchangeOWA_MSExchangeOWA](https://docs.microsoft.com/en-us/exchange/)<br/> [Win32_PerfRawData_MSExchangeAutodiscover_MSExchangeAutodiscover](https://docs.microsoft.com/en-us/exchange/)<br/> [Win32_PerfRawData_MSExchangeWorkloadManagementWorkloads_MSExchangeWorkloadManagementWorkloads](https://docs.microsoft.com/en-us/exchange/)<br/> [Win32_PerfRawData_MSExchangeRpcClientAccess_MSExchangeRpcClientAccess](https://docs.microsoft.com/en-us/exchange/)<br/>
Enabled by default? | No

## Flags
Since the official WMI class names are extremely long, the following shorthands are used instead.
`--collectors.exchange.class-list` lists these shorthands along with the real WMI class names

* `ldap`
* `transport_queues`
* `database_instances`
* `http_proxy`
* `activesync`
* `availability_service`
* `owa`
* `autodiscover`
* `management_workloads`
* `rpc`

### `--collectors.exchange.list`
List all available MS Exchange WMI class long- and short-names

### `--collectors.exchange.classes-enabled`
Comma-separated list of exchange WMI classes from which the exporter should collect data. 
If no classes are given, all classes will be queried.

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`rpc_avg_latency` | The latency (ms), averaged for the past 1024 packets | counter |
`rpc_requests` | Number of client requests currently being processed by  the RPC Client Access service | gauge |
`rpc_active_user_count` | Number of unique users that have shown some kind of activity in the last 2 minutes | gauge |
`rpc_connection_count` | Total number of client connections maintained | counter |
`rpc_ops_per_sec` | The rate (ops/s) at wich RPC operations occur | counter | 
`rpc_user_count` | Number of users | counter |
`ldap_read_time` | Time (in ms) to send an LDAP read request and receive a response | gauge | name
`ldap_search_time` | Time (in ms) to send an LDAP search request and receive a response | gauge | name
`ldap_timeout_errors_per_sec` | LDAP timeout errors per second | gauge | name
`ldap_long_running_ops_per_min` | Long Running LDAP operations pr minute | gauge | name
`ldap_searches_time_limit_exceeded_per_min` | LDAP searches time limit exceeded per minute | gauge | name
`transport_queues_external_active_remote_delivery` | External Active Remote Delivery Queue length | gauge | name
`transport_queues_internal_active_remote_delivery` | Internal Active Remote Delivery Queue length | gauge | name
`transport_queues_active_mailbox_delivery` | Active Mailbox Delivery Queue length | gauge | name
`transport_queues_retry_mailbox_delivery` | Retry Mailbox Delivery Queue length | gauge | name
`transport_queues_unreachable` | Unreachable Queue length | gauge | name
`transport_queues_external_largest_delivery` | External Largest Delivery Queue length | gauge | name
`transport_queues_internal_largest_delivery` | Internal Largest Delivery Queue length | gauge | name
`transport_queues_poison` | Poison Queue length | gauge | name
`iodb_reads_avg_latency` | Average time (in ms) per database read operation | gauge | name
`iodb_writes_avg_latency` | Average time (in ms) per database write opreation | gauge | name
`iodb_log_writes_avg_latency` | Average time (in ms) per Log write operation | gauge | name
`iodb_reads_recovery_avg_latency` | Average time (in ms) per passive database read operation  | gauge | name
`iodb_writes_recovery_avg_latency` | Average time (in ms) per passive database write operation | gauge | name
`http_proxy_mailbox_server_locator_avg_latency` | Average latency (ms) of MailboxServerLocator web service calls | gauge | name
`http_proxy_avg_auth_latency` | Average time spent authenticating CAS requests over the last 200 samples | gauge | name
`http_proxy_avg_client_access_server_proccessing_latency` | Average latency (ms) of CAS processing time over the last 200 requests | gauge | name
`http_proxy_mailbox_server_proxy_failure_rate` | Percentage of connection failures between this CAS and MBX servers over the last 200 samples | gauge | name
`http_proxy_outstanding_proxy_requests` | Number of concurrent outstanding proxy requests | gauge | name
`http_proxy_requests_per_sec` | Number of proxy requests processed each second | gauge | name
`activesync_requests_per_sec` | Number of HTTP requests received from the client via ASP.NET per second. Used to determine current user load | gauge |
`activesync_ping_cmds_pending` | Number of ping commands currently pending in the queue | gauge |
`activesync_sync_cmds_pending` | Number of sync commands processed per second. Clients use this command to synchronize items within a folder | gauge |
`avail_service_requests_per_sec` | Number of requests serviced per second | gauge |
`owa_current_unique_users` | Number of unique users currently logged on to Outlook Web App | gauge |
`owa_requests_per_sec` | Number of requests handled by Outlook Web App per second | gauge |
`autodiscover_requests_per_sec` | Number of autodiscover service requests processed each second | gauge |
`workload_active_tasks` | Number of active tasks currently running in the background for workload management | gauge |
`workload_completed_tasks` | Number of workload management tasks that have been completed | counter |
`workload_queued_tasks` | Number of workload management tasks that are currently queued up waiting to be processed | counter |

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_



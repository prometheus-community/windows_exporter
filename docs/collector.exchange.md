# exchange collector

The exchange collector collects metrics from MS Exchange hosts through perflib
=======


|||
-|-
Metric name prefix  | `exchange`
Classes 			| [Win32_PerfRawData_MSExchangeADAccess_MSExchangeADAccessProcesses](https://docs.microsoft.com/en-us/exchange/)<br/> [Win32_PerfRawData_MSExchangeTransportQueues_MSExchangeTransportueues](https://docs.microsoft.com/en-us/exchange/)<br/> [Win32_PerfRawData_ESE_MSExchangeDatabaseInstances](https://docs.microsoft.com/en-us/exchange/)<br/> [Win32_PerfRawData_MSExchangeHttpProxy_MSExchangeHttpProxy](https://docs.microsoft.com/en-us/exchange/)<br/> [Win32_PerfRawData_MSExchangeActiveSync_MSExchangeActiveSync](https://docs.microsoft.com/en-us/exchange/)<br/> [Win32_PerfRawData_MSExchangeAvailabilityService_MSExchangeAvailabilityService](https://docs.microsoft.com/en-us/exchange/)<br/> [Win32_PerfRawData_MSExchangeOWA_MSExchangeOWA](https://docs.microsoft.com/en-us/exchange/)<br/> [Win32_PerfRawData_MSExchangeAutodiscover_MSExchangeAutodiscover](https://docs.microsoft.com/en-us/exchange/)<br/> [Win32_PerfRawData_MSExchangeWorkloadManagementWorkloads_MSExchangeWorkloadManagementWorkloads](https://docs.microsoft.com/en-us/exchange/)<br/> [Win32_PerfRawData_MSExchangeRpcClientAccess_MSExchangeRpcClientAccess](https://docs.microsoft.com/en-us/exchange/)<br/>
Enabled by default? | No

## Flags

### `--collectors.exchange.list`
Lists the Perflib Objects that are queried for data along with the perlfib object id

### `--collectors.exchange.enabled`
Comma-separated list of collectors to use, for example: `--collectors.exchange.enabled=AvailabilityService,OutlookWebAccess`. Matching is case-sensitive. Depending on the exchange installation not all performance counters are available. Use `--collectors.exchange.list` to obtain a list of supported collectors.

## Metrics
Name          | Description
--------------|---------------
`windows_exchange_rpc_avg_latency_sec` | The latency (sec), averaged for the past 1024 packets
`windows_exchange_rpc_requests` | Number of client requests currently being processed by  the RPC Client Access service
`windows_exchange_rpc_active_user_count` | Number of unique users that have shown some kind of activity in the last 2 minutes
`windows_exchange_rpc_connection_count` | Total number of client connections maintained
`windows_exchange_rpc_operations_total` | The rate at which RPC operations occur
`windows_exchange_rpc_user_count` | Number of users
`windows_exchange_ldap_read_time_sec` | Time (sec) to send an LDAP read request and receive a response
`windows_exchange_ldap_search_time_sec` | Time (sec) to send an LDAP search request and receive a response
`windows_exchange_ldap_write_time_sec` | Time (sec) to send an LDAP Add/Modify/Delete request and receive a response
`windows_exchange_ldap_timeout_errors_total` | Total number of LDAP timeout errors
`windows_exchange_ldap_long_running_ops_per_sec` | Long Running LDAP operations per second
`windows_exchange_transport_queues_external_active_remote_delivery` | External Active Remote Delivery Queue length
`windows_exchange_transport_queues_internal_active_remote_delivery` | Internal Active Remote Delivery Queue length
`windows_exchange_transport_queues_active_mailbox_delivery` | Active Mailbox Delivery Queue length
`windows_exchange_transport_queues_retry_mailbox_delivery` | Retry Mailbox Delivery Queue length
`windows_exchange_transport_queues_unreachable` | Unreachable Queue length
`windows_exchange_transport_queues_external_largest_delivery` | External Largest Delivery Queue length
`windows_exchange_transport_queues_internal_largest_delivery` | Internal Largest Delivery Queue length 
`windows_exchange_transport_queues_poison` | Poison Queue length
`windows_exchange_http_proxy_mailbox_server_locator_avg_latency_sec` | Average latency (sec) of MailboxServerLocator web service calls
`windows_exchange_http_proxy_avg_auth_latency` | Average time spent authenticating CAS requests over the last 200 samples
`windows_exchange_http_proxy_outstanding_proxy_requests` | Number of concurrent outstanding proxy requests
`windows_exchange_http_proxy_requests_total` | Number of proxy requests processed each second
`windows_exchange_avail_service_requests_per_sec` | Number of requests serviced per second
`windows_exchange_owa_current_unique_users` | Number of unique users currently logged on to Outlook Web App
`windows_exchange_owa_requests_total` | Number of requests handled by Outlook Web App per second
`windows_exchange_autodiscover_requests_total` | Number of autodiscover service requests processed each second
`windows_exchange_workload_active_tasks` | Number of active tasks currently running in the background for workload management
`windows_exchange_workload_completed_tasks` | Number of workload management tasks that have been completed
`windows_exchange_workload_queued_tasks` | Number of workload management tasks that are currently queued up waiting to be processed
`windows_exchange_workload_yielded_tasks` | The total number of tasks that have been yielded by a workload
`windows_exchange_workload_is_active` | Active indicates whether the workload is in an active (1) or paused (0) state
`windows_exchange_activesync_requests_total` | Num HTTP requests received from the client via ASP.NET per sec. Shows Current user load
`windows_exchange_http_proxy_avg_cas_proccessing_latency_sec` | Average latency (sec) of CAS processing time over the last 200 reqs
`windows_exchange_http_proxy_mailbox_proxy_failure_rate` | % of failures between this CAS and MBX servers over the last 200 sample
`windows_exchange_activesync_ping_cmds_pending` | Number of ping commands currently pending in the queue
`windows_exchange_activesync_sync_cmds_total` | Number of sync commands processed per second. Clients use this command to synchronize items within a folder

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_


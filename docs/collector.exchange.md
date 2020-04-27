# exchange collector

The exchange collector exposes metrics about the MS Exchange server

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
`wmi_exchange_rpc_avg_latency_sec` | The latency (seconds), averaged for the past 1024 packets | gauge |
`wmi_exchange_rpc_requests` | Number of client requests currently being processed by  the RPC Client Access service | gauge |
`wmi_exchange_rpc_active_user_count` | Number of unique users that have shown some kind of activity in the last 2 minutes | gauge |
`wmi_exchange_rpc_connection_count` | Total number of client connections maintained | gauge |
`wmi_exchange_rpc_ops_per_sec` | The rate (ops/s) at wich RPC operations occur | counter | 
`wmi_exchange_rpc_user_count` | Number of users | gauge |
`wmi_exchange_ldap_read_time_sec` | Time (seconds) to send an LDAP read request and receive a response | gauge | name
`wmi_exchange_ldap_search_time_se` | Time (seconds) to send an LDAP search request and receive a response | gauge | name
`wmi_exchange_ldap_timeout_errors_per_sec` | LDAP timeout errors per second | gauge | name
`wmi_exchange_ldap_long_running_ops_per_sec` | Long Running LDAP operations per second | gauge | name
`wmi_exchange_ldap_searches_time_limit_exceeded_per_min` | LDAP searches time limit exceeded per minute | gauge | name
`wmi_exchange_transport_queues_external_active_remote_delivery` | External Active Remote Delivery Queue length | gauge | name
`wmi_exchange_transport_queues_internal_active_remote_delivery` | Internal Active Remote Delivery Queue length | gauge | name
`wmi_exchange_transport_queues_active_mailbox_delivery` | Active Mailbox Delivery Queue length | gauge | name
`wmi_exchange_transport_queues_retry_mailbox_delivery` | Retry Mailbox Delivery Queue length | gauge | name
`wmi_exchange_transport_queues_unreachable` | Unreachable Queue length | gauge | name
`wmi_exchange_transport_queues_external_largest_delivery` | External Largest Delivery Queue length | gauge | name
`wmi_exchange_transport_queues_internal_largest_delivery` | Internal Largest Delivery Queue length | gauge | name
`wmi_exchange_transport_queues_poison_sec` | Poison Queue length | gauge | name
`wmi_exchange_iodb_reads_avg_latency_sec` | Average time (seconds) per database read operation | counter | name
`wmi_exchange_iodb_writes_avg_latency_sec` | Average time (seconds) per database write opreation | counter | name
`wmi_exchange_iodb_log_writes_avg_latency_sec` | Average time (seconds) per Log write operation | counter | name
`wmi_exchange_iodb_reads_recovery_avg_latency_sec` | Average time (seconds) per passive database read operation  | counter | name
`wmi_exchange_iodb_writes_recovery_avg_latency_sec` | Average time (seconds) per passive database write operation | counter | name

`wmi_exchange_http_proxy_mailbox_server_locator_avg_latency_sec` | Average latency (seconds) of MailboxServerLocator web service calls | counter | name
`wmi_exchange_http_proxy_avg_auth_latency` | Average time spent authenticating CAS requests over the last 200 samples | gauge | name
`wmi_exchange_http_proxy_avg_client_access_server_proccessing_latency_sec` | Average latency (seconds) of CAS processing time over the last 200 requests | gauge | name
`wmi_exchange_http_proxy_mailbox_server_proxy_failure_rate` | Percentage of connection failures between this CAS and MBX servers over the last 200 samples | gauge | name
`wmi_exchange_http_proxy_outstanding_proxy_requests` | Number of concurrent outstanding proxy requests | gauge | name
`wmi_exchange_http_proxy_requests_per_sec` | Number of proxy requests processed each second | counter | name

`wmi_exchange_activesync_requests_per_sec` | Number of HTTP requests received from the client via ASP.NET per second. Used to determine current user load | counter |

`wmi_exchange_activesync_ping_cmds_pending` | Number of ping commands currently pending in the queue | counter |

`wmi_exchange_activesync_sync_cmds_pending` | Number of sync commands processed per second. Clients use this command to synchronize items within a folder | counter |

`wmi_exchange_avail_service_requests_per_sec` | Number of requests serviced per second | counter |

`wmi_exchange_owa_current_unique_users` | Number of unique users currently logged on to Outlook Web App | gauge |
`wmi_exchange_owa_requests_per_sec` | Number of requests handled by Outlook Web App per second | counter |

`wmi_exchange_autodiscover_requests_per_sec` | Number of autodiscover service requests processed each second | counter |

`wmi_exchange_workload_active_tasks` | Number of active tasks currently running in the background for workload management | gauge |
`wmi_exchange_workload_completed_tasks` | Number of workload management tasks that have been completed | counter |
`wmi_exchange_workload_queued_tasks` | Number of workload management tasks that are currently queued up waiting to be processed | counter |

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_



# service collector

The service collector exposes information as metrics about Windows Services

|||
-|-
Metric name prefix  | `service_info`
Classes             | -
Enabled by default? | Yes



## Flags

None

## Metrics

| Name                              | Description                                                             | Type  | Labels           |
|-----------------------------------|-------------------------------------------------------------------------|-------|------------------|
| `windows_service_info_run_as`     | Contains service information run as user in labels, constant 1          | gauge | name, run_as     |
| `windows_service_info_start_mode` | The start mode of the service, 1 if the current start mode, 0 otherwise | gauge | name, start_mode |

For the values of the `start_mode`, `status` and `run_as` labels, see below.


### Start modes

A service can have the following start modes:
- `boot`
- `system`
- `auto`
- `manual`
- `disabled`

Note that there is some overlap with service state.

### Run As

Account name under which a service runs. Depending on the service type, the account name may be in the form of "DomainName\Username" or UPN format ("Username@DomainName").


### Example metric

```
# HELP windows_service_info_run_as The start mode of the service (StartMode)
# TYPE windows_service_info_run_as gauge
windows_service_info_run_as{name="AGSService",run_as="LocalSystem"} 1
windows_service_info_run_as{name="AJRouter",run_as="NT AUTHORITY\\LocalService"} 1
windows_service_info_run_as{name="ALG",run_as="NT AUTHORITY\\LocalService"} 1
# HELP windows_service_info_start_mode The start mode of the service (StartMode)
# TYPE windows_service_info_start_mode gauge
windows_service_info_start_mode{name="AGSService",start_mode="auto"} 0
windows_service_info_start_mode{name="AGSService",start_mode="boot"} 0
windows_service_info_start_mode{name="AGSService",start_mode="disabled"} 1
windows_service_info_start_mode{name="AGSService",start_mode="manual"} 0
windows_service_info_start_mode{name="AGSService",start_mode="system"} 0
windows_service_info_start_mode{name="AJRouter",start_mode="auto"} 0
windows_service_info_start_mode{name="AJRouter",start_mode="boot"} 0
windows_service_info_start_mode{name="AJRouter",start_mode="disabled"} 0
windows_service_info_start_mode{name="AJRouter",start_mode="manual"} 1
windows_service_info_start_mode{name="AJRouter",start_mode="system"} 0
windows_service_info_start_mode{name="ALG",start_mode="auto"} 0
windows_service_info_start_mode{name="ALG",start_mode="boot"} 0
windows_service_info_start_mode{name="ALG",start_mode="disabled"} 0
windows_service_info_start_mode{name="ALG",start_mode="manual"} 1
windows_service_info_start_mode{name="ALG",start_mode="system"} 0   
```
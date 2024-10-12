# logon collector

The logon collector exposes metrics detailing the active user logon sessions.

|                     |           |
|---------------------|-----------|
| Metric name prefix  | `logon`   |
| Source              | Win32 API |
| Enabled by default? | No        |

## Flags

None

## Metrics

| Name                                      | Description                                | Type  | Labels                             |
|-------------------------------------------|--------------------------------------------|-------|------------------------------------|
| `windows_logon_session_timestamp_seconds` | timestamp of the logon session in seconds. | gauge | `domain`, `id`, `type`, `username` |

### Example metric
Query the total number of interactive logon sessions
```
# HELP windows_logon_session_timestamp_seconds timestamp of the logon session in seconds
# TYPE windows_logon_session_timestamp_seconds gauge
windows_logon_session_timestamp_seconds{domain="",id="103590",type="System",username=""} 1.728759837e+09
windows_logon_session_timestamp_seconds{domain="Font Driver Host",id="104592",type="Interactive",username="UMFD-0"} 1.728759837e+09
windows_logon_session_timestamp_seconds{domain="Font Driver Host",id="124850",type="Interactive",username="UMFD-1"} 1.728759838e+09
windows_logon_session_timestamp_seconds{domain="JOK-PC",id="521539",type="Interactive",username="Jan"} 1.728759839e+09
windows_logon_session_timestamp_seconds{domain="JOK-PC",id="521983",type="Interactive",username="Jan"} 1.728759839e+09
windows_logon_session_timestamp_seconds{domain="NT-AUTORITÄT",id="997",type="Service",username="Lokaler Dienst"} 1.728759838e+09
windows_logon_session_timestamp_seconds{domain="WORKGROUP",id="996",type="Service",username="JOK-PC$"} 1.728759838e+09
windows_logon_session_timestamp_seconds{domain="WORKGROUP",id="999",type="System",username="JOK-PC$"} 1.728759837e+09
windows_logon_session_timestamp_seconds{domain="Window Manager",id="148473",type="Interactive",username="DWM-1"} 1.728759838e+09
windows_logon_session_timestamp_seconds{domain="Window Manager",id="148582",type="Interactive",username="DWM-1"} 1.728759838e+09
```

### Possible values for `type`

- System
- Interactive
- Network
- Batch
- Service
- Proxy
- Unlock
- NetworkCleartext
- NewCredentials
- RemoteInteractive
- CachedInteractive
- CachedRemoteInteractive
- CachedUnlock

## Useful queries
Query the total number of local and remote (I.E. Terminal Services) interactive sessions.
```
count(windows_logon_logon_type{type=~"Interactive|RemoteInteractive"}) by (type)
```

## Alerting examples
_This collector doesn’t yet have alerting examples, we would appreciate your help adding them!_

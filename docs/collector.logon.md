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
| `windows_logon_session_logon_timestamp_seconds` | timestamp of the logon session in seconds. | gauge | `domain`, `id`, `type`, `username` |

### Example metric
Query the total number of interactive logon sessions
```
# HELP windows_logon_session_logon_timestamp_seconds timestamp of the logon session in seconds.
# TYPE windows_logon_session_logon_timestamp_seconds gauge
windows_logon_session_logon_timestamp_seconds{domain="",id="0x0:0x8c54",type="System",username=""} 1.72876928e+09
windows_logon_session_logon_timestamp_seconds{domain="Font Driver Host",id="0x0:0x991a",type="Interactive",username="UMFD-1"} 1.728769282e+09
windows_logon_session_logon_timestamp_seconds{domain="Font Driver Host",id="0x0:0x9933",type="Interactive",username="UMFD-0"} 1.728769282e+09
windows_logon_session_logon_timestamp_seconds{domain="Font Driver Host",id="0x0:0x994a",type="Interactive",username="UMFD-0"} 1.728769282e+09
windows_logon_session_logon_timestamp_seconds{domain="Font Driver Host",id="0x0:0x999d",type="Interactive",username="UMFD-1"} 1.728769282e+09
windows_logon_session_logon_timestamp_seconds{domain="Font Driver Host",id="0x0:0xbf25a",type="Interactive",username="UMFD-2"} 1.728769532e+09
windows_logon_session_logon_timestamp_seconds{domain="Font Driver Host",id="0x0:0xbf290",type="Interactive",username="UMFD-2"} 1.728769532e+09
windows_logon_session_logon_timestamp_seconds{domain="JKROEPKE",id="0x0:0x130241",type="Network",username="vm-jok-dev$"} 1.728769625e+09
windows_logon_session_logon_timestamp_seconds{domain="JKROEPKE",id="0x0:0x24f7c9",type="Network",username="vm-jok-dev$"} 1.728770121e+09
windows_logon_session_logon_timestamp_seconds{domain="JKROEPKE",id="0x0:0x276846",type="Network",username="vm-jok-dev$"} 1.728770195e+09
windows_logon_session_logon_timestamp_seconds{domain="JKROEPKE",id="0x0:0x3e4",type="Service",username="vm-jok-dev$"} 1.728769283e+09
windows_logon_session_logon_timestamp_seconds{domain="JKROEPKE",id="0x0:0x3e7",type="System",username="vm-jok-dev$"} 1.728769279e+09
windows_logon_session_logon_timestamp_seconds{domain="JKROEPKE",id="0x0:0x71d0f",type="Network",username="vm-jok-dev$"} 1.728769324e+09
windows_logon_session_logon_timestamp_seconds{domain="JKROEPKE",id="0x0:0x720a3",type="Network",username="vm-jok-dev$"} 1.728769324e+09
windows_logon_session_logon_timestamp_seconds{domain="JKROEPKE",id="0x0:0x725cb",type="Network",username="vm-jok-dev$"} 1.728769324e+09
windows_logon_session_logon_timestamp_seconds{domain="JKROEPKE",id="0x0:0x753d8",type="Network",username="vm-jok-dev$"} 1.728769325e+09
windows_logon_session_logon_timestamp_seconds{domain="JKROEPKE",id="0x0:0xa3913",type="Network",username="vm-jok-dev$"} 1.728769385e+09
windows_logon_session_logon_timestamp_seconds{domain="JKROEPKE",id="0x0:0xbe7f2",type="Network",username="jok"} 1.728769531e+09
windows_logon_session_logon_timestamp_seconds{domain="JKROEPKE",id="0x0:0xc76c4",type="RemoteInteractive",username="jok"} 1.728769533e+09
windows_logon_session_logon_timestamp_seconds{domain="NT AUTHORITY",id="0x0:0x3e3",type="Service",username="IUSR"} 1.728769295e+09
windows_logon_session_logon_timestamp_seconds{domain="NT AUTHORITY",id="0x0:0x3e5",type="Service",username="LOCAL SERVICE"} 1.728769283e+09
windows_logon_session_logon_timestamp_seconds{domain="NT Service",id="0x0:0xae4c7",type="Service",username="MSSQLSERVER"} 1.728769425e+09
windows_logon_session_logon_timestamp_seconds{domain="NT Service",id="0x0:0xb42f1",type="Service",username="SQLTELEMETRY"} 1.728769431e+09
windows_logon_session_logon_timestamp_seconds{domain="Window Manager",id="0x0:0xbfbac",type="Interactive",username="DWM-2"} 1.728769532e+09
windows_logon_session_logon_timestamp_seconds{domain="Window Manager",id="0x0:0xbfc72",type="Interactive",username="DWM-2"} 1.728769532e+09
windows_logon_session_logon_timestamp_seconds{domain="Window Manager",id="0x0:0xdedd",type="Interactive",username="DWM-1"} 1.728769283e+09
windows_logon_session_logon_timestamp_seconds{domain="Window Manager",id="0x0:0xdefd",type="Interactive",username="DWM-1"} 1.728769283e+09
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

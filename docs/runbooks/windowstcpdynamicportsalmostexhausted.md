---
title: WindowsTCPDynamicPortsAlmostExhausted
---

## Meaning

Connections of one address family (`af`) have held more than
`tcpDynamicPortUsageThreshold` (80 percent by default) of the TCP dynamic port
range for 10 minutes. The share divides `windows_tcp_dynamic_ports_in_use`, the
distinct local ports inside the range that a connection in any state holds,
including `TIME_WAIT`, by `windows_tcp_dynamic_port_range_size`. The alert
needs the `tcp` collector with its `dynamic_ports` sub-collector.

## Impact

Windows assigns the local port of an outgoing connection from this range. Once
it is exhausted, new outgoing connections fail with `WSAENOBUFS` (10055),
an explicit `bind` to port 0 with `WSAEADDRINUSE` (10048), and the System event log records event 4227 or 4231
from Tcpip. Services that open many short connections, such as web servers
calling backends, proxies, Exchange or SQL clients, fail first. Incoming
connections to listening ports are not affected.

## Diagnosis

1. Compare `windows_tcp_dynamic_ports_in_use` with
   `windows_tcp_connections{state="TIME_WAIT"}` and
   `windows_tcp_connections{state="ESTABLISHED"}` of the same `af`. Many
   `TIME_WAIT` connections point to a client that opens a new connection per
   request; many established ones to a connection leak.
2. On the server, find the processes that hold the ports:

   ```powershell
   netsh int ipv4 show dynamicport tcp
   netsh int ipv4 show excludedportrange tcp
   netstat -anoq | findstr BOUND
   Get-NetTCPConnection | Group-Object -Property State, OwningProcess | Sort-Object Count -Descending | Select-Object -First 10 Count, Name
   Get-Process -Id <OwningProcess>
   ```

3. Excluded port ranges inside the dynamic range (often reserved by Hyper-V,
   WinNAT or container networking) leave fewer usable ports than the range
   size, so exhaustion can happen below 100 percent. Sockets in state `BOUND`
   (bound, not connected) also hold ports but are most likely not counted by
   the metric; many of them for one process point to a socket leak.
4. Check the System event log for Tcpip events 4227 and 4231.

## Mitigation

Fix the client that opens the connections: reuse connections (HTTP
keep-alive, connection pools) instead of opening one per request, and close
leaked ones. Restarting the owning process frees its ports at once. As a short
term measure, enlarge the range with
`netsh int ipv4 set dynamicport tcp start=<port> num=<count>` (and `ipv6`),
keeping clear of ports that services listen on, and restart the exporter so
it reads the new range. Confirm the share stays below the threshold.

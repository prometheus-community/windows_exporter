# tcp collector

Read the root [`AGENTS.md`](../../../AGENTS.md) first.

## Connection table

`connections_state` and `dynamic_ports` both come from
`iphlpapi.GetTCPConnectionStats`, which reads `GetExtendedTcpTable`
(`TCP_TABLE_OWNER_PID_ALL`) once per address family and scrape and counts the
states and the distinct local ports in a range in one pass. Keep it a single
read when adding data from the table; on busy servers it has tens of thousands
of rows. Ports in the rows are in network byte order (`BigEndianUint32`).

## Dynamic port range

`dynamic_ports` reads the range once in Build from the `Internet` instance of
`MSFT_NetTCPSetting` (`root/StandardCimv2`) and applies it to IPv4 and IPv6.
`netsh int ipv4|ipv6 show dynamicport tcp` reports it per address family,
`Get-NetTCPSetting` per setting (`Internet`, `Datacenter`, `Compat`, custom),
and `Set-NetTCPSetting` can change each setting on Server 2016 and later.
Which of them the kernel uses for a connection depends on the transport filter.
If a probe on Windows shows that netsh and the `Internet` setting can differ,
read the range per address family instead. A count counts every state:
a `TIME_WAIT` connection still holds its local port, and a listener on an
ephemeral port takes it out of the pool.

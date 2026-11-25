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

`dynamic_ports` reads the range on every scrape from NSI
(`internal/headers/nsi`): `NsiGetParameter` on the TCP module
`{eb004a03-9b1a-11d4-9123-0050047759bc}`, object index 20, active store,
read-write struct, no key, 4 bytes. The first port is a big-endian uint16 at
offset 0, the number of ports a little-endian uint16 at offset 2
(`c0 00 00 40` = 49152/16384; `c0 01 ff 3f` after
`netsh int ipv4 set dynamicport tcp start=49153 num=16383`). UDP has the same
layout in object 4 of module `{eb004a02-…}`. Found by probing Windows 11 build
26300 and Windows Server 2025 build 26100 in November 2026; the API is
undocumented, so `nsi` has a layout test and a live test, and Build falls back
to the `Internet` instance of `MSFT_NetTCPSetting` (read once) when NSI fails.
There is one range per protocol: setting it for IPv4 with netsh changed IPv6
and every `MSFT_NetTCPSetting` instance too. If neither source works, Build
logs a warning and drops `dynamic_ports` instead of failing the collector.

Open questions: whether IPv4 and IPv6 also share one ephemeral port pool (then
the per-`af` counts should be summed for the alert), and whether
`GetExtendedTcpTable` lists sockets in state `BOUND` (bound, not connected,
shown by `netstat -anoq`); the docs assume it does not. A count counts every
state: a `TIME_WAIT` connection still holds its local port, and a listener on
an ephemeral port takes it out of the pool.

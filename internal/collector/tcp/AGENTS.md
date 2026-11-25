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
(`internal/headers/nsi`, `NsiGetAllParameters` on object 20 of the TCP
module, the same call as local `netsh int ipv4 show dynamicport tcp`). See
[`internal/headers/nsi/AGENTS.md`](../../headers/nsi/AGENTS.md) for the call
path, the layout and how it was found. NSI is undocumented, so Build falls
back to the `Internet` instance of `MSFT_NetTCPSetting` (read once) when NSI
fails, and logs a warning. There is one range per protocol, shared by IPv4 and
IPv6: setting it for IPv4 with netsh changed IPv6 and every
`MSFT_NetTCPSetting` instance too. If neither source works, Build drops `dynamic_ports` instead of failing
the collector.

IPv4 and IPv6 connections take their ports from one pool per protocol and
compartment, so `dynamic_ports_in_use` has no `af` label: both tables add to
one `iphlpapi.PortSet`, and a scrape where one table fails reports no
`dynamic_ports` metrics. Open question: whether `GetExtendedTcpTable` lists
sockets in state `BOUND` (bound, not connected, shown by `netstat -anoq`); the
docs assume it does not. A count counts every
state: a `TIME_WAIT` connection still holds its local port, and a listener on
an ephemeral port takes it out of the pool.

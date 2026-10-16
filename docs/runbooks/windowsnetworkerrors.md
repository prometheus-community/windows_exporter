---
title: WindowsNetworkErrors
---

## Meaning

Packets with inbound or outbound errors make up more than the configured share
of all packets on an interface for 15 minutes. The default
`networkErrorRatioThreshold` is 1 percent. The ratio divides the error rates
(`windows:net_received_errors:rate` plus `windows:net_outbound_errors:rate`) by
the packet rates (`windows:net_received_packets:rate` plus
`windows:net_sent_packets:rate`), so a quiet interface with a handful of errors
alerts while a busy interface with the same error count does not. Interfaces
that carry no packets produce no sample and cannot alert. The alert identifies
the interface using its `nic` label and keeps firing for `keepFiringFor` after
the ratio drops below the threshold.

## Impact

Packet errors may cause retransmissions, lower throughput, and intermittent
connection failures.

## Diagnosis

1. Inspect `windows:net_received_errors:rate` and
   `windows:net_outbound_errors:rate` separately to identify the direction, and
   compare them with `windows:net_received_packets:rate` and
   `windows:net_sent_packets:rate` to see whether a low packet rate inflates the
   ratio.
2. Compare traffic and `windows:net_received_utilization:ratio` and
   `windows:net_sent_utilization:ratio` with the normal baseline. These ratios
   use bandwidth in bytes per second and keep each direction separate.
3. Inspect local adapter state and counters:

   ```powershell
   Get-NetAdapter
   Get-NetAdapterStatistics
   ```

4. Inspect switch-port counters, cabling, link negotiation, NIC drivers, and
   System event logs. For virtual NICs, inspect the host and virtual switch.

## Mitigation

Repair faulty links or switch ports and correct driver or adapter configuration
problems. Apply driver and firmware fixes through the normal change process.
Relieve congestion if it contributes. Verify errors stop increasing and affected
application connections recover; cumulative counters do not need to return to
zero for the ratio alert to resolve. Raise `networkErrorRatioThreshold` only for
interfaces whose error share is known and accepted.

---
title: WindowsNetworkErrors
---

## Meaning

The sum of inbound and outbound packet error rates exceeds the configured
threshold for 15 minutes. The default is one error per second per interface.
The alert identifies the interface using its `nic` label.

## Impact

Packet errors may cause retransmissions, lower throughput, and intermittent
connection failures.

## Diagnosis

1. Inspect `windows:net_received_errors:rate` and
   `windows:net_outbound_errors:rate` separately to identify the direction.
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
zero for the rate alert to resolve.

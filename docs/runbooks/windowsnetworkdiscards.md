---
title: WindowsNetworkDiscards
---

## Meaning

The sum of inbound and outbound discarded packet rates exceeds the configured
threshold for 15 minutes. The default is one discard per second per interface.
Discards count packets dropped without a packet error, such as buffer exhaustion.

## Impact

Dropped packets can cause retransmissions, reduced throughput, and application
timeouts. Some drops may be expected under a deliberate traffic policy.

## Diagnosis

1. Inspect `windows:net_received_discarded:rate` and
   `windows:net_outbound_discarded:rate` to identify the direction and affected
   `nic`.
2. Compare traffic, directional utilization ratios, and
   `windows_net_output_queue_length_packets` with the normal baseline.
3. Inspect `Get-NetAdapterStatistics`, switch-port drops, adapter buffers,
   virtual switch counters, and network QoS policies.
4. Correlate drops with CPU pressure, traffic bursts, hypervisor contention, and
   application failures. An unknown or zero bandwidth estimate produces no
   utilization sample and should not be interpreted as zero traffic.

## Mitigation

Relieve congestion, distribute traffic, or add bandwidth. Correct buffer, driver,
or virtual switch problems when evidence identifies them. Review intentional
drop policies with the network owner before changing thresholds. Confirm discard
rates and application retransmissions return to the expected baseline.

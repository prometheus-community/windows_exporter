---
title: WindowsNetworkUtilizationHigh
---

## Meaning

Received or sent throughput on the interface `nic` exceeds the configured share
of its link bandwidth for 15 minutes. The default is 90%. The recordings
`windows:net_received_utilization:ratio` and
`windows:net_sent_utilization:ratio` divide the byte rate of each direction by
`windows_net_current_bandwidth_bytes`.

## Impact

The link is close to saturation in at least one direction. Further traffic
queues at the adapter, which raises latency and causes discards and TCP
retransmissions for every application on the host.

## Diagnosis

1. Inspect both utilization ratios to find the direction, and
   `windows:net_received_bytes:rate` and `windows:net_sent_bytes:rate` for the
   absolute volume.
2. Check `windows_net_current_bandwidth_bytes`. A link that negotiated a lower
   speed than expected, for example 100 Mbit/s instead of 1 Gbit/s, shows high
   utilization at normal traffic; compare it with `Get-NetAdapter`.
3. Correlate with `windows:net_received_discarded:rate`,
   `windows:net_outbound_discarded:rate` and
   `windows_net_output_queue_length_packets` to see whether the saturation
   already drops packets.
4. Identify the traffic with Resource Monitor, `Get-NetTCPConnection`, or
   switch-port statistics: backups, replication, file copies and monitoring
   bursts are common causes.

## Mitigation

Move or schedule the bulk transfer, apply QoS, add bandwidth, or fix a link that
negotiated the wrong speed. Confirm utilization returns below the threshold and
discards stop. Tune the threshold for interfaces that run near line rate by
design, such as storage replication links.

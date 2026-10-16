---
title: WindowsNetworkInterfaceFlapping
---

## Meaning

The operational status of the interface `nic` changed more than twice within
the rate window, and did so for two minutes. The alert counts the changes of
`windows_net_nic_operation_status{status="up"}` between `1` and `0`.

## Impact

Every transition drops the connections that use the interface and triggers
route, DHCP and DNS registration changes. A flapping uplink on a server causes
intermittent application failures that are hard to attribute.

## Diagnosis

1. Inspect `windows_net_nic_operation_status` for the interface to see the
   status it flips to; `down` and `lower layer down` point at the link,
   `dormant` at authentication such as 802.1X.
2. Check the System event log for NDIS, the adapter driver and
   `Microsoft-Windows-NetworkProfile` events, and the switch port for link
   state changes, errors and port security actions.
3. Check the cable, SFP, speed and duplex negotiation, and whether a teaming or
   failover configuration switches between members.
4. On virtual machines, check the hypervisor for virtual switch changes, live
   migrations, or a host adapter that flaps.

## Mitigation

Replace the cable or transceiver, fix the speed and duplex settings, update the
adapter driver, or correct the switch port and authentication configuration.
Exclude an interface that flaps by design, such as a Wi-Fi or VPN adapter,
with the `nicSelector`. Confirm the status stays `up`.

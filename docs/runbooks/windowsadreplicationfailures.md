---
title: WindowsADReplicationFailures
---

## Meaning

The rate of replication synchronization requests minus successful requests
exceeds the configured threshold for 15 minutes. The default is any sustained
positive failure rate. This alert is opt-in and requires the `ad` collector.

## Impact

Failed replication may leave domain controllers with stale directory data,
affecting authentication, DNS, policies, and directory changes.

## Diagnosis

1. Inspect rates of `windows_ad_replication_sync_requests_total` and
   `windows_ad_replication_sync_requests_success_total` on the affected controller.
   Counter resets around a restart may temporarily distort their difference.
2. Use the Active Directory diagnostic tools to identify partners and error codes:

   ```powershell
   repadmin /replsummary
   repadmin /showrepl
   dcdiag /test:Replications /test:DNS
   ```

3. Check Directory Service events, DNS resolution, RPC connectivity, replication
   permissions, and clock synchronization on both partners.
4. Check for an offline or decommissioned controller and review recent topology,
   firewall, or credential changes.

## Mitigation

Resolve the specific error reported by replication diagnostics. Restore partner
connectivity, DNS, time synchronization, or permissions as appropriate. Handle
retired domain controllers through the documented Active Directory removal
process. Coordinate topology and recovery changes with the directory owner.
Confirm replication succeeds and the pending queue drains.

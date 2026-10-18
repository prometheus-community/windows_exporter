---
title: WindowsDHCPScopeAlmostExhausted
---

## Meaning

The enabled DHCP scope `scope` has fewer free addresses than the configured
share of its address pool for 15 minutes. The warning fires below
`dhcpScopeFreeWarningThreshold` (10 percent by default) and the critical alert
below `dhcpScopeFreeCriticalThreshold` (2 percent); both fire below the
critical threshold, so inhibit the warning in Alertmanager. The share divides
`windows_dhcp_scope_addresses_free` by the sum of free and in-use addresses,
and the alert only counts scopes whose `windows_dhcp_scope_state` is enabled.
On a failover pair the free and in-use counts are the totals of both servers.
The alert needs the `dhcp` collector.

## Impact

Once the scope is exhausted, new clients receive no address and cannot use
the network. Existing clients keep their lease until it expires and then lose
their address too.

## Diagnosis

1. Compare the free and in-use addresses of the scope over the last days to
   see whether the exhaustion is a trend or a sudden burst, and inspect
   `windows_dhcp_scope_reserved_address` and
   `windows_dhcp_scope_pending_offers` for reservations and offers in flight.
2. On the server, read the scope and its leases:

   ```powershell
   Get-DhcpServerv4ScopeStatistics -ScopeId <scope>
   Get-DhcpServerv4Scope -ScopeId <scope> | Select-Object ScopeId, SubnetMask, StartRange, EndRange, LeaseDuration, State
   Get-DhcpServerv4Lease -ScopeId <scope> | Group-Object AddressState
   Get-DhcpServerv4ExclusionRange -ScopeId <scope>
   ```

3. Look for a long lease duration on a network with many short-lived clients,
   large exclusion ranges, a flood of leases from one device that changes its
   MAC address, or a failover partner that holds the free addresses.

## Mitigation

Shorten the lease duration for networks with transient clients, remove
exclusion ranges that are not needed, add a range through a superscope or a
larger subnet, or reclaim bad leases. On a failover pair, check the split of
the addresses between the partners. Confirm the free share stays above the
threshold.

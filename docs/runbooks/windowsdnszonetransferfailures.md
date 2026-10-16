---
title: WindowsDNSZoneTransferFailures
---

## Meaning

The DNS server has failed zone transfers for 15 minutes, measured as the rate
of `windows_dns_zone_transfer_failures_total` over `rateInterval` above
`dnsZoneTransferFailureRateThreshold`, zero by default. The counter counts
transfers this server sends as the primary to its secondaries, including
requests it refuses, and it has no zone label. The alert needs the `dns`
collector and keeps firing for `keepFiringFor` after the failures stop. Raise
the threshold on servers that refuse transfer requests from scanners all the
time.

## Impact

The secondary servers of the zone serve stale records. When the zone on a
secondary expires, it stops answering for the zone altogether. Active
Directory integrated zones replicate through Active Directory and are not
affected unless they also have secondaries.

## Diagnosis

1. Inspect `windows_dns_zone_transfer_failures_total` next to
   `windows_dns_zone_transfer_requests_received_total` and
   `windows_dns_zone_transfer_success_sent_total` to see what share of the
   transfers fail, and whether the failures started with a change.
2. On this primary, find the refused or failed requests and the zone
   settings; the DNS Server log records a refused transfer request with the
   address of the requester:

   ```powershell
   Get-WinEvent -LogName 'DNS Server' -MaxEvents 100 | Where-Object Message -match 'zone transfer'
   Get-DnsServerZone | Where-Object { $_.ZoneType -eq 'Primary' } | Get-DnsServerZoneTransferPolicy
   Get-DnsServerZone -Name '<zone>' | Select-Object ZoneName, SecureSecondaries, SecondaryServers, NotifyServers
   ```

   On the secondary, events 6522 to 6527 of the DNS Server log say why its
   transfer of the zone failed: refused by the primary, invalid response or
   an expired zone.

3. Check that the requesting secondary is in the zone transfer list, that TCP
   port 53 is open between the servers, and that the zone is not paused or
   broken. Failures from unknown addresses are refused transfers and may be
   expected.

## Mitigation

Add the secondary to the allowed servers with `Set-DnsServerPrimaryZone -SecureSecondaries`
and the server list, fix the network path, or remove a decommissioned
secondary from the zone's notify list so it stops requesting transfers.
Confirm the failure counter stops rising and the secondaries show a current
serial number.

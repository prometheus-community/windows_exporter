---
title: WindowsADReplicationPartnerFailing
---

## Meaning

Inbound replication of one naming context from one source domain controller
has failed on every attempt for at least an hour. The alert names the naming
context, the partner and its site. The alert only fires where the `ad`
collector is enabled, which is off by default, with its `replication_partners`
sub-collector, which is on by default.

## Impact

The domain controller does not receive changes of that naming context from
the partner. Depending on the topology, other paths may still deliver them, but
users and computers that authenticate against this controller can see stale
passwords, group memberships, policies or DNS records. A partner that does not
replicate for longer than the tombstone lifetime (180 days by default) can no
longer replicate at all.

## Diagnosis

1. Read the error code of the last attempt and the time of the last success:

   ```promql
   windows_ad_replication_partner_last_result
   time() - windows_ad_replication_partner_last_success_timestamp_seconds
   ```

   `net helpmsg <code>` prints the message of a code. Frequent ones are 1722
   (RPC server unavailable), 1256 (remote system not available), 8453
   (replication access denied), 8524 (DNS lookup failure), 8456/8457
   (replication disabled on the source or destination) and 8614 (tombstone
   lifetime exceeded).
2. Confirm on the domain controller:

   ```powershell
   repadmin /showrepl
   repadmin /replsummary
   dcdiag /test:Replications /test:DNS
   ```

3. Check name resolution of the partner's `_msdcs` CNAME record, RPC and LDAP
   connectivity through firewalls, time synchronization and the Directory
   Service event log on both domain controllers.
4. Check whether the partner was decommissioned without metadata cleanup or is
   shut down for a longer period.

## Mitigation

Fix the cause behind the error code, for example DNS registration, firewall
rules or time synchronization, then trigger replication with
`repadmin /syncall /AdeP` and confirm that the consecutive failures return to
0. Remove metadata of a domain controller that is gone for good with
`ntdsutil` or Active Directory Sites and Services. Do not force replication
from a partner that exceeded the tombstone lifetime; follow Microsoft's
guidance for lingering objects instead.

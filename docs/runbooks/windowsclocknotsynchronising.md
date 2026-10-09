---
title: WindowsClockNotSynchronising
---

## Meaning

Windows Time reports no active NTP time source for ten minutes. This alert is
generated only with `enableTime: true` and requires the `time` collector and its
`ntp` subcollector. Enable it on hosts expected to synchronize through NTP;
hosts intentionally using another clock source may legitimately report zero.

## Impact

The clock can drift without a working time source. Clock errors can affect
Kerberos authentication, TLS validation, scheduled work, and event timestamps.

## Diagnosis

1. Inspect `windows_time_ntp_client_time_sources` and the exporter collector
   success metric. Missing metrics indicate a different problem than a zero value.
2. On the affected host, inspect Windows Time:

   ```powershell
   Get-Service W32Time
   w32tm /query /status
   w32tm /query /source
   w32tm /query /peers
   w32tm /query /configuration
   ```

3. Check the Windows Time-Service event log, source DNS resolution, and network
   access to the source over UDP port 123.
4. For domain members, inspect the domain time hierarchy and the PDC emulator's
   upstream source. For virtual machines, check hypervisor time integration.

## Mitigation

Restore the expected Windows Time service, configured source, and UDP reachability.
Respect domain policy; configure external peers on the appropriate authoritative
server rather than overriding every domain member. After correcting the cause,
request rediscovery with `w32tm /resync /rediscover` and verify a valid source and
successful synchronization. Follow [WindowsClockSkewDetected](../windowsclockskewdetected/)
if a significant offset remains.

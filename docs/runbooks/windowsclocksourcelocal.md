---
title: WindowsClockSourceLocal
---

## Meaning

Windows Time reports `Local CMOS Clock` or `NoSync` as the source of the system
clock for ten minutes. The alert carries the source in its `type` label. It
needs the `time` collector with its `clock_source` subcollector and covers
domain members too, whose source is `NT5DS` when they synchronize from the
domain hierarchy; the NTP source count of `WindowsClockNotSynchronising` does
not see those hosts. A host that is meant to be the free-running reference of
an isolated network reports the local clock by design; exclude it with
`alerts+: { WindowsClockSourceLocal: false }` on that configuration.

## Impact

The clock drifts without an external source. Clock errors break Kerberos
authentication beyond the five minute tolerance, TLS validation, scheduled
work, log correlation and distributed systems that depend on time order.

## Diagnosis

1. Inspect `windows_time_clock_sync_source` for the host, and
   `windows_time_computed_time_offset_seconds` to see how far the clock has
   already drifted.
2. On the host, check the configured and the effective source:

   ```powershell
   w32tm /query /source
   w32tm /query /status
   w32tm /query /configuration
   Get-Service W32Time
   ```

3. `Local CMOS Clock` after a reboot often means the Windows Time service
   started before the network was up or the configured peers were unreachable.
   `NoSync` means the time provider is configured not to synchronize.
4. For domain members, check that a domain controller is reachable and that
   the PDC emulator synchronizes from an external source. For virtual machines,
   check whether hypervisor time integration overrides the configured source.
   A Hyper-V guest outside a domain that is kept in time by the integration
   service reports `NoSync` while `w32tm /query /source` shows the
   `VM IC Time Synchronization Provider`; that is by design, so turn the alert
   off for such guests with `alerts+: { WindowsClockSourceLocal: false }` or
   configure an NTP source in the guest.

## Mitigation

Configure the intended source and restart synchronization with
`w32tm /config /syncfromflags:domhier /update` on domain members or
`w32tm /config /manualpeerlist:"<servers>" /syncfromflags:manual /update` on
stand-alone hosts, then `w32tm /resync /rediscover`. Fix network access to the
peers over UDP port 123. Confirm the source changes to `NT5DS` or `NTP` and the
offset returns to normal; follow [WindowsClockSkewDetected](../windowsclockskewdetected/)
if it does not.

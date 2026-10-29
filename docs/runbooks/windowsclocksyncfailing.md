---
title: WindowsClockSyncFailing
---

## Meaning

The last synchronization attempt of Windows Time has not succeeded for an hour.
`windows_time_sync_last_result` names the outcome that W32Time reports:
`no_data` (no time source answered), `stale_data` (the answers were too old) or
`change_too_big` (the offset exceeded the configured maximum correction).
`shutdown` is rarely seen: while W32Time stops, the exporter cannot query its
status and publishes no sync status. This alert is generated only with
`collectors+: { time: true }` and requires the `time` collector and its
`sync_status` subcollector.

The result changes only with the next attempt. Stand-alone hosts with the
default `time.windows.com` peer poll once a week, so one failed attempt can
keep this alert firing for days. Hosts that intentionally run on their local
clock may never report a successful attempt; turn the alert off for them with
`alerts+: { WindowsClockSyncFailing: false }` and rely on
[WindowsClockSourceLocal](../windowsclocksourcelocal/).

## Impact

The clock runs on the last correction and drifts. Clock errors can affect
Kerberos authentication, TLS validation, scheduled work, and event timestamps.

## Diagnosis

1. Inspect `windows_time_sync_last_result{result!="success"} == 1`,
   `windows_time_sync_last_success_timestamp_seconds` and
   `windows_time_sync_source_info` for the host. The age of the last success
   is `time() - windows_time_sync_last_success_timestamp_seconds`.
2. On the affected host, inspect Windows Time:

   ```powershell
   Get-Service W32Time
   w32tm /query /status
   w32tm /query /source
   w32tm /query /peers
   w32tm /query /configuration
   ```

3. For `no_data` and `stale_data`, check the System log (source
   `Time-Service`), the `Microsoft-Windows-Time-Service/Operational` log,
   DNS resolution of the source, and network access to it over UDP port 123.
   For domain members, check the domain time hierarchy; for virtual machines,
   the hypervisor time integration.
4. For `change_too_big`, compare the local clock with the source. The offset
   is larger than `MaxPosPhaseCorrection` or `MaxNegPhaseCorrection` allow.
   By default this happens only on domain controllers (48 hours) and
   stand-alone hosts (15 hours); domain members correct any offset.

## Mitigation

Restore the time source and its reachability, then run
`w32tm /resync /rediscover` and confirm that `w32tm /query /status` shows a
recent `Last Successful Sync Time`.
For `change_too_big`, set the clock manually or let an administrator review the
phase correction limits before resynchronizing; respect domain policy instead of
overriding every domain member. Follow
[WindowsClockSkewDetected](../windowsclockskewdetected/) if a significant offset
remains.

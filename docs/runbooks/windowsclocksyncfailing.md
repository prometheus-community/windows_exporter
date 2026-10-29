---
title: WindowsClockSyncFailing
---

## Meaning

The last synchronization attempt of Windows Time has not succeeded for an hour.
The `result` label names the outcome that W32Time reports: `no_data` (no time
source answered), `stale_data` (the answers were too old), `change_too_big`
(the offset exceeded the configured maximum correction) or `shutdown` (the
service was stopping). This alert is generated only with
`collectors+: { time: true }` and requires the `time` collector and its
`sync_status` subcollector.

## Impact

The clock runs on the last correction and drifts. Clock errors can affect
Kerberos authentication, TLS validation, scheduled work, and event timestamps.

## Diagnosis

1. Inspect `windows_time_sync_last_result`,
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

3. For `no_data` and `stale_data`, check the Windows Time-Service event log,
   DNS resolution of the source, and network access to it over UDP port 123.
   For domain members, check the domain time hierarchy; for virtual machines,
   the hypervisor time integration.
4. For `change_too_big`, compare the local clock with the source. The offset
   is larger than `MaxPosPhaseCorrection` or `MaxNegPhaseCorrection` allow.

## Mitigation

Restore the time source and its reachability, then run
`w32tm /resync /rediscover` and confirm that `w32tm /query /status` shows a
recent `Last Successful Sync Time`.
For `change_too_big`, set the clock manually or let an administrator review the
phase correction limits before resynchronizing; respect domain policy instead of
overriding every domain member. Follow
[WindowsClockSkewDetected](../windowsclockskewdetected/) if a significant offset
remains.

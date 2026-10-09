---
title: WindowsClockSkewDetected
---

## Meaning

The absolute clock offset reported by Windows Time exceeds the configured
threshold for ten minutes. The default is 0.05 seconds (50 ms). This opt-in alert
requires the `time` collector with its `ntp` subcollector. The offset is relative
to the selected source and does not independently validate that source's accuracy.

## Impact

Clock skew can affect authentication, certificate validity, scheduled work, and
the ordering of events across hosts. The default threshold is an early warning;
application tolerances vary.

## Diagnosis

1. Inspect `windows_time_computed_time_offset_seconds`, source count, and
   `windows_time_ntp_round_trip_delay_seconds` for the affected host.
2. Check synchronization state and the selected source:

   ```powershell
   w32tm /query /status
   w32tm /query /source
   w32tm /query /configuration
   ```

3. Compare with a trusted source using
   `w32tm /stripchart /computer:<trusted-time-server> /samples:5 /dataonly`.
   Check source health, network delay, Windows Time-Service events, and domain
   or hypervisor time configuration.
4. Consult [WindowsClockNotSynchronising](../windowsclocknotsynchronising/) when
   no NTP source is available.

## Mitigation

Restore a healthy source and correct time configuration before requesting
`w32tm /resync /rediscover`. Coordinate large clock corrections with application
owners because abrupt changes can disrupt running workloads. Avoid setting the
clock manually as a recurring fix. Confirm the reported offset decreases and
remains within the application's required tolerance.

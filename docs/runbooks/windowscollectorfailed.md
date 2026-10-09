---
title: WindowsCollectorFailed
---

## Meaning

The collector identified by the `collector` label has reported
`windows_exporter_collector_success == 0` for five minutes.

## Impact

Metrics from this collector are incomplete or missing. Alerts relying on those
metrics may stop evaluating even though the exporter remains reachable.

## Diagnosis

1. Find the failing collector and compare the affected hosts:

   ```promql
   windows_exporter_collector_success{job="windows_exporter"} == 0
   ```

2. Review exporter logs for that collector's error and duration. Check whether
   the failure started after an exporter upgrade, Windows update, or account
   permission change.
3. Check the collector's documentation for supported Windows roles, required
   privileges, and performance counters. Confirm the required service or role
   exists and is running on the host.
4. Compare `windows_exporter_collector_duration_seconds` with the scrape timeout.
   Test the relevant performance counters locally with Windows Performance
   Monitor.

## Mitigation

Restore the required service, privileges, or performance counter availability.
Limit expensive collector queries when collection exceeds the timeout, and
adjust timeouts only after measuring normal collection time. Disable a collector
on hosts where its role is absent. If the problem began with an upgrade, check
release notes and apply a fix or roll back through the normal change process.
Confirm the collector succeeds and its dependent metrics resume.

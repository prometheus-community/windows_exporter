---
title: WindowsCollectorTimeout
---

## Meaning

The collector identified by the `collector` label has reported
`windows_exporter_collector_timeout == 1` for ten minutes. The collector did
not finish within the scrape timeout, or its call from an earlier scrape is
still running and later scrapes skip it until it returns.

## Impact

The metrics of this collector are missing or stale on every scrape that times
out, so dashboards show gaps and alerts on those metrics stop evaluating. A
collector that never returns can hold a scrape slot and delay the other
collectors of the host.

## Diagnosis

1. Find the collector and compare how long it usually takes:

   ```promql
   windows_exporter_collector_timeout{job="windows_exporter"} == 1
   windows_exporter_collector_duration_seconds{job="windows_exporter"}
   ```

2. Compare the duration with the scrape timeout of the job. The exporter
   subtracts `--scrape.timeout-margin` from the `scrape_timeout` that
   Prometheus sends.
3. Check the exporter log for the collector. WMI and Management Infrastructure
   queries, the Task Scheduler and Windows Update are slow on hosts with many
   objects or a busy WMI provider; the collector's documentation lists the
   expensive options.
4. Check whether the host was under CPU, disk or memory pressure at the same
   time, and whether a Windows service the collector depends on hangs, for
   example `Winmgmt` for WMI-based collectors.

## Mitigation

Raise the `scrape_timeout` of the job within its `scrape_interval`, narrow the
collector with its include and exclude flags, or move the collector to a
separate job with a longer timeout. Restart a hanging provider service when the
log identifies it. Disable the collector on hosts where its data is not needed.
Confirm `windows_exporter_collector_timeout` returns to `0` and the collector's
metrics are published again.

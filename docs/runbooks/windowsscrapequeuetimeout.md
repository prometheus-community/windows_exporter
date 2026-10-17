---
title: WindowsScrapeQueueTimeout
---

## Meaning

`windows_exporter_scrape_queue_timeout == 1` for ten minutes. Scrapes run one
at a time; a scrape whose timeout ends while it waits for an earlier scrape to
finish runs no collector and reports this metric as `1`. On such a scrape
`windows_exporter_collector_success` is `0` for every collector.

## Impact

The affected scrapes return no collector metrics, so the host has gaps in all
of its series and `WindowsCollectorFailed` can fire for every collector at
once. The exporter itself is reachable.

## Diagnosis

1. Confirm the queue timeout and find the scrape that holds the slot:

   ```promql
   windows_exporter_scrape_queue_timeout{job="windows_exporter"} == 1
   windows_exporter_collector_timeout{job="windows_exporter"} == 1
   scrape_duration_seconds{job="windows_exporter"}
   ```

2. Check whether more than one Prometheus server or job scrapes the same
   exporter, or whether a job scrapes it more often than its scrape duration
   allows. Compare `scrape_duration_seconds` with the `scrape_interval`.
3. Identify the slow collector with `windows_exporter_collector_duration_seconds`
   and follow `WindowsCollectorTimeout` for it.

## Mitigation

Scrape each exporter from one job, or give the jobs a `scrape_interval` longer
than a full scrape takes. Shorten the slow collector or raise the
`scrape_timeout` so a scrape finishes before the next one starts. Confirm
`windows_exporter_scrape_queue_timeout` stays `0` and collector success returns
to `1`.

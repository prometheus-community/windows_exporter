---
title: WindowsDiskFillingUp
---

## Meaning

Free space is below the warning threshold and a linear forecast predicts the
volume will fill within the configured horizon. Defaults use six hours of
history, a 24-hour prediction horizon, and a one-hour alert duration.

## Impact

Continued growth can exhaust the volume before routine capacity maintenance,
causing write failures or an application outage.

## Diagnosis

1. Inspect the free-space trend for the volume and verify local free space with
   `Get-Volume`. Windows disk space counters can lag by 10–15 minutes.
2. Inspect the forecast using the configured selector and window, for example:

   ```promql
   predict_linear(windows_logical_disk_free_bytes{job="windows_exporter", instance="host:9182", volume="C:"}[6h], 24 * 3600)
   ```

3. Identify the application, logs, backups, or temporary data causing growth.
   Check whether a recent bulk write distorted the trend or growth is sustained.
4. Compare absolute remaining bytes with the application's write rate. A linear
   forecast is an estimate and does not model scheduled cleanup or sudden growth.

## Mitigation

Restore retention and rotation, reduce write growth, or expand the volume before
it fills. Coordinate data cleanup with the owner. If the forecast reflects a
known temporary job, verify its end time and remaining capacity before silencing
the alert. Follow [WindowsDiskAlmostFull](../windowsdiskalmostfull/) when free
space is already critically low.

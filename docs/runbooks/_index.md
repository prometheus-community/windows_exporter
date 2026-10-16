---
title: Windows exporter runbooks
---

These runbooks describe alerts from the [Windows exporter mixin](https://github.com/prometheus-community/windows_exporter/tree/master/contrib/mixin).
Use the alert's `job`, `instance`, and any `cluster` or environment labels to find
the affected host. Disk alerts also identify a `volume`; network alerts identify
a `nic`. Adjust example PromQL selectors to match your scrape configuration.

The default thresholds are starting points. Check the generated alert rules for
your configured thresholds and durations. The rules are grouped by collector
area, and each alert is generated only when the collectors it reads are enabled
in the mixin configuration. Active Directory and NTP clock alerts need the `ad`
and `time` collectors, which are off by default.

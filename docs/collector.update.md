# update collector

The update collector exposes the Windows Update service metrics. Note that the Windows Update service (`wuauserv`) must be running, else metric collection will fail.

The Windows Update service is responsible for managing the installation of updates for the operating system and other Microsoft software. The service can be configured to automatically download and install updates, or to notify the user when updates are available.


|                     |                        |
|---------------------|------------------------|
| Metric name prefix  | `update`               |
| Data source         | Windows Update service |
| Enabled by default? | No                     |

## Flags

### `--collector.updates.online`
Whether to search for updates online. If set to `false`, the collector will only list updates that are already found by the Windows Update service.
Set to `true` to search for updates online, which will take longer to complete.

### `--collector.updates.scrape-interval`
Define the interval of scraping Windows Update information

## Metrics

| Name                           | Description                                   | Type  | Labels                        |
|--------------------------------|-----------------------------------------------|-------|-------------------------------|
| `windows_updates_pending_info` | Expose information single pending update item | gauge | `category`,`severity`,`title` |
| `windows_updates_scrape_query_duration_seconds` | Duration of the last scrape query to the Windows Update API | gauge |  |
| `windows_updates_scrape_timestamp_seconds` | Timestamp of the last scrape | gauge |  |

### Example metrics
```
# HELP windows_updates_pending Pending Windows Updates
# TYPE windows_updates_pending gauge
windows_updates_pending{category="Drivers",severity="",title="Intel Corporation - Bluetooth - 23.60.5.10"} 1
# HELP windows_updates_scrape_query_duration_seconds Duration of the last scrape query to the Windows Update API
# TYPE windows_updates_scrape_query_duration_seconds gauge
windows_updates_scrape_query_duration_seconds 2.8161838
# HELP windows_updates_scrape_timestamp_seconds Timestamp of the last scrape
# TYPE windows_updates_scrape_timestamp_seconds gauge
windows_updates_scrape_timestamp_seconds 1.727539734e+09
```

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

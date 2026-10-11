# update collector

The update collector exposes the Windows Update service metrics. Note that the Windows Update service (`wuauserv`) must be running, else metric collection will fail.

The Windows Update service is responsible for managing the installation of updates for the operating system and other Microsoft software. The service can be configured to automatically download and install updates, or to notify the user when updates are available.


|                     |                        |
|---------------------|------------------------|
| Metric name prefix  | `update`               |
| Data source         | Windows Update service |
| Enabled by default? | No                     |


## Flags

### `--collector.update.online`
Whether to search for updates online. If set to `false` via `--no-collector.update.online`, the collector will only list updates that are already found by the Windows Update service.
Set to `true` via `--collector.update.online` to search for updates online, which will take longer to complete.

Default: `false`

### `--collector.update.scrape-interval`
Define the interval of scraping Windows Update information

Default: `6h0m0s`

## Metrics

Pending updates include optional updates, such as the monthly non-security preview update. The `deployment_action` label is `installation` for regular updates and `optional_installation` for optional ones. Filter with `deployment_action="installation"` to ignore optional updates.

| Name                                           | Description                                                      | Type  | Labels                                                            |
|------------------------------------------------|------------------------------------------------------------------|-------|-------------------------------------------------------------------|
| `windows_update_pending_info`                  | Expose information for a single pending update item              | gauge | `id`,`revision`,`category`,`severity`,`title`,`deployment_action` |
| `windows_update_pending_published_timestamp`   | Expose last published timestamp for a single pending update item | gauge | `id`,`revision`                                                   |
| `windows_update_scrape_query_duration_seconds` | Duration of the last scrape query to the Windows Update API      | gauge |                                                                   |
| `windows_update_scrape_timestamp_seconds`      | Timestamp of the last scrape                                     | gauge |                                                                   |

### Example metrics
```
# HELP windows_update_pending_info Expose information for a single pending update item
# TYPE windows_update_pending_info gauge
windows_update_pending_info{category="Definition Updates",deployment_action="installation",id="a32ca1d0-ddd4-486b-b708-d941db4f1051",revision="204",severity="",title="Update for Windows Security platform - KB5007651 (Version 10.0.27840.1000)"} 1
windows_update_pending_info{category="Definition Updates",deployment_action="installation",id="b50a64de-a0bb-465b-9842-9963b6eee21e",revision="200",severity="",title="Security Intelligence Update for Microsoft Defender Antivirus - KB2267602 (Version 1.429.146.0) - Current Channel (Broad)"} 1
windows_update_pending_info{category="Updates",deployment_action="optional_installation",id="4c5a2079-dd6c-4b78-a20b-07b2d5ea2217",revision="1",severity="",title="2026-09 Preview Update (KB5124010) (26300.9550)"} 1
# HELP windows_update_pending_published_timestamp Expose last published timestamp for a single pending update item
# TYPE windows_update_pending_published_timestamp gauge
windows_update_pending_published_timestamp{id="a32ca1d0-ddd4-486b-b708-d941db4f1051",revision="204"} 1.747872e+09
windows_update_pending_published_timestamp{id="b50a64de-a0bb-465b-9842-9963b6eee21e",revision="200"} 1.7479584e+09
windows_update_pending_published_timestamp{id="4c5a2079-dd6c-4b78-a20b-07b2d5ea2217",revision="1"} 1.7900352e+09
# HELP windows_update_scrape_query_duration_seconds Duration of the last scrape query to the Windows Update API
# TYPE windows_update_scrape_query_duration_seconds gauge
windows_update_scrape_query_duration_seconds 2.8161838
# HELP windows_update_scrape_timestamp_seconds Timestamp of the last scrape
# TYPE windows_update_scrape_timestamp_seconds gauge
windows_update_scrape_timestamp_seconds 1.727539734e+09
```

## Useful queries

Add extended information like cmdline or owner to other process metrics.

```
windows_update_pending_published_timestamp * on(id, revision) group_left(severity, title) windows_update_pending_info
```

Count pending updates without optional updates, such as the monthly non-security preview update.

```
count(windows_update_pending_info{deployment_action="installation"})
```

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

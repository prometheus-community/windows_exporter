# pagefile collector

The pagefile collector exposes metrics about the pagefile usage

|||
-|-
Metric name prefix  | `pagefile`
Data source         | Performance counter `Paging File` (`% Usage` and its base), `GetPerformanceInfo` (page size)
Enabled by default? | No

## Flags

None

## Metrics

| Name                           | Description                                                                                                                 | Type  | Labels |
|--------------------------------|-----------------------------------------------------------------------------------------------------------------------------|-------|--------|
| `windows_pagefile_free_bytes`  | Number of bytes that can be mapped into the operating system paging files without causing any other pages to be swapped out | gauge | `file` |
| `windows_pagefile_limit_bytes` | Current size of the paging file in bytes                                                                                    | gauge | `file` |

Both values come from the `% Usage` counter: its raw value is the number of pages in use and its base is the size of
the paging file in pages. They are multiplied by the system page size. A paging file without a valid size is left out.

### Example metric

```
# HELP windows_pagefile_free_bytes Number of bytes that can be mapped into the operating system paging files without causing any other pages to be swapped out
# TYPE windows_pagefile_free_bytes gauge
windows_pagefile_free_bytes{file="C:\\pagefile.sys"} 6.025797632e+09
# HELP windows_pagefile_limit_bytes Current size of the paging file in bytes
# TYPE windows_pagefile_limit_bytes gauge
windows_pagefile_limit_bytes{file="C:\\pagefile.sys"} 6.442450944e+09
```

## Useful queries
_This collector does not yet have useful queries, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

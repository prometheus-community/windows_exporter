# pagefile collector

The pagefile collector exposes metrics about the pagefile usage

|||
-|-
Metric name prefix  | `pagefile`
Classes             | [`Win32_OperatingSystem`](https://msdn.microsoft.com/en-us/library/aa394239)
Enabled by default? | Yes

## Flags

None

## Metrics

| Name                           | Description                                                                                                                 | Type  | Labels |
|--------------------------------|-----------------------------------------------------------------------------------------------------------------------------|-------|--------|
| `windows_pagefile_free_bytes`  | Number of bytes that can be mapped into the operating system paging files without causing any other pages to be swapped out | gauge | `file` |
| `windows_pagefile_limit_bytes` | Number of bytes that can be stored in the operating system paging files. 0 (zero) indicates that there are no paging files  | gauge | `file` |


### Example metric

```
# HELP windows_pagefile_free_bytes OperatingSystem.FreeSpaceInPagingFiles
# TYPE windows_pagefile_free_bytes gauge
windows_pagefile_free_bytes{file="C:\\pagefile.sys"} 6.025797632e+09
# HELP windows_pagefile_limit_bytes OperatingSystem.SizeStoredInPagingFiles
# TYPE windows_pagefile_limit_bytes gauge
windows_pagefile_limit_bytes{file="C:\\pagefile.sys"} 6.442450944e+09
```

## Useful queries
_This collector does not yet have useful queries, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

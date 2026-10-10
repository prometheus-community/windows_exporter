# mscluster agent notes

- Subcollectors are moving from the `root\MSCluster` WMI provider (ClusWMI.dll)
  to ClusAPI through [`internal/headers/clusapi`](../../headers/clusapi). The
  WMI provider is a thin layer over ClusAPI: most WMI properties are common
  properties from `CLUSCTL_<OBJECT>_GET_RO_COMMON_PROPERTIES` and
  `CLUSCTL_<OBJECT>_GET_COMMON_PROPERTIES`; `Characteristics`, `Flags` and
  `State` come from their own control codes and `GetCluster<Object>State`.
- Keep WMI types when publishing: `sint32` properties (for example
  `FailbackWindowStart`) are converted from the raw 32-bit value with
  `int32`, `uint32` properties keep the unsigned bit pattern.
- Each ClusAPI subcollector owns its own `clusapi.Cluster`. `Close` and `Build`
  go through `closeSources`, which keeps a source whose `Close` failed so a
  later call can retry. Grafana Alloy calls `Build`/`Close` repeatedly.
- Every native subcollector has a fixture test that checks gathered metric
  families and a parity test against the WMI class. The CI cluster
  (`setup-mscluster` in `.github/workflows/ci.yml`) is a single node without
  storage or administrative access point, so it has no resources and no
  cluster shared volumes. Parity tests must accept empty sets in that case.
- Use `ClusterOpenEnum` and the `Open*Ex` functions (Windows Server 2008 R2+),
  not `ClusterOpenEnumEx` (Windows Server 2016+), to keep Windows Server
  2012 R2 support.

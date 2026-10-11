# mscluster agent notes

- All subcollectors read through ClusAPI
  ([`internal/headers/clusapi`](../../headers/clusapi)) instead of the
  `root\MSCluster` WMI provider (ClusWMI.dll). The WMI provider is a thin layer
  over ClusAPI: most WMI properties are common properties from
  `CLUSCTL_<OBJECT>_GET_RO_COMMON_PROPERTIES` and
  `CLUSCTL_<OBJECT>_GET_COMMON_PROPERTIES`; `Characteristics`, `Flags` and
  `State` come from their own control codes and `GetCluster<Object>State`.
  Some WMI values are synthesized or never filled (published as 0 from NULL);
  the parity tests show which ones, and the collector reproduces them.
- Keep WMI types when publishing: `sint32` properties (for example
  `FailbackWindowStart`) are converted from the raw 32-bit value with
  `int32`, `uint32` properties keep the unsigned bit pattern. Unknown states
  (-1) are values, as in WMI, not errors.
- Each subcollector owns its own `clusapi.Cluster`. A `Cluster` runs one call
  at a time without holding its lock: ClusAPI RPCs cannot be cancelled, so a
  second call fails with `clusapi.ErrBusy`, and `Close` leaves the handle to
  the running call. `callSource` stops waiting at the scrape budget, so one
  stuck RPC does not hold `Collect` and the other subcollectors.
- `closeSources` drops a source even when its `Close` fails, so `Build` can
  recreate it. Grafana Alloy calls `Build`/`Close` repeatedly.
- Every subcollector has a fixture test that checks gathered metric families
  and a parity test against its WMI class. The CI cluster (`setup-mscluster`
  in `.github/workflows/ci.yml`) is a single node without storage or
  administrative access point. It has online, offline and failed resources in
  two groups (`CIResources`, `CIFailedResources`) on a private network of an
  internal Hyper-V switch; the parity tests require them. It has no clustered
  disks, so shared volume parity still compares empty sets.
- Property lists returned by the cluster service end with an extra
  `CLUSPROP_SYNTAX_ENDMARK` after the last value list; fixtures built only from
  the documentation miss this. The parser validates only the structure and
  accepts zero padding. `TestDumpPropertyLists` logs real lists as hex when
  `WINDOWS_EXPORTER_TEST_CLUSAPI_DUMP` is set on a cluster node.
- Each `Cluster` owns one control buffer (`controlScratch`). Like the handle,
  it belongs to the running call, so calls never share it and no `sync.Pool`
  is needed. `ParseProperties` returns `Property.Data` slices into it: copy
  values out (`copyResourceProperties`, `readCommonProperties`) before the
  next control, and never keep a parsed list past that point. `get` zeroes
  the bytes, so a reused buffer reads like a new one.
- `clusapi.Open` reports an unconfigured cluster or a missing ClusAPI with
  `errors.ErrUnsupported`; the shared collector test treats that as an
  unsupported role on hosts where `mscluster` is not a required fixture.
- Use `ClusterOpenEnum` and the `Open*Ex` functions (Windows Server 2008 R2+),
  not `ClusterOpenEnumEx` (Windows Server 2016+), to keep Windows Server
  2012 R2 support.
- Cluster, network, node, group and resource metrics are reused between
  scrapes through `internal/metriccache` when every input is unchanged. A
  cache value must hold every input of the cached series: field values and
  whether each field is published (`appendObjectFieldsKey`, which resolves the
  build-dependent rules of `objectField`), owner node and node names. Objects
  with missing fields are not cached, so they report the same errors on every
  scrape. Shared volume sizes change constantly and are not cached.

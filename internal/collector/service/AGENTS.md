# service collector

- `EnumServicesStatusEx` returns state and PID of all services in one call. The
  configuration needs `OpenService`, `QueryServiceConfig` and
  `CloseServiceHandle` per service, which costs CPU in services.exe as well, so
  it is cached in [`config_cache.go`](config_cache.go). Never cache state or PID.
- `NotifyServiceStatusChange` does not report configuration changes. The cache
  watches `HKLM\SYSTEM\CurrentControlSet\Services` with
  `RegNotifyChangeKeyValue` instead. Arm it with `REG_NOTIFY_THREAD_AGNOSTIC`:
  without it the registration ends when the arming OS thread exits, and Go
  goroutines are not bound to one.
- The delayed auto-start flag and the triggers are queried with
  `QueryServiceConfig2` on the same handle and cached with the configuration.
  The SCM ignores the delayed flag unless the start type is auto, so it is only
  queried for auto-start services. `x/sys/windows` has no
  `SERVICE_TRIGGER_INFO`; the mirror in `service.go` only reads `cTriggers`.
- The delayed flag is the `DelayedAutostart` value of the service key, so a
  change updates the key's last write time. Triggers are stored in the
  `TriggerInfo` subkey. Adding the first trigger creates it, which updates the
  last write time of the service key (`TestCollectorPicksUpConfigChanges`
  checks this). That a later trigger change made with `sc triggerinfo` also
  does is assumed, not verified: the SCM is believed to rewrite the whole
  `TriggerInfo` subkey. If it only changes deeper keys, the cache picks the
  change up after at most `configCacheMaxAge`.
- The subtree notification fires every few seconds on a normal machine (W32Time,
  Tcpip, VSS, driver `Enum` keys, Hyper-V). Treat it as a hint only and compare
  the last write time of each service's own key, which changes only for its own
  values and direct subkeys.
- Registry write times have a resolution of one clock tick, so two writes in
  the same tick leave the time unchanged. Only cache a configuration once its
  key is older than `configCacheMinKeyAge`, and never cache an unavailable one
  (for example a missing MUI display name), which can change without a write.
- The cache tests use a temporary key below `HKEY_CURRENT_USER`, so they run
  without elevation. Registry write times have a coarse resolution; sleep before
  a write that must get a newer time.
- `TestCollectorPicksUpConfigChanges` creates a real service and changes it
  through the SCM. It needs an elevated process: it skips locally without one
  and fails in CI, where `service` is in `WINDOWS_EXPORTER_TEST_COLLECTORS`.
- `EnumServicesStatusEx` fills at most 256 KiB per call (`BOUNDED_DWORD_256K`
  in [MS-SCMR](https://winprotocoldocs-bhdugrdyduf5h2e4.b02.azurefd.net/MS-SCMR/%5bMS-SCMR%5d.pdf)).
  `queryAllServices` passes a resume index and keeps calling while the SCM
  returns `ERROR_MORE_DATA` with services; it only grows the buffer, up to
  that limit, when not one service fits. RDS hosts add per-user service
  instances for every session, so their list can exceed one call.
  `TestQueryAllServicesResumes` forces several calls with a small buffer.
- `OpenService` fails with `ERROR_ACCESS_DENIED` when the service's DACL denies
  `SERVICE_QUERY_CONFIG` to the exporter's account, and with
  `ERROR_SERVICE_DOES_NOT_EXIST` for a service deleted after the
  enumeration. The first publishes the state without the configuration
  (`TestCollectorPublishesServiceWithDeniedConfig`); the second skips the
  service with a debug log.

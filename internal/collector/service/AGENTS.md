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

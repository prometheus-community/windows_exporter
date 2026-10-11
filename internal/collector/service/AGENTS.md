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
- The subtree notification fires every few seconds on a normal machine (W32Time,
  Tcpip, VSS, driver `Enum` keys, Hyper-V). Treat it as a hint only and compare
  the last write time of each service's own key, which changes only for its own
  values and direct subkeys.
- The cache tests use a temporary key below `HKEY_CURRENT_USER`, so they run
  without elevation. Registry write times have a coarse resolution; sleep before
  a write that must get a newer time.
- `TestCollectorPicksUpConfigChanges` creates a real service and changes it
  through the SCM. It needs an elevated process: it skips locally without one
  and fails in CI, where `service` is in `WINDOWS_EXPORTER_TEST_COLLECTORS`.

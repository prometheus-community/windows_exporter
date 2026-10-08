(import '../mixin.libsonnet') {
  _config+:: {
    windowsExporterSelector: 'job=~"windows.*", cluster="test"',
    volumeSelector: 'volume!~"HarddiskVolume.*"',
    nicSelector: 'nic!="Loopback"',
    rateInterval: '2m',
    cpuHighUsageThreshold: 80,
    memoryHighUsageThreshold: 85,
    diskFreeWarningThreshold: 20,
    diskFreeCriticalThreshold: 10,
    criticalSeverity: 'page',
    enableActiveDirectory: true,
    adPendingReplicationThreshold: 10,
    adReplicationFailureRateThreshold: 0.5,
  },
}

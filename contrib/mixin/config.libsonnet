{
  _config+:: {
    // Prometheus label matchers, without enclosing braces.
    windowsExporterSelector: 'job="windows_exporter"',
    volumeSelector: 'volume!=""',
    nicSelector: 'nic!=""',
    rateInterval: '5m',

    // Opt in only for targets with the ad collector enabled.
    enableActiveDirectory: false,

    cpuHighUsageThreshold: 90,
    memoryHighUsageThreshold: 90,
    diskFreeWarningThreshold: 10,
    diskFreeCriticalThreshold: 5,
    diskPredictionWindow: '6h',
    diskPredictionHours: 24,
    criticalSeverity: 'critical',
    adPendingReplicationThreshold: 50,
    adReplicationFailureRateThreshold: 0,
  },
}

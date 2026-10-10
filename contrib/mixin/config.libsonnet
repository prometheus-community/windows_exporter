{
  _config+:: {
    // Prometheus label matchers, without enclosing braces.
    windowsExporterSelector: 'job="windows_exporter"',
    volumeSelector: 'volume!=""',
    nicSelector: 'nic!=""',
    rateInterval: '5m',

    // Opt in only for targets with the ad collector enabled.
    enableActiveDirectory: false,
    // Requires the time collector with its ntp subcollector enabled.
    enableTime: false,

    // Collectors enabled on the windows_exporter targets, by collector name.
    // The dashboard only renders the tabs, rows and panels of enabled
    // collectors. The defaults match the exporter's default collectors.
    collectors: {
      cpu: true,
      diskdrive: false,
      gpu: false,
      hyperv: false,
      logical_disk: true,
      memory: true,
      net: true,
      os: true,
      physical_disk: true,
      process: false,
      scheduled_task: false,
      service: true,
      smb: false,
      smbclient: false,
      system: true,
      tcp: false,
      time: $._config.enableTime,
      udp: false,
      update: false,
    },

    runbookURLPattern: 'https://prometheus-community.github.io/windows_exporter/runbooks/%s/',

    cpuHighUsageThreshold: 90,
    memoryHighUsageThreshold: 90,
    memoryCommitHighUsageThreshold: 90,
    diskLatencyThresholdSeconds: 0.05,
    networkErrorRateThreshold: 1,
    networkDiscardRateThreshold: 1,
    clockOffsetThresholdSeconds: 0.05,
    diskFreeWarningThreshold: 10,
    diskFreeCriticalThreshold: 5,
    diskPredictionWindow: '6h',
    diskPredictionHours: 24,
    criticalSeverity: 'critical',
    adPendingReplicationThreshold: 50,
    adReplicationFailureRateThreshold: 0,
  },
}

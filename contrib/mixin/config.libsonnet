{
  _config+:: {
    // Prometheus label matchers, without enclosing braces.
    windowsExporterSelector: 'job="windows_exporter"',
    volumeSelector: 'volume!=""',
    nicSelector: 'nic!=""',
    serviceSelector: 'service!=""',
    taskSelector: 'task!=""',
    // ClusterPerformanceHistory is the volume of the cluster's own health history.
    sharedVolumeSelector: 'shared_volume!="ClusterPerformanceHistory"',
    rateInterval: '5m',

    // Opt in only for targets with the ad collector enabled. Sets collectors.ad.
    enableActiveDirectory: false,
    // Requires the time collector with its ntp subcollector enabled. Sets collectors.time.
    enableTime: false,

    // Grafana dashboard metadata. A UID is the v2 metadata.name; change the
    // UIDs to import the dashboards next to the sample ones.
    // Every tab also renders as a dashboard of its own with the UID
    // <dashboardUIDPrefix>-<tab> and the title "<dashboardTitlePrefix> / <tab>",
    // for example windows-exporter-cpu and "Windows Exporter / CPU".
    enableTabDashboards: true,
    dashboardUIDPrefix: 'windows-exporter',
    dashboardTitlePrefix: 'Windows Exporter',
    // The full dashboard has all tabs.
    dashboardUID: self.dashboardUIDPrefix + '-full',
    dashboardTitle: self.dashboardTitlePrefix + ' / Full',
    dashboardDescription: 'Fleet overview and per-host details for Windows hosts monitored by windows_exporter.',
    // The dashboards link to each other through their tags.
    dashboardTags: ['prometheus', 'windows', 'windows_exporter'],

    // Dashboard tabs that don't belong to a single collector. Turn them off
    // for a compact dashboard with only the collector tabs.
    enableFleetTab: true,
    enableOverviewTab: true,
    // Tabs, by the slug in their UID, that render only as dashboards of their
    // own, even with enableTabDashboards: false, and not in the full
    // dashboard. Set it to [] to add them to the full dashboard.
    separateDashboardTabs: ['mssql'],

    // Collectors enabled on the windows_exporter targets, by collector name.
    // The dashboard only renders the tabs, rows and panels of enabled
    // collectors, and the alerts only the rules whose collectors are enabled.
    // The defaults match the exporter's default collectors.
    collectors: {
      ad: $._config.enableActiveDirectory,
      cpu: true,
      dhcp: false,
      diskdrive: false,
      dns: false,
      gpu: false,
      hyperv: false,
      license: false,
      logical_disk: true,
      memory: true,
      mscluster: false,
      mssql: false,
      net: true,
      os: true,
      pagefile: false,
      physical_disk: true,
      process: false,
      scheduled_task: false,
      service: true,
      smb: false,
      smbclient: false,
      storage_spaces: false,
      system: true,
      tcp: false,
      time: $._config.enableTime,
      udp: false,
      update: false,
    },

    // Alerts by name. Set one to false to leave it out of the generated
    // rules, for example alerts+: { WindowsHostRebooted: false }. Names that
    // are not listed here fail the build; names left out of a replaced object
    // stay enabled.
    alerts: {
      WindowsExporterDown: true,
      WindowsCollectorFailed: true,
      WindowsCollectorTimeout: true,
      WindowsScrapeQueueTimeout: true,
      WindowsCPUHighUsage: true,
      WindowsCPUSaturation: true,
      WindowsMemoryHighUsage: true,
      WindowsMemoryCommitHighUsage: true,
      WindowsMemoryPagingHigh: true,
      WindowsDiskReadLatencyHigh: true,
      WindowsDiskWriteLatencyHigh: true,
      WindowsDiskAlmostFull: true,
      WindowsDiskFillingUp: true,
      WindowsNetworkErrors: true,
      WindowsNetworkDiscards: true,
      WindowsNetworkUtilizationHigh: true,
      WindowsNetworkInterfaceFlapping: true,
      WindowsServiceNotRunning: true,
      WindowsServiceCrashLooping: true,
      WindowsHostRebooted: true,
      WindowsADPendingReplication: true,
      WindowsADReplicationFailures: true,
      WindowsClockNotSynchronising: true,
      WindowsClockSkewDetected: true,
      WindowsClockSourceLocal: true,
      WindowsDiskDriveNotOK: true,
      WindowsPagefileAlmostFull: true,
      WindowsLicenseNotGenuine: true,
      WindowsScheduledTaskFailed: true,
      WindowsUpdatesPending: true,
      WindowsHyperVWMIUnhealthy: true,
      WindowsHyperVReplicationCritical: true,
      WindowsHyperVReplicationWarning: true,
      WindowsClusterNodeDown: true,
      WindowsClusterResourceFailed: true,
      WindowsClusterResourceGroupFailed: true,
      WindowsClusterSharedVolumeAlmostFull: true,
      WindowsStoragePoolUnhealthy: true,
      WindowsStorageVirtualDiskUnhealthy: true,
      WindowsDHCPScopeAlmostExhausted: true,
      WindowsDNSZoneTransferFailures: true,
    },

    runbookURLPattern: 'https://prometheus-community.github.io/windows_exporter/runbooks/%s/',

    // Threshold alerts keep firing this long after their condition clears,
    // so a value oscillating around a threshold does not flap. Needs
    // Prometheus 2.42 or later; set it to '' to leave keep_firing_for out.
    keepFiringFor: '5m',

    cpuHighUsageThreshold: 90,
    // Threads waiting for a processor, per logical processor.
    cpuQueuePerCoreThreshold: 2,
    memoryHighUsageThreshold: 90,
    memoryCommitHighUsageThreshold: 90,
    // Pages read from or written to disk per second.
    memoryPagingThreshold: 1000,
    diskLatencyThresholdSeconds: 0.05,
    // Average latency per I/O above which a Hyper-V virtual disk counts as slow.
    hypervDiskLatencyThresholdSeconds: 0.01,
    // Percent of all packets on the interface.
    networkErrorRatioThreshold: 1,
    networkDiscardRatioThreshold: 1,
    // Percent of the interface bandwidth, in either direction.
    networkUtilizationThreshold: 90,
    clockOffsetThresholdSeconds: 0.05,
    diskFreeWarningThreshold: 10,
    diskFreeCriticalThreshold: 5,
    diskPredictionWindow: '6h',
    diskPredictionHours: 24,
    // Restarts of one service within 15 minutes.
    serviceRestartThreshold: 2,
    criticalSeverity: 'critical',
    infoSeverity: 'info',
    adPendingReplicationThreshold: 50,
    adReplicationFailureRateThreshold: 0,
    // Percent of the page file in use.
    pagefileUsageThreshold: 80,
    // Percent of free addresses in an active DHCP scope.
    dhcpScopeFreeWarningThreshold: 10,
    dhcpScopeFreeCriticalThreshold: 2,
    // How long updates may wait for installation before WindowsUpdatesPending fires.
    updatePendingDuration: '7d',
    // Failed zone transfers per second.
    dnsZoneTransferFailureRateThreshold: 0,
  },
}

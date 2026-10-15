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
    // collectors. The defaults match the exporter's default collectors.
    collectors: {
      cpu: true,
      diskdrive: false,
      gpu: false,
      hyperv: false,
      logical_disk: true,
      memory: true,
      mssql: false,
      net: true,
      os: true,
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

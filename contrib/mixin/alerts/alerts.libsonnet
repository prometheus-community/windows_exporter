{
  local c = self._config,
  local q = (import '../lib/queries.libsonnet')(c),
  local known = std.objectFields((import '../config.libsonnet')._config.alerts),
  local unknown = std.setDiff(std.set(std.objectFields(c.alerts)), std.set(known)),
  assert std.length(unknown) == 0 : 'Unknown alerts in _config.alerts: ' + std.join(', ', unknown),
  assert !std.objectHas(c, 'networkErrorRateThreshold') && !std.objectHas(c, 'networkDiscardRateThreshold') :
         'networkErrorRateThreshold and networkDiscardRateThreshold (packets per second) were replaced by networkErrorRatioThreshold and networkDiscardRatioThreshold (percent of all packets)',
  // A collector missing from a replaced collectors object counts as disabled;
  // an alert missing from a replaced alerts object counts as enabled.
  local collectorEnabled(collector) = std.get(c.collectors, collector, false),
  local alertEnabled(name) = std.get(c.alerts, name, true),
  local keepFiring = if c.keepFiringFor != '' then { keep_firing_for: c.keepFiringFor } else {},
  // alert renders a rule when every collector it needs is enabled and the
  // alert is not turned off. Threshold alerts set keepFiringFor so a value
  // oscillating around the threshold does not flap; state alerts and
  // one-shot alerts leave it out.
  local alert(name, expr, duration, severity, summary, description, collectors=[], threshold=true) =
    if std.all([collectorEnabled(collector) for collector in collectors]) && alertEnabled(name) then [
      {
        alert: name,
        expr: expr,
        labels: { severity: severity },
        annotations: { summary: summary, description: description },
      } + (if duration != '0m' then { 'for': duration } else {}) + (if threshold then keepFiring else {}),
    ] else [],
  local group(name, rules) = if std.length(rules) > 0 then [{ name: name, rules: rules }] else [],
  // Percent of free addresses in a scope; a scope without addresses has no sample.
  local dhcpScopeFree = '(%s * 100)' % q.ratio(
    q.metric('windows_dhcp_scope_addresses_free'),
    '(%s + %s)' % [q.metric('windows_dhcp_scope_addresses_free'), q.metric('windows_dhcp_scope_addresses_in_use')]
  ),
  local dhcpScopeActive = q.metric('windows_dhcp_scope_state', ['state=~"Enabled|EnabledSwitched"']),
  prometheusAlerts+:: {
    groups:
      group('windows-exporter', std.flattenArrays([
        alert(
          'WindowsExporterDown',
          q.metric('up') + ' == 0',
          '5m',
          c.criticalSeverity,
          'Windows exporter is unreachable',
          'Prometheus cannot scrape Windows exporter on {{ $labels.instance }} (job {{ $labels.job }}).',
          threshold=false
        ),
        alert(
          'WindowsCollectorFailed',
          q.metric('windows_exporter_collector_success') + ' == 0',
          '5m',
          'warning',
          'Windows collector failed',
          'Collector {{ $labels.collector }} on {{ $labels.instance }} has failed for at least 5 minutes.',
          threshold=false
        ),
        alert(
          'WindowsCollectorTimeout',
          q.metric('windows_exporter_collector_timeout') + ' == 1',
          '10m',
          'warning',
          'Windows collector exceeds the scrape timeout',
          'Collector {{ $labels.collector }} on {{ $labels.instance }} has not finished within the scrape timeout for 10 minutes.',
          threshold=false
        ),
        alert(
          'WindowsScrapeQueueTimeout',
          q.metric('windows_exporter_scrape_queue_timeout') + ' == 1',
          '10m',
          'warning',
          'Windows exporter scrapes time out in the queue',
          'Scrapes of {{ $labels.instance }} have timed out waiting for an earlier scrape to finish for 10 minutes.',
          threshold=false
        ),
      ]))
      + group('windows-cpu', std.flattenArrays([
        alert(
          'WindowsCPUHighUsage',
          q.metric('windows:cpu_usage:ratio') + ' * 100 > %s' % c.cpuHighUsageThreshold,
          '15m',
          'warning',
          'Windows CPU usage is high',
          'CPU usage on {{ $labels.instance }} has exceeded %s%% for 15 minutes.' % c.cpuHighUsageThreshold,
          collectors=['cpu']
        ),
        alert(
          'WindowsCPUSaturation',
          q.metric('windows:cpu_queue_per_core:ratio') + ' > %s' % c.cpuQueuePerCoreThreshold,
          '15m',
          'warning',
          'Windows processor queue is long',
          'More than %s threads per logical processor have been waiting for CPU time on {{ $labels.instance }} for 15 minutes.' % c.cpuQueuePerCoreThreshold,
          collectors=['cpu', 'system']
        ),
      ]))
      + group('windows-memory', std.flattenArrays([
        alert(
          'WindowsMemoryHighUsage',
          q.metric('windows:memory_usage:ratio') + ' * 100 > %s' % c.memoryHighUsageThreshold,
          '15m',
          'warning',
          'Windows memory usage is high',
          'Physical memory usage on {{ $labels.instance }} has exceeded %s%% for 15 minutes.' % c.memoryHighUsageThreshold,
          collectors=['memory']
        ),
        alert(
          'WindowsMemoryCommitHighUsage',
          q.metric('windows:memory_committed:ratio') + ' * 100 > %s' % c.memoryCommitHighUsageThreshold,
          '15m',
          'warning',
          'Windows memory commit usage is high',
          'Committed memory on {{ $labels.instance }} has exceeded %s%% of the commit limit for 15 minutes.' % c.memoryCommitHighUsageThreshold,
          collectors=['memory']
        ),
        alert(
          'WindowsMemoryPagingHigh',
          q.metric('windows:memory_swap_page_operations:rate') + ' > %s' % c.memoryPagingThreshold,
          '15m',
          'warning',
          'Windows memory paging is high',
          'Paging on {{ $labels.instance }} has exceeded %s pages per second for 15 minutes.' % c.memoryPagingThreshold,
          collectors=['memory']
        ),
        alert(
          'WindowsPagefileAlmostFull',
          '(1 - %s) * 100 > %s' % [
            q.ratio(q.metric('windows_pagefile_free_bytes'), q.metric('windows_pagefile_limit_bytes')),
            c.pagefileUsageThreshold,
          ],
          '15m',
          'warning',
          'Windows page file is almost full',
          'Page file {{ $labels.file }} on {{ $labels.instance }} has been more than %s%% in use for 15 minutes.' % c.pagefileUsageThreshold,
          collectors=['pagefile']
        ),
      ]))
      + group('windows-disk', std.flattenArrays([
        alert(
          'WindowsDisk%sLatencyHigh' % direction[0],
          q.metric('windows:logical_disk_%s_latency:seconds' % direction[1], [c.volumeSelector]) + ' > %g' % c.diskLatencyThresholdSeconds,
          '15m',
          'warning',
          'Windows disk %s latency is high' % direction[1],
          'Volume {{ $labels.volume }} on {{ $labels.instance }} has averaged more than %g seconds per %s operation for 15 minutes.' % [c.diskLatencyThresholdSeconds, direction[1]],
          collectors=['logical_disk']
        )
        for direction in [['Read', 'read'], ['Write', 'write']]
      ] + [
        alert(
          'WindowsDiskAlmostFull',
          q.metric('windows:logical_disk_free:ratio', [c.volumeSelector]) + ' * 100 < %s' % entry[1],
          '15m',
          entry[0],
          'Windows disk space is low',
          'Volume {{ $labels.volume }} on {{ $labels.instance }} has less than %s%% free space. Windows may delay space metrics by 10–15 minutes.' % entry[1],
          collectors=['logical_disk']
        )
        for entry in [
          ['warning', c.diskFreeWarningThreshold],
          [c.criticalSeverity, c.diskFreeCriticalThreshold],
        ]
      ] + [
        alert(
          'WindowsDiskFillingUp',
          q.metric('windows:logical_disk_free:ratio', [c.volumeSelector]) + ' * 100 < %s and predict_linear(%s[%s], %s) < 0' % [
            c.diskFreeWarningThreshold,
            q.metric('windows_logical_disk_free_bytes', [c.volumeSelector]),
            c.diskPredictionWindow,
            c.diskPredictionHours * 3600,
          ],
          '1h',
          'warning',
          'Windows disk is predicted to fill up',
          'Volume {{ $labels.volume }} on {{ $labels.instance }} is predicted to fill within %s hours.' % c.diskPredictionHours,
          collectors=['logical_disk']
        ),
        alert(
          'WindowsDiskDriveNotOK',
          // One series per status with the current status set to 1, so the
          // alert carries the status the drive reports. Some USB and virtual
          // drives report Unknown for good.
          q.metric('windows_diskdrive_status', ['status!~"OK|Unknown"']) + ' == 1',
          '5m',
          c.criticalSeverity,
          'Windows disk drive is not healthy',
          'Disk drive {{ $labels.device_id }} on {{ $labels.instance }} has reported status {{ $labels.status }} for 5 minutes.',
          collectors=['diskdrive'],
          threshold=false
        ),
      ]))
      + group('windows-network', std.flattenArrays([
        alert(
          'WindowsNetwork%s' % outcome[0],
          q.ratio(
            '(%s + %s)' % [
              q.metric('windows:net_received_%s:rate' % outcome[1], [c.nicSelector]),
              q.metric('windows:net_outbound_%s:rate' % outcome[1], [c.nicSelector]),
            ],
            '(%s + %s)' % [
              q.metric('windows:net_received_packets:rate', [c.nicSelector]),
              q.metric('windows:net_sent_packets:rate', [c.nicSelector]),
            ]
          ) + ' * 100 > %g' % outcome[2],
          '15m',
          'warning',
          'Windows network packet %s are high' % outcome[3],
          'Packet %s on interface {{ $labels.nic }} on {{ $labels.instance }} have exceeded %g%% of all packets for 15 minutes.' % [outcome[3], outcome[2]],
          collectors=['net']
        )
        for outcome in [
          ['Errors', 'errors', c.networkErrorRatioThreshold, 'errors'],
          ['Discards', 'discarded', c.networkDiscardRatioThreshold, 'discards'],
        ]
      ] + [
        alert(
          'WindowsNetworkUtilizationHigh',
          '(%s * 100 > %s) or (%s * 100 > %s)' % [
            q.metric('windows:net_received_utilization:ratio', [c.nicSelector]),
            c.networkUtilizationThreshold,
            q.metric('windows:net_sent_utilization:ratio', [c.nicSelector]),
            c.networkUtilizationThreshold,
          ],
          '15m',
          'warning',
          'Windows network interface utilization is high',
          'Traffic on interface {{ $labels.nic }} on {{ $labels.instance }} has exceeded %s%% of its bandwidth in at least one direction for 15 minutes.' % c.networkUtilizationThreshold,
          collectors=['net']
        ),
        alert(
          'WindowsNetworkInterfaceFlapping',
          'max without (status) (changes(%s[%s])) > 2' % [
            q.metric('windows_net_nic_operation_status', [c.nicSelector, 'status="up"']),
            c.rateInterval,
          ],
          '2m',
          'warning',
          'Windows network interface is flapping',
          'Interface {{ $labels.nic }} on {{ $labels.instance }} has changed its operational status more than twice in %s.' % c.rateInterval,
          collectors=['net']
        ),
      ]))
      + group('windows-service', std.flattenArrays([
        alert(
          'WindowsServiceNotRunning',
          // The matches ignore only the labels of the right-hand metric, so
          // every target label, including ones a selector adds, takes part.
          std.join(' ', [
            'max without (state) (%s == 0)' % q.metric('windows_service_state', [c.serviceSelector, 'state="running"']),
            'and ignoring (start_mode) (%s == 1)' % q.metric('windows_service_start_mode', [c.serviceSelector, 'start_mode="auto"']),
            'unless ignoring (display_name, run_as, path_name, delayed_start, trigger_start) (%s == 1)' % q.metric('windows_service_info', [c.serviceSelector, 'trigger_start="true"']),
          ]),
          '15m',
          'warning',
          'Windows automatic service is not running',
          'Service {{ $labels.service }} on {{ $labels.instance }} starts automatically but has not been running for 15 minutes.',
          collectors=['service'],
          threshold=false
        ),
        alert(
          'WindowsServiceCrashLooping',
          // Every service process has its own series, so the series seen in the
          // window count the processes the service ran as.
          'count without (process_id) (count_over_time(%s[15m])) - 1 > %s' % [
            q.metric('windows_service_process', [c.serviceSelector]),
            c.serviceRestartThreshold,
          ],
          '0m',
          'warning',
          'Windows service is restarting repeatedly',
          'Service {{ $labels.service }} on {{ $labels.instance }} has restarted more than %s times in 15 minutes.' % c.serviceRestartThreshold,
          collectors=['service']
        ),
      ]))
      + group('windows-system', std.flattenArrays([
        alert(
          'WindowsHostRebooted',
          // Fires once per reboot: for the first ten minutes after the boot,
          // and only on a host that was monitored before it. The exporter is
          // down while the host reboots, so the second operand takes the last
          // boot time of the hour before, not the sample of ten minutes ago.
          '(time() - %s) < 600 and (time() - last_over_time(%s[1h] offset 10m)) > 600' % [
            q.metric('windows_system_boot_time_timestamp'),
            q.metric('windows_system_boot_time_timestamp'),
          ],
          '0m',
          c.infoSeverity,
          'Windows host has rebooted',
          '{{ $labels.instance }} rebooted {{ $value | humanizeDuration }} ago.',
          collectors=['system'],
          threshold=false
        ),
        alert(
          'WindowsLicenseNotGenuine',
          q.metric('windows_license_status', ['state!="genuine"']) + ' == 1',
          '1h',
          'warning',
          'Windows license is not genuine',
          'Windows on {{ $labels.instance }} has reported license status {{ $labels.state }} for 1 hour.',
          collectors=['license'],
          threshold=false
        ),
      ]))
      + group('windows-active-directory', std.flattenArrays([
        alert(
          'WindowsADPendingReplication',
          q.metric('windows_ad_replication_pending_operations') + ' > %s' % c.adPendingReplicationThreshold,
          '15m',
          'warning',
          'Active Directory replication queue is high',
          'Pending replication operations on {{ $labels.instance }} have exceeded %s for 15 minutes.' % c.adPendingReplicationThreshold,
          collectors=['ad']
        ),
        alert(
          'WindowsADReplicationFailures',
          '(' + q.rate('windows_ad_replication_sync_requests_total') + ' - ' + q.rate('windows_ad_replication_sync_requests_success_total') + ') > %s' % c.adReplicationFailureRateThreshold,
          '15m',
          'warning',
          'Active Directory replication requests are failing',
          'Replication synchronization requests on {{ $labels.instance }} have been failing for 15 minutes.',
          collectors=['ad']
        ),
      ]))
      + group('windows-time', std.flattenArrays([
        alert(
          'WindowsClockNotSynchronising',
          q.metric('windows_time_ntp_client_time_sources') + ' == 0',
          '10m',
          'warning',
          'Windows clock has no NTP time source',
          'Windows Time on {{ $labels.instance }} has had no active NTP time source for 10 minutes.',
          collectors=['time'],
          threshold=false
        ),
        alert(
          'WindowsClockSkewDetected',
          'abs(%s) > %g' % [q.metric('windows_time_computed_time_offset_seconds'), c.clockOffsetThresholdSeconds],
          '10m',
          'warning',
          'Windows clock offset is high',
          'Windows Time on {{ $labels.instance }} reports a clock offset above %g seconds for 10 minutes.' % c.clockOffsetThresholdSeconds,
          collectors=['time']
        ),
        alert(
          'WindowsClockSourceLocal',
          // Covers domain members (NT5DS) too, which the NTP source count misses.
          q.metric('windows_time_clock_sync_source', ['type=~"Local CMOS Clock|NoSync"']) + ' == 1',
          '10m',
          'warning',
          'Windows clock has no external time source',
          'Windows Time on {{ $labels.instance }} has used {{ $labels.type }} as its clock source for 10 minutes.',
          collectors=['time'],
          threshold=false
        ),
      ]))
      + group('windows-scheduled-task', std.flattenArrays([
        alert(
          'WindowsScheduledTaskFailed',
          // unknown is any result code without a named status, including the
          // exit codes of the task's program.
          // A disabled task keeps its last result, so disabling a broken task
          // clears the alert.
          '%s == 1 unless ignoring (status, state) (%s == 1)' % [
            q.metric('windows_scheduled_task_last_result_status', [c.taskSelector, 'status=~"unknown|terminated"']),
            q.metric('windows_scheduled_task_state', [c.taskSelector, 'state="disabled"']),
          ],
          '1h',
          'warning',
          'Windows scheduled task did not succeed',
          'The last run of task {{ $labels.task }} on {{ $labels.instance }} ended with result {{ $labels.status }}, and the task has not succeeded in the hour since.',
          collectors=['scheduled_task'],
          threshold=false
        ),
      ]))
      + group('windows-update', std.flattenArrays([
        alert(
          'WindowsUpdatesPending',
          'count without (id, revision, category, severity, title) (%s) > 0' % q.metric('windows_update_pending_info', ['deployment_action="installation"']),
          c.updatePendingDuration,
          c.infoSeverity,
          'Windows updates are waiting for installation',
          '{{ $labels.instance }} has had {{ $value }} updates waiting for installation for %s.' % c.updatePendingDuration,
          collectors=['update'],
          threshold=false
        ),
      ]))
      + group('windows-hyperv', std.flattenArrays([
        alert(
          'WindowsHyperVWMIUnhealthy',
          q.metric('windows_hyperv_wmi_health') + ' == 0',
          '10m',
          'warning',
          'Hyper-V WMI namespace is not responding',
          'The Hyper-V WMI namespace on {{ $labels.instance }} has not responded for 10 minutes; Hyper-V metrics are missing.',
          collectors=['hyperv'],
          threshold=false
        ),
        alert(
          'WindowsHyperVReplicationCritical',
          q.metric('windows_hyperv_replica_vm_health') + ' == 3',
          '5m',
          c.criticalSeverity,
          'Hyper-V replication is critical',
          'Replication of virtual machine {{ $labels.vm }} ({{ $labels.relationship }}) on {{ $labels.instance }} has been in critical health for 5 minutes.',
          collectors=['hyperv'],
          threshold=false
        ),
        alert(
          'WindowsHyperVReplicationWarning',
          q.metric('windows_hyperv_replica_vm_health') + ' == 2',
          '15m',
          'warning',
          'Hyper-V replication has a warning',
          'Replication of virtual machine {{ $labels.vm }} ({{ $labels.relationship }}) on {{ $labels.instance }} has been in warning health for 15 minutes.',
          collectors=['hyperv'],
          threshold=false
        ),
      ]))
      // Every cluster node reports the state of the whole cluster, so these
      // alerts fire once per node; group them by cluster labels in Alertmanager.
      + group('windows-cluster', std.flattenArrays([
        alert(
          'WindowsClusterNodeDown',
          q.metric('windows_mscluster_node_state') + ' == 1',
          '5m',
          c.criticalSeverity,
          'Failover cluster node is down',
          'Cluster node {{ $labels.node }} has been down for 5 minutes, as reported by {{ $labels.instance }}.',
          collectors=['mscluster'],
          threshold=false
        ),
        alert(
          'WindowsClusterResourceFailed',
          q.metric('windows_mscluster_resource_state') + ' == 4',
          '5m',
          c.criticalSeverity,
          'Failover cluster resource has failed',
          'Cluster resource {{ $labels.resource }} ({{ $labels.type }}) in group {{ $labels.resource_group }} has been in the failed state for 5 minutes, as reported by {{ $labels.instance }}.',
          collectors=['mscluster'],
          threshold=false
        ),
        alert(
          'WindowsClusterResourceGroupFailed',
          q.metric('windows_mscluster_resourcegroup_state') + ' == 2',
          '5m',
          c.criticalSeverity,
          'Failover cluster resource group has failed',
          'Cluster resource group {{ $labels.resource_group }} has been in the failed state for 5 minutes, as reported by {{ $labels.instance }}.',
          collectors=['mscluster'],
          threshold=false
        ),
        alert(
          'WindowsClusterSharedVolumeAlmostFull',
          q.ratio(q.metric('windows_mscluster_shared_volumes_free_bytes', [c.sharedVolumeSelector]), q.metric('windows_mscluster_shared_volumes_total_bytes', [c.sharedVolumeSelector])) + ' * 100 < %s' % c.diskFreeWarningThreshold,
          '15m',
          'warning',
          'Cluster shared volume is almost full',
          'Cluster shared volume {{ $labels.shared_volume }} has had less than %s%% free space for 15 minutes, as reported by {{ $labels.instance }}.' % c.diskFreeWarningThreshold,
          collectors=['mscluster']
        ),
        alert(
          'WindowsClusterSharedVolumeAlmostFull',
          q.ratio(q.metric('windows_mscluster_shared_volumes_free_bytes', [c.sharedVolumeSelector]), q.metric('windows_mscluster_shared_volumes_total_bytes', [c.sharedVolumeSelector])) + ' * 100 < %s' % c.diskFreeCriticalThreshold,
          '15m',
          c.criticalSeverity,
          'Cluster shared volume is almost full',
          'Cluster shared volume {{ $labels.shared_volume }} has had less than %s%% free space for 15 minutes, as reported by {{ $labels.instance }}.' % c.diskFreeCriticalThreshold,
          collectors=['mscluster']
        ),
      ]))
      // On Storage Spaces Direct every node reports the clustered pools and
      // virtual disks, so these alerts fire once per node as well.
      + group('windows-storage-spaces', std.flattenArrays([
        // The primordial pool holds the disks that are not in a concrete pool
        // yet; only the info metric tells it apart.
        alert(
          'WindowsStoragePoolUnhealthy',
          '(%s == 2) and ignoring (primordial) %s' % [q.metric('windows_storage_spaces_pool_health_status'), q.metric('windows_storage_spaces_pool_info', ['primordial="false"'])],
          '10m',
          c.criticalSeverity,
          'Storage pool is unhealthy',
          'Storage pool {{ $labels.pool }} on {{ $labels.instance }} has reported unhealthy status for 10 minutes.',
          collectors=['storage_spaces'],
          threshold=false
        ),
        alert(
          'WindowsStoragePoolUnhealthy',
          '(%s == 1) and ignoring (primordial) %s' % [q.metric('windows_storage_spaces_pool_health_status'), q.metric('windows_storage_spaces_pool_info', ['primordial="false"'])],
          '10m',
          'warning',
          'Storage pool has a health warning',
          'Storage pool {{ $labels.pool }} on {{ $labels.instance }} has reported warning status for 10 minutes.',
          collectors=['storage_spaces'],
          threshold=false
        ),
        alert(
          'WindowsStorageVirtualDiskUnhealthy',
          q.metric('windows_storage_spaces_virtual_disk_health_status') + ' == 2',
          '10m',
          c.criticalSeverity,
          'Storage Spaces virtual disk is unhealthy',
          'Virtual disk {{ $labels.virtual_disk }} on {{ $labels.instance }} has reported unhealthy status for 10 minutes.',
          collectors=['storage_spaces'],
          threshold=false
        ),
        alert(
          'WindowsStorageVirtualDiskUnhealthy',
          q.metric('windows_storage_spaces_virtual_disk_health_status') + ' == 1',
          '10m',
          'warning',
          'Storage Spaces virtual disk has a health warning',
          'Virtual disk {{ $labels.virtual_disk }} on {{ $labels.instance }} has reported warning status for 10 minutes.',
          collectors=['storage_spaces'],
          threshold=false
        ),
      ]))
      + group('windows-dhcp', std.flattenArrays([
        alert(
          'WindowsDHCPScopeAlmostExhausted',
          dhcpScopeFree + ' < %s and ignoring (state) (%s == 1)' % [c.dhcpScopeFreeWarningThreshold, dhcpScopeActive],
          '15m',
          'warning',
          'DHCP scope is almost exhausted',
          'DHCP scope {{ $labels.scope }} on {{ $labels.instance }} has had less than %s%% free addresses for 15 minutes.' % c.dhcpScopeFreeWarningThreshold,
          collectors=['dhcp']
        ),
        alert(
          'WindowsDHCPScopeAlmostExhausted',
          dhcpScopeFree + ' < %s and ignoring (state) (%s == 1)' % [c.dhcpScopeFreeCriticalThreshold, dhcpScopeActive],
          '15m',
          c.criticalSeverity,
          'DHCP scope is almost exhausted',
          'DHCP scope {{ $labels.scope }} on {{ $labels.instance }} has had less than %s%% free addresses for 15 minutes.' % c.dhcpScopeFreeCriticalThreshold,
          collectors=['dhcp']
        ),
      ]))
      + group('windows-dns', std.flattenArrays([
        alert(
          'WindowsDNSZoneTransferFailures',
          q.rate('windows_dns_zone_transfer_failures_total') + ' > %g' % c.dnsZoneTransferFailureRateThreshold,
          '15m',
          'warning',
          'DNS zone transfers are failing',
          'Zone transfers from DNS server {{ $labels.instance }} have been failing for 15 minutes.',
          collectors=['dns']
        ),
      ])),
  },
}

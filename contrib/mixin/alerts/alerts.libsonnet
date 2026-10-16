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
      ])),
  },
}

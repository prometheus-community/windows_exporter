{
  local c = self._config,
  local q = (import '../lib/queries.libsonnet')(c),
  prometheusRules+:: {
    groups: [
      {
        name: 'windows-exporter-recording',
        rules: [
          { record: 'windows:cpu_usage:ratio', expr: q.cpuUsage },
          { record: 'windows:cpu_cores:count', expr: 'count without (core, mode) (%s)' % q.metric('windows_cpu_time_total', ['mode="idle"']) },
          { record: 'windows:cpu_time:ratio_rate', expr: 'avg without (core) (%s)' % q.rate('windows_cpu_time_total') },
          { record: 'windows:memory_usage:ratio', expr: q.memoryUsage },
          {
            record: 'windows:memory_used:bytes',
            expr: '%s - %s' % [q.metric('windows_memory_physical_total_bytes'), q.metric('windows_memory_available_bytes')],
          },
          {
            record: 'windows:memory_cached:bytes',
            expr: std.join(' + ', [
              q.metric('windows_memory_%s_bytes' % component)
              for component in ['cache', 'modified_page_list', 'standby_cache_core', 'standby_cache_normal_priority', 'standby_cache_reserve']
            ]),
          },
          { record: 'windows:memory_committed:ratio', expr: q.memoryCommitted },
          { record: 'windows:memory_swap_page_operations:rate', expr: q.rate('windows_memory_swap_page_operations_total') },
          { record: 'windows:logical_disk_free:ratio', expr: q.diskFree },
          {
            record: 'windows:logical_disk_used:bytes',
            expr: '%s - %s' % [
              q.metric('windows_logical_disk_size_bytes', [c.volumeSelector]),
              q.metric('windows_logical_disk_free_bytes', [c.volumeSelector]),
            ],
          },
          {
            record: 'windows:logical_disk_busy:ratio',
            expr: 'clamp(1 - %s, 0, 1)' % q.rate('windows_logical_disk_idle_seconds_total', [c.volumeSelector]),
          },
        ] + [
          {
            record: 'windows:logical_disk_%s_bytes:rate' % direction,
            expr: q.rate('windows_logical_disk_%s_bytes_total' % direction, [c.volumeSelector]),
          }
          for direction in ['read', 'write']
        ] + [
          {
            record: 'windows:logical_disk_%s_operations:rate' % direction,
            expr: q.rate('windows_logical_disk_%ss_total' % direction, [c.volumeSelector]),
          }
          for direction in ['read', 'write']
        ] + [
          { record: 'windows:logical_disk_%s_latency:seconds' % direction, expr: q.diskLatency(direction) }
          for direction in ['read', 'write']
        ] + [
          {
            record: 'windows:net_%s_bytes:rate' % direction,
            expr: q.rate('windows_net_bytes_%s_total' % direction, [c.nicSelector]),
          }
          for direction in ['received', 'sent']
        ] + [
          { record: 'windows:net_%s_utilization:ratio' % direction, expr: q.networkUtilization(direction) }
          for direction in ['received', 'sent']
        ] + [
          {
            record: 'windows:net_%s_%s:rate' % [direction, outcome],
            expr: q.rate('windows_net_packets_%s_%s_total' % [direction, outcome], [c.nicSelector]),
          }
          for direction in ['received', 'outbound']
          for outcome in ['errors', 'discarded']
        ],
      },
    ],
  },
}

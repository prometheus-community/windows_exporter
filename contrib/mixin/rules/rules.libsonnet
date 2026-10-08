{
  local c = self._config,
  local q = (import '../lib/queries.libsonnet')(c),
  prometheusRules+:: {
    groups: [
      {
        name: 'windows-exporter-recording',
        rules: [
          { record: 'windows:cpu_usage:ratio', expr: q.cpuUsage },
          { record: 'windows:memory_usage:ratio', expr: q.memoryUsage },
          { record: 'windows:logical_disk_free:ratio', expr: q.diskFree },
        ] + [
          {
            record: 'windows:logical_disk_%s_bytes:rate' % direction,
            expr: q.rate('windows_logical_disk_%s_bytes_total' % direction, [c.volumeSelector]),
          }
          for direction in ['read', 'write']
        ] + [
          {
            record: 'windows:net_%s_bytes:rate' % direction,
            expr: q.rate('windows_net_bytes_%s_total' % direction, [c.nicSelector]),
          }
          for direction in ['received', 'sent']
        ],
      },
    ],
  },
}

function(c)
  {
    selector(extra=[]): std.join(', ', [s for s in [c.windowsExporterSelector] + extra if s != '']),
    metric(name, extra=[]): '%s{%s}' % [name, self.selector(extra)],
    rate(name, extra=[]): 'rate(%s[%s])' % [self.metric(name, extra), c.rateInterval],
    cpuUsage:
      '1 - avg without (core, mode) (%s)' % self.rate('windows_cpu_time_total', ['mode="idle"']),
    memoryUsage:
      '1 - %s / %s' % [
        self.metric('windows_memory_available_bytes'),
        self.metric('windows_memory_physical_total_bytes'),
      ],
    diskFree:
      '%s / %s' % [
        self.metric('windows_logical_disk_free_bytes', [c.volumeSelector]),
        self.metric('windows_logical_disk_size_bytes', [c.volumeSelector]),
      ],
  }

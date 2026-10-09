function(c)
  {
    selector(extra=[]): std.join(', ', [s for s in [c.windowsExporterSelector] + extra if s != '']),
    metric(name, extra=[]): '%s{%s}' % [name, self.selector(extra)],
    rate(name, extra=[]): 'rate(%s[%s])' % [self.metric(name, extra), c.rateInterval],
    // Drop undefined ratios instead of recording NaN or +Inf for idle devices.
    ratio(numerator, denominator): '(%s / (%s > 0))' % [numerator, denominator],
    cpuUsage:
      '1 - avg without (core, mode) (%s)' % self.rate('windows_cpu_time_total', ['mode="idle"']),
    memoryUsage:
      '1 - %s' % self.ratio(
        self.metric('windows_memory_available_bytes'),
        self.metric('windows_memory_physical_total_bytes')
      ),
    diskFree:
      self.ratio(
        self.metric('windows_logical_disk_free_bytes', [c.volumeSelector]),
        self.metric('windows_logical_disk_size_bytes', [c.volumeSelector])
      ),
    memoryCommitted:
      self.ratio(self.metric('windows_memory_committed_bytes'), self.metric('windows_memory_commit_limit')),
    diskLatency(direction):
      self.ratio(
        self.rate('windows_logical_disk_%s_seconds_total' % direction, [c.volumeSelector]),
        self.rate('windows_logical_disk_%ss_total' % direction, [c.volumeSelector])
      ),
    networkUtilization(direction):
      self.ratio(
        self.rate('windows_net_bytes_%s_total' % direction, [c.nicSelector]),
        self.metric('windows_net_current_bandwidth_bytes', [c.nicSelector])
      ),
  }

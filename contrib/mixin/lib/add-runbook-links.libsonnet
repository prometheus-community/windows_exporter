{
  prometheusAlerts+:: {
    groups: [
      group {
        rules: [
          if std.objectHas(rule, 'alert') && !(std.objectHas(rule, 'annotations') && std.objectHas(rule.annotations, 'runbook_url')) then
            rule {
              annotations+: {
                runbook_url: $._config.runbookURLPattern % std.asciiLower(rule.alert),
              },
            }
          else rule
          for rule in group.rules
        ],
      }
      for group in super.groups
    ],
  },
}

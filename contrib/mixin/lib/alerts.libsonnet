// Helpers shared by the alert files in alerts/. Takes the mixin _config.
function(c)
  {
    // A collector missing from a replaced collectors object counts as disabled;
    // an alert missing from a replaced alerts object counts as enabled.
    collectorEnabled(collector): std.get(c.collectors, collector, false),
    alertEnabled(name): std.get(c.alerts, name, true),
    // group renders the rules of one collector as a rule group, or nothing
    // when none of them is enabled. Each rule is a plain Prometheus alerting
    // rule plus two hidden fields: collectors:: lists the collectors the rule
    // reads when it differs from the group's, and state:: true marks state
    // alerts (== 0, == 1) and one-shot alerts, which must resolve at once.
    // Every other alert is a threshold alert and keeps firing for
    // keepFiringFor after its condition clears, so a value oscillating around
    // the threshold does not flap.
    group(name, collectors, rules):
      local enabled = [
        rule + (if c.keepFiringFor != '' && !std.get(rule, 'state', false) then { keep_firing_for: c.keepFiringFor } else {})
        for rule in rules
        if self.alertEnabled(rule.alert) && std.all([self.collectorEnabled(collector) for collector in std.get(rule, 'collectors', collectors)])
      ];
      if std.length(enabled) > 0 then [{ name: name, rules: enabled }] else [],
  }

local linked = {
  _config:: { runbookURLPattern: 'https://example.com/runbooks/%s/' },
  prometheusAlerts:: {
    groups: [
      {
        name: 'custom',
        rules: [
          { alert: 'WindowsCustomAlert', expr: 'up == 0', annotations: { summary: 'Existing summary' } },
          { alert: 'WindowsOverride', expr: 'up == 0', annotations: { runbook_url: 'https://example.com/override' } },
          { record: 'windows:custom:sum', expr: 'sum(up)' },
          { alert: 'WindowsNoAnnotations', expr: 'up == 0' },
        ],
      },
    ],
  },
} + (import '../lib/add-runbook-links.libsonnet');
local rules = linked.prometheusAlerts.groups[0].rules;
assert rules[0].annotations == { summary: 'Existing summary', runbook_url: 'https://example.com/runbooks/windowscustomalert/' };
assert rules[1].annotations.runbook_url == 'https://example.com/override';
assert !std.objectHas(rules[2], 'annotations');
assert rules[3].annotations.runbook_url == 'https://example.com/runbooks/windowsnoannotations/';
assert rules[0].expr == 'up == 0' && linked.prometheusAlerts.groups[0].name == 'custom';
{ result: 'Runbook link composition passed' }

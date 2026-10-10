// Renders dashboard/windows-exporter-dashboard.json. Run with jsonnet -J vendor -S.
// The sample dashboard shows every collector; mixin users select theirs with
// _config.collectors.
local mixin = (import 'mixin.libsonnet') + {
  _config+:: { collectors: { [collector]: true for collector in std.objectFields(super.collectors) } },
};

(import 'lib/manifest.libsonnet')(mixin.grafanaDashboards['windows-exporter.json'])

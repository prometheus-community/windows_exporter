// Renders the sample dashboards into ../../dashboard. Run with
// jsonnet -J vendor -S -m <directory>. The sample dashboards show every
// collector; mixin users select theirs with _config.collectors.
local manifest = import 'lib/manifest.libsonnet';
local mixin = (import 'mixin.libsonnet') + {
  _config+:: { collectors: { [collector]: true for collector in std.objectFields(super.collectors) } },
};

// The full dashboard keeps the file name it had before the tab dashboards.
{
  [if name == 'windows-exporter.json' then 'windows-exporter-dashboard.json' else name]: manifest(mixin.grafanaDashboards[name])
  for name in std.objectFields(mixin.grafanaDashboards)
}

// Renders dashboard/windows-exporter-dashboard.json. Run with jsonnet -J vendor -S.
(import 'lib/manifest.libsonnet')((import 'mixin.libsonnet').grafanaDashboards['windows-exporter.json'])

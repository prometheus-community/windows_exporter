{
  local dashboards = (import 'windows-exporter.libsonnet')($._config),
  grafanaDashboards+:: { 'windows-exporter.json': dashboards.full } + dashboards.tabs,
}

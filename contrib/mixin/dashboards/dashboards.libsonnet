{
  grafanaDashboards+:: {
    'windows-exporter.json': (import 'windows-exporter.libsonnet')($._config),
  },
}

//go:build windows

package perfdata

import "github.com/prometheus-community/windows_exporter/internal/perfdata"

type Object struct {
	Object        string             `json:"object"         yaml:"object"`
	Instances     []string           `json:"instances"      yaml:"instances"`
	Counters      map[string]Counter `json:"counters"       yaml:"counters"`
	InstanceLabel string             `json:"instance_label" yaml:"instance_label"` //nolint:tagliatelle

	collector *perfdata.Collector
}

type Counter struct {
	Type string `json:"type" yaml:"type"`
}

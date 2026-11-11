// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package registry

import (
	"github.com/prometheus-community/windows_exporter/internal/config/yamlstring"
	"github.com/prometheus-community/windows_exporter/internal/multiplier"
	"github.com/prometheus/client_golang/prometheus"
	"go.yaml.in/yaml/v3"
	winregistry "golang.org/x/sys/windows/registry"
)

type Key struct {
	Name   string  `json:"name"   yaml:"name"`
	Key    string  `json:"key"    yaml:"key"`
	Values []Value `json:"values" yaml:"values"`

	// resolved at Build time
	hive    winregistry.Key
	subPath string
	label   string
}

type Value struct {
	Name       string            `json:"name"       yaml:"name"`
	Metric     string            `json:"metric"     yaml:"metric"`
	Help       string            `json:"help"       yaml:"help"`
	Type       string            `json:"type"       yaml:"type"`
	Labels     map[string]string `json:"labels"     yaml:"labels"`
	Multiplier *float64          `json:"multiplier" yaml:"multiplier"`

	// resolved at Build time
	desc       *prometheus.Desc
	metricType prometheus.ValueType
	multiplier multiplier.Multiplier
}

// UnmarshalYAML decodes the registry block of the configuration file, where
// the keys are kept as a string. See [yamlstring.DecodeBlock].
func (c *Config) UnmarshalYAML(node *yaml.Node) error {
	return yamlstring.DecodeBlock(node, "keys", &c.Keys)
}

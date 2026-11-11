// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package performancecounter

import (
	"github.com/prometheus-community/windows_exporter/internal/config/yamlstring"
	"github.com/prometheus-community/windows_exporter/internal/multiplier"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus/client_golang/prometheus"
	"go.yaml.in/yaml/v3"
)

type Object struct {
	Name          string          `json:"name"           yaml:"name"`
	Object        string          `json:"object"         yaml:"object"`
	Type          pdh.CounterType `json:"type"           yaml:"type"`
	Instances     []string        `json:"instances"      yaml:"instances"`
	Counters      []Counter       `json:"counters"       yaml:"counters"`
	InstanceLabel string          `json:"instance_label" yaml:"instance_label"`

	collector *pdh.Collector[pdh.Row]
	// valueIndex maps a counter name to its index in pdh.Row.Values.
	valueIndex map[string]int
}

type Counter struct {
	Name       string            `json:"name"       yaml:"name"`
	Type       string            `json:"type"       yaml:"type"`
	Metric     string            `json:"metric"     yaml:"metric"`
	Labels     map[string]string `json:"labels"     yaml:"labels"`
	Multiplier *float64          `json:"multiplier" yaml:"multiplier"`

	// Resolved at Build time.
	multiplier multiplier.Multiplier
	// desc describes the series without an instance label. instanceDesc
	// adds the instance label as variable label; it is nil if the object
	// has no instances or the counter sets the instance label itself.
	desc         *prometheus.Desc
	instanceDesc *prometheus.Desc
}

// https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/54691ebe11bb9ec32b4e35cd31fcb94a352de134/receiver/windowsperfcountersreceiver/README.md?plain=1#L150

// UnmarshalYAML decodes the performancecounter block of the configuration file,
// where the objects are kept as a string. See [yamlstring.DecodeBlock].
func (c *Config) UnmarshalYAML(node *yaml.Node) error {
	return yamlstring.DecodeBlock(node, "objects", &c.Objects)
}

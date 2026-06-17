// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows

package registry

import (
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
	Name   string            `json:"name"   yaml:"name"`
	Metric string            `json:"metric" yaml:"metric"`
	Type   string            `json:"type"   yaml:"type"`
	Labels map[string]string `json:"labels" yaml:"labels"`

	// resolved at Build time
	desc       *prometheus.Desc
	metricType prometheus.ValueType
}

// UnmarshalYAML is a no-op, so the strict config file validation accepts the
// keys being provided as a string. See the performancecounter collector.
func (*Config) UnmarshalYAML(*yaml.Node) error {
	return nil
}

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

package wmi

import (
	"github.com/prometheus-community/windows_exporter/internal/config/yamlstring"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus/client_golang/prometheus"
	"go.yaml.in/yaml/v3"
)

type Query struct {
	Name            string          `json:"name"             yaml:"name"`
	Namespace       string          `json:"namespace"        yaml:"namespace"`
	Class           string          `json:"class"            yaml:"class"`
	Where           string          `json:"where"            yaml:"where"`
	LabelProperties []LabelProperty `json:"label_properties" yaml:"label_properties"`
	Properties      []Property      `json:"properties"       yaml:"properties"`

	// Resolved at Build time.
	namespace  mi.Namespace
	query      mi.Query
	wql        string
	labelNames []string
}

type LabelProperty struct {
	Name  string `json:"name"  yaml:"name"`
	Label string `json:"label" yaml:"label"`

	// Resolved at Build time.
	elementName mi.ElementName
}

type Property struct {
	Name   string            `json:"name"   yaml:"name"`
	Metric string            `json:"metric" yaml:"metric"`
	Help   string            `json:"help"   yaml:"help"`
	Type   string            `json:"type"   yaml:"type"`
	Labels map[string]string `json:"labels" yaml:"labels"`

	// Resolved at Build time.
	elementName mi.ElementName
	desc        *prometheus.Desc
	metricType  prometheus.ValueType
}

// UnmarshalYAML decodes the wmi block of the configuration file, where the
// queries are kept as a string. See [yamlstring.DecodeBlock].
func (c *Config) UnmarshalYAML(node *yaml.Node) error {
	return yamlstring.DecodeBlock(node, "queries", &c.Queries)
}

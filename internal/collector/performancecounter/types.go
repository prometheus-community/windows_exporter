// Copyright 2024 The Prometheus Authors
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

package performancecounter

import (
	"github.com/prometheus-community/windows_exporter/internal/pdh"
)

type Object struct {
	Name          string          `json:"name"           yaml:"name"`
	Object        string          `json:"object"         yaml:"object"`
	Type          pdh.CounterType `json:"type"           yaml:"type"`
	Instances     []string        `json:"instances"      yaml:"instances"`
	Counters      []Counter       `json:"counters"       yaml:"counters"`
	InstanceLabel string          `json:"instance_label" yaml:"instance_label"` //nolint:tagliatelle

	collector      *pdh.Collector
	perfDataObject any
}

type Counter struct {
	Name   string            `json:"name"   yaml:"name"`
	Type   string            `json:"type"   yaml:"type"`
	Metric string            `json:"metric" yaml:"metric"`
	Labels map[string]string `json:"labels" yaml:"labels"`
}

// https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/54691ebe11bb9ec32b4e35cd31fcb94a352de134/receiver/windowsperfcountersreceiver/README.md?plain=1#L150

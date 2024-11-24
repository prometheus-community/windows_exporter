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

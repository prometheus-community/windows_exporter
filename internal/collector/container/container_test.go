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

package container_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/container"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/prometheus/client_golang/prometheus"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, container.Name, container.NewWithFlags)
}

func TestCollector(t *testing.T) {
	metrics := testutils.TestCollector(t, container.New, nil)
	testutils.RequireFixtureMetric(t, metrics, container.Name, "windows_container_available", prometheus.Labels{"container": "hostprocess", "namespace": "default", "hostprocess": "true"})
	testutils.RequireFixtureMetric(t, metrics, container.Name, "windows_container_available", prometheus.Labels{"container": "nanoserver", "namespace": "default", "hostprocess": "false"})
}

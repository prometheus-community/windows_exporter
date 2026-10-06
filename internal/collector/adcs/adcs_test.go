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

package adcs_test

import (
	"os"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/adcs"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, adcs.Name, adcs.NewWithFlags)
}

func TestCollector(t *testing.T) {
	metrics := testutils.TestCollector(t, adcs.New, nil)

	template := os.Getenv("WINDOWS_EXPORTER_TEST_ADCS_TEMPLATE")
	if template == "" {
		return
	}

	issued := testutils.RequireFixtureMetric(t, metrics, adcs.Name, "windows_adcs_issued_requests_total", prometheus.Labels{"cert_template": template})
	require.NotNil(t, issued)
	require.Positive(t, issued.GetCounter().GetValue(), "domain fixture certificate was not issued")
}

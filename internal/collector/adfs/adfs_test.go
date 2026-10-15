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

package adfs_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/adfs"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/stretchr/testify/require"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, adfs.Name, adfs.NewWithFlags)
}

func TestCollector(t *testing.T) {
	families := testutils.TestCollector(t, adfs.New, nil)

	// The query times need the time base of their PERF_AVERAGE_TIMER counter.
	for _, name := range []string{
		"windows_adfs_db_artifact_queries_total",
		"windows_adfs_db_artifact_query_time_seconds_total",
		"windows_adfs_db_config_queries_total",
		"windows_adfs_db_config_query_time_seconds_total",
	} {
		require.Contains(t, families, name)
	}
}

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

package memory_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/memory"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, memory.Name, memory.NewWithFlags)
}

func TestCollector(t *testing.T) {
	families := testutils.TestCollector(t, memory.New, nil)

	// "Pool Paged Allocs" and "Pool Nonpaged Allocs" are PERF_COUNTER_RAWCOUNT,
	// an instantaneous value.
	for _, name := range []string{"windows_memory_pool_paged_allocs", "windows_memory_pool_nonpaged_allocs"} {
		require.Contains(t, families, name)
		require.Equal(t, dto.MetricType_GAUGE, families[name].GetType(), name)
		require.NotContains(t, families, name+"_total")
	}
}

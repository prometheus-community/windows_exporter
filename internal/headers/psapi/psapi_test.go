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

package psapi_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/headers/psapi"
	"github.com/stretchr/testify/require"
)

func TestGetPerformanceInfo(t *testing.T) {
	t.Parallel()

	info, err := psapi.GetPerformanceInfo()
	require.NoError(t, err)
	require.NotZero(t, info.PageSize)
}

// TestGetPerformanceInfoDoesNotAllocate checks that the result struct, whose
// address is passed to the native call, stays on the stack.
func TestGetPerformanceInfoDoesNotAllocate(t *testing.T) {
	allocs := testing.AllocsPerRun(100, func() {
		_, _ = psapi.GetPerformanceInfo()
	})

	require.Zero(t, allocs)
}

func BenchmarkGetPerformanceInfo(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		_, _ = psapi.GetPerformanceInfo()
	}
}

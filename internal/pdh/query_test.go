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

package pdh_test

import (
	"log/slog"
	"sync"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/stretchr/testify/require"
)

type systemValues struct {
	Processes float64 `perfdata:"Processes"`
	Threads   float64 `perfdata:"Threads"`
}

type memoryValues struct {
	AvailableBytes float64 `perfdata:"Available Bytes"`
}

func TestQuery(t *testing.T) {
	t.Parallel()

	query, err := pdh.NewQuery()
	require.NoError(t, err)

	system, err := pdh.NewQueryCollector[systemValues](slog.New(slog.DiscardHandler), query, pdh.CounterTypeRaw, "System", nil)
	require.NoError(t, err)

	memory, err := pdh.NewQueryCollector[memoryValues](slog.New(slog.DiscardHandler), query, pdh.CounterTypeRaw, "Memory", nil)
	require.NoError(t, err)

	var (
		systemData []systemValues
		memoryData []memoryValues
	)

	// One sample serves both collectors.
	require.NoError(t, query.Collect())
	require.NoError(t, system.Collect(&systemData))
	require.NoError(t, memory.Collect(&memoryData))
	require.Len(t, systemData, 1)
	require.Len(t, memoryData, 1)
	require.NotZero(t, systemData[0].Processes)
	require.NotZero(t, memoryData[0].AvailableBytes)

	// Closing one collector removes only its counters.
	system.Close()
	require.ErrorIs(t, system.Collect(&systemData), pdh.ErrPerformanceCounterNotInitialized)
	require.NoError(t, query.Collect())
	require.NoError(t, memory.Collect(&memoryData))
	require.NotZero(t, memoryData[0].AvailableBytes)

	memory.Close()
	query.Close()
	query.Close()

	require.ErrorIs(t, query.Collect(), pdh.ErrPerformanceCounterNotInitialized)

	_, err = pdh.NewQueryCollector[memoryValues](slog.New(slog.DiscardHandler), query, pdh.CounterTypeRaw, "Memory", nil)
	require.ErrorIs(t, err, pdh.ErrPerformanceCounterNotInitialized)

	_, err = pdh.NewQueryCollector[memoryValues](slog.New(slog.DiscardHandler), nil, pdh.CounterTypeRaw, "Memory", nil)
	require.ErrorIs(t, err, pdh.ErrPerformanceCounterNotInitialized)
}

// A collector whose object is missing reports the error, and the other
// collector of the query keeps working.
func TestQueryMissingObject(t *testing.T) {
	t.Parallel()

	query, err := pdh.NewQuery()
	require.NoError(t, err)

	defer query.Close()

	missing, err := pdh.NewQueryCollector[memoryValues](slog.New(slog.DiscardHandler), query, pdh.CounterTypeRaw, "Does Not Exist", nil)
	require.Error(t, err)

	defer missing.Close()

	memory, err := pdh.NewQueryCollector[memoryValues](slog.New(slog.DiscardHandler), query, pdh.CounterTypeRaw, "Memory", nil)
	require.NoError(t, err)

	defer memory.Close()

	var data []memoryValues

	require.NoError(t, query.Collect())
	require.NoError(t, memory.Collect(&data))
	require.NotEmpty(t, data)
}

func TestQueryConcurrentCollect(t *testing.T) {
	t.Parallel()

	query, err := pdh.NewQuery()
	require.NoError(t, err)

	system, err := pdh.NewQueryCollector[systemValues](slog.New(slog.DiscardHandler), query, pdh.CounterTypeRaw, "System", nil)
	require.NoError(t, err)

	memory, err := pdh.NewQueryCollector[memoryValues](slog.New(slog.DiscardHandler), query, pdh.CounterTypeRaw, "Memory", nil)
	require.NoError(t, err)

	var wg sync.WaitGroup

	for range 8 {
		wg.Go(func() {
			var (
				systemData []systemValues
				memoryData []memoryValues
			)

			for range 20 {
				if err := query.Collect(); err != nil {
					t.Error(err)

					return
				}

				if err := system.Collect(&systemData); err != nil {
					t.Error(err)
				}

				if err := memory.Collect(&memoryData); err != nil {
					t.Error(err)
				}
			}
		})
	}

	wg.Wait()

	system.Close()
	memory.Close()
	query.Close()
}

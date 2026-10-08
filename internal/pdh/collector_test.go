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
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/stretchr/testify/require"
)

type process struct {
	Name        string
	ThreadCount float64 `perfdata:"Thread Count"`
}

func TestCollector(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		object    string
		instances []string
	}{
		{
			object:    "Process",
			instances: []string{"*"},
		},
	} {
		t.Run(tc.object, func(t *testing.T) {
			t.Parallel()

			performanceData, err := pdh.NewCollector[process](slog.New(slog.DiscardHandler), pdh.CounterTypeRaw, tc.object, tc.instances)
			require.NoError(t, err)

			time.Sleep(100 * time.Millisecond)

			var data []process

			err = performanceData.Collect(&data)
			require.NoError(t, err)
			require.NotEmpty(t, data)

			err = performanceData.Collect(&data)
			require.NoError(t, err)
			require.NotEmpty(t, data)

			for _, instance := range data {
				if instance.Name == "Idle" || instance.Name == "Secure System" {
					continue
				}

				require.NotZerof(t, instance.ThreadCount, "object: %s, instance: %s, counter: %s", tc.object, instance, instance.ThreadCount)
			}
		})
	}
}

type processorInformation struct {
	Name          string
	ProcessorTime float64 `perfdata:"% Processor Time"`
	IdleTime      float64 `perfdata:"% Idle Time"`
}

// TestCollectorExplicitInstances verifies that a collector built with an explicit
// list of instances returns one row per requested instance. Each instance has its
// own counter handle, and PDH writes every single-item result into the same buffer,
// so the instance names must not be resolved through a cache keyed by buffer address.
func TestCollectorExplicitInstances(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.DiscardHandler)

	discovery, err := pdh.NewCollector[processorInformation](logger, pdh.CounterTypeRaw, "Processor Information", pdh.InstancesAll)
	require.NoError(t, err)

	t.Cleanup(discovery.Close)

	var all []processorInformation

	require.NoError(t, discovery.Collect(&all))

	instances := make([]string, 0, 3)

	for _, row := range all {
		if len(instances) == cap(instances) {
			break
		}

		instances = append(instances, row.Name)
	}

	if len(instances) < 2 {
		t.Skipf("need at least 2 Processor Information instances, got %v", instances)
	}

	for _, counterType := range []pdh.CounterType{pdh.CounterTypeRaw, pdh.CounterTypeFormatted} {
		t.Run(string(counterType), func(t *testing.T) {
			t.Parallel()

			collector, err := pdh.NewCollector[processorInformation](logger, counterType, "Processor Information", instances)
			require.NoError(t, err)

			t.Cleanup(collector.Close)

			// Formatted rate counters need two samples before they report a value.
			time.Sleep(100 * time.Millisecond)

			var data []processorInformation

			require.NoError(t, collector.Collect(&data))

			names := make([]string, 0, len(data))
			for _, row := range data {
				names = append(names, row.Name)
			}

			require.ElementsMatch(t, instances, names)
		})
	}
}

func TestNewCollectorNonStruct(t *testing.T) {
	t.Parallel()

	_, err := pdh.NewCollector[int](slog.New(slog.DiscardHandler), pdh.CounterTypeRaw, "Process", pdh.InstancesAll)
	require.Error(t, err)
}

func TestDynamicCollector(t *testing.T) {
	t.Parallel()

	counters := []string{"Thread Count", "Handle Count"}

	collector, err := pdh.NewDynamicCollector(slog.New(slog.DiscardHandler), pdh.CounterTypeRaw, "Process", pdh.InstancesAll, counters)
	require.NoError(t, err)

	t.Cleanup(collector.Close)

	var data []pdh.Row

	require.NoError(t, collector.Collect(&data))
	require.NotEmpty(t, data)

	for _, row := range data {
		require.NotEmpty(t, row.Name)
		require.Len(t, row.Values, len(counters))

		if row.Name == "Idle" || row.Name == "Secure System" {
			continue
		}

		require.NotZerof(t, row.Values[0], "instance: %s, counter: %s", row.Name, counters[0])
	}
}

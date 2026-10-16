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
	"time"

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus/client_golang/prometheus"
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

func TestCollectorConcurrentCollect(t *testing.T) {
	t.Parallel()

	for _, counterType := range []pdh.CounterType{pdh.CounterTypeRaw, pdh.CounterTypeFormatted} {
		t.Run(string(counterType), func(t *testing.T) {
			t.Parallel()

			collector, err := pdh.NewCollector[process](slog.New(slog.DiscardHandler), counterType, "Process", pdh.InstancesAll)
			require.NoError(t, err)

			t.Cleanup(collector.Close)

			const callers = 8

			data := make([][]process, callers)
			errs := make([]error, callers)
			start := make(chan struct{})

			var wg sync.WaitGroup

			for i := range callers {
				wg.Go(func() {
					<-start

					for range 4 {
						if errs[i] = collector.Collect(&data[i]); errs[i] != nil {
							return
						}
					}
				})
			}

			close(start)
			wg.Wait()

			for i := range callers {
				require.NoError(t, errs[i])
				require.NotEmpty(t, data[i])
			}
		})
	}
}

func TestCollectorConcurrentClose(t *testing.T) {
	t.Parallel()

	collector, err := pdh.NewCollector[process](slog.New(slog.DiscardHandler), pdh.CounterTypeRaw, "Process", pdh.InstancesAll)
	require.NoError(t, err)

	t.Cleanup(collector.Close)

	const callers = 8

	errs := make([]error, callers)
	start := make(chan struct{})

	var wg sync.WaitGroup

	for i := range callers {
		wg.Go(func() {
			<-start

			var data []process

			errs[i] = collector.Collect(&data)
		})
	}

	wg.Go(func() {
		<-start
		collector.Close()
	})

	close(start)
	wg.Wait()

	for _, err := range errs {
		if err != nil {
			require.ErrorIs(t, err, pdh.ErrPerformanceCounterNotInitialized)
		}
	}

	var data []process

	require.ErrorIs(t, collector.Collect(&data), pdh.ErrPerformanceCounterNotInitialized)
	collector.Close()
}

func TestCollectorInitializationError(t *testing.T) {
	t.Parallel()

	type invalidProcess struct {
		ThreadCount float64 `perfdata:"Thread Count"`
		HandleCount string  `perfdata:"Handle Count"`
	}

	// NewCollector releases what it opened and returns no collector with an error.
	collector, err := pdh.NewCollector[invalidProcess](slog.New(slog.DiscardHandler), pdh.CounterTypeRaw, "Process", pdh.InstancesAll)
	require.ErrorContains(t, err, "field HandleCount must be a float64")
	require.Nil(t, collector)

	var data []invalidProcess

	require.ErrorIs(t, collector.Collect(&data), pdh.ErrPerformanceCounterNotInitialized)

	// Closing a nil collector is a no-op.
	collector.Close()
}

// TestCollectorInitializationErrors checks that every constructor error
// returns a nil collector.
func TestCollectorInitializationErrors(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.DiscardHandler)

	t.Run("missing counter", func(t *testing.T) {
		t.Parallel()

		type missingCounter struct {
			ThreadCount float64 `perfdata:"Thread Count"`
			Missing     float64 `perfdata:"No Such Counter"`
		}

		collector, err := pdh.NewCollector[missingCounter](logger, pdh.CounterTypeRaw, "Process", pdh.InstancesAll)
		require.ErrorIs(t, err, pdh.NewPdhError(pdh.CstatusNoCounter))
		require.Nil(t, collector)
	})

	t.Run("missing object", func(t *testing.T) {
		t.Parallel()

		collector, err := pdh.NewDynamicCollector(logger, pdh.CounterTypeRaw, "No Such Object", pdh.InstancesAll, []string{"Thread Count"})
		require.ErrorIs(t, err, pdh.NewPdhError(pdh.CstatusNoObject))
		require.Nil(t, collector)
	})

	t.Run("NUL in counter path", func(t *testing.T) {
		t.Parallel()

		collector, err := pdh.NewDynamicCollector(logger, pdh.CounterTypeRaw, "Process", pdh.InstancesAll, []string{"Thread\x00Count"})
		require.ErrorIs(t, err, pdh.NewPdhError(pdh.CstatusBadCountername))
		require.Nil(t, collector)
	})

	t.Run("invalid perfdata_min_build", func(t *testing.T) {
		t.Parallel()

		type notANumber struct {
			ThreadCount float64 `perfdata:"Thread Count" perfdata_min_build:"abc"`
		}

		type tooLarge struct {
			// Truncated to uint16, this was build 0 and the counter was never skipped.
			ThreadCount float64 `perfdata:"Thread Count" perfdata_min_build:"65536"`
		}

		type negative struct {
			ThreadCount float64 `perfdata:"Thread Count" perfdata_min_build:"-1"`
		}

		_, err := pdh.NewCollector[notANumber](logger, pdh.CounterTypeRaw, "Process", pdh.InstancesAll)
		require.ErrorContains(t, err, `field ThreadCount: invalid perfdata_min_build "abc"`)

		_, err = pdh.NewCollector[tooLarge](logger, pdh.CounterTypeRaw, "Process", pdh.InstancesAll)
		require.ErrorContains(t, err, `field ThreadCount: invalid perfdata_min_build "65536"`)

		collector, err := pdh.NewCollector[negative](logger, pdh.CounterTypeRaw, "Process", pdh.InstancesAll)
		require.ErrorContains(t, err, `field ThreadCount: invalid perfdata_min_build "-1"`)
		require.Nil(t, collector)
	})

	t.Run("future counter is skipped", func(t *testing.T) {
		t.Parallel()

		type futureCounter struct {
			ThreadCount float64 `perfdata:"Thread Count"`
			Future      float64 `perfdata:"No Such Counter" perfdata_min_build:"65535"`
		}

		collector, err := pdh.NewCollector[futureCounter](logger, pdh.CounterTypeRaw, "Process", pdh.InstancesAll)
		require.NoError(t, err)

		t.Cleanup(collector.Close)
	})
}

// TestQueryCollectorInitializationError checks that a failing collector
// removes its counters from a shared query and leaves the other collectors of
// the query working.
func TestQueryCollectorInitializationError(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.DiscardHandler)

	query, err := pdh.NewQuery()
	require.NoError(t, err)

	t.Cleanup(query.Close)

	type system struct {
		Processes float64 `perfdata:"Processes"`
	}

	type missingCounter struct {
		Processes float64 `perfdata:"Processes"`
		Missing   float64 `perfdata:"No Such Counter"`
	}

	collector, err := pdh.NewQueryCollector[system](logger, query, pdh.CounterTypeRaw, "System", nil)
	require.NoError(t, err)

	t.Cleanup(collector.Close)

	failed, err := pdh.NewQueryCollector[missingCounter](logger, query, pdh.CounterTypeRaw, "System", nil)
	require.ErrorIs(t, err, pdh.NewPdhError(pdh.CstatusNoCounter))
	require.Nil(t, failed)

	require.NoError(t, query.Collect())

	var data []system

	require.NoError(t, collector.Collect(&data))
	require.Len(t, data, 1)
	require.Positive(t, data[0].Processes)
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

// TestNewCollectorInvalidInput checks that invalid input returns no collector,
// so there is nothing for the caller to close.
func TestNewCollectorInvalidInput(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.DiscardHandler)

	collector, err := pdh.NewCollector[process](logger, "invalid", "Process", pdh.InstancesAll)
	require.ErrorContains(t, err, "invalid result type")
	require.Nil(t, collector)

	collector2, err := pdh.NewCollector[struct{ Name string }](logger, pdh.CounterTypeRaw, "Process", pdh.InstancesAll)
	require.ErrorContains(t, err, "no counters configured")
	require.Nil(t, collector2)
}

type processIO struct {
	Name          string
	ThreadCount   float64 `perfdata:"Thread Count"`
	IOReadBytesPS float64 `perfdata:"IO Read Bytes/sec"`
}

// TestCollectorMetricType checks that every counter carries its own metric
// type, independent of the other counters of the object.
func TestCollectorMetricType(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		resultType  pdh.CounterType
		threadCount prometheus.ValueType
		ioReadBytes prometheus.ValueType
	}{
		{pdh.CounterTypeRaw, prometheus.GaugeValue, prometheus.CounterValue},
		{pdh.CounterTypeFormatted, prometheus.GaugeValue, prometheus.GaugeValue},
	} {
		t.Run(string(tc.resultType), func(t *testing.T) {
			t.Parallel()

			collector, err := pdh.NewCollector[processIO](slog.New(slog.DiscardHandler), tc.resultType, "Process", pdh.InstancesAll)
			require.NoError(t, err)

			t.Cleanup(collector.Close)

			metricType, ok := collector.MetricType("Thread Count")
			require.True(t, ok)
			require.Equal(t, tc.threadCount, metricType)

			metricType, ok = collector.MetricType("IO Read Bytes/sec")
			require.True(t, ok)
			require.Equal(t, tc.ioReadBytes, metricType)

			_, ok = collector.MetricType("unknown")
			require.False(t, ok)
		})
	}
}

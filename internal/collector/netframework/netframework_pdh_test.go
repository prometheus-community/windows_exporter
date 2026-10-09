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

package netframework

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

type fixturePerfCollector[T any] struct {
	rows   []T
	err    error
	closed int
}

func (c *fixturePerfCollector[T]) Collect(dst *[]T) error {
	*dst = append((*dst)[:0], c.rows...)

	return c.err
}

func (c *fixturePerfCollector[T]) Close() { c.closed++ }

type fixturePromCollector struct {
	collect func(chan<- prometheus.Metric)
}

func (fixturePromCollector) Describe(chan<- *prometheus.Desc)      {}
func (c fixturePromCollector) Collect(ch chan<- prometheus.Metric) { c.collect(ch) }

type expectedMetric struct {
	name       string
	metricType dto.MetricType
	value      float64
	area       string
}

func assertFixtureMetrics(t *testing.T, collect func(chan<- prometheus.Metric, time.Duration) error, expected []expectedMetric, withPID bool) {
	t.Helper()

	registry := prometheus.NewRegistry()

	var collectErr error

	registry.MustRegister(fixturePromCollector{collect: func(ch chan<- prometheus.Metric) { collectErr = collect(ch, time.Second) }})
	families, err := registry.Gather()
	require.NoError(t, err)
	require.NoError(t, collectErr)

	wanted := make(map[string]expectedMetric)

	for _, process := range []string{"worker", "worker#1", "report_Total"} {
		for _, metric := range expected {
			wanted[metric.name+"|"+process+"|"+metric.area] = metric
		}
	}

	gotCount := 0

	for _, family := range families {
		for _, metric := range family.GetMetric() {
			labels := make(map[string]string)
			for _, label := range metric.GetLabel() {
				labels[label.GetName()] = label.GetValue()
			}

			key := family.GetName() + "|" + labels["process"] + "|" + labels["area"]
			want, ok := wanted[key]
			require.True(t, ok, "unexpected metric %s", key)
			require.Equal(t, want.metricType, family.GetType(), key)

			value := metric.GetGauge().GetValue()
			if family.GetType() == dto.MetricType_COUNTER {
				value = metric.GetCounter().GetValue()
			}

			require.InDelta(t, want.value, value, 1e-12, key)

			labelCount := 1

			if withPID {
				require.Equal(t, "123", labels["process_id"])

				labelCount++
			}

			if want.area != "" {
				labelCount++
			}

			require.Len(t, labels, labelCount)
			delete(wanted, key)

			gotCount++
		}
	}

	require.Empty(t, wanted)
	require.Equal(t, len(expected)*3, gotCount)
}

func TestClrExceptionsMetricCompatibility(t *testing.T) {
	c := &Collector{perfFrequency: 1000}
	c.describeClrExceptions()

	row := perfDataClrExceptions{
		NumberofExcepsThrown:    11,
		NumberofFiltersPersec:   12,
		NumberofFinallysPersec:  13,
		ThrowToCatchDepthPersec: 14,
	}
	rows := make([]perfDataClrExceptions, 0, 4)

	for _, name := range []string{"_Global_", "worker", "worker#1", "report_Total"} {
		row.Name = name
		rows = append(rows, row)
	}

	c.perfClrExceptions = &fixturePerfCollector[perfDataClrExceptions]{rows: rows}
	assertFixtureMetrics(t, c.collectClrExceptions, []expectedMetric{
		{"windows_netframework_clrexceptions_exceptions_thrown_total", dto.MetricType_COUNTER, 11, ""},
		{"windows_netframework_clrexceptions_exceptions_filters_total", dto.MetricType_COUNTER, 12, ""},
		{"windows_netframework_clrexceptions_exceptions_finallys_total", dto.MetricType_COUNTER, 13, ""},
		{"windows_netframework_clrexceptions_throw_to_catch_depth_total", dto.MetricType_COUNTER, 14, ""},
	}, false)
}

func TestClrInteropMetricCompatibility(t *testing.T) {
	c := &Collector{perfFrequency: 1000}
	c.describeClrInterop()

	row := perfDataClrInterop{
		NumberofCCWs:        11,
		NumberofStubs:       12,
		Numberofmarshalling: 13,
	}
	rows := make([]perfDataClrInterop, 0, 4)

	for _, name := range []string{"_Global_", "worker", "worker#1", "report_Total"} {
		row.Name = name
		rows = append(rows, row)
	}

	c.perfClrInterop = &fixturePerfCollector[perfDataClrInterop]{rows: rows}
	assertFixtureMetrics(t, c.collectClrInterop, []expectedMetric{
		{"windows_netframework_clrinterop_com_callable_wrappers_total", dto.MetricType_COUNTER, 11, ""},
		{"windows_netframework_clrinterop_interop_marshalling_total", dto.MetricType_COUNTER, 13, ""},
		{"windows_netframework_clrinterop_interop_stubs_created_total", dto.MetricType_COUNTER, 12, ""},
	}, false)
}

func TestClrJITMetricCompatibility(t *testing.T) {
	c := &Collector{perfFrequency: 1000}
	c.describeClrJIT()

	row := perfDataClrJIT{
		NumberofMethodsJitted:      11,
		PercentTimeinJit:           50,
		StandardJitFailures:        13,
		TotalNumberofILBytesJitted: 14,
	}
	rows := make([]perfDataClrJIT, 0, 4)

	for _, name := range []string{"_Global_", "worker", "worker#1", "report_Total"} {
		row.Name = name
		rows = append(rows, row)
	}

	c.perfClrJIT = &fixturePerfCollector[perfDataClrJIT]{rows: rows}
	assertFixtureMetrics(t, c.collectClrJIT, []expectedMetric{
		{"windows_netframework_clrjit_jit_methods_total", dto.MetricType_COUNTER, 11, ""},
		{"windows_netframework_clrjit_jit_time_percent", dto.MetricType_GAUGE, 0.05, ""},
		{"windows_netframework_clrjit_jit_standard_failures_total", dto.MetricType_GAUGE, 13, ""},
		{"windows_netframework_clrjit_jit_il_bytes_total", dto.MetricType_COUNTER, 14, ""},
	}, false)
}

func TestClrLoadingMetricCompatibility(t *testing.T) {
	c := &Collector{perfFrequency: 1000}
	c.describeClrLoading()

	row := perfDataClrLoading{
		BytesinLoaderHeap:         11,
		CurrentAssemblies:         12,
		CurrentClassesLoaded:      13,
		Currentappdomains:         14,
		TotalAppdomains:           15,
		TotalAssemblies:           16,
		TotalClassesLoaded:        17,
		TotalNumberofLoadFailures: 18,
		Totalappdomainsunloaded:   19,
	}
	rows := make([]perfDataClrLoading, 0, 4)

	for _, name := range []string{"_Global_", "worker", "worker#1", "report_Total"} {
		row.Name = name
		rows = append(rows, row)
	}

	c.perfClrLoading = &fixturePerfCollector[perfDataClrLoading]{rows: rows}
	assertFixtureMetrics(t, c.collectClrLoading, []expectedMetric{
		{"windows_netframework_clrloading_loader_heap_size_bytes", dto.MetricType_GAUGE, 11, ""},
		{"windows_netframework_clrloading_appdomains_loaded_current", dto.MetricType_GAUGE, 14, ""},
		{"windows_netframework_clrloading_assemblies_loaded_current", dto.MetricType_GAUGE, 12, ""},
		{"windows_netframework_clrloading_classes_loaded_current", dto.MetricType_GAUGE, 13, ""},
		{"windows_netframework_clrloading_appdomains_loaded_total", dto.MetricType_COUNTER, 15, ""},
		{"windows_netframework_clrloading_appdomains_unloaded_total", dto.MetricType_COUNTER, 19, ""},
		{"windows_netframework_clrloading_assemblies_loaded_total", dto.MetricType_COUNTER, 16, ""},
		{"windows_netframework_clrloading_classes_loaded_total", dto.MetricType_COUNTER, 17, ""},
		{"windows_netframework_clrloading_class_load_failures_total", dto.MetricType_COUNTER, 18, ""},
	}, false)
}

func TestClrLocksAndThreadsMetricCompatibility(t *testing.T) {
	c := &Collector{perfFrequency: 1000}
	c.describeClrLocksAndThreads()

	row := perfDataClrLocksAndThreads{
		CurrentQueueLength:               11,
		NumberofcurrentlogicalThreads:    12,
		NumberofcurrentphysicalThreads:   13,
		Numberofcurrentrecognizedthreads: 14,
		Numberoftotalrecognizedthreads:   15,
		QueueLengthPeak:                  16,
		TotalNumberofContentions:         17,
	}
	rows := make([]perfDataClrLocksAndThreads, 0, 4)

	for _, name := range []string{"_Global_", "worker", "worker#1", "report_Total"} {
		row.Name = name
		rows = append(rows, row)
	}

	c.perfClrLocksAndThreads = &fixturePerfCollector[perfDataClrLocksAndThreads]{rows: rows}
	assertFixtureMetrics(t, c.collectClrLocksAndThreads, []expectedMetric{
		{"windows_netframework_clrlocksandthreads_current_queue_length", dto.MetricType_GAUGE, 11, ""},
		{"windows_netframework_clrlocksandthreads_current_logical_threads", dto.MetricType_GAUGE, 12, ""},
		{"windows_netframework_clrlocksandthreads_physical_threads_current", dto.MetricType_GAUGE, 13, ""},
		{"windows_netframework_clrlocksandthreads_recognized_threads_current", dto.MetricType_GAUGE, 14, ""},
		{"windows_netframework_clrlocksandthreads_recognized_threads_total", dto.MetricType_COUNTER, 15, ""},
		{"windows_netframework_clrlocksandthreads_queue_length_total", dto.MetricType_COUNTER, 16, ""},
		{"windows_netframework_clrlocksandthreads_contentions_total", dto.MetricType_COUNTER, 17, ""},
	}, false)
}

func TestClrMemoryMetricCompatibility(t *testing.T) {
	c := &Collector{perfFrequency: 1000}
	c.describeClrMemory()

	row := perfDataClrMemory{
		AllocatedBytesPersec:      11,
		FinalizationSurvivors:     12,
		Gen0PromotedBytesPerSec:   13,
		Gen0heapsize:              14,
		Gen1PromotedBytesPerSec:   15,
		Gen1heapsize:              16,
		Gen2heapsize:              17,
		LargeObjectHeapsize:       18,
		NumberGCHandles:           19,
		NumberGen0Collections:     20,
		NumberGen1Collections:     21,
		NumberGen2Collections:     22,
		NumberInducedGC:           23,
		NumberTotalcommittedBytes: 24,
		NumberTotalreservedBytes:  25,
		NumberofPinnedObjects:     26,
		NumberofSinkBlocksinuse:   27,
		PercentTimeinGC:           50,
		PercentTimeinGC_base:      200,
		ProcessID:                 123,
	}
	rows := make([]perfDataClrMemory, 0, 4)

	for _, name := range []string{"_Global_", "worker", "worker#1", "report_Total"} {
		row.Name = name
		rows = append(rows, row)
	}

	c.perfClrMemory = &fixturePerfCollector[perfDataClrMemory]{rows: rows}
	assertFixtureMetrics(t, c.collectClrMemory, []expectedMetric{
		{"windows_netframework_clrmemory_allocated_bytes_total", dto.MetricType_COUNTER, 11, ""},
		{"windows_netframework_clrmemory_finalization_survivors", dto.MetricType_GAUGE, 12, ""},
		{"windows_netframework_clrmemory_heap_size_bytes", dto.MetricType_GAUGE, 14, "Gen0"},
		{"windows_netframework_clrmemory_promoted_bytes", dto.MetricType_GAUGE, 13, "Gen0"},
		{"windows_netframework_clrmemory_heap_size_bytes", dto.MetricType_GAUGE, 16, "Gen1"},
		{"windows_netframework_clrmemory_promoted_bytes", dto.MetricType_GAUGE, 15, "Gen1"},
		{"windows_netframework_clrmemory_heap_size_bytes", dto.MetricType_GAUGE, 17, "Gen2"},
		{"windows_netframework_clrmemory_heap_size_bytes", dto.MetricType_GAUGE, 18, "LOH"},
		{"windows_netframework_clrmemory_number_gc_handles", dto.MetricType_GAUGE, 19, ""},
		{"windows_netframework_clrmemory_collections_total", dto.MetricType_COUNTER, 20, "Gen0"},
		{"windows_netframework_clrmemory_collections_total", dto.MetricType_COUNTER, 21, "Gen1"},
		{"windows_netframework_clrmemory_collections_total", dto.MetricType_COUNTER, 22, "Gen2"},
		{"windows_netframework_clrmemory_induced_gc_total", dto.MetricType_COUNTER, 23, ""},
		{"windows_netframework_clrmemory_number_pinned_objects", dto.MetricType_GAUGE, 26, ""},
		{"windows_netframework_clrmemory_number_sink_blocksinuse", dto.MetricType_GAUGE, 27, ""},
		{"windows_netframework_clrmemory_committed_bytes", dto.MetricType_GAUGE, 24, ""},
		{"windows_netframework_clrmemory_reserved_bytes", dto.MetricType_GAUGE, 25, ""},
		{"windows_netframework_clrmemory_gc_time_percent", dto.MetricType_GAUGE, 25, ""},
	}, true)
}

func TestClrRemotingMetricCompatibility(t *testing.T) {
	c := &Collector{perfFrequency: 1000}
	c.describeClrRemoting()

	row := perfDataClrRemoting{
		Channels:                       11,
		ContextBoundClassesLoaded:      12,
		ContextBoundObjectsAllocPersec: 13,
		ContextProxies:                 14,
		Contexts:                       15,
		TotalRemoteCalls:               16,
	}
	rows := make([]perfDataClrRemoting, 0, 4)

	for _, name := range []string{"_Global_", "worker", "worker#1", "report_Total"} {
		row.Name = name
		rows = append(rows, row)
	}

	c.perfClrRemoting = &fixturePerfCollector[perfDataClrRemoting]{rows: rows}
	assertFixtureMetrics(t, c.collectClrRemoting, []expectedMetric{
		{"windows_netframework_clrremoting_channels_total", dto.MetricType_COUNTER, 11, ""},
		{"windows_netframework_clrremoting_context_bound_classes_loaded", dto.MetricType_GAUGE, 12, ""},
		{"windows_netframework_clrremoting_context_bound_objects_total", dto.MetricType_COUNTER, 13, ""},
		{"windows_netframework_clrremoting_context_proxies_total", dto.MetricType_COUNTER, 14, ""},
		{"windows_netframework_clrremoting_contexts", dto.MetricType_GAUGE, 15, ""},
		{"windows_netframework_clrremoting_remote_calls_total", dto.MetricType_COUNTER, 16, ""},
	}, false)
}

func TestClrSecurityMetricCompatibility(t *testing.T) {
	c := &Collector{perfFrequency: 1000}
	c.describeClrSecurity()

	row := perfDataClrSecurity{
		NumberLinkTimeChecks:  11,
		PercentTimeinRTchecks: 50,
		StackWalkDepth:        13,
		TotalRuntimeChecks:    14,
	}
	rows := make([]perfDataClrSecurity, 0, 4)

	for _, name := range []string{"_Global_", "worker", "worker#1", "report_Total"} {
		row.Name = name
		rows = append(rows, row)
	}

	c.perfClrSecurity = &fixturePerfCollector[perfDataClrSecurity]{rows: rows}
	assertFixtureMetrics(t, c.collectClrSecurity, []expectedMetric{
		{"windows_netframework_clrsecurity_link_time_checks_total", dto.MetricType_COUNTER, 11, ""},
		{"windows_netframework_clrsecurity_rt_checks_time_percent", dto.MetricType_GAUGE, 0.05, ""},
		{"windows_netframework_clrsecurity_stack_walk_depth", dto.MetricType_GAUGE, 13, ""},
		{"windows_netframework_clrsecurity_runtime_checks_total", dto.MetricType_COUNTER, 14, ""},
	}, false)
}

func TestGCZeroBaseDoesNotPublishInvalidRatio(t *testing.T) {
	c := &Collector{}
	c.describeClrMemory()
	c.perfClrMemory = &fixturePerfCollector[perfDataClrMemory]{rows: []perfDataClrMemory{{Name: "worker", ProcessID: 123, AllocatedBytesPersec: 99}}}
	registry := prometheus.NewRegistry()
	registry.MustRegister(fixturePromCollector{collect: func(ch chan<- prometheus.Metric) { require.NoError(t, c.collectClrMemory(ch, time.Second)) }})
	families, err := registry.Gather()
	require.NoError(t, err)

	foundAllocated := false

	for _, family := range families {
		require.NotEqual(t, "windows_netframework_clrmemory_gc_time_percent", family.GetName())

		if family.GetName() == "windows_netframework_clrmemory_allocated_bytes_total" {
			require.InDelta(t, 99.0, family.GetMetric()[0].GetCounter().GetValue(), 1e-12)

			foundAllocated = true
		}
	}

	require.True(t, foundAllocated)
}

func TestMissingObjectErrorClassification(t *testing.T) {
	missing := pdh.NewPdhError(pdh.CstatusNoObject)
	unrelated := errors.New("access denied")

	require.True(t, onlyMissingObject(fmt.Errorf("initialization: %w", errors.Join(missing, missing))))
	require.False(t, onlyMissingObject(errors.Join(missing, unrelated)))
	require.False(t, onlyMissingObject(fmt.Errorf("initialization: %w", errors.Join(missing, unrelated))))
	require.False(t, onlyMissingObject(pdh.NewPdhError(pdh.CstatusNoCounter)))
}

func TestNoDataIsReturned(t *testing.T) {
	c := &Collector{}
	c.describeClrExceptions()
	c.perfClrExceptions = unavailablePerfCollector[perfDataClrExceptions]{}
	ch := make(chan prometheus.Metric, 16)
	require.ErrorIs(t, c.collectClrExceptions(ch, time.Second), pdh.ErrNoData)
	require.Empty(t, ch)
}

func TestCloseReleasesInitializedCollectorOnce(t *testing.T) {
	c := &Collector{}
	fixture := &fixturePerfCollector[perfDataClrExceptions]{}
	c.closeFns = []func(){func() {
		if c.perfClrExceptions != nil {
			c.perfClrExceptions.Close()
		}
	}}
	c.perfClrExceptions = fixture
	require.NoError(t, c.Close())
	require.NoError(t, c.Close())
	require.Equal(t, 1, fixture.closed)
}

func BenchmarkClrMemoryMetricPublication(b *testing.B) {
	c := &Collector{}
	c.describeClrMemory()

	rows := make([]perfDataClrMemory, 100)
	for i := range rows {
		rows[i] = perfDataClrMemory{Name: fmt.Sprintf("worker#%d", i), ProcessID: float64(i), PercentTimeinGC: 50, PercentTimeinGC_base: 200}
	}

	c.perfClrMemory = &fixturePerfCollector[perfDataClrMemory]{rows: rows}
	ch := make(chan prometheus.Metric, 2100)

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		if err := c.collectClrMemory(ch, time.Second); err != nil {
			b.Fatal(err)
		}

		for len(ch) != 0 {
			<-ch
		}
	}
}

func TestCollectRetainsSuccessfulChildMetrics(t *testing.T) {
	c := &Collector{}
	c.describeClrExceptions()
	c.describeClrInterop()
	c.perfClrExceptions = &fixturePerfCollector[perfDataClrExceptions]{rows: []perfDataClrExceptions{{Name: "worker", NumberofExcepsThrown: 37}}}
	failure := errors.New("provider failed")
	c.perfClrInterop = &fixturePerfCollector[perfDataClrInterop]{err: failure}
	c.collectorFns = []func(chan<- prometheus.Metric, time.Duration) error{c.collectClrExceptions, c.collectClrInterop}
	registry := prometheus.NewRegistry()

	var collectErr error

	registry.MustRegister(fixturePromCollector{collect: func(ch chan<- prometheus.Metric) { collectErr = c.Collect(ch, time.Second) }})
	families, err := registry.Gather()
	require.NoError(t, err)
	require.ErrorIs(t, collectErr, failure)
	require.Len(t, families, 4)

	for _, family := range families {
		if family.GetName() == "windows_netframework_clrexceptions_exceptions_thrown_total" {
			require.InDelta(t, 37.0, family.GetMetric()[0].GetCounter().GetValue(), 1e-12)
		}
	}
}

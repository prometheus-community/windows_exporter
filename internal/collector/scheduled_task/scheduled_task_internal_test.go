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

package scheduled_task

import (
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/require"
)

func TestCollectMetrics(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		result     int64
		state      TaskState
		stateLabel string
		status     string
	}{
		{
			name: "success", result: 0,
			state: TASK_STATE_READY, stateLabel: "ready", status: "success",
		},
		{
			name: "application error", result: 1,
			state: TASK_STATE_READY, stateLabel: "ready", status: "unknown",
		},
		{
			name: "ready", result: 0x41300,
			state: TASK_STATE_READY, stateLabel: "ready", status: "ready",
		},
		{
			name: "running", result: 0x41301,
			state: TASK_STATE_RUNNING, stateLabel: "running", status: "running",
		},
		{
			name: "disabled", result: 0x41302,
			state: TASK_STATE_DISABLED, stateLabel: "disabled", status: "disabled",
		},
		{
			name: "never run", result: 0x41303,
			state: TASK_STATE_UNKNOWN, stateLabel: "unknown", status: "has_not_run",
		},
		{
			name: "no more runs", result: 0x41304,
			state: TASK_STATE_READY, stateLabel: "ready", status: "no_more_runs",
		},
		{
			name: "not scheduled", result: 0x41305,
			state: TASK_STATE_READY, stateLabel: "ready", status: "not_scheduled",
		},
		{
			name: "terminated", result: 0x41306,
			state: TASK_STATE_DISABLED, stateLabel: "disabled", status: "terminated",
		},
		{
			name: "no valid triggers", result: 0x41307,
			state: TASK_STATE_READY, stateLabel: "ready", status: "no_valid_triggers",
		},
		{
			name: "event trigger", result: 0x41308,
			state: TASK_STATE_READY, stateLabel: "ready", status: "event_trigger",
		},
		{
			name: "queued", result: 0x41325,
			state: TASK_STATE_QUEUED, stateLabel: "queued", status: "queued",
		},
		{
			name: "signed HRESULT", result: -2147216629,
			state: TASK_STATE_READY, stateLabel: "ready", status: "unknown",
		},
		{
			name: "unsigned HRESULT", result: 0x8004130B,
			state: TASK_STATE_READY, stateLabel: "ready", status: "unknown",
		},
		{
			name: "unrecognized result", result: -559038737,
			state: TASK_STATE_READY, stateLabel: "ready", status: "unknown",
		},
		{
			name: "maximum code", result: -1,
			state: TASK_STATE_READY, stateLabel: "ready", status: "unknown",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			task := ScheduledTask{
				Path:            "/Test/Task",
				State:           tc.state,
				MissedRunsCount: 3,
				LastTaskResult:  TaskResult(tc.result),
			}
			families := gatherTaskMetrics(t, nil, ScheduledTasks{task})

			states := families["windows_scheduled_task_state"]
			require.NotNil(t, states)
			require.Equal(t, dto.MetricType_GAUGE, states.GetType())
			require.Len(t, states.GetMetric(), 5)

			expectedStates := map[string]float64{"disabled": 0, "queued": 0, "ready": 0, "running": 0, "unknown": 0}
			expectedStates[tc.stateLabel] = 1

			for _, metric := range states.GetMetric() {
				require.Len(t, metric.GetLabel(), 2)
				require.Equal(t, "state", metric.GetLabel()[0].GetName())
				require.Equal(t, "task", metric.GetLabel()[1].GetName())
				require.Equal(t, task.Path, metric.GetLabel()[1].GetValue())
				state := metric.GetLabel()[0].GetValue()
				want, ok := expectedStates[state]
				require.True(t, ok, "unexpected or duplicate state %s", state)
				require.InDelta(t, want, metric.GetGauge().GetValue(), 0)
				delete(expectedStates, state)
			}

			require.Empty(t, expectedStates)

			statuses := families["windows_scheduled_task_last_result_status"]
			require.NotNil(t, statuses)
			require.Equal(t, dto.MetricType_GAUGE, statuses.GetType())
			require.Len(t, statuses.GetMetric(), 12)

			expectedStatuses := map[string]float64{
				"success": 0, "ready": 0, "running": 0, "disabled": 0,
				"has_not_run": 0, "no_more_runs": 0, "not_scheduled": 0, "terminated": 0,
				"no_valid_triggers": 0, "event_trigger": 0, "queued": 0, "unknown": 0,
			}
			expectedStatuses[tc.status] = 1

			for _, metric := range statuses.GetMetric() {
				require.Len(t, metric.GetLabel(), 2)
				require.Equal(t, "status", metric.GetLabel()[0].GetName())
				require.Equal(t, "task", metric.GetLabel()[1].GetName())
				require.Equal(t, task.Path, metric.GetLabel()[1].GetValue())
				status := metric.GetLabel()[0].GetValue()
				want, ok := expectedStatuses[status]
				require.True(t, ok, "unexpected or duplicate status %s", status)
				require.InDelta(t, want, metric.GetGauge().GetValue(), 0)
				delete(expectedStatuses, status)
			}

			require.Empty(t, expectedStatuses)

			if tc.result == 0x41303 {
				require.Len(t, families, 2)
				require.NotContains(t, families, "windows_scheduled_task_last_result")
				require.NotContains(t, families, "windows_scheduled_task_missed_runs")

				return
			}

			require.Len(t, families, 4)
			lastResult := families["windows_scheduled_task_last_result"]
			require.NotNil(t, lastResult)
			require.Len(t, lastResult.GetMetric(), 1)

			wantSuccess := 0.0
			if tc.result == 0 {
				wantSuccess = 1
			}

			require.InDelta(t, wantSuccess, lastResult.GetMetric()[0].GetGauge().GetValue(), 0)

			missedRuns := families["windows_scheduled_task_missed_runs"]
			require.NotNil(t, missedRuns)
			require.Len(t, missedRuns.GetMetric(), 1)
			require.InDelta(t, task.MissedRunsCount, missedRuns.GetMetric()[0].GetGauge().GetValue(), 0)
		})
	}
}

func gatherTaskMetrics(t *testing.T, config *Config, tasks ScheduledTasks) map[string]*dto.MetricFamily {
	t.Helper()

	c := New(config)
	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), nil))

	registry := prometheus.NewPedanticRegistry()
	require.NoError(t, registry.Register(taskMetricCollector{collector: c, tasks: tasks}))
	families, err := registry.Gather()
	require.NoError(t, err)

	result := make(map[string]*dto.MetricFamily, len(families))

	for _, family := range families {
		result[family.GetName()] = family
	}

	return result
}

type taskMetricCollector struct {
	collector *Collector
	tasks     ScheduledTasks
}

func (c taskMetricCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c taskMetricCollector) Collect(ch chan<- prometheus.Metric) {
	c.collector.collectMetrics(ch, c.tasks)
}

func TestCollectMetricsCache(t *testing.T) {
	t.Parallel()

	tasks := ScheduledTasks{
		{Path: "/Unchanged", State: TASK_STATE_READY, MissedRunsCount: 1, LastTaskResult: SCHED_S_SUCCESS},
		{Path: "/StateChanges", State: TASK_STATE_READY, LastTaskResult: SCHED_S_SUCCESS},
		{Path: "/ResultChanges", State: TASK_STATE_READY, LastTaskResult: SCHED_S_TASK_HAS_NOT_RUN},
		{Path: "/MissedRunsChange", State: TASK_STATE_READY, LastTaskResult: 1},
		{Path: "/Deleted", State: TASK_STATE_DISABLED, LastTaskResult: SCHED_S_TASK_DISABLED},
	}

	changed := ScheduledTasks{
		tasks[0],
		{Path: "/StateChanges", State: TASK_STATE_RUNNING, LastTaskResult: SCHED_S_SUCCESS},
		{Path: "/ResultChanges", State: TASK_STATE_READY, LastTaskResult: SCHED_S_SUCCESS},
		{Path: "/MissedRunsChange", State: TASK_STATE_READY, MissedRunsCount: 2, LastTaskResult: 1},
		{Path: "/Added", State: TASK_STATE_QUEUED, LastTaskResult: SCHED_S_TASK_QUEUED},
	}

	c := New(nil)
	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), nil))

	var previous map[string][]prometheus.Metric

	for i, scrape := range []ScheduledTasks{tasks, tasks, changed, changed} {
		// Collect once outside the registry, which calls Collect for Describe too.
		collected := make(chan prometheus.Metric, len(scrape)*19)
		c.collectMetrics(collected, scrape)
		close(collected)

		var sent []prometheus.Metric
		for metric := range collected {
			sent = append(sent, metric)
		}

		cached := exposition(t, func(ch chan<- prometheus.Metric) {
			for _, metric := range sent {
				ch <- metric
			}
		})

		uncached := exposition(t, func(ch chan<- prometheus.Metric) {
			for _, task := range scrape {
				for _, metric := range c.taskMetrics(task) {
					ch <- metric
				}
			}
		})

		require.Equal(t, uncached, cached, "scrape %d", i)

		// Unchanged tasks send the metrics of the previous scrape again.
		current := make(map[string][]prometheus.Metric, len(scrape))

		for _, task := range scrape {
			n := 19
			if task.LastTaskResult == SCHED_S_TASK_HAS_NOT_RUN {
				n = 17
			}

			current[task.Path], sent = sent[:n], sent[n:]

			// Scrapes 1 and 3 repeat the previous one, scrape 2 changes all tasks but /Unchanged.
			old, ok := previous[task.Path]
			wantReused := ok && (i%2 == 1 || task.Path == "/Unchanged")
			require.Equal(t, wantReused, ok && old[0] == current[task.Path][0], "scrape %d task %s", i, task.Path)
		}

		require.Empty(t, sent)

		previous = current
	}

	// Deleted tasks are dropped from the cache.
	scrape := c.metricCache.Begin()
	_, ok := scrape.Load("/Deleted", taskValues{state: TASK_STATE_DISABLED, lastTaskResult: SCHED_S_TASK_DISABLED})
	require.False(t, ok)
	_, ok = scrape.Load("/Added", taskValues{state: TASK_STATE_QUEUED, lastTaskResult: SCHED_S_TASK_QUEUED})
	require.True(t, ok)

	// Build creates new descriptors, so the cache must not return metrics built from the old ones.
	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), nil))

	scrape = c.metricCache.Begin()
	_, ok = scrape.Load("/Added", taskValues{state: TASK_STATE_QUEUED, lastTaskResult: SCHED_S_TASK_QUEUED})
	require.False(t, ok)
}

// exposition gathers the metrics sent by collect and returns them in the text format.
func exposition(t *testing.T, collect func(ch chan<- prometheus.Metric)) string {
	t.Helper()

	registry := prometheus.NewPedanticRegistry()
	require.NoError(t, registry.Register(collectorFunc(collect)))

	families, err := registry.Gather()
	require.NoError(t, err)

	var text strings.Builder

	for _, family := range families {
		_, err := expfmt.MetricFamilyToText(&text, family)
		require.NoError(t, err)
	}

	return text.String()
}

type collectorFunc func(ch chan<- prometheus.Metric)

func (f collectorFunc) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(f, ch)
}

func (f collectorFunc) Collect(ch chan<- prometheus.Metric) {
	f(ch)
}

// BenchmarkCollectMetrics measures building and sending the metrics of 265
// tasks, roughly the number of tasks on a Windows 11 workstation. With a warm
// cache the task values did not change since the last scrape; with a cold cache
// every metric is built again.
func BenchmarkCollectMetrics(b *testing.B) {
	for _, warm := range []bool{true, false} {
		b.Run(fmt.Sprintf("warm=%t", warm), func(b *testing.B) {
			c := New(nil)
			require.NoError(b, c.Build(slog.New(slog.DiscardHandler), nil))

			tasks := benchmarkTasks(265)
			ch := make(chan prometheus.Metric, len(tasks)*19)

			b.ReportAllocs()

			for b.Loop() {
				if !warm {
					c.metricCache.Reset()
				}

				c.collectMetrics(ch, tasks)

				for len(ch) > 0 {
					<-ch
				}
			}
		})
	}
}

func benchmarkTasks(n int) ScheduledTasks {
	results := []TaskResult{SCHED_S_SUCCESS, SCHED_S_TASK_READY, SCHED_S_TASK_HAS_NOT_RUN, 1}
	tasks := make(ScheduledTasks, n)

	for i := range tasks {
		tasks[i] = ScheduledTask{
			Path:            fmt.Sprintf("/Microsoft/Windows/Folder%d/Task%d", i/4, i),
			State:           TaskState(i % 5),
			MissedRunsCount: float64(i % 3),
			LastTaskResult:  results[i%len(results)],
		}
	}

	return tasks
}

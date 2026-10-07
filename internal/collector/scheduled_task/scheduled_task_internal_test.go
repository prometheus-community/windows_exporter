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
	"regexp"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
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

func TestCollectMetricsFilters(t *testing.T) {
	t.Parallel()

	tasks := ScheduledTasks{
		{Path: "/Included/Task", State: TASK_STATE_READY},
		{Path: "/Included/Excluded", State: TASK_STATE_READY},
		{Path: "/Other/Task", State: TASK_STATE_READY},
	}

	for _, tc := range []struct {
		name   string
		config *Config
		tasks  ScheduledTasks
		paths  []string
	}{
		{name: "empty", tasks: ScheduledTasks{}, paths: []string{}},
		{name: "defaults", tasks: tasks, paths: []string{"/Included/Task", "/Included/Excluded", "/Other/Task"}},
		{
			name:   "include",
			config: &Config{TaskInclude: regexp.MustCompile(`^/Included/`)},
			tasks:  tasks,
			paths:  []string{"/Included/Task", "/Included/Excluded"},
		},
		{
			name: "exclude takes precedence",
			config: &Config{
				TaskInclude: regexp.MustCompile(`^/Included/`),
				TaskExclude: regexp.MustCompile(`Excluded$`),
			},
			tasks: tasks,
			paths: []string{"/Included/Task"},
		},
		{
			name:   "all excluded",
			config: &Config{TaskExclude: regexp.MustCompile(`.*`)},
			tasks:  tasks,
			paths:  []string{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			families := gatherTaskMetrics(t, tc.config, tc.tasks)
			if len(tc.paths) == 0 {
				require.Empty(t, families)

				return
			}

			require.Len(t, families, 4)

			for _, family := range families {
				paths := make([]string, 0, len(family.GetMetric()))

				for _, metric := range family.GetMetric() {
					for _, label := range metric.GetLabel() {
						if label.GetName() == "task" {
							paths = append(paths, label.GetValue())
						}
					}
				}

				samplesPerTask := 1
				if family.GetName() == "windows_scheduled_task_state" {
					samplesPerTask = 5
				}

				if family.GetName() == "windows_scheduled_task_last_result_status" {
					samplesPerTask = 12
				}

				expected := make([]string, 0, samplesPerTask*len(tc.paths))
				for range samplesPerTask {
					expected = append(expected, tc.paths...)
				}

				require.ElementsMatch(t, expected, paths, "%s", family.GetName())
			}
		})
	}
}

func gatherTaskMetrics(t *testing.T, config *Config, tasks ScheduledTasks) map[string]*dto.MetricFamily {
	t.Helper()

	c := New(config)
	require.NoError(t, c.Build(nil, nil))

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

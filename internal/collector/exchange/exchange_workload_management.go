// Copyright 2024 The Prometheus Authors
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

package exchange

import (
	"errors"
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	activeTasks    = "ActiveTasks"
	completedTasks = "CompletedTasks"
	queuedTasks    = "QueuedTasks"
	yieldedTasks   = "YieldedTasks"
	isActive       = "Active"
)

func (c *Collector) buildWorkloadManagementWorkloads() error {
	counters := []string{
		activeTasks,
		completedTasks,
		queuedTasks,
		yieldedTasks,
		isActive,
	}

	var err error

	c.perfDataCollectorWorkloadManagementWorkloads, err = perfdata.NewCollector("MSExchange WorkloadManagement Workloads", perfdata.InstancesAll, counters)
	if err != nil {
		return fmt.Errorf("failed to create MSExchange WorkloadManagement Workloads collector: %w", err)
	}

	c.activeTasks = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "workload_active_tasks"),
		"Number of active tasks currently running in the background for workload management",
		[]string{"name"},
		nil,
	)
	c.completedTasks = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "workload_completed_tasks"),
		"Number of workload management tasks that have been completed",
		[]string{"name"},
		nil,
	)
	c.queuedTasks = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "workload_queued_tasks"),
		"Number of workload management tasks that are currently queued up waiting to be processed",
		[]string{"name"},
		nil,
	)
	c.yieldedTasks = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "workload_yielded_tasks"),
		"The total number of tasks that have been yielded by a workload",
		[]string{"name"},
		nil,
	)
	c.isActive = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "workload_is_active"),
		"Active indicates whether the workload is in an active (1) or paused (0) state",
		[]string{"name"},
		nil,
	)

	return nil
}

func (c *Collector) collectWorkloadManagementWorkloads(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollectorWorkloadManagementWorkloads.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect MSExchange WorkloadManagement Workloads: %w", err)
	}

	if len(perfData) == 0 {
		return errors.New("perflib query for MSExchange WorkloadManagement Workloads returned empty result set")
	}

	for name, data := range perfData {
		labelName := c.toLabelName(name)

		ch <- prometheus.MustNewConstMetric(
			c.activeTasks,
			prometheus.GaugeValue,
			data[activeTasks].FirstValue,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.completedTasks,
			prometheus.CounterValue,
			data[completedTasks].FirstValue,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.queuedTasks,
			prometheus.CounterValue,
			data[queuedTasks].FirstValue,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.yieldedTasks,
			prometheus.CounterValue,
			data[yieldedTasks].FirstValue,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.isActive,
			prometheus.GaugeValue,
			data[isActive].FirstValue,
			labelName,
		)
	}

	return nil
}

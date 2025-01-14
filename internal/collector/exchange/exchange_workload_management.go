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
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorWorkloadManagementWorkloads struct {
	perfDataCollectorWorkloadManagementWorkloads *pdh.Collector
	perfDataObjectWorkloadManagementWorkloads    []perfDataCounterValuesWorkloadManagementWorkloads

	activeTasks    *prometheus.Desc
	isActive       *prometheus.Desc
	completedTasks *prometheus.Desc
	queuedTasks    *prometheus.Desc
	yieldedTasks   *prometheus.Desc
}

type perfDataCounterValuesWorkloadManagementWorkloads struct {
	Name string

	ActiveTasks    float64 `perfdata:"ActiveTasks"`
	CompletedTasks float64 `perfdata:"CompletedTasks"`
	QueuedTasks    float64 `perfdata:"QueuedTasks"`
	YieldedTasks   float64 `perfdata:"YieldedTasks"`
	IsActive       float64 `perfdata:"Active"`
}

func (c *Collector) buildWorkloadManagementWorkloads() error {
	var err error

	c.perfDataCollectorWorkloadManagementWorkloads, err = pdh.NewCollector[perfDataCounterValuesWorkloadManagementWorkloads](pdh.CounterTypeRaw, "MSExchange WorkloadManagement Workloads", pdh.InstancesAll)
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
	err := c.perfDataCollectorWorkloadManagementWorkloads.Collect(&c.perfDataObjectWorkloadManagementWorkloads)
	if err != nil {
		return fmt.Errorf("failed to collect MSExchange WorkloadManagement Workloads: %w", err)
	}

	for _, data := range c.perfDataObjectWorkloadManagementWorkloads {
		labelName := c.toLabelName(data.Name)

		ch <- prometheus.MustNewConstMetric(
			c.activeTasks,
			prometheus.GaugeValue,
			data.ActiveTasks,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.completedTasks,
			prometheus.CounterValue,
			data.CompletedTasks,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.queuedTasks,
			prometheus.CounterValue,
			data.QueuedTasks,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.yieldedTasks,
			prometheus.CounterValue,
			data.YieldedTasks,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.isActive,
			prometheus.GaugeValue,
			data.IsActive,
			labelName,
		)
	}

	return nil
}

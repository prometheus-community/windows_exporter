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

package mssql

import (
	"errors"
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorWaitStats struct {
	waitStatsPerfDataCollectors map[mssqlInstance]*pdh.Collector
	waitStatsPerfDataObject     []perfDataCounterValuesWaitStats

	waitStatsLockWaits                     *prometheus.Desc
	waitStatsMemoryGrantQueueWaits         *prometheus.Desc
	waitStatsThreadSafeMemoryObjectsWaits  *prometheus.Desc
	waitStatsLogWriteWaits                 *prometheus.Desc
	waitStatsLogBufferWaits                *prometheus.Desc
	waitStatsNetworkIOWaits                *prometheus.Desc
	waitStatsPageIOLatchWaits              *prometheus.Desc
	waitStatsPageLatchWaits                *prometheus.Desc
	waitStatsNonPageLatchWaits             *prometheus.Desc
	waitStatsWaitForTheWorkerWaits         *prometheus.Desc
	waitStatsWorkspaceSynchronizationWaits *prometheus.Desc
	waitStatsTransactionOwnershipWaits     *prometheus.Desc
}

type perfDataCounterValuesWaitStats struct {
	Name string

	WaitStatsLockWaits                     float64 `perfdata:"Lock waits"`
	WaitStatsMemoryGrantQueueWaits         float64 `perfdata:"Memory grant queue waits"`
	WaitStatsThreadSafeMemoryObjectsWaits  float64 `perfdata:"Thread-safe memory objects waits"`
	WaitStatsLogWriteWaits                 float64 `perfdata:"Log write waits"`
	WaitStatsLogBufferWaits                float64 `perfdata:"Log buffer waits"`
	WaitStatsNetworkIOWaits                float64 `perfdata:"Network IO waits"`
	WaitStatsPageIOLatchWaits              float64 `perfdata:"Page IO latch waits"`
	WaitStatsPageLatchWaits                float64 `perfdata:"Page latch waits"`
	WaitStatsNonpageLatchWaits             float64 `perfdata:"Non-Page latch waits"`
	WaitStatsWaitForTheWorkerWaits         float64 `perfdata:"Wait for the worker"`
	WaitStatsWorkspaceSynchronizationWaits float64 `perfdata:"Workspace synchronization waits"`
	WaitStatsTransactionOwnershipWaits     float64 `perfdata:"Transaction ownership waits"`
}

func (c *Collector) buildWaitStats() error {
	var err error

	c.waitStatsPerfDataCollectors = make(map[mssqlInstance]*pdh.Collector, len(c.mssqlInstances))
	errs := make([]error, 0, len(c.mssqlInstances))

	for _, sqlInstance := range c.mssqlInstances {
		c.waitStatsPerfDataCollectors[sqlInstance], err = pdh.NewCollector[perfDataCounterValuesWaitStats](pdh.CounterTypeRaw, c.mssqlGetPerfObjectName(sqlInstance, "Wait Statistics"), pdh.InstancesAll)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to create Wait Statistics collector for instance %s: %w", sqlInstance.name, err))
		}
	}

	c.waitStatsLockWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_lock_waits"),
		"(WaitStats.LockWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)

	c.waitStatsMemoryGrantQueueWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_memory_grant_queue_waits"),
		"(WaitStats.MemoryGrantQueueWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)
	c.waitStatsThreadSafeMemoryObjectsWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_thread_safe_memory_objects_waits"),
		"(WaitStats.ThreadSafeMemoryObjectsWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)
	c.waitStatsLogWriteWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_log_write_waits"),
		"(WaitStats.LogWriteWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)
	c.waitStatsLogBufferWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_log_buffer_waits"),
		"(WaitStats.LogBufferWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)
	c.waitStatsNetworkIOWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_network_io_waits"),
		"(WaitStats.NetworkIOWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)
	c.waitStatsPageIOLatchWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_page_io_latch_waits"),
		"(WaitStats.PageIOLatchWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)
	c.waitStatsPageLatchWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_page_latch_waits"),
		"(WaitStats.PageLatchWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)
	c.waitStatsNonPageLatchWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_nonpage_latch_waits"),
		"(WaitStats.NonpageLatchWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)
	c.waitStatsWaitForTheWorkerWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_wait_for_the_worker_waits"),
		"(WaitStats.WaitForTheWorkerWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)
	c.waitStatsWorkspaceSynchronizationWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_workspace_synchronization_waits"),
		"(WaitStats.WorkspaceSynchronizationWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)
	c.waitStatsTransactionOwnershipWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "waitstats_transaction_ownership_waits"),
		"(WaitStats.TransactionOwnershipWaits)",
		[]string{"mssql_instance", "item"},
		nil,
	)

	return errors.Join(errs...)
}

func (c *Collector) collectWaitStats(ch chan<- prometheus.Metric) error {
	return c.collect(ch, subCollectorWaitStats, c.waitStatsPerfDataCollectors, c.collectWaitStatsInstance)
}

func (c *Collector) collectWaitStatsInstance(ch chan<- prometheus.Metric, sqlInstance mssqlInstance, perfDataCollector *pdh.Collector) error {
	err := perfDataCollector.Collect(&c.waitStatsPerfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect %s metrics: %w", c.mssqlGetPerfObjectName(sqlInstance, "Wait Statistics"), err)
	}

	for _, data := range c.waitStatsPerfDataObject {
		ch <- prometheus.MustNewConstMetric(
			c.waitStatsLockWaits,
			prometheus.CounterValue,
			data.WaitStatsLockWaits,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsMemoryGrantQueueWaits,
			prometheus.CounterValue,
			data.WaitStatsMemoryGrantQueueWaits,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsThreadSafeMemoryObjectsWaits,
			prometheus.CounterValue,
			data.WaitStatsThreadSafeMemoryObjectsWaits,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsLogWriteWaits,
			prometheus.CounterValue,
			data.WaitStatsLogWriteWaits,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsLogBufferWaits,
			prometheus.CounterValue,
			data.WaitStatsLogBufferWaits,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsNetworkIOWaits,
			prometheus.CounterValue,
			data.WaitStatsNetworkIOWaits,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsPageIOLatchWaits,
			prometheus.CounterValue,
			data.WaitStatsPageIOLatchWaits,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsPageLatchWaits,
			prometheus.CounterValue,
			data.WaitStatsPageLatchWaits,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsNonPageLatchWaits,
			prometheus.CounterValue,
			data.WaitStatsNonpageLatchWaits,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsWaitForTheWorkerWaits,
			prometheus.CounterValue,
			data.WaitStatsWaitForTheWorkerWaits,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsWorkspaceSynchronizationWaits,
			prometheus.CounterValue,
			data.WaitStatsWorkspaceSynchronizationWaits,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsTransactionOwnershipWaits,
			prometheus.CounterValue,
			data.WaitStatsTransactionOwnershipWaits,
			sqlInstance.name, data.Name,
		)
	}

	return nil
}

func (c *Collector) closeWaitStats() {
	for _, perfDataCollector := range c.waitStatsPerfDataCollectors {
		perfDataCollector.Close()
	}
}

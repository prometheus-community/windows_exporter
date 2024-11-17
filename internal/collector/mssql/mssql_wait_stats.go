//go:build windows

package mssql

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorWaitStats struct {
	waitStatsPerfDataCollectors map[string]*perfdata.Collector

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

const (
	waitStatsLockWaits                     = "Lock waits"
	waitStatsMemoryGrantQueueWaits         = "Memory grant queue waits"
	waitStatsThreadSafeMemoryObjectsWaits  = "Thread-safe memory objects waits"
	waitStatsLogWriteWaits                 = "Log write waits"
	waitStatsLogBufferWaits                = "Log buffer waits"
	waitStatsNetworkIOWaits                = "Network IO waits"
	waitStatsPageIOLatchWaits              = "Page IO latch waits"
	waitStatsPageLatchWaits                = "Page latch waits"
	waitStatsNonpageLatchWaits             = "Non-Page latch waits"
	waitStatsWaitForTheWorkerWaits         = "Wait for the worker"
	waitStatsWorkspaceSynchronizationWaits = "Workspace synchronization waits"
	waitStatsTransactionOwnershipWaits     = "Transaction ownership waits"
)

func (c *Collector) buildWaitStats() error {
	var err error

	c.waitStatsPerfDataCollectors = make(map[string]*perfdata.Collector, len(c.mssqlInstances))
	counters := []string{
		waitStatsLockWaits,
		waitStatsMemoryGrantQueueWaits,
		waitStatsThreadSafeMemoryObjectsWaits,
		waitStatsLogWriteWaits,
		waitStatsLogBufferWaits,
		waitStatsNetworkIOWaits,
		waitStatsPageIOLatchWaits,
		waitStatsPageLatchWaits,
		waitStatsNonpageLatchWaits,
		waitStatsWaitForTheWorkerWaits,
		waitStatsWorkspaceSynchronizationWaits,
		waitStatsTransactionOwnershipWaits,
	}

	for sqlInstance := range c.mssqlInstances {
		c.waitStatsPerfDataCollectors[sqlInstance], err = perfdata.NewCollector(c.mssqlGetPerfObjectName(sqlInstance, "Wait Statistics"), perfdata.InstanceAll, counters)
		if err != nil {
			return fmt.Errorf("failed to create Wait Statistics collector for instance %s: %w", sqlInstance, err)
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

	return nil
}

func (c *Collector) collectWaitStats(ch chan<- prometheus.Metric) error {
	return c.collect(ch, subCollectorWaitStats, c.waitStatsPerfDataCollectors, c.collectWaitStatsInstance)
}

func (c *Collector) collectWaitStatsInstance(ch chan<- prometheus.Metric, sqlInstance string, perfDataCollector *perfdata.Collector) error {
	perfData, err := perfDataCollector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect %s metrics: %w", c.mssqlGetPerfObjectName(sqlInstance, "Wait Statistics"), err)
	}

	for item, data := range perfData {
		ch <- prometheus.MustNewConstMetric(
			c.waitStatsLockWaits,
			prometheus.CounterValue,
			data[waitStatsLockWaits].FirstValue,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsMemoryGrantQueueWaits,
			prometheus.CounterValue,
			data[waitStatsMemoryGrantQueueWaits].FirstValue,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsThreadSafeMemoryObjectsWaits,
			prometheus.CounterValue,
			data[waitStatsThreadSafeMemoryObjectsWaits].FirstValue,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsLogWriteWaits,
			prometheus.CounterValue,
			data[waitStatsLogWriteWaits].FirstValue,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsLogBufferWaits,
			prometheus.CounterValue,
			data[waitStatsLogBufferWaits].FirstValue,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsNetworkIOWaits,
			prometheus.CounterValue,
			data[waitStatsNetworkIOWaits].FirstValue,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsPageIOLatchWaits,
			prometheus.CounterValue,
			data[waitStatsPageIOLatchWaits].FirstValue,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsPageLatchWaits,
			prometheus.CounterValue,
			data[waitStatsPageLatchWaits].FirstValue,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsNonPageLatchWaits,
			prometheus.CounterValue,
			data[waitStatsNonpageLatchWaits].FirstValue,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsWaitForTheWorkerWaits,
			prometheus.CounterValue,
			data[waitStatsWaitForTheWorkerWaits].FirstValue,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsWorkspaceSynchronizationWaits,
			prometheus.CounterValue,
			data[waitStatsWorkspaceSynchronizationWaits].FirstValue,
			sqlInstance, item,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitStatsTransactionOwnershipWaits,
			prometheus.CounterValue,
			data[waitStatsTransactionOwnershipWaits].FirstValue,
			sqlInstance, item,
		)
	}

	return nil
}

func (c *Collector) closeWaitStats() {
	for _, perfDataCollector := range c.waitStatsPerfDataCollectors {
		perfDataCollector.Close()
	}
}

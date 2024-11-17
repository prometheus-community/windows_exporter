//go:build windows

package mssql

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorMemoryManager struct {
	memMgrPerfDataCollectors map[string]*perfdata.Collector

	memMgrConnectionMemoryKB       *prometheus.Desc
	memMgrDatabaseCacheMemoryKB    *prometheus.Desc
	memMgrExternalBenefitOfMemory  *prometheus.Desc
	memMgrFreeMemoryKB             *prometheus.Desc
	memMgrGrantedWorkspaceMemoryKB *prometheus.Desc
	memMgrLockBlocks               *prometheus.Desc
	memMgrLockBlocksAllocated      *prometheus.Desc
	memMgrLockMemoryKB             *prometheus.Desc
	memMgrLockOwnerBlocks          *prometheus.Desc
	memMgrLockOwnerBlocksAllocated *prometheus.Desc
	memMgrLogPoolMemoryKB          *prometheus.Desc
	memMgrMaximumWorkspaceMemoryKB *prometheus.Desc
	memMgrMemoryGrantsOutstanding  *prometheus.Desc
	memMgrMemoryGrantsPending      *prometheus.Desc
	memMgrOptimizerMemoryKB        *prometheus.Desc
	memMgrReservedServerMemoryKB   *prometheus.Desc
	memMgrSQLCacheMemoryKB         *prometheus.Desc
	memMgrStolenServerMemoryKB     *prometheus.Desc
	memMgrTargetServerMemoryKB     *prometheus.Desc
	memMgrTotalServerMemoryKB      *prometheus.Desc
}

const (
	memMgrConnectionMemoryKB       = "Connection Memory (KB)"
	memMgrDatabaseCacheMemoryKB    = "Database Cache Memory (KB)"
	memMgrExternalBenefitOfMemory  = "External benefit of memory"
	memMgrFreeMemoryKB             = "Free Memory (KB)"
	memMgrGrantedWorkspaceMemoryKB = "Granted Workspace Memory (KB)"
	memMgrLockBlocks               = "Lock Blocks"
	memMgrLockBlocksAllocated      = "Lock Blocks Allocated"
	memMgrLockMemoryKB             = "Lock Memory (KB)"
	memMgrLockOwnerBlocks          = "Lock Owner Blocks"
	memMgrLockOwnerBlocksAllocated = "Lock Owner Blocks Allocated"
	memMgrLogPoolMemoryKB          = "Log Pool Memory (KB)"
	memMgrMaximumWorkspaceMemoryKB = "Maximum Workspace Memory (KB)"
	memMgrMemoryGrantsOutstanding  = "Memory Grants Outstanding"
	memMgrMemoryGrantsPending      = "Memory Grants Pending"
	memMgrOptimizerMemoryKB        = "Optimizer Memory (KB)"
	memMgrReservedServerMemoryKB   = "Reserved Server Memory (KB)"
	memMgrSQLCacheMemoryKB         = "SQL Cache Memory (KB)"
	memMgrStolenServerMemoryKB     = "Stolen Server Memory (KB)"
	memMgrTargetServerMemoryKB     = "Target Server Memory (KB)"
	memMgrTotalServerMemoryKB      = "Total Server Memory (KB)"
)

func (c *Collector) buildMemoryManager() error {
	var err error

	c.memMgrPerfDataCollectors = make(map[string]*perfdata.Collector, len(c.mssqlInstances))
	counters := []string{
		memMgrConnectionMemoryKB,
		memMgrDatabaseCacheMemoryKB,
		memMgrExternalBenefitOfMemory,
		memMgrFreeMemoryKB,
		memMgrGrantedWorkspaceMemoryKB,
		memMgrLockBlocks,
		memMgrLockBlocksAllocated,
		memMgrLockMemoryKB,
		memMgrLockOwnerBlocks,
		memMgrLockOwnerBlocksAllocated,
		memMgrLogPoolMemoryKB,
		memMgrMaximumWorkspaceMemoryKB,
		memMgrMemoryGrantsOutstanding,
		memMgrMemoryGrantsPending,
		memMgrOptimizerMemoryKB,
		memMgrReservedServerMemoryKB,
		memMgrSQLCacheMemoryKB,
		memMgrStolenServerMemoryKB,
		memMgrTargetServerMemoryKB,
		memMgrTotalServerMemoryKB,
	}

	for sqlInstance := range c.mssqlInstances {
		c.memMgrPerfDataCollectors[sqlInstance], err = perfdata.NewCollector(c.mssqlGetPerfObjectName(sqlInstance, "Memory Manager"), perfdata.InstanceAll, counters)
		if err != nil {
			return fmt.Errorf("failed to create Locks collector for instance %s: %w", sqlInstance, err)
		}
	}

	c.memMgrConnectionMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_connection_memory_bytes"),
		"(MemoryManager.ConnectionMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrDatabaseCacheMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_database_cache_memory_bytes"),
		"(MemoryManager.DatabaseCacheMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrExternalBenefitOfMemory = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_external_benefit_of_memory"),
		"(MemoryManager.Externalbenefitofmemory)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrFreeMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_free_memory_bytes"),
		"(MemoryManager.FreeMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrGrantedWorkspaceMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_granted_workspace_memory_bytes"),
		"(MemoryManager.GrantedWorkspaceMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrLockBlocks = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_lock_blocks"),
		"(MemoryManager.LockBlocks)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrLockBlocksAllocated = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_allocated_lock_blocks"),
		"(MemoryManager.LockBlocksAllocated)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrLockMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_lock_memory_bytes"),
		"(MemoryManager.LockMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrLockOwnerBlocks = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_lock_owner_blocks"),
		"(MemoryManager.LockOwnerBlocks)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrLockOwnerBlocksAllocated = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_allocated_lock_owner_blocks"),
		"(MemoryManager.LockOwnerBlocksAllocated)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrLogPoolMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_log_pool_memory_bytes"),
		"(MemoryManager.LogPoolMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrMaximumWorkspaceMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_maximum_workspace_memory_bytes"),
		"(MemoryManager.MaximumWorkspaceMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrMemoryGrantsOutstanding = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_outstanding_memory_grants"),
		"(MemoryManager.MemoryGrantsOutstanding)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrMemoryGrantsPending = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_pending_memory_grants"),
		"(MemoryManager.MemoryGrantsPending)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrOptimizerMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_optimizer_memory_bytes"),
		"(MemoryManager.OptimizerMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrReservedServerMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_reserved_server_memory_bytes"),
		"(MemoryManager.ReservedServerMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrSQLCacheMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_sql_cache_memory_bytes"),
		"(MemoryManager.SQLCacheMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrStolenServerMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_stolen_server_memory_bytes"),
		"(MemoryManager.StolenServerMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrTargetServerMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_target_server_memory_bytes"),
		"(MemoryManager.TargetServerMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.memMgrTotalServerMemoryKB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memmgr_total_server_memory_bytes"),
		"(MemoryManager.TotalServerMemoryKB)",
		[]string{"mssql_instance"},
		nil,
	)

	return nil
}

func (c *Collector) collectMemoryManager(ch chan<- prometheus.Metric) error {
	return c.collect(ch, subCollectorMemoryManager, c.memMgrPerfDataCollectors, c.collectMemoryManagerInstance)
}

func (c *Collector) collectMemoryManagerInstance(ch chan<- prometheus.Metric, sqlInstance string, perfDataCollector *perfdata.Collector) error {
	perfData, err := perfDataCollector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect %s metrics: %w", c.mssqlGetPerfObjectName(sqlInstance, "Memory Manager"), err)
	}

	data, ok := perfData[perfdata.EmptyInstance]
	if !ok {
		return fmt.Errorf("perflib query for %s returned empty result set", c.mssqlGetPerfObjectName(sqlInstance, "Memory Manager"))
	}

	ch <- prometheus.MustNewConstMetric(
		c.memMgrConnectionMemoryKB,
		prometheus.GaugeValue,
		data[memMgrConnectionMemoryKB].FirstValue*1024,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrDatabaseCacheMemoryKB,
		prometheus.GaugeValue,
		data[memMgrDatabaseCacheMemoryKB].FirstValue*1024,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrExternalBenefitOfMemory,
		prometheus.GaugeValue,
		data[memMgrExternalBenefitOfMemory].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrFreeMemoryKB,
		prometheus.GaugeValue,
		data[memMgrFreeMemoryKB].FirstValue*1024,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrGrantedWorkspaceMemoryKB,
		prometheus.GaugeValue,
		data[memMgrGrantedWorkspaceMemoryKB].FirstValue*1024,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrLockBlocks,
		prometheus.GaugeValue,
		data[memMgrLockBlocks].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrLockBlocksAllocated,
		prometheus.GaugeValue,
		data[memMgrLockBlocksAllocated].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrLockMemoryKB,
		prometheus.GaugeValue,
		data[memMgrLockMemoryKB].FirstValue*1024,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrLockOwnerBlocks,
		prometheus.GaugeValue,
		data[memMgrLockOwnerBlocks].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrLockOwnerBlocksAllocated,
		prometheus.GaugeValue,
		data[memMgrLockOwnerBlocksAllocated].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrLogPoolMemoryKB,
		prometheus.GaugeValue,
		data[memMgrLogPoolMemoryKB].FirstValue*1024,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrMaximumWorkspaceMemoryKB,
		prometheus.GaugeValue,
		data[memMgrMaximumWorkspaceMemoryKB].FirstValue*1024,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrMemoryGrantsOutstanding,
		prometheus.GaugeValue,
		data[memMgrMemoryGrantsOutstanding].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrMemoryGrantsPending,
		prometheus.GaugeValue,
		data[memMgrMemoryGrantsPending].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrOptimizerMemoryKB,
		prometheus.GaugeValue,
		data[memMgrOptimizerMemoryKB].FirstValue*1024,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrReservedServerMemoryKB,
		prometheus.GaugeValue,
		data[memMgrReservedServerMemoryKB].FirstValue*1024,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrSQLCacheMemoryKB,
		prometheus.GaugeValue,
		data[memMgrSQLCacheMemoryKB].FirstValue*1024,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrStolenServerMemoryKB,
		prometheus.GaugeValue,
		data[memMgrStolenServerMemoryKB].FirstValue*1024,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrTargetServerMemoryKB,
		prometheus.GaugeValue,
		data[memMgrTargetServerMemoryKB].FirstValue*1024,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrTotalServerMemoryKB,
		prometheus.GaugeValue,
		data[memMgrTotalServerMemoryKB].FirstValue*1024,
		sqlInstance,
	)

	return nil
}

func (c *Collector) closeMemoryManager() {
	for _, perfDataCollector := range c.memMgrPerfDataCollectors {
		perfDataCollector.Close()
	}
}

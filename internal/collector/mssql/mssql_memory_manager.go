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

type collectorMemoryManager struct {
	memMgrPerfDataCollectors map[mssqlInstance]*pdh.Collector
	memMgrPerfDataObject     []perfDataCounterValuesMemMgr

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

type perfDataCounterValuesMemMgr struct {
	MemMgrConnectionMemoryKB       float64 `perfdata:"Connection Memory (KB)"`
	MemMgrDatabaseCacheMemoryKB    float64 `perfdata:"Database Cache Memory (KB)"`
	MemMgrExternalBenefitOfMemory  float64 `perfdata:"External benefit of memory"`
	MemMgrFreeMemoryKB             float64 `perfdata:"Free Memory (KB)"`
	MemMgrGrantedWorkspaceMemoryKB float64 `perfdata:"Granted Workspace Memory (KB)"`
	MemMgrLockBlocks               float64 `perfdata:"Lock Blocks"`
	MemMgrLockBlocksAllocated      float64 `perfdata:"Lock Blocks Allocated"`
	MemMgrLockMemoryKB             float64 `perfdata:"Lock Memory (KB)"`
	MemMgrLockOwnerBlocks          float64 `perfdata:"Lock Owner Blocks"`
	MemMgrLockOwnerBlocksAllocated float64 `perfdata:"Lock Owner Blocks Allocated"`
	MemMgrLogPoolMemoryKB          float64 `perfdata:"Log Pool Memory (KB)"`
	MemMgrMaximumWorkspaceMemoryKB float64 `perfdata:"Maximum Workspace Memory (KB)"`
	MemMgrMemoryGrantsOutstanding  float64 `perfdata:"Memory Grants Outstanding"`
	MemMgrMemoryGrantsPending      float64 `perfdata:"Memory Grants Pending"`
	MemMgrOptimizerMemoryKB        float64 `perfdata:"Optimizer Memory (KB)"`
	MemMgrReservedServerMemoryKB   float64 `perfdata:"Reserved Server Memory (KB)"`
	MemMgrSQLCacheMemoryKB         float64 `perfdata:"SQL Cache Memory (KB)"`
	MemMgrStolenServerMemoryKB     float64 `perfdata:"Stolen Server Memory (KB)"`
	MemMgrTargetServerMemoryKB     float64 `perfdata:"Target Server Memory (KB)"`
	MemMgrTotalServerMemoryKB      float64 `perfdata:"Total Server Memory (KB)"`
}

func (c *Collector) buildMemoryManager() error {
	var err error

	c.memMgrPerfDataCollectors = make(map[mssqlInstance]*pdh.Collector, len(c.mssqlInstances))
	errs := make([]error, 0, len(c.mssqlInstances))

	for _, sqlInstance := range c.mssqlInstances {
		c.memMgrPerfDataCollectors[sqlInstance], err = pdh.NewCollector[perfDataCounterValuesMemMgr](pdh.CounterTypeRaw, c.mssqlGetPerfObjectName(sqlInstance, "Memory Manager"), pdh.InstancesAll)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to create Memory Manager collector for instance %s: %w", sqlInstance.name, err))
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

	return errors.Join(errs...)
}

func (c *Collector) collectMemoryManager(ch chan<- prometheus.Metric) error {
	return c.collect(ch, subCollectorMemoryManager, c.memMgrPerfDataCollectors, c.collectMemoryManagerInstance)
}

func (c *Collector) collectMemoryManagerInstance(ch chan<- prometheus.Metric, sqlInstance mssqlInstance, perfDataCollector *pdh.Collector) error {
	err := perfDataCollector.Collect(&c.memMgrPerfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect %s metrics: %w", c.mssqlGetPerfObjectName(sqlInstance, "Memory Manager"), err)
	}

	ch <- prometheus.MustNewConstMetric(
		c.memMgrConnectionMemoryKB,
		prometheus.GaugeValue,
		c.memMgrPerfDataObject[0].MemMgrConnectionMemoryKB*1024,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrDatabaseCacheMemoryKB,
		prometheus.GaugeValue,
		c.memMgrPerfDataObject[0].MemMgrDatabaseCacheMemoryKB*1024,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrExternalBenefitOfMemory,
		prometheus.GaugeValue,
		c.memMgrPerfDataObject[0].MemMgrExternalBenefitOfMemory,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrFreeMemoryKB,
		prometheus.GaugeValue,
		c.memMgrPerfDataObject[0].MemMgrFreeMemoryKB*1024,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrGrantedWorkspaceMemoryKB,
		prometheus.GaugeValue,
		c.memMgrPerfDataObject[0].MemMgrGrantedWorkspaceMemoryKB*1024,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrLockBlocks,
		prometheus.GaugeValue,
		c.memMgrPerfDataObject[0].MemMgrLockBlocks,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrLockBlocksAllocated,
		prometheus.GaugeValue,
		c.memMgrPerfDataObject[0].MemMgrLockBlocksAllocated,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrLockMemoryKB,
		prometheus.GaugeValue,
		c.memMgrPerfDataObject[0].MemMgrLockMemoryKB*1024,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrLockOwnerBlocks,
		prometheus.GaugeValue,
		c.memMgrPerfDataObject[0].MemMgrLockOwnerBlocks,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrLockOwnerBlocksAllocated,
		prometheus.GaugeValue,
		c.memMgrPerfDataObject[0].MemMgrLockOwnerBlocksAllocated,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrLogPoolMemoryKB,
		prometheus.GaugeValue,
		c.memMgrPerfDataObject[0].MemMgrLogPoolMemoryKB*1024,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrMaximumWorkspaceMemoryKB,
		prometheus.GaugeValue,
		c.memMgrPerfDataObject[0].MemMgrMaximumWorkspaceMemoryKB*1024,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrMemoryGrantsOutstanding,
		prometheus.GaugeValue,
		c.memMgrPerfDataObject[0].MemMgrMemoryGrantsOutstanding,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrMemoryGrantsPending,
		prometheus.GaugeValue,
		c.memMgrPerfDataObject[0].MemMgrMemoryGrantsPending,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrOptimizerMemoryKB,
		prometheus.GaugeValue,
		c.memMgrPerfDataObject[0].MemMgrOptimizerMemoryKB*1024,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrReservedServerMemoryKB,
		prometheus.GaugeValue,
		c.memMgrPerfDataObject[0].MemMgrReservedServerMemoryKB*1024,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrSQLCacheMemoryKB,
		prometheus.GaugeValue,
		c.memMgrPerfDataObject[0].MemMgrSQLCacheMemoryKB*1024,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrStolenServerMemoryKB,
		prometheus.GaugeValue,
		c.memMgrPerfDataObject[0].MemMgrStolenServerMemoryKB*1024,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrTargetServerMemoryKB,
		prometheus.GaugeValue,
		c.memMgrPerfDataObject[0].MemMgrTargetServerMemoryKB*1024,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMgrTotalServerMemoryKB,
		prometheus.GaugeValue,
		c.memMgrPerfDataObject[0].MemMgrTotalServerMemoryKB*1024,
		sqlInstance.name,
	)

	return nil
}

func (c *Collector) closeMemoryManager() {
	for _, perfDataCollector := range c.memMgrPerfDataCollectors {
		perfDataCollector.Close()
	}
}

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

package mssql

import (
	"errors"
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorTransactions struct {
	transactionsPerfDataCollectors map[mssqlInstance]*pdh.Collector[perfDataCounterValuesTransactions]
	transactionsPerfDataObject     []perfDataCounterValuesTransactions

	transactionsTempDbFreeSpaceBytes             *prometheus.Desc
	transactionsLongestTransactionRunningSeconds *prometheus.Desc
	transactionsNonSnapshotVersionActive         *prometheus.Desc
	transactionsSnapshotActive                   *prometheus.Desc
	transactionsActive                           *prometheus.Desc
	transactionsUpdateConflictsTotal             *prometheus.Desc
	transactionsUpdateSnapshotTotal              *prometheus.Desc
	transactionsUpdateSnapshotActive             *prometheus.Desc
	transactionsVersionCleanupBytes              *prometheus.Desc
	transactionsVersionGenerationBytes           *prometheus.Desc
	transactionsVersionStoreSizeBytes            *prometheus.Desc
	transactionsVersionStoreUnits                *prometheus.Desc
	transactionsVersionStoreCreationUnits        *prometheus.Desc
	transactionsVersionStoreTruncationUnits      *prometheus.Desc
}

type perfDataCounterValuesTransactions struct {
	TransactionsFreeSpaceintempdbKB            float64 `perfdata:"Free Space in tempdb (KB)"`
	TransactionsLongestTransactionRunningTime  float64 `perfdata:"Longest Transaction Running Time"`
	TransactionsNonSnapshotVersionTransactions float64 `perfdata:"NonSnapshot Version Transactions"`
	TransactionsSnapshotTransactions           float64 `perfdata:"Snapshot Transactions"`
	TransactionsTransactions                   float64 `perfdata:"Transactions"`
	TransactionsUpdateconflictratio            float64 `perfdata:"Update conflict ratio"`
	TransactionsUpdateconflictratioBase        float64 `perfdata:"Update conflict ratio,secondvalue"`
	TransactionsUpdateSnapshotTransactions     float64 `perfdata:"Update Snapshot Transactions"`
	TransactionsVersionCleanuprateKBPers       float64 `perfdata:"Version Cleanup rate (KB/s)"`
	TransactionsVersionGenerationrateKBPers    float64 `perfdata:"Version Generation rate (KB/s)"`
	TransactionsVersionStoreSizeKB             float64 `perfdata:"Version Store Size (KB)"`
	TransactionsVersionStoreunitcount          float64 `perfdata:"Version Store unit count"`
	TransactionsVersionStoreunitcreation       float64 `perfdata:"Version Store unit creation"`
	TransactionsVersionStoreunittruncation     float64 `perfdata:"Version Store unit truncation"`
}

func (c *Collector) buildTransactions() error {
	var err error

	c.transactionsPerfDataCollectors = make(map[mssqlInstance]*pdh.Collector[perfDataCounterValuesTransactions], len(c.mssqlInstances))
	errs := make([]error, 0, len(c.mssqlInstances))

	for _, sqlInstance := range c.mssqlInstances {
		c.transactionsPerfDataCollectors[sqlInstance], err = pdh.NewCollector[perfDataCounterValuesTransactions](c.logger, pdh.CounterTypeRaw, c.mssqlGetPerfObjectName(sqlInstance, "Transactions"), nil)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to create Transactions collector for instance %s: %w", sqlInstance.name, err))
		}
	}

	c.transactionsTempDbFreeSpaceBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_tempdb_free_space_bytes"),
		"(Transactions.FreeSpaceInTempDbKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.transactionsLongestTransactionRunningSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_longest_transaction_running_seconds"),
		"(Transactions.LongestTransactionRunningTime)",
		[]string{"mssql_instance"},
		nil,
	)
	c.transactionsNonSnapshotVersionActive = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_nonsnapshot_version_active"),
		"(Transactions.NonSnapshotVersionTransactions)",
		[]string{"mssql_instance"},
		nil,
	)
	c.transactionsSnapshotActive = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_snapshot_active"),
		"(Transactions.SnapshotTransactions)",
		[]string{"mssql_instance"},
		nil,
	)
	c.transactionsActive = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_active"),
		"(Transactions.Transactions)",
		[]string{"mssql_instance"},
		nil,
	)
	c.transactionsUpdateConflictsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_update_conflicts_total"),
		"(Transactions.UpdateConflictRatio) Total number of update snapshot transactions that had an update conflict",
		[]string{"mssql_instance"},
		nil,
	)
	c.transactionsUpdateSnapshotTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_update_snapshot_transactions_total"),
		"(Transactions.UpdateConflictRatio_Base) Total number of update snapshot transactions",
		[]string{"mssql_instance"},
		nil,
	)
	c.transactionsUpdateSnapshotActive = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_update_snapshot_active"),
		"(Transactions.UpdateSnapshotTransactions)",
		[]string{"mssql_instance"},
		nil,
	)
	c.transactionsVersionCleanupBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_version_cleanup_bytes_total"),
		"(Transactions.VersionCleanupRateKBs) Total size of row versions removed from the version store, in bytes",
		[]string{"mssql_instance"},
		nil,
	)
	c.transactionsVersionGenerationBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_version_generation_bytes_total"),
		"(Transactions.VersionGenerationRateKBs) Total size of row versions added to the version store, in bytes",
		[]string{"mssql_instance"},
		nil,
	)
	c.transactionsVersionStoreSizeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_version_store_size_bytes"),
		"(Transactions.VersionStoreSizeKB)",
		[]string{"mssql_instance"},
		nil,
	)
	c.transactionsVersionStoreUnits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_version_store_units"),
		"(Transactions.VersionStoreUnitCount)",
		[]string{"mssql_instance"},
		nil,
	)
	c.transactionsVersionStoreCreationUnits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_version_store_creation_units"),
		"(Transactions.VersionStoreUnitCreation)",
		[]string{"mssql_instance"},
		nil,
	)
	c.transactionsVersionStoreTruncationUnits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_version_store_truncation_units"),
		"(Transactions.VersionStoreUnitTruncation)",
		[]string{"mssql_instance"},
		nil,
	)

	return errors.Join(errs...)
}

func (c *Collector) collectTransactions(ch chan<- prometheus.Metric) error {
	return c.collect(ch, subCollectorTransactions, c.transactionsPerfDataCollectors, c.collectTransactionsInstance)
}

// Win32_PerfRawData_MSSQLSERVER_Transactions docs:
// - https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-transactions-object
func (c *Collector) collectTransactionsInstance(ch chan<- prometheus.Metric, sqlInstance mssqlInstance, perfDataCollector *pdh.Collector[perfDataCounterValuesTransactions]) error {
	err := perfDataCollector.Collect(&c.transactionsPerfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect %s metrics: %w", c.mssqlGetPerfObjectName(sqlInstance, "Transactions"), err)
	}

	c.emitTransactionsMetrics(ch, sqlInstance)

	return nil
}

func (c *Collector) emitTransactionsMetrics(ch chan<- prometheus.Metric, sqlInstance mssqlInstance) {
	ch <- prometheus.MustNewConstMetric(
		c.transactionsTempDbFreeSpaceBytes,
		prometheus.GaugeValue,
		c.transactionsPerfDataObject[0].TransactionsFreeSpaceintempdbKB*1024,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.transactionsLongestTransactionRunningSeconds,
		prometheus.GaugeValue,
		c.transactionsPerfDataObject[0].TransactionsLongestTransactionRunningTime,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.transactionsNonSnapshotVersionActive,
		prometheus.GaugeValue,
		c.transactionsPerfDataObject[0].TransactionsNonSnapshotVersionTransactions,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.transactionsSnapshotActive,
		prometheus.GaugeValue,
		c.transactionsPerfDataObject[0].TransactionsSnapshotTransactions,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.transactionsActive,
		prometheus.GaugeValue,
		c.transactionsPerfDataObject[0].TransactionsTransactions,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.transactionsUpdateConflictsTotal,
		prometheus.CounterValue,
		c.transactionsPerfDataObject[0].TransactionsUpdateconflictratio,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.transactionsUpdateSnapshotTotal,
		prometheus.CounterValue,
		c.transactionsPerfDataObject[0].TransactionsUpdateconflictratioBase,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.transactionsUpdateSnapshotActive,
		prometheus.GaugeValue,
		c.transactionsPerfDataObject[0].TransactionsUpdateSnapshotTransactions,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.transactionsVersionCleanupBytes,
		prometheus.CounterValue,
		c.transactionsPerfDataObject[0].TransactionsVersionCleanuprateKBPers*1024,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.transactionsVersionGenerationBytes,
		prometheus.CounterValue,
		c.transactionsPerfDataObject[0].TransactionsVersionGenerationrateKBPers*1024,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.transactionsVersionStoreSizeBytes,
		prometheus.GaugeValue,
		c.transactionsPerfDataObject[0].TransactionsVersionStoreSizeKB*1024,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.transactionsVersionStoreUnits,
		prometheus.GaugeValue,
		c.transactionsPerfDataObject[0].TransactionsVersionStoreunitcount,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.transactionsVersionStoreCreationUnits,
		prometheus.CounterValue,
		c.transactionsPerfDataObject[0].TransactionsVersionStoreunitcreation,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.transactionsVersionStoreTruncationUnits,
		prometheus.CounterValue,
		c.transactionsPerfDataObject[0].TransactionsVersionStoreunittruncation,
		sqlInstance.name,
	)
}

func (c *Collector) closeTransactions() {
	for _, perfDataCollector := range c.transactionsPerfDataCollectors {
		perfDataCollector.Close()
	}
}

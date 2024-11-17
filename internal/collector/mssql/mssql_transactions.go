//go:build windows

package mssql

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorTransactions struct {
	transactionsPerfDataCollectors map[string]*perfdata.Collector

	transactionsTempDbFreeSpaceBytes             *prometheus.Desc
	transactionsLongestTransactionRunningSeconds *prometheus.Desc
	transactionsNonSnapshotVersionActiveTotal    *prometheus.Desc
	transactionsSnapshotActiveTotal              *prometheus.Desc
	transactionsActive                           *prometheus.Desc
	transactionsUpdateConflictsTotal             *prometheus.Desc
	transactionsUpdateSnapshotActiveTotal        *prometheus.Desc
	transactionsVersionCleanupRateBytes          *prometheus.Desc
	transactionsVersionGenerationRateBytes       *prometheus.Desc
	transactionsVersionStoreSizeBytes            *prometheus.Desc
	transactionsVersionStoreUnits                *prometheus.Desc
	transactionsVersionStoreCreationUnits        *prometheus.Desc
	transactionsVersionStoreTruncationUnits      *prometheus.Desc
}

const (
	transactionsFreeSpaceintempdbKB            = "Free Space in tempdb (KB)"
	transactionsLongestTransactionRunningTime  = "Longest Transaction Running Time"
	transactionsNonSnapshotVersionTransactions = "NonSnapshot Version Transactions"
	transactionsSnapshotTransactions           = "Snapshot Transactions"
	transactionsTransactions                   = "Transactions"
	transactionsUpdateconflictratio            = "Update conflict ratio"
	transactionsUpdateSnapshotTransactions     = "Update Snapshot Transactions"
	transactionsVersionCleanuprateKBPers       = "Version Cleanup rate (KB/s)"
	transactionsVersionGenerationrateKBPers    = "Version Generation rate (KB/s)"
	transactionsVersionStoreSizeKB             = "Version Store Size (KB)"
	transactionsVersionStoreunitcount          = "Version Store unit count"
	transactionsVersionStoreunitcreation       = "Version Store unit creation"
	transactionsVersionStoreunittruncation     = "Version Store unit truncation"
)

func (c *Collector) buildTransactions() error {
	var err error

	c.transactionsPerfDataCollectors = make(map[string]*perfdata.Collector, len(c.mssqlInstances))
	counters := []string{
		transactionsFreeSpaceintempdbKB,
		transactionsLongestTransactionRunningTime,
		transactionsNonSnapshotVersionTransactions,
		transactionsSnapshotTransactions,
		transactionsTransactions,
		transactionsUpdateconflictratio,
		transactionsUpdateSnapshotTransactions,
		transactionsVersionCleanuprateKBPers,
		transactionsVersionGenerationrateKBPers,
		transactionsVersionStoreSizeKB,
		transactionsVersionStoreunitcount,
		transactionsVersionStoreunitcreation,
		transactionsVersionStoreunittruncation,
	}

	for sqlInstance := range c.mssqlInstances {
		c.transactionsPerfDataCollectors[sqlInstance], err = perfdata.NewCollector(c.mssqlGetPerfObjectName(sqlInstance, "Transactions"), nil, counters)
		if err != nil {
			return fmt.Errorf("failed to create Transactions collector for instance %s: %w", sqlInstance, err)
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
	c.transactionsNonSnapshotVersionActiveTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_nonsnapshot_version_active_total"),
		"(Transactions.NonSnapshotVersionTransactions)",
		[]string{"mssql_instance"},
		nil,
	)
	c.transactionsSnapshotActiveTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_snapshot_active_total"),
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
		"(Transactions.UpdateConflictRatio)",
		[]string{"mssql_instance"},
		nil,
	)
	c.transactionsUpdateSnapshotActiveTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_update_snapshot_active_total"),
		"(Transactions.UpdateSnapshotTransactions)",
		[]string{"mssql_instance"},
		nil,
	)
	c.transactionsVersionCleanupRateBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_version_cleanup_rate_bytes"),
		"(Transactions.VersionCleanupRateKBs)",
		[]string{"mssql_instance"},
		nil,
	)
	c.transactionsVersionGenerationRateBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transactions_version_generation_rate_bytes"),
		"(Transactions.VersionGenerationRateKBs)",
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

	return nil
}

func (c *Collector) collectTransactions(ch chan<- prometheus.Metric) error {
	return c.collect(ch, subCollectorTransactions, c.transactionsPerfDataCollectors, c.collectTransactionsInstance)
}

// Win32_PerfRawData_MSSQLSERVER_Transactions docs:
// - https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-transactions-object
func (c *Collector) collectTransactionsInstance(ch chan<- prometheus.Metric, sqlInstance string, perfDataCollector *perfdata.Collector) error {
	perfData, err := perfDataCollector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect %s metrics: %w", c.mssqlGetPerfObjectName(sqlInstance, "Transactions"), err)
	}

	data, ok := perfData[perfdata.EmptyInstance]
	if !ok {
		return fmt.Errorf("perflib query for %s returned empty result set", c.mssqlGetPerfObjectName(sqlInstance, "Transactions"))
	}

	ch <- prometheus.MustNewConstMetric(
		c.transactionsTempDbFreeSpaceBytes,
		prometheus.GaugeValue,
		data[transactionsFreeSpaceintempdbKB].FirstValue*1024,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.transactionsLongestTransactionRunningSeconds,
		prometheus.GaugeValue,
		data[transactionsLongestTransactionRunningTime].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.transactionsNonSnapshotVersionActiveTotal,
		prometheus.CounterValue,
		data[transactionsNonSnapshotVersionTransactions].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.transactionsSnapshotActiveTotal,
		prometheus.CounterValue,
		data[transactionsSnapshotTransactions].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.transactionsActive,
		prometheus.GaugeValue,
		data[transactionsTransactions].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.transactionsUpdateConflictsTotal,
		prometheus.CounterValue,
		data[transactionsUpdateconflictratio].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.transactionsUpdateSnapshotActiveTotal,
		prometheus.CounterValue,
		data[transactionsUpdateSnapshotTransactions].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.transactionsVersionCleanupRateBytes,
		prometheus.GaugeValue,
		data[transactionsVersionCleanuprateKBPers].FirstValue*1024,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.transactionsVersionGenerationRateBytes,
		prometheus.GaugeValue,
		data[transactionsVersionGenerationrateKBPers].FirstValue*1024,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.transactionsVersionStoreSizeBytes,
		prometheus.GaugeValue,
		data[transactionsVersionStoreSizeKB].FirstValue*1024,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.transactionsVersionStoreUnits,
		prometheus.CounterValue,
		data[transactionsVersionStoreunitcount].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.transactionsVersionStoreCreationUnits,
		prometheus.CounterValue,
		data[transactionsVersionStoreunitcreation].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.transactionsVersionStoreTruncationUnits,
		prometheus.CounterValue,
		data[transactionsVersionStoreunittruncation].FirstValue,
		sqlInstance,
	)

	return nil
}

func (c *Collector) closeTransactions() {
	for _, perfDataCollector := range c.transactionsPerfDataCollectors {
		perfDataCollector.Close()
	}
}

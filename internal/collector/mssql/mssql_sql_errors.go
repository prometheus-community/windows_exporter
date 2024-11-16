//go:build windows

package mssql

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorSQLErrors struct {
	sqlErrorsPerfDataCollectors map[string]*perfdata.Collector

	// Win32_PerfRawData_{instance}_SQLServerSQLErrors
	sqlErrorsTotal *prometheus.Desc
}

const (
	sqlErrorsErrorsPerSec = "Errors/sec"
)

func (c *Collector) buildSQLErrors() error {
	var err error

	c.genStatsPerfDataCollectors = make(map[string]*perfdata.Collector, len(c.mssqlInstances))
	counters := []string{
		sqlErrorsErrorsPerSec,
	}

	for sqlInstance := range c.mssqlInstances {
		c.genStatsPerfDataCollectors[sqlInstance], err = perfdata.NewCollector(c.mssqlGetPerfObjectName(sqlInstance, "SQL Errors"), perfdata.InstanceAll, counters)
		if err != nil {
			return fmt.Errorf("failed to create SQL Errors collector for instance %s: %w", sqlInstance, err)
		}
	}

	// Win32_PerfRawData_{instance}_SQLServerSQLErrors
	c.sqlErrorsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sql_errors_total"),
		"(SQLErrors.Total)",
		[]string{"mssql_instance", "resource"},
		nil,
	)

	return nil
}

func (c *Collector) collectSQLErrors(ch chan<- prometheus.Metric) error {
	return c.collect(ch, subCollectorSQLErrors, c.dbReplicaPerfDataCollectors, c.collectSQLErrorsInstance)
}

func (c *Collector) collectSQLErrorsInstance(ch chan<- prometheus.Metric, sqlInstance string, perfDataCollector *perfdata.Collector) error {
	perfData, err := perfDataCollector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect %s metrics: %w", c.mssqlGetPerfObjectName(sqlInstance, "SQL Errors"), err)
	}

	for resource, data := range perfData {
		ch <- prometheus.MustNewConstMetric(
			c.sqlErrorsTotal,
			prometheus.CounterValue,
			data[sqlErrorsErrorsPerSec].FirstValue,
			sqlInstance, resource,
		)
	}

	return nil
}

func (c *Collector) closeSQLErrors() {
	for _, perfDataCollector := range c.sqlErrorsPerfDataCollectors {
		perfDataCollector.Close()
	}
}

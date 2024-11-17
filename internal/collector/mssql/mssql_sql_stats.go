//go:build windows

package mssql

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorSQLStats struct {
	sqlStatsPerfDataCollectors map[string]*perfdata.Collector

	sqlStatsAutoParamAttempts       *prometheus.Desc
	sqlStatsBatchRequests           *prometheus.Desc
	sqlStatsFailedAutoParams        *prometheus.Desc
	sqlStatsForcedParameterizations *prometheus.Desc
	sqlStatsGuidedplanexecutions    *prometheus.Desc
	sqlStatsMisguidedplanexecutions *prometheus.Desc
	sqlStatsSafeAutoParams          *prometheus.Desc
	sqlStatsSQLAttentionrate        *prometheus.Desc
	sqlStatsSQLCompilations         *prometheus.Desc
	sqlStatsSQLReCompilations       *prometheus.Desc
	sqlStatsUnsafeAutoParams        *prometheus.Desc
}

const (
	sqlStatsAutoParamAttemptsPerSec       = "Auto-Param Attempts/sec"
	sqlStatsBatchRequestsPerSec           = "Batch Requests/sec"
	sqlStatsFailedAutoParamsPerSec        = "Failed Auto-Params/sec"
	sqlStatsForcedParameterizationsPerSec = "Forced Parameterizations/sec"
	sqlStatsGuidedplanexecutionsPerSec    = "Guided plan executions/sec"
	sqlStatsMisguidedplanexecutionsPerSec = "Misguided plan executions/sec"
	sqlStatsSafeAutoParamsPerSec          = "Safe Auto-Params/sec"
	sqlStatsSQLAttentionrate              = "SQL Attention rate"
	sqlStatsSQLCompilationsPerSec         = "SQL Compilations/sec"
	sqlStatsSQLReCompilationsPerSec       = "SQL Re-Compilations/sec"
	sqlStatsUnsafeAutoParamsPerSec        = "Unsafe Auto-Params/sec"
)

func (c *Collector) buildSQLStats() error {
	var err error

	c.genStatsPerfDataCollectors = make(map[string]*perfdata.Collector, len(c.mssqlInstances))
	counters := []string{
		sqlStatsAutoParamAttemptsPerSec,
		sqlStatsBatchRequestsPerSec,
		sqlStatsFailedAutoParamsPerSec,
		sqlStatsForcedParameterizationsPerSec,
		sqlStatsGuidedplanexecutionsPerSec,
		sqlStatsMisguidedplanexecutionsPerSec,
		sqlStatsSafeAutoParamsPerSec,
		sqlStatsSQLAttentionrate,
		sqlStatsSQLCompilationsPerSec,
		sqlStatsSQLReCompilationsPerSec,
		sqlStatsUnsafeAutoParamsPerSec,
	}

	for sqlInstance := range c.mssqlInstances {
		c.genStatsPerfDataCollectors[sqlInstance], err = perfdata.NewCollector(c.mssqlGetPerfObjectName(sqlInstance, "SQL Statistics"), nil, counters)
		if err != nil {
			return fmt.Errorf("failed to create SQL Statistics collector for instance %s: %w", sqlInstance, err)
		}
	}

	c.sqlStatsAutoParamAttempts = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_auto_parameterization_attempts"),
		"(SQLStatistics.AutoParamAttempts)",
		[]string{"mssql_instance"},
		nil,
	)
	c.sqlStatsBatchRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_batch_requests"),
		"(SQLStatistics.BatchRequests)",
		[]string{"mssql_instance"},
		nil,
	)
	c.sqlStatsFailedAutoParams = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_failed_auto_parameterization_attempts"),
		"(SQLStatistics.FailedAutoParams)",
		[]string{"mssql_instance"},
		nil,
	)
	c.sqlStatsForcedParameterizations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_forced_parameterizations"),
		"(SQLStatistics.ForcedParameterizations)",
		[]string{"mssql_instance"},
		nil,
	)
	c.sqlStatsGuidedplanexecutions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_guided_plan_executions"),
		"(SQLStatistics.Guidedplanexecutions)",
		[]string{"mssql_instance"},
		nil,
	)
	c.sqlStatsMisguidedplanexecutions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_misguided_plan_executions"),
		"(SQLStatistics.Misguidedplanexecutions)",
		[]string{"mssql_instance"},
		nil,
	)
	c.sqlStatsSafeAutoParams = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_safe_auto_parameterization_attempts"),
		"(SQLStatistics.SafeAutoParams)",
		[]string{"mssql_instance"},
		nil,
	)
	c.sqlStatsSQLAttentionrate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_sql_attentions"),
		"(SQLStatistics.SQLAttentions)",
		[]string{"mssql_instance"},
		nil,
	)
	c.sqlStatsSQLCompilations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_sql_compilations"),
		"(SQLStatistics.SQLCompilations)",
		[]string{"mssql_instance"},
		nil,
	)
	c.sqlStatsSQLReCompilations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_sql_recompilations"),
		"(SQLStatistics.SQLReCompilations)",
		[]string{"mssql_instance"},
		nil,
	)
	c.sqlStatsUnsafeAutoParams = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sqlstats_unsafe_auto_parameterization_attempts"),
		"(SQLStatistics.UnsafeAutoParams)",
		[]string{"mssql_instance"},
		nil,
	)

	return nil
}

func (c *Collector) collectSQLStats(ch chan<- prometheus.Metric) error {
	return c.collect(ch, subCollectorSQLStats, c.sqlStatsPerfDataCollectors, c.collectSQLStatsInstance)
}

func (c *Collector) collectSQLStatsInstance(ch chan<- prometheus.Metric, sqlInstance string, perfDataCollector *perfdata.Collector) error {
	perfData, err := perfDataCollector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect %s metrics: %w", c.mssqlGetPerfObjectName(sqlInstance, "SQL Statistics"), err)
	}

	data, ok := perfData[perfdata.EmptyInstance]
	if !ok {
		return fmt.Errorf("perflib query for %s returned empty result set", c.mssqlGetPerfObjectName(sqlInstance, "SQL Statistics"))
	}

	ch <- prometheus.MustNewConstMetric(
		c.sqlStatsAutoParamAttempts,
		prometheus.CounterValue,
		data[sqlStatsAutoParamAttemptsPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.sqlStatsBatchRequests,
		prometheus.CounterValue,
		data[sqlStatsBatchRequestsPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.sqlStatsFailedAutoParams,
		prometheus.CounterValue,
		data[sqlStatsFailedAutoParamsPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.sqlStatsForcedParameterizations,
		prometheus.CounterValue,
		data[sqlStatsForcedParameterizationsPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.sqlStatsGuidedplanexecutions,
		prometheus.CounterValue,
		data[sqlStatsGuidedplanexecutionsPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.sqlStatsMisguidedplanexecutions,
		prometheus.CounterValue,
		data[sqlStatsMisguidedplanexecutionsPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.sqlStatsSafeAutoParams,
		prometheus.CounterValue,
		data[sqlStatsSafeAutoParamsPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.sqlStatsSQLAttentionrate,
		prometheus.CounterValue,
		data[sqlStatsSQLAttentionrate].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.sqlStatsSQLCompilations,
		prometheus.CounterValue,
		data[sqlStatsSQLCompilationsPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.sqlStatsSQLReCompilations,
		prometheus.CounterValue,
		data[sqlStatsSQLReCompilationsPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.sqlStatsUnsafeAutoParams,
		prometheus.CounterValue,
		data[sqlStatsUnsafeAutoParamsPerSec].FirstValue,
		sqlInstance,
	)

	return nil
}

func (c *Collector) closeSQLStats() {
	for _, perfDataCollector := range c.sqlStatsPerfDataCollectors {
		perfDataCollector.Close()
	}
}

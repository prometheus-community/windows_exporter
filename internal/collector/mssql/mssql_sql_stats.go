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

type collectorSQLStats struct {
	sqlStatsPerfDataCollectors map[mssqlInstance]*pdh.Collector
	sqlStatsPerfDataObject     []perfDataCounterValuesSqlStats

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

type perfDataCounterValuesSqlStats struct {
	SqlStatsAutoParamAttemptsPerSec       float64 `perfdata:"Auto-Param Attempts/sec"`
	SqlStatsBatchRequestsPerSec           float64 `perfdata:"Batch Requests/sec"`
	SqlStatsFailedAutoParamsPerSec        float64 `perfdata:"Failed Auto-Params/sec"`
	SqlStatsForcedParameterizationsPerSec float64 `perfdata:"Forced Parameterizations/sec"`
	SqlStatsGuidedplanexecutionsPerSec    float64 `perfdata:"Guided plan executions/sec"`
	SqlStatsMisguidedplanexecutionsPerSec float64 `perfdata:"Misguided plan executions/sec"`
	SqlStatsSafeAutoParamsPerSec          float64 `perfdata:"Safe Auto-Params/sec"`
	SqlStatsSQLAttentionrate              float64 `perfdata:"SQL Attention rate"`
	SqlStatsSQLCompilationsPerSec         float64 `perfdata:"SQL Compilations/sec"`
	SqlStatsSQLReCompilationsPerSec       float64 `perfdata:"SQL Re-Compilations/sec"`
	SqlStatsUnsafeAutoParamsPerSec        float64 `perfdata:"Unsafe Auto-Params/sec"`
}

func (c *Collector) buildSQLStats() error {
	var err error

	c.sqlStatsPerfDataCollectors = make(map[mssqlInstance]*pdh.Collector, len(c.mssqlInstances))
	errs := make([]error, 0, len(c.mssqlInstances))

	for _, sqlInstance := range c.mssqlInstances {
		c.sqlStatsPerfDataCollectors[sqlInstance], err = pdh.NewCollector[perfDataCounterValuesSqlStats](pdh.CounterTypeRaw, c.mssqlGetPerfObjectName(sqlInstance, "SQL Statistics"), nil)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to create SQL Statistics collector for instance %s: %w", sqlInstance.name, err))
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

	return errors.Join(errs...)
}

func (c *Collector) collectSQLStats(ch chan<- prometheus.Metric) error {
	return c.collect(ch, subCollectorSQLStats, c.sqlStatsPerfDataCollectors, c.collectSQLStatsInstance)
}

func (c *Collector) collectSQLStatsInstance(ch chan<- prometheus.Metric, sqlInstance mssqlInstance, perfDataCollector *pdh.Collector) error {
	err := perfDataCollector.Collect(&c.sqlStatsPerfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect %s metrics: %w", c.mssqlGetPerfObjectName(sqlInstance, "SQL Statistics"), err)
	}

	ch <- prometheus.MustNewConstMetric(
		c.sqlStatsAutoParamAttempts,
		prometheus.CounterValue,
		c.sqlStatsPerfDataObject[0].SqlStatsAutoParamAttemptsPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.sqlStatsBatchRequests,
		prometheus.CounterValue,
		c.sqlStatsPerfDataObject[0].SqlStatsBatchRequestsPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.sqlStatsFailedAutoParams,
		prometheus.CounterValue,
		c.sqlStatsPerfDataObject[0].SqlStatsFailedAutoParamsPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.sqlStatsForcedParameterizations,
		prometheus.CounterValue,
		c.sqlStatsPerfDataObject[0].SqlStatsForcedParameterizationsPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.sqlStatsGuidedplanexecutions,
		prometheus.CounterValue,
		c.sqlStatsPerfDataObject[0].SqlStatsGuidedplanexecutionsPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.sqlStatsMisguidedplanexecutions,
		prometheus.CounterValue,
		c.sqlStatsPerfDataObject[0].SqlStatsMisguidedplanexecutionsPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.sqlStatsSafeAutoParams,
		prometheus.CounterValue,
		c.sqlStatsPerfDataObject[0].SqlStatsSafeAutoParamsPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.sqlStatsSQLAttentionrate,
		prometheus.CounterValue,
		c.sqlStatsPerfDataObject[0].SqlStatsSQLAttentionrate,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.sqlStatsSQLCompilations,
		prometheus.CounterValue,
		c.sqlStatsPerfDataObject[0].SqlStatsSQLCompilationsPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.sqlStatsSQLReCompilations,
		prometheus.CounterValue,
		c.sqlStatsPerfDataObject[0].SqlStatsSQLReCompilationsPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.sqlStatsUnsafeAutoParams,
		prometheus.CounterValue,
		c.sqlStatsPerfDataObject[0].SqlStatsUnsafeAutoParamsPerSec,
		sqlInstance.name,
	)

	return nil
}

func (c *Collector) closeSQLStats() {
	for _, perfDataCollector := range c.sqlStatsPerfDataCollectors {
		perfDataCollector.Close()
	}
}

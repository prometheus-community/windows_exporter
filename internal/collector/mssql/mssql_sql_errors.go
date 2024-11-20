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
	errs := make([]error, 0, len(c.mssqlInstances))
	counters := []string{
		sqlErrorsErrorsPerSec,
	}

	for sqlInstance := range c.mssqlInstances {
		c.genStatsPerfDataCollectors[sqlInstance], err = perfdata.NewCollector(c.mssqlGetPerfObjectName(sqlInstance, "SQL Errors"), perfdata.InstancesAll, counters)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to create SQL Errors collector for instance %s: %w", sqlInstance, err))
		}
	}

	// Win32_PerfRawData_{instance}_SQLServerSQLErrors
	c.sqlErrorsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sql_errors_total"),
		"(SQLErrors.Total)",
		[]string{"mssql_instance", "resource"},
		nil,
	)

	return errors.Join(errs...)
}

func (c *Collector) collectSQLErrors(ch chan<- prometheus.Metric) error {
	return c.collect(ch, subCollectorSQLErrors, c.dbReplicaPerfDataCollectors, c.collectSQLErrorsInstance)
}

func (c *Collector) collectSQLErrorsInstance(ch chan<- prometheus.Metric, sqlInstance string, perfDataCollector *perfdata.Collector) error {
	if perfDataCollector == nil {
		return types.ErrPerfCounterCollectorNotInitialized
	}

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

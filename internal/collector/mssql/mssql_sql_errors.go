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

type collectorSQLErrors struct {
	sqlErrorsPerfDataCollectors map[mssqlInstance]*pdh.Collector
	sqlErrorsPerfDataObject     []perfDataCounterValuesSqlErrors

	// Win32_PerfRawData_{instance}_SQLServerSQLErrors
	sqlErrorsTotal *prometheus.Desc
}

type perfDataCounterValuesSqlErrors struct {
	Name string

	SqlErrorsErrorsPerSec float64 `perfdata:"Errors/sec"`
}

func (c *Collector) buildSQLErrors() error {
	var err error

	c.sqlErrorsPerfDataCollectors = make(map[mssqlInstance]*pdh.Collector, len(c.mssqlInstances))
	errs := make([]error, 0, len(c.mssqlInstances))

	for _, sqlInstance := range c.mssqlInstances {
		c.sqlErrorsPerfDataCollectors[sqlInstance], err = pdh.NewCollector[perfDataCounterValuesSqlErrors](pdh.CounterTypeRaw, c.mssqlGetPerfObjectName(sqlInstance, "SQL Errors"), pdh.InstancesAll)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to create SQL Errors collector for instance %s: %w", sqlInstance.name, err))
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
	return c.collect(ch, subCollectorSQLErrors, c.sqlErrorsPerfDataCollectors, c.collectSQLErrorsInstance)
}

func (c *Collector) collectSQLErrorsInstance(ch chan<- prometheus.Metric, sqlInstance mssqlInstance, perfDataCollector *pdh.Collector) error {
	err := perfDataCollector.Collect(&c.sqlErrorsPerfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect %s metrics: %w", c.mssqlGetPerfObjectName(sqlInstance, "SQL Errors"), err)
	}

	for _, data := range c.sqlErrorsPerfDataObject {
		ch <- prometheus.MustNewConstMetric(
			c.sqlErrorsTotal,
			prometheus.CounterValue,
			data.SqlErrorsErrorsPerSec,
			sqlInstance.name, data.Name,
		)
	}

	return nil
}

func (c *Collector) closeSQLErrors() {
	for _, perfDataCollector := range c.sqlErrorsPerfDataCollectors {
		perfDataCollector.Close()
	}
}

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

type collectorLocks struct {
	locksPerfDataCollectors map[mssqlInstance]*pdh.Collector
	locksPerfDataObject     []perfDataCounterValuesLocks

	// Win32_PerfRawData_{instance}_SQLServerLocks
	locksWaitTime             *prometheus.Desc
	locksCount                *prometheus.Desc
	locksLockRequests         *prometheus.Desc
	locksLockTimeouts         *prometheus.Desc
	locksLockTimeoutstimeout0 *prometheus.Desc
	locksLockWaits            *prometheus.Desc
	locksLockWaitTimeMS       *prometheus.Desc
	locksNumberOfDeadlocks    *prometheus.Desc
}

type perfDataCounterValuesLocks struct {
	Name string

	LocksAverageWaitTimeMS          float64 `perfdata:"Average Wait Time (ms)"`
	LocksAverageWaitTimeMSBase      float64 `perfdata:"Average Wait Time Base,secondvalue"`
	LocksLockRequestsPerSec         float64 `perfdata:"Lock Requests/sec"`
	LocksLockTimeoutsPerSec         float64 `perfdata:"Lock Timeouts/sec"`
	LocksLockTimeoutsTimeout0PerSec float64 `perfdata:"Lock Timeouts (timeout > 0)/sec"`
	LocksLockWaitsPerSec            float64 `perfdata:"Lock Waits/sec"`
	LocksLockWaitTimeMS             float64 `perfdata:"Lock Wait Time (ms)"`
	LocksNumberOfDeadlocksPerSec    float64 `perfdata:"Number of Deadlocks/sec"`
}

func (c *Collector) buildLocks() error {
	var err error

	c.locksPerfDataCollectors = make(map[mssqlInstance]*pdh.Collector, len(c.mssqlInstances))
	errs := make([]error, 0, len(c.mssqlInstances))

	for _, sqlInstance := range c.mssqlInstances {
		c.locksPerfDataCollectors[sqlInstance], err = pdh.NewCollector[perfDataCounterValuesLocks](pdh.CounterTypeRaw, c.mssqlGetPerfObjectName(sqlInstance, "Locks"), pdh.InstancesAll)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to create Locks collector for instance %s: %w", sqlInstance.name, err))
		}
	}

	c.locksWaitTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "locks_wait_time_seconds"),
		"(Locks.AverageWaitTimems Total time in seconds which locks have been holding resources)",
		[]string{"mssql_instance", "resource"},
		nil,
	)
	c.locksCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "locks_count"),
		"(Locks.AverageWaitTimems_Base count of how often requests have run into locks)",
		[]string{"mssql_instance", "resource"},
		nil,
	)
	c.locksLockRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "locks_lock_requests"),
		"(Locks.LockRequests)",
		[]string{"mssql_instance", "resource"},
		nil,
	)
	c.locksLockTimeouts = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "locks_lock_timeouts"),
		"(Locks.LockTimeouts)",
		[]string{"mssql_instance", "resource"},
		nil,
	)
	c.locksLockTimeoutstimeout0 = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "locks_lock_timeouts_excluding_NOWAIT"),
		"(Locks.LockTimeoutstimeout0)",
		[]string{"mssql_instance", "resource"},
		nil,
	)
	c.locksLockWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "locks_lock_waits"),
		"(Locks.LockWaits)",
		[]string{"mssql_instance", "resource"},
		nil,
	)
	c.locksLockWaitTimeMS = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "locks_lock_wait_seconds"),
		"(Locks.LockWaitTimems)",
		[]string{"mssql_instance", "resource"},
		nil,
	)
	c.locksNumberOfDeadlocks = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "locks_deadlocks"),
		"(Locks.NumberOfDeadlocks)",
		[]string{"mssql_instance", "resource"},
		nil,
	)

	return errors.Join(errs...)
}

func (c *Collector) collectLocks(ch chan<- prometheus.Metric) error {
	return c.collect(ch, subCollectorLocks, c.locksPerfDataCollectors, c.collectLocksInstance)
}

func (c *Collector) collectLocksInstance(ch chan<- prometheus.Metric, sqlInstance mssqlInstance, perfDataCollector *pdh.Collector) error {
	err := perfDataCollector.Collect(&c.locksPerfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect %s metrics: %w", c.mssqlGetPerfObjectName(sqlInstance, "Locks"), err)
	}

	for _, data := range c.locksPerfDataObject {
		ch <- prometheus.MustNewConstMetric(
			c.locksWaitTime,
			prometheus.GaugeValue,
			data.LocksAverageWaitTimeMS/1000.0,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.locksCount,
			prometheus.GaugeValue,
			data.LocksAverageWaitTimeMSBase/1000.0,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.locksLockRequests,
			prometheus.CounterValue,
			data.LocksLockRequestsPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.locksLockTimeouts,
			prometheus.CounterValue,
			data.LocksLockTimeoutsPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.locksLockTimeoutstimeout0,
			prometheus.CounterValue,
			data.LocksLockTimeoutsTimeout0PerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.locksLockWaits,
			prometheus.CounterValue,
			data.LocksLockWaitsPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.locksLockWaitTimeMS,
			prometheus.GaugeValue,
			data.LocksLockWaitTimeMS/1000.0,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.locksNumberOfDeadlocks,
			prometheus.CounterValue,
			data.LocksNumberOfDeadlocksPerSec,
			sqlInstance.name, data.Name,
		)
	}

	return nil
}

func (c *Collector) closeLocks() {
	for _, perfDataCollector := range c.locksPerfDataCollectors {
		perfDataCollector.Close()
	}
}

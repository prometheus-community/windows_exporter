//go:build windows

package mssql

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorLocks struct {
	locksPerfDataCollectors map[string]*perfdata.Collector

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

const (
	locksAverageWaitTimeMS          = "Average Wait Time (ms)"
	locksAverageWaitTimeMSBase      = "Average Wait Time Base"
	locksLockRequestsPerSec         = "Lock Requests/sec"
	locksLockTimeoutsPerSec         = "Lock Timeouts/sec"
	locksLockTimeoutsTimeout0PerSec = "Lock Timeouts (timeout > 0)/sec"
	locksLockWaitsPerSec            = "Lock Waits/sec"
	locksLockWaitTimeMS             = "Lock Wait Time (ms)"
	locksNumberOfDeadlocksPerSec    = "Number of Deadlocks/sec"
)

func (c *Collector) buildLocks() error {
	var err error

	c.locksPerfDataCollectors = make(map[string]*perfdata.Collector, len(c.mssqlInstances))
	counters := []string{
		locksAverageWaitTimeMS,
		locksAverageWaitTimeMSBase,
		locksLockRequestsPerSec,
		locksLockTimeoutsPerSec,
		locksLockTimeoutsTimeout0PerSec,
		locksLockWaitsPerSec,
		locksLockWaitTimeMS,
		locksNumberOfDeadlocksPerSec,
	}

	for sqlInstance := range c.mssqlInstances {
		c.locksPerfDataCollectors[sqlInstance], err = perfdata.NewCollector(c.mssqlGetPerfObjectName(sqlInstance, "Locks"), perfdata.InstanceAll, counters)
		if err != nil {
			return fmt.Errorf("failed to create Locks collector for instance %s: %w", sqlInstance, err)
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

	return nil
}

func (c *Collector) collectLocks(ch chan<- prometheus.Metric) error {
	return c.collect(ch, subCollectorLocks, c.locksPerfDataCollectors, c.collectLocksInstance)
}

func (c *Collector) collectLocksInstance(ch chan<- prometheus.Metric, sqlInstance string, perfDataCollector *perfdata.Collector) error {
	perfData, err := perfDataCollector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect %s metrics: %w", c.mssqlGetPerfObjectName(sqlInstance, "Locks"), err)
	}

	for lockResourceName, data := range perfData {
		ch <- prometheus.MustNewConstMetric(
			c.locksWaitTime,
			prometheus.GaugeValue,
			data[locksAverageWaitTimeMS].FirstValue/1000.0,
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.locksCount,
			prometheus.GaugeValue,
			data[locksAverageWaitTimeMSBase].SecondValue/1000.0,
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.locksLockRequests,
			prometheus.CounterValue,
			data[locksLockRequestsPerSec].FirstValue,
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.locksLockTimeouts,
			prometheus.CounterValue,
			data[locksLockTimeoutsPerSec].FirstValue,
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.locksLockTimeoutstimeout0,
			prometheus.CounterValue,
			data[locksLockTimeoutsTimeout0PerSec].FirstValue,
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.locksLockWaits,
			prometheus.CounterValue,
			data[locksLockWaitsPerSec].FirstValue,
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.locksLockWaitTimeMS,
			prometheus.GaugeValue,
			data[locksLockWaitTimeMS].FirstValue/1000.0,
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.locksNumberOfDeadlocks,
			prometheus.CounterValue,
			data[locksNumberOfDeadlocksPerSec].FirstValue,
			sqlInstance, lockResourceName,
		)
	}

	return nil
}

func (c *Collector) closeLocks() {
	for _, perfDataCollector := range c.locksPerfDataCollectors {
		perfDataCollector.Close()
	}
}

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

type collectorGeneralStatistics struct {
	genStatsPerfDataCollectors map[mssqlInstance]*pdh.Collector
	genStatsPerfDataObject     []perfDataCounterValuesGenStats

	genStatsActiveTempTables              *prometheus.Desc
	genStatsConnectionReset               *prometheus.Desc
	genStatsEventNotificationsDelayedDrop *prometheus.Desc
	genStatsHTTPAuthenticatedRequests     *prometheus.Desc
	genStatsLogicalConnections            *prometheus.Desc
	genStatsLogins                        *prometheus.Desc
	genStatsLogouts                       *prometheus.Desc
	genStatsMarsDeadlocks                 *prometheus.Desc
	genStatsNonAtomicYieldRate            *prometheus.Desc
	genStatsProcessesBlocked              *prometheus.Desc
	genStatsSOAPEmptyRequests             *prometheus.Desc
	genStatsSOAPMethodInvocations         *prometheus.Desc
	genStatsSOAPSessionInitiateRequests   *prometheus.Desc
	genStatsSOAPSessionTerminateRequests  *prometheus.Desc
	genStatsSOAPSQLRequests               *prometheus.Desc
	genStatsSOAPWSDLRequests              *prometheus.Desc
	genStatsSQLTraceIOProviderLockWaits   *prometheus.Desc
	genStatsTempDBRecoveryUnitID          *prometheus.Desc
	genStatsTempDBrowSetID                *prometheus.Desc
	genStatsTempTablesCreationRate        *prometheus.Desc
	genStatsTempTablesForDestruction      *prometheus.Desc
	genStatsTraceEventNotificationQueue   *prometheus.Desc
	genStatsTransactions                  *prometheus.Desc
	genStatsUserConnections               *prometheus.Desc
}

type perfDataCounterValuesGenStats struct {
	GenStatsActiveTempTables              float64 `perfdata:"Active Temp Tables"`
	GenStatsConnectionResetPerSec         float64 `perfdata:"Connection Reset/sec"`
	GenStatsEventNotificationsDelayedDrop float64 `perfdata:"Event Notifications Delayed Drop"`
	GenStatsHTTPAuthenticatedRequests     float64 `perfdata:"HTTP Authenticated Requests"`
	GenStatsLogicalConnections            float64 `perfdata:"Logical Connections"`
	GenStatsLoginsPerSec                  float64 `perfdata:"Logins/sec"`
	GenStatsLogoutsPerSec                 float64 `perfdata:"Logouts/sec"`
	GenStatsMarsDeadlocks                 float64 `perfdata:"Mars Deadlocks"`
	GenStatsNonatomicYieldRate            float64 `perfdata:"Non-atomic yield rate"`
	GenStatsProcessesBlocked              float64 `perfdata:"Processes blocked"`
	GenStatsSOAPEmptyRequests             float64 `perfdata:"SOAP Empty Requests"`
	GenStatsSOAPMethodInvocations         float64 `perfdata:"SOAP Method Invocations"`
	GenStatsSOAPSessionInitiateRequests   float64 `perfdata:"SOAP Session Initiate Requests"`
	GenStatsSOAPSessionTerminateRequests  float64 `perfdata:"SOAP Session Terminate Requests"`
	GenStatsSOAPSQLRequests               float64 `perfdata:"SOAP SQL Requests"`
	GenStatsSOAPWSDLRequests              float64 `perfdata:"SOAP WSDL Requests"`
	GenStatsSQLTraceIOProviderLockWaits   float64 `perfdata:"SQL Trace IO Provider Lock Waits"`
	GenStatsTempdbRecoveryUnitID          float64 `perfdata:"Tempdb recovery unit id"`
	GenStatsTempdbRowsetID                float64 `perfdata:"Tempdb rowset id"`
	GenStatsTempTablesCreationRate        float64 `perfdata:"Temp Tables Creation Rate"`
	GenStatsTempTablesForDestruction      float64 `perfdata:"Temp Tables For Destruction"`
	GenStatsTraceEventNotificationQueue   float64 `perfdata:"Trace Event Notification Queue"`
	GenStatsTransactions                  float64 `perfdata:"Transactions"`
	GenStatsUserConnections               float64 `perfdata:"User Connections"`
}

func (c *Collector) buildGeneralStatistics() error {
	var err error

	c.genStatsPerfDataCollectors = make(map[mssqlInstance]*pdh.Collector, len(c.mssqlInstances))
	errs := make([]error, 0, len(c.mssqlInstances))

	for _, sqlInstance := range c.mssqlInstances {
		c.genStatsPerfDataCollectors[sqlInstance], err = pdh.NewCollector[perfDataCounterValuesGenStats](pdh.CounterTypeRaw, c.mssqlGetPerfObjectName(sqlInstance, "General Statistics"), nil)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to create General Statistics collector for instance %s: %w", sqlInstance.name, err))
		}
	}

	// Win32_PerfRawData_{instance}_SQLServerGeneralStatistics
	c.genStatsActiveTempTables = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_active_temp_tables"),
		"(GeneralStatistics.ActiveTempTables)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsConnectionReset = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_connection_resets"),
		"(GeneralStatistics.ConnectionReset)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsEventNotificationsDelayedDrop = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_event_notifications_delayed_drop"),
		"(GeneralStatistics.EventNotificationsDelayedDrop)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsHTTPAuthenticatedRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_http_authenticated_requests"),
		"(GeneralStatistics.HTTPAuthenticatedRequests)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsLogicalConnections = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_logical_connections"),
		"(GeneralStatistics.LogicalConnections)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsLogins = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_logins"),
		"(GeneralStatistics.Logins)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsLogouts = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_logouts"),
		"(GeneralStatistics.Logouts)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsMarsDeadlocks = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_mars_deadlocks"),
		"(GeneralStatistics.MarsDeadlocks)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsNonAtomicYieldRate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_non_atomic_yields"),
		"(GeneralStatistics.Nonatomicyields)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsProcessesBlocked = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_blocked_processes"),
		"(GeneralStatistics.Processesblocked)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsSOAPEmptyRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_soap_empty_requests"),
		"(GeneralStatistics.SOAPEmptyRequests)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsSOAPMethodInvocations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_soap_method_invocations"),
		"(GeneralStatistics.SOAPMethodInvocations)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsSOAPSessionInitiateRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_soap_session_initiate_requests"),
		"(GeneralStatistics.SOAPSessionInitiateRequests)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsSOAPSessionTerminateRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_soap_session_terminate_requests"),
		"(GeneralStatistics.SOAPSessionTerminateRequests)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsSOAPSQLRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_soapsql_requests"),
		"(GeneralStatistics.SOAPSQLRequests)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsSOAPWSDLRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_soapwsdl_requests"),
		"(GeneralStatistics.SOAPWSDLRequests)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsSQLTraceIOProviderLockWaits = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_sql_trace_io_provider_lock_waits"),
		"(GeneralStatistics.SQLTraceIOProviderLockWaits)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsTempDBRecoveryUnitID = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_tempdb_recovery_unit_ids_generated"),
		"(GeneralStatistics.Tempdbrecoveryunitid)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsTempDBrowSetID = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_tempdb_rowset_ids_generated"),
		"(GeneralStatistics.Tempdbrowsetid)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsTempTablesCreationRate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_temp_tables_creations"),
		"(GeneralStatistics.TempTablesCreations)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsTempTablesForDestruction = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_temp_tables_awaiting_destruction"),
		"(GeneralStatistics.TempTablesForDestruction)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsTraceEventNotificationQueue = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_trace_event_notification_queue_size"),
		"(GeneralStatistics.TraceEventNotificationQueue)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsTransactions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_transactions"),
		"(GeneralStatistics.Transactions)",
		[]string{"mssql_instance"},
		nil,
	)
	c.genStatsUserConnections = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "genstats_user_connections"),
		"(GeneralStatistics.UserConnections)",
		[]string{"mssql_instance"},
		nil,
	)

	return errors.Join(errs...)
}

func (c *Collector) collectGeneralStatistics(ch chan<- prometheus.Metric) error {
	return c.collect(ch, subCollectorGeneralStatistics, c.genStatsPerfDataCollectors, c.collectGeneralStatisticsInstance)
}

func (c *Collector) collectGeneralStatisticsInstance(ch chan<- prometheus.Metric, sqlInstance mssqlInstance, perfDataCollector *pdh.Collector) error {
	err := perfDataCollector.Collect(&c.genStatsPerfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect %s metrics: %w", c.mssqlGetPerfObjectName(sqlInstance, "General Statistics"), err)
	}

	ch <- prometheus.MustNewConstMetric(
		c.genStatsActiveTempTables,
		prometheus.GaugeValue,
		c.genStatsPerfDataObject[0].GenStatsActiveTempTables,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsConnectionReset,
		prometheus.CounterValue,
		c.genStatsPerfDataObject[0].GenStatsConnectionResetPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsEventNotificationsDelayedDrop,
		prometheus.GaugeValue,
		c.genStatsPerfDataObject[0].GenStatsEventNotificationsDelayedDrop,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsHTTPAuthenticatedRequests,
		prometheus.GaugeValue,
		c.genStatsPerfDataObject[0].GenStatsHTTPAuthenticatedRequests,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsLogicalConnections,
		prometheus.GaugeValue,
		c.genStatsPerfDataObject[0].GenStatsLogicalConnections,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsLogins,
		prometheus.CounterValue,
		c.genStatsPerfDataObject[0].GenStatsLoginsPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsLogouts,
		prometheus.CounterValue,
		c.genStatsPerfDataObject[0].GenStatsLogoutsPerSec,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsMarsDeadlocks,
		prometheus.GaugeValue,
		c.genStatsPerfDataObject[0].GenStatsMarsDeadlocks,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsNonAtomicYieldRate,
		prometheus.CounterValue,
		c.genStatsPerfDataObject[0].GenStatsNonatomicYieldRate,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsProcessesBlocked,
		prometheus.GaugeValue,
		c.genStatsPerfDataObject[0].GenStatsProcessesBlocked,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsSOAPEmptyRequests,
		prometheus.GaugeValue,
		c.genStatsPerfDataObject[0].GenStatsSOAPEmptyRequests,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsSOAPMethodInvocations,
		prometheus.GaugeValue,
		c.genStatsPerfDataObject[0].GenStatsSOAPMethodInvocations,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsSOAPSessionInitiateRequests,
		prometheus.GaugeValue,
		c.genStatsPerfDataObject[0].GenStatsSOAPSessionInitiateRequests,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsSOAPSessionTerminateRequests,
		prometheus.GaugeValue,
		c.genStatsPerfDataObject[0].GenStatsSOAPSessionTerminateRequests,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsSOAPSQLRequests,
		prometheus.GaugeValue,
		c.genStatsPerfDataObject[0].GenStatsSOAPSQLRequests,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsSOAPWSDLRequests,
		prometheus.GaugeValue,
		c.genStatsPerfDataObject[0].GenStatsSOAPWSDLRequests,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsSQLTraceIOProviderLockWaits,
		prometheus.GaugeValue,
		c.genStatsPerfDataObject[0].GenStatsSQLTraceIOProviderLockWaits,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsTempDBRecoveryUnitID,
		prometheus.GaugeValue,
		c.genStatsPerfDataObject[0].GenStatsTempdbRecoveryUnitID,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsTempDBrowSetID,
		prometheus.GaugeValue,
		c.genStatsPerfDataObject[0].GenStatsTempdbRowsetID,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsTempTablesCreationRate,
		prometheus.CounterValue,
		c.genStatsPerfDataObject[0].GenStatsTempTablesCreationRate,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsTempTablesForDestruction,
		prometheus.GaugeValue,
		c.genStatsPerfDataObject[0].GenStatsTempTablesForDestruction,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsTraceEventNotificationQueue,
		prometheus.GaugeValue,
		c.genStatsPerfDataObject[0].GenStatsTraceEventNotificationQueue,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsTransactions,
		prometheus.GaugeValue,
		c.genStatsPerfDataObject[0].GenStatsTransactions,
		sqlInstance.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsUserConnections,
		prometheus.GaugeValue,
		c.genStatsPerfDataObject[0].GenStatsUserConnections,
		sqlInstance.name,
	)

	return nil
}

func (c *Collector) closeGeneralStatistics() {
	for _, perfDataCollector := range c.genStatsPerfDataCollectors {
		perfDataCollector.Close()
	}
}

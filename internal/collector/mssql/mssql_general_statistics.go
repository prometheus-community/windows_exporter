//go:build windows

package mssql

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorGeneralStatistics struct {
	genStatsPerfDataCollectors map[string]*perfdata.Collector

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

const (
	genStatsActiveTempTables              = "Active Temp Tables"
	genStatsConnectionResetPerSec         = "Connection Reset/sec"
	genStatsEventNotificationsDelayedDrop = "Event Notifications Delayed Drop"
	genStatsHTTPAuthenticatedRequests     = "HTTP Authenticated Requests"
	genStatsLogicalConnections            = "Logical Connections"
	genStatsLoginsPerSec                  = "Logins/sec"
	genStatsLogoutsPerSec                 = "Logouts/sec"
	genStatsMarsDeadlocks                 = "Mars Deadlocks"
	genStatsNonatomicYieldRate            = "Non-atomic yield rate"
	genStatsProcessesBlocked              = "Processes blocked"
	genStatsSOAPEmptyRequests             = "SOAP Empty Requests"
	genStatsSOAPMethodInvocations         = "SOAP Method Invocations"
	genStatsSOAPSessionInitiateRequests   = "SOAP Session Initiate Requests"
	genStatsSOAPSessionTerminateRequests  = "SOAP Session Terminate Requests"
	genStatsSOAPSQLRequests               = "SOAP SQL Requests"
	genStatsSOAPWSDLRequests              = "SOAP WSDL Requests"
	genStatsSQLTraceIOProviderLockWaits   = "SQL Trace IO Provider Lock Waits"
	genStatsTempdbRecoveryUnitID          = "Tempdb recovery unit id"
	genStatsTempdbRowsetID                = "Tempdb rowset id"
	genStatsTempTablesCreationRate        = "Temp Tables Creation Rate"
	genStatsTempTablesForDestruction      = "Temp Tables For Destruction"
	genStatsTraceEventNotificationQueue   = "Trace Event Notification Queue"
	genStatsTransactions                  = "Transactions"
	genStatsUserConnections               = "User Connections"
)

func (c *Collector) buildGeneralStatistics() error {
	var err error

	c.genStatsPerfDataCollectors = make(map[string]*perfdata.Collector, len(c.mssqlInstances))
	counters := []string{
		genStatsActiveTempTables,
		genStatsConnectionResetPerSec,
		genStatsEventNotificationsDelayedDrop,
		genStatsHTTPAuthenticatedRequests,
		genStatsLogicalConnections,
		genStatsLoginsPerSec,
		genStatsLogoutsPerSec,
		genStatsMarsDeadlocks,
		genStatsNonatomicYieldRate,
		genStatsProcessesBlocked,
		genStatsSOAPEmptyRequests,
		genStatsSOAPMethodInvocations,
		genStatsSOAPSessionInitiateRequests,
		genStatsSOAPSessionTerminateRequests,
		genStatsSOAPSQLRequests,
		genStatsSOAPWSDLRequests,
		genStatsSQLTraceIOProviderLockWaits,
		genStatsTempdbRecoveryUnitID,
		genStatsTempdbRowsetID,
		genStatsTempTablesCreationRate,
		genStatsTempTablesForDestruction,
		genStatsTraceEventNotificationQueue,
		genStatsTransactions,
		genStatsUserConnections,
	}

	for sqlInstance := range c.mssqlInstances {
		c.genStatsPerfDataCollectors[sqlInstance], err = perfdata.NewCollector(c.mssqlGetPerfObjectName(sqlInstance, "General Statistics"), nil, counters)
		if err != nil {
			return fmt.Errorf("failed to create General Statistics collector for instance %s: %w", sqlInstance, err)
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

	return nil
}

func (c *Collector) collectGeneralStatistics(ch chan<- prometheus.Metric) error {
	return c.collect(ch, subCollectorGeneralStatistics, c.genStatsPerfDataCollectors, c.collectGeneralStatisticsInstance)
}

func (c *Collector) collectGeneralStatisticsInstance(ch chan<- prometheus.Metric, sqlInstance string, perfDataCollector *perfdata.Collector) error {
	perfData, err := perfDataCollector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect %s metrics: %w", c.mssqlGetPerfObjectName(sqlInstance, "General Statistics"), err)
	}

	data, ok := perfData[perfdata.EmptyInstance]
	if !ok {
		return fmt.Errorf("perflib query for %s returned empty result set", c.mssqlGetPerfObjectName(sqlInstance, "General Statistics"))
	}

	ch <- prometheus.MustNewConstMetric(
		c.genStatsActiveTempTables,
		prometheus.GaugeValue,
		data[genStatsActiveTempTables].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsConnectionReset,
		prometheus.CounterValue,
		data[genStatsConnectionResetPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsEventNotificationsDelayedDrop,
		prometheus.GaugeValue,
		data[genStatsEventNotificationsDelayedDrop].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsHTTPAuthenticatedRequests,
		prometheus.GaugeValue,
		data[genStatsHTTPAuthenticatedRequests].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsLogicalConnections,
		prometheus.GaugeValue,
		data[genStatsLogicalConnections].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsLogins,
		prometheus.CounterValue,
		data[genStatsLoginsPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsLogouts,
		prometheus.CounterValue,
		data[genStatsLogoutsPerSec].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsMarsDeadlocks,
		prometheus.GaugeValue,
		data[genStatsMarsDeadlocks].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsNonAtomicYieldRate,
		prometheus.CounterValue,
		data[genStatsNonatomicYieldRate].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsProcessesBlocked,
		prometheus.GaugeValue,
		data[genStatsProcessesBlocked].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsSOAPEmptyRequests,
		prometheus.GaugeValue,
		data[genStatsSOAPEmptyRequests].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsSOAPMethodInvocations,
		prometheus.GaugeValue,
		data[genStatsSOAPMethodInvocations].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsSOAPSessionInitiateRequests,
		prometheus.GaugeValue,
		data[genStatsSOAPSessionInitiateRequests].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsSOAPSessionTerminateRequests,
		prometheus.GaugeValue,
		data[genStatsSOAPSessionTerminateRequests].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsSOAPSQLRequests,
		prometheus.GaugeValue,
		data[genStatsSOAPSQLRequests].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsSOAPWSDLRequests,
		prometheus.GaugeValue,
		data[genStatsSOAPWSDLRequests].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsSQLTraceIOProviderLockWaits,
		prometheus.GaugeValue,
		data[genStatsSQLTraceIOProviderLockWaits].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsTempDBRecoveryUnitID,
		prometheus.GaugeValue,
		data[genStatsTempdbRecoveryUnitID].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsTempDBrowSetID,
		prometheus.GaugeValue,
		data[genStatsTempdbRowsetID].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsTempTablesCreationRate,
		prometheus.CounterValue,
		data[genStatsTempTablesCreationRate].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsTempTablesForDestruction,
		prometheus.GaugeValue,
		data[genStatsTempTablesForDestruction].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsTraceEventNotificationQueue,
		prometheus.GaugeValue,
		data[genStatsTraceEventNotificationQueue].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsTransactions,
		prometheus.GaugeValue,
		data[genStatsTransactions].FirstValue,
		sqlInstance,
	)

	ch <- prometheus.MustNewConstMetric(
		c.genStatsUserConnections,
		prometheus.GaugeValue,
		data[genStatsUserConnections].FirstValue,
		sqlInstance,
	)

	return nil
}

func (c *Collector) closeGeneralStatistics() {
	for _, perfDataCollector := range c.genStatsPerfDataCollectors {
		perfDataCollector.Close()
	}
}

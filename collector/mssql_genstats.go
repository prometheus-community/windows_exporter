// returns data points from the following classes:
// - Win32_PerfRawData_MSSQLSERVER_SQLServerGeneralStatistics
//   https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-general-statistics-object

package collector

import (
	"fmt"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	Factories["mssql_genstats"] = NewMSSQLGenStatsCollector
}

// MSSQLGenStatsCollector is a Prometheus collector for Win32_PerfRawData_{instance}_SQLServerGeneralStatistics metrics
type MSSQLGenStatsCollector struct {
	ActiveTempTables              *prometheus.Desc
	ConnectionResetPersec         *prometheus.Desc
	EventNotificationsDelayedDrop *prometheus.Desc
	HTTPAuthenticatedRequests     *prometheus.Desc
	LogicalConnections            *prometheus.Desc
	LoginsPersec                  *prometheus.Desc
	LogoutsPersec                 *prometheus.Desc
	MarsDeadlocks                 *prometheus.Desc
	Nonatomicyieldrate            *prometheus.Desc
	Processesblocked              *prometheus.Desc
	SOAPEmptyRequests             *prometheus.Desc
	SOAPMethodInvocations         *prometheus.Desc
	SOAPSessionInitiateRequests   *prometheus.Desc
	SOAPSessionTerminateRequests  *prometheus.Desc
	SOAPSQLRequests               *prometheus.Desc
	SOAPWSDLRequests              *prometheus.Desc
	SQLTraceIOProviderLockWaits   *prometheus.Desc
	Tempdbrecoveryunitid          *prometheus.Desc
	Tempdbrowsetid                *prometheus.Desc
	TempTablesCreationRate        *prometheus.Desc
	TempTablesForDestruction      *prometheus.Desc
	TraceEventNotificationQueue   *prometheus.Desc
	Transactions                  *prometheus.Desc
	UserConnections               *prometheus.Desc

	sqlInstances sqlInstancesType
}

// NewMSSQLGenStatsCollector ...
func NewMSSQLGenStatsCollector() (Collector, error) {

	const subsystem = "mssql_genstats"
	return &MSSQLGenStatsCollector{

		// Win32_PerfRawData_{instance}_SQLServerGeneralStatistics
		ActiveTempTables: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "active_temp_tables"),
			"(GeneralStatistics.ActiveTempTables)",
			[]string{"instance"},
			nil,
		),
		ConnectionResetPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "connection_resets"),
			"(GeneralStatistics.ConnectionReset)",
			[]string{"instance"},
			nil,
		),
		EventNotificationsDelayedDrop: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "event_notifications_delayed_drop"),
			"(GeneralStatistics.EventNotificationsDelayedDrop)",
			[]string{"instance"},
			nil,
		),
		HTTPAuthenticatedRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "http_authenticated_requests"),
			"(GeneralStatistics.HTTPAuthenticatedRequests)",
			[]string{"instance"},
			nil,
		),
		LogicalConnections: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "logical_connections"),
			"(GeneralStatistics.LogicalConnections)",
			[]string{"instance"},
			nil,
		),
		LoginsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "logins"),
			"(GeneralStatistics.Logins)",
			[]string{"instance"},
			nil,
		),
		LogoutsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "logouts"),
			"(GeneralStatistics.Logouts)",
			[]string{"instance"},
			nil,
		),
		MarsDeadlocks: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "mars_deadlocks"),
			"(GeneralStatistics.MarsDeadlocks)",
			[]string{"instance"},
			nil,
		),
		Nonatomicyieldrate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "non_atomic_yields"),
			"(GeneralStatistics.Nonatomicyields)",
			[]string{"instance"},
			nil,
		),
		Processesblocked: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "genstatss_blocked_processes"),
			"(GeneralStatistics.Processesblocked)",
			[]string{"instance"},
			nil,
		),
		SOAPEmptyRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "soap_empty_requests"),
			"(GeneralStatistics.SOAPEmptyRequests)",
			[]string{"instance"},
			nil,
		),
		SOAPMethodInvocations: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "soap_method_invocations"),
			"(GeneralStatistics.SOAPMethodInvocations)",
			[]string{"instance"},
			nil,
		),
		SOAPSessionInitiateRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "soap_session_initiate_requests"),
			"(GeneralStatistics.SOAPSessionInitiateRequests)",
			[]string{"instance"},
			nil,
		),
		SOAPSessionTerminateRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "soap_session_terminate_requests"),
			"(GeneralStatistics.SOAPSessionTerminateRequests)",
			[]string{"instance"},
			nil,
		),
		SOAPSQLRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "soapsql_requests"),
			"(GeneralStatistics.SOAPSQLRequests)",
			[]string{"instance"},
			nil,
		),
		SOAPWSDLRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "soapwsdl_requests"),
			"(GeneralStatistics.SOAPWSDLRequests)",
			[]string{"instance"},
			nil,
		),
		SQLTraceIOProviderLockWaits: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sql_trace_io_provider_lock_waits"),
			"(GeneralStatistics.SQLTraceIOProviderLockWaits)",
			[]string{"instance"},
			nil,
		),
		Tempdbrecoveryunitid: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "tempdb_recovery_unit_ids_generated"),
			"(GeneralStatistics.Tempdbrecoveryunitid)",
			[]string{"instance"},
			nil,
		),
		Tempdbrowsetid: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "tempdb_rowset_ids_generated"),
			"(GeneralStatistics.Tempdbrowsetid)",
			[]string{"instance"},
			nil,
		),
		TempTablesCreationRate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "temp_tables_creations"),
			"(GeneralStatistics.TempTablesCreations)",
			[]string{"instance"},
			nil,
		),
		TempTablesForDestruction: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "temp_tables_awaiting_destruction"),
			"(GeneralStatistics.TempTablesForDestruction)",
			[]string{"instance"},
			nil,
		),
		TraceEventNotificationQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "trace_event_notification_queue_size"),
			"(GeneralStatistics.TraceEventNotificationQueue)",
			[]string{"instance"},
			nil,
		),
		Transactions: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "transactions"),
			"(GeneralStatistics.Transactions)",
			[]string{"instance"},
			nil,
		),
		UserConnections: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "user_connections"),
			"(GeneralStatistics.UserConnections)",
			[]string{"instance"},
			nil,
		),

		sqlInstances: getMSSQLInstances(),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *MSSQLGenStatsCollector) Collect(ch chan<- prometheus.Metric) error {
	for instance := range c.sqlInstances {
		log.Debugf("mssql collector iterating sql instance %s.", instance)

		// Win32_PerfRawData_{instance}_SQLServerGeneralStatistics
		if desc, err := c.collectGeneralStatistics(ch, instance); err != nil {
			log.Error("failed collecting MSSQL GeneralStatistics metrics:", desc, instance, err)
			return err
		}
	}

	return nil
}

type win32PerfRawDataSQLServerGeneralStatistics struct {
	ActiveTempTables              uint64
	ConnectionResetPersec         uint64
	EventNotificationsDelayedDrop uint64
	HTTPAuthenticatedRequests     uint64
	LogicalConnections            uint64
	LoginsPersec                  uint64
	LogoutsPersec                 uint64
	MarsDeadlocks                 uint64
	Nonatomicyieldrate            uint64
	Processesblocked              uint64
	SOAPEmptyRequests             uint64
	SOAPMethodInvocations         uint64
	SOAPSessionInitiateRequests   uint64
	SOAPSessionTerminateRequests  uint64
	SOAPSQLRequests               uint64
	SOAPWSDLRequests              uint64
	SQLTraceIOProviderLockWaits   uint64
	Tempdbrecoveryunitid          uint64
	Tempdbrowsetid                uint64
	TempTablesCreationRate        uint64
	TempTablesForDestruction      uint64
	TraceEventNotificationQueue   uint64
	Transactions                  uint64
	UserConnections               uint64
}

func (c *MSSQLGenStatsCollector) collectGeneralStatistics(ch chan<- prometheus.Metric, sqlInstance string) (*prometheus.Desc, error) {
	var dst []win32PerfRawDataSQLServerGeneralStatistics
	class := fmt.Sprintf("Win32_PerfRawData_%s_SQLServerGeneralStatistics", sqlInstance)
	q := queryAllForClass(&dst, class)

	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	if len(dst) > 0 {
		v := dst[0]
		ch <- prometheus.MustNewConstMetric(
			c.ActiveTempTables,
			prometheus.GaugeValue,
			float64(v.ActiveTempTables),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ConnectionResetPersec,
			prometheus.CounterValue,
			float64(v.ConnectionResetPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.EventNotificationsDelayedDrop,
			prometheus.GaugeValue,
			float64(v.EventNotificationsDelayedDrop),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.HTTPAuthenticatedRequests,
			prometheus.GaugeValue,
			float64(v.HTTPAuthenticatedRequests),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogicalConnections,
			prometheus.GaugeValue,
			float64(v.LogicalConnections),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LoginsPersec,
			prometheus.CounterValue,
			float64(v.LoginsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogoutsPersec,
			prometheus.CounterValue,
			float64(v.LogoutsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MarsDeadlocks,
			prometheus.GaugeValue,
			float64(v.MarsDeadlocks),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Nonatomicyieldrate,
			prometheus.CounterValue,
			float64(v.Nonatomicyieldrate),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Processesblocked,
			prometheus.GaugeValue,
			float64(v.Processesblocked),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SOAPEmptyRequests,
			prometheus.GaugeValue,
			float64(v.SOAPEmptyRequests),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SOAPMethodInvocations,
			prometheus.GaugeValue,
			float64(v.SOAPMethodInvocations),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SOAPSessionInitiateRequests,
			prometheus.GaugeValue,
			float64(v.SOAPSessionInitiateRequests),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SOAPSessionTerminateRequests,
			prometheus.GaugeValue,
			float64(v.SOAPSessionTerminateRequests),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SOAPSQLRequests,
			prometheus.GaugeValue,
			float64(v.SOAPSQLRequests),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SOAPWSDLRequests,
			prometheus.GaugeValue,
			float64(v.SOAPWSDLRequests),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SQLTraceIOProviderLockWaits,
			prometheus.GaugeValue,
			float64(v.SQLTraceIOProviderLockWaits),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Tempdbrecoveryunitid,
			prometheus.GaugeValue,
			float64(v.Tempdbrecoveryunitid),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Tempdbrowsetid,
			prometheus.GaugeValue,
			float64(v.Tempdbrowsetid),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TempTablesCreationRate,
			prometheus.CounterValue,
			float64(v.TempTablesCreationRate),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TempTablesForDestruction,
			prometheus.GaugeValue,
			float64(v.TempTablesForDestruction),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TraceEventNotificationQueue,
			prometheus.GaugeValue,
			float64(v.TraceEventNotificationQueue),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Transactions,
			prometheus.GaugeValue,
			float64(v.Transactions),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.UserConnections,
			prometheus.GaugeValue,
			float64(v.UserConnections),
			sqlInstance,
		)
	}

	return nil, nil
}

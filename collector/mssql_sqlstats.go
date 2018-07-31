// returns data points from the following classes:
// - Win32_PerfRawData_MSSQLSERVER_SQLServerSQLStatistics
//   https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-sql-statistics-object

package collector

import (
	"fmt"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	Factories["mssql_sqlstats"] = NewMSSQLStatsCollector
}

// MSSQLStatsCollector is a Prometheus collector for Win32_PerfRawData_{instance}_SQLServerSQLStatistics metrics
type MSSQLStatsCollector struct {
	AutoParamAttemptsPersec       *prometheus.Desc
	BatchRequestsPersec           *prometheus.Desc
	FailedAutoParamsPersec        *prometheus.Desc
	ForcedParameterizationsPersec *prometheus.Desc
	GuidedplanexecutionsPersec    *prometheus.Desc
	MisguidedplanexecutionsPersec *prometheus.Desc
	SafeAutoParamsPersec          *prometheus.Desc
	SQLAttentionrate              *prometheus.Desc
	SQLCompilationsPersec         *prometheus.Desc
	SQLReCompilationsPersec       *prometheus.Desc
	UnsafeAutoParamsPersec        *prometheus.Desc

	sqlInstances sqlInstancesType
}

// NewMSSQLStatsCollector ...
func NewMSSQLStatsCollector() (Collector, error) {

	const subsystem = "mssql_stats"
	return &MSSQLStatsCollector{

		// Win32_PerfRawData_{instance}_SQLServerSQLStatistics
		AutoParamAttemptsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "auto_parameterization_attempts"),
			"(SQLStatistics.AutoParamAttempts)",
			[]string{"instance"},
			nil,
		),
		BatchRequestsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "batch_requests"),
			"(SQLStatistics.BatchRequests)",
			[]string{"instance"},
			nil,
		),
		FailedAutoParamsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "failed_auto_parameterization_attempts"),
			"(SQLStatistics.FailedAutoParams)",
			[]string{"instance"},
			nil,
		),
		ForcedParameterizationsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "forced_parameterizations"),
			"(SQLStatistics.ForcedParameterizations)",
			[]string{"instance"},
			nil,
		),
		GuidedplanexecutionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "guided_plan_executions"),
			"(SQLStatistics.Guidedplanexecutions)",
			[]string{"instance"},
			nil,
		),
		MisguidedplanexecutionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "misguided_plan_executions"),
			"(SQLStatistics.Misguidedplanexecutions)",
			[]string{"instance"},
			nil,
		),
		SafeAutoParamsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "safe_auto_parameterization_attempts"),
			"(SQLStatistics.SafeAutoParams)",
			[]string{"instance"},
			nil,
		),
		SQLAttentionrate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sql_attentions"),
			"(SQLStatistics.SQLAttentions)",
			[]string{"instance"},
			nil,
		),
		SQLCompilationsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sql_compilations"),
			"(SQLStatistics.SQLCompilations)",
			[]string{"instance"},
			nil,
		),
		SQLReCompilationsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sql_recompilations"),
			"(SQLStatistics.SQLReCompilations)",
			[]string{"instance"},
			nil,
		),
		UnsafeAutoParamsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "unsafe_auto_parameterization_attempts"),
			"(SQLStatistics.UnsafeAutoParams)",
			[]string{"instance"},
			nil,
		),

		sqlInstances: getMSSQLInstances(),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *MSSQLStatsCollector) Collect(ch chan<- prometheus.Metric) error {
	for instance := range c.sqlInstances {
		log.Debugf("mssql collector iterating sql instance %s.", instance)

		// Win32_PerfRawData_{instance}_SQLServerSQLStatistics
		if desc, err := c.collectSQLStats(ch, instance); err != nil {
			log.Error("failed collecting MSSQL SQLStats metrics:", desc, err)
			return err
		}
	}

	return nil
}

type win32PerfRawDataSQLServerSQLStatistics struct {
	AutoParamAttemptsPersec       uint64
	BatchRequestsPersec           uint64
	FailedAutoParamsPersec        uint64
	ForcedParameterizationsPersec uint64
	GuidedplanexecutionsPersec    uint64
	MisguidedplanexecutionsPersec uint64
	SafeAutoParamsPersec          uint64
	SQLAttentionrate              uint64
	SQLCompilationsPersec         uint64
	SQLReCompilationsPersec       uint64
	UnsafeAutoParamsPersec        uint64
}

func (c *MSSQLStatsCollector) collectSQLStats(ch chan<- prometheus.Metric, sqlInstance string) (*prometheus.Desc, error) {
	var dst []win32PerfRawDataSQLServerSQLStatistics
	class := fmt.Sprintf("Win32_PerfRawData_%s_SQLServerSQLStatistics", sqlInstance)
	q := queryAllForClass(&dst, class)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	if len(dst) > 0 {
		v := dst[0]

		ch <- prometheus.MustNewConstMetric(
			c.AutoParamAttemptsPersec,
			prometheus.CounterValue,
			float64(v.AutoParamAttemptsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BatchRequestsPersec,
			prometheus.CounterValue,
			float64(v.BatchRequestsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FailedAutoParamsPersec,
			prometheus.CounterValue,
			float64(v.FailedAutoParamsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ForcedParameterizationsPersec,
			prometheus.CounterValue,
			float64(v.ForcedParameterizationsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GuidedplanexecutionsPersec,
			prometheus.CounterValue,
			float64(v.GuidedplanexecutionsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MisguidedplanexecutionsPersec,
			prometheus.CounterValue,
			float64(v.MisguidedplanexecutionsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SafeAutoParamsPersec,
			prometheus.CounterValue,
			float64(v.SafeAutoParamsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SQLAttentionrate,
			prometheus.CounterValue,
			float64(v.SQLAttentionrate),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SQLCompilationsPersec,
			prometheus.CounterValue,
			float64(v.SQLCompilationsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SQLReCompilationsPersec,
			prometheus.CounterValue,
			float64(v.SQLReCompilationsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.UnsafeAutoParamsPersec,
			prometheus.CounterValue,
			float64(v.UnsafeAutoParamsPersec),
			sqlInstance,
		)
	}

	return nil, nil
}

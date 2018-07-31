// returns data points from the following classes:
// - Win32_PerfRawData_MSSQLSERVER_SQLServerLocks
//   https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-locks-object

package collector

import (
	"fmt"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	Factories["mssql_locks"] = NewMSSQLLocksCollector
}

// MSSQLLocksCollector is a Prometheus collector for Win32_PerfRawData_{instance}_SQLServerLocks metrics
type MSSQLLocksCollector struct {
	AverageWaitTimems          *prometheus.Desc
	LockRequestsPersec         *prometheus.Desc
	LockTimeoutsPersec         *prometheus.Desc
	LockTimeoutstimeout0Persec *prometheus.Desc
	LockWaitsPersec            *prometheus.Desc
	LockWaitTimems             *prometheus.Desc
	NumberofDeadlocksPersec    *prometheus.Desc

	sqlInstances sqlInstancesType
}

// NewMSSQLLocksCollector ...
func NewMSSQLLocksCollector() (Collector, error) {

	const subsystem = "mssql_locks"
	return &MSSQLLocksCollector{

		// Win32_PerfRawData_{instance}_SQLServerLocks
		AverageWaitTimems: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "average_wait_seconds"),
			"(Locks.AverageWaitTimems)",
			[]string{"instance", "resource"},
			nil,
		),
		LockRequestsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_requests"),
			"(Locks.LockRequests)",
			[]string{"instance", "resource"},
			nil,
		),
		LockTimeoutsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_timeouts"),
			"(Locks.LockTimeouts)",
			[]string{"instance", "resource"},
			nil,
		),
		LockTimeoutstimeout0Persec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_timeouts_excluding_NOWAIT"),
			"(Locks.LockTimeoutstimeout0)",
			[]string{"instance", "resource"},
			nil,
		),
		LockWaitsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_waits"),
			"(Locks.LockWaits)",
			[]string{"instance", "resource"},
			nil,
		),
		LockWaitTimems: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_wait_seconds"),
			"(Locks.LockWaitTimems)",
			[]string{"instance", "resource"},
			nil,
		),
		NumberofDeadlocksPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "deadlocks"),
			"(Locks.NumberofDeadlocks)",
			[]string{"instance", "resource"},
			nil,
		),

		sqlInstances: getMSSQLInstances(),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *MSSQLLocksCollector) Collect(ch chan<- prometheus.Metric) error {
	for instance := range c.sqlInstances {
		log.Debugf("mssql collector iterating sql instance %s.", instance)

		// Win32_PerfRawData_{instance}_SQLServerLocks
		if desc, err := c.collectLocks(ch, instance); err != nil {
			log.Error("failed collecting MSSQL Locks metrics:", desc, err)
			return err
		}
	}

	return nil
}

type win32PerfRawDataSQLServerLocks struct {
	Name                       string
	AverageWaitTimems          uint64
	LockRequestsPersec         uint64
	LockTimeoutsPersec         uint64
	LockTimeoutstimeout0Persec uint64
	LockWaitsPersec            uint64
	LockWaitTimems             uint64
	NumberofDeadlocksPersec    uint64
}

func (c *MSSQLLocksCollector) collectLocks(ch chan<- prometheus.Metric, sqlInstance string) (*prometheus.Desc, error) {
	var dst []win32PerfRawDataSQLServerLocks
	class := fmt.Sprintf("Win32_PerfRawData_%s_SQLServerLocks", sqlInstance)
	q := queryAllForClassWhere(&dst, class, `Name <> '_Total'`)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	for _, v := range dst {
		lockResourceName := v.Name

		ch <- prometheus.MustNewConstMetric(
			c.AverageWaitTimems,
			prometheus.GaugeValue,
			float64(v.AverageWaitTimems)/1000.0,
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockRequestsPersec,
			prometheus.CounterValue,
			float64(v.LockRequestsPersec),
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockTimeoutsPersec,
			prometheus.CounterValue,
			float64(v.LockTimeoutsPersec),
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockTimeoutstimeout0Persec,
			prometheus.CounterValue,
			float64(v.LockTimeoutstimeout0Persec),
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockWaitsPersec,
			prometheus.CounterValue,
			float64(v.LockWaitsPersec),
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockWaitTimems,
			prometheus.GaugeValue,
			float64(v.LockWaitTimems)/1000.0,
			sqlInstance, lockResourceName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberofDeadlocksPersec,
			prometheus.CounterValue,
			float64(v.NumberofDeadlocksPersec),
			sqlInstance, lockResourceName,
		)
	}

	return nil, nil
}

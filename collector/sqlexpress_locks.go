// returns data points from Win32_PerfRawData_MSSQLSQLEXPRESS_MSSQLSQLEXPRESSLocks
// <add link to documentation here> - Win32_PerfRawData_MSSQLSQLEXPRESS_MSSQLSQLEXPRESSLocks class
package collector

import (
	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	Factories["win32_perfrawdata_mssqlsqlexpress_mssqlsqlexpresssqlstatistics"] = NewWin32_PerfRawData_MSSQLSQLEXPRESS_MSSQLSQLEXPRESSSQLStatisticsCollector
}

// A Win32_PerfRawData_MSSQLSQLEXPRESS_MSSQLSQLEXPRESSSQLStatisticsCollector is a Prometheus collector for WMI Win32_PerfRawData_MSSQLSQLEXPRESS_MSSQLSQLEXPRESSLocks metrics
type Win32_PerfRawData_MSSQLSQLEXPRESS_MSSQLSQLEXPRESSSQLStatisticsCollector struct {
	AverageWaitTimems          *prometheus.Desc
	LockRequestsPersec         *prometheus.Desc
	LockTimeoutsPersec         *prometheus.Desc
	LockTimeoutstimeout0Persec *prometheus.Desc
	LockWaitsPersec            *prometheus.Desc
	LockWaitTimems             *prometheus.Desc
	NumberofDeadlocksPersec    *prometheus.Desc
}

// NewWin32_PerfRawData_MSSQLSQLEXPRESS_MSSQLSQLEXPRESSSQLStatisticsCollector ...
func NewWin32_PerfRawData_MSSQLSQLEXPRESS_MSSQLSQLEXPRESSSQLStatisticsCollector() (Collector, error) {
	const subsystem = "win32_perfrawdata_mssqlsqlexpress_mssqlsqlexpresssqlstatistics"
	return &Win32_PerfRawData_MSSQLSQLEXPRESS_MSSQLSQLEXPRESSSQLStatisticsCollector{
		AverageWaitTimems: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "average_wait_timems"),
			"(AverageWaitTimems)",
			nil,
			nil,
		),
		LockRequestsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_requests_persec"),
			"(LockRequestsPersec)",
			nil,
			nil,
		),
		LockTimeoutsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_timeouts_persec"),
			"(LockTimeoutsPersec)",
			nil,
			nil,
		),
		LockTimeoutstimeout0Persec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_timeoutstimeout0_persec"),
			"(LockTimeoutstimeout0Persec)",
			nil,
			nil,
		),
		LockWaitsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_waits_persec"),
			"(LockWaitsPersec)",
			nil,
			nil,
		),
		LockWaitTimems: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_wait_timems"),
			"(LockWaitTimems)",
			nil,
			nil,
		),
		NumberofDeadlocksPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "numberof_deadlocks_persec"),
			"(NumberofDeadlocksPersec)",
			nil,
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Win32_PerfRawData_MSSQLSQLEXPRESS_MSSQLSQLEXPRESSSQLStatisticsCollector) Collect(ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting win32_perfrawdata_mssqlsqlexpress_mssqlsqlexpresssqlstatistics metrics:", desc, err)
		return err
	}
	return nil
}

type Win32_PerfRawData_MSSQLSQLEXPRESS_MSSQLSQLEXPRESSLocks struct {
	Name string

	AverageWaitTimems          uint64
	LockRequestsPersec         uint64
	LockTimeoutsPersec         uint64
	LockTimeoutstimeout0Persec uint64
	LockWaitsPersec            uint64
	LockWaitTimems             uint64
	NumberofDeadlocksPersec    uint64
}

func (c *Win32_PerfRawData_MSSQLSQLEXPRESS_MSSQLSQLEXPRESSSQLStatisticsCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_MSSQLSQLEXPRESS_MSSQLSQLEXPRESSLocks
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	ch <- prometheus.MustNewConstMetric(
		c.AverageWaitTimems,
		prometheus.GaugeValue,
		float64(dst[0].AverageWaitTimems),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LockRequestsPersec,
		prometheus.GaugeValue,
		float64(dst[0].LockRequestsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LockTimeoutsPersec,
		prometheus.GaugeValue,
		float64(dst[0].LockTimeoutsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LockTimeoutstimeout0Persec,
		prometheus.GaugeValue,
		float64(dst[0].LockTimeoutstimeout0Persec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LockWaitsPersec,
		prometheus.GaugeValue,
		float64(dst[0].LockWaitsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.LockWaitTimems,
		prometheus.GaugeValue,
		float64(dst[0].LockWaitTimems),
	)

	ch <- prometheus.MustNewConstMetric(
		c.NumberofDeadlocksPersec,
		prometheus.GaugeValue,
		float64(dst[0].NumberofDeadlocksPersec),
	)

	return nil, nil
}

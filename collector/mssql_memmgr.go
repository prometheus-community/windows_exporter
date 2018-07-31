// returns data points from the following classes:
// - Win32_PerfRawData_MSSQLSERVER_SQLServerMemoryManager
//   https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-memory-manager-object

package collector

import (
	"fmt"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	Factories["mssql_memmgr"] = NewMSSQLMemMgrCollector
}

// MSSQLMemMgrCollector is a Prometheus collector for Win32_PerfRawData_{instance}_SQLServerMemoryManager metrics
type MSSQLMemMgrCollector struct {
	ConnectionMemoryKB       *prometheus.Desc
	DatabaseCacheMemoryKB    *prometheus.Desc
	Externalbenefitofmemory  *prometheus.Desc
	FreeMemoryKB             *prometheus.Desc
	GrantedWorkspaceMemoryKB *prometheus.Desc
	LockBlocks               *prometheus.Desc
	LockBlocksAllocated      *prometheus.Desc
	LockMemoryKB             *prometheus.Desc
	LockOwnerBlocks          *prometheus.Desc
	LockOwnerBlocksAllocated *prometheus.Desc
	LogPoolMemoryKB          *prometheus.Desc
	MaximumWorkspaceMemoryKB *prometheus.Desc
	MemoryGrantsOutstanding  *prometheus.Desc
	MemoryGrantsPending      *prometheus.Desc
	OptimizerMemoryKB        *prometheus.Desc
	ReservedServerMemoryKB   *prometheus.Desc
	SQLCacheMemoryKB         *prometheus.Desc
	StolenServerMemoryKB     *prometheus.Desc
	TargetServerMemoryKB     *prometheus.Desc
	TotalServerMemoryKB      *prometheus.Desc

	sqlInstances sqlInstancesType
}

// NewMSSQLMemMgrCollector ...
func NewMSSQLMemMgrCollector() (Collector, error) {

	const subsystem = "mssql_memmgr"
	return &MSSQLMemMgrCollector{

		// Win32_PerfRawData_{instance}_SQLServerMemoryManager
		ConnectionMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "connection_memory_bytes"),
			"(MemoryManager.ConnectionMemoryKB)",
			[]string{"instance"},
			nil,
		),
		DatabaseCacheMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "database_cache_memory_bytes"),
			"(MemoryManager.DatabaseCacheMemoryKB)",
			[]string{"instance"},
			nil,
		),
		Externalbenefitofmemory: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "external_benefit_of_memory"),
			"(MemoryManager.Externalbenefitofmemory)",
			[]string{"instance"},
			nil,
		),
		FreeMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "free_memory_bytes"),
			"(MemoryManager.FreeMemoryKB)",
			[]string{"instance"},
			nil,
		),
		GrantedWorkspaceMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "granted_workspace_memory_bytes"),
			"(MemoryManager.GrantedWorkspaceMemoryKB)",
			[]string{"instance"},
			nil,
		),
		LockBlocks: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_blocks"),
			"(MemoryManager.LockBlocks)",
			[]string{"instance"},
			nil,
		),
		LockBlocksAllocated: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "allocated_lock_blocks"),
			"(MemoryManager.LockBlocksAllocated)",
			[]string{"instance"},
			nil,
		),
		LockMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_memory_bytes"),
			"(MemoryManager.LockMemoryKB)",
			[]string{"instance"},
			nil,
		),
		LockOwnerBlocks: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lock_owner_blocks"),
			"(MemoryManager.LockOwnerBlocks)",
			[]string{"instance"},
			nil,
		),
		LockOwnerBlocksAllocated: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "allocated_lock_owner_blocks"),
			"(MemoryManager.LockOwnerBlocksAllocated)",
			[]string{"instance"},
			nil,
		),
		LogPoolMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_pool_memory_bytes"),
			"(MemoryManager.LogPoolMemoryKB)",
			[]string{"instance"},
			nil,
		),
		MaximumWorkspaceMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "maximum_workspace_memory_bytes"),
			"(MemoryManager.MaximumWorkspaceMemoryKB)",
			[]string{"instance"},
			nil,
		),
		MemoryGrantsOutstanding: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "outstanding_memory_grants"),
			"(MemoryManager.MemoryGrantsOutstanding)",
			[]string{"instance"},
			nil,
		),
		MemoryGrantsPending: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pending_memory_grants"),
			"(MemoryManager.MemoryGrantsPending)",
			[]string{"instance"},
			nil,
		),
		OptimizerMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "optimizer_memory_bytes"),
			"(MemoryManager.OptimizerMemoryKB)",
			[]string{"instance"},
			nil,
		),
		ReservedServerMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "reserved_server_memory_bytes"),
			"(MemoryManager.ReservedServerMemoryKB)",
			[]string{"instance"},
			nil,
		),
		SQLCacheMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sql_cache_memory_bytes"),
			"(MemoryManager.SQLCacheMemoryKB)",
			[]string{"instance"},
			nil,
		),
		StolenServerMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "stolen_server_memory_bytes"),
			"(MemoryManager.StolenServerMemoryKB)",
			[]string{"instance"},
			nil,
		),
		TargetServerMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "target_server_memory_bytes"),
			"(MemoryManager.TargetServerMemoryKB)",
			[]string{"instance"},
			nil,
		),
		TotalServerMemoryKB: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "total_server_memory_bytes"),
			"(MemoryManager.TotalServerMemoryKB)",
			[]string{"instance"},
			nil,
		),

		sqlInstances: getMSSQLInstances(),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *MSSQLMemMgrCollector) Collect(ch chan<- prometheus.Metric) error {
	for instance := range c.sqlInstances {
		log.Debugf("mssql collector iterating sql instance %s.", instance)

		// Win32_PerfRawData_{instance}_SQLServerMemoryManager
		if desc, err := c.collectMemoryManager(ch, instance); err != nil {
			log.Error("failed collecting MSSQL MemoryManager metrics:", desc, err)
			return err
		}
	}

	return nil
}

type win32PerfRawDataSQLServerMemoryManager struct {
	ConnectionMemoryKB       uint64
	DatabaseCacheMemoryKB    uint64
	Externalbenefitofmemory  uint64
	FreeMemoryKB             uint64
	GrantedWorkspaceMemoryKB uint64
	LockBlocks               uint64
	LockBlocksAllocated      uint64
	LockMemoryKB             uint64
	LockOwnerBlocks          uint64
	LockOwnerBlocksAllocated uint64
	LogPoolMemoryKB          uint64
	MaximumWorkspaceMemoryKB uint64
	MemoryGrantsOutstanding  uint64
	MemoryGrantsPending      uint64
	OptimizerMemoryKB        uint64
	ReservedServerMemoryKB   uint64
	SQLCacheMemoryKB         uint64
	StolenServerMemoryKB     uint64
	TargetServerMemoryKB     uint64
	TotalServerMemoryKB      uint64
}

func (c *MSSQLMemMgrCollector) collectMemoryManager(ch chan<- prometheus.Metric, sqlInstance string) (*prometheus.Desc, error) {
	var dst []win32PerfRawDataSQLServerMemoryManager
	class := fmt.Sprintf("Win32_PerfRawData_%s_SQLServerMemoryManager", sqlInstance)
	q := queryAllForClass(&dst, class)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	if len(dst) > 0 {
		v := dst[0]

		ch <- prometheus.MustNewConstMetric(
			c.ConnectionMemoryKB,
			prometheus.GaugeValue,
			float64(v.ConnectionMemoryKB*1024),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabaseCacheMemoryKB,
			prometheus.GaugeValue,
			float64(v.DatabaseCacheMemoryKB*1024),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Externalbenefitofmemory,
			prometheus.GaugeValue,
			float64(v.Externalbenefitofmemory),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FreeMemoryKB,
			prometheus.GaugeValue,
			float64(v.FreeMemoryKB*1024),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GrantedWorkspaceMemoryKB,
			prometheus.GaugeValue,
			float64(v.GrantedWorkspaceMemoryKB*1024),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockBlocks,
			prometheus.GaugeValue,
			float64(v.LockBlocks),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockBlocksAllocated,
			prometheus.GaugeValue,
			float64(v.LockBlocksAllocated),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockMemoryKB,
			prometheus.GaugeValue,
			float64(v.LockMemoryKB*1024),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockOwnerBlocks,
			prometheus.GaugeValue,
			float64(v.LockOwnerBlocks),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LockOwnerBlocksAllocated,
			prometheus.GaugeValue,
			float64(v.LockOwnerBlocksAllocated),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogPoolMemoryKB,
			prometheus.GaugeValue,
			float64(v.LogPoolMemoryKB*1024),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MaximumWorkspaceMemoryKB,
			prometheus.GaugeValue,
			float64(v.MaximumWorkspaceMemoryKB*1024),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MemoryGrantsOutstanding,
			prometheus.GaugeValue,
			float64(v.MemoryGrantsOutstanding),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MemoryGrantsPending,
			prometheus.GaugeValue,
			float64(v.MemoryGrantsPending),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.OptimizerMemoryKB,
			prometheus.GaugeValue,
			float64(v.OptimizerMemoryKB*1024),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReservedServerMemoryKB,
			prometheus.GaugeValue,
			float64(v.ReservedServerMemoryKB*1024),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SQLCacheMemoryKB,
			prometheus.GaugeValue,
			float64(v.SQLCacheMemoryKB*1024),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.StolenServerMemoryKB,
			prometheus.GaugeValue,
			float64(v.StolenServerMemoryKB*1024),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TargetServerMemoryKB,
			prometheus.GaugeValue,
			float64(v.TargetServerMemoryKB*1024),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalServerMemoryKB,
			prometheus.GaugeValue,
			float64(v.TotalServerMemoryKB*1024),
			sqlInstance,
		)
	}

	return nil, nil
}

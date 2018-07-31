// returns data points from the following classes:
// - Win32_PerfRawData_MSSQLSERVER_SQLServerBufferManager
//   https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-buffer-manager-object

package collector

import (
	"fmt"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	Factories["mssql_bufman"] = NewMSSQLBufManCollector
}

// MSSQLBufManCollector is a Prometheus collector for Win32_PerfRawData_{instance}_SQLServerBufferManager metrics
type MSSQLBufManCollector struct {
	// Win32_PerfRawData_{instance}_SQLServerBufferManager
	BackgroundwriterpagesPersec   *prometheus.Desc
	Buffercachehitratio           *prometheus.Desc
	CheckpointpagesPersec         *prometheus.Desc
	Databasepages                 *prometheus.Desc
	Extensionallocatedpages       *prometheus.Desc
	Extensionfreepages            *prometheus.Desc
	Extensioninuseaspercentage    *prometheus.Desc
	ExtensionoutstandingIOcounter *prometheus.Desc
	ExtensionpageevictionsPersec  *prometheus.Desc
	ExtensionpagereadsPersec      *prometheus.Desc
	Extensionpageunreferencedtime *prometheus.Desc
	ExtensionpagewritesPersec     *prometheus.Desc
	FreeliststallsPersec          *prometheus.Desc
	IntegralControllerSlope       *prometheus.Desc
	LazywritesPersec              *prometheus.Desc
	Pagelifeexpectancy            *prometheus.Desc
	PagelookupsPersec             *prometheus.Desc
	PagereadsPersec               *prometheus.Desc
	PagewritesPersec              *prometheus.Desc
	ReadaheadpagesPersec          *prometheus.Desc
	ReadaheadtimePersec           *prometheus.Desc
	Targetpages                   *prometheus.Desc

	sqlInstances sqlInstancesType
}

// NewMSSQLBufManCollector ...
func NewMSSQLBufManCollector() (Collector, error) {

	const subsystem = "mssql_bufman"
	return &MSSQLBufManCollector{

		// Win32_PerfRawData_{instance}_SQLServerBufferManager
		BackgroundwriterpagesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "background_writer_pages"),
			"(BufferManager.Backgroundwriterpages)",
			[]string{"instance"},
			nil,
		),
		Buffercachehitratio: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "buffer_cache_hit_ratio"),
			"(BufferManager.Buffercachehitratio)",
			[]string{"instance"},
			nil,
		),
		CheckpointpagesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "checkpoint_pages"),
			"(BufferManager.Checkpointpages)",
			[]string{"instance"},
			nil,
		),
		Databasepages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "database_pages"),
			"(BufferManager.Databasepages)",
			[]string{"instance"},
			nil,
		),
		Extensionallocatedpages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "extension_allocated_pages"),
			"(BufferManager.Extensionallocatedpages)",
			[]string{"instance"},
			nil,
		),
		Extensionfreepages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "extension_free_pages"),
			"(BufferManager.Extensionfreepages)",
			[]string{"instance"},
			nil,
		),
		Extensioninuseaspercentage: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "extension_in_use_as_percentage"),
			"(BufferManager.Extensioninuseaspercentage)",
			[]string{"instance"},
			nil,
		),
		ExtensionoutstandingIOcounter: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "extension_outstanding_io"),
			"(BufferManager.ExtensionoutstandingIOcounter)",
			[]string{"instance"},
			nil,
		),
		ExtensionpageevictionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "extension_page_evictions"),
			"(BufferManager.Extensionpageevictions)",
			[]string{"instance"},
			nil,
		),
		ExtensionpagereadsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "extension_page_reads"),
			"(BufferManager.Extensionpagereads)",
			[]string{"instance"},
			nil,
		),
		Extensionpageunreferencedtime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "extension_page_unreferenced_seconds"),
			"(BufferManager.Extensionpageunreferencedtime)",
			[]string{"instance"},
			nil,
		),
		ExtensionpagewritesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "extension_page_writes"),
			"(BufferManager.Extensionpagewrites)",
			[]string{"instance"},
			nil,
		),
		FreeliststallsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "free_list_stalls"),
			"(BufferManager.Freeliststalls)",
			[]string{"instance"},
			nil,
		),
		IntegralControllerSlope: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "integral_controller_slope"),
			"(BufferManager.IntegralControllerSlope)",
			[]string{"instance"},
			nil,
		),
		LazywritesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lazywrites"),
			"(BufferManager.Lazywrites)",
			[]string{"instance"},
			nil,
		),
		Pagelifeexpectancy: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "page_life_expectancy_seconds"),
			"(BufferManager.Pagelifeexpectancy)",
			[]string{"instance"},
			nil,
		),
		PagelookupsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "page_lookups"),
			"(BufferManager.Pagelookups)",
			[]string{"instance"},
			nil,
		),
		PagereadsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "page_reads"),
			"(BufferManager.Pagereads)",
			[]string{"instance"},
			nil,
		),
		PagewritesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "page_writes"),
			"(BufferManager.Pagewrites)",
			[]string{"instance"},
			nil,
		),
		ReadaheadpagesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "read_ahead_pages"),
			"(BufferManager.Readaheadpages)",
			[]string{"instance"},
			nil,
		),
		ReadaheadtimePersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "read_ahead_issuing_seconds"),
			"(BufferManager.Readaheadtime)",
			[]string{"instance"},
			nil,
		),
		Targetpages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "target_pages"),
			"(BufferManager.Targetpages)",
			[]string{"instance"},
			nil,
		),

		sqlInstances: getMSSQLInstances(),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *MSSQLBufManCollector) Collect(ch chan<- prometheus.Metric) error {
	for instance := range c.sqlInstances {
		log.Debugf("mssql_bufman collector iterating sql instance %s.", instance)

		// Win32_PerfRawData_{instance}_SQLServerBufferManager
		if desc, err := c.collectBufferManager(ch, instance); err != nil {
			log.Error("failed collecting MSSQL BufferManager metrics:", desc, err)
			return err
		}
	}

	return nil
}

type win32PerfRawDataSQLServerBufferManager struct {
	BackgroundwriterpagesPersec   uint64
	Buffercachehitratio           uint64
	CheckpointpagesPersec         uint64
	Databasepages                 uint64
	Extensionallocatedpages       uint64
	Extensionfreepages            uint64
	Extensioninuseaspercentage    uint64
	ExtensionoutstandingIOcounter uint64
	ExtensionpageevictionsPersec  uint64
	ExtensionpagereadsPersec      uint64
	Extensionpageunreferencedtime uint64
	ExtensionpagewritesPersec     uint64
	FreeliststallsPersec          uint64
	IntegralControllerSlope       uint64
	LazywritesPersec              uint64
	Pagelifeexpectancy            uint64
	PagelookupsPersec             uint64
	PagereadsPersec               uint64
	PagewritesPersec              uint64
	ReadaheadpagesPersec          uint64
	ReadaheadtimePersec           uint64
	Targetpages                   uint64
}

func (c *MSSQLBufManCollector) collectBufferManager(ch chan<- prometheus.Metric, sqlInstance string) (*prometheus.Desc, error) {
	var dst []win32PerfRawDataSQLServerBufferManager
	class := fmt.Sprintf("Win32_PerfRawData_%s_SQLServerBufferManager", sqlInstance)
	q := queryAllForClass(&dst, class)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	if len(dst) > 0 {
		v := dst[0]

		ch <- prometheus.MustNewConstMetric(
			c.BackgroundwriterpagesPersec,
			prometheus.CounterValue,
			float64(v.BackgroundwriterpagesPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Buffercachehitratio,
			prometheus.GaugeValue,
			float64(v.Buffercachehitratio),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.CheckpointpagesPersec,
			prometheus.CounterValue,
			float64(v.CheckpointpagesPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Databasepages,
			prometheus.GaugeValue,
			float64(v.Databasepages),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Extensionallocatedpages,
			prometheus.GaugeValue,
			float64(v.Extensionallocatedpages),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Extensionfreepages,
			prometheus.GaugeValue,
			float64(v.Extensionfreepages),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Extensioninuseaspercentage,
			prometheus.GaugeValue,
			float64(v.Extensioninuseaspercentage),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ExtensionoutstandingIOcounter,
			prometheus.GaugeValue,
			float64(v.ExtensionoutstandingIOcounter),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ExtensionpageevictionsPersec,
			prometheus.CounterValue,
			float64(v.ExtensionpageevictionsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ExtensionpagereadsPersec,
			prometheus.CounterValue,
			float64(v.ExtensionpagereadsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Extensionpageunreferencedtime,
			prometheus.GaugeValue,
			float64(v.Extensionpageunreferencedtime),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ExtensionpagewritesPersec,
			prometheus.CounterValue,
			float64(v.ExtensionpagewritesPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FreeliststallsPersec,
			prometheus.CounterValue,
			float64(v.FreeliststallsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.IntegralControllerSlope,
			prometheus.GaugeValue,
			float64(v.IntegralControllerSlope),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LazywritesPersec,
			prometheus.CounterValue,
			float64(v.LazywritesPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Pagelifeexpectancy,
			prometheus.GaugeValue,
			float64(v.Pagelifeexpectancy),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PagelookupsPersec,
			prometheus.CounterValue,
			float64(v.PagelookupsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PagereadsPersec,
			prometheus.CounterValue,
			float64(v.PagereadsPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PagewritesPersec,
			prometheus.CounterValue,
			float64(v.PagewritesPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReadaheadpagesPersec,
			prometheus.CounterValue,
			float64(v.ReadaheadpagesPersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReadaheadtimePersec,
			prometheus.CounterValue,
			float64(v.ReadaheadtimePersec),
			sqlInstance,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Targetpages,
			prometheus.GaugeValue,
			float64(v.Targetpages),
			sqlInstance,
		)
	}

	return nil, nil
}

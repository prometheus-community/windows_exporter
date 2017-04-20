// returns data points from Win32_PerfRawData_APPPOOLCountersProvider_APPPOOLWAS
// <add link to documentation here> - Win32_PerfRawData_APPPOOLCountersProvider_APPPOOLWAS class
package collector

import (
	"log"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	Factories["iis_apppool"] = NewIISAppPoolCollector
}

// A IISAppPoolCollector is a Prometheus collector for WMI Win32_PerfRawData_APPPOOLCountersProvider_APPPOOLWAS metrics
type IISAppPoolCollector struct {
	CurrentApplicationPoolState        *prometheus.Desc
	CurrentApplicationPoolUptime       *prometheus.Desc
	CurrentWorkerProcesses             *prometheus.Desc
	MaximumWorkerProcesses             *prometheus.Desc
	RecentWorkerProcessFailures        *prometheus.Desc
	TimeSinceLastWorkerProcessFailure  *prometheus.Desc
	TotalApplicationPoolRecycles       *prometheus.Desc
	TotalApplicationPoolUptime         *prometheus.Desc
	TotalWorkerProcessesCreated        *prometheus.Desc
	TotalWorkerProcessFailures         *prometheus.Desc
	TotalWorkerProcessPingFailures     *prometheus.Desc
	TotalWorkerProcessShutdownFailures *prometheus.Desc
	TotalWorkerProcessStartupFailures  *prometheus.Desc
}

// NewIISAppPoolCollector ...
func NewIISAppPoolCollector() (Collector, error) {
	const subsystem = "iis_apppool"
	return &IISAppPoolCollector{
		CurrentApplicationPoolState: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "current_application_pool_state"),
			"The current status of the application pool (1 - Uninitialized, 2 - Initialized, 3 - Running, 4 - Disabling, 5 - Disabled, 6 - Shutdown Pending, 7 - Delete Pending) (CurrentApplicationPoolState)",
			[]string{"app"},
			nil,
		),
		CurrentApplicationPoolUptime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "current_application_pool_uptime"),
			"The length of time, in seconds, that the application pool has been running since it was started (CurrentApplicationPoolUptime)",
			[]string{"app"},
			nil,
		),
		CurrentWorkerProcesses: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "current_worker_processes"),
			"The current number of worker processes that are running in the application pool (CurrentWorkerProcesses)",
			[]string{"app"},
			nil,
		),
		MaximumWorkerProcesses: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "maximum_worker_processes"),
			"The maximum number of worker processes that have been created for the application pool since Windows Process Activation Service (WAS) started (MaximumWorkerProcesses)",
			[]string{"app"},
			nil,
		),
		RecentWorkerProcessFailures: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "recent_worker_process_failures"),
			"The number of times that worker processes for the application pool failed during the rapid-fail protection interval (RecentWorkerProcessFailures)",
			[]string{"app"},
			nil,
		),
		TimeSinceLastWorkerProcessFailure: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "time_since_last_worker_process_failure"),
			"The length of time, in seconds, since the last worker process failure occurred for the application pool (TimeSinceLastWorkerProcessFailure)",
			[]string{"app"},
			nil,
		),
		TotalApplicationPoolRecycles: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "total_application_pool_recycles"),
			"The number of times that the application pool has been recycled since Windows Process Activation Service (WAS) started (TotalApplicationPoolRecycles)",
			[]string{"app"},
			nil,
		),
		TotalApplicationPoolUptime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "total_application_pool_uptime"),
			"The length of time, in seconds, that the application pool has been running since Windows Process Activation Service (WAS) started (TotalApplicationPoolUptime)",
			[]string{"app"},
			nil,
		),
		TotalWorkerProcessesCreated: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "total_worker_processes_created"),
			"The number of worker processes created for the application pool since Windows Process Activation Service (WAS) started (TotalWorkerProcessesCreated)",
			[]string{"app"},
			nil,
		),
		TotalWorkerProcessFailures: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "total_worker_process_failures"),
			"The number of times that worker processes have crashed since the application pool was started (TotalWorkerProcessFailures)",
			[]string{"app"},
			nil,
		),
		TotalWorkerProcessPingFailures: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "total_worker_process_ping_failures"),
			"The number of times that Windows Process Activation Service (WAS) did not receive a response to ping messages sent to a worker process (TotalWorkerProcessPingFailures)",
			[]string{"app"},
			nil,
		),
		TotalWorkerProcessShutdownFailures: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "total_worker_process_shutdown_failures"),
			"The number of times that Windows Process Activation Service (WAS) failed to shut down a worker process (TotalWorkerProcessShutdownFailures)",
			[]string{"app"},
			nil,
		),
		TotalWorkerProcessStartupFailures: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "total_worker_process_startup_failures"),
			"The number of times that Windows Process Activation Service (WAS) failed to start a worker process (TotalWorkerProcessStartupFailures)",
			[]string{"app"},
			nil,
		),


		appWhitelistPattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *siteWhitelist)),
		appBlacklistPattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *siteBlacklist)),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *IISAppPoolCollector) Collect(ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Println("[ERROR] failed collecting apppool metrics:", desc, err)
		return err
	}
	return nil
}

type Win32_PerfRawData_APPPOOLCountersProvider_APPPOOLWAS struct {
	Name string

	CurrentApplicationPoolState        uint32
	CurrentApplicationPoolUptime       uint64
	CurrentWorkerProcesses             uint32
	MaximumWorkerProcesses             uint32
	RecentWorkerProcessFailures        uint32
	TimeSinceLastWorkerProcessFailure  uint64
	TotalApplicationPoolRecycles       uint32
	TotalApplicationPoolUptime         uint64
	TotalWorkerProcessesCreated        uint32
	TotalWorkerProcessFailures         uint32
	TotalWorkerProcessPingFailures     uint32
	TotalWorkerProcessShutdownFailures uint32
	TotalWorkerProcessStartupFailures  uint32
}

func (c *IISAppPoolCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_APPPOOLCountersProvider_APPPOOLWAS
	q := wmi.CreateQuery(&dst, "")
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	for _, app := range dst {
		if app.Name == "_Total" ||
			c.appBlacklistPattern.MatchString(app.Name) ||
			!c.appWhitelistPattern.MatchString(app.Name) {
			continue
		}

		// Guages
		ch <- prometheus.MustNewConstMetric(
			c.CurrentApplicationPoolState,
			prometheus.GaugeValue,
			float64(app.CurrentApplicationPoolState),
			app.NAME
		)

		ch <- prometheus.MustNewConstMetric(
			c.CurrentApplicationPoolUptime,
			prometheus.GaugeValue,
			float64(app.CurrentApplicationPoolUptime),
			app.NAME
		)

		ch <- prometheus.MustNewConstMetric(
			c.CurrentWorkerProcesses,
			prometheus.GaugeValue,
			float64(app.CurrentWorkerProcesses),
			app.NAME
		)

		ch <- prometheus.MustNewConstMetric(
			c.MaximumWorkerProcesses,
			prometheus.GaugeValue,
			float64(app.MaximumWorkerProcesses),
			app.NAME
		)

		ch <- prometheus.MustNewConstMetric(
			c.RecentWorkerProcessFailures,
			prometheus.GaugeValue,
			float64(app.RecentWorkerProcessFailures),
			app.NAME
		)

		ch <- prometheus.MustNewConstMetric(
			c.TimeSinceLastWorkerProcessFailure,
			prometheus.GaugeValue,
			float64(app.TimeSinceLastWorkerProcessFailure),
			app.NAME
		)

		// Counters
		ch <- prometheus.MustNewConstMetric(
			c.TotalApplicationPoolRecycles,
			prometheus.CounterValue,
			float64(app.TotalApplicationPoolRecycles),
			app.NAME
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalApplicationPoolUptime,
			prometheus.CounterValue,
			float64(app.TotalApplicationPoolUptime),
			app.NAME
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalWorkerProcessesCreated,
			prometheus.CounterValue,
			float64(app.TotalWorkerProcessesCreated),
			app.NAME
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalWorkerProcessFailures,
			prometheus.CounterValue,
			float64(app.TotalWorkerProcessFailures),
			app.NAME
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalWorkerProcessPingFailures,
			prometheus.CounterValue,
			float64(app.TotalWorkerProcessPingFailures),
			app.NAME
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalWorkerProcessShutdownFailures,
			prometheus.CounterValue,
			float64(app.TotalWorkerProcessShutdownFailures),
			app.NAME
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalWorkerProcessStartupFailures,
			prometheus.CounterValue,
			float64(app.TotalWorkerProcessStartupFailures),
			app.NAME
		)

	}

	return nil, nil
}

//go:build windows

package iis

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorAppPoolWAS struct {
	perfDataCollectorAppPoolWAS *perfdata.Collector

	currentApplicationPoolState        *prometheus.Desc
	currentApplicationPoolUptime       *prometheus.Desc
	currentWorkerProcesses             *prometheus.Desc
	maximumWorkerProcesses             *prometheus.Desc
	recentWorkerProcessFailures        *prometheus.Desc
	timeSinceLastWorkerProcessFailure  *prometheus.Desc
	totalApplicationPoolRecycles       *prometheus.Desc
	totalApplicationPoolUptime         *prometheus.Desc
	totalWorkerProcessesCreated        *prometheus.Desc
	totalWorkerProcessFailures         *prometheus.Desc
	totalWorkerProcessPingFailures     *prometheus.Desc
	totalWorkerProcessShutdownFailures *prometheus.Desc
	totalWorkerProcessStartupFailures  *prometheus.Desc
}

const (
	CurrentApplicationPoolState        = "Current Application Pool State"
	CurrentApplicationPoolUptime       = "Current Application Pool Uptime"
	CurrentWorkerProcesses             = "Current Worker Processes"
	MaximumWorkerProcesses             = "Maximum Worker Processes"
	RecentWorkerProcessFailures        = "Recent Worker Process Failures"
	TimeSinceLastWorkerProcessFailure  = "Time Since Last Worker Process Failure"
	TotalApplicationPoolRecycles       = "Total Application Pool Recycles"
	TotalApplicationPoolUptime         = "Total Application Pool Uptime"
	TotalWorkerProcessesCreated        = "Total Worker Processes Created"
	TotalWorkerProcessFailures         = "Total Worker Process Failures"
	TotalWorkerProcessPingFailures     = "Total Worker Process Ping Failures"
	TotalWorkerProcessShutdownFailures = "Total Worker Process Shutdown Failures"
	TotalWorkerProcessStartupFailures  = "Total Worker Process Startup Failures"
)

var applicationStates = map[uint32]string{
	1: "Uninitialized",
	2: "Initialized",
	3: "Running",
	4: "Disabling",
	5: "Disabled",
	6: "Shutdown Pending",
	7: "Delete Pending",
}

func (c *Collector) buildAppPoolWAS() error {
	var err error

	c.perfDataCollectorAppPoolWAS, err = perfdata.NewCollector("APP_POOL_WAS", perfdata.InstanceAll, []string{
		CurrentApplicationPoolState,
		CurrentApplicationPoolUptime,
		CurrentWorkerProcesses,
		MaximumWorkerProcesses,
		RecentWorkerProcessFailures,
		TimeSinceLastWorkerProcessFailure,
		TotalApplicationPoolRecycles,
		TotalApplicationPoolUptime,
		TotalWorkerProcessesCreated,
		TotalWorkerProcessFailures,
		TotalWorkerProcessPingFailures,
		TotalWorkerProcessShutdownFailures,
		TotalWorkerProcessStartupFailures,
	})
	if err != nil {
		return fmt.Errorf("failed to create APP_POOL_WAS collector: %w", err)
	}

	// APP_POOL_WAS
	c.currentApplicationPoolState = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_application_pool_state"),
		"The current status of the application pool (1 - Uninitialized, 2 - Initialized, 3 - Running, 4 - Disabling, 5 - Disabled, 6 - Shutdown Pending, 7 - Delete Pending) (CurrentApplicationPoolState)",
		[]string{"app", "state"},
		nil,
	)
	c.currentApplicationPoolUptime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_application_pool_start_time"),
		"The unix timestamp for the application pool start time (CurrentApplicationPoolUptime)",
		[]string{"app"},
		nil,
	)
	c.currentWorkerProcesses = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_worker_processes"),
		"The current number of worker processes that are running in the application pool (CurrentWorkerProcesses)",
		[]string{"app"},
		nil,
	)
	c.maximumWorkerProcesses = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "maximum_worker_processes"),
		"The maximum number of worker processes that have been created for the application pool since Windows Process Activation Service (WAS) started (MaximumWorkerProcesses)",
		[]string{"app"},
		nil,
	)
	c.recentWorkerProcessFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "recent_worker_process_failures"),
		"The number of times that worker processes for the application pool failed during the rapid-fail protection interval (RecentWorkerProcessFailures)",
		[]string{"app"},
		nil,
	)
	c.timeSinceLastWorkerProcessFailure = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "time_since_last_worker_process_failure"),
		"The length of time, in seconds, since the last worker process failure occurred for the application pool (TimeSinceLastWorkerProcessFailure)",
		[]string{"app"},
		nil,
	)
	c.totalApplicationPoolRecycles = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "total_application_pool_recycles"),
		"The number of times that the application pool has been recycled since Windows Process Activation Service (WAS) started (TotalApplicationPoolRecycles)",
		[]string{"app"},
		nil,
	)
	c.totalApplicationPoolUptime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "total_application_pool_start_time"),
		"The unix timestamp for the application pool of when the Windows Process Activation Service (WAS) started (TotalApplicationPoolUptime)",
		[]string{"app"},
		nil,
	)
	c.totalWorkerProcessesCreated = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "total_worker_processes_created"),
		"The number of worker processes created for the application pool since Windows Process Activation Service (WAS) started (TotalWorkerProcessesCreated)",
		[]string{"app"},
		nil,
	)
	c.totalWorkerProcessFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "total_worker_process_failures"),
		"The number of times that worker processes have crashed since the application pool was started (TotalWorkerProcessFailures)",
		[]string{"app"},
		nil,
	)
	c.totalWorkerProcessPingFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "total_worker_process_ping_failures"),
		"The number of times that Windows Process Activation Service (WAS) did not receive a response to ping messages sent to a worker process (TotalWorkerProcessPingFailures)",
		[]string{"app"},
		nil,
	)
	c.totalWorkerProcessShutdownFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "total_worker_process_shutdown_failures"),
		"The number of times that Windows Process Activation Service (WAS) failed to shut down a worker process (TotalWorkerProcessShutdownFailures)",
		[]string{"app"},
		nil,
	)
	c.totalWorkerProcessStartupFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "total_worker_process_startup_failures"),
		"The number of times that Windows Process Activation Service (WAS) failed to start a worker process (TotalWorkerProcessStartupFailures)",
		[]string{"app"},
		nil,
	)

	return nil
}

func (c *Collector) collectAppPoolWAS(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollectorWebService.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect APP_POOL_WAS metrics: %w", err)
	}

	deduplicateIISNames(perfData)

	for name, app := range perfData {
		if c.config.AppExclude.MatchString(name) || !c.config.AppInclude.MatchString(name) {
			continue
		}

		for key, label := range applicationStates {
			isCurrentState := 0.0
			if key == uint32(app[CurrentApplicationPoolState].FirstValue) {
				isCurrentState = 1.0
			}

			ch <- prometheus.MustNewConstMetric(
				c.currentApplicationPoolState,
				prometheus.GaugeValue,
				isCurrentState,
				name,
				label,
			)
		}

		ch <- prometheus.MustNewConstMetric(
			c.currentApplicationPoolUptime,
			prometheus.GaugeValue,
			app[CurrentApplicationPoolUptime].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.currentWorkerProcesses,
			prometheus.GaugeValue,
			app[CurrentWorkerProcesses].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.maximumWorkerProcesses,
			prometheus.GaugeValue,
			app[MaximumWorkerProcesses].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.recentWorkerProcessFailures,
			prometheus.GaugeValue,
			app[RecentWorkerProcessFailures].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.timeSinceLastWorkerProcessFailure,
			prometheus.GaugeValue,
			app[TimeSinceLastWorkerProcessFailure].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalApplicationPoolRecycles,
			prometheus.CounterValue,
			app[TotalApplicationPoolRecycles].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalApplicationPoolUptime,
			prometheus.CounterValue,
			app[TotalApplicationPoolUptime].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalWorkerProcessesCreated,
			prometheus.CounterValue,
			app[TotalWorkerProcessesCreated].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalWorkerProcessFailures,
			prometheus.CounterValue,
			app[TotalWorkerProcessFailures].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalWorkerProcessPingFailures,
			prometheus.CounterValue,
			app[TotalWorkerProcessPingFailures].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalWorkerProcessShutdownFailures,
			prometheus.CounterValue,
			app[TotalWorkerProcessShutdownFailures].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalWorkerProcessStartupFailures,
			prometheus.CounterValue,
			app[TotalWorkerProcessStartupFailures].FirstValue,
			name,
		)
	}

	return nil
}

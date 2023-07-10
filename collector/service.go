//go:build windows
// +build windows

package collector

import (
	"fmt"
	"strings"
	"syscall"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/mgr"
)

const (
	FlagServiceWhereClause = "collector.service.services-where"
	FlagServiceUseAPI      = "collector.service.use-api"
)

var (
	serviceWhereClause *string
	useAPI             *bool
)

// A serviceCollector is a Prometheus collector for WMI Win32_Service metrics
type serviceCollector struct {
	logger log.Logger

	Information *prometheus.Desc
	State       *prometheus.Desc
	StartMode   *prometheus.Desc
	Status      *prometheus.Desc

	queryWhereClause string
}

// newServiceCollectorFlags ...
func newServiceCollectorFlags(app *kingpin.Application) {
	serviceWhereClause = app.Flag(
		FlagServiceWhereClause,
		"WQL 'where' clause to use in WMI metrics query. Limits the response to the services you specify and reduces the size of the response.",
	).Default("").String()
	useAPI = app.Flag(
		FlagServiceUseAPI,
		"Use API calls to collect service data instead of WMI. Flag 'collector.service.services-where' won't be effective.",
	).Default("false").Bool()
}

// newserviceCollector ...
func newserviceCollector(logger log.Logger) (Collector, error) {
	const subsystem = "service"
	logger = log.With(logger, "collector", subsystem)

	if *serviceWhereClause == "" {
		_ = level.Warn(logger).Log("msg", "No where-clause specified for service collector. This will generate a very large number of metrics!")
	}
	if *useAPI {
		_ = level.Warn(logger).Log("msg", "API collection is enabled.")
	}

	return &serviceCollector{
		logger: logger,

		Information: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "info"),
			"A metric with a constant '1' value labeled with service information",
			[]string{"name", "display_name", "process_id", "run_as"},
			nil,
		),
		State: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "state"),
			"The state of the service (State)",
			[]string{"name", "state"},
			nil,
		),
		StartMode: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "start_mode"),
			"The start mode of the service (StartMode)",
			[]string{"name", "start_mode"},
			nil,
		),
		Status: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "status"),
			"The status of the service (Status)",
			[]string{"name", "status"},
			nil,
		),
		queryWhereClause: *serviceWhereClause,
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *serviceCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if *useAPI {
		if err := c.collectAPI(ch); err != nil {
			_ = level.Error(c.logger).Log("msg", "failed collecting API service metrics:", "err", err)
			return err
		}
	} else {
		if err := c.collectWMI(ch); err != nil {
			_ = level.Error(c.logger).Log("msg", "failed collecting WMI service metrics:", "err", err)
			return err
		}
	}
	return nil
}

// Win32_Service docs:
// - https://msdn.microsoft.com/en-us/library/aa394418(v=vs.85).aspx
type Win32_Service struct {
	DisplayName string
	Name        string
	ProcessId   uint32
	State       string
	Status      string
	StartMode   string
	StartName   *string
}

var (
	allStates = []string{
		"stopped",
		"start pending",
		"stop pending",
		"running",
		"continue pending",
		"pause pending",
		"paused",
		"unknown",
	}
	apiStateValues = map[uint]string{
		windows.SERVICE_CONTINUE_PENDING: "continue pending",
		windows.SERVICE_PAUSE_PENDING:    "pause pending",
		windows.SERVICE_PAUSED:           "paused",
		windows.SERVICE_RUNNING:          "running",
		windows.SERVICE_START_PENDING:    "start pending",
		windows.SERVICE_STOP_PENDING:     "stop pending",
		windows.SERVICE_STOPPED:          "stopped",
	}
	allStartModes = []string{
		"boot",
		"system",
		"auto",
		"manual",
		"disabled",
	}
	apiStartModeValues = map[uint32]string{
		windows.SERVICE_AUTO_START:   "auto",
		windows.SERVICE_BOOT_START:   "boot",
		windows.SERVICE_DEMAND_START: "manual",
		windows.SERVICE_DISABLED:     "disabled",
		windows.SERVICE_SYSTEM_START: "system",
	}
	allStatuses = []string{
		"ok",
		"error",
		"degraded",
		"unknown",
		"pred fail",
		"starting",
		"stopping",
		"service",
		"stressed",
		"nonrecover",
		"no contact",
		"lost comm",
	}
)

func (c *serviceCollector) collectWMI(ch chan<- prometheus.Metric) error {
	var dst []Win32_Service
	q := queryAllWhere(&dst, c.queryWhereClause, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}
	for _, service := range dst {
		pid := fmt.Sprintf("%d", uint64(service.ProcessId))

		runAs := ""
		if service.StartName != nil {
			runAs = *service.StartName
		}
		ch <- prometheus.MustNewConstMetric(
			c.Information,
			prometheus.GaugeValue,
			1.0,
			strings.ToLower(service.Name),
			service.DisplayName,
			pid,
			runAs,
		)

		for _, state := range allStates {
			isCurrentState := 0.0
			if state == strings.ToLower(service.State) {
				isCurrentState = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				c.State,
				prometheus.GaugeValue,
				isCurrentState,
				strings.ToLower(service.Name),
				state,
			)
		}

		for _, startMode := range allStartModes {
			isCurrentStartMode := 0.0
			if startMode == strings.ToLower(service.StartMode) {
				isCurrentStartMode = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				c.StartMode,
				prometheus.GaugeValue,
				isCurrentStartMode,
				strings.ToLower(service.Name),
				startMode,
			)
		}

		for _, status := range allStatuses {
			isCurrentStatus := 0.0
			if status == strings.ToLower(service.Status) {
				isCurrentStatus = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				c.Status,
				prometheus.GaugeValue,
				isCurrentStatus,
				strings.ToLower(service.Name),
				status,
			)
		}
	}
	return nil
}

func (c *serviceCollector) collectAPI(ch chan<- prometheus.Metric) error {
	svcmgrConnection, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer svcmgrConnection.Disconnect() //nolint:errcheck

	// List All Services from the Services Manager.
	serviceList, err := svcmgrConnection.ListServices()
	if err != nil {
		return err
	}

	// Iterate through the Services List.
	for _, service := range serviceList {
		// Get UTF16 service name.
		serviceName, err := syscall.UTF16PtrFromString(service)
		if err != nil {
			_ = level.Warn(c.logger).Log("msg", fmt.Sprintf("Service %s get name error:  %#v", service, err))
			continue
		}

		// Open connection for service handler.
		serviceHandle, err := windows.OpenService(svcmgrConnection.Handle, serviceName, windows.GENERIC_READ)
		if err != nil {
			_ = level.Warn(c.logger).Log("msg", fmt.Sprintf("Open service %s error:  %#v", service, err))
			continue
		}

		// Create handle for each service.
		serviceManager := &mgr.Service{Name: service, Handle: serviceHandle}
		defer serviceManager.Close()

		// Get Service Configuration.
		serviceConfig, err := serviceManager.Config()
		if err != nil {
			_ = level.Warn(c.logger).Log("msg", fmt.Sprintf("Get ervice %s config error:  %#v", service, err))
			continue
		}

		// Get Service Current Status.
		serviceStatus, err := serviceManager.Query()
		if err != nil {
			_ = level.Warn(c.logger).Log("msg", fmt.Sprintf("Get service %s status error:  %#v", service, err))
			continue
		}

		pid := fmt.Sprintf("%d", uint64(serviceStatus.ProcessId))

		ch <- prometheus.MustNewConstMetric(
			c.Information,
			prometheus.GaugeValue,
			1.0,
			strings.ToLower(service),
			serviceConfig.DisplayName,
			pid,
			serviceConfig.ServiceStartName,
		)

		for _, state := range apiStateValues {
			isCurrentState := 0.0
			if state == apiStateValues[uint(serviceStatus.State)] {
				isCurrentState = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				c.State,
				prometheus.GaugeValue,
				isCurrentState,
				strings.ToLower(service),
				state,
			)
		}

		for _, startMode := range apiStartModeValues {
			isCurrentStartMode := 0.0
			if startMode == apiStartModeValues[serviceConfig.StartType] {
				isCurrentStartMode = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				c.StartMode,
				prometheus.GaugeValue,
				isCurrentStartMode,
				strings.ToLower(service),
				startMode,
			)
		}
	}
	return nil
}

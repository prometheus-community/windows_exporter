//go:build windows

package service

import (
	"fmt"
	"strconv"
	"strings"
	"syscall"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/mgr"
)

const (
	Name                   = "service"
	FlagServiceWhereClause = "collector.service.services-where"
	FlagServiceUseAPI      = "collector.service.use-api"
)

type Config struct {
	ServiceWhereClause string `yaml:"service_where_clause"`
	UseAPI             bool   `yaml:"use_api"`
}

var ConfigDefaults = Config{
	ServiceWhereClause: "",
	UseAPI:             false,
}

// A collector is a Prometheus collector for WMI Win32_Service metrics
type collector struct {
	logger log.Logger

	serviceWhereClause *string
	useAPI             *bool

	Information *prometheus.Desc
	State       *prometheus.Desc
	StartMode   *prometheus.Desc
	Status      *prometheus.Desc

	queryWhereClause string
}

func New(logger log.Logger, config *Config) types.Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &collector{
		serviceWhereClause: &config.ServiceWhereClause,
		useAPI:             &config.UseAPI,
	}
	c.SetLogger(logger)
	return c
}

func NewWithFlags(app *kingpin.Application) types.Collector {
	return &collector{
		serviceWhereClause: app.Flag(
			FlagServiceWhereClause,
			"WQL 'where' clause to use in WMI metrics query. Limits the response to the services you specify and reduces the size of the response.",
		).Default(ConfigDefaults.ServiceWhereClause).String(),
		useAPI: app.Flag(
			FlagServiceUseAPI,
			"Use API calls to collect service data instead of WMI. Flag 'collector.service.services-where' won't be effective.",
		).Default(strconv.FormatBool(ConfigDefaults.UseAPI)).Bool(),
	}
}

func (c *collector) GetName() string {
	return Name
}

func (c *collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *collector) GetPerfCounter() ([]string, error) {
	return []string{}, nil
}

func (c *collector) Build() error {
	if *c.serviceWhereClause == "" && !*c.useAPI {
		_ = level.Warn(c.logger).Log("msg", "No where-clause specified for service collector. This will generate a very large number of metrics!")
	}
	if *c.useAPI {
		_ = level.Warn(c.logger).Log("msg", "API collection is enabled.")
	}

	c.Information = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "info"),
		"A metric with a constant '1' value labeled with service information",
		[]string{"name", "display_name", "process_id", "run_as", "start_mode", "current_state"},
		nil,
	)
	c.State = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "state"),
		"The state of the service (1 - Stopped, 2 - Start Pending, 3 - Stop Pending, 4 - Running, 5 - Continue Pending, 6 - Pause Pending, 7 - Paused)",
		[]string{"name", "display_name"},
		nil,
	)
	c.StartMode = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "start_mode"),
		"The start mode of the service (0 - Boot, 1 - System, 2 - Automatic, 3 - Manual, 4 - Disabled)",
		[]string{"name", "display_name"},
		nil,
	)
	c.queryWhereClause = *c.serviceWhereClause
	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if *c.useAPI {
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
	// https://learn.microsoft.com/en-us/dotnet/api/system.serviceprocess.servicecontrollerstatus
	apiStateValues = map[uint]string{
		windows.SERVICE_CONTINUE_PENDING: "continue pending",
		windows.SERVICE_PAUSE_PENDING:    "pause pending",
		windows.SERVICE_PAUSED:           "paused",
		windows.SERVICE_RUNNING:          "running",
		windows.SERVICE_START_PENDING:    "start pending",
		windows.SERVICE_STOP_PENDING:     "stop pending",
		windows.SERVICE_STOPPED:          "stopped",
	}

	// https://learn.microsoft.com/en-us/dotnet/api/system.serviceprocess.servicestartmode
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

func (c *collector) collectWMI(ch chan<- prometheus.Metric) error {
	var dst []Win32_Service
	q := wmi.QueryAllWhere(&dst, c.queryWhereClause, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}
	for _, service := range dst {
		pid := fmt.Sprintf("%d", uint64(service.ProcessId))

		var stateCode uint
		for code, state := range apiStateValues {
			if state == strings.ToLower(service.State) {
				stateCode = code
				ch <- prometheus.MustNewConstMetric(
					c.State,
					prometheus.GaugeValue,
					float64(code),
					strings.ToLower(service.Name),
					service.DisplayName,
				)
			}
		}

		var startCode uint32
		for code, startModeLower := range apiStartModeValues {
			if startModeLower == strings.ToLower(service.StartMode) {
				startCode = code
				ch <- prometheus.MustNewConstMetric(
					c.StartMode,
					prometheus.GaugeValue,
					float64(code),
					strings.ToLower(service.Name),
					service.DisplayName,
				)
			}
		}

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
			fmt.Sprint(startCode),
			fmt.Sprint(stateCode),
		)
	}
	return nil
}

func (c *collector) collectAPI(ch chan<- prometheus.Metric) error {
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
		(func() {
			// Get UTF16 service name.
			serviceName, err := syscall.UTF16PtrFromString(service)
			if err != nil {
				_ = level.Warn(c.logger).Log("msg", fmt.Sprintf("Service %s get name error:  %#v", service, err))
				return
			}

			// Open connection for service handler.
			serviceHandle, err := windows.OpenService(svcmgrConnection.Handle, serviceName, windows.GENERIC_READ)
			if err != nil {
				_ = level.Warn(c.logger).Log("msg", fmt.Sprintf("Open service %s error:  %#v", service, err))
				return
			}

			// Create handle for each service.
			serviceManager := &mgr.Service{Name: service, Handle: serviceHandle}
			defer serviceManager.Close()

			// Get Service Configuration.
			serviceConfig, err := serviceManager.Config()
			if err != nil {
				_ = level.Warn(c.logger).Log("msg", fmt.Sprintf("Get service %s config error:  %#v", service, err))
				return
			}

			// Get Service Current Status.
			serviceStatus, err := serviceManager.Query()
			if err != nil {
				_ = level.Warn(c.logger).Log("msg", fmt.Sprintf("Get service %s status error:  %#v", service, err))
				return
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
				fmt.Sprint(serviceConfig.StartType),
				fmt.Sprint(serviceStatus.State),
			)

			ch <- prometheus.MustNewConstMetric(
				c.State,
				prometheus.GaugeValue,
				float64(serviceStatus.State),
				strings.ToLower(service),
				serviceConfig.DisplayName,
			)

			ch <- prometheus.MustNewConstMetric(
				c.StartMode,
				prometheus.GaugeValue,
				float64(serviceConfig.StartType),
				strings.ToLower(service),
				serviceConfig.DisplayName,
			)
		})()
	}
	return nil
}

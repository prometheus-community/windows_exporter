//go:build windows

package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/utils"
	"github.com/prometheus-community/windows_exporter/pkg/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/mgr"
)

const Name = "service"

type Config struct {
	ServiceWhereClause string `yaml:"service_where_clause"`
	UseAPI             bool   `yaml:"use_api"`
	V2                 bool   `yaml:"v2"`
}

var ConfigDefaults = Config{
	ServiceWhereClause: "",
	UseAPI:             false,
	V2:                 false,
}

// A Collector is a Prometheus Collector for WMI Win32_Service metrics
type Collector struct {
	logger log.Logger

	serviceWhereClause *string
	useAPI             *bool
	v2                 *bool

	Information *prometheus.Desc
	State       *prometheus.Desc
	StartMode   *prometheus.Desc
	Status      *prometheus.Desc
	StateV2     *prometheus.Desc
}

func New(logger log.Logger, config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		serviceWhereClause: &config.ServiceWhereClause,
		useAPI:             &config.UseAPI,
	}
	c.SetLogger(logger)

	return c
}

func NewWithFlags(app *kingpin.Application) *Collector {
	return &Collector{
		serviceWhereClause: app.Flag(
			"collector.service.services-where",
			"WQL 'where' clause to use in WMI metrics query. Limits the response to the services you specify and reduces the size of the response.",
		).Default(ConfigDefaults.ServiceWhereClause).String(),
		useAPI: app.Flag(
			"collector.service.use-api",
			"Use API calls to collect service data instead of WMI. Flag 'collector.service.services-where' won't be effective.",
		).Default(strconv.FormatBool(ConfigDefaults.UseAPI)).Bool(),
		v2: app.Flag(
			"collector.service.v2",
			"Enable V2 service collector. This collector can services state much more efficiently, can't provide general service information.",
		).Default(strconv.FormatBool(ConfigDefaults.V2)).Bool(),
	}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *Collector) GetPerfCounter() ([]string, error) {
	return []string{}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build() error {
	if utils.IsEmpty(c.serviceWhereClause) {
		_ = level.Warn(c.logger).Log("msg", "No where-clause specified for service collector. This will generate a very large number of metrics!")
	}
	if *c.useAPI {
		_ = level.Warn(c.logger).Log("msg", "API collection is enabled.")
	}

	c.Information = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "info"),
		"A metric with a constant '1' value labeled with service information",
		[]string{"name", "display_name", "process_id", "run_as"},
		nil,
	)
	c.State = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "state"),
		"The state of the service (State)",
		[]string{"name", "state"},
		nil,
	)
	c.StartMode = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "start_mode"),
		"The start mode of the service (StartMode)",
		[]string{"name", "start_mode"},
		nil,
	)
	c.Status = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "status"),
		"The status of the service (Status)",
		[]string{"name", "status"},
		nil,
	)
	c.StateV2 = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "state"),
		"The state of the service (State)",
		[]string{"name", "display_name", "status"},
		nil,
	)

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var err error

	switch {
	case *c.useAPI:
		if err = c.collectAPI(ch); err != nil {
			_ = level.Error(c.logger).Log("msg", "failed collecting API service metrics:", "err", err)
		}
	case *c.v2:
		if err = c.collectAPIV2(ch); err != nil {
			_ = level.Error(c.logger).Log("msg", "failed collecting API service metrics:", "err", err)
		}
	default:
		if err = c.collectWMI(ch); err != nil {
			_ = level.Error(c.logger).Log("msg", "failed collecting WMI service metrics:", "err", err)
		}
	}

	return err
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
	apiStateValues = map[uint32]string{
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

func (c *Collector) collectWMI(ch chan<- prometheus.Metric) error {
	var dst []Win32_Service
	q := wmi.QueryAllWhere(&dst, *c.serviceWhereClause, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}
	for _, service := range dst {
		pid := strconv.FormatUint(uint64(service.ProcessId), 10)

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

func (c *Collector) collectAPI(ch chan<- prometheus.Metric) error {
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

			pid := strconv.FormatUint(uint64(serviceStatus.ProcessId), 10)

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
				if state == apiStateValues[uint32(serviceStatus.State)] {
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
		})()
	}
	return nil
}

func (c *Collector) collectAPIV2(ch chan<- prometheus.Metric) error {
	services, err := c.queryAllServiceStates()
	if err != nil {
		_ = level.Warn(c.logger).Log("msg", "Failed to query services", "err", err)
		return err
	}

	if services == nil {
		_ = level.Warn(c.logger).Log("msg", "No services queried")
		return nil
	}

	var isCurrentState float64

	for _, svc := range services {
		for state, stateValue := range apiStateValues {
			isCurrentState = 0.0
			if state == svc.ServiceStatusProcess.CurrentState {
				isCurrentState = 1.0
			}

			ch <- prometheus.MustNewConstMetric(
				c.StateV2,
				prometheus.GaugeValue,
				isCurrentState,
				windows.UTF16PtrToString(svc.ServiceName),
				windows.UTF16PtrToString(svc.DisplayName),
				stateValue,
			)
		}
	}

	return nil
}

// queryAllServiceStates returns all service states of the current Windows system
// This is realized by ask Service Manager directly.
//
// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.
//
// Source: https://github.com/DataDog/datadog-agent/blob/afbd8b6c87939c92610c654cb07fdfd439e4fb27/pkg/util/winutil/scmmonitor.go#L61-L96
func (c *Collector) queryAllServiceStates() ([]windows.ENUM_SERVICE_STATUS_PROCESS, error) {
	// EnumServiceStatusEx requires only SC_MANAGER_ENUM_SERVICE.
	h, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_ENUMERATE_SERVICE)
	if err != nil {
		return nil, fmt.Errorf("failed to open scm: %w", err)
	}

	m := &mgr.Mgr{Handle: h}
	defer func() {
		if err := m.Disconnect(); err != nil {
			_ = level.Warn(c.logger).Log("msg", "Failed to disconnect from scm", "err", err)
		}
	}()

	var bytesNeeded, servicesReturned uint32
	var buf []byte
	for {
		var p *byte
		if len(buf) > 0 {
			p = &buf[0]
		}
		err = windows.EnumServicesStatusEx(m.Handle, windows.SC_ENUM_PROCESS_INFO,
			windows.SERVICE_WIN32, windows.SERVICE_STATE_ALL,
			p, uint32(len(buf)), &bytesNeeded, &servicesReturned, nil, nil)
		if err == nil {
			break
		}
		if !errors.Is(err, windows.ERROR_MORE_DATA) {
			return nil, fmt.Errorf("failed to enum services %w", err)
		}
		if bytesNeeded <= uint32(len(buf)) {
			return nil, err
		}
		buf = make([]byte, bytesNeeded)
	}

	if servicesReturned == 0 {
		return nil, nil
	}

	services := unsafe.Slice((*windows.ENUM_SERVICE_STATUS_PROCESS)(unsafe.Pointer(&buf[0])), servicesReturned)

	return services, nil
}

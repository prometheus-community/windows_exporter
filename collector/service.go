//go:build windows
// +build windows

package collector

import (
	"bytes"
	"fmt"
	"strings"
	"syscall"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/mgr"
)

const (
	FlagServiceWhereClause = "collector.service.services-where"
	FlagServiceList        = "collector.service.services-list"
	FlagServiceUseAPI      = "collector.service.use-api"
)

type ServiceDef struct {
	Name         string
	CustomLabels map[string]string
}

var (
	serviceWhereClause *string
	serviceList        *string
	useAPI             *bool
	services           = make(map[string]*ServiceDef)
	services_labels    = make([]string, 0)
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
	serviceList = app.Flag(
		FlagServiceList,
		"comma separated list of service name used to build WQL 'where' clause to use in WMI metrics query. Limits the response to the services you specify and reduces the size of the response.",
	).Default("").String()
}

// Build service list for name
func expandServiceWhere(services string) string {

	separated := strings.Split(services, ",")
	unique := map[string]bool{}
	for _, s := range separated {
		s = strings.TrimSpace(s)
		if s != "" {
			unique[s] = true
		}
	}
	var b bytes.Buffer

	i := 0
	for s := range unique {
		i += 1
		b.WriteString("Name='")
		b.WriteString(s)
		b.WriteString("'")
		if i < len(unique) {
			b.WriteString(" or ")
		}

	}
	return b.String()
}

func ServiceBuildHook() map[string]config.CfgHook {
	config_hooks := &config.CfgHook{
		ConfigAttrs: []string{"collector", "service", "services"},
		Hook:        ServiceBuildMap,
	}
	entry := make(map[string]config.CfgHook)
	entry["services-list"] = *config_hooks
	return entry
}

func ServiceBuildMap(logger log.Logger, data interface{}) map[string]string {
	ret := make(map[string]string)
	switch typed := data.(type) {
	case map[interface{}]interface{}:
		ret = flatten(data)

	// form is a dict of service's name maybe with custom labels
	case map[string]interface{}:
		for name, raw_labels := range typed {
			labels := flatten(raw_labels)
			service := &ServiceDef{
				Name:         name,
				CustomLabels: labels,
			}
			services[strings.ToLower(service.Name)] = service
		}

	// form is a list of services' name
	case []interface{}:
		for _, fv := range typed {
			var name string
			tmp := flatten(fv)
			// tmp is a pseudo map with a key set and no value: collect first key
			for k := range tmp {
				name = k
				break
			}
			service := &ServiceDef{
				Name:         name,
				CustomLabels: nil,
			}
			services[strings.ToLower(service.Name)] = service

			// serviceHash[fmt.Sprint(idx)] = flatten(fv)
		}

	default:
		ret["unknown_parameter"] = "1"
		// serviceHash[fmt.Sprint(typed)] = 1
	}

	// build a list of all custom labels for each service
	exists := make(map[string]bool)
	for _, svc := range services {
		// fill the exists map with all labels' names
		for name := range svc.CustomLabels {
			exists[name] = true
		}
	}
	// check custom labels: each name must be present for each service
	for _, svc := range services {
		for name := range svc.CustomLabels {
			if _, ok := exists[name]; !ok {
				_ = level.Warn(logger).Log("errmsg", "label: %s not present for service '%s'", name, svc.Name)
				exists[name] = false
			}
		}
	}
	services_labels = make([]string, 0, len(exists))
	for name := range exists {
		services_labels = append(services_labels, name)
	}
	return ret
}

// convert map value to something that can be use as labels of service def
func flatten(data interface{}) map[string]string {
	ret := make(map[string]string)
	switch typed := data.(type) {
	// value is a map of something that we hope are strings
	case map[string]interface{}:
		for fk, fv := range typed {
			ret[fk] = fmt.Sprint(fv)
		}
	// value is a string or nothing (list)
	default:
		if typed != nil {
			ret[fmt.Sprint(typed)] = "1"
		} else {
			ret = nil
		}
	}
	return ret
}

// newserviceCollector ...
func newserviceCollector(logger log.Logger) (Collector, error) {
	const subsystem = "service"
	logger = log.With(logger, "collector", subsystem)

	if *serviceWhereClause == "" && *serviceList == "" && len(services) == 0 {
		_ = level.Warn(logger).Log("msg", "No where-clause specified for service collector. This will generate a very large number of metrics!")
	}

	if *serviceWhereClause == "" && *serviceList != "" {
		*serviceWhereClause = expandServiceWhere(*serviceList)
	} else if len(services) > 0 {
		servList := make([]string, 0, len(services))
		for _, svc := range services {
			if svc != nil {
				servList = append(servList, svc.Name)
			}
		}
		svc := strings.Join(servList, ",")
		*serviceWhereClause = expandServiceWhere(svc)
	}

	if *serviceWhereClause != "" {
		_ = level.Debug(logger).Log("msg", fmt.Sprintf("serviceWhereClause='%s'", *serviceWhereClause))
	}

	if *useAPI {
		_ = level.Warn(logger).Log("msg", "API collection is enabled.")
	}

	var var_labels [4][]string
	var_labels[0] = []string{"name", "display_name", "process_id", "run_as"}
	var_labels[1] = []string{"name", "state"}
	var_labels[2] = []string{"name", "start_mode"}
	var_labels[3] = []string{"name", "status"}
	if len(services_labels) > 0 {
		for _, label := range services_labels {
			for idx, metric_label := range var_labels {
				var_labels[idx] = append(metric_label, label)
			}
		}
	}

	return &serviceCollector{
		logger: logger,

		Information: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "info"),
			"A metric with a constant '1' value labeled with service information",
			var_labels[0],
			nil,
		),
		State: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "state"),
			"The state of the service (State)",
			var_labels[1],
			nil,
		),
		StartMode: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "start_mode"),
			"The start mode of the service (StartMode)",
			var_labels[2],
			nil,
		),
		Status: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "status"),
			"The status of the service (Status)",
			var_labels[3],
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
		// build the custom labels with their values for the service
		var custom_labels []string
		service_name := strings.ToLower(service.Name)
		if len(services_labels) > 0 {
			custom_labels = make([]string, len(services_labels))
			if svc, ok := services[service_name]; ok {
				// we find the service name in service definition list
				for idx, label := range services_labels {
					if val, tst := svc.CustomLabels[label]; tst {
						custom_labels[idx] = val
					}
				}
			} else {
				// we don't find the service name !?!? not possible but...
				for idx := range services_labels {
					custom_labels[idx] = ""
				}
			}
		}

		// Information metric
		labels := make([]string, 4)
		// service name
		labels[0] = service_name
		// service display name
		labels[1] = service.DisplayName
		// service pid
		labels[2] = fmt.Sprintf("%d", uint64(service.ProcessId))
		// service runAs
		if service.StartName != nil {
			labels[3] = *service.StartName
		} else {
			labels[3] = ""
		}
		if len(custom_labels) > 0 {
			labels = append(labels, custom_labels...)
		}

		ch <- prometheus.MustNewConstMetric(
			c.Information,
			prometheus.GaugeValue,
			1.0, labels...,
		)

		// State metric
		labels = make([]string, 2)
		// service name
		labels[0] = service_name
		labels[1] = ""
		if len(custom_labels) > 0 {
			labels = append(labels, custom_labels...)
		}

		for _, state := range allStates {
			isCurrentState := 0.0
			if state == strings.ToLower(service.State) {
				isCurrentState = 1.0
			}
			labels[1] = state
			ch <- prometheus.MustNewConstMetric(
				c.State,
				prometheus.GaugeValue,
				isCurrentState,
				labels...,
			)
		}

		// StartMode metric
		for _, startMode := range allStartModes {
			isCurrentStartMode := 0.0
			if startMode == strings.ToLower(service.StartMode) {
				isCurrentStartMode = 1.0
			}
			labels[1] = startMode
			ch <- prometheus.MustNewConstMetric(
				c.StartMode,
				prometheus.GaugeValue,
				isCurrentStartMode,
				labels...,
			)
		}

		// service status metric
		for _, status := range allStatuses {
			isCurrentStatus := 0.0
			if status == strings.ToLower(service.Status) {
				isCurrentStatus = 1.0
			}
			labels[1] = status
			ch <- prometheus.MustNewConstMetric(
				c.Status,
				prometheus.GaugeValue,
				isCurrentStatus,
				labels...,
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

		// build the custom labels with their values for the service
		var custom_labels []string
		service_name := strings.ToLower(service)
		if len(services_labels) > 0 {
			custom_labels = make([]string, len(services_labels))
			if svc, ok := services[service_name]; ok {
				// we find the service name in service definition list
				for idx, label := range services_labels {
					if val, tst := svc.CustomLabels[label]; tst {
						custom_labels[idx] = val
					}
				}
			} else {
				// we don't find the service name !?!? not possible but...
				for idx := range services_labels {
					custom_labels[idx] = ""
				}
			}
		}

		// Information metric
		labels := make([]string, 4)
		// service name
		labels[0] = service_name
		// service display name
		labels[1] = serviceConfig.DisplayName
		// service pid
		labels[2] = fmt.Sprintf("%d", uint64(serviceStatus.ProcessId))
		// service runAs
		labels[3] = serviceConfig.ServiceStartName

		if len(custom_labels) > 0 {
			labels = append(labels, custom_labels...)
		}
		ch <- prometheus.MustNewConstMetric(
			c.Information,
			prometheus.GaugeValue,
			1.0,
			labels...,
		)

		// State metric
		labels = make([]string, 2)
		// service name
		labels[0] = service_name
		labels[1] = ""
		if len(custom_labels) > 0 {
			labels = append(labels, custom_labels...)
		}

		for _, state := range apiStateValues {
			isCurrentState := 0.0
			if state == apiStateValues[uint(serviceStatus.State)] {
				isCurrentState = 1.0
			}
			labels[1] = state
			ch <- prometheus.MustNewConstMetric(
				c.State,
				prometheus.GaugeValue,
				isCurrentState,
				labels...,
			)
		}

		// StartMode metric
		for _, startMode := range apiStartModeValues {
			isCurrentStartMode := 0.0
			if startMode == apiStartModeValues[serviceConfig.StartType] {
				isCurrentStartMode = 1.0
			}
			labels[1] = startMode
			ch <- prometheus.MustNewConstMetric(
				c.StartMode,
				prometheus.GaugeValue,
				isCurrentStartMode,
				labels...,
			)
		}
	}
	return nil
}

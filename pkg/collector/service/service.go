//go:build windows

package service

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"unsafe"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/mgr"
)

const Name = "service"

type Config struct {
	ServiceInclude *regexp.Regexp `yaml:"service_include"`
	ServiceExclude *regexp.Regexp `yaml:"service_exclude"`
}

var ConfigDefaults = Config{
	ServiceInclude: types.RegExpAny,
	ServiceExclude: types.RegExpEmpty,
}

// A Collector is a Prometheus Collector for service metrics.
type Collector struct {
	config Config

	state     *prometheus.Desc
	processID *prometheus.Desc
	info      *prometheus.Desc
	startMode *prometheus.Desc

	serviceManagerHandle *mgr.Mgr
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	if config.ServiceExclude == nil {
		config.ServiceExclude = ConfigDefaults.ServiceExclude
	}

	if config.ServiceInclude == nil {
		config.ServiceInclude = ConfigDefaults.ServiceInclude
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(app *kingpin.Application) *Collector {
	c := &Collector{
		config: ConfigDefaults,
	}

	var serviceExclude, serviceInclude string

	app.Flag(
		"collector.service.exclude",
		"Regexp of service to exclude. Service name (not the display name!) must both match include and not match exclude to be included.",
	).Default(c.config.ServiceExclude.String()).StringVar(&serviceExclude)

	app.Flag(
		"collector.service.include",
		"Regexp of service to include. Process name (not the display name!) must both match include and not match exclude to be included.",
	).Default(c.config.ServiceInclude.String()).StringVar(&serviceInclude)

	app.Action(func(*kingpin.ParseContext) error {
		var err error

		c.config.ServiceExclude, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", serviceExclude))
		if err != nil {
			return fmt.Errorf("collector.process.exclude: %w", err)
		}

		c.config.ServiceInclude, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", serviceInclude))
		if err != nil {
			return fmt.Errorf("collector.process.include: %w", err)
		}

		return nil
	})

	return c
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) GetPerfCounter(_ *slog.Logger) ([]string, error) {
	return []string{}, nil
}

func (c *Collector) Build(logger *slog.Logger, _ *wmi.Client) error {
	logger = logger.With(slog.String("collector", Name))

	if c.config.ServiceInclude.String() == "^(?:.*)$" && c.config.ServiceExclude.String() == "^(?:)$" {
		logger.Warn("No filters specified for service collector. This will generate a very large number of metrics!")
	}

	c.info = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "info"),
		"A metric with a constant '1' value labeled with service information",
		[]string{"name", "display_name", "run_as", "path_name"},
		nil,
	)
	c.state = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "state"),
		"The state of the service (State)",
		[]string{"name", "status"},
		nil,
	)
	c.startMode = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "start_mode"),
		"The start mode of the service (StartMode)",
		[]string{"name", "start_mode"},
		nil,
	)
	c.processID = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "process"),
		"Process of started service. The value is the creation time of the process as a unix timestamp.",
		[]string{"name", "process_id"},
		nil,
	)

	// EnumServiceStatusEx requires only SC_MANAGER_ENUM_SERVICE.
	handle, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_ENUMERATE_SERVICE)
	if err != nil {
		return fmt.Errorf("failed to open scm: %w", err)
	}

	c.serviceManagerHandle = &mgr.Mgr{Handle: handle}

	return nil
}

func (c *Collector) Close(logger *slog.Logger) error {
	if err := c.serviceManagerHandle.Disconnect(); err != nil {
		logger.Warn("Failed to disconnect from scm",
			slog.Any("err", err),
		)
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	logger = logger.With(slog.String("collector", Name))

	if err := c.collect(logger, ch); err != nil {
		logger.Error("failed collecting API service metrics:",
			slog.Any("err", err),
		)

		return fmt.Errorf("failed collecting API service metrics: %w", err)
	}

	return nil
}

func (c *Collector) collect(logger *slog.Logger, ch chan<- prometheus.Metric) error {
	services, err := c.queryAllServices()
	if err != nil {
		logger.Warn("Failed to query services",
			slog.Any("err", err),
		)

		return err
	}

	if services == nil {
		logger.Warn("No services queried")

		return nil
	}

	// Iterate through the Services List.
	for _, service := range services {
		serviceName := windows.UTF16PtrToString(service.ServiceName)
		if c.config.ServiceExclude.MatchString(serviceName) ||
			!c.config.ServiceInclude.MatchString(serviceName) {
			continue
		}

		if err := c.collectService(ch, logger, service); err != nil {
			logger.Warn("failed collecting service info",
				slog.Any("err", err),
				slog.String("service", windows.UTF16PtrToString(service.ServiceName)),
			)
		}
	}

	return nil
}

var apiStateValues = map[uint32]string{
	windows.SERVICE_CONTINUE_PENDING: "continue pending",
	windows.SERVICE_PAUSE_PENDING:    "pause pending",
	windows.SERVICE_PAUSED:           "paused",
	windows.SERVICE_RUNNING:          "running",
	windows.SERVICE_START_PENDING:    "start pending",
	windows.SERVICE_STOP_PENDING:     "stop pending",
	windows.SERVICE_STOPPED:          "stopped",
}

var apiStartModeValues = map[uint32]string{
	windows.SERVICE_AUTO_START:   "auto",
	windows.SERVICE_BOOT_START:   "boot",
	windows.SERVICE_DEMAND_START: "manual",
	windows.SERVICE_DISABLED:     "disabled",
	windows.SERVICE_SYSTEM_START: "system",
}

func (c *Collector) collectService(ch chan<- prometheus.Metric, logger *slog.Logger, service windows.ENUM_SERVICE_STATUS_PROCESS) error {
	// Open connection for service handler.
	serviceHandle, err := windows.OpenService(c.serviceManagerHandle.Handle, service.ServiceName, windows.SERVICE_QUERY_CONFIG)
	if err != nil {
		return fmt.Errorf("failed to open service: %w", err)
	}

	serviceNameString := windows.UTF16PtrToString(service.ServiceName)

	// Create handle for each service.
	serviceManager := &mgr.Service{Name: serviceNameString, Handle: serviceHandle}
	defer func(serviceManager *mgr.Service) {
		if err := serviceManager.Close(); err != nil {
			logger.Warn("failed to close service handle",
				slog.Any("err", err),
				slog.String("service", serviceNameString),
			)
		}
	}(serviceManager)

	// Get Service Configuration.
	serviceConfig, err := serviceManager.Config()
	if err != nil {
		if !errors.Is(err, windows.ERROR_FILE_NOT_FOUND) && !errors.Is(err, windows.ERROR_MUI_FILE_NOT_FOUND) {
			return fmt.Errorf("failed to get service configuration: %w", err)
		}

		logger.Debug("failed collecting service",
			slog.Any("err", err),
			slog.String("service", serviceNameString),
		)
	}

	ch <- prometheus.MustNewConstMetric(
		c.info,
		prometheus.GaugeValue,
		1.0,
		serviceNameString,
		serviceConfig.DisplayName,
		serviceConfig.ServiceStartName,
		serviceConfig.BinaryPathName,
	)

	var (
		isCurrentStartMode float64
		isCurrentState     float64
	)

	for _, startMode := range apiStartModeValues {
		isCurrentStartMode = 0.0
		if startMode == apiStartModeValues[serviceConfig.StartType] {
			isCurrentStartMode = 1.0
		}
		ch <- prometheus.MustNewConstMetric(
			c.startMode,
			prometheus.GaugeValue,
			isCurrentStartMode,
			serviceNameString,
			startMode,
		)
	}

	for state, stateValue := range apiStateValues {
		isCurrentState = 0.0
		if state == service.ServiceStatusProcess.CurrentState {
			isCurrentState = 1.0
		}

		ch <- prometheus.MustNewConstMetric(
			c.state,
			prometheus.GaugeValue,
			isCurrentState,
			serviceNameString,
			stateValue,
		)
	}

	processID := strconv.FormatUint(uint64(service.ServiceStatusProcess.ProcessId), 10)

	if processID != "0" { //nolint: nestif
		processStartTime, err := getProcessStartTime(logger, service.ServiceStatusProcess.ProcessId)
		if err != nil {
			if errors.Is(err, windows.ERROR_ACCESS_DENIED) {
				logger.Debug("failed to get process start time",
					slog.String("service", serviceNameString),
					slog.Any("err", err),
				)
			} else {
				logger.Warn("failed to get process start time",
					slog.String("service", serviceNameString),
					slog.Any("err", err),
				)
			}
		} else {
			ch <- prometheus.MustNewConstMetric(
				c.processID,
				prometheus.GaugeValue,
				float64(processStartTime/1_000_000_000),
				serviceNameString,
				processID,
			)
		}
	}

	return nil
}

// queryAllServices returns all service states of the current Windows system
// This is realized by ask Service Manager directly.
func (c *Collector) queryAllServices() ([]windows.ENUM_SERVICE_STATUS_PROCESS, error) {
	var (
		bytesNeeded      uint32
		servicesReturned uint32
		resumeHandle     uint32
	)

	if err := windows.EnumServicesStatusEx(
		c.serviceManagerHandle.Handle,
		windows.SC_STATUS_PROCESS_INFO,
		windows.SERVICE_WIN32,
		windows.SERVICE_STATE_ALL,
		nil,
		0,
		&bytesNeeded,
		&servicesReturned,
		&resumeHandle,
		nil,
	); !errors.Is(err, windows.ERROR_MORE_DATA) {
		return nil, fmt.Errorf("could not fetch buffer size for EnumServicesStatusEx: %w", err)
	}

	buf := make([]byte, bytesNeeded)
	if err := windows.EnumServicesStatusEx(
		c.serviceManagerHandle.Handle,
		windows.SC_STATUS_PROCESS_INFO,
		windows.SERVICE_WIN32,
		windows.SERVICE_STATE_ALL,
		&buf[0],
		bytesNeeded,
		&bytesNeeded,
		&servicesReturned,
		&resumeHandle,
		nil,
	); err != nil {
		return nil, fmt.Errorf("could not query windows service list: %w", err)
	}

	if servicesReturned == 0 {
		return nil, nil
	}

	services := unsafe.Slice((*windows.ENUM_SERVICE_STATUS_PROCESS)(unsafe.Pointer(&buf[0])), int(servicesReturned))

	return services, nil
}

func getProcessStartTime(logger *slog.Logger, pid uint32) (uint64, error) {
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return 0, fmt.Errorf("failed to open process %w", err)
	}

	defer func(handle windows.Handle) {
		err := windows.CloseHandle(handle)
		if err != nil {
			logger.Warn("failed to close process handle",
				slog.Any("err", err),
			)
		}
	}(handle)

	var creation windows.Filetime

	var exit windows.Filetime

	var krn windows.Filetime

	var user windows.Filetime

	err = windows.GetProcessTimes(handle, &creation, &exit, &krn, &user)
	if err != nil {
		return 0, fmt.Errorf("failed to get process times %w", err)
	}

	return uint64(creation.Nanoseconds()), nil
}

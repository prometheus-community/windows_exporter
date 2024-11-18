//go:build windows

package service

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"sync"
	"unsafe"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
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

	logger *slog.Logger

	apiStateValues     map[uint32]string
	apiStartModeValues map[uint32]string

	state     *prometheus.Desc
	processID *prometheus.Desc
	info      *prometheus.Desc
	startMode *prometheus.Desc

	// serviceConfigPoolBytes is a pool of byte slices used to avoid allocations
	// ref: https://victoriametrics.com/blog/go-sync-pool/
	serviceConfigPoolBytes sync.Pool

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
	).Default("").StringVar(&serviceExclude)

	app.Flag(
		"collector.service.include",
		"Regexp of service to include. Process name (not the display name!) must both match include and not match exclude to be included.",
	).Default(".+").StringVar(&serviceInclude)

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

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	c.logger = logger.With(slog.String("collector", Name))

	if c.config.ServiceInclude.String() == "^(?:.*)$" && c.config.ServiceExclude.String() == "^(?:)$" {
		c.logger.Warn("No filters specified for service collector. This will generate a very large number of metrics!")
	}

	c.serviceConfigPoolBytes = sync.Pool{
		New: func() any {
			return new([]byte)
		},
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
		[]string{"name", "state"},
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

	c.apiStateValues = map[uint32]string{
		windows.SERVICE_CONTINUE_PENDING: "continue pending",
		windows.SERVICE_PAUSE_PENDING:    "pause pending",
		windows.SERVICE_PAUSED:           "paused",
		windows.SERVICE_RUNNING:          "running",
		windows.SERVICE_START_PENDING:    "start pending",
		windows.SERVICE_STOP_PENDING:     "stop pending",
		windows.SERVICE_STOPPED:          "stopped",
	}

	c.apiStartModeValues = map[uint32]string{
		windows.SERVICE_AUTO_START:   "auto",
		windows.SERVICE_BOOT_START:   "boot",
		windows.SERVICE_DEMAND_START: "manual",
		windows.SERVICE_DISABLED:     "disabled",
		windows.SERVICE_SYSTEM_START: "system",
	}

	// EnumServiceStatusEx requires only SC_MANAGER_ENUM_SERVICE.
	handle, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_ENUMERATE_SERVICE)
	if err != nil {
		return fmt.Errorf("failed to open scm: %w", err)
	}

	c.serviceManagerHandle = &mgr.Mgr{Handle: handle}

	return nil
}

func (c *Collector) Close() error {
	if err := c.serviceManagerHandle.Disconnect(); err != nil {
		c.logger.Warn("Failed to disconnect from scm",
			slog.Any("err", err),
		)
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	services, err := c.queryAllServices()
	if err != nil {
		return fmt.Errorf("failed to query services: %w", err)
	}

	if len(services) == 0 {
		c.logger.Warn("No services queried")

		return nil
	}

	servicesCh := make(chan windows.ENUM_SERVICE_STATUS_PROCESS, len(services))
	wg := sync.WaitGroup{}
	wg.Add(len(services))

	for range 4 {
		go func(ch chan<- prometheus.Metric, wg *sync.WaitGroup) {
			for service := range servicesCh {
				c.collectWorker(ch, service)
				wg.Done()
			}
		}(ch, &wg)
	}

	for _, service := range services {
		servicesCh <- service
	}

	close(servicesCh)

	wg.Wait()

	return nil
}

func (c *Collector) collectWorker(ch chan<- prometheus.Metric, service windows.ENUM_SERVICE_STATUS_PROCESS) {
	serviceName := windows.UTF16PtrToString(service.ServiceName)

	if c.config.ServiceExclude.MatchString(serviceName) || !c.config.ServiceInclude.MatchString(serviceName) {
		return
	}

	if err := c.collectService(ch, serviceName, service); err != nil {
		c.logger.Warn("failed collecting service info",
			slog.Any("err", err),
			slog.String("service", serviceName),
		)
	}
}

func (c *Collector) collectService(ch chan<- prometheus.Metric, serviceName string, service windows.ENUM_SERVICE_STATUS_PROCESS) error {
	// Open connection for service handler.
	serviceHandle, err := windows.OpenService(c.serviceManagerHandle.Handle, service.ServiceName, windows.SERVICE_QUERY_CONFIG)
	if err != nil {
		return fmt.Errorf("failed to open service: %w", err)
	}

	// Create handle for each service.
	serviceManager := &mgr.Service{Name: serviceName, Handle: serviceHandle}
	defer func(serviceManager *mgr.Service) {
		if err := serviceManager.Close(); err != nil {
			c.logger.Warn("failed to close service handle",
				slog.Any("err", err),
				slog.String("service", serviceName),
			)
		}
	}(serviceManager)

	// Get Service Configuration.
	serviceConfig, err := c.getServiceConfig(serviceManager)
	if err != nil {
		if !errors.Is(err, windows.ERROR_FILE_NOT_FOUND) && !errors.Is(err, windows.ERROR_MUI_FILE_NOT_FOUND) {
			return fmt.Errorf("failed to get service configuration: %w", err)
		}

		c.logger.Debug("failed collecting service config",
			slog.Any("err", err),
			slog.String("service", serviceName),
		)
	}

	ch <- prometheus.MustNewConstMetric(
		c.info,
		prometheus.GaugeValue,
		1.0,
		serviceName,
		serviceConfig.DisplayName,
		serviceConfig.ServiceStartName,
		serviceConfig.BinaryPathName,
	)

	var (
		isCurrentStartMode float64
		isCurrentState     float64
	)

	for _, startMode := range c.apiStartModeValues {
		isCurrentStartMode = 0.0
		if startMode == c.apiStartModeValues[serviceConfig.StartType] {
			isCurrentStartMode = 1.0
		}
		ch <- prometheus.MustNewConstMetric(
			c.startMode,
			prometheus.GaugeValue,
			isCurrentStartMode,
			serviceName,
			startMode,
		)
	}

	for state, stateValue := range c.apiStateValues {
		isCurrentState = 0.0
		if state == service.ServiceStatusProcess.CurrentState {
			isCurrentState = 1.0
		}

		ch <- prometheus.MustNewConstMetric(
			c.state,
			prometheus.GaugeValue,
			isCurrentState,
			serviceName,
			stateValue,
		)
	}

	if service.ServiceStatusProcess.ProcessId == 0 {
		return nil
	}

	processID := strconv.FormatUint(uint64(service.ServiceStatusProcess.ProcessId), 10)

	processStartTime, err := c.getProcessStartTime(service.ServiceStatusProcess.ProcessId)
	if err == nil {
		ch <- prometheus.MustNewConstMetric(
			c.processID,
			prometheus.GaugeValue,
			float64(processStartTime/1_000_000_000),
			serviceName,
			processID,
		)

		return nil
	}

	if errors.Is(err, windows.ERROR_ACCESS_DENIED) {
		c.logger.Debug("failed to get process start time",
			slog.String("service", serviceName),
			slog.Any("err", err),
		)
	} else {
		c.logger.Warn("failed to get process start time",
			slog.String("service", serviceName),
			slog.Any("err", err),
		)
	}

	return nil
}

// queryAllServices returns all service states of the current Windows system
// This is realized by ask Service Manager directly.
func (c *Collector) queryAllServices() ([]windows.ENUM_SERVICE_STATUS_PROCESS, error) {
	var (
		bytesNeeded      uint32
		servicesReturned uint32
		err              error
	)

	buf := make([]byte, 1024*100)

	for {
		err = windows.EnumServicesStatusEx(
			c.serviceManagerHandle.Handle,
			windows.SC_STATUS_PROCESS_INFO,
			windows.SERVICE_WIN32,
			windows.SERVICE_STATE_ALL,
			&buf[0],
			uint32(len(buf)),
			&bytesNeeded,
			&servicesReturned,
			nil,
			nil,
		)

		if err == nil {
			break
		}

		if !errors.Is(err, windows.ERROR_MORE_DATA) {
			return nil, err
		}

		if bytesNeeded <= uint32(len(buf)) {
			return nil, err
		}

		buf = make([]byte, bytesNeeded)
	}

	if servicesReturned == 0 {
		return []windows.ENUM_SERVICE_STATUS_PROCESS{}, nil
	}

	services := unsafe.Slice((*windows.ENUM_SERVICE_STATUS_PROCESS)(unsafe.Pointer(&buf[0])), int(servicesReturned))

	return services, nil
}

func (c *Collector) getProcessStartTime(pid uint32) (uint64, error) {
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return 0, fmt.Errorf("failed to open process %w", err)
	}

	defer func(handle windows.Handle) {
		err := windows.CloseHandle(handle)
		if err != nil {
			c.logger.Warn("failed to close process handle",
				slog.Any("err", err),
			)
		}
	}(handle)

	var (
		creation windows.Filetime
		exit     windows.Filetime
		krn      windows.Filetime
		user     windows.Filetime
	)

	err = windows.GetProcessTimes(handle, &creation, &exit, &krn, &user)
	if err != nil {
		return 0, fmt.Errorf("failed to get process times %w", err)
	}

	return uint64(creation.Nanoseconds()), nil
}

// getServiceConfig is an optimized variant of [mgr.Service] that only
// retrieves the necessary information.
func (c *Collector) getServiceConfig(service *mgr.Service) (mgr.Config, error) {
	var serviceConfig *windows.QUERY_SERVICE_CONFIG

	bytesNeeded := uint32(1024)

	buf, ok := c.serviceConfigPoolBytes.Get().(*[]byte)
	if !ok || len(*buf) == 0 {
		*buf = make([]byte, bytesNeeded)
	} else {
		bytesNeeded = uint32(cap(*buf))
	}

	for {
		serviceConfig = (*windows.QUERY_SERVICE_CONFIG)(unsafe.Pointer(&(*buf)[0]))

		err := windows.QueryServiceConfig(service.Handle, serviceConfig, bytesNeeded, &bytesNeeded)
		if err == nil {
			break
		}

		if !errors.Is(err, windows.ERROR_INSUFFICIENT_BUFFER) && !errors.Is(err, windows.ERROR_MORE_DATA) {
			return mgr.Config{}, err
		}

		if bytesNeeded <= uint32(len(*buf)) {
			return mgr.Config{}, fmt.Errorf("win32 reports buffer too small (%d), but buffer is large enough (%d): %w", uint32(cap(*buf)), bytesNeeded, err)
		}

		*buf = make([]byte, bytesNeeded)
	}

	c.serviceConfigPoolBytes.Put(buf)

	return mgr.Config{
		BinaryPathName:   windows.UTF16PtrToString(serviceConfig.BinaryPathName),
		DisplayName:      windows.UTF16PtrToString(serviceConfig.DisplayName),
		StartType:        serviceConfig.StartType,
		ServiceStartName: windows.UTF16PtrToString(serviceConfig.ServiceStartName),
	}, nil
}

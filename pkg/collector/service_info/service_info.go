//go:build windows

package service_info

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"syscall"
	"unsafe"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/mgr"
)

const Name = "service_info"

type Config struct{}

var ConfigDefaults = Config{}

var apiStartModeValues = map[uint32]string{
	windows.SERVICE_AUTO_START:   "auto",
	windows.SERVICE_BOOT_START:   "boot",
	windows.SERVICE_DEMAND_START: "manual",
	windows.SERVICE_DISABLED:     "disabled",
	windows.SERVICE_SYSTEM_START: "system",
}

// A Collector is a Prometheus Collector for WMI Win32_Service metrics
type Collector struct {
	logger log.Logger

	runAs     *prometheus.Desc
	startMode *prometheus.Desc

	serviceManagerHandle *mgr.Mgr
}

func New(logger log.Logger, config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{}
	c.SetLogger(logger)

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
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

func (c *Collector) Build() error {
	c.runAs = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "run_as"),
		"The start mode of the service (StartMode)",
		[]string{"name", "run_as"},
		nil,
	)
	c.startMode = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "start_mode"),
		"The start mode of the service (StartMode)",
		[]string{"name", "start_mode"},
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

func (c *Collector) Close() error {
	if err := c.serviceManagerHandle.Disconnect(); err != nil {
		_ = level.Warn(c.logger).Log("msg", "Failed to disconnect from scm", "err", err)
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var err error

	if err = c.collect(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting API service metrics", "err", err)
	}

	return err
}

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	// List All Services from the Services Manager.
	serviceList, err := c.queryAllServices()
	if err != nil {
		return err
	}

	serviceWorker := make(chan *uint16, len(serviceList))

	// Iterate through the Services List.
	for _, service := range serviceList {
		serviceWorker <- service.ServiceName
	}

	close(serviceWorker)

	wg := sync.WaitGroup{}
	wg.Add(runtime.NumCPU())

	for range runtime.NumCPU() {
		go func() {
			defer wg.Done()

			for serviceName := range serviceWorker {
				if err := c.collectServiceInfo(ch, serviceName); err != nil {
					_ = level.Warn(c.logger).Log("msg", "failed collecting service info", "err", err, "service", windows.UTF16PtrToString(serviceName))
				}
			}
		}()
	}

	wg.Wait()

	return nil
}

func (c *Collector) collectServiceInfo(ch chan<- prometheus.Metric, serviceName *uint16) error {
	// Open connection for service handler.
	serviceHandle, err := windows.OpenService(c.serviceManagerHandle.Handle, serviceName, windows.SERVICE_QUERY_CONFIG)
	if err != nil {
		return fmt.Errorf("failed to open service: %w", err)
	}

	serviceNameString := windows.UTF16PtrToString(serviceName)

	// Create handle for each service.
	serviceManager := &mgr.Service{Name: serviceNameString, Handle: serviceHandle}
	defer func(serviceManager *mgr.Service) {
		if err := serviceManager.Close(); err != nil {
			_ = level.Warn(c.logger).Log("msg", "failed to close service handle", "err", err)
		}
	}(serviceManager)

	// Get Service Configuration.
	serviceConfig, err := serviceManager.Config()
	if err != nil {
		if !errors.Is(err, windows.ERROR_FILE_NOT_FOUND) && !errors.Is(err, windows.ERROR_MUI_FILE_NOT_FOUND) {
			return fmt.Errorf("failed to get service configuration: %w", err)
		}

		_ = level.Debug(c.logger).Log("msg", "failed collecting service info", "err", err, "service", serviceName)
	}

	ch <- prometheus.MustNewConstMetric(
		c.runAs,
		prometheus.GaugeValue,
		1.0,
		serviceNameString,
		serviceConfig.ServiceStartName,
	)

	for _, startMode := range apiStartModeValues {
		isCurrentStartMode := 0.0
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

	return nil
}

// queryAllServices returns all service states of the current Windows system
// This is realized by ask Service Manager directly.
func (c *Collector) queryAllServices() ([]windows.ENUM_SERVICE_STATUS_PROCESS, error) {
	var bytesNeeded, servicesReturned uint32
	var buf []byte
	var err error
	for {
		var p *byte
		if len(buf) > 0 {
			p = &buf[0]
		}
		err = windows.EnumServicesStatusEx(c.serviceManagerHandle.Handle, windows.SC_ENUM_PROCESS_INFO,
			windows.SERVICE_WIN32, windows.SERVICE_STATE_ALL,
			p, uint32(len(buf)), &bytesNeeded, &servicesReturned, nil, nil)
		if err == nil {
			break
		}
		if !errors.Is(err, syscall.ERROR_MORE_DATA) {
			return nil, err
		}
		if bytesNeeded <= uint32(len(buf)) {
			return nil, err
		}
		buf = make([]byte, bytesNeeded)
	}

	if servicesReturned == 0 {
		return nil, nil
	}
	services := unsafe.Slice((*windows.ENUM_SERVICE_STATUS_PROCESS)(unsafe.Pointer(&buf[0])), int(servicesReturned))

	return services, nil
}

//go:build windows

package service

import (
	"errors"
	"fmt"
	"strconv"
	"unsafe"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/mgr"
)

const Name = "service"

type Config struct {
}

var ConfigDefaults = Config{}

var apiStateValues = map[uint32]string{
	windows.SERVICE_CONTINUE_PENDING: "continue pending",
	windows.SERVICE_PAUSE_PENDING:    "pause pending",
	windows.SERVICE_PAUSED:           "paused",
	windows.SERVICE_RUNNING:          "running",
	windows.SERVICE_START_PENDING:    "start pending",
	windows.SERVICE_STOP_PENDING:     "stop pending",
	windows.SERVICE_STOPPED:          "stopped",
}

// A Collector is a Prometheus Collector for WMI Win32_Service metrics.
type Collector struct {
	config Config
	logger log.Logger

	state     *prometheus.Desc
	processID *prometheus.Desc

	serviceManagerHandle *mgr.Mgr
}

func New(logger log.Logger, config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		config: *config,
	}

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
	c.state = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "state"),
		"The state of the service (State)",
		[]string{"name", "display_name", "status"},
		nil,
	)
	c.processID = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "process_id"),
		"Process ID of started service",
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
		_ = level.Error(c.logger).Log("msg", "failed collecting API service metrics:", "err", err)
	}

	return err
}

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	services, err := c.queryAllServices()
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
				c.state,
				prometheus.GaugeValue,
				isCurrentState,
				windows.UTF16PtrToString(svc.ServiceName),
				windows.UTF16PtrToString(svc.DisplayName),
				stateValue,
			)
		}

		ch <- prometheus.MustNewConstMetric(
			c.processID,
			prometheus.GaugeValue,
			1.0,
			windows.UTF16PtrToString(svc.ServiceName),
			strconv.FormatUint(uint64(svc.ServiceStatusProcess.ProcessId), 10),
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

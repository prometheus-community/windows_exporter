//go:build windows

package cpu_info

import (
	"errors"
	"log/slog"
	"strconv"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
)

const (
	Name = "cpu_info"
)

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for a few WMI metrics in Win32_Processor.
type Collector struct {
	config Config

	wmiClient *wmi.Client

	cpuInfo                   *prometheus.Desc
	cpuCoreCount              *prometheus.Desc
	cpuEnabledCoreCount       *prometheus.Desc
	cpuLogicalProcessorsCount *prometheus.Desc
	cpuThreadCount            *prometheus.Desc
	cpuL2CacheSize            *prometheus.Desc
	cpuL3CacheSize            *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) GetPerfCounter(_ *slog.Logger) ([]string, error) {
	return []string{}, nil
}

func (c *Collector) Close(_ *slog.Logger) error {
	return nil
}

func (c *Collector) Build(_ *slog.Logger, wmiClient *wmi.Client) error {
	if wmiClient == nil || wmiClient.SWbemServicesClient == nil {
		return errors.New("wmiClient or SWbemServicesClient is nil")
	}

	c.wmiClient = wmiClient
	c.cpuInfo = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, "", Name),
		"Labelled CPU information as provided by Win32_Processor",
		[]string{
			"architecture",
			"device_id",
			"description",
			"family",
			"name",
		},
		nil,
	)
	c.cpuThreadCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "thread"),
		"Number of threads per CPU",
		[]string{
			"device_id",
		},
		nil,
	)
	c.cpuCoreCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "core"),
		"Number of cores per CPU",
		[]string{
			"device_id",
		},
		nil,
	)
	c.cpuEnabledCoreCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "enabled_core"),
		"Number of enabled cores per CPU",
		[]string{
			"device_id",
		},
		nil,
	)
	c.cpuLogicalProcessorsCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "logical_processor"),
		"Number of logical processors per CPU",
		[]string{
			"device_id",
		},
		nil,
	)
	c.cpuL2CacheSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "l2_cache_size"),
		"Size of L2 cache per CPU",
		[]string{
			"device_id",
		},
		nil,
	)
	c.cpuL3CacheSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "l3_cache_size"),
		"Size of L3 cache per CPU",
		[]string{
			"device_id",
		},
		nil,
	)

	return nil
}

type win32Processor struct {
	Architecture              uint32
	DeviceID                  string
	Description               string
	Family                    uint16
	L2CacheSize               uint32
	L3CacheSize               uint32
	Name                      string
	ThreadCount               uint32
	NumberOfCores             uint32
	NumberOfEnabledCore       uint32
	NumberOfLogicalProcessors uint32
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	logger = logger.With(slog.String("collector", Name))
	if err := c.collect(ch); err != nil {
		logger.Error("failed collecting cpu_info metrics",
			slog.Any("err", err),
		)

		return err
	}

	return nil
}

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	var dst []win32Processor
	// We use a static query here because the provided methods in wmi.go all issue a SELECT *;
	// This results in the time-consuming LoadPercentage field being read which seems to measure each CPU
	// serially over a 1 second interval, so the scrape time is at least 1s * num_sockets
	if err := c.wmiClient.Query("SELECT Architecture, DeviceId, Description, Family, L2CacheSize, L3CacheSize, Name, ThreadCount, NumberOfCores, NumberOfEnabledCore, NumberOfLogicalProcessors FROM Win32_Processor", &dst); err != nil {
		return err
	}

	if len(dst) == 0 {
		return errors.New("WMI query returned empty result set")
	}

	// Some CPUs end up exposing trailing spaces for certain strings, so clean them up
	for _, processor := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.cpuInfo,
			prometheus.GaugeValue,
			1.0,
			strconv.Itoa(int(processor.Architecture)),
			strings.TrimRight(processor.DeviceID, " "),
			strings.TrimRight(processor.Description, " "),
			strconv.Itoa(int(processor.Family)),
			strings.TrimRight(processor.Name, " "),
		)
		ch <- prometheus.MustNewConstMetric(
			c.cpuCoreCount,
			prometheus.GaugeValue,
			float64(processor.NumberOfCores),
			strings.TrimRight(processor.DeviceID, " "),
		)
		ch <- prometheus.MustNewConstMetric(
			c.cpuEnabledCoreCount,
			prometheus.GaugeValue,
			float64(processor.NumberOfEnabledCore),
			strings.TrimRight(processor.DeviceID, " "),
		)
		ch <- prometheus.MustNewConstMetric(
			c.cpuLogicalProcessorsCount,
			prometheus.GaugeValue,
			float64(processor.NumberOfLogicalProcessors),
			strings.TrimRight(processor.DeviceID, " "),
		)
		ch <- prometheus.MustNewConstMetric(
			c.cpuThreadCount,
			prometheus.GaugeValue,
			float64(processor.ThreadCount),
			strings.TrimRight(processor.DeviceID, " "),
		)
		ch <- prometheus.MustNewConstMetric(
			c.cpuL2CacheSize,
			prometheus.GaugeValue,
			float64(processor.L2CacheSize),
			strings.TrimRight(processor.DeviceID, " "),
		)
		ch <- prometheus.MustNewConstMetric(
			c.cpuL3CacheSize,
			prometheus.GaugeValue,
			float64(processor.L3CacheSize),
			strings.TrimRight(processor.DeviceID, " "),
		)
	}

	return nil
}

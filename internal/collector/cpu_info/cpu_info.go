//go:build windows

package cpu_info

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "cpu_info"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for a few WMI metrics in Win32_Processor.
type Collector struct {
	config    Config
	miSession *mi.Session
	miQuery   mi.Query

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

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(_ *slog.Logger, miSession *mi.Session) error {
	if miSession == nil {
		return errors.New("miSession is nil")
	}

	miQuery, err := mi.NewQuery("SELECT Architecture, DeviceId, Description, Family, L2CacheSize, L3CacheSize, Name, ThreadCount, NumberOfCores, NumberOfEnabledCore, NumberOfLogicalProcessors FROM Win32_Processor")
	if err != nil {
		return fmt.Errorf("failed to create WMI query: %w", err)
	}

	c.miQuery = miQuery
	c.miSession = miSession

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

type miProcessor struct {
	Architecture              uint32 `mi:"Architecture"`
	DeviceID                  string `mi:"DeviceID"`
	Description               string `mi:"Description"`
	Family                    uint16 `mi:"Family"`
	L2CacheSize               uint32 `mi:"L2CacheSize"`
	L3CacheSize               uint32 `mi:"L3CacheSize"`
	Name                      string `mi:"Name"`
	ThreadCount               uint32 `mi:"ThreadCount"`
	NumberOfCores             uint32 `mi:"NumberOfCores"`
	NumberOfEnabledCore       uint32 `mi:"NumberOfEnabledCore"`
	NumberOfLogicalProcessors uint32 `mi:"NumberOfLogicalProcessors"`

	Total int
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	var dst []miProcessor
	if err := c.miSession.Query(&dst, mi.NamespaceRootCIMv2, c.miQuery); err != nil {
		return fmt.Errorf("WMI query failed: %w", err)
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

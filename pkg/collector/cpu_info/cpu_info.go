//go:build windows

package cpu_info

import (
	"errors"
	"strconv"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	Name = "cpu_info"
	// If you are adding additional labels to the metric, make sure that they get added in here as well. See below for explanation.
	win32ProcessorQuery = "SELECT Architecture, DeviceId, Description, Family, L2CacheSize, L3CacheSize, Name FROM Win32_Processor"
)

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for a few WMI metrics in Win32_Processor.
type Collector struct {
	logger log.Logger

	cpuInfo *prometheus.Desc
}

func New(logger log.Logger, _ *Config) *Collector {
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

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build() error {
	c.cpuInfo = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, "", Name),
		"Labelled CPU information as provided by Win32_Processor",
		[]string{
			"architecture",
			"device_id",
			"description",
			"family",
			"l2_cache_size",
			"l3_cache_size",
			"name",
		},
		nil,
	)

	return nil
}

type win32_Processor struct {
	Architecture uint32
	DeviceID     string
	Description  string
	Family       uint16
	L2CacheSize  uint32
	L3CacheSize  uint32
	Name         string
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.collect(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting cpu_info metrics", "err", err)
		return err
	}
	return nil
}

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	var dst []win32_Processor
	// We use a static query here because the provided methods in wmi.go all issue a SELECT *;
	// This results in the time-consuming LoadPercentage field being read which seems to measure each CPU
	// serially over a 1 second interval, so the scrape time is at least 1s * num_sockets
	if err := wmi.Query(win32ProcessorQuery, &dst); err != nil {
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
			strconv.Itoa(int(processor.L2CacheSize)),
			strconv.Itoa(int(processor.L3CacheSize)),
			strings.TrimRight(processor.Name, " "),
		)
	}

	return nil
}

//go:build windows
// +build windows

package collector

import (
	"errors"
	"strconv"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
)

// If you are adding additional labels to the metric, make sure that they get added in here as well. See below for explanation.
const (
	win32ProcessorQuery = "SELECT Architecture, DeviceId, Description, Family, L2CacheSize, L3CacheSize, Name FROM Win32_Processor"
)

// A CpuInfoCollector is a Prometheus collector for a few WMI metrics in Win32_Processor
type CpuInfoCollector struct {
	logger  log.Logger
	CpuInfo *prometheus.Desc
}

func newCpuInfoCollector(logger log.Logger) (Collector, error) {
	const subsystem = "cpu_info"

	return &CpuInfoCollector{
		logger: log.With(logger, "collector", subsystem),

		CpuInfo: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", subsystem),
			"Labeled CPU information as provided provided by Win32_Processor",
			[]string{
				"architecture",
				"device_id",
				"description",
				"family",
				"l2_cache_size",
				"l3_cache_size",
				"name"},
			nil,
		),
	}, nil
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
func (c *CpuInfoCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting cpu_info metrics", "desc", desc, "err", err)
		return err
	}
	return nil
}

func (c *CpuInfoCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []win32_Processor
	// We use a static query here because the provided methods in wmi.go all issue a SELECT *;
	// This results in the time consuming LoadPercentage field being read which seems to measure each CPU
	// serially over a 1 second interval, so the scrape time is at least 1s * num_sockets
	if err := wmi.Query(win32ProcessorQuery, &dst); err != nil {
		return nil, err
	}
	if len(dst) == 0 {
		return nil, errors.New("WMI query returned empty result set")
	}

	// Some CPUs end up exposing trailing spaces for certain strings, so clean them up
	for _, processor := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.CpuInfo,
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

	return nil, nil
}

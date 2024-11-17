//go:build windows

package pagefile

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/headers/psapi"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "pagefile"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI metrics.
type Collector struct {
	config Config

	perfDataCollector *perfdata.Collector

	pagingFreeBytes  *prometheus.Desc
	pagingLimitBytes *prometheus.Desc
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
	c.perfDataCollector.Close()

	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *mi.Session) error {
	var err error

	c.perfDataCollector, err = perfdata.NewCollector("Paging File", perfdata.InstanceAll, []string{
		usage,
	})
	if err != nil {
		return fmt.Errorf("failed to create Paging File collector: %w", err)
	}

	c.pagingLimitBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "limit_bytes"),
		"Number of bytes that can be stored in the operating system paging files. 0 (zero) indicates that there are no paging files",
		[]string{"file"},
		nil,
	)

	c.pagingFreeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "free_bytes"),
		"Number of bytes that can be mapped into the operating system paging files without causing any other pages to be swapped out",
		[]string{"file"},
		nil,
	)

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	data, err := c.perfDataCollector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect Paging File metrics: %w", err)
	}

	gpi, err := psapi.GetPerformanceInfo()
	if err != nil {
		return err
	}

	for fileName, pageFile := range data {
		fileString := strings.ReplaceAll(fileName, `\??\`, "")
		file, err := os.Stat(fileString)

		var fileSize float64

		// For unknown reasons, Windows doesn't always create a page file. Continue collection rather than aborting.
		if err == nil {
			fileSize = float64(file.Size())
		}

		ch <- prometheus.MustNewConstMetric(
			c.pagingFreeBytes,
			prometheus.GaugeValue,
			fileSize-(pageFile[usage].FirstValue*float64(gpi.PageSize)),
			fileString,
		)

		ch <- prometheus.MustNewConstMetric(
			c.pagingLimitBytes,
			prometheus.GaugeValue,
			fileSize,
			fileString,
		)
	}

	return nil
}

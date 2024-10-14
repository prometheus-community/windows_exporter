//go:build windows

package paging

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/headers/psapi"
	v1 "github.com/prometheus-community/windows_exporter/internal/perfdata/v1"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
	"golang.org/x/sys/windows/registry"
)

const Name = "paging"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI metrics.
type Collector struct {
	config Config

	pagingFreeBytes  *prometheus.Desc
	pagingLimitBytes *prometheus.Desc
}

type pagingFileCounter struct {
	Name      string
	Usage     float64 `perflib:"% Usage"`
	UsagePeak float64 `perflib:"% Usage Peak"`
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
	return []string{"Paging File"}, nil
}

func (c *Collector) Close(_ *slog.Logger) error {
	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *wmi.Client) error {
	c.pagingLimitBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "paging_limit_bytes"),
		"OperatingSystem.SizeStoredInPagingFiles",
		nil,
		nil,
	)

	c.pagingFreeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "paging_free_bytes"),
		"OperatingSystem.FreeSpaceInPagingFiles",
		nil,
		nil,
	)

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	logger = logger.With(slog.String("collector", Name))

	return c.collectPaging(ctx, logger, ch)
}

func (c *Collector) collectPaging(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	// Get total allocation of paging files across all disks.
	memManKey, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\Session Manager\Memory Management`, registry.QUERY_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key SYSTEM\\CurrentControlSet\\Control\\Session Manager\\Memory Management: %w", err)
	}

	defer memManKey.Close()

	pagingFiles, _, err := memManKey.GetStringsValue("ExistingPageFiles")
	if err != nil {
		return fmt.Errorf("failed to read registry key ExistingPageFiles: %w", err)
	}

	var fsipf float64

	for _, pagingFile := range pagingFiles {
		fileString := strings.ReplaceAll(pagingFile, `\??\`, "")
		file, err := os.Stat(fileString)
		// For unknown reasons, Windows doesn't always create a page file. Continue collection rather than aborting.
		if err != nil {
			logger.Debug(fmt.Sprintf("Failed to read page file (reason: %s): %s\n", err, fileString))
		} else {
			fsipf += float64(file.Size())
		}
	}

	gpi, err := psapi.GetPerformanceInfo()
	if err != nil {
		return err
	}

	pfc := make([]pagingFileCounter, 0)
	if err = v1.UnmarshalObject(ctx.PerfObjects["Paging File"], &pfc, logger); err != nil {
		return err
	}

	// Get current page file usage.
	var pfbRaw float64

	for _, pageFile := range pfc {
		if strings.Contains(strings.ToLower(pageFile.Name), "_total") {
			continue
		}

		pfbRaw += pageFile.Usage
	}

	// Subtract from total page file allocation on disk.
	pfb := fsipf - (pfbRaw * float64(gpi.PageSize))

	ch <- prometheus.MustNewConstMetric(
		c.pagingFreeBytes,
		prometheus.GaugeValue,
		pfb,
	)

	ch <- prometheus.MustNewConstMetric(
		c.pagingLimitBytes,
		prometheus.GaugeValue,
		fsipf,
	)

	return nil
}

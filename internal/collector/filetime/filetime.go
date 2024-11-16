//go:build windows

package filetime

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/alecthomas/kingpin/v2"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "filetime"

type Config struct {
	FilePatterns []string
}

var ConfigDefaults = Config{
	FilePatterns: []string{},
}

// A Collector is a Prometheus Collector for collecting file times.
type Collector struct {
	config Config

	logger    *slog.Logger
	fileMTime *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	if config.FilePatterns == nil {
		config.FilePatterns = ConfigDefaults.FilePatterns
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
	c.config.FilePatterns = make([]string, 0)

	var filePatterns string

	app.Flag(
		"collector.filetime.file-patterns",
		"Comma-separated list of file patterns. Each pattern is a glob pattern that can contain `*`, `?`, and `**` (recursive). See https://github.com/bmatcuk/doublestar#patterns",
	).Default(strings.Join(ConfigDefaults.FilePatterns, ",")).StringVar(&filePatterns)

	app.Action(func(*kingpin.ParseContext) error {
		// doublestar.Glob() requires forward slashes
		c.config.FilePatterns = strings.Split(filepath.ToSlash(filePatterns), ",")

		return nil
	})

	return c
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	c.logger = logger.With(slog.String("collector", Name))

	c.logger.Info("filetime collector is in an experimental state! It may subject to change.")

	c.fileMTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mtime_timestamp_seconds"),
		"File modification time",
		[]string{"file"},
		nil,
	)

	for _, filePattern := range c.config.FilePatterns {
		basePath, pattern := doublestar.SplitPattern(filePattern)

		_, err := doublestar.Glob(os.DirFS(basePath), pattern, doublestar.WithFilesOnly())
		if err != nil {
			return fmt.Errorf("invalid glob pattern: %w", err)
		}
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	wg := sync.WaitGroup{}

	for _, filePattern := range c.config.FilePatterns {
		wg.Add(1)

		go func(filePattern string) {
			defer wg.Done()

			if err := c.collectGlobFilePath(ch, filePattern); err != nil {
				c.logger.Error("failed collecting metrics for filepath",
					slog.String("filepath", filePattern),
					slog.Any("err", err),
				)
			}
		}(filePattern)
	}

	wg.Wait()

	return nil
}

func (c *Collector) collectGlobFilePath(ch chan<- prometheus.Metric, filePattern string) error {
	basePath, pattern := doublestar.SplitPattern(filePattern)
	basePathFS := os.DirFS(basePath)

	matches, err := doublestar.Glob(basePathFS, pattern, doublestar.WithFilesOnly())
	if err != nil {
		return fmt.Errorf("failed to glob: %w", err)
	}

	for _, match := range matches {
		filePath := filepath.Join(basePath, match)

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			c.logger.Warn("failed to state file",
				slog.String("file", filePath),
				slog.Any("err", err),
			)

			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.fileMTime,
			prometheus.GaugeValue,
			float64(fileInfo.ModTime().UTC().Unix()),
			filePath,
		)
	}

	return nil
}

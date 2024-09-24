//go:build windows

package filetime

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
	"golang.org/x/sys/windows"
)

const Name = "filetime"

type Config struct {
	filePatterns      []string
	EnableGlobPattern bool
}

var ConfigDefaults = Config{
	filePatterns:      []string{},
	EnableGlobPattern: false,
}

// A Collector is a Prometheus Collector for collecting file times
type Collector struct {
	config Config

	fileATime *prometheus.Desc
	fileMTime *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	if config.filePatterns == nil {
		config.filePatterns = ConfigDefaults.filePatterns
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
	c.config.filePatterns = make([]string, 0)

	var filePatterns string

	app.Flag(
		"collectors.filetime.file-patterns",
		"Comma-separated list of file paths. The file paths can include wildcard characters, for example, an asterisk (*) or a question mark (?).",
	).Default(strings.Join(ConfigDefaults.filePatterns, ",")).StringVar(&filePatterns)

	app.Flag(
		"collectors.filetime.enable-glob-pattern",
		"Enable glob syntax in collectors.filetime.filepaths. This supports the standard glob syntax, including recursive pattern, but access time is not supported.",
	).Default(strconv.FormatBool(c.config.EnableGlobPattern)).BoolVar(&c.config.EnableGlobPattern)

	app.Action(func(*kingpin.ParseContext) error {
		if c.config.EnableGlobPattern {
			filePatterns = strings.ReplaceAll(filePatterns, "\\", "/")
		}

		c.config.filePatterns = strings.Split(filePatterns, ",")

		return nil
	})

	return c
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

func (c *Collector) Build(_ *slog.Logger, _ *wmi.Client) error {
	c.fileATime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "atime_timestamp_seconds"),
		"File access time",
		[]string{"file"},
		nil,
	)
	c.fileMTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mtime_timestamp_seconds"),
		"File modification time",
		[]string{"file"},
		nil,
	)

	if c.config.EnableGlobPattern {
		for _, filePattern := range c.config.filePatterns {
			basePath, pattern := doublestar.SplitPattern(filePattern)

			_, err := doublestar.Glob(os.DirFS(basePath), pattern, doublestar.WithFilesOnly())
			if err != nil {
				return fmt.Errorf("invalid glob pattern: %w", err)
			}
		}
	} else {
		var data windows.Win32finddata

		for _, filePattern := range c.config.filePatterns {
			firstFile, err := windows.FindFirstFile(windows.StringToUTF16Ptr(filePattern), &data)
			if err != nil {
				return fmt.Errorf("invalid pattern: %w", err)
			}

			if err = windows.FindClose(firstFile); err != nil {
				return fmt.Errorf("failed to close handle: %w", err)
			}
		}
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	logger = logger.With(slog.String("collector", Name))

	var err error

	if c.config.EnableGlobPattern {
		err = c.collectGlob(logger, ch)
	} else {
		err = c.collectWin32(logger, ch)
	}

	return err
}

// collectWin32 collects file times for each file path in the config. It using Win32 FindFirstFile and FindNextFile.
func (c *Collector) collectGlob(logger *slog.Logger, ch chan<- prometheus.Metric) error {
	wg := sync.WaitGroup{}

	for _, filePattern := range c.config.filePatterns {
		wg.Add(1)

		go func(filePattern string) {
			defer wg.Done()

			if err := c.collectGlobFilePath(logger, ch, filePattern); err != nil {
				logger.Error("failed collecting metrics for filepath",
					slog.String("filepath", filePattern),
					slog.Any("err", err),
				)
			}
		}(filePattern)
	}

	wg.Wait()

	return nil
}

func (c *Collector) collectGlobFilePath(logger *slog.Logger, ch chan<- prometheus.Metric, filePattern string) error {
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
			logger.Warn("failed to state file",
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

// collectWin32 collects file times for each file path in the config. It using Win32 FindFirstFile and FindNextFile.
func (c *Collector) collectWin32(logger *slog.Logger, ch chan<- prometheus.Metric) error {
	wg := sync.WaitGroup{}

	for _, filePath := range c.config.filePatterns {
		wg.Add(1)

		go func(filePath string) {
			defer wg.Done()

			if err := c.collectWin32FilePath(logger, ch, filePath); err != nil {
				logger.Error("failed collecting metrics for filepath",
					slog.String("filepath", filePath),
					slog.Any("err", err),
				)
			}
		}(filePath)
	}

	wg.Wait()

	return nil
}

func (c *Collector) collectWin32FilePath(logger *slog.Logger, ch chan<- prometheus.Metric, filePattern string) error {
	var data windows.Win32finddata
	handle, err := windows.FindFirstFile(windows.StringToUTF16Ptr(filePattern), &data)
	if err != nil {
		return fmt.Errorf("invalid pattern: %w", err)
	}

	defer func() {
		if err := windows.FindClose(handle); err != nil {
			logger.Warn("failed to close handle",
				slog.String("filepath", filePattern),
				slog.Any("err", err),
			)
		}
	}()

	for {
		fileName := windows.UTF16ToString(data.FileName[:])
		if fileName != "." && fileName != ".." {
			fileName = filepath.Join(filepath.Dir(filePattern), fileName)

			ch <- prometheus.MustNewConstMetric(
				c.fileATime,
				prometheus.GaugeValue,
				float64(time.Unix(0, data.LastAccessTime.Nanoseconds()).Unix()),
				fileName,
			)

			ch <- prometheus.MustNewConstMetric(
				c.fileMTime,
				prometheus.GaugeValue,
				float64(time.Unix(0, data.LastWriteTime.Nanoseconds()).Unix()),
				fileName,
			)
		}

		if err = windows.FindNextFile(handle, &data); err != nil {
			if errors.Is(err, windows.ERROR_NO_MORE_FILES) {
				break
			}

			return fmt.Errorf("failed to find next file: %w", err)
		}
	}

	return nil
}

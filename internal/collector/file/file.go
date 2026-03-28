// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows

package file

import (
	"fmt"
	"io/fs"
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

const Name = "file"

type Config struct {
	FilePatterns []string `yaml:"file-patterns"`
}

//nolint:gochecknoglobals
var ConfigDefaults = Config{
	FilePatterns: []string{},
}

// A Collector is a Prometheus Collector for collecting file times.
type Collector struct {
	config Config

	logger    *slog.Logger
	fileMTime *prometheus.Desc
	fileSize  *prometheus.Desc
	fileCount *prometheus.Desc
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

	var filePatterns string

	app.Flag(
		"collector.file.file-patterns",
		"Comma-separated list of file patterns. Each pattern is a glob pattern that can contain `*`, `?`, and `**` (recursive). See https://github.com/bmatcuk/doublestar#patterns",
	).Default(strings.Join(ConfigDefaults.FilePatterns, ",")).StringVar(&filePatterns)

	app.Action(func(*kingpin.ParseContext) error {
		for p := range strings.SplitSeq(filePatterns, ",") {
			if p != "" {
				c.config.FilePatterns = append(c.config.FilePatterns, p)
			}
		}

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

	c.logger.Info("file collector is in an experimental state! It may subject to change.")

	c.fileMTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mtime_timestamp_seconds"),
		"File modification time",
		[]string{"file"},
		nil,
	)

	c.fileSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "size_bytes"),
		"File size",
		[]string{"file"},
		nil,
	)

	c.fileCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "count"),
		"Number of files matching the pattern",
		[]string{"pattern"},
		nil,
	)

	for _, filePattern := range c.config.FilePatterns {
		if filePattern == "" {
			continue
		}

		if !doublestar.ValidatePattern(filepath.ToSlash(filePattern)) {
			return fmt.Errorf("invalid glob pattern: %s", filePattern)
		}
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	wg := sync.WaitGroup{}

	var mu sync.Mutex

	seenFiles := make(map[string]struct{})

	for _, filePattern := range c.config.FilePatterns {
		wg.Add(1)

		go func(filePattern string) {
			defer wg.Done()

			count, err := c.collectGlobFilePath(ch, filePattern, &mu, seenFiles)
			if err != nil {
				c.logger.Error("failed collecting metrics for filepath",
					slog.String("filepath", filePattern),
					slog.Any("err", err),
				)

				return
			}

			ch <- prometheus.MustNewConstMetric(
				c.fileCount,
				prometheus.GaugeValue,
				float64(count),
				filePattern,
			)
		}(filePattern)
	}

	wg.Wait()

	return nil
}

func (c *Collector) collectGlobFilePath(ch chan<- prometheus.Metric, filePattern string, mu *sync.Mutex, seenFiles map[string]struct{}) (int, error) {
	basePath, pattern := doublestar.SplitPattern(filepath.ToSlash(filePattern))
	basePathFS := os.DirFS(basePath)

	var count int

	err := doublestar.GlobWalk(basePathFS, pattern, func(path string, d fs.DirEntry) error {
		filePath := filepath.Join(basePath, path)

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			c.logger.Warn("failed to state file",
				slog.String("file", filePath),
				slog.Any("err", err),
			)

			return nil
		}

		count++

		mu.Lock()

		_, alreadySeen := seenFiles[filePath]
		if !alreadySeen {
			seenFiles[filePath] = struct{}{}
		}
		mu.Unlock()

		if alreadySeen {
			return nil
		}

		ch <- prometheus.MustNewConstMetric(
			c.fileMTime,
			prometheus.GaugeValue,
			float64(fileInfo.ModTime().UTC().UnixMicro())/1e6,
			filePath,
		)

		ch <- prometheus.MustNewConstMetric(
			c.fileSize,
			prometheus.GaugeValue,
			float64(fileInfo.Size()),
			filePath,
		)

		return nil
	}, doublestar.WithFilesOnly(), doublestar.WithCaseInsensitive())
	if err != nil {
		return count, fmt.Errorf("failed to glob: %w", err)
	}

	return count, nil
}

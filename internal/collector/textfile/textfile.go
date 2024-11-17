//go:build windows

// Copyright 2015 The Prometheus Authors
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

package textfile

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/dimchansky/utfbom"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

const Name = "textfile"

type Config struct {
	TextFileDirectories []string `yaml:"text_file_directories"`
}

var ConfigDefaults = Config{
	TextFileDirectories: []string{getDefaultPath()},
}

type Collector struct {
	config Config
	logger *slog.Logger

	// Only set for testing to get predictable output.
	mTime *float64

	mTimeDesc *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	if config.TextFileDirectories == nil {
		config.TextFileDirectories = ConfigDefaults.TextFileDirectories
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

	var textFileDirectories string

	app.Flag(
		"collector.textfile.directories",
		"Directory or Directories to read text files with metrics from.",
	).Default(strings.Join(ConfigDefaults.TextFileDirectories, ",")).StringVar(&textFileDirectories)

	app.Action(func(*kingpin.ParseContext) error {
		c.config.TextFileDirectories = strings.Split(textFileDirectories, ",")

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

	c.logger.Info("textfile Collector directories: " + strings.Join(c.config.TextFileDirectories, ","))

	c.mTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, "textfile", "mtime_seconds"),
		"Unixtime mtime of textfiles successfully read.",
		[]string{"file"},
		nil,
	)

	return nil
}

// Given a slice of metric families, determine if any two entries are duplicates.
// Duplicates will be detected where the metric name, labels and label values are identical.
func duplicateMetricEntry(metricFamilies []*dto.MetricFamily) bool {
	uniqueMetrics := make(map[string]map[string]string)

	for _, metricFamily := range metricFamilies {
		metricName := metricFamily.GetName()

		for _, metric := range metricFamily.GetMetric() {
			metricLabels := metric.GetLabel()
			labels := make(map[string]string)

			for _, label := range metricLabels {
				labels[label.GetName()] = label.GetValue()
			}
			// Check if key is present before appending
			_, mapContainsKey := uniqueMetrics[metricName]

			// Duplicate metric found with identical labels & label values
			if mapContainsKey && reflect.DeepEqual(uniqueMetrics[metricName], labels) {
				return true
			}

			uniqueMetrics[metricName] = labels
		}
	}

	return false
}

func (c *Collector) convertMetricFamily(logger *slog.Logger, metricFamily *dto.MetricFamily, ch chan<- prometheus.Metric) {
	var valType prometheus.ValueType

	var val float64

	allLabelNames := map[string]struct{}{}

	for _, metric := range metricFamily.GetMetric() {
		labels := metric.GetLabel()
		for _, label := range labels {
			if _, ok := allLabelNames[label.GetName()]; !ok {
				allLabelNames[label.GetName()] = struct{}{}
			}
		}
	}

	for _, metric := range metricFamily.GetMetric() {
		if metric.TimestampMs != nil {
			logger.Warn(fmt.Sprintf("Ignoring unsupported custom timestamp on textfile Collector metric %v", metric))
		}

		labels := metric.GetLabel()

		var names []string

		var values []string

		for _, label := range labels {
			names = append(names, label.GetName())
			values = append(values, label.GetValue())
		}

		for k := range allLabelNames {
			present := false

			for _, name := range names {
				if k == name {
					present = true

					break
				}
			}

			if !present {
				names = append(names, k)
				values = append(values, "")
			}
		}

		metricType := metricFamily.GetType()
		switch metricType {
		case dto.MetricType_COUNTER:
			valType = prometheus.CounterValue
			val = metric.GetCounter().GetValue()

		case dto.MetricType_GAUGE:
			valType = prometheus.GaugeValue
			val = metric.GetGauge().GetValue()

		case dto.MetricType_UNTYPED:
			valType = prometheus.UntypedValue
			val = metric.GetUntyped().GetValue()

		case dto.MetricType_SUMMARY:
			quantiles := map[float64]float64{}
			for _, q := range metric.GetSummary().GetQuantile() {
				quantiles[q.GetQuantile()] = q.GetValue()
			}
			ch <- prometheus.MustNewConstSummary(
				prometheus.NewDesc(
					metricFamily.GetName(),
					metricFamily.GetHelp(),
					names, nil,
				),
				metric.GetSummary().GetSampleCount(),
				metric.GetSummary().GetSampleSum(),
				quantiles, values...,
			)
		case dto.MetricType_HISTOGRAM:
			buckets := map[float64]uint64{}
			for _, b := range metric.GetHistogram().GetBucket() {
				buckets[b.GetUpperBound()] = b.GetCumulativeCount()
			}
			ch <- prometheus.MustNewConstHistogram(
				prometheus.NewDesc(
					metricFamily.GetName(),
					metricFamily.GetHelp(),
					names, nil,
				),
				metric.GetHistogram().GetSampleCount(),
				metric.GetHistogram().GetSampleSum(),
				buckets, values...,
			)
		default:
			logger.Error("unknown metric type for file")

			continue
		}

		if metricType == dto.MetricType_GAUGE || metricType == dto.MetricType_COUNTER || metricType == dto.MetricType_UNTYPED {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					metricFamily.GetName(),
					metricFamily.GetHelp(),
					names, nil,
				),
				valType, val, values...,
			)
		}
	}
}

func (c *Collector) exportMTimes(mTimes map[string]time.Time, ch chan<- prometheus.Metric) {
	// Export the mtimes of the successful files.
	if len(mTimes) > 0 {
		// Sorting is needed for predictable output comparison in tests.
		filenames := make([]string, 0, len(mTimes))
		for filename := range mTimes {
			filenames = append(filenames, filename)
		}

		sort.Strings(filenames)

		for _, filename := range filenames {
			mtime := float64(mTimes[filename].UnixNano() / 1e9)
			if c.mTime != nil {
				mtime = *c.mTime
			}
			ch <- prometheus.MustNewConstMetric(c.mTimeDesc, prometheus.GaugeValue, mtime, filename)
		}
	}
}

type carriageReturnFilteringReader struct {
	r io.Reader
}

// Read returns data from the underlying io.Reader, but with \r filtered out.
func (cr carriageReturnFilteringReader) Read(p []byte) (int, error) {
	buf := make([]byte, len(p))
	n, err := cr.r.Read(buf)

	if err != nil && err != io.EOF {
		return n, err
	}

	pi := 0

	for i := range n {
		if buf[i] != '\r' {
			p[pi] = buf[i]
			pi++
		}
	}

	return pi, err
}

// Collect implements the Collector interface.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	mTimes := map[string]time.Time{}

	// Create empty metricFamily slice here and append parsedFamilies to it inside the loop.
	// Once loop is complete, raise error if any duplicates are present.
	// This will ensure that duplicate metrics are correctly detected between multiple .prom files.
	var metricFamilies []*dto.MetricFamily

	// Iterate over files and accumulate their metrics.
	for _, directory := range c.config.TextFileDirectories {
		err := filepath.WalkDir(directory, func(path string, dirEntry os.DirEntry, err error) error {
			if err != nil {
				return fmt.Errorf("error reading directory: %w", err)
			}

			if !dirEntry.IsDir() && strings.HasSuffix(dirEntry.Name(), ".prom") {
				c.logger.Debug("Processing file: " + path)

				families_array, err := scrapeFile(path, c.logger)
				if err != nil {
					c.logger.Error(fmt.Sprintf("Error scraping file: %q. Skip File.", path),
						slog.Any("err", err),
					)

					return nil
				}

				fileInfo, err := os.Stat(path)
				if err != nil {
					c.logger.Error(fmt.Sprintf("Error reading file info: %q. Skip File.", path),
						slog.Any("err", err),
					)

					return nil
				}

				if _, hasName := mTimes[fileInfo.Name()]; hasName {
					c.logger.Error(fmt.Sprintf("Duplicate filename detected: %q. Skip File.", path))

					return nil
				}

				mTimes[fileInfo.Name()] = fileInfo.ModTime()

				metricFamilies = append(metricFamilies, families_array...)
			}

			return nil
		})
		if err != nil && directory != "" {
			c.logger.Error("Error reading textfile Collector directory: "+directory,
				slog.Any("err", err),
			)
		}
	}

	// If duplicates are detected across *multiple* files, return error.
	if duplicateMetricEntry(metricFamilies) {
		c.logger.Error("Duplicate metrics detected across multiple files")
	} else {
		for _, mf := range metricFamilies {
			c.convertMetricFamily(c.logger, mf, ch)
		}
	}

	c.exportMTimes(mTimes, ch)

	return nil
}

func scrapeFile(path string, logger *slog.Logger) ([]*dto.MetricFamily, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	var parser expfmt.TextParser

	r, encoding := utfbom.Skip(carriageReturnFilteringReader{r: file})
	if err = checkBOM(encoding); err != nil {
		return nil, err
	}

	parsedFamilies, err := parser.TextToMetricFamilies(r)

	closeErr := file.Close()
	if closeErr != nil {
		logger.Warn("error closing file "+path,
			slog.Any("err", closeErr),
		)
	}

	if err != nil {
		return nil, err
	}

	// Use temporary array to check for duplicates
	families_array := make([]*dto.MetricFamily, 0, len(parsedFamilies))

	for _, mf := range parsedFamilies {
		families_array = append(families_array, mf)

		for _, m := range mf.GetMetric() {
			if m.TimestampMs != nil {
				return nil, errors.New("textfile contains unsupported client-side timestamps")
			}
		}

		if mf.Help == nil {
			help := "Metric read from " + path
			mf.Help = &help
		}
	}

	// If duplicate metrics are detected in a *single* file, skip processing of file metrics
	if duplicateMetricEntry(families_array) {
		return nil, errors.New("duplicate metrics detected")
	}

	return families_array, nil
}

func checkBOM(encoding utfbom.Encoding) error {
	if encoding == utfbom.Unknown || encoding == utfbom.UTF8 {
		return nil
	}

	return errors.New(encoding.String())
}

func getDefaultPath() string {
	execPath, _ := os.Executable()

	return filepath.Join(filepath.Dir(execPath), "textfile_inputs")
}

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

//go:build !notextfile

package textfile

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/dimchansky/utfbom"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

const (
	Name                    = "textfile"
	FlagTextFileDirectory   = "collector.textfile.directory"
	FlagTextFileDirectories = "collector.textfile.directories"
)

type Config struct {
	TextFileDirectories string `yaml:"text_file_directories"`
}

var ConfigDefaults = Config{
	TextFileDirectories: getDefaultPath(),
}

type collector struct {
	logger log.Logger

	textFileDirectory   *string
	textFileDirectories *string

	directories string
	// Only set for testing to get predictable output.
	mtime *float64

	MtimeDesc *prometheus.Desc
}

func New(logger log.Logger, config *Config) types.Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	textFileDirectory := ""
	c := &collector{
		textFileDirectories: &config.TextFileDirectories,
		textFileDirectory:   &textFileDirectory,
	}
	c.SetLogger(logger)
	return c
}

func NewWithFlags(app *kingpin.Application) types.Collector {
	return &collector{
		textFileDirectory: app.Flag(
			FlagTextFileDirectory,
			"DEPRECATED: Use --collector.textfile.directories",
		).Default("").Hidden().String(),
		textFileDirectories: app.Flag(
			FlagTextFileDirectories,
			"Directory or Directories to read text files with metrics from.",
		).Default(ConfigDefaults.TextFileDirectories).String(),
	}
}

func (c *collector) GetName() string {
	return Name
}

func (c *collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *collector) GetPerfCounter() ([]string, error) {
	return []string{}, nil
}

func (c *collector) Build() error {
	c.directories = ""
	if *c.textFileDirectory != "" || *c.textFileDirectories != "" {
		c.directories = *c.textFileDirectory + "," + *c.textFileDirectories
		c.directories = strings.Trim(c.directories, ",")
	}
	_ = level.Info(c.logger).Log("msg", fmt.Sprintf("textfile collector directories: %s", c.directories))

	c.MtimeDesc = prometheus.NewDesc(
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
		metric_name := *metricFamily.Name
		for _, metric := range metricFamily.Metric {
			metric_labels := metric.GetLabel()
			labels := make(map[string]string)
			for _, label := range metric_labels {
				labels[label.GetName()] = label.GetValue()
			}
			// Check if key is present before appending
			_, mapContainsKey := uniqueMetrics[metric_name]

			// Duplicate metric found with identical labels & label values
			if mapContainsKey == true && reflect.DeepEqual(uniqueMetrics[metric_name], labels) {
				return true
			}
			uniqueMetrics[metric_name] = labels
		}
	}
	return false
}

func (c *collector) convertMetricFamily(metricFamily *dto.MetricFamily, ch chan<- prometheus.Metric) {
	var valType prometheus.ValueType
	var val float64

	allLabelNames := map[string]struct{}{}
	for _, metric := range metricFamily.Metric {
		labels := metric.GetLabel()
		for _, label := range labels {
			if _, ok := allLabelNames[label.GetName()]; !ok {
				allLabelNames[label.GetName()] = struct{}{}
			}
		}
	}

	for _, metric := range metricFamily.Metric {
		if metric.TimestampMs != nil {
			_ = level.Warn(c.logger).Log("msg", fmt.Sprintf("Ignoring unsupported custom timestamp on textfile collector metric %v", metric))
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
			if present == false {
				names = append(names, k)
				values = append(values, "")
			}
		}

		metricType := metricFamily.GetType()
		switch metricType {
		case dto.MetricType_COUNTER:
			valType = prometheus.CounterValue
			val = metric.Counter.GetValue()

		case dto.MetricType_GAUGE:
			valType = prometheus.GaugeValue
			val = metric.Gauge.GetValue()

		case dto.MetricType_UNTYPED:
			valType = prometheus.UntypedValue
			val = metric.Untyped.GetValue()

		case dto.MetricType_SUMMARY:
			quantiles := map[float64]float64{}
			for _, q := range metric.Summary.Quantile {
				quantiles[q.GetQuantile()] = q.GetValue()
			}
			ch <- prometheus.MustNewConstSummary(
				prometheus.NewDesc(
					*metricFamily.Name,
					metricFamily.GetHelp(),
					names, nil,
				),
				metric.Summary.GetSampleCount(),
				metric.Summary.GetSampleSum(),
				quantiles, values...,
			)
		case dto.MetricType_HISTOGRAM:
			buckets := map[float64]uint64{}
			for _, b := range metric.Histogram.Bucket {
				buckets[b.GetUpperBound()] = b.GetCumulativeCount()
			}
			ch <- prometheus.MustNewConstHistogram(
				prometheus.NewDesc(
					*metricFamily.Name,
					metricFamily.GetHelp(),
					names, nil,
				),
				metric.Histogram.GetSampleCount(),
				metric.Histogram.GetSampleSum(),
				buckets, values...,
			)
		default:
			_ = level.Error(c.logger).Log("msg", "unknown metric type for file")
			continue
		}
		if metricType == dto.MetricType_GAUGE || metricType == dto.MetricType_COUNTER || metricType == dto.MetricType_UNTYPED {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					*metricFamily.Name,
					metricFamily.GetHelp(),
					names, nil,
				),
				valType, val, values...,
			)
		}
	}
}

func (c *collector) exportMTimes(mtimes map[string]time.Time, ch chan<- prometheus.Metric) {
	// Export the mtimes of the successful files.
	if len(mtimes) > 0 {
		// Sorting is needed for predictable output comparison in tests.
		filenames := make([]string, 0, len(mtimes))
		for filename := range mtimes {
			filenames = append(filenames, filename)
		}
		sort.Strings(filenames)

		for _, filename := range filenames {
			mtime := float64(mtimes[filename].UnixNano() / 1e9)
			if c.mtime != nil {
				mtime = *c.mtime
			}
			ch <- prometheus.MustNewConstMetric(c.MtimeDesc, prometheus.GaugeValue, mtime, filename)
		}
	}
}

type carriageReturnFilteringReader struct {
	r io.Reader
}

// Read returns data from the underlying io.Reader, but with \r filtered out
func (cr carriageReturnFilteringReader) Read(p []byte) (int, error) {
	buf := make([]byte, len(p))
	n, err := cr.r.Read(buf)

	if err != nil && err != io.EOF {
		return n, err
	}

	pi := 0
	for i := 0; i < n; i++ {
		if buf[i] != '\r' {
			p[pi] = buf[i]
			pi++
		}
	}

	return pi, err
}

// Collect implements the Collector interface.
func (c *collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	errorMetric := 0.0
	mtimes := map[string]time.Time{}
	// Create empty metricFamily slice here and append parsedFamilies to it inside the loop.
	// Once loop is complete, raise error if any duplicates are present.
	// This will ensure that duplicate metrics are correctly detected between multiple .prom files.
	metricFamilies := []*dto.MetricFamily{}

	// Iterate over files and accumulate their metrics.
	for _, directory := range strings.Split(c.directories, ",") {
		err := filepath.WalkDir(directory, func(path string, dirEntry os.DirEntry, err error) error {
			if err != nil {
				_ = level.Error(c.logger).Log("msg", fmt.Sprintf("Error reading directory: %s", path), "err", err)
				errorMetric = 1.0
				return nil
			}
			if !dirEntry.IsDir() && strings.HasSuffix(dirEntry.Name(), ".prom") {
				_ = level.Debug(c.logger).Log("msg", fmt.Sprintf("Processing file: %s", path))
				families_array, err := scrapeFile(path, c.logger)
				if err != nil {
					_ = level.Error(c.logger).Log("msg", fmt.Sprintf("Error scraping file: %q. Skip File.", path), "err", err)
					errorMetric = 1.0
					return nil
				}
				fileInfo, err := os.Stat(path)
				if err != nil {
					_ = level.Error(c.logger).Log("msg", fmt.Sprintf("Error reading file info: %q. Skip File.", path), "err", err)
					errorMetric = 1.0
					return nil
				}
				if _, hasName := mtimes[fileInfo.Name()]; hasName {
					_ = level.Error(c.logger).Log("msg", fmt.Sprintf("Duplicate filename detected: %q. Skip File.", path))
					errorMetric = 1.0
					return nil
				}
				mtimes[fileInfo.Name()] = fileInfo.ModTime()
				metricFamilies = append(metricFamilies, families_array...)
			}
			return nil
		})
		if err != nil && directory != "" {
			_ = level.Error(c.logger).Log("msg", fmt.Sprintf("Error reading textfile collector directory: %s", c.directories), "err", err)
			errorMetric = 1.0
		}
	}

	// If duplicates are detected across *multiple* files, return error.
	if duplicateMetricEntry(metricFamilies) {
		_ = level.Error(c.logger).Log("msg", "Duplicate metrics detected across multiple files")
		errorMetric = 1.0
	} else {
		for _, mf := range metricFamilies {
			c.convertMetricFamily(mf, ch)
		}
	}

	c.exportMTimes(mtimes, ch)
	// Export if there were errors.
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, "textfile", "scrape_error"),
			"1 if there was an error opening or reading a file, 0 otherwise",
			nil, nil,
		),
		prometheus.GaugeValue, errorMetric,
	)
	return nil
}

func scrapeFile(path string, log log.Logger) ([]*dto.MetricFamily, error) {
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
		_ = level.Warn(log).Log("msg", fmt.Sprintf("Error closing file %q", path), "err", closeErr)
	}
	if err != nil {
		return nil, err
	}

	// Use temporary array to check for duplicates
	var families_array []*dto.MetricFamily

	for _, mf := range parsedFamilies {
		families_array = append(families_array, mf)
		for _, m := range mf.Metric {
			if m.TimestampMs != nil {
				return nil, fmt.Errorf("textfile contains unsupported client-side timestamps")
			}
		}
		if mf.Help == nil {
			help := fmt.Sprintf("Metric read from %s", path)
			mf.Help = &help
		}
	}

	// If duplicate metrics are detected in a *single* file, skip processing of file metrics
	if duplicateMetricEntry(families_array) {
		return nil, fmt.Errorf("duplicate metrics detected")
	}
	return families_array, nil
}

func checkBOM(encoding utfbom.Encoding) error {
	if encoding == utfbom.Unknown || encoding == utfbom.UTF8 {
		return nil
	}

	return fmt.Errorf(encoding.String())
}

func getDefaultPath() string {
	execPath, _ := os.Executable()
	return filepath.Join(filepath.Dir(execPath), "textfile_inputs")
}

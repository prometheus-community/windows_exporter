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
// +build !notextfile

package collector

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/dimchansky/utfbom"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	textFileDirectory = kingpin.Flag(
		"collector.textfile.directory",
		"Directory to read text files with metrics from.",
	).Default(getDefaultPath()).String()

	mtimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "textfile", "mtime_seconds"),
		"Unixtime mtime of textfiles successfully read.",
		[]string{"file"},
		nil,
	)
)

type textFileCollector struct {
	path string
	// Only set for testing to get predictable output.
	mtime *float64
}

func init() {
	registerCollector("textfile", NewTextFileCollector)
}

// NewTextFileCollector returns a new Collector exposing metrics read from files
// in the given textfile directory.
func NewTextFileCollector() (Collector, error) {
	return &textFileCollector{
		path: *textFileDirectory,
	}, nil
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

func convertMetricFamily(metricFamily *dto.MetricFamily, ch chan<- prometheus.Metric) {
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
			log.Warnf("Ignoring unsupported custom timestamp on textfile collector metric %v", metric)
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
			log.Errorf("unknown metric type for file")
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

func (c *textFileCollector) exportMTimes(mtimes map[string]time.Time, ch chan<- prometheus.Metric) {
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
			ch <- prometheus.MustNewConstMetric(mtimeDesc, prometheus.GaugeValue, mtime, filename)
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

// Update implements the Collector interface.
func (c *textFileCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	error := 0.0
	mtimes := map[string]time.Time{}

	// Iterate over files and accumulate their metrics.
	files, err := ioutil.ReadDir(c.path)
	if err != nil && c.path != "" {
		log.Errorf("Error reading textfile collector directory %q: %s", c.path, err)
		error = 1.0
	}

	// Create empty metricFamily slice here and append parsedFamilies to it inside the loop.
	// Once loop is complete, raise error if any duplicates are present.
	// This will ensure that duplicate metrics are correctly detected between multiple .prom files.
	var metricFamilies = []*dto.MetricFamily{}
fileLoop:
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".prom") {
			continue
		}
		path := filepath.Join(c.path, f.Name())
		log.Debugf("Processing file %q", path)
		file, err := os.Open(path)
		if err != nil {
			log.Errorf("Error opening %q: %v", path, err)
			error = 1.0
			continue
		}
		var parser expfmt.TextParser
		r, encoding := utfbom.Skip(carriageReturnFilteringReader{r: file})
		if err = checkBOM(encoding); err != nil {
			log.Errorf("Invalid file encoding detected in %s: %s - file must be UTF8", path, err.Error())
			error = 1.0
			continue
		}
		parsedFamilies, err := parser.TextToMetricFamilies(r)
		closeErr := file.Close()
		if closeErr != nil {
			log.Warnf("Error closing file: %v", err)
		}
		if err != nil {
			log.Errorf("Error parsing %q: %v", path, err)
			error = 1.0
			continue
		}

		// Use temporary array to check for duplicates
		var families_array []*dto.MetricFamily

		for _, mf := range parsedFamilies {
			families_array = append(families_array, mf)
			for _, m := range mf.Metric {
				if m.TimestampMs != nil {
					log.Errorf("Textfile %q contains unsupported client-side timestamps, skipping entire file", path)
					error = 1.0
					continue fileLoop
				}
			}
			if mf.Help == nil {
				help := fmt.Sprintf("Metric read from %s", path)
				mf.Help = &help
			}
		}

		// If duplicate metrics are detected in a *single* file, skip processing of file metrics
		if duplicateMetricEntry(families_array) {
			log.Errorf("Duplicate metrics detected in file %s. Skipping file processing.", f.Name())
			error = 1.0
			continue
		}

		// Only set this once it has been parsed and validated, so that
		// a failure does not appear fresh.
		mtimes[f.Name()] = f.ModTime()

		for _, metricFamily := range parsedFamilies {
			metricFamilies = append(metricFamilies, metricFamily)
		}
	}

	// If duplicates are detected across *multiple* files, return error.
	if duplicateMetricEntry(metricFamilies) {
		log.Errorf("Duplicate metrics detected across multiple files")
		error = 1.0
	} else {
		for _, mf := range metricFamilies {
			convertMetricFamily(mf, ch)
		}
	}

	c.exportMTimes(mtimes, ch)

	// Export if there were errors.
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "textfile", "scrape_error"),
			"1 if there was an error opening or reading a file, 0 otherwise",
			nil, nil,
		),
		prometheus.GaugeValue, error,
	)
	return nil
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

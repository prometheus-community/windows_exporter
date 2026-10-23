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

package performancecounter

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"math"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/multiplier"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"go.yaml.in/yaml/v3"
)

const Name = "performancecounter"

var (
	reNonAlphaNum = regexp.MustCompile(`[^a-zA-Z0-9]`)

	//nolint:gochecknoglobals // strings.NewReplacer is safe for concurrent use
	stringReplacer = strings.NewReplacer(
		"%", "percent",
		"(", "",
		")", "",
	)
)

type Config struct {
	Objects []Object `yaml:"objects"`
}

//nolint:gochecknoglobals
var ConfigDefaults = Config{
	Objects: make([]Object, 0),
}

// A Collector is a Prometheus collector for performance counter metrics.
type Collector struct {
	config Config

	logger *slog.Logger

	objects []Object

	// meta
	subCollectorScrapeDurationDesc *prometheus.Desc
	subCollectorScrapeSuccessDesc  *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	if config.Objects == nil {
		config.Objects = ConfigDefaults.Objects
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

	var objects string

	app.Flag(
		"collector.performancecounter.objects",
		"Objects of performance data to observe. See docs for more information on how to use this flag. By default, no objects are observed.",
	).Default("").StringVar(&objects)

	app.Action(func(*kingpin.ParseContext) error {
		if objects == "" {
			return nil
		}

		// Reject unknown keys like the configuration file does, so that a typo
		// fails instead of silently dropping the setting.
		decoder := yaml.NewDecoder(strings.NewReader(objects))
		decoder.KnownFields(true)

		if err := decoder.Decode(&c.config.Objects); err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("failed to parse objects %s: %w", objects, err)
		}

		return nil
	})

	return c
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	for _, object := range c.objects {
		object.collector.Close()
	}

	return nil
}

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	c.logger = logger.With(slog.String("collector", Name))
	c.objects = make([]Object, 0, len(c.config.Objects))
	names := make([]string, 0, len(c.config.Objects))

	var errs []error

	for i, object := range c.config.Objects {
		if object.Name == "" {
			return errors.New("object name is required")
		}

		if object.Object == "" {
			errs = append(errs, fmt.Errorf("object %s: object is required", object.Name))

			continue
		}

		if slices.Contains(names, object.Name) {
			errs = append(errs, fmt.Errorf("object %s: name is duplicated", object.Name))

			continue
		}

		names = append(names, object.Name)

		if object.InstanceLabel == "" {
			object.InstanceLabel = "object_instance"
		}

		if object.Instances != nil && !validLabelName(object.InstanceLabel) {
			errs = append(errs, fmt.Errorf("object %s: %q is not a valid instance label", object.Name, object.InstanceLabel))

			continue
		}

		counters := make([]Counter, 0, len(object.Counters))
		counterNames := make([]string, 0, len(object.Counters))
		object.valueIndex = make(map[string]int, len(object.Counters))

		for j, counter := range object.Counters {
			if counter.Metric == "" {
				counter.Metric = sanitizeMetricName(
					fmt.Sprintf("%s_%s_%s_%s", types.Namespace, Name, object.Object, counter.Name),
				)
				c.config.Objects[i].Counters[j].Metric = counter.Metric
			}

			if counter.Name == "" {
				errs = append(errs, fmt.Errorf("object %s: counter name is required", object.Name))

				continue
			}

			m, err := multiplier.New(counter.Multiplier)
			if err != nil {
				errs = append(errs, fmt.Errorf("object %s: counter %s: %w", object.Name, counter.Name, err))

				continue
			}

			object.Counters[j].multiplier = m

			if _, ok := object.valueIndex[counter.Name]; ok {
				errs = append(errs, fmt.Errorf("object %s: counter name %s is duplicated", object.Name, counter.Name))

				continue
			}

			if k := slices.IndexFunc(counters, func(other Counter) bool {
				return other.Metric == counter.Metric && maps.Equal(other.Labels, counter.Labels)
			}); k != -1 {
				errs = append(errs, fmt.Errorf("object %s: counters %s and %s produce identical series %s, set a different metric name or labels",
					object.Name, counters[k].Name, counter.Name, counter.Metric,
				))

				continue
			}

			if err := object.buildDescs(&object.Counters[j]); err != nil {
				errs = append(errs, fmt.Errorf("object %s: counter %s: %w", object.Name, counter.Name, err))

				continue
			}

			counters = append(counters, counter)

			object.valueIndex[counter.Name] = len(counterNames)
			counterNames = append(counterNames, counter.Name)
		}

		if object.Type == "" {
			object.Type = pdh.CounterTypeRaw
		}

		collector, err := pdh.NewDynamicCollector(c.logger, object.Type, object.Object, object.Instances, counterNames)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed collector for %s: %w", object.Name, err))
		}

		object.collector = collector

		c.objects = append(c.objects, object)
	}

	c.subCollectorScrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "collector_duration_seconds"),
		"windows_exporter: Duration of an performancecounter child collection.",
		[]string{"collector"},
		nil,
	)
	c.subCollectorScrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "collector_success"),
		"windows_exporter: Whether a performancecounter child collector was successful.",
		[]string{"collector"},
		nil,
	)

	return errors.Join(errs...)
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric, _ time.Duration) error {
	var errs []error

	for _, perfDataObject := range c.objects {
		startTime := time.Now()
		err := c.collectObject(ch, perfDataObject)
		duration := time.Since(startTime)
		success := 1.0

		if err != nil {
			errs = append(errs, fmt.Errorf("failed to collect object %s: %w", perfDataObject.Name, err))
			success = 0.0

			c.logger.Debug("performancecounter object collection failed",
				slog.String("object", perfDataObject.Name),
				slog.Duration("duration", duration),
				slog.Any("err", err),
			)
		} else {
			c.logger.Debug("performancecounter object collection succeeded",
				slog.String("object", perfDataObject.Name),
				slog.Duration("duration", duration),
			)
		}

		ch <- prometheus.MustNewConstMetric(
			c.subCollectorScrapeSuccessDesc,
			prometheus.GaugeValue,
			success,
			perfDataObject.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.subCollectorScrapeDurationDesc,
			prometheus.GaugeValue,
			duration.Seconds(),
			perfDataObject.Name,
		)
	}

	return errors.Join(errs...)
}

func (c *Collector) collectObject(ch chan<- prometheus.Metric, perfDataObject Object) error {
	var rows []pdh.Row

	err := perfDataObject.collector.Collect(&rows)
	if err != nil {
		return fmt.Errorf("failed to collect data: %w", err)
	}

	// Some objects, e.g. Process, report the same instance name more than
	// once. Each occurrence is published with a "#N" suffix, see the docs.
	pdh.SuffixDuplicates(rows,
		func(row *pdh.Row) *string { return &row.Name },
		func(row *pdh.Row) int { return row.Occurrence },
	)

	metricTypes := counterMetricTypes(perfDataObject)

	var errs []error

	for _, row := range rows {
		for j, counter := range perfDataObject.Counters {
			valueIndex, ok := perfDataObject.valueIndex[counter.Name]
			if !ok {
				errs = append(errs, fmt.Errorf("%s not found in collected data", counter.Name))

				continue
			}

			collectedCounterValue := row.Values[valueIndex]

			// The counter has no valid value for this instance in this sample.
			// Omit it instead of reporting zero, which looks like a counter reset.
			if math.IsNaN(collectedCounterValue) {
				continue
			}

			value := counter.multiplier.Apply(collectedCounterValue)

			if counter.instanceDesc != nil && row.Name != pdh.InstanceEmpty {
				ch <- prometheus.MustNewConstMetric(counter.instanceDesc, metricTypes[j], value, row.Name)
			} else {
				ch <- prometheus.MustNewConstMetric(counter.desc, metricTypes[j], value)
			}
		}
	}

	return errors.Join(errs...)
}

// buildDescs creates the descs of counter once, so that a scrape doesn't
// build one desc per series. The instance becomes a variable label; a
// const label of the same name takes precedence over it.
func (o *Object) buildDescs(counter *Counter) error {
	const help = "windows_exporter: custom Performance Counter metric"

	counter.desc = prometheus.NewDesc(counter.Metric, help, nil, counter.Labels)
	if err := counter.desc.Err(); err != nil {
		return err
	}

	counter.instanceDesc = nil

	if _, ok := counter.Labels[o.InstanceLabel]; o.Instances == nil || ok {
		return nil
	}

	counter.instanceDesc = prometheus.NewDesc(counter.Metric, help, []string{o.InstanceLabel}, counter.Labels)

	return counter.instanceDesc.Err()
}

// validLabelName reports whether prometheus accepts name as label name.
func validLabelName(name string) bool {
	return model.UTF8Validation.IsValidLabelName(name) && !strings.HasPrefix(name, model.ReservedLabelPrefix)
}

// counterMetricTypes returns the metric type of every counter of the object.
// An explicit type in the configuration takes precedence over the type that
// PDH reports for the counter.
func counterMetricTypes(perfDataObject Object) []prometheus.ValueType {
	metricTypes := make([]prometheus.ValueType, len(perfDataObject.Counters))

	for i, counter := range perfDataObject.Counters {
		switch counter.Type {
		case "counter":
			metricTypes[i] = prometheus.CounterValue
		case "gauge":
			metricTypes[i] = prometheus.GaugeValue
		default:
			metricType, ok := perfDataObject.collector.MetricType(counter.Name)
			if !ok {
				metricType = prometheus.GaugeValue
			}

			metricTypes[i] = metricType
		}
	}

	return metricTypes
}

func sanitizeMetricName(name string) string {
	return strings.Trim(reNonAlphaNum.ReplaceAllString(strings.ToLower(stringReplacer.Replace(name)), "_"), "_")
}

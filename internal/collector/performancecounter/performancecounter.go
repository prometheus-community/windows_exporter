// Copyright 2024 The Prometheus Authors
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
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v3"
)

const Name = "performancecounter"

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

	metricNameReplacer *strings.Replacer

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

		if err := yaml.Unmarshal([]byte(objects), &c.config.Objects); err != nil {
			return fmt.Errorf("failed to parse objects: %w", err)
		}

		return nil
	})

	return c
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	for _, object := range c.config.Objects {
		object.collector.Close()
	}

	return nil
}

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	c.logger = logger.With(slog.String("collector", Name))

	names := make([]string, 0, len(c.config.Objects))
	errs := make([]error, 0, len(c.config.Objects))

	for i, object := range c.config.Objects {
		if object.Name == "" {
			return fmt.Errorf("object name is required")
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

		counters := make([]string, 0, len(object.Counters))
		for j, counter := range object.Counters {
			if counter.Name == "" {
				errs = append(errs, errors.New("counter name is required"))

				continue
			}

			if slices.Contains(counters, counter.Name) {
				errs = append(errs, fmt.Errorf("counter name %s is duplicated", counter.Name))

				continue
			}

			counters = append(counters, counter.Name)

			if counter.Metric == "" {
				c.config.Objects[i].Counters[j].Metric = c.sanitizeMetricName(
					fmt.Sprintf("%s_%s_%s_%s", types.Namespace, Name, object.Object, counter.Name),
				)
			}
		}

		collector, err := perfdata.NewCollector(object.Object, object.Instances, counters)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed collector for %s: %w", object.Name, err))

			continue
		}

		if object.InstanceLabel == "" {
			c.config.Objects[i].InstanceLabel = "instance"
		}

		c.config.Objects[i].collector = collector
	}

	c.metricNameReplacer = strings.NewReplacer(
		".", "",
		"%", "",
		"/", "_",
		" ", "_",
		"-", "_",
	)

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

	return errors.Join()
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	errs := make([]error, 0, len(c.config.Objects))

	for _, perfDataObject := range c.config.Objects {
		startTime := time.Now()
		err := c.collectObject(ch, perfDataObject)
		duration := time.Since(startTime)
		success := 1.0

		if err != nil {
			errs = append(errs, fmt.Errorf("failed to collect object %s: %w", perfDataObject.Name, err))
			success = 0.0

			c.logger.Debug(fmt.Sprintf("performancecounter collector %s failed after %s", perfDataObject.Name, duration),
				slog.Any("err", err),
			)
		} else {
			c.logger.Debug(fmt.Sprintf("performancecounter collector %s succeeded after %s", perfDataObject.Name, duration))
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

	return nil
}

func (c *Collector) collectObject(ch chan<- prometheus.Metric, perfDataObject Object) error {
	collectedPerfData, err := perfDataObject.collector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect data: %w", err)
	}

	var errs []error

	for collectedInstance, collectedInstanceCounters := range collectedPerfData {
		for _, counter := range perfDataObject.Counters {
			collectedCounterValue, ok := collectedInstanceCounters[counter.Name]
			if !ok {
				errs = append(errs, fmt.Errorf("counter %s not found in collected data", counter.Name))

				continue
			}

			labels := make(prometheus.Labels, len(counter.Labels)+2)
			labels["collector"] = perfDataObject.Name

			if collectedInstance != perfdata.InstanceEmpty {
				labels[perfDataObject.InstanceLabel] = collectedInstance
			}

			for key, value := range counter.Labels {
				labels[key] = value
			}

			var metricType prometheus.ValueType

			switch counter.Type {
			case "counter":
				metricType = prometheus.CounterValue
			case "gauge":
				metricType = prometheus.GaugeValue
			default:
				metricType = collectedCounterValue.Type
			}

			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					counter.Metric,
					"windows_exporter: custom Performance Counter metric",
					nil,
					labels,
				),
				metricType,
				collectedCounterValue.FirstValue,
			)

			if collectedCounterValue.SecondValue != 0 {
				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc(
						counter.Metric+"_second",
						"windows_exporter: custom Performance Counter metric",
						nil,
						labels,
					),
					metricType,
					collectedCounterValue.SecondValue,
				)
			}
		}
	}

	return errors.Join()
}

func (c *Collector) sanitizeMetricName(name string) string {
	return strings.Trim(c.metricNameReplacer.Replace(strings.ToLower(name)), "_")
}

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
	"reflect"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v3"
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

		if err := yaml.Unmarshal([]byte(objects), &c.config.Objects); err != nil {
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
	for _, object := range c.config.Objects {
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
		counters := make([]string, 0, len(object.Counters))
		fields := make([]reflect.StructField, 0, len(object.Counters)+2)

		for j, counter := range object.Counters {
			if counter.Metric == "" {
				c.config.Objects[i].Counters[j].Metric = sanitizeMetricName(
					fmt.Sprintf("%s_%s_%s_%s", types.Namespace, Name, object.Object, counter.Name),
				)
			}

			if counter.Name == "" {
				errs = append(errs, errors.New("counter name is required"))
				c.config.Objects = slices.Delete(c.config.Objects, i, 1)

				continue
			}

			if slices.Contains(counters, counter.Name) {
				errs = append(errs, fmt.Errorf("counter name %s is duplicated", counter.Name))

				continue
			}

			counters = append(counters, counter.Name)

			field, err := func(name string) (_ reflect.StructField, err error) {
				defer func() {
					if r := recover(); r != nil {
						err = fmt.Errorf("failed to create field for %s: %v", name, r)
					}
				}()

				return reflect.StructField{
					Name: strings.ToUpper(sanitizeMetricName(name)),
					Type: reflect.TypeOf(float64(0)),
					Tag:  reflect.StructTag(fmt.Sprintf(`perfdata:"%s"`, name)),
				}, nil
			}(counter.Name)
			if err != nil {
				errs = append(errs, err)

				continue
			}

			fields = append(fields, field)
		}

		if object.Instances != nil {
			fields = append(fields, reflect.StructField{
				Name: "Name",
				Type: reflect.TypeOf(""),
			})
		}

		fields = append(fields, reflect.StructField{
			Name: "MetricType",
			Type: reflect.TypeOf(prometheus.ValueType(0)),
		})

		valueType := reflect.StructOf(fields)

		if object.Type == "" {
			object.Type = pdh.CounterTypeRaw
		}

		collector, err := pdh.NewCollectorWithReflection(object.Type, object.Object, object.Instances, valueType)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed collector for %s: %w", object.Name, err))
		}

		if object.InstanceLabel == "" {
			object.InstanceLabel = "instance"
		}

		object.collector = collector
		object.perfDataObject = reflect.New(reflect.SliceOf(valueType)).Interface()

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
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	var errs []error

	for _, perfDataObject := range c.objects {
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

	return errors.Join(errs...)
}

func (c *Collector) collectObject(ch chan<- prometheus.Metric, perfDataObject Object) error {
	err := perfDataObject.collector.Collect(perfDataObject.perfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect data: %w", err)
	}

	var errs []error

	sliceValue := reflect.ValueOf(perfDataObject.perfDataObject).Elem().Interface()
	for i := range reflect.ValueOf(sliceValue).Len() {
		for _, counter := range perfDataObject.Counters {
			val := reflect.ValueOf(sliceValue).Index(i)

			field := val.FieldByName(strings.ToUpper(sanitizeMetricName(counter.Name)))
			if !field.IsValid() {
				errs = append(errs, fmt.Errorf("%s not found in collected data", counter.Name))

				continue
			}

			if field.Kind() != reflect.Float64 {
				errs = append(errs, fmt.Errorf("failed to cast %s to float64", counter.Name))

				continue
			}

			collectedCounterValue := field.Float()

			field = val.FieldByName("MetricType")
			if !field.IsValid() {
				errs = append(errs, errors.New("field MetricType not found in collected data"))

				continue
			}

			if field.Kind() != reflect.TypeOf(prometheus.ValueType(0)).Kind() {
				errs = append(errs, fmt.Errorf("failed to cast MetricType for %s to prometheus.ValueType", counter.Name))

				continue
			}

			metricType, _ := field.Interface().(prometheus.ValueType)

			labels := make(prometheus.Labels, len(counter.Labels)+1)

			if perfDataObject.Instances != nil {
				field := val.FieldByName("Name")
				if !field.IsValid() {
					errs = append(errs, errors.New("field Name not found in collected data"))

					continue
				}

				if field.Kind() != reflect.String {
					errs = append(errs, fmt.Errorf("failed to cast Name for %s to string", counter.Name))

					continue
				}

				collectedInstance := field.String()
				if collectedInstance != pdh.InstanceEmpty {
					labels[perfDataObject.InstanceLabel] = collectedInstance
				}
			}

			for key, value := range counter.Labels {
				labels[key] = value
			}

			switch counter.Type {
			case "counter":
				metricType = prometheus.CounterValue
			case "gauge":
				metricType = prometheus.GaugeValue
			}

			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					counter.Metric,
					"windows_exporter: custom Performance Counter metric",
					nil,
					labels,
				),
				metricType,
				collectedCounterValue,
			)
		}
	}

	return errors.Join(errs...)
}

func sanitizeMetricName(name string) string {
	return strings.Trim(reNonAlphaNum.ReplaceAllString(strings.ToLower(stringReplacer.Replace(name)), "_"), "_")
}

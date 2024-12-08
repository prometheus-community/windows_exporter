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
	"fmt"
	"log/slog"
	"strings"

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

	for i, object := range c.config.Objects {
		counters := make([]string, 0, len(object.Counters))
		for j, counter := range object.Counters {
			counters = append(counters, counter.Name)

			if counter.Metric == "" {
				c.config.Objects[i].Counters[j].Metric = sanitizeMetricName(fmt.Sprintf("%s_%s_%s_%s", types.Namespace, Name, object.Object, counter.Name))
			}
		}

		collector, err := perfdata.NewCollector(object.Object, object.Instances, counters)
		if err != nil {
			return fmt.Errorf("failed to create v2 collector: %w", err)
		}

		if object.InstanceLabel == "" {
			c.config.Objects[i].InstanceLabel = "instance"
		}

		c.config.Objects[i].collector = collector
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	for _, perfDataObject := range c.config.Objects {
		collectedPerfData, err := perfDataObject.collector.Collect()
		if err != nil {
			return fmt.Errorf("failed to collect data: %w", err)
		}

		for collectedInstance, collectedInstanceCounters := range collectedPerfData {
			for _, counter := range perfDataObject.Counters {
				collectedCounterValue, ok := collectedInstanceCounters[counter.Name]
				if !ok {
					c.logger.Warn(fmt.Sprintf("counter %s not found in collected data", counter.Name))

					continue
				}

				labels := make(prometheus.Labels, len(counter.Labels)+1)
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
	}

	return nil
}

func sanitizeMetricName(name string) string {
	replacer := strings.NewReplacer(
		".", "",
		"%", "",
		"/", "_",
		" ", "_",
		"-", "_",
	)

	return strings.Trim(replacer.Replace(strings.ToLower(name)), "_")
}

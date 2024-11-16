//go:build windows

package perfdata

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "perfdata"

type Config struct {
	Objects []Object `yaml:"objects"`
}

var ConfigDefaults = Config{
	Objects: make([]Object, 0),
}

// A Collector is a Prometheus collector for perfdata metrics.
type Collector struct {
	config Config
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
		"collector.perfdata.objects",
		"Objects of performance data to observe. See docs for more information on how to use this flag. By default, no objects are observed.",
	).Default("").StringVar(&objects)

	app.Action(func(*kingpin.ParseContext) error {
		if objects == "" {
			return nil
		}

		if err := json.Unmarshal([]byte(objects), &c.config.Objects); err != nil {
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
	logger.Warn("The perfdata collector is in an experimental state! The configuration may change in future. Please report any issues.")

	for i, object := range c.config.Objects {
		collector, err := perfdata.NewCollector(object.Object, object.Instances, slices.Sorted(maps.Keys(object.Counters)))
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
	for _, object := range c.config.Objects {
		data, err := object.collector.Collect()
		if err != nil {
			return fmt.Errorf("failed to collect data: %w", err)
		}

		for instance, counters := range data {
			for counter, value := range counters {
				var labels prometheus.Labels
				if instance != perfdata.EmptyInstance {
					labels = prometheus.Labels{object.InstanceLabel: instance}
				}

				metricType := value.Type

				if val, ok := object.Counters[counter]; ok {
					switch val.Type {
					case "counter":
						metricType = prometheus.CounterValue
					case "gauge":
						metricType = prometheus.GaugeValue
					}
				}

				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc(
						sanitizeMetricName(fmt.Sprintf("%s_perfdata_%s_%s", types.Namespace, object.Object, counter)),
						fmt.Sprintf("Performance data for \\%s\\%s", object.Object, counter),
						nil,
						labels,
					),
					metricType,
					value.FirstValue,
				)
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

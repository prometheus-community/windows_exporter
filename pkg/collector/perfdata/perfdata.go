//go:build windows

package perfdata

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/pdh"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v3"
)

const (
	Name = "perfdata"

	FlagPerfDataObjects               = "collector.perfdata.objects"
	FlagPerfDataIgnoredErrors         = "collector.perfdata.ignored-errors"
	FlagPerfDataUseWildcardsExpansion = "collector.perfdata.use-wildcards-expansion"
)

type Config struct {
	Objects               []pdh.PerfObject `yaml:"objects"`
	IgnoredErrors         []string         `yaml:"ignoredErrors"`
	UseWildcardsExpansion bool             `yaml:"useWildcardsExpansion"`
}

var ConfigDefaults = Config{
	Objects:               make([]pdh.PerfObject, 0),
	IgnoredErrors:         make([]string, 0),
	UseWildcardsExpansion: true,
}

// A collector is a Prometheus collector for perfdata metrics
type collector struct {
	logger log.Logger

	ignoredErrors         *[]string
	objectsPlain          *string
	objects               []pdh.PerfObject
	useWildcardsExpansion *bool

	perfCounters pdh.WinPerfCounters

	metrics map[string]map[string]metricMetadata
}

type metricMetadata struct {
	desc      *prometheus.Desc
	valueType prometheus.ValueType
}

func New(logger log.Logger, config *Config) types.Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &collector{
		ignoredErrors:         &config.IgnoredErrors,
		useWildcardsExpansion: &config.UseWildcardsExpansion,
		objects:               config.Objects,
	}
	c.SetLogger(logger)
	return c
}

func NewWithFlags(app *kingpin.Application) types.Collector {
	return &collector{
		ignoredErrors: app.Flag(
			FlagPerfDataIgnoredErrors,
			"IgnoredErrors accepts a list of PDH error codes, if this error is encountered it will be ignored. For example, you can provide \"PDH_NO_DATA\" to ignore performance counters with no instances, but by default no errors are ignored.",
		).Default("").Strings(),
		objectsPlain: app.Flag(
			FlagPerfDataObjects,
			"Objects of performance data to observe. See docs for more information on how to use this flag. By default, no objects are observed.",
		).Default("").String(),
		useWildcardsExpansion: app.Flag(
			FlagPerfDataUseWildcardsExpansion,
			"Wildcards can be used in the instance name and the counter name. Instance indexes will also be returned in the instance name.",
		).Default(strconv.FormatBool(ConfigDefaults.UseWildcardsExpansion)).Bool(),
	}
}

func (c *collector) GetName() string {
	return Name
}

func (c *collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

// GetPerfCounter implements the [types.Collector] interface.
func (c *collector) GetPerfCounter() ([]string, error) {
	return []string{}, nil
}

func (c *collector) Build() error {
	_ = level.Warn(c.logger).Log("msg", "perfdata collector is in an experimental state! The configuration may change in future. Please report any issues.")

	if *c.objectsPlain != "" {
		err := yaml.Unmarshal([]byte(*c.objectsPlain), &c.objects)
		if err != nil {
			return fmt.Errorf("error parsing object flag: %w", err)
		}
	}

	objects := make([]pdh.PerfObject, len(c.objects))

	for i, object := range c.objects {
		object.UseRawValues = true

		objects[i] = object
	}

	c.perfCounters = pdh.WinPerfCounters{
		Log:                        c.logger,
		UsePerfCounterTime:         false,
		UseWildcardsExpansion:      *c.useWildcardsExpansion,
		LocalizeWildcardsExpansion: true,
		IgnoredErrors:              *c.ignoredErrors,
		Object:                     objects,
	}

	err := c.perfCounters.Init()
	if err != nil {
		return fmt.Errorf("failed to initialize perf data: %w", err)
	}

	var perfCounterInfos map[string]pdh.CounterInfos
	perfCounterInfos, err = c.perfCounters.GetInfo()
	if err != nil {
		return fmt.Errorf("failed to get perf data info: %w", err)
	}

	counterInfos, ok := perfCounterInfos["localhost"]
	if !ok {
		return errors.New("missing perf data")
	}

	c.metrics = map[string]map[string]metricMetadata{}
	for objectName, objectCounters := range counterInfos {
		subSystem := sanitizeMetricName(objectName)

		c.metrics[objectName] = map[string]metricMetadata{}

		for counterName, counterInfo := range objectCounters {
			name := sanitizeMetricName(counterName)
			metadata := metricMetadata{}
			metadata.valueType, err = perflib.GetPrometheusValueType(counterInfo.CounterType)
			if err != nil {
				return fmt.Errorf("failed to get prometheus value type: %w", err)
			}

			metadata.desc = prometheus.NewDesc(
				prometheus.BuildFQName(types.Namespace+"_perfdata", subSystem, name),
				counterInfo.ExplainText,
				[]string{"instance"},
				nil,
			)

			c.metrics[objectName][counterName] = metadata
		}
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.collect(ch); err != nil {
		_ = level.Error(c.logger).Log("failed collecting perfdata metrics", "err", err)
		return err
	}
	return nil
}

func (c *collector) collect(ch chan<- prometheus.Metric) error {
	acc, err := c.perfCounters.Gather()
	if err != nil {
		return fmt.Errorf("failed to gather perf data: %w", err)
	}

	hostResult, ok := acc["localhost"]
	if !ok {
		return errors.New("missing perf data")
	}

	for objectName, objectCounters := range hostResult {
		for counterName, counters := range objectCounters {
			for instance, value := range counters {
				ch <- prometheus.MustNewConstMetric(
					c.metrics[objectName][counterName].desc,
					c.metrics[objectName][counterName].valueType,
					value,
					instance,
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
	)

	return strings.Trim(replacer.Replace(strings.ToLower(name)), "_")
}

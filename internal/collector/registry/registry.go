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

package registry

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"go.yaml.in/yaml/v3"
	winregistry "golang.org/x/sys/windows/registry"
)

const Name = "registry"

var reNonAlphaNum = regexp.MustCompile(`[^a-zA-Z0-9]`)

type Config struct {
	Keys []Key `yaml:"keys"`
}

//nolint:gochecknoglobals
var ConfigDefaults = Config{
	Keys: make([]Key, 0),
}

// A Collector is a Prometheus collector for Windows registry values.
type Collector struct {
	config Config

	logger *slog.Logger

	keys []Key

	keySuccessDesc *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	if config.Keys == nil {
		config.Keys = ConfigDefaults.Keys
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

	var keys string

	app.Flag(
		"collector.registry.keys",
		"Registry keys to collect REG_DWORD and REG_QWORD values from. See docs for more information on how to use this flag. By default, no keys are collected.",
	).Default("").StringVar(&keys)

	app.Action(func(*kingpin.ParseContext) error {
		if keys == "" {
			return nil
		}

		if err := yaml.Unmarshal([]byte(keys), &c.config.Keys); err != nil {
			return fmt.Errorf("failed to parse keys %s: %w", keys, err)
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

	c.keySuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "key_success"),
		"Whether the registry key could be read successfully.",
		[]string{"key"},
		nil,
	)

	c.keys = make([]Key, 0, len(c.config.Keys))
	labels := make([]string, 0, len(c.config.Keys))

	var errs []error

	// seenMetricHelps and seenMetricTypes track the first-seen help text and
	// value type for each metric name. Prometheus requires all metrics sharing
	// a name to carry identical help text and type; mismatches cause silent
	// metric drops at scrape time. Catching them here turns that into a loud
	// build-time error.
	seenMetricHelps := make(map[string]string)
	seenMetricTypes := make(map[string]prometheus.ValueType)

	for _, key := range c.config.Keys {
		if key.Key == "" {
			errs = append(errs, errors.New("key is required"))

			continue
		}

		hive, subPath, label, err := parseKeyPath(key.Key)
		if err != nil {
			errs = append(errs, err)

			continue
		}

		if slices.Contains(labels, label) {
			errs = append(errs, fmt.Errorf("key %s is duplicated", label))

			continue
		}

		labels = append(labels, label)

		if len(key.Values) == 0 {
			errs = append(errs, fmt.Errorf("key %s has no values configured", label))

			continue
		}

		// group identifies the key in logs and seeds auto-generated metric names.
		// It defaults to the normalized key path when no explicit name is given.
		group := key.Name
		if group == "" {
			group = label
		}

		values := make([]Value, 0, len(key.Values))
		valueNames := make([]string, 0, len(key.Values))

		for _, value := range key.Values {
			if value.Name == "" {
				errs = append(errs, fmt.Errorf("value name is required for key %s", label))

				continue
			}

			// Registry value names are matched case-insensitively, so lowercase
			// them to keep reads consistent across systems.
			value.Name = strings.ToLower(value.Name)

			if slices.Contains(valueNames, value.Name) {
				errs = append(errs, fmt.Errorf("value name %q of key %s is duplicated", value.Name, label))

				continue
			}

			valueNames = append(valueNames, value.Name)

			// If no metric name is given, derive one from the key group and value
			// name, mirroring the performancecounter collector.
			if value.Metric == "" {
				value.Metric = sanitizeMetricName(
					fmt.Sprintf("%s_%s_%s_%s", types.Namespace, Name, group, value.Name),
				)
			}

			switch value.Type {
			case "", "gauge":
				value.metricType = prometheus.GaugeValue
			case "counter":
				value.metricType = prometheus.CounterValue
			default:
				errs = append(errs, fmt.Errorf("value %q of key %s has invalid type %q, must be \"gauge\" or \"counter\"", value.Name, label, value.Type))

				continue
			}

			help := value.Help
			if help == "" {
				help = "windows_exporter: custom registry metric"
			}

			if prevHelp, seen := seenMetricHelps[value.Metric]; seen {
				if prevHelp != help {
					errs = append(errs, fmt.Errorf(
						"value %q of key %s: metric %q must have the same help text as other values sharing this metric name (got %q, want %q)",
						value.Name, label, value.Metric, help, prevHelp,
					))

					continue
				}

				if seenMetricTypes[value.Metric] != value.metricType {
					errs = append(errs, fmt.Errorf(
						"value %q of key %s: metric %q must have the same type as other values sharing this metric name",
						value.Name, label, value.Metric,
					))

					continue
				}
			} else {
				seenMetricHelps[value.Metric] = help
				seenMetricTypes[value.Metric] = value.metricType
			}

			value.desc = prometheus.NewDesc(
				value.Metric,
				help,
				nil,
				value.Labels,
			)

			values = append(values, value)
		}

		// Build only resolves and validates the static configuration; it does not
		// open the key. Whether a key can actually be opened and read depends on
		// runtime conditions (existence, ACLs, transient locks) that are reported
		// per key via the key_success metric at scrape time. Opening keys here
		// would turn a recoverable, observable per-key condition into a fatal
		// collector build error that aborts the whole exporter.
		key.hive = hive
		key.subPath = subPath
		key.label = label
		key.Values = values

		c.keys = append(c.keys, key)
	}

	return errors.Join(errs...)
}

// Collect sends the metric values for each configured registry key
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric, _ time.Duration) error {
	var errs []error

	for _, key := range c.keys {
		err := c.collectKey(ch, key)
		success := 1.0

		if err != nil {
			errs = append(errs, err)
			success = 0.0

			c.logger.Debug("failed to collect registry key",
				slog.String("key", key.label),
				slog.Any("err", err),
			)
		}

		ch <- prometheus.MustNewConstMetric(
			c.keySuccessDesc,
			prometheus.GaugeValue,
			success,
			key.label,
		)
	}

	return errors.Join(errs...)
}

func (c *Collector) collectKey(ch chan<- prometheus.Metric, key Key) error {
	rk, err := winregistry.OpenKey(key.hive, key.subPath, winregistry.QUERY_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key %s: %w", key.label, err)
	}

	defer func() {
		_ = rk.Close()
	}()

	var errs []error

	for _, value := range key.Values {
		val, _, err := rk.GetIntegerValue(value.Name)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to read value %q of registry key %s: %w", value.Name, key.label, err))

			continue
		}

		ch <- prometheus.MustNewConstMetric(
			value.desc,
			value.metricType,
			float64(val),
		)
	}

	return errors.Join(errs...)
}

// parseKeyPath splits a full registry key path into its hive and sub path,
// accepting short (HKLM) and long (HKEY_LOCAL_MACHINE) hive names as well as
// forward slashes as separators. The returned label is the normalized path
// with the short hive name and backslashes, lowercased so that the same key
// produces an identical label regardless of how it is cased in the registry or
// the configuration. The registry matches keys case-insensitively, so
// lowercasing the label can never collide with a distinct key.
func parseKeyPath(path string) (winregistry.Key, string, string, error) {
	normalized := strings.Trim(strings.ReplaceAll(path, "/", `\`), `\`)

	hiveName, subPath, _ := strings.Cut(normalized, `\`)

	var (
		hive  winregistry.Key
		label string
	)

	switch strings.ToUpper(hiveName) {
	case "HKLM", "HKEY_LOCAL_MACHINE":
		hive, label = winregistry.LOCAL_MACHINE, "hklm"
	case "HKCU", "HKEY_CURRENT_USER":
		hive, label = winregistry.CURRENT_USER, "hkcu"
	case "HKU", "HKEY_USERS":
		hive, label = winregistry.USERS, "hku"
	case "HKCR", "HKEY_CLASSES_ROOT":
		hive, label = winregistry.CLASSES_ROOT, "hkcr"
	case "HKCC", "HKEY_CURRENT_CONFIG":
		hive, label = winregistry.CURRENT_CONFIG, "hkcc"
	default:
		return 0, "", "", fmt.Errorf("unknown registry hive %q in key %q", hiveName, path)
	}

	if subPath != "" {
		label += `\` + strings.ToLower(subPath)
	}

	return hive, subPath, label, nil
}

// sanitizeMetricName turns an arbitrary string into a valid Prometheus metric
// name by lowercasing it, replacing every non-alphanumeric character with an
// underscore, and trimming leading and trailing underscores.
func sanitizeMetricName(name string) string {
	return strings.Trim(reNonAlphaNum.ReplaceAllString(strings.ToLower(name), "_"), "_")
}

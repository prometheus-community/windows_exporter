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

package wmi

import (
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"go.yaml.in/yaml/v3"
)

const (
	Name = "wmi"

	defaultNamespace = "root/CIMv2"

	// minQueryTimeout is the least time left in the scrape for which a query
	// is still started.
	minQueryTimeout = time.Millisecond
)

var errScrapeTimeout = errors.New("scrape timeout exceeded")

var (
	reNonAlphaNum = regexp.MustCompile(`[^a-zA-Z0-9]`)

	// reIdentifier matches WQL class and property names. They are interpolated
	// into the generated SELECT statement, so anything else is rejected.
	reIdentifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
)

type Config struct {
	Queries []Query `yaml:"queries"`
}

//nolint:gochecknoglobals
var ConfigDefaults = Config{
	Queries: make([]Query, 0),
}

// A Collector is a Prometheus collector for user-defined WMI queries.
type Collector struct {
	config Config

	logger    *slog.Logger
	miSession *mi.Session

	queries []Query

	querySuccessDesc  *prometheus.Desc
	queryDurationDesc *prometheus.Desc
}

// metricSignature is what every property sharing a metric name must agree
// on. Prometheus rejects a metric family whose members differ in help text,
// type or label names.
type metricSignature struct {
	help       string
	metricType prometheus.ValueType
	labelNames string
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	if config.Queries == nil {
		config.Queries = ConfigDefaults.Queries
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

	var queries string

	app.Flag(
		"collector.wmi.queries",
		"WMI queries to collect numeric properties from. See docs for more information on how to use this flag. By default, no queries are run.",
	).Default("").StringVar(&queries)

	app.Action(func(*kingpin.ParseContext) error {
		if strings.TrimSpace(queries) == "" {
			return nil
		}

		// Reject unknown keys, so that a typo like "propertys" fails instead
		// of silently dropping the setting.
		decoder := yaml.NewDecoder(strings.NewReader(queries))
		decoder.KnownFields(true)

		if err := decoder.Decode(&c.config.Queries); err != nil {
			return fmt.Errorf("failed to parse queries %s: %w", queries, err)
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

//nolint:gocognit,cyclop,funlen,maintidx // validation of the user configuration is a flat list of checks
func (c *Collector) Build(logger *slog.Logger, miSession *mi.Session) error {
	c.logger = logger.With(slog.String("collector", Name))

	c.logger.Info("wmi collector is in an experimental state! It may subject to change.")

	c.querySuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "query_success"),
		"Whether the WMI query and all of its configured properties could be read successfully.",
		[]string{"name"},
		nil,
	)
	c.queryDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "query_duration_seconds"),
		"Duration of the WMI query.",
		[]string{"name"},
		nil,
	)

	c.queries = make([]Query, 0, len(c.config.Queries))

	if len(c.config.Queries) == 0 {
		return nil
	}

	if miSession == nil {
		return errors.New("miSession is nil")
	}

	c.miSession = miSession

	// names maps the sanitized query name, which seeds auto-generated metric
	// names, to the configured one. Names that only differ in case or special
	// characters would produce the same metrics.
	names := make(map[string]string, len(c.config.Queries))
	signatures := make(map[string]metricSignature)

	// seriesQueries maps the series of queries without label properties to
	// the query producing it. Such a query emits exactly one series per
	// property, so a collision with another query is certain. Series of
	// queries with label properties depend on the instances returned at
	// runtime and cannot be checked here.
	seriesQueries := make(map[string]string)

	reservedMetrics := []string{
		prometheus.BuildFQName(types.Namespace, Name, "query_success"),
		prometheus.BuildFQName(types.Namespace, Name, "query_duration_seconds"),
	}

	var errs []error

	for _, query := range c.config.Queries {
		if query.Name == "" {
			errs = append(errs, errors.New("query name is required"))

			continue
		}

		if prev, ok := names[sanitizeMetricName(query.Name)]; ok {
			if prev == query.Name {
				errs = append(errs, fmt.Errorf("query %s: name is duplicated", query.Name))
			} else {
				errs = append(errs, fmt.Errorf("query %s: name produces the same metric names as query %s", query.Name, prev))
			}

			continue
		}

		names[sanitizeMetricName(query.Name)] = query.Name

		if !reIdentifier.MatchString(query.Class) {
			errs = append(errs, fmt.Errorf("query %s: class %q must be a valid WMI class name", query.Name, query.Class))

			continue
		}

		if len(query.Properties) == 0 {
			errs = append(errs, fmt.Errorf("query %s: no properties configured", query.Name))

			continue
		}

		if query.Namespace == "" {
			query.Namespace = defaultNamespace
		}

		// MI expects forward slashes, but backslashes are the common notation
		// in WMI tooling.
		query.Namespace = strings.ReplaceAll(query.Namespace, `\`, "/")

		var err error

		query.namespace, err = mi.NewNamespace(query.Namespace)
		if err != nil {
			errs = append(errs, fmt.Errorf("query %s: invalid namespace %q: %w", query.Name, query.Namespace, err))

			continue
		}

		// selected collects the property names for the SELECT statement. WMI
		// property names are case-insensitive, so they are compared lowercased.
		selected := make([]string, 0, len(query.LabelProperties)+len(query.Properties))
		selectedLower := make([]string, 0, cap(selected))
		addSelected := func(name string) {
			if !slices.Contains(selectedLower, strings.ToLower(name)) {
				selected = append(selected, name)
				selectedLower = append(selectedLower, strings.ToLower(name))
			}
		}

		queryErrs := len(errs)
		labelProperties := make([]LabelProperty, 0, len(query.LabelProperties))
		query.labelNames = make([]string, 0, len(query.LabelProperties))

		for _, labelProperty := range query.LabelProperties {
			if !reIdentifier.MatchString(labelProperty.Name) {
				errs = append(errs, fmt.Errorf("query %s: label property %q must be a valid WMI property name", query.Name, labelProperty.Name))

				continue
			}

			if labelProperty.Label == "" {
				labelProperty.Label = sanitizeMetricName(labelProperty.Name)
			}

			if slices.Contains(query.labelNames, labelProperty.Label) {
				errs = append(errs, fmt.Errorf("query %s: label %q is duplicated", query.Name, labelProperty.Label))

				continue
			}

			query.labelNames = append(query.labelNames, labelProperty.Label)
			labelProperties = append(labelProperties, labelProperty)

			addSelected(labelProperty.Name)
		}

		properties := make([]Property, 0, len(query.Properties))
		propertyNames := make([]string, 0, len(query.Properties))

		for _, property := range query.Properties {
			if !reIdentifier.MatchString(property.Name) {
				errs = append(errs, fmt.Errorf("query %s: property %q must be a valid WMI property name", query.Name, property.Name))

				continue
			}

			if slices.Contains(propertyNames, strings.ToLower(property.Name)) {
				errs = append(errs, fmt.Errorf("query %s: property %q is duplicated", query.Name, property.Name))

				continue
			}

			propertyNames = append(propertyNames, strings.ToLower(property.Name))

			// If no metric name is given, derive one from the query name and
			// property name, mirroring the registry collector.
			if property.Metric == "" {
				property.Metric = sanitizeMetricName(
					fmt.Sprintf("%s_%s_%s_%s", types.Namespace, Name, query.Name, property.Name),
				)
			}

			if slices.Contains(reservedMetrics, property.Metric) {
				errs = append(errs, fmt.Errorf("query %s: property %q: metric %q is reserved by the wmi collector", query.Name, property.Name, property.Metric))

				continue
			}

			switch property.Type {
			case "", "gauge":
				property.metricType = prometheus.GaugeValue
			case "counter":
				property.metricType = prometheus.CounterValue
			default:
				errs = append(errs, fmt.Errorf("query %s: property %q has invalid type %q, must be \"gauge\" or \"counter\"", query.Name, property.Name, property.Type))

				continue
			}

			if label, ok := firstCommonLabel(query.labelNames, property.Labels); ok {
				errs = append(errs, fmt.Errorf("query %s: property %q: constant label %q is already set by a label property", query.Name, property.Name, label))

				continue
			}

			if k := slices.IndexFunc(properties, func(other Property) bool {
				return other.Metric == property.Metric && maps.Equal(other.Labels, property.Labels)
			}); k != -1 {
				errs = append(errs, fmt.Errorf("query %s: properties %q and %q produce identical series %s, set a different metric name or labels",
					query.Name, properties[k].Name, property.Name, property.Metric,
				))

				continue
			}

			series := seriesKey(property.Metric, property.Labels)

			if prev, ok := seriesQueries[series]; ok && len(query.labelNames) == 0 {
				errs = append(errs, fmt.Errorf("query %s: property %q produces the same series %s as query %s, set a different metric name or labels",
					query.Name, property.Name, property.Metric, prev,
				))

				continue
			}

			if property.Help == "" {
				property.Help = "windows_exporter: custom WMI metric"
			}

			labelNames := append(slices.Collect(maps.Keys(property.Labels)), query.labelNames...)
			slices.Sort(labelNames)

			signature := metricSignature{
				help:       property.Help,
				metricType: property.metricType,
				labelNames: strings.Join(labelNames, ","),
			}

			if prev, seen := signatures[property.Metric]; seen && prev != signature {
				errs = append(errs, fmt.Errorf(
					"query %s: property %q: metric %q must have the same help text, type and label names as other properties sharing this metric name",
					query.Name, property.Name, property.Metric,
				))

				continue
			}

			property.desc = prometheus.NewDesc(
				property.Metric,
				property.Help,
				query.labelNames,
				property.Labels,
			)

			// Creating a sample metric surfaces invalid metric or label names,
			// such as reserved "__" labels, now instead of panicking at scrape
			// time.
			if _, err := prometheus.NewConstMetric(property.desc, property.metricType, 0, make([]string, len(query.labelNames))...); err != nil {
				errs = append(errs, fmt.Errorf("query %s: property %q: %w", query.Name, property.Name, err))

				continue
			}

			signatures[property.Metric] = signature

			if len(query.labelNames) == 0 {
				seriesQueries[series] = query.Name
			}

			properties = append(properties, property)

			addSelected(property.Name)
		}

		if len(errs) > queryErrs {
			continue
		}

		query.LabelProperties = labelProperties
		query.Properties = properties

		query.Where = strings.TrimSpace(query.Where)

		query.wql = fmt.Sprintf("SELECT %s FROM %s", strings.Join(selected, ", "), query.Class)
		if query.Where != "" {
			query.wql += " WHERE " + query.Where
		}

		query.query, err = mi.NewQuery(query.wql)
		if err != nil {
			errs = append(errs, fmt.Errorf("query %s: invalid query %q: %w", query.Name, query.wql, err))

			continue
		}

		// Build only validates the static configuration and does not run the
		// query. Whether a namespace, class or property exists and is readable
		// is a runtime condition, reported per query via query_success.
		c.queries = append(c.queries, query)
	}

	return errors.Join(errs...)
}

// Collect sends the metric values for each configured WMI query
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric, maxScrapeDuration time.Duration) error {
	// The queries run one after another and share the scrape timeout, so each
	// query only gets the time that is left. A non-positive timeout means no
	// deadline.
	var deadline time.Time

	if maxScrapeDuration > 0 {
		deadline = time.Now().Add(maxScrapeDuration)
	}

	var errs []error

	for _, query := range c.queries {
		startTime := time.Now()
		err := c.collectQuery(ch, query, deadline)
		duration := time.Since(startTime)
		success := 1.0

		if err != nil {
			// A failing query is reported via query_success. Like the registry
			// collector, it does not fail the whole collector.
			errs = append(errs, fmt.Errorf("failed to collect query %s: %w", query.Name, errors.Join(err, types.ErrNoData)))
			success = 0.0

			c.logger.Warn("wmi query failed",
				slog.String("query", query.Name),
				slog.String("wql", query.wql),
				slog.Duration("duration", duration),
				slog.Any("err", err),
			)
		} else {
			c.logger.Debug(fmt.Sprintf("wmi query %s succeeded after %s", query.Name, duration))
		}

		ch <- prometheus.MustNewConstMetric(
			c.querySuccessDesc,
			prometheus.GaugeValue,
			success,
			query.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.queryDurationDesc,
			prometheus.GaugeValue,
			duration.Seconds(),
			query.Name,
		)
	}

	return errors.Join(errs...)
}

func (c *Collector) collectQuery(ch chan<- prometheus.Metric, query Query, deadline time.Time) error {
	// Zero means no timeout, which leaves the MI default in place.
	var timeout time.Duration

	if !deadline.IsZero() {
		timeout = time.Until(deadline)

		// MI timeouts have microsecond resolution and treat zero as unset,
		// so a query is not started without a usable amount of time left.
		if timeout < minQueryTimeout {
			return errScrapeTimeout
		}
	}

	labelValues := make([]string, len(query.LabelProperties))

	// A property that cannot be read usually fails for every instance, so
	// each property reports at most one error per scrape.
	propertyErrs := make(map[string]error)

	err := c.miSession.QueryFunc(query.namespace, query.query, timeout, func(instance *mi.Instance) error {
		// The MI timeout does not stop a query that keeps returning
		// instances, so the deadline is also checked per instance.
		if !deadline.IsZero() && time.Now().After(deadline) {
			return errScrapeTimeout
		}

		for i, labelProperty := range query.LabelProperties {
			element, err := instance.GetElement(labelProperty.Name)
			if err != nil {
				return fmt.Errorf("failed to read label property %s: %w", labelProperty.Name, err)
			}

			if element.IsNull() {
				labelValues[i] = ""

				continue
			}

			labelValues[i], err = element.String()
			if err != nil {
				return fmt.Errorf("failed to read label property %s: %w", labelProperty.Name, err)
			}
		}

		for _, property := range query.Properties {
			if _, failed := propertyErrs[property.Name]; failed {
				continue
			}

			element, err := instance.GetElement(property.Name)
			if err != nil {
				propertyErrs[property.Name] = fmt.Errorf("failed to read property %s: %w", property.Name, err)

				continue
			}

			// WMI reports properties without a value as null. Exporting them
			// as 0 would be indistinguishable from a real 0, so skip them.
			if element.IsNull() {
				continue
			}

			value, err := element.Float64()
			if err != nil {
				propertyErrs[property.Name] = fmt.Errorf("failed to read property %s: %w", property.Name, err)

				continue
			}

			ch <- prometheus.MustNewConstMetric(
				property.desc,
				property.metricType,
				value,
				labelValues...,
			)
		}

		return nil
	})
	if err != nil {
		return err
	}

	errs := make([]error, 0, len(propertyErrs))

	for _, property := range query.Properties {
		if err, ok := propertyErrs[property.Name]; ok {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// firstCommonLabel returns the first label name in labelNames that is also a
// key of labels.
func firstCommonLabel(labelNames []string, labels map[string]string) (string, bool) {
	for _, name := range labelNames {
		if _, ok := labels[name]; ok {
			return name, true
		}
	}

	return "", false
}

// sanitizeMetricName turns an arbitrary string into a valid Prometheus metric
// or label name by lowercasing it, replacing every non-alphanumeric character
// with an underscore, and trimming leading and trailing underscores.
func sanitizeMetricName(name string) string {
	return strings.Trim(reNonAlphaNum.ReplaceAllString(strings.ToLower(name), "_"), "_")
}

// seriesKey identifies the series of a property without label properties: its
// metric name and constant labels.
func seriesKey(metric string, labels map[string]string) string {
	var sb strings.Builder

	sb.WriteString(metric)

	for _, name := range slices.Sorted(maps.Keys(labels)) {
		sb.WriteString("\xff" + name + "=" + labels[name])
	}

	return sb.String()
}

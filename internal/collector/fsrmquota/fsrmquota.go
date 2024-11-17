//go:build windows

package fsrmquota

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "fsrmquota"

type Config struct{}

var ConfigDefaults = Config{}

type Collector struct {
	config    Config
	miSession *mi.Session
	miQuery   mi.Query

	quotasCount *prometheus.Desc
	peakUsage   *prometheus.Desc
	size        *prometheus.Desc
	usage       *prometheus.Desc

	description     *prometheus.Desc
	disabled        *prometheus.Desc
	matchesTemplate *prometheus.Desc
	softLimit       *prometheus.Desc
	template        *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(_ *slog.Logger, miSession *mi.Session) error {
	if miSession == nil {
		return errors.New("miSession is nil")
	}

	miQuery, err := mi.NewQuery("SELECT * FROM MSFT_FSRMQuota")
	if err != nil {
		return fmt.Errorf("failed to create WMI query: %w", err)
	}

	c.miQuery = miQuery
	c.miSession = miSession

	c.quotasCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "count"),
		"Number of Quotas",
		nil,
		nil,
	)
	c.peakUsage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "peak_usage_bytes"),
		"The highest amount of disk space usage charged to this quota. (PeakUsage)",
		[]string{"path", "template"},
		nil,
	)
	c.size = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "size_bytes"),
		"The size of the quota. (Size)",
		[]string{"path", "template"},
		nil,
	)
	c.usage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "usage_bytes"),
		"The current amount of disk space usage charged to this quota. (Usage)",
		[]string{"path", "template"},
		nil,
	)
	c.description = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "description"),
		"Description of the quota (Description)",
		[]string{"path", "template", "description"},
		nil,
	)
	c.disabled = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "disabled"),
		"If 1, the quota is disabled. The default value is 0. (Disabled)",
		[]string{"path", "template"},
		nil,
	)
	c.softLimit = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "softlimit"),
		"If 1, the quota is a soft limit. If 0, the quota is a hard limit. The default value is 0. Optional (SoftLimit)",
		[]string{"path", "template"},
		nil,
	)
	c.template = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "template"),
		"Quota template name. (Template)",
		[]string{"path", "template"},
		nil,
	)
	c.matchesTemplate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "matchestemplate"),
		"If 1, the property values of this quota match those values of the template from which it was derived. (MatchesTemplate)",
		[]string{"path", "template"},
		nil,
	)

	return nil
}

// MSFT_FSRMQuota docs:
// https://docs.microsoft.com/en-us/previous-versions/windows/desktop/fsrm/msft-fsrmquota
type MSFT_FSRMQuota struct {
	Name string `mi:"Name"`

	Path        string `mi:"Path"`
	PeakUsage   uint64 `mi:"PeakUsage"`
	Size        uint64 `mi:"Size"`
	Usage       uint64 `mi:"Usage"`
	Description string `mi:"Description"`
	Template    string `mi:"Template"`
	// Threshold string `mi:"Threshold"`
	Disabled        bool `mi:"Disabled"`
	MatchesTemplate bool `mi:"MatchesTemplate"`
	SoftLimit       bool `mi:"SoftLimit"`
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	var dst []MSFT_FSRMQuota
	if err := c.miSession.Query(&dst, mi.NamespaceRootWindowsFSRM, c.miQuery); err != nil {
		return fmt.Errorf("WMI query failed: %w", err)
	}

	var count int

	for _, quota := range dst {
		count++
		path := quota.Path
		template := quota.Template
		Description := quota.Description

		ch <- prometheus.MustNewConstMetric(
			c.peakUsage,
			prometheus.GaugeValue,
			float64(quota.PeakUsage),
			path,
			template,
		)
		ch <- prometheus.MustNewConstMetric(
			c.size,
			prometheus.GaugeValue,
			float64(quota.Size),
			path,
			template,
		)
		ch <- prometheus.MustNewConstMetric(
			c.usage,
			prometheus.GaugeValue,
			float64(quota.Usage),
			path,
			template,
		)
		ch <- prometheus.MustNewConstMetric(
			c.description,
			prometheus.GaugeValue,
			1.0,
			path, template, Description,
		)
		ch <- prometheus.MustNewConstMetric(
			c.disabled,
			prometheus.GaugeValue,
			utils.BoolToFloat(quota.Disabled),
			path,
			template,
		)
		ch <- prometheus.MustNewConstMetric(
			c.matchesTemplate,
			prometheus.GaugeValue,
			utils.BoolToFloat(quota.MatchesTemplate),
			path,
			template,
		)
		ch <- prometheus.MustNewConstMetric(
			c.softLimit,
			prometheus.GaugeValue,
			utils.BoolToFloat(quota.SoftLimit),
			path,
			template,
		)
	}

	ch <- prometheus.MustNewConstMetric(
		c.quotasCount,
		prometheus.GaugeValue,
		float64(count),
	)

	return nil
}

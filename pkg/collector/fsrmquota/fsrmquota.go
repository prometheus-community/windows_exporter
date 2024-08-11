//go:build windows

package fsrmquota

import (
	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/utils"
	"github.com/prometheus-community/windows_exporter/pkg/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "fsrmquota"

type Config struct{}

var ConfigDefaults = Config{}

type Collector struct {
	config Config
	logger log.Logger

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

func New(logger log.Logger, config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		config: *config,
	}

	c.SetLogger(logger)

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *Collector) GetPerfCounter() ([]string, error) {
	return []string{}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build() error {
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

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.collect(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting fsrmquota metrics", "err", err)
		return err
	}
	return nil
}

// MSFT_FSRMQuota docs:
// https://docs.microsoft.com/en-us/previous-versions/windows/desktop/fsrm/msft-fsrmquota
type MSFT_FSRMQuota struct {
	Name string

	Path        string
	PeakUsage   uint64
	Size        uint64
	Usage       uint64
	Description string
	Template    string
	// Threshold             string
	Disabled        bool
	MatchesTemplate bool
	SoftLimit       bool
}

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	var dst []MSFT_FSRMQuota
	q := wmi.QueryAll(&dst, c.logger)

	var count int

	if err := wmi.QueryNamespace(q, &dst, "root/microsoft/windows/fsrm"); err != nil {
		return err
	}

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

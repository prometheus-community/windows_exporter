package collector

import (
	"github.com/StackExchange/wmi"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("fsrmquota", newFSRMQuotaCollector)
}

type FSRMQuotaCollector struct {
	QuotasCount *prometheus.Desc
	Path        *prometheus.Desc
	PeakUsage   *prometheus.Desc
	Size        *prometheus.Desc
	Usage       *prometheus.Desc

	Description     *prometheus.Desc
	Disabled        *prometheus.Desc
	MatchesTemplate *prometheus.Desc
	SoftLimit       *prometheus.Desc
	Template        *prometheus.Desc
}

func newFSRMQuotaCollector() (Collector, error) {
	const subsystem = "fsrmquota"
	return &FSRMQuotaCollector{
		QuotasCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "count"),
			"Number of Quotas",
			nil,
			nil,
		),
		PeakUsage: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "peak_usage_bytes"),
			"The highest amount of disk space usage charged to this quota. (PeakUsage)",
			[]string{"path", "template"},
			nil,
		),
		Size: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "size_bytes"),
			"The size of the quota. (Size)",
			[]string{"path", "template"},
			nil,
		),
		Usage: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "usage_bytes"),
			"The current amount of disk space usage charged to this quota. (Usage)",
			[]string{"path", "template"},
			nil,
		),
		Description: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "description"),
			"Description of the quota (Description)",
			[]string{"path", "template", "description"},
			nil,
		),
		Disabled: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "disabled"),
			"If 1, the quota is disabled. The default value is 0. (Disabled)",
			[]string{"path", "template"},
			nil,
		),
		SoftLimit: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "softlimit"),
			"If 1, the quota is a soft limit. If 0, the quota is a hard limit. The default value is 0. Optional (SoftLimit)",
			[]string{"path", "template"},
			nil,
		),
		Template: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "template"),
			"Quota template name. (Template)",
			[]string{"path", "template"},
			nil,
		),
		MatchesTemplate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "matchestemplate"),
			"If 1, the property values of this quota match those values of the template from which it was derived. (MatchesTemplate)",
			[]string{"path", "template"},
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *FSRMQuotaCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting fsrmquota metrics:", desc, err)
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
	//Threshold             string
	Disabled        bool
	MatchesTemplate bool
	SoftLimit       bool
}

func (c *FSRMQuotaCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []MSFT_FSRMQuota
	q := queryAll(&dst)

	var count int

	if err := wmi.QueryNamespace(q, &dst, "root/microsoft/windows/fsrm"); err != nil {
		return nil, err
	}

	for _, quota := range dst {

		count++
		path := quota.Path
		template := quota.Template
		Description := quota.Description

		ch <- prometheus.MustNewConstMetric(
			c.PeakUsage,
			prometheus.GaugeValue,
			float64(quota.PeakUsage),
			path,
			template,
		)
		ch <- prometheus.MustNewConstMetric(
			c.Size,
			prometheus.GaugeValue,
			float64(quota.Size),
			path,
			template,
		)
		ch <- prometheus.MustNewConstMetric(
			c.Usage,
			prometheus.GaugeValue,
			float64(quota.Usage),
			path,
			template,
		)
		ch <- prometheus.MustNewConstMetric(
			c.Description,
			prometheus.GaugeValue,
			1.0,
			path, template, Description,
		)
		ch <- prometheus.MustNewConstMetric(
			c.Disabled,
			prometheus.GaugeValue,
			boolToFloat(quota.Disabled),
			path,
			template,
		)
		ch <- prometheus.MustNewConstMetric(
			c.MatchesTemplate,
			prometheus.GaugeValue,
			boolToFloat(quota.MatchesTemplate),
			path,
			template,
		)
		ch <- prometheus.MustNewConstMetric(
			c.SoftLimit,
			prometheus.GaugeValue,
			boolToFloat(quota.SoftLimit),
			path,
			template,
		)
	}

	ch <- prometheus.MustNewConstMetric(
		c.QuotasCount,
		prometheus.GaugeValue,
		float64(count),
	)
	return nil, nil
}

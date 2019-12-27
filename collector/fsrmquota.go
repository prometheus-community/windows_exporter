package collector

import (
	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	Factories["fsrmquota"] = NewMSFT_FSRMQuotaCollector
}

// A MSFT_FSRMQuotaCollector is a Prometheus collector for WMI MSFT_FSRMQuota metrics
type MSFT_FSRMQuotaCollector struct {
	QuotasCount *prometheus.Desc
	Path      *prometheus.Desc
	PeakUsage *prometheus.Desc
	Size      *prometheus.Desc
	Usage     *prometheus.Desc
	
	
	Description           *prometheus.Desc
	Disabled              *prometheus.Desc
	MatchesTemplate       *prometheus.Desc
	SoftLimit             *prometheus.Desc
	Template              *prometheus.Desc
	Threshold             *prometheus.Desc
	
}

// NewMSFT_FSRMQuotaCollector ...
func NewMSFT_FSRMQuotaCollector() (Collector, error) {
	const subsystem = "fsrmquota"
	return &MSFT_FSRMQuotaCollector{
		QuotasCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "count"),
			"Number of Quotas",
			nil,
			nil,
		),
		PeakUsage: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "peak_usage_bytes"),
			"The highest amount of disk space usage charged to this quota. (PeakUsage)",
			[]string{"Path","Template"},
			nil,
		),
		Path: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "path"),
			"A string that represents a valid local path to a folder. (Path)",
			[]string{"Path","Template"},
			nil,
		),
		Size: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "size_bytes"),
			"The size of the quota. (Size)",
			[]string{"Path","Template"},
			nil,
		),
		Usage: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "usage"),
			"The current amount of disk space usage charged to this quota. (Usage)",
			[]string{"Path","Template"},
			nil,
		),
		Description: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "description"),
			"Description of the quota (Description)",
			[]string{"Path","Template","Description"},
			nil,
		),
		Disabled: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "disabled"),
			"If 1, the quota is disabled. The default value is 0. (Disabled)",
			[]string{"Path","Template"},
			nil,
		),
		SoftLimit: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "softlimit"),
			"If 1, the quota is a soft limit. If 0, the quota is a hard limit. The default value is 0. Optional (SoftLimit)",
			[]string{"Path","Template"},
			nil,
		),
		Template: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "template"),
			"Quota template name. (Template)",
			[]string{"Path","Template"},
			nil,
		),
		MatchesTemplate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "matchestemplate"),
			"If 1, the property values of this quota match those values of the template from which it was derived. (MatchesTemplate)",
			[]string{"Path","Template"},
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *MSFT_FSRMQuotaCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting msft_fsrmquota metrics:", desc, err)
		return err
	}
	return nil
}

// MSFT_FSRMQuota docs:
// https://docs.microsoft.com/en-us/windows-server/storage/fsrm/fsrm-overview
type MSFT_FSRMQuota struct {
	Name string

	Path      string
	PeakUsage uint64
	Size      uint64
	Usage     uint64
	Description           string
	Template              string
	//Threshold             string
	Disabled              bool
	MatchesTemplate       bool
	SoftLimit             bool

}

func (c *MSFT_FSRMQuotaCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []MSFT_FSRMQuota
	q := queryAll(&dst)
	
	var count int

	if err := wmi.QueryNamespace(q, &dst, "root/microsoft/windows/fsrm"); err != nil {
		return nil, err
	}
	
		for _, quota := range dst {
		
		count++
		Path := quota.Path
		Template := quota.Template
		Description := quota.Description
		
	ch <- prometheus.MustNewConstMetric(
		c.PeakUsage,
		prometheus.GaugeValue,
		float64(quota.PeakUsage),
		Path,
		Template,
	)

	ch <- prometheus.MustNewConstMetric(
		c.Size,
		prometheus.GaugeValue,
		float64(quota.Size),
		Path,
		Template,
	)

	ch <- prometheus.MustNewConstMetric(
		c.Usage,
		prometheus.GaugeValue,
		float64(quota.Usage),
		Path,
		Template,
	)
	ch <- prometheus.MustNewConstMetric(
		c.Description,
		prometheus.GaugeValue,
		1.0,
		Path,Template,Description,
	)
	ch <- prometheus.MustNewConstMetric(
		c.Template,
		prometheus.GaugeValue,
		1.0,
		Path,
		Template,
	)
	
	if quota.Disabled {
				ch <- prometheus.MustNewConstMetric(c.Disabled,
					prometheus.GaugeValue, 1.0, Path,Template)
	} else {
				ch <- prometheus.MustNewConstMetric(c.Disabled,
					prometheus.GaugeValue, 0.0, Path,Template)
	}
	if quota.MatchesTemplate {
				ch <- prometheus.MustNewConstMetric(
					c.MatchesTemplate,
					prometheus.GaugeValue,
					1.0,
					Path,
					Template,
				)
	} else {
				ch <- prometheus.MustNewConstMetric(
					c.MatchesTemplate,
					prometheus.GaugeValue,
					0.0,
					Path,
					Template,
				)
	}	
	if quota.SoftLimit {
				ch <- prometheus.MustNewConstMetric(
					c.SoftLimit,
					prometheus.GaugeValue,
					1.0,
					Path,
					Template,
					)
				
	} else {
				ch <- prometheus.MustNewConstMetric(
					c.SoftLimit,
					prometheus.GaugeValue,
					0.0,
					Path,
					Template,
				)
	}
	}
	
	ch <- prometheus.MustNewConstMetric(
		c.QuotasCount,
		prometheus.GaugeValue,
		float64(count),
	)
	return nil, nil
}
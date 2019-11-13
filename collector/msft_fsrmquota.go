package collector

import (
	//"strings"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	Factories["msft_fsrmquota"] = NewMSFT_FSRMQuotaCollector
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
	const subsystem = "msft_fsrmquota"
	return &MSFT_FSRMQuotaCollector{
		QuotasCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "count"),
			"Number of Quotas",
			nil,
			nil,
		),
		PeakUsage: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "peak_usage"),
			"The highest amount of disk space usage charged to this quota. (PeakUsage)",
			[]string{"quotaPath","quotaTemplate"},
			nil,
		),
		Path: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "path"),
			"A string that represents a valid local path to a folder. Must not exceed the value of MAX_PATH. Required. (Path)",
			[]string{"quotaPath","quotaTemplate"},
			nil,
		),
		Size: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "size"),
			"The size of the quota. If the Template property is not provided then the Size property must be provided (Size)",
			[]string{"quotaPath","quotaTemplate"},
			nil,
		),
		Usage: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "usage"),
			"The current amount of disk space usage charged to this quota. (Usage)",
			[]string{"quotaPath","quotaTemplate"},
			nil,
		),
		Description: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "description"),
			"A string up to 1KB in size. Optional. The default value is an empty string. (Description)",
			[]string{"quotaPath","quotaTemplate","quotaDescription"},
			nil,
		),
		Disabled: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "disabled"),
			"If True, the quota is disabled. The default value is False. (Disabled)",
			[]string{"quotaPath","quotaTemplate"},
			nil,
		),
		SoftLimit: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "softlimit"),
			"If True, the quota is a soft limit. If False, the quota is a hard limit. The default value is False. Optional (SoftLimit)",
			[]string{"quotaPath","quotaTemplate"},
			nil,
		),
		Template: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "template"),
			"A valid quota template name. Up to 1KB in size. Optional (Template)",
			[]string{"quotaPath","quotaTemplate"},
			nil,
		),
		MatchesTemplate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "matchestemplate"),
			"If True, the property values of this quota match those values of the template from which it was derived. (MatchesTemplate)",
			[]string{"quotaPath","quotaTemplate"},
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
// - <add link to documentation here>
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
		quotaPath := quota.Path
		quotaTemplate := quota.Template
		quotaDescription := quota.Description
		
	ch <- prometheus.MustNewConstMetric(
		c.PeakUsage,
		prometheus.GaugeValue,
		float64(quota.PeakUsage),
		quotaPath,
		quotaTemplate,
	)

	ch <- prometheus.MustNewConstMetric(
		c.Size,
		prometheus.GaugeValue,
		float64(quota.Size),
		quotaPath,
		quotaTemplate,
	)

	ch <- prometheus.MustNewConstMetric(
		c.Usage,
		prometheus.GaugeValue,
		float64(quota.Usage),
		quotaPath,
		quotaTemplate,
	)
	ch <- prometheus.MustNewConstMetric(
		c.Description,
		prometheus.GaugeValue,
		1.0,
		quotaPath,quotaTemplate,quotaDescription,
	)
	ch <- prometheus.MustNewConstMetric(
		c.Template,
		prometheus.GaugeValue,
		1.0,
		quotaPath,
		quotaTemplate,
	)
	
	if quota.Disabled {
				ch <- prometheus.MustNewConstMetric(c.Disabled,
					prometheus.GaugeValue, 1.0, quotaPath,quotaTemplate)
	} else {
				ch <- prometheus.MustNewConstMetric(c.Disabled,
					prometheus.GaugeValue, 0.0, quotaPath,quotaTemplate)
	}
	if quota.MatchesTemplate {
				ch <- prometheus.MustNewConstMetric(
					c.MatchesTemplate,
					prometheus.GaugeValue,
					1.0,
					quotaPath,
					quotaTemplate,
				)
	} else {
				ch <- prometheus.MustNewConstMetric(
					c.MatchesTemplate,
					prometheus.GaugeValue,
					0.0,
					quotaPath,
					quotaTemplate,
				)
	}	
	if quota.SoftLimit {
				ch <- prometheus.MustNewConstMetric(
					c.SoftLimit,
					prometheus.GaugeValue,
					1.0,
					quotaPath,
					quotaTemplate,
					)
				
	} else {
				ch <- prometheus.MustNewConstMetric(
					c.SoftLimit,
					prometheus.GaugeValue,
					0.0,
					quotaPath,
					quotaTemplate,
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

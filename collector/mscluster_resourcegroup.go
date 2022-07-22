package collector

import (
	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("mscluster_resourcegroup", newMSCluster_ResourceGroupCollector)
}

// A MSCluster_ResourceGroupCollector is a Prometheus collector for WMI MSCluster_ResourceGroup metrics
type MSCluster_ResourceGroupCollector struct {
	AutoFailbackType    *prometheus.Desc
	Characteristics     *prometheus.Desc
	ColdStartSetting    *prometheus.Desc
	DefaultOwner        *prometheus.Desc
	FailbackWindowEnd   *prometheus.Desc
	FailbackWindowStart *prometheus.Desc
	FailoverPeriod      *prometheus.Desc
	FailoverThreshold   *prometheus.Desc
	FaultDomain         *prometheus.Desc
	Flags               *prometheus.Desc
	GroupType           *prometheus.Desc
	PlacementOptions    *prometheus.Desc
	Priority            *prometheus.Desc
	ResiliencyPeriod    *prometheus.Desc
	State               *prometheus.Desc
}

func newMSCluster_ResourceGroupCollector() (Collector, error) {
	const subsystem = "mscluster_resourcegroup"
	return &MSCluster_ResourceGroupCollector{
		AutoFailbackType: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "auto_failback_type"),
			"Provides access to the group's AutoFailbackType property.",
			[]string{"name"},
			nil,
		),
		Characteristics: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "characteristics"),
			"Provides the characteristics of the group.",
			[]string{"name"},
			nil,
		),
		ColdStartSetting: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cold_start_setting"),
			"Indicates whether a group can start after a cluster cold start.",
			[]string{"name"},
			nil,
		),
		DefaultOwner: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "default_owner"),
			"Number of the last node the resource group was activated on or explicitly moved to.",
			[]string{"name"},
			nil,
		),
		FailbackWindowEnd: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "failback_window_end"),
			"The FailbackWindowEnd property provides the latest time that the group can be moved back to the node identified as its preferred node.",
			[]string{"name"},
			nil,
		),
		FailbackWindowStart: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "failback_window_start"),
			"The FailbackWindowStart property provides the earliest time (that is, local time as kept by the cluster) that the group can be moved back to the node identified as its preferred node.",
			[]string{"name"},
			nil,
		),
		FailoverPeriod: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "failover_period"),
			"The FailoverPeriod property specifies a number of hours during which a maximum number of failover attempts, specified by the FailoverThreshold property, can occur.",
			[]string{"name"},
			nil,
		),
		FailoverThreshold: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "failover_threshold"),
			"The FailoverThreshold property specifies the maximum number of failover attempts.",
			[]string{"name"},
			nil,
		),
		Flags: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "flags"),
			"Provides access to the flags set for the group. ",
			[]string{"name"},
			nil,
		),
		GroupType: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "group_type"),
			"The Type of the resource group.",
			[]string{"name"},
			nil,
		),
		Priority: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "priority"),
			"Priority value of the resource group",
			[]string{"name"},
			nil,
		),
		ResiliencyPeriod: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "resiliency_period"),
			"The resiliency period for this group, in seconds.",
			[]string{"name"},
			nil,
		),
		State: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "state"),
			"The current state of the resource group. -1: Unknown; 0: Online; 1: Offline; 2: Failed; 3: Partial Online; 4: Pending",
			[]string{"name"},
			nil,
		),
	}, nil
}

// MSCluster_ResourceGroup docs:
// - https://docs.microsoft.com/en-us/previous-versions/windows/desktop/cluswmi/mscluster-resourcegroup
//
type MSCluster_ResourceGroup struct {
	Name string

	AutoFailbackType    uint
	Characteristics     uint
	ColdStartSetting    uint
	DefaultOwner        uint
	FailbackWindowEnd   int
	FailbackWindowStart int
	FailoverPeriod      uint
	FailoverThreshold   uint
	Flags               uint
	GroupType           uint
	Priority            uint
	ResiliencyPeriod    uint
	State               uint
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *MSCluster_ResourceGroupCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	var dst []MSCluster_ResourceGroup
	q := queryAll(&dst)
	if err := wmi.QueryNamespace(q, &dst, "root/MSCluster"); err != nil {
		return err
	}

	for _, v := range dst {

		ch <- prometheus.MustNewConstMetric(
			c.AutoFailbackType,
			prometheus.GaugeValue,
			float64(v.AutoFailbackType),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Characteristics,
			prometheus.GaugeValue,
			float64(v.Characteristics),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ColdStartSetting,
			prometheus.GaugeValue,
			float64(v.ColdStartSetting),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DefaultOwner,
			prometheus.GaugeValue,
			float64(v.DefaultOwner),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FailbackWindowEnd,
			prometheus.GaugeValue,
			float64(v.FailbackWindowEnd),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FailbackWindowStart,
			prometheus.GaugeValue,
			float64(v.FailbackWindowStart),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FailoverPeriod,
			prometheus.GaugeValue,
			float64(v.FailoverPeriod),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FailoverThreshold,
			prometheus.GaugeValue,
			float64(v.FailoverThreshold),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Flags,
			prometheus.GaugeValue,
			float64(v.Flags),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GroupType,
			prometheus.GaugeValue,
			float64(v.GroupType),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Priority,
			prometheus.GaugeValue,
			float64(v.Priority),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ResiliencyPeriod,
			prometheus.GaugeValue,
			float64(v.ResiliencyPeriod),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.State,
			prometheus.GaugeValue,
			float64(v.State),
			v.Name,
		)

	}

	return nil
}

package mscluster

import (
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

const nameResourceGroup = Name + "_resourcegroup"

// msClusterResourceGroup represents the MSCluster_ResourceGroup WMI class
// - https://docs.microsoft.com/en-us/previous-versions/windows/desktop/cluswmi/mscluster-resourcegroup
type msClusterResourceGroup struct {
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
	OwnerNode           string
	Priority            uint
	ResiliencyPeriod    uint
	State               uint
}

func (c *Collector) buildResourceGroup() {
	c.resourceGroupAutoFailbackType = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResourceGroup, "auto_failback_type"),
		"Provides access to the group's AutoFailbackType property.",
		[]string{"name"},
		nil,
	)
	c.resourceGroupCharacteristics = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResourceGroup, "characteristics"),
		"Provides the characteristics of the group.",
		[]string{"name"},
		nil,
	)
	c.resourceGroupColdStartSetting = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResourceGroup, "cold_start_setting"),
		"Indicates whether a group can start after a cluster cold start.",
		[]string{"name"},
		nil,
	)
	c.resourceGroupDefaultOwner = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResourceGroup, "default_owner"),
		"Number of the last node the resource group was activated on or explicitly moved to.",
		[]string{"name"},
		nil,
	)
	c.resourceGroupFailbackWindowEnd = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResourceGroup, "failback_window_end"),
		"The FailbackWindowEnd property provides the latest time that the group can be moved back to the node identified as its preferred node.",
		[]string{"name"},
		nil,
	)
	c.resourceGroupFailbackWindowStart = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResourceGroup, "failback_window_start"),
		"The FailbackWindowStart property provides the earliest time (that is, local time as kept by the cluster) that the group can be moved back to the node identified as its preferred node.",
		[]string{"name"},
		nil,
	)
	c.resourceGroupFailOverPeriod = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResourceGroup, "failover_period"),
		"The FailoverPeriod property specifies a number of hours during which a maximum number of failover attempts, specified by the FailoverThreshold property, can occur.",
		[]string{"name"},
		nil,
	)
	c.resourceGroupFailOverThreshold = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResourceGroup, "failover_threshold"),
		"The FailoverThreshold property specifies the maximum number of failover attempts.",
		[]string{"name"},
		nil,
	)
	c.resourceGroupFlags = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResourceGroup, "flags"),
		"Provides access to the flags set for the group. ",
		[]string{"name"},
		nil,
	)
	c.resourceGroupGroupType = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResourceGroup, "group_type"),
		"The Type of the resource group.",
		[]string{"name"},
		nil,
	)
	c.resourceGroupOwnerNode = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResourceGroup, "owner_node"),
		"The node hosting the resource group. 0: Not hosted; 1: Hosted",
		[]string{"node_name", "name"},
		nil,
	)
	c.resourceGroupOwnerNode = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResourceGroup, "owner_node"),
		"The node hosting the resource group. 0: Not hosted; 1: Hosted",
		[]string{"node_name", "name"},
		nil,
	)
	c.resourceGroupPriority = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResourceGroup, "priority"),
		"Priority value of the resource group",
		[]string{"name"},
		nil,
	)
	c.resourceGroupResiliencyPeriod = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResourceGroup, "resiliency_period"),
		"The resiliency period for this group, in seconds.",
		[]string{"name"},
		nil,
	)
	c.resourceGroupState = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResourceGroup, "state"),
		"The current state of the resource group. -1: Unknown; 0: Online; 1: Offline; 2: Failed; 3: Partial Online; 4: Pending",
		[]string{"name"},
		nil,
	)
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) collectResourceGroup(ch chan<- prometheus.Metric, nodeNames []string) error {
	var dst []msClusterResourceGroup

	if err := c.wmiClient.Query("SELECT * FROM MSCluster_ResourceGroup", &dst, nil, "root/MSCluster"); err != nil {
		return err
	}

	for _, v := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.resourceGroupAutoFailbackType,
			prometheus.GaugeValue,
			float64(v.AutoFailbackType),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.resourceGroupCharacteristics,
			prometheus.GaugeValue,
			float64(v.Characteristics),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.resourceGroupColdStartSetting,
			prometheus.GaugeValue,
			float64(v.ColdStartSetting),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.resourceGroupDefaultOwner,
			prometheus.GaugeValue,
			float64(v.DefaultOwner),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.resourceGroupFailbackWindowEnd,
			prometheus.GaugeValue,
			float64(v.FailbackWindowEnd),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.resourceGroupFailbackWindowStart,
			prometheus.GaugeValue,
			float64(v.FailbackWindowStart),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.resourceGroupFailOverPeriod,
			prometheus.GaugeValue,
			float64(v.FailoverPeriod),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.resourceGroupFailOverThreshold,
			prometheus.GaugeValue,
			float64(v.FailoverThreshold),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.resourceGroupFlags,
			prometheus.GaugeValue,
			float64(v.Flags),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.resourceGroupGroupType,
			prometheus.GaugeValue,
			float64(v.GroupType),
			v.Name,
		)

		for _, nodeName := range nodeNames {
			isCurrentState := 0.0
			if v.OwnerNode == nodeName {
				isCurrentState = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				c.resourceGroupOwnerNode,
				prometheus.GaugeValue,
				isCurrentState,
				nodeName, v.Name,
			)
		}

		ch <- prometheus.MustNewConstMetric(
			c.resourceGroupPriority,
			prometheus.GaugeValue,
			float64(v.Priority),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.resourceGroupResiliencyPeriod,
			prometheus.GaugeValue,
			float64(v.ResiliencyPeriod),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.resourceGroupState,
			prometheus.GaugeValue,
			float64(v.State),
			v.Name,
		)
	}

	return nil
}

//go:build windows

package mscluster

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
)

const nameNetwork = Name + "_network"

type collectorNetwork struct {
	networkCharacteristics *prometheus.Desc
	networkFlags           *prometheus.Desc
	networkMetric          *prometheus.Desc
	networkRole            *prometheus.Desc
	networkState           *prometheus.Desc
}

// msClusterNetwork represents the MSCluster_Network WMI class
// - https://docs.microsoft.com/en-us/previous-versions/windows/desktop/cluswmi/mscluster-network
type msClusterNetwork struct {
	Name string `mi:"Name"`

	Characteristics uint `mi:"Characteristics"`
	Flags           uint `mi:"Flags"`
	Metric          uint `mi:"Metric"`
	Role            uint `mi:"Role"`
	State           uint `mi:"State"`
}

func (c *Collector) buildNetwork() {
	c.networkCharacteristics = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameNetwork, "characteristics"),
		"Provides the characteristics of the network.",
		[]string{"name"},
		nil,
	)
	c.networkFlags = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameNetwork, "flags"),
		"Provides access to the flags set for the node. ",
		[]string{"name"},
		nil,
	)
	c.networkMetric = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameNetwork, "metric"),
		"The metric of a cluster network (networks with lower values are used first). If this value is set, then the AutoMetric property is set to false.",
		[]string{"name"},
		nil,
	)
	c.networkRole = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameNetwork, "role"),
		"Provides access to the network's Role property. The Role property describes the role of the network in the cluster. 0: None; 1: Cluster; 2: Client; 3: Both ",
		[]string{"name"},
		nil,
	)
	c.networkState = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameNetwork, "state"),
		"Provides the current state of the network. 1-1: Unknown; 0: Unavailable; 1: Down; 2: Partitioned; 3: Up",
		[]string{"name"},
		nil,
	)
}

// Collect sends the metric values for each metric
// to the provided prometheus metric channel.
func (c *Collector) collectNetwork(ch chan<- prometheus.Metric) error {
	var dst []msClusterNetwork

	if err := c.miSession.Query(&dst, mi.NamespaceRootMSCluster, utils.Must(mi.NewQuery("SELECT * MSCluster_Node"))); err != nil {
		return fmt.Errorf("WMI query failed: %w", err)
	}

	for _, v := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.networkCharacteristics,
			prometheus.GaugeValue,
			float64(v.Characteristics),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.networkFlags,
			prometheus.GaugeValue,
			float64(v.Flags),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.networkMetric,
			prometheus.GaugeValue,
			float64(v.Metric),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.networkRole,
			prometheus.GaugeValue,
			float64(v.Role),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.networkState,
			prometheus.GaugeValue,
			float64(v.State),
			v.Name,
		)
	}

	return nil
}

//go:build windows

package mscluster

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
)

const nameNode = Name + "_node"

type collectorNode struct {
	nodeBuildNumber           *prometheus.Desc
	nodeCharacteristics       *prometheus.Desc
	nodeDetectedCloudPlatform *prometheus.Desc
	nodeDynamicWeight         *prometheus.Desc
	nodeFlags                 *prometheus.Desc
	nodeMajorVersion          *prometheus.Desc
	nodeMinorVersion          *prometheus.Desc
	nodeNeedsPreventQuorum    *prometheus.Desc
	nodeNodeDrainStatus       *prometheus.Desc
	nodeNodeHighestVersion    *prometheus.Desc
	nodeNodeLowestVersion     *prometheus.Desc
	nodeNodeWeight            *prometheus.Desc
	nodeState                 *prometheus.Desc
	nodeStatusInformation     *prometheus.Desc
}

// msClusterNode represents the MSCluster_Node WMI class
// - https://docs.microsoft.com/en-us/previous-versions/windows/desktop/cluswmi/mscluster-node
type msClusterNode struct {
	Name string `mi:"Name"`

	BuildNumber           uint `mi:"BuildNumber"`
	Characteristics       uint `mi:"Characteristics"`
	DetectedCloudPlatform uint `mi:"DetectedCloudPlatform"`
	DynamicWeight         uint `mi:"DynamicWeight"`
	Flags                 uint `mi:"Flags"`
	MajorVersion          uint `mi:"MajorVersion"`
	MinorVersion          uint `mi:"MinorVersion"`
	NeedsPreventQuorum    uint `mi:"NeedsPreventQuorum"`
	NodeDrainStatus       uint `mi:"NodeDrainStatus"`
	NodeHighestVersion    uint `mi:"NodeHighestVersion"`
	NodeLowestVersion     uint `mi:"NodeLowestVersion"`
	NodeWeight            uint `mi:"NodeWeight"`
	State                 uint `mi:"State"`
	StatusInformation     uint `mi:"StatusInformation"`
}

func (c *Collector) buildNode() {
	c.nodeBuildNumber = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameNode, "build_number"),
		"Provides access to the node's BuildNumber property.",
		[]string{"name"},
		nil,
	)
	c.nodeCharacteristics = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameNode, "characteristics"),
		"Provides access to the characteristics set for the node.",
		[]string{"name"},
		nil,
	)
	c.nodeDetectedCloudPlatform = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameNode, "detected_cloud_platform"),
		"(DetectedCloudPlatform)",
		[]string{"name"},
		nil,
	)
	c.nodeDynamicWeight = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameNode, "dynamic_weight"),
		"The dynamic vote weight of the node adjusted by dynamic quorum feature.",
		[]string{"name"},
		nil,
	)
	c.nodeFlags = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameNode, "flags"),
		"Provides access to the flags set for the node.",
		[]string{"name"},
		nil,
	)
	c.nodeMajorVersion = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameNode, "major_version"),
		"Provides access to the node's MajorVersion property, which specifies the major portion of the Windows version installed.",
		[]string{"name"},
		nil,
	)
	c.nodeMinorVersion = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameNode, "minor_version"),
		"Provides access to the node's MinorVersion property, which specifies the minor portion of the Windows version installed.",
		[]string{"name"},
		nil,
	)
	c.nodeNeedsPreventQuorum = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameNode, "needs_prevent_quorum"),
		"Whether the cluster service on that node should be started with prevent quorum flag.",
		[]string{"name"},
		nil,
	)
	c.nodeNodeDrainStatus = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameNode, "node_drain_status"),
		"The current node drain status of a node. 0: Not Initiated; 1: In Progress; 2: Completed; 3: Failed",
		[]string{"name"},
		nil,
	)
	c.nodeNodeHighestVersion = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameNode, "node_highest_version"),
		"Provides access to the node's NodeHighestVersion property, which specifies the highest possible version of the cluster service with which the node can join or communicate.",
		[]string{"name"},
		nil,
	)
	c.nodeNodeLowestVersion = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameNode, "node_lowest_version"),
		"Provides access to the node's NodeLowestVersion property, which specifies the lowest possible version of the cluster service with which the node can join or communicate.",
		[]string{"name"},
		nil,
	)
	c.nodeNodeWeight = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameNode, "node_weight"),
		"The vote weight of the node.",
		[]string{"name"},
		nil,
	)
	c.nodeState = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameNode, "state"),
		"Returns the current state of a node. -1: Unknown; 0: Up; 1: Down; 2: Paused; 3: Joining",
		[]string{"name"},
		nil,
	)
	c.nodeStatusInformation = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameNode, "status_information"),
		"The isolation or quarantine status of the node.",
		[]string{"name"},
		nil,
	)
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) collectNode(ch chan<- prometheus.Metric) ([]string, error) {
	var dst []msClusterNode

	if err := c.miSession.Query(&dst, mi.NamespaceRootMSCluster, utils.Must(mi.NewQuery("SELECT * FROM MSCluster_Node"))); err != nil {
		return nil, fmt.Errorf("WMI query failed: %w", err)
	}

	nodeNames := make([]string, 0, len(dst))

	for _, v := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.nodeBuildNumber,
			prometheus.GaugeValue,
			float64(v.BuildNumber),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.nodeCharacteristics,
			prometheus.GaugeValue,
			float64(v.Characteristics),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.nodeDetectedCloudPlatform,
			prometheus.GaugeValue,
			float64(v.DetectedCloudPlatform),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.nodeDynamicWeight,
			prometheus.GaugeValue,
			float64(v.DynamicWeight),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.nodeFlags,
			prometheus.GaugeValue,
			float64(v.Flags),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.nodeMajorVersion,
			prometheus.GaugeValue,
			float64(v.MajorVersion),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.nodeMinorVersion,
			prometheus.GaugeValue,
			float64(v.MinorVersion),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.nodeNeedsPreventQuorum,
			prometheus.GaugeValue,
			float64(v.NeedsPreventQuorum),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.nodeNodeDrainStatus,
			prometheus.GaugeValue,
			float64(v.NodeDrainStatus),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.nodeNodeHighestVersion,
			prometheus.GaugeValue,
			float64(v.NodeHighestVersion),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.nodeNodeLowestVersion,
			prometheus.GaugeValue,
			float64(v.NodeLowestVersion),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.nodeNodeWeight,
			prometheus.GaugeValue,
			float64(v.NodeWeight),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.nodeState,
			prometheus.GaugeValue,
			float64(v.State),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.nodeStatusInformation,
			prometheus.GaugeValue,
			float64(v.StatusInformation),
			v.Name,
		)

		nodeNames = append(nodeNames, v.Name)
	}

	return nodeNames, nil
}

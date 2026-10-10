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

package mscluster

import (
	"time"

	"github.com/prometheus-community/windows_exporter/internal/headers/clusapi"
	"github.com/prometheus-community/windows_exporter/internal/osversion"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const nameNode = Name + "_node"

type nodeSource interface {
	Nodes(deadline time.Time) ([]clusapi.Object, error)
	Close() error
}

type collectorNode struct {
	nodeSource nodeSource

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

// msClusterNode represents the MSCluster_Node WMI class. The collector reads
// ClusAPI; the parity test compares against this WMI model.
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

func (c *Collector) buildNode() error {
	source, err := clusapi.Open()
	if err != nil {
		return err
	}

	c.nodeSource = source
	c.buildNodeDescriptors()

	return nil
}

func (c *Collector) buildNodeDescriptors() {
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
// to the provided prometheus Metric channel. It returns the node names that
// the resource and resource group subcollectors use for owner_node.
func (c *Collector) collectNode(ch chan<- prometheus.Metric, maxScrapeDuration time.Duration) ([]string, error) {
	var deadline time.Time
	if maxScrapeDuration > 0 {
		deadline = time.Now().Add(maxScrapeDuration)
	}

	nodes, resultErr := c.nodeSource.Nodes(deadline)

	return c.publishNodes(ch, nodes, osversion.Build(), resultErr)
}

func (c *Collector) publishNodes(ch chan<- prometheus.Metric, nodes []clusapi.Object, build uint16, resultErr error) ([]string, error) {
	fields := []objectField{
		{name: "BuildNumber", desc: c.nodeBuildNumber},
		{name: "Characteristics", desc: c.nodeCharacteristics},
		// The previous WMI query selected DetectedCloudPlatform only on Windows
		// Server 2022 and newer and published 0 on older builds.
		{name: "DetectedCloudPlatform", desc: c.nodeDetectedCloudPlatform, minBuild: osversion.LTSC2022, older: zeroIfMissing},
		{name: "DynamicWeight", desc: c.nodeDynamicWeight},
		{name: "Flags", desc: c.nodeFlags},
		{name: "MajorVersion", desc: c.nodeMajorVersion},
		{name: "MinorVersion", desc: c.nodeMinorVersion},
		{name: "NeedsPreventQuorum", desc: c.nodeNeedsPreventQuorum},
		{name: "NodeDrainStatus", desc: c.nodeNodeDrainStatus},
		{name: "NodeHighestVersion", desc: c.nodeNodeHighestVersion},
		{name: "NodeLowestVersion", desc: c.nodeNodeLowestVersion},
		{name: "NodeWeight", desc: c.nodeNodeWeight},
		{name: "State", desc: c.nodeState},
		{name: "StatusInformation", desc: c.nodeStatusInformation, minBuild: osversion.LTSC2016},
	}

	nodeNames := make([]string, 0, len(nodes))

	for _, node := range nodes {
		if node.Name == "" {
			continue
		}

		resultErr = publishObjectFields(ch, "node", node, fields, build, resultErr)
		nodeNames = append(nodeNames, node.Name)
	}

	return nodeNames, resultErr
}

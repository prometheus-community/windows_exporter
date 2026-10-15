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

const nameResourceGroup = Name + "_resourcegroup"

type resourceGroupSource interface {
	Groups(deadline time.Time) ([]clusapi.Object, error)
	Close() error
}

type collectorResourceGroup struct {
	resourceGroupSource resourceGroupSource

	// resourceGroupMetricCache reuses the field and owner node metrics of
	// unchanged groups; see objectMetricCache.
	resourceGroupMetricCache objectMetricCache

	resourceGroupAutoFailbackType    *prometheus.Desc
	resourceGroupCharacteristics     *prometheus.Desc
	resourceGroupColdStartSetting    *prometheus.Desc
	resourceGroupDefaultOwner        *prometheus.Desc
	resourceGroupFailbackWindowEnd   *prometheus.Desc
	resourceGroupFailbackWindowStart *prometheus.Desc
	resourceGroupFailOverPeriod      *prometheus.Desc
	resourceGroupFailOverThreshold   *prometheus.Desc
	resourceGroupFlags               *prometheus.Desc
	resourceGroupGroupType           *prometheus.Desc
	resourceGroupOwnerNode           *prometheus.Desc
	resourceGroupPriority            *prometheus.Desc
	resourceGroupResiliencyPeriod    *prometheus.Desc
	resourceGroupState               *prometheus.Desc
}

// msClusterResourceGroup represents the MSCluster_ResourceGroup WMI class. The
// collector reads ClusAPI; the parity test compares against this WMI model.
// - https://docs.microsoft.com/en-us/previous-versions/windows/desktop/cluswmi/mscluster-resourcegroup
type msClusterResourceGroup struct {
	Name string `mi:"Name"`

	AutoFailbackType    uint   `mi:"AutoFailbackType"`
	Characteristics     uint   `mi:"Characteristics"`
	ColdStartSetting    uint   `mi:"ColdStartSetting"`
	DefaultOwner        uint   `mi:"DefaultOwner"`
	FailbackWindowEnd   int    `mi:"FailbackWindowEnd"`
	FailbackWindowStart int    `mi:"FailbackWindowStart"`
	FailoverPeriod      uint   `mi:"FailoverPeriod"`
	FailoverThreshold   uint   `mi:"FailoverThreshold"`
	Flags               uint   `mi:"Flags"`
	GroupType           uint   `mi:"GroupType"`
	OwnerNode           string `mi:"OwnerNode"`
	Priority            uint   `mi:"Priority"`
	ResiliencyPeriod    uint   `mi:"ResiliencyPeriod"`
	State               uint   `mi:"State"`
}

func (c *Collector) buildResourceGroup() error {
	source, err := clusapi.Open()
	if err != nil {
		return err
	}

	c.resourceGroupSource = source
	c.buildResourceGroupDescriptors()

	return nil
}

func (c *Collector) buildResourceGroupDescriptors() {
	c.resourceGroupAutoFailbackType = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResourceGroup, "auto_failback_type"),
		"Provides access to the group's AutoFailbackType property.",
		[]string{"resource_group"},
		nil,
	)
	c.resourceGroupCharacteristics = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResourceGroup, "characteristics"),
		"Provides the characteristics of the group.",
		[]string{"resource_group"},
		nil,
	)
	c.resourceGroupColdStartSetting = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResourceGroup, "cold_start_setting"),
		"Indicates whether a group can start after a cluster cold start.",
		[]string{"resource_group"},
		nil,
	)
	c.resourceGroupDefaultOwner = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResourceGroup, "default_owner"),
		"Number of the last node the resource group was activated on or explicitly moved to.",
		[]string{"resource_group"},
		nil,
	)
	c.resourceGroupFailbackWindowEnd = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResourceGroup, "failback_window_end"),
		"The FailbackWindowEnd property provides the latest time that the group can be moved back to the node identified as its preferred node.",
		[]string{"resource_group"},
		nil,
	)
	c.resourceGroupFailbackWindowStart = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResourceGroup, "failback_window_start"),
		"The FailbackWindowStart property provides the earliest time (that is, local time as kept by the cluster) that the group can be moved back to the node identified as its preferred node.",
		[]string{"resource_group"},
		nil,
	)
	c.resourceGroupFailOverPeriod = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResourceGroup, "failover_period_seconds"),
		"The FailoverPeriod property specifies a number of seconds during which a maximum number of failover attempts, specified by the FailoverThreshold property, can occur. Windows stores it in hours.",
		[]string{"resource_group"},
		nil,
	)
	c.resourceGroupFailOverThreshold = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResourceGroup, "failover_threshold"),
		"The FailoverThreshold property specifies the maximum number of failover attempts.",
		[]string{"resource_group"},
		nil,
	)
	c.resourceGroupFlags = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResourceGroup, "flags"),
		"Provides access to the flags set for the group. ",
		[]string{"resource_group"},
		nil,
	)
	c.resourceGroupGroupType = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResourceGroup, "group_type"),
		"The Type of the resource group.",
		[]string{"resource_group"},
		nil,
	)
	c.resourceGroupOwnerNode = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResourceGroup, "owner_node"),
		"The node hosting the resource group. 0: Not hosted; 1: Hosted",
		[]string{"node", "resource_group"},
		nil,
	)
	c.resourceGroupPriority = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResourceGroup, "priority"),
		"Priority value of the resource group",
		[]string{"resource_group"},
		nil,
	)
	c.resourceGroupResiliencyPeriod = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResourceGroup, "resiliency_period_seconds"),
		"The resiliency period for this group, in seconds.",
		[]string{"resource_group"},
		nil,
	)
	c.resourceGroupState = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResourceGroup, "state"),
		"The current state of the resource group. -1: Unknown; 0: Online; 1: Offline; 2: Failed; 3: Partial Online; 4: Pending",
		[]string{"resource_group"},
		nil,
	)

	// Cached metrics refer to the descriptors of an earlier build.
	c.resourceGroupMetricCache.Reset()
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) collectResourceGroup(ch chan<- prometheus.Metric, maxScrapeDuration time.Duration, nodeNames []string) error {
	if c.resourceGroupSource == nil {
		return errNotBuilt(subCollectorResourceGroup)
	}

	groups, resultErr := callSource(maxScrapeDuration, c.resourceGroupSource.Groups)

	return c.publishResourceGroups(ch, groups, nodeNames, osversion.Build(), resultErr)
}

func (c *Collector) publishResourceGroups(ch chan<- prometheus.Metric, groups []clusapi.Object, nodeNames []string, build uint16, resultErr error) error {
	fields := []objectField{
		{name: "AutoFailbackType", desc: c.resourceGroupAutoFailbackType},
		{name: "Characteristics", desc: c.resourceGroupCharacteristics},
		{name: "ColdStartSetting", desc: c.resourceGroupColdStartSetting, minBuild: osversion.LTSC2016, older: omitIfMissing},
		{name: "DefaultOwner", desc: c.resourceGroupDefaultOwner},
		{name: "FailbackWindowEnd", desc: c.resourceGroupFailbackWindowEnd, signed: true},
		{name: "FailbackWindowStart", desc: c.resourceGroupFailbackWindowStart, signed: true},
		{name: "FailoverPeriod", desc: c.resourceGroupFailOverPeriod, scale: 3600},
		{name: "FailoverThreshold", desc: c.resourceGroupFailOverThreshold},
		{name: "Flags", desc: c.resourceGroupFlags},
		{name: "GroupType", desc: c.resourceGroupGroupType},
		{name: "Priority", desc: c.resourceGroupPriority},
		{name: "ResiliencyPeriod", desc: c.resourceGroupResiliencyPeriod, minBuild: osversion.LTSC2016},
		{name: "State", desc: c.resourceGroupState},
	}

	scrape := c.resourceGroupMetricCache.Begin()

	var nodeNamesKey []byte
	for _, nodeName := range nodeNames {
		nodeNamesKey = appendKeyString(nodeNamesKey, nodeName)
	}

	var key []byte

	for _, group := range groups {
		if group.Name == "" {
			continue
		}

		var ok bool

		key, ok = appendObjectFieldsKey(key[:0], group, fields, build)
		if !ok {
			// A group with missing fields is not cached, so that it reports
			// its errors on every scrape.
			resultErr = publishObjectFields(ch, "group", group, fields, build, resultErr)

			for _, metric := range c.resourceGroupOwnerNodeMetrics(nil, group, nodeNames) {
				ch <- metric
			}

			continue
		}

		if group.OwnerNodeValid {
			key = append(key, 1)
			key = appendKeyString(key, group.OwnerNode)
			key = append(key, nodeNamesKey...)
		} else {
			key = append(key, 0)
		}

		metrics := scrape.Get(group.Name, string(key), func() []prometheus.Metric {
			metrics := objectFieldMetrics(group, fields, build, len(nodeNames))

			return c.resourceGroupOwnerNodeMetrics(metrics, group, nodeNames)
		})

		for _, metric := range metrics {
			ch <- metric
		}
	}

	commitScrape(scrape, len(groups), resultErr)

	return resultErr
}

// resourceGroupOwnerNodeMetrics appends the owner node metrics of a group, one
// per node, to metrics. A group without a valid owner node has none.
func (c *Collector) resourceGroupOwnerNodeMetrics(metrics []prometheus.Metric, group clusapi.Object, nodeNames []string) []prometheus.Metric {
	if !group.OwnerNodeValid {
		return metrics
	}

	for _, nodeName := range nodeNames {
		isCurrentState := 0.0
		if group.OwnerNode == nodeName {
			isCurrentState = 1.0
		}

		metrics = append(metrics, prometheus.MustNewConstMetric(c.resourceGroupOwnerNode, prometheus.GaugeValue, isCurrentState, nodeName, group.Name))
	}

	return metrics
}

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
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const nameNetwork = Name + "_network"

type networkSource interface {
	Networks(deadline time.Time) ([]clusapi.Object, error)
	Close() error
}

type collectorNetwork struct {
	networkSource networkSource

	networkCharacteristics *prometheus.Desc
	networkFlags           *prometheus.Desc
	networkMetric          *prometheus.Desc
	networkRole            *prometheus.Desc
	networkState           *prometheus.Desc
}

// msClusterNetwork represents the MSCluster_Network WMI class. The collector
// reads ClusAPI; the parity test compares against this WMI model.
// - https://docs.microsoft.com/en-us/previous-versions/windows/desktop/cluswmi/mscluster-network
type msClusterNetwork struct {
	Name string `mi:"Name"`

	Characteristics uint `mi:"Characteristics"`
	Flags           uint `mi:"Flags"`
	Metric          uint `mi:"Metric"`
	Role            uint `mi:"Role"`
	State           uint `mi:"State"`
}

func (c *Collector) buildNetwork() error {
	source, err := clusapi.Open()
	if err != nil {
		return err
	}

	c.networkSource = source
	c.buildNetworkDescriptors()

	return nil
}

func (c *Collector) buildNetworkDescriptors() {
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
func (c *Collector) collectNetwork(ch chan<- prometheus.Metric, maxScrapeDuration time.Duration) error {
	var deadline time.Time
	if maxScrapeDuration > 0 {
		deadline = time.Now().Add(maxScrapeDuration)
	}

	networks, resultErr := c.networkSource.Networks(deadline)

	return c.publishNetworks(ch, networks, resultErr)
}

func (c *Collector) publishNetworks(ch chan<- prometheus.Metric, networks []clusapi.Object, resultErr error) error {
	fields := []objectField{
		{name: "Characteristics", desc: c.networkCharacteristics},
		{name: "Flags", desc: c.networkFlags},
		{name: "Metric", desc: c.networkMetric},
		{name: "Role", desc: c.networkRole},
		{name: "State", desc: c.networkState},
	}

	for _, network := range networks {
		if network.Name == "" {
			continue
		}

		// All network properties exist since Windows Server 2012, so the
		// build does not matter.
		resultErr = publishObjectFields(ch, "network", network, fields, 0, resultErr)
	}

	return resultErr
}

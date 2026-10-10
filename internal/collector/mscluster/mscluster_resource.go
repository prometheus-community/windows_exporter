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
	"errors"
	"fmt"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/headers/clusapi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const nameResource = Name + "_resource"

type resourceSource interface {
	Resources(deadline time.Time) ([]clusapi.Resource, error)
	Close() error
}

type collectorResource struct {
	resourceSource resourceSource

	resourceCharacteristics        *prometheus.Desc
	resourceDeadlockTimeout        *prometheus.Desc
	resourceEmbeddedFailureAction  *prometheus.Desc
	resourceFlags                  *prometheus.Desc
	resourceIsAlivePollInterval    *prometheus.Desc
	resourceLooksAlivePollInterval *prometheus.Desc
	resourceMonitorProcessId       *prometheus.Desc
	resourceOwnerNode              *prometheus.Desc
	resourcePendingTimeout         *prometheus.Desc
	resourceResourceClass          *prometheus.Desc
	resourceRestartAction          *prometheus.Desc
	resourceRestartDelay           *prometheus.Desc
	resourceRestartPeriod          *prometheus.Desc
	resourceRestartThreshold       *prometheus.Desc
	resourceRetryPeriodOnFailure   *prometheus.Desc
	resourceState                  *prometheus.Desc
	resourceSubClass               *prometheus.Desc
}

// msClusterResource represents the MSCluster_Resource WMI class
// - https://docs.microsoft.com/en-us/previous-versions/windows/desktop/cluswmi/mscluster-resource
type msClusterResource struct {
	Name       string `mi:"Name"`
	Type       string `mi:"Type"`
	OwnerGroup string `mi:"OwnerGroup"`
	OwnerNode  string `mi:"OwnerNode"`

	Characteristics        uint `mi:"Characteristics"`
	DeadlockTimeout        uint `mi:"DeadlockTimeout"`
	EmbeddedFailureAction  uint `mi:"EmbeddedFailureAction"`
	Flags                  uint `mi:"Flags"`
	IsAlivePollInterval    uint `mi:"IsAlivePollInterval"`
	LooksAlivePollInterval uint `mi:"LooksAlivePollInterval"`
	MonitorProcessId       uint `mi:"MonitorProcessId"`
	PendingTimeout         uint `mi:"PendingTimeout"`
	ResourceClass          uint `mi:"ResourceClass"`
	RestartAction          uint `mi:"RestartAction"`
	RestartDelay           uint `mi:"RestartDelay"`
	RestartPeriod          uint `mi:"RestartPeriod"`
	RestartThreshold       uint `mi:"RestartThreshold"`
	RetryPeriodOnFailure   uint `mi:"RetryPeriodOnFailure"`
	State                  uint `mi:"State"`
	Subclass               uint `mi:"Subclass"`
}

func (c *Collector) buildResource() error {
	source, err := clusapi.Open()
	if err != nil {
		return err
	}

	c.resourceSource = source
	c.buildResourceDescriptors()

	return nil
}

func (c *Collector) buildResourceDescriptors() {
	c.resourceCharacteristics = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResource, "characteristics"),
		"Provides the characteristics of the object.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.resourceDeadlockTimeout = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResource, "deadlock_timeout"),
		"Indicates the length of time to wait, in milliseconds, before declaring a deadlock in any call into a resource.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.resourceEmbeddedFailureAction = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResource, "embedded_failure_action"),
		"The time, in milliseconds, that a resource should remain in a failed state before the Cluster service attempts to restart it.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.resourceFlags = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResource, "flags"),
		"Provides access to the flags set for the object.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.resourceIsAlivePollInterval = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResource, "is_alive_poll_interval"),
		"Provides access to the resource's IsAlivePollInterval property, which is the recommended interval in milliseconds at which the Cluster Service should poll the resource to determine whether it is operational. If the property is set to 0xFFFFFFFF, the Cluster Service uses the IsAlivePollInterval property for the resource type associated with the resource.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.resourceLooksAlivePollInterval = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResource, "looks_alive_poll_interval"),
		"Provides access to the resource's LooksAlivePollInterval property, which is the recommended interval in milliseconds at which the Cluster Service should poll the resource to determine whether it appears operational. If the property is set to 0xFFFFFFFF, the Cluster Service uses the LooksAlivePollInterval property for the resource type associated with the resource.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.resourceMonitorProcessId = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResource, "monitor_process_id"),
		"Provides the process ID of the resource host service that is currently hosting the resource.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.resourceOwnerNode = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResource, "owner_node"),
		"The node hosting the resource. 0: Not hosted; 1: Hosted",
		[]string{"type", "owner_group", "node_name", "name"},
		nil,
	)
	c.resourceOwnerNode = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResource, "owner_node"),
		"The node hosting the resource. 0: Not hosted; 1: Hosted",
		[]string{"type", "owner_group", "node_name", "name"},
		nil,
	)
	c.resourcePendingTimeout = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResource, "pending_timeout"),
		"Provides access to the resource's PendingTimeout property. If a resource cannot be brought online or taken offline in the number of milliseconds specified by the PendingTimeout property, the resource is forcibly terminated.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.resourceResourceClass = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResource, "resource_class"),
		"Gets or sets the resource class of a resource. 0: Unknown; 1: Storage; 2: Network; 32768: Unknown ",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.resourceRestartAction = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResource, "restart_action"),
		"Provides access to the resource's RestartAction property, which is the action to be taken by the Cluster Service if the resource fails.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.resourceRestartDelay = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResource, "restart_delay"),
		"Indicates the time delay before a failed resource is restarted.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.resourceRestartPeriod = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResource, "restart_period"),
		"Provides access to the resource's RestartPeriod property, which is interval of time, in milliseconds, during which a specified number of restart attempts can be made on a nonresponsive resource.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.resourceRestartThreshold = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResource, "restart_threshold"),
		"Provides access to the resource's RestartThreshold property which is the maximum number of restart attempts that can be made on a resource within an interval defined by the RestartPeriod property before the Cluster Service initiates the action specified by the RestartAction property.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.resourceRetryPeriodOnFailure = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResource, "retry_period_on_failure"),
		"Provides access to the resource's RetryPeriodOnFailure property, which is the interval of time (in milliseconds) that a resource should remain in a failed state before the Cluster service attempts to restart it.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.resourceState = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResource, "state"),
		"The current state of the resource. -1: Unknown; 0: Inherited; 1: Initializing; 2: Online; 3: Offline; 4: Failed; 128: Pending; 129: Online Pending; 130: Offline Pending ",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.resourceSubClass = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameResource, "subclass"),
		"Provides the list of references to nodes that can be the owner of this resource.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) collectResource(ch chan<- prometheus.Metric, maxScrapeDuration time.Duration, nodeNames []string) error {
	if c.resourceSource == nil {
		return errNotBuilt(subCollectorResource)
	}

	resources, resultErr := callSource(maxScrapeDuration, c.resourceSource.Resources)

	return c.publishResources(ch, resources, nodeNames, resultErr)
}

func (c *Collector) publishResources(ch chan<- prometheus.Metric, resources []clusapi.Resource, nodeNames []string, resultErr error) error {
	for _, resource := range resources {
		if !resource.IdentityValid {
			continue
		}

		fields := []struct {
			name string
			desc *prometheus.Desc
		}{
			{"Characteristics", c.resourceCharacteristics},
			{"DeadlockTimeout", c.resourceDeadlockTimeout},
			{"EmbeddedFailureAction", c.resourceEmbeddedFailureAction},
			{"Flags", c.resourceFlags},
			{"IsAlivePollInterval", c.resourceIsAlivePollInterval},
			{"LooksAlivePollInterval", c.resourceLooksAlivePollInterval},
			{"MonitorProcessId", c.resourceMonitorProcessId},
			{"PendingTimeout", c.resourcePendingTimeout},
			{"ResourceClass", c.resourceResourceClass},
			{"RestartAction", c.resourceRestartAction},
			{"RestartDelay", c.resourceRestartDelay},
			{"RestartPeriod", c.resourceRestartPeriod},
			{"RestartThreshold", c.resourceRestartThreshold},
			{"RetryPeriodOnFailure", c.resourceRetryPeriodOnFailure},
			{"State", c.resourceState},
			{"Subclass", c.resourceSubClass},
		}
		for _, field := range fields {
			value, exists := resource.Values[field.name]
			if !exists {
				resultErr = errors.Join(resultErr, fmt.Errorf("resource %q: missing DWORD property %s", resource.Name, field.name))

				continue
			}

			ch <- prometheus.MustNewConstMetric(field.desc, prometheus.GaugeValue, float64(value), resource.Type, resource.OwnerGroup, resource.Name)
		}

		for _, nodeName := range nodeNames {
			value := 0.0
			if resource.OwnerNode == nodeName {
				value = 1
			}

			ch <- prometheus.MustNewConstMetric(c.resourceOwnerNode, prometheus.GaugeValue, value, resource.Type, resource.OwnerGroup, nodeName, resource.Name)
		}
	}

	return resultErr
}

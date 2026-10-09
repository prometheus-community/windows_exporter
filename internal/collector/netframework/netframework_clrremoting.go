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

package netframework

import (
	"fmt"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

func (c *Collector) describeClrRemoting() {
	c.channels = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrRemoting+"_channels_total"),
		"Displays the total number of remoting channels registered across all application domains since application started.",
		[]string{"process"},
		nil,
	)
	c.contextBoundClassesLoaded = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrRemoting+"_context_bound_classes_loaded"),
		"Displays the current number of context-bound classes that are loaded.",
		[]string{"process"},
		nil,
	)
	c.contextBoundObjects = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrRemoting+"_context_bound_objects_total"),
		"Displays the total number of context-bound objects allocated.",
		[]string{"process"},
		nil,
	)
	c.contextProxies = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrRemoting+"_context_proxies_total"),
		"Displays the total number of remoting proxy objects in this process since it started.",
		[]string{"process"},
		nil,
	)
	c.contexts = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrRemoting+"_contexts"),
		"Displays the current number of remoting contexts in the application.",
		[]string{"process"},
		nil,
	)
	c.totalRemoteCalls = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrRemoting+"_remote_calls_total"),
		"Displays the total number of remote procedure calls invoked since the application started.",
		[]string{"process"},
		nil,
	)
}

func (c *Collector) buildClrRemoting() error {
	c.describeClrRemoting()

	var err error

	c.perfClrRemoting, err = newPerfCollector[perfDataClrRemoting](c.logger, ".NET CLR Remoting")

	return err
}

type perfDataClrRemoting struct {
	Name                           string
	Channels                       float64 `perfdata:"Channels"`
	ContextBoundClassesLoaded      float64 `perfdata:"Context-Bound Classes Loaded"`
	ContextBoundObjectsAllocPersec float64 `perfdata:"Context-Bound Objects Alloc / sec"`
	ContextProxies                 float64 `perfdata:"Context Proxies"`
	Contexts                       float64 `perfdata:"Contexts"`
	TotalRemoteCalls               float64 `perfdata:"Total Remote Calls"`
}

func (c *Collector) collectClrRemoting(ch chan<- prometheus.Metric, _ time.Duration) error {
	var dst []perfDataClrRemoting
	if err := c.perfClrRemoting.Collect(&dst); err != nil {
		return fmt.Errorf("failed to collect .NET CLR Remoting: %w", err)
	}

	for _, process := range dst {
		if process.Name == "_Global_" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.channels,
			prometheus.CounterValue,
			float64(process.Channels),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.contextBoundClassesLoaded,
			prometheus.GaugeValue,
			float64(process.ContextBoundClassesLoaded),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.contextBoundObjects,
			prometheus.CounterValue,
			float64(process.ContextBoundObjectsAllocPersec),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.contextProxies,
			prometheus.CounterValue,
			float64(process.ContextProxies),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.contexts,
			prometheus.GaugeValue,
			float64(process.Contexts),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.totalRemoteCalls,
			prometheus.CounterValue,
			float64(process.TotalRemoteCalls),
			process.Name,
		)
	}

	return nil
}

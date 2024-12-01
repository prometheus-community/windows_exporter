// Copyright 2024 The Prometheus Authors
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

package mssql

import (
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorInstance struct {
	instances *prometheus.GaugeVec
}

func (c *Collector) buildInstance() error {
	c.instances = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: types.Namespace,
			Subsystem: Name,
			Name:      "instance_info",
			Help:      "A metric with a constant '1' value labeled with mssql instance information",
		},
		[]string{"edition", "mssql_instance", "patch", "version"},
	)

	for _, instance := range c.mssqlInstances {
		c.instances.WithLabelValues(instance.edition, instance.name, instance.patchVersion, instance.majorVersion.String()).Set(1)
	}

	return nil
}

func (c *Collector) collectInstance(ch chan<- prometheus.Metric) error {
	c.instances.Collect(ch)

	return nil
}

func (c *Collector) closeInstance() {
}

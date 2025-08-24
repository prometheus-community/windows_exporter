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

package mssql

import (
	"fmt"
	"log/slog"

	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows/registry"
)

type collectorInstance struct {
	instances *prometheus.Desc
}

func (c *Collector) buildInstance() error {
	c.instances = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "instance_info"),
		"A metric with a constant '1' value labeled with mssql instance information",
		[]string{"edition", "mssql_instance", "patch", "version"},
		nil,
	)

	return nil
}

func (c *Collector) collectInstance(ch chan<- prometheus.Metric) error {
	for _, instance := range c.mssqlInstances {
		regKeyName := fmt.Sprintf(`Software\Microsoft\Microsoft SQL Server\%s\Setup`, instance.name)

		regKey, err := registry.OpenKey(registry.LOCAL_MACHINE, regKeyName, registry.QUERY_VALUE)
		if err != nil {
			c.logger.Debug(fmt.Sprintf("couldn't open registry %s:", regKeyName),
				slog.Any("err", err),
			)

			continue
		}

		patchVersion, _, err := regKey.GetStringValue("PatchLevel")
		_ = regKey.Close()

		if err != nil {
			c.logger.Debug("couldn't get version from registry",
				slog.Any("err", err),
			)

			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.instances,
			prometheus.GaugeValue,
			1,
			instance.edition,
			instance.name,
			patchVersion,
			instance.majorVersion.String(),
		)
	}

	return nil
}

func (c *Collector) closeInstance() {
}

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
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const nameVirtualDisk = Name + "_virtualdisk"

type collectorVirtualDisk struct {
	virtualDiskMIQuery mi.Query

	virtualDiskInfo              *prometheus.Desc
	virtualDiskOperationalStatus *prometheus.Desc
	virtualDiskHealthStatus      *prometheus.Desc
	virtualDiskSize              *prometheus.Desc
	virtualDiskFootprintOnPool   *prometheus.Desc
	virtualDiskStorageEfficiency *prometheus.Desc
}

// msftVirtualDisk represents the MSFT_VirtualDisk WMI class
type msftVirtualDisk struct {
	FriendlyName      string `mi:"FriendlyName"`
	OperationalStatus string `mi:"OperationalStatus"`
	HealthStatus      string `mi:"HealthStatus"`
	Size              uint64 `mi:"Size"`
	FootprintOnPool   uint64 `mi:"FootprintOnPool"`
}

func (c *Collector) buildVirtualDisk() error {
	virtualDiskMIQuery, err := mi.NewQuery("SELECT FriendlyName, OperationalStatus, HealthStatus, Size, FootprintOnPool FROM MSFT_VirtualDisk")
	if err != nil {
		return fmt.Errorf("failed to create WMI query: %w", err)
	}

	c.virtualDiskMIQuery = virtualDiskMIQuery

	c.virtualDiskInfo = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameVirtualDisk, "info"),
		"Virtual Disk information (value is always 1)",
		[]string{"name"},
		nil,
	)

	c.virtualDiskOperationalStatus = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameVirtualDisk, "operational_status"),
		"Operational status of the Virtual Disk. 1: OK, 2: Detached",
		[]string{"name", "status"},
		nil,
	)

	c.virtualDiskHealthStatus = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameVirtualDisk, "health_status"),
		"Health status of the Virtual Disk. 1: Healthy, 2: Warning, 3: Unknown",
		[]string{"name", "status"},
		nil,
	)

	c.virtualDiskSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameVirtualDisk, "size_bytes"),
		"Total size of the Virtual Disk in bytes",
		[]string{"name"},
		nil,
	)

	c.virtualDiskFootprintOnPool = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameVirtualDisk, "footprint_on_pool_bytes"),
		"Physical space used by the Virtual Disk on the storage pool in bytes",
		[]string{"name"},
		nil,
	)

	c.virtualDiskStorageEfficiency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameVirtualDisk, "storage_efficiency_percent"),
		"Storage efficiency percentage (Size / FootprintOnPool * 100)",
		[]string{"name"},
		nil,
	)

	return nil
}

func (c *Collector) collectVirtualDisk(ch chan<- prometheus.Metric) error {
	var dst []msftVirtualDisk
	if err := c.miSession.Query(&dst, mi.NamespaceRootStorage, c.virtualDiskMIQuery); err != nil {
		return fmt.Errorf("WMI query failed: %w", err)
	}

	for _, vdisk := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.virtualDiskInfo,
			prometheus.GaugeValue,
			1.0,
			vdisk.FriendlyName,
		)

		// Map OperationalStatus to numeric value
		var opStatusValue float64
		switch vdisk.OperationalStatus {
		case "OK":
			opStatusValue = 1
		case "Detached":
			opStatusValue = 2
		default:
			opStatusValue = 0 // Unknown
		}

		ch <- prometheus.MustNewConstMetric(
			c.virtualDiskOperationalStatus,
			prometheus.GaugeValue,
			opStatusValue,
			vdisk.FriendlyName,
			vdisk.OperationalStatus,
		)

		// Map HealthStatus to numeric value
		var healthStatusValue float64
		switch vdisk.HealthStatus {
		case "Healthy":
			healthStatusValue = 1
		case "Warning":
			healthStatusValue = 2
		case "Unknown":
			healthStatusValue = 3
		default:
			healthStatusValue = 3 // Treat anything else as Unknown
		}

		ch <- prometheus.MustNewConstMetric(
			c.virtualDiskHealthStatus,
			prometheus.GaugeValue,
			healthStatusValue,
			vdisk.FriendlyName,
			vdisk.HealthStatus,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualDiskSize,
			prometheus.GaugeValue,
			float64(vdisk.Size),
			vdisk.FriendlyName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualDiskFootprintOnPool,
			prometheus.GaugeValue,
			float64(vdisk.FootprintOnPool),
			vdisk.FriendlyName,
		)

		// Calculate storage efficiency: Size / FootprintOnPool * 100
		var storageEfficiency float64
		if vdisk.FootprintOnPool > 0 {
			storageEfficiency = (float64(vdisk.Size) / float64(vdisk.FootprintOnPool)) * 100
		}

		ch <- prometheus.MustNewConstMetric(
			c.virtualDiskStorageEfficiency,
			prometheus.GaugeValue,
			storageEfficiency,
			vdisk.FriendlyName,
		)
	}

	return nil
}

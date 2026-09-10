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
	"time"

	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const nameStoragePool = Name + "_storagepool"

type collectorStoragePool struct {
	storagePoolMIQuery mi.Query

	storagePoolInfo          *prometheus.Desc
	storagePoolHealthStatus  *prometheus.Desc
	storagePoolSize          *prometheus.Desc
	storagePoolAllocatedSize *prometheus.Desc
}

// msftStoragePool represents the MSFT_StoragePool WMI class
type msftStoragePool struct {
	FriendlyName  string `mi:"FriendlyName"`
	UniqueId      string `mi:"UniqueId"`
	HealthStatus  uint16 `mi:"HealthStatus"`
	Size          uint64 `mi:"Size"`
	AllocatedSize uint64 `mi:"AllocatedSize"`
	// OperationalStatus []uint16 `mi:"OperationalStatus"`  Not supported by mi query: https://github.com/prometheus-community/windows_exporter/pull/2296#issuecomment-3736584632
	// ThinProvisioningAlertThresholds []uint16 `mi:"ThinProvisioningAlertThresholds"`  Not supported by mi query (array), see OperationalStatus above.
}

func (c *Collector) buildStoragePool() error {
	wmiSelect := "FriendlyName,UniqueId,HealthStatus,Size,AllocatedSize"

	storagePoolMIQuery, err := mi.NewQuery(fmt.Sprintf("SELECT %s FROM MSFT_StoragePool", wmiSelect))
	if err != nil {
		return fmt.Errorf("failed to create WMI query: %w", err)
	}

	c.storagePoolMIQuery = storagePoolMIQuery

	c.storagePoolInfo = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameStoragePool, "info"),
		"Storage pool information (value is always 1)",
		[]string{"name", "unique_id"},
		nil,
	)

	c.storagePoolHealthStatus = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameStoragePool, "health_status"),
		"Health status of the storage pool. 0: Healthy, 1: Warning, 2: Unhealthy, 5: Unknown",
		[]string{"name", "unique_id"},
		nil,
	)

	c.storagePoolSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameStoragePool, "size_bytes"),
		"Total size of the storage pool in bytes",
		[]string{"name", "unique_id"},
		nil,
	)

	c.storagePoolAllocatedSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameStoragePool, "allocated_size_bytes"),
		"Allocated size of the storage pool in bytes",
		[]string{"name", "unique_id"},
		nil,
	)

	var dst []msftStoragePool

	if err := c.miSession.Query(&dst, mi.NamespaceRootStorage, c.storagePoolMIQuery, 0); err != nil {
		return fmt.Errorf("WMI query failed: %w", err)
	}

	return nil
}

func (c *Collector) collectStoragePool(ch chan<- prometheus.Metric, maxScrapeDuration time.Duration) error {
	var dst []msftStoragePool

	if err := c.miSession.Query(&dst, mi.NamespaceRootStorage, c.storagePoolMIQuery, maxScrapeDuration); err != nil {
		return fmt.Errorf("WMI query failed: %w", err)
	}

	for _, pool := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.storagePoolInfo,
			prometheus.GaugeValue,
			1.0,
			pool.FriendlyName,
			pool.UniqueId,
		)

		ch <- prometheus.MustNewConstMetric(
			c.storagePoolHealthStatus,
			prometheus.GaugeValue,
			float64(pool.HealthStatus),
			pool.FriendlyName,
			pool.UniqueId,
		)

		ch <- prometheus.MustNewConstMetric(
			c.storagePoolSize,
			prometheus.GaugeValue,
			float64(pool.Size),
			pool.FriendlyName,
			pool.UniqueId,
		)

		ch <- prometheus.MustNewConstMetric(
			c.storagePoolAllocatedSize,
			prometheus.GaugeValue,
			float64(pool.AllocatedSize),
			pool.FriendlyName,
			pool.UniqueId,
		)
	}

	return nil
}

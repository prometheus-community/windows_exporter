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
	"strings"

	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const nameS2D = Name + "_csv"

type collectorS2D struct {
	s2dMIQuery mi.Query

	s2dInfo      *prometheus.Desc
	s2dTotalSize *prometheus.Desc
	s2dFreeSpace *prometheus.Desc
}

// msClusterDiskPartition represents the MSCluster_DiskPartition WMI class
type msClusterDiskPartition struct {
	Name      string `mi:"Name"`
	Path      string `mi:"Path"`
	TotalSize uint64 `mi:"TotalSize"`
	FreeSpace uint64 `mi:"FreeSpace"`
	Volume    string `mi:"VolumeLabel"`
}

func (c *Collector) buildS2D() error {
	s2dMIQuery, err := mi.NewQuery("SELECT Name, Path, TotalSize, FreeSpace, VolumeLabel FROM MSCluster_DiskPartition")
	if err != nil {
		return fmt.Errorf("failed to create WMI query: %w", err)
	}

	c.s2dMIQuery = s2dMIQuery

	c.s2dInfo = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameS2D, "info"),
		"Cluster Shared Volumes information",
		[]string{"name", "path", "volume"},
		nil,
	)

	c.s2dTotalSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameS2D, "total_bytes"),
		"Total size of the Cluster Shared Volume in bytes",
		[]string{"name", "path", "volume"},
		nil,
	)

	c.s2dFreeSpace = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameS2D, "free_bytes"),
		"Free space on the Cluster Shared Volume in bytes",
		[]string{"name", "path", "volume"},
		nil,
	)

	var dst []msClusterDiskPartition
	if err := c.miSession.Query(&dst, mi.NamespaceRootMSCluster, c.s2dMIQuery); err != nil {
		return fmt.Errorf("WMI query failed: %w", err)
	}

	return nil
}

func (c *Collector) collectS2D(ch chan<- prometheus.Metric) error {
	var dst []msClusterDiskPartition
	if err := c.miSession.Query(&dst, mi.NamespaceRootMSCluster, c.s2dMIQuery); err != nil {
		return fmt.Errorf("WMI query failed: %w", err)
	}

	for _, partition := range dst {
		name := strings.TrimRight(partition.Name, " ")
		path := strings.TrimRight(partition.Path, " ")
		volume := strings.TrimRight(partition.Volume, " ")

		ch <- prometheus.MustNewConstMetric(
			c.s2dInfo,
			prometheus.GaugeValue,
			1.0,
			name,
			path,
			volume,
		)

		ch <- prometheus.MustNewConstMetric(
			c.s2dTotalSize,
			prometheus.GaugeValue,
			float64(partition.TotalSize)*1024, // Convert from KB to bytes
			name,
			path,
			volume,
		)

		ch <- prometheus.MustNewConstMetric(
			c.s2dFreeSpace,
			prometheus.GaugeValue,
			float64(partition.FreeSpace)*1024, // Convert from KB to bytes
			name,
			path,
			volume,
		)
	}

	return nil
}

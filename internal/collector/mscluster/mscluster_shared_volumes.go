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

const nameSharedVolumes = Name + "_shared_volumes"

type collectorSharedVolumes struct {
	sharedVolumesMIQuery mi.Query

	sharedVolumesInfo      *prometheus.Desc
	sharedVolumesTotalSize *prometheus.Desc
	sharedVolumesFreeSpace *prometheus.Desc
}

// msClusterDiskPartition represents the MSCluster_DiskPartition WMI class
type msClusterDiskPartition struct {
	Name      string `mi:"Name"`
	Path      string `mi:"Path"`
	TotalSize uint64 `mi:"TotalSize"`
	FreeSpace uint64 `mi:"FreeSpace"`
	Volume    string `mi:"VolumeLabel"`
}

func (c *Collector) buildSharedVolumes() error {
	sharedVolumesMIQuery, err := mi.NewQuery("SELECT Name, Path, TotalSize, FreeSpace, VolumeLabel FROM MSCluster_DiskPartition")
	if err != nil {
		return fmt.Errorf("failed to create WMI query: %w", err)
	}

	c.sharedVolumesMIQuery = sharedVolumesMIQuery

	c.sharedVolumesInfo = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameSharedVolumes, "info"),
		"Cluster Shared Volumes information (value is always 1)",
		[]string{"name", "path"},
		nil,
	)

	c.sharedVolumesTotalSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameSharedVolumes, "total_bytes"),
		"Total size of the Cluster Shared Volume in bytes",
		[]string{"name"},
		nil,
	)

	c.sharedVolumesFreeSpace = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameSharedVolumes, "free_bytes"),
		"Free space on the Cluster Shared Volume in bytes",
		[]string{"name"},
		nil,
	)

	var dst []msClusterDiskPartition
	if err := c.miSession.Query(&dst, mi.NamespaceRootMSCluster, c.sharedVolumesMIQuery); err != nil {
		return fmt.Errorf("WMI query failed: %w", err)
	}

	return nil
}

func (c *Collector) collectSharedVolumes(ch chan<- prometheus.Metric) error {
	var dst []msClusterDiskPartition
	if err := c.miSession.Query(&dst, mi.NamespaceRootMSCluster, c.sharedVolumesMIQuery); err != nil {
		return fmt.Errorf("WMI query failed: %w", err)
	}

	for _, partition := range dst {
		volume := strings.TrimRight(partition.Volume, " ")

		ch <- prometheus.MustNewConstMetric(
			c.sharedVolumesInfo,
			prometheus.GaugeValue,
			1.0,
			volume,
			partition.Path,
		)

		ch <- prometheus.MustNewConstMetric(
			c.sharedVolumesTotalSize,
			prometheus.GaugeValue,
			float64(partition.TotalSize)*1024*1024, // Convert from KB to bytes
			volume,
		)

		ch <- prometheus.MustNewConstMetric(
			c.sharedVolumesFreeSpace,
			prometheus.GaugeValue,
			float64(partition.FreeSpace)*1024*1024, // Convert from KB to bytes
			volume,
		)
	}

	return nil
}

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
	"time"

	"github.com/prometheus-community/windows_exporter/internal/headers/clusapi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows"
)

const nameSharedVolumes = Name + "_shared_volumes"

type sharedVolumesSource interface {
	DiskPartitions(deadline time.Time) ([]clusapi.Partition, error)
	Close() error
}

type collectorSharedVolumes struct {
	sharedVolumesSource sharedVolumesSource

	sharedVolumesInfo      *prometheus.Desc
	sharedVolumesTotalSize *prometheus.Desc
	sharedVolumesFreeSpace *prometheus.Desc
}

// msClusterDiskPartition represents the MSCluster_DiskPartition WMI class. The
// collector reads ClusAPI; the parity test compares against this WMI model.
// TotalSize and FreeSpace are megabytes.
type msClusterDiskPartition struct {
	Name       string `mi:"Name"`
	Path       string `mi:"Path"`
	TotalSize  uint64 `mi:"TotalSize"`
	FreeSpace  uint64 `mi:"FreeSpace"`
	Volume     string `mi:"VolumeLabel"`
	VolumeGuid string `mi:"VolumeGuid"`
}

func (c *Collector) buildSharedVolumes() error {
	source, err := clusapi.Open()
	if err != nil {
		return err
	}

	c.sharedVolumesSource = source
	c.buildSharedVolumesDescriptors()

	return nil
}

func (c *Collector) buildSharedVolumesDescriptors() {
	c.sharedVolumesInfo = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameSharedVolumes, "info"),
		"Cluster Shared Volumes information (value is always 1)",
		[]string{"name", "path", "volume_guid"},
		nil,
	)

	c.sharedVolumesTotalSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameSharedVolumes, "total_bytes"),
		"Total size of the Cluster Shared Volume in bytes",
		[]string{"name", "volume_guid"},
		nil,
	)

	c.sharedVolumesFreeSpace = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameSharedVolumes, "free_bytes"),
		"Free space on the Cluster Shared Volume in bytes",
		[]string{"name", "volume_guid"},
		nil,
	)
}

func (c *Collector) collectSharedVolumes(ch chan<- prometheus.Metric, maxScrapeDuration time.Duration) error {
	if c.sharedVolumesSource == nil {
		return errNotBuilt(subCollectorSharedVolumes)
	}

	partitions, resultErr := callSource(maxScrapeDuration, c.sharedVolumesSource.DiskPartitions)

	return c.publishSharedVolumes(ch, partitions, resultErr)
}

func (c *Collector) publishSharedVolumes(ch chan<- prometheus.Metric, partitions []clusapi.Partition, resultErr error) error {
	for _, partition := range partitions {
		volume := strings.TrimRight(partition.VolumeLabel, " ")
		volumeGUID := formatVolumeGUID(partition.VolumeGUID)

		ch <- prometheus.MustNewConstMetric(
			c.sharedVolumesInfo,
			prometheus.GaugeValue,
			1.0,
			volume,
			partition.DeviceName,
			volumeGUID,
		)

		// MSCluster_DiskPartition reports whole megabytes; keep that resolution.
		ch <- prometheus.MustNewConstMetric(
			c.sharedVolumesTotalSize,
			prometheus.GaugeValue,
			float64(partition.TotalBytes>>20)*1024*1024,
			volume,
			volumeGUID,
		)

		ch <- prometheus.MustNewConstMetric(
			c.sharedVolumesFreeSpace,
			prometheus.GaugeValue,
			float64(partition.FreeBytes>>20)*1024*1024,
			volume,
			volumeGUID,
		)
	}

	return resultErr
}

// formatVolumeGUID matches MSCluster_DiskPartition.VolumeGuid: lower case,
// without braces.
func formatVolumeGUID(guid windows.GUID) string {
	return fmt.Sprintf("%08x-%04x-%04x-%x-%x",
		guid.Data1, guid.Data2, guid.Data3, guid.Data4[:2], guid.Data4[2:])
}

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

package csvv

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "csvv"

type Config struct{}

//nolint:gochecknoglobals
var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for CSV volume metrics from MSCluster_DiskPartition.
type Collector struct {
	config    Config
	miSession *mi.Session
	miQuery   mi.Query

	csvvTotalSize *prometheus.Desc
	csvvFreeSpace *prometheus.Desc
	csvvInfo      *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(_ *slog.Logger, miSession *mi.Session) error {
	c.csvvInfo = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "info"),
		"CSV volume information",
		[]string{
			"path",
			"volume",
		},
		nil,
	)

	c.csvvTotalSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "total_bytes"),
		"Total size of the CSV volume in bytes",
		[]string{
			"path",
			"volume",
		},
		nil,
	)

	c.csvvFreeSpace = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "free_bytes"),
		"Free space on the CSV volume in bytes",
		[]string{
			"path",
			"volume",
		},
		nil,
	)

	if miSession == nil {
		return errors.New("miSession is nil")
	}

	miQuery, err := mi.NewQuery("SELECT Path, TotalSize, FreeSpace, FileSystem, VolumeLabel FROM MSCluster_DiskPartition")
	if err != nil {
		return fmt.Errorf("failed to create WMI query: %w", err)
	}

	c.miQuery = miQuery
	c.miSession = miSession

	var dst []miDiskPartition
	if err := c.miSession.Query(&dst, mi.NamespaceRootMSCluster, c.miQuery); err != nil {
		return fmt.Errorf("WMI query failed: %w", err)
	}

	return nil
}

type miDiskPartition struct {
	Path      string `mi:"Path"`
	TotalSize uint64 `mi:"TotalSize"`
	FreeSpace uint64 `mi:"FreeSpace"`
	Volume    string `mi:"VolumeLabel"`
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	var dst []miDiskPartition
	if err := c.miSession.Query(&dst, mi.NamespaceRootMSCluster, c.miQuery); err != nil {
		return fmt.Errorf("WMI query failed: %w", err)
	}

	for _, partition := range dst {
		path := strings.TrimRight(partition.Path, " ")
		volume := strings.TrimRight(partition.volume, " ")

		ch <- prometheus.MustNewConstMetric(
			c.csvvInfo,
			prometheus.GaugeValue,
			1.0,
			path,
			volume,
		)

		ch <- prometheus.MustNewConstMetric(
			c.csvvTotalSize,
			prometheus.GaugeValue,
			float64(partition.TotalSize)*1024, // Convert from KB to bytes
			path,
			volume,
		)

		ch <- prometheus.MustNewConstMetric(
			c.csvvFreeSpace,
			prometheus.GaugeValue,
			float64(partition.FreeSpace)*1024, // Convert from KB to bytes
			path,
			volume,
		)
	}

	return nil
}

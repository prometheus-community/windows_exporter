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

package csv

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "csv"

type Config struct{}

//nolint:gochecknoglobals
var ConfigDefaults = Config{}

type Collector struct {
	config Config
	logger *slog.Logger

	capacityBytes  *prometheus.Desc
	freeSpaceBytes *prometheus.Desc
	usedSpaceBytes *prometheus.Desc
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

func NewWithFlags(app *kingpin.Application) *Collector {
	return &Collector{
		config: ConfigDefaults,
	}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	c.logger = logger.With(slog.String("collector", Name))

	c.capacityBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "capacity_bytes"),
		"Total capacity of the Cluster Shared Volume in bytes",
		[]string{"volume"},
		nil,
	)
	c.freeSpaceBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "free_space_bytes"),
		"Free space of the Cluster Shared Volume in bytes",
		[]string{"volume"},
		nil,
	)
	c.usedSpaceBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "used_space_bytes"),
		"Used space of the Cluster Shared Volume in bytes",
		[]string{"volume"},
		nil,
	)

	return nil
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	csvInfos, err := c.getClusterSharedVolumes()
	if err != nil {
		return fmt.Errorf("failed to get cluster shared volumes: %w", err)
	}

	for _, info := range csvInfos {
		ch <- prometheus.MustNewConstMetric(
			c.capacityBytes,
			prometheus.GaugeValue,
			info.CapacityBytes,
			info.FriendlyVolumeName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.freeSpaceBytes,
			prometheus.GaugeValue,
			info.FreeSpaceBytes,
			info.FriendlyVolumeName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.usedSpaceBytes,
			prometheus.GaugeValue,
			info.UsedSpaceBytes,
			info.FriendlyVolumeName,
		)
	}

	return nil
}

func (c *Collector) getClusterSharedVolumes() ([]ClusterSharedVolumeInfo, error) {
	psCommand := `
		Get-ClusterSharedVolume |
		ForEach-Object {
			$partition = $_.SharedVolumeInfo.Partition
			$friendlyName = $_.SharedVolumeInfo.FriendlyVolumeName
			$capacity = $partition.Size
			$usedSpace = $partition.UsedSpace
			$freeSpace = $partition.FreeSpace
			"$friendlyName|$capacity|$freeSpace|$usedSpace"
		}
	`

	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", psCommand)
	output, err := cmd.Output()
	if err != nil {
		c.logger.Error("failed to execute PowerShell command",
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("failed to execute PowerShell command: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	csvInfos := make([]ClusterSharedVolumeInfo, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) != 4 {
			c.logger.Warn("invalid CSV output line",
				slog.String("line", line),
			)
			continue
		}

		capacity, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			c.logger.Warn("failed to parse capacity",
				slog.String("value", parts[1]),
				slog.Any("error", err),
			)
			continue
		}

		freeSpace, err := strconv.ParseFloat(parts[2], 64)
		if err != nil {
			c.logger.Warn("failed to parse free space",
				slog.String("value", parts[2]),
				slog.Any("error", err),
			)
			continue
		}

		usedSpace, err := strconv.ParseFloat(parts[3], 64)
		if err != nil {
			c.logger.Warn("failed to parse used space",
				slog.String("value", parts[3]),
				slog.Any("error", err),
			)
			continue
		}

		csvInfos = append(csvInfos, ClusterSharedVolumeInfo{
			FriendlyVolumeName: parts[0],
			CapacityBytes:      capacity,
			FreeSpaceBytes:     freeSpace,
			UsedSpaceBytes:     usedSpace,
		})
	}

	return csvInfos, nil
}

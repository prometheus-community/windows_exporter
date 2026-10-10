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

package diskdrive

import (
	"cmp"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "diskdrive"

type Config struct{}

//nolint:gochecknoglobals
var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for the Win32_DiskDrive properties
// DeviceID, Model, Caption, Name, Partitions, Size, Status and Availability.
// The values are read natively when they match WMI at Build time.
type Collector struct {
	config Config
	logger *slog.Logger

	miSession *mi.Session
	miQuery   mi.Query

	// useNative is set by Build when the native reader reproduced the WMI
	// result exactly.
	useNative  bool
	nativeRead func() ([]diskDrive, error)

	availability *prometheus.Desc
	diskInfo     *prometheus.Desc
	partitions   *prometheus.Desc
	size         *prometheus.Desc
	status       *prometheus.Desc
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

func (c *Collector) Build(logger *slog.Logger, miSession *mi.Session) error {
	c.logger = logger.With(slog.String("collector", Name))
	c.useNative = false

	c.diskInfo = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "info"),
		"General drive information",
		[]string{
			"device_id",
			"model",
			"caption",
			"name",
		},
		nil,
	)
	c.status = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "status"),
		"Status of the drive",
		[]string{"name", "status"},
		nil,
	)
	c.size = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "size"),
		"Size of the disk drive. It is calculated by multiplying the total number of cylinders, tracks in each cylinder, sectors in each track, and bytes in each sector.",
		[]string{"name"},
		nil,
	)
	c.partitions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "partitions"),
		"Number of partitions",
		[]string{"name"},
		nil,
	)
	c.availability = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "availability"),
		"Availability Status",
		[]string{"name", "availability"},
		nil,
	)

	if miSession == nil {
		return errors.New("miSession is nil")
	}

	miQuery, err := mi.NewQuery("SELECT DeviceID, Model, Caption, Name, Partitions, Size, Status, Availability FROM WIN32_DiskDrive")
	if err != nil {
		return fmt.Errorf("failed to create WMI query: %w", err)
	}

	c.miQuery = miQuery
	c.miSession = miSession

	wmiDrives, err := c.readWMI(0)
	if err != nil {
		return err
	}

	if c.nativeRead == nil {
		c.nativeRead = readNativeDiskDrives
	}

	// The native reader replaces WMI only after it reproduced the complete WMI
	// result on this host. The properties are rebuilt from the CIMWin32
	// implementation of a recent Windows build; older provider builds keep WMI
	// if they behave differently.
	nativeDrives, err := c.nativeRead()

	switch {
	case err != nil:
		c.logger.Info("native disk drive enumeration failed, using WMI",
			slog.Any("err", err),
		)
	case !equalDiskDrives(nativeDrives, wmiDrives):
		c.logger.Info("native disk drive properties differ from Win32_DiskDrive, using WMI",
			slog.Any("wmi", wmiDrives),
			slog.Any("native", nativeDrives),
		)
	default:
		c.logger.Debug("native disk drive properties match Win32_DiskDrive, using native APIs")

		c.useNative = true
	}

	return nil
}

func (c *Collector) readWMI(maxScrapeDuration time.Duration) ([]diskDrive, error) {
	var dst []diskDrive
	if err := c.miSession.Query(&dst, mi.NamespaceRootCIMv2, c.miQuery, maxScrapeDuration); err != nil {
		return nil, fmt.Errorf("WMI query failed: %w", err)
	}

	return dst, nil
}

func (c *Collector) readDiskDrives(maxScrapeDuration time.Duration) ([]diskDrive, error) {
	if !c.useNative {
		return c.readWMI(maxScrapeDuration)
	}

	drives, err := c.nativeRead()
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate disk drives: %w", err)
	}

	return drives, nil
}

// equalDiskDrives reports whether both results contain the same drives with
// identical raw property values, independent of their order.
func equalDiskDrives(a, b []diskDrive) bool {
	if len(a) != len(b) {
		return false
	}

	a = slices.Clone(a)
	b = slices.Clone(b)

	slices.SortFunc(a, compareDiskDrives)
	slices.SortFunc(b, compareDiskDrives)

	return slices.Equal(a, b)
}

func compareDiskDrives(a, b diskDrive) int {
	return cmp.Or(
		cmp.Compare(a.DeviceID, b.DeviceID),
		cmp.Compare(a.Name, b.Name),
		cmp.Compare(a.Model, b.Model),
		cmp.Compare(a.Caption, b.Caption),
		cmp.Compare(a.Size, b.Size),
		cmp.Compare(a.Partitions, b.Partitions),
		cmp.Compare(a.Status, b.Status),
		cmp.Compare(a.Availability, b.Availability),
	)
}

type diskDrive struct {
	DeviceID     string `mi:"DeviceID"`
	Model        string `mi:"Model"`
	Size         uint64 `mi:"Size"`
	Name         string `mi:"Name"`
	Caption      string `mi:"Caption"`
	Partitions   uint32 `mi:"Partitions"`
	Status       string `mi:"Status"`
	Availability uint16 `mi:"Availability"`
}

//nolint:gochecknoglobals
var (
	allDiskStatus = []string{
		"OK",
		"Error",
		"Degraded",
		"Unknown",
		"Pred Fail",
		"Starting",
		"Stopping",
		"Service",
		"Stressed",
		"Nonrecover",
		"No Contact",
		"Lost Comm",
	}

	availMap = map[int]string{
		1:  "Other",
		2:  "Unknown",
		3:  "Running / Full Power",
		4:  "Warning",
		5:  "In Test",
		6:  "Not Applicable",
		7:  "Power Off",
		8:  "Off line",
		9:  "Off Duty",
		10: "Degraded",
		11: "Not Installed",
		12: "Install Error",
		13: "Power Save - Unknown",
		14: "Power Save - Low Power Mode",
		15: "Power Save - Standby",
		16: "Power Cycle",
		17: "Power Save - Warning",
		18: "Paused",
		19: "Not Ready",
		20: "Not Configured",
		21: "Quiesced",
	}
)

// Collect sends the metric values for each metric to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric, maxScrapeDuration time.Duration) error {
	dst, err := c.readDiskDrives(maxScrapeDuration)
	if err != nil {
		return err
	}

	if len(dst) == 0 {
		return errors.New("no disk drives found")
	}

	for _, disk := range dst {
		distName := strings.Trim(disk.Name, `\.`)

		ch <- prometheus.MustNewConstMetric(
			c.diskInfo,
			prometheus.GaugeValue,
			1.0,
			strings.Trim(disk.DeviceID, `\.`),
			strings.TrimRight(disk.Model, " "),
			strings.TrimRight(disk.Caption, " "),
			distName,
		)

		for _, status := range allDiskStatus {
			isCurrentState := 0.0
			if status == disk.Status {
				isCurrentState = 1.0
			}

			ch <- prometheus.MustNewConstMetric(
				c.status,
				prometheus.GaugeValue,
				isCurrentState,
				distName,
				status,
			)
		}

		ch <- prometheus.MustNewConstMetric(
			c.size,
			prometheus.GaugeValue,
			float64(disk.Size),
			distName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.partitions,
			prometheus.GaugeValue,
			float64(disk.Partitions),
			distName,
		)

		for availNum, val := range availMap {
			isCurrentState := 0.0
			if availNum == int(disk.Availability) {
				isCurrentState = 1.0
			}

			ch <- prometheus.MustNewConstMetric(
				c.availability,
				prometheus.GaugeValue,
				isCurrentState,
				distName,
				val,
			)
		}
	}

	return nil
}

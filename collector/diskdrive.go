//go:build windows
// +build windows

package collector

import (
	"errors"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
)

const (
	win32DiskQuery = "SELECT DeviceID, Model, Caption, Name, Partitions, Size, Status, Availability FROM WIN32_DiskDrive"
)

// A DiskDriveInfoCollector is a Prometheus collector for a few WMI metrics in Win32_DiskDrive
type DiskDriveInfoCollector struct {
	logger log.Logger

	DiskInfo     *prometheus.Desc
	Status       *prometheus.Desc
	Size         *prometheus.Desc
	Partitions   *prometheus.Desc
	Availability *prometheus.Desc
}

func newDiskDriveInfoCollector(logger log.Logger) (Collector, error) {
	const subsystem = "diskdrive"

	return &DiskDriveInfoCollector{
		logger: log.With(logger, "collector", subsystem),

		DiskInfo: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "info"),
			"General drive information",
			[]string{
				"device_id",
				"model",
				"caption",
				"name",
			},
			nil,
		),

		Status: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "status"),
			"Status of the drive",
			[]string{
				"name", "status"},
			nil,
		),

		Size: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "size"),
			"Size of the disk drive. It is calculated by multiplying the total number of cylinders, tracks in each cylinder, sectors in each track, and bytes in each sector.",
			[]string{"name"},
			nil,
		),

		Partitions: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "partitions"),
			"Number of partitions",
			[]string{"name"},
			nil,
		),

		Availability: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "availability"),
			"Availability Status",
			[]string{
				"name", "availability"},
			nil,
		),
	}, nil
}

type Win32_DiskDrive struct {
	DeviceID     string
	Model        string
	Size         uint64
	Name         string
	Caption      string
	Partitions   uint32
	Status       string
	Availability uint16
}

var (
	allDiskStatus = []string{
		"OK",
		"Error",
		"Degraded",
		"Unknown",
		"Pred fail",
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
func (c *DiskDriveInfoCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting disk_drive_info metrics", "desc", desc, "err", err)
		return err
	}
	return nil
}

func (c *DiskDriveInfoCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_DiskDrive

	if err := wmi.Query(win32DiskQuery, &dst); err != nil {
		return nil, err
	}
	if len(dst) == 0 {
		return nil, errors.New("WMI query returned empty result set")
	}

	for _, disk := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.DiskInfo,
			prometheus.GaugeValue,
			1.0,
			strings.Trim(disk.DeviceID, "\\.\\"),
			strings.TrimRight(disk.Model, " "),
			strings.TrimRight(disk.Caption, " "),
			strings.TrimRight(disk.Name, "\\.\\"),
		)

		for _, status := range allDiskStatus {
			isCurrentState := 0.0
			if status == disk.Status {
				isCurrentState = 1.0
			}

			ch <- prometheus.MustNewConstMetric(
				c.Status,
				prometheus.GaugeValue,
				isCurrentState,
				strings.Trim(disk.Name, "\\.\\"),
				status,
			)
		}

		ch <- prometheus.MustNewConstMetric(
			c.Size,
			prometheus.GaugeValue,
			float64(disk.Size),
			strings.Trim(disk.Name, "\\.\\"),
		)

		ch <- prometheus.MustNewConstMetric(
			c.Partitions,
			prometheus.GaugeValue,
			float64(disk.Partitions),
			strings.Trim(disk.Name, "\\.\\"),
		)

		for availNum, val := range availMap {
			isCurrentState := 0.0
			if availNum == int(disk.Availability) {
				isCurrentState = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				c.Availability,
				prometheus.GaugeValue,
				isCurrentState,
				strings.Trim(disk.Name, "\\.\\"),
				val,
			)
		}
	}

	return nil, nil
}

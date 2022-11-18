//go:build windows
// +build windows

package collector

import (
	"errors"
	"strings"

	"github.com/StackExchange/wmi"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("disk_drive", newDiskDriveInfoCollector)
}

const (
	win32DiskQuery = "SELECT DeviceID, Model, Caption, Name, Partitions, Size, Status, Availability FROM WIN32_DiskDrive"
)

// A DiskDriveInfoCollector is a Prometheus collector for a few WMI metrics in Win32_DiskDrive
type DiskDriveInfoCollector struct {
	DiskInfo     *prometheus.Desc
	Status       *prometheus.Desc
	Size         *prometheus.Desc
	Partitions   *prometheus.Desc
	Availability *prometheus.Desc
}

func newDiskDriveInfoCollector() (Collector, error) {
	const subsystem = "disk_drive"

	return &DiskDriveInfoCollector{
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
		log.Error("failed collecting disk_drive_info metrics:", desc, err)
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

	for _, processor := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.DiskInfo,
			prometheus.GaugeValue,
			1.0,
			strings.Trim(processor.DeviceID, "\\.\\"),
			strings.TrimRight(processor.Model, " "),
			strings.TrimRight(processor.Caption, " "),
			strings.TrimRight(processor.Name, "\\.\\"),
		)

		for _, status := range allDiskStatus {
			isCurrentState := 0.0
			if status == processor.Status {
				isCurrentState = 1.0
			}

			ch <- prometheus.MustNewConstMetric(
				c.Status,
				prometheus.GaugeValue,
				isCurrentState,
				strings.Trim(processor.Name, "\\.\\"),
				status,
			)
		}

		ch <- prometheus.MustNewConstMetric(
			c.Size,
			prometheus.CounterValue,
			float64(processor.Size),
			strings.Trim(processor.Name, "\\.\\"),
		)

		ch <- prometheus.MustNewConstMetric(
			c.Partitions,
			prometheus.CounterValue,
			float64(processor.Partitions),
			strings.Trim(processor.Name, "\\.\\"),
		)

		for availNum, val := range availMap {
			isCurrentState := 0.0
			if availNum == int(processor.Availability) {
				isCurrentState = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				c.Availability,
				prometheus.GaugeValue,
				isCurrentState,
				strings.Trim(processor.Name, "\\.\\"),
				val,
			)
		}
	}

	return nil, nil
}

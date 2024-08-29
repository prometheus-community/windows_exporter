//go:build windows

package diskdrive

import (
	"errors"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
)

const (
	Name           = "diskdrive"
	win32DiskQuery = "SELECT DeviceID, Model, Caption, Name, Partitions, Size, Status, Availability FROM WIN32_DiskDrive"
)

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for a few WMI metrics in Win32_DiskDrive.
type Collector struct {
	config    Config
	wmiClient *wmi.Client

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

func (c *Collector) GetPerfCounter(_ log.Logger) ([]string, error) {
	return []string{}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(_ log.Logger, wmiClient *wmi.Client) error {
	if wmiClient == nil || wmiClient.SWbemServicesClient == nil {
		return errors.New("wmiClient or SWbemServicesClient is nil")
	}

	c.wmiClient = wmiClient
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

	return nil
}

type win32_DiskDrive struct {
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
func (c *Collector) Collect(_ *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	if err := c.collect(ch); err != nil {
		_ = level.Error(logger).Log("msg", "failed collecting disk_drive_info metrics", "err", err)
		return err
	}
	return nil
}

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	var dst []win32_DiskDrive

	if err := c.wmiClient.Query(win32DiskQuery, &dst); err != nil {
		return err
	}
	if len(dst) == 0 {
		return errors.New("WMI query returned empty result set")
	}

	for _, disk := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.diskInfo,
			prometheus.GaugeValue,
			1.0,
			strings.Trim(disk.DeviceID, "\\.\\"), //nolint:staticcheck
			strings.TrimRight(disk.Model, " "),
			strings.TrimRight(disk.Caption, " "),
			strings.TrimRight(disk.Name, "\\.\\"), //nolint:staticcheck
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
				strings.Trim(disk.Name, "\\.\\"), //nolint:staticcheck
				status,
			)
		}

		ch <- prometheus.MustNewConstMetric(
			c.size,
			prometheus.GaugeValue,
			float64(disk.Size),
			strings.Trim(disk.Name, "\\.\\"), //nolint:staticcheck
		)

		ch <- prometheus.MustNewConstMetric(
			c.partitions,
			prometheus.GaugeValue,
			float64(disk.Partitions),
			strings.Trim(disk.Name, "\\.\\"), //nolint:staticcheck
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
				strings.Trim(disk.Name, "\\.\\"), //nolint:staticcheck
				val,
			)
		}
	}

	return nil
}

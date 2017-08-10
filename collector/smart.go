// returns data points from MSStorageDriver_ATAPISmartData class
// in order to function, vendor-specific drivers for your HDD will need to be installed
// vendor-specific code based on https://exchange.nagios.org/directory/Plugins/Operating-Systems/Windows/NRPE/check_smartwmi-SMART-Monitoring-for-Windows-by-using-builtin-WMI/details

package collector

import (
	"log"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	Factories["smart"] = NewSMARTCollector
}

// A SMARTCollector is a Prometheus collector for WMI metrics
type SMARTCollector struct {
	SelfTestStatus *prometheus.Desc
	TotalTime      *prometheus.Desc
	Capability     *prometheus.Desc
}

// NewSMARTCollector ...
func NewSMARTCollector() (Collector, error) {
	const subsystem = "smart"

	return &SMARTCollector{
		SelfTestStatus: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "selftest_status"),
			"The self test status code (SMART.SelfTestStatus)",
			[]string{"volume"},
			nil,
		),
		TotalTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "total_time"),
			"Total time used (SMART.TotalTime)",
			[]string{"volume"},
			nil,
		),
		Capability: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "capability"),
			"Smart capability (SMART.SmartCapability)",
			[]string{"volume"},
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *SMARTCollector) Collect(ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Println("[ERROR] failed collecting smart metrics:", desc, err)
		return err
	}
	return nil
}

type MSStorageDriver_ATAPISmartData struct {
	InstanceName    string
	Active          bool
	SelfTestStatus  uint64
	TotalTime       uint64
	SmartCapability uint64
	// VendorSpecific  []uint8 // TODO parse this when https://github.com/StackExchange/wmi/pull/30 is merged
}

func (c *SMARTCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []MSStorageDriver_ATAPISmartData
	if err := wmi.QueryNamespace(wmi.CreateQuery(&dst, ""), &dst, `root\wmi`); err != nil {
		return nil, err
	}

	for _, disk := range dst {
		if !disk.Active {
			// exclude non-active disks
			continue
		}
		volume := disk.InstanceName
		ch <- prometheus.MustNewConstMetric(
			c.SelfTestStatus,
			prometheus.GaugeValue,
			float64(disk.SelfTestStatus),
			volume,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalTime,
			prometheus.GaugeValue,
			float64(disk.TotalTime),
			volume,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Capability,
			prometheus.GaugeValue,
			float64(disk.SmartCapability),
			volume,
		)
	}

	return nil, nil
}

package mscluster_resourcegroup

import (
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/wmi"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "mscluster_resourcegroup"

type Config struct{}

var ConfigDefaults = Config{}

// A collector is a Prometheus collector for WMI MSCluster_ResourceGroup metrics
type collector struct {
	logger log.Logger

	AutoFailbackType    *prometheus.Desc
	Characteristics     *prometheus.Desc
	ColdStartSetting    *prometheus.Desc
	DefaultOwner        *prometheus.Desc
	FailbackWindowEnd   *prometheus.Desc
	FailbackWindowStart *prometheus.Desc
	FailoverPeriod      *prometheus.Desc
	FailoverThreshold   *prometheus.Desc
	FaultDomain         *prometheus.Desc
	Flags               *prometheus.Desc
	GroupType           *prometheus.Desc
	PlacementOptions    *prometheus.Desc
	Priority            *prometheus.Desc
	ResiliencyPeriod    *prometheus.Desc
	State               *prometheus.Desc
}

func New(logger log.Logger, _ *Config) types.Collector {
	c := &collector{}
	c.SetLogger(logger)
	return c
}

func NewWithFlags(_ *kingpin.Application) types.Collector {
	return &collector{}
}

func (c *collector) GetName() string {
	return Name
}

func (c *collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *collector) GetPerfCounter() ([]string, error) {
	return []string{"Memory"}, nil
}

func (c *collector) Build() error {
	c.AutoFailbackType = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "auto_failback_type"),
		"Provides access to the group's AutoFailbackType property.",
		[]string{"name"},
		nil,
	)
	c.Characteristics = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "characteristics"),
		"Provides the characteristics of the group.",
		[]string{"name"},
		nil,
	)
	c.ColdStartSetting = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cold_start_setting"),
		"Indicates whether a group can start after a cluster cold start.",
		[]string{"name"},
		nil,
	)
	c.DefaultOwner = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "default_owner"),
		"Number of the last node the resource group was activated on or explicitly moved to.",
		[]string{"name"},
		nil,
	)
	c.FailbackWindowEnd = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "failback_window_end"),
		"The FailbackWindowEnd property provides the latest time that the group can be moved back to the node identified as its preferred node.",
		[]string{"name"},
		nil,
	)
	c.FailbackWindowStart = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "failback_window_start"),
		"The FailbackWindowStart property provides the earliest time (that is, local time as kept by the cluster) that the group can be moved back to the node identified as its preferred node.",
		[]string{"name"},
		nil,
	)
	c.FailoverPeriod = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "failover_period"),
		"The FailoverPeriod property specifies a number of hours during which a maximum number of failover attempts, specified by the FailoverThreshold property, can occur.",
		[]string{"name"},
		nil,
	)
	c.FailoverThreshold = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "failover_threshold"),
		"The FailoverThreshold property specifies the maximum number of failover attempts.",
		[]string{"name"},
		nil,
	)
	c.Flags = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "flags"),
		"Provides access to the flags set for the group. ",
		[]string{"name"},
		nil,
	)
	c.GroupType = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "group_type"),
		"The Type of the resource group.",
		[]string{"name"},
		nil,
	)
	c.Priority = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "priority"),
		"Priority value of the resource group",
		[]string{"name"},
		nil,
	)
	c.ResiliencyPeriod = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "resiliency_period"),
		"The resiliency period for this group, in seconds.",
		[]string{"name"},
		nil,
	)
	c.State = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "state"),
		"The current state of the resource group. -1: Unknown; 0: Online; 1: Offline; 2: Failed; 3: Partial Online; 4: Pending",
		[]string{"name"},
		nil,
	)
	return nil
}

// MSCluster_ResourceGroup docs:
// - https://docs.microsoft.com/en-us/previous-versions/windows/desktop/cluswmi/mscluster-resourcegroup
type MSCluster_ResourceGroup struct {
	Name string

	AutoFailbackType    uint
	Characteristics     uint
	ColdStartSetting    uint
	DefaultOwner        uint
	FailbackWindowEnd   int
	FailbackWindowStart int
	FailoverPeriod      uint
	FailoverThreshold   uint
	Flags               uint
	GroupType           uint
	Priority            uint
	ResiliencyPeriod    uint
	State               uint
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var dst []MSCluster_ResourceGroup
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.QueryNamespace(q, &dst, "root/MSCluster"); err != nil {
		return err
	}

	for _, v := range dst {

		ch <- prometheus.MustNewConstMetric(
			c.AutoFailbackType,
			prometheus.GaugeValue,
			float64(v.AutoFailbackType),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Characteristics,
			prometheus.GaugeValue,
			float64(v.Characteristics),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ColdStartSetting,
			prometheus.GaugeValue,
			float64(v.ColdStartSetting),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DefaultOwner,
			prometheus.GaugeValue,
			float64(v.DefaultOwner),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FailbackWindowEnd,
			prometheus.GaugeValue,
			float64(v.FailbackWindowEnd),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FailbackWindowStart,
			prometheus.GaugeValue,
			float64(v.FailbackWindowStart),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FailoverPeriod,
			prometheus.GaugeValue,
			float64(v.FailoverPeriod),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FailoverThreshold,
			prometheus.GaugeValue,
			float64(v.FailoverThreshold),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Flags,
			prometheus.GaugeValue,
			float64(v.Flags),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GroupType,
			prometheus.GaugeValue,
			float64(v.GroupType),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Priority,
			prometheus.GaugeValue,
			float64(v.Priority),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ResiliencyPeriod,
			prometheus.GaugeValue,
			float64(v.ResiliencyPeriod),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.State,
			prometheus.GaugeValue,
			float64(v.State),
			v.Name,
		)

	}

	return nil
}

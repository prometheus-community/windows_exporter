package mscluster_resourcegroup

import (
	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/prometheus-community/windows_exporter/pkg/collector/mscluster_node"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "mscluster_resourcegroup"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI MSCluster_ResourceGroup metrics
type Collector struct {
	logger log.Logger

	autoFailbackType    *prometheus.Desc
	characteristics     *prometheus.Desc
	coldStartSetting    *prometheus.Desc
	defaultOwner        *prometheus.Desc
	failbackWindowEnd   *prometheus.Desc
	failbackWindowStart *prometheus.Desc
	failOverPeriod      *prometheus.Desc
	failOverThreshold   *prometheus.Desc
	flags               *prometheus.Desc
	groupType           *prometheus.Desc
	ownerNode           *prometheus.Desc
	priority            *prometheus.Desc
	resiliencyPeriod    *prometheus.Desc
	state               *prometheus.Desc
}

func New(logger log.Logger, _ *Config) *Collector {
	c := &Collector{}
	c.SetLogger(logger)

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *Collector) GetPerfCounter() ([]string, error) {
	return []string{"Memory"}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build() error {
	c.autoFailbackType = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "auto_failback_type"),
		"Provides access to the group's AutoFailbackType property.",
		[]string{"name"},
		nil,
	)
	c.characteristics = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "characteristics"),
		"Provides the characteristics of the group.",
		[]string{"name"},
		nil,
	)
	c.coldStartSetting = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cold_start_setting"),
		"Indicates whether a group can start after a cluster cold start.",
		[]string{"name"},
		nil,
	)
	c.defaultOwner = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "default_owner"),
		"Number of the last node the resource group was activated on or explicitly moved to.",
		[]string{"name"},
		nil,
	)
	c.failbackWindowEnd = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "failback_window_end"),
		"The FailbackWindowEnd property provides the latest time that the group can be moved back to the node identified as its preferred node.",
		[]string{"name"},
		nil,
	)
	c.failbackWindowStart = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "failback_window_start"),
		"The FailbackWindowStart property provides the earliest time (that is, local time as kept by the cluster) that the group can be moved back to the node identified as its preferred node.",
		[]string{"name"},
		nil,
	)
	c.failOverPeriod = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "failover_period"),
		"The FailoverPeriod property specifies a number of hours during which a maximum number of failover attempts, specified by the FailoverThreshold property, can occur.",
		[]string{"name"},
		nil,
	)
	c.failOverThreshold = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "failover_threshold"),
		"The FailoverThreshold property specifies the maximum number of failover attempts.",
		[]string{"name"},
		nil,
	)
	c.flags = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "flags"),
		"Provides access to the flags set for the group. ",
		[]string{"name"},
		nil,
	)
	c.groupType = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "group_type"),
		"The Type of the resource group.",
		[]string{"name"},
		nil,
	)
	c.ownerNode = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "owner_node"),
		"The node hosting the resource group. 0: Not hosted; 1: Hosted",
		[]string{"node_name", "name"},
		nil,
	)
	c.ownerNode = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "owner_node"),
		"The node hosting the resource group. 0: Not hosted; 1: Hosted",
		[]string{"node_name", "name"},
		nil,
	)
	c.priority = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "priority"),
		"Priority value of the resource group",
		[]string{"name"},
		nil,
	)
	c.resiliencyPeriod = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "resiliency_period"),
		"The resiliency period for this group, in seconds.",
		[]string{"name"},
		nil,
	)
	c.state = prometheus.NewDesc(
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
	OwnerNode           string
	Priority            uint
	ResiliencyPeriod    uint
	State               uint
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var dst []MSCluster_ResourceGroup
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.QueryNamespace(q, &dst, "root/MSCluster"); err != nil {
		return err
	}

	for _, v := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.autoFailbackType,
			prometheus.GaugeValue,
			float64(v.AutoFailbackType),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.characteristics,
			prometheus.GaugeValue,
			float64(v.Characteristics),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.coldStartSetting,
			prometheus.GaugeValue,
			float64(v.ColdStartSetting),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.defaultOwner,
			prometheus.GaugeValue,
			float64(v.DefaultOwner),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.failbackWindowEnd,
			prometheus.GaugeValue,
			float64(v.FailbackWindowEnd),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.failbackWindowStart,
			prometheus.GaugeValue,
			float64(v.FailbackWindowStart),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.failOverPeriod,
			prometheus.GaugeValue,
			float64(v.FailoverPeriod),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.failOverThreshold,
			prometheus.GaugeValue,
			float64(v.FailoverThreshold),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.flags,
			prometheus.GaugeValue,
			float64(v.Flags),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.groupType,
			prometheus.GaugeValue,
			float64(v.GroupType),
			v.Name,
		)

		if mscluster_node.NodeName != nil {
			for _, node_name := range mscluster_node.NodeName {
				isCurrentState := 0.0
				if v.OwnerNode == node_name {
					isCurrentState = 1.0
				}
				ch <- prometheus.MustNewConstMetric(
					c.ownerNode,
					prometheus.GaugeValue,
					isCurrentState,
					node_name, v.Name,
				)
			}
		}

		ch <- prometheus.MustNewConstMetric(
			c.priority,
			prometheus.GaugeValue,
			float64(v.Priority),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.resiliencyPeriod,
			prometheus.GaugeValue,
			float64(v.ResiliencyPeriod),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.state,
			prometheus.GaugeValue,
			float64(v.State),
			v.Name,
		)
	}

	return nil
}

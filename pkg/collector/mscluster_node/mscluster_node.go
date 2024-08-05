package mscluster_node

import (
	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "mscluster_node"

type Config struct{}

var ConfigDefaults = Config{}

// Variable used by mscluster_resource and mscluster_resourcegroup
var NodeName []string

// A Collector is a Prometheus Collector for WMI MSCluster_Node metrics
type Collector struct {
	logger log.Logger

	BuildNumber           *prometheus.Desc
	Characteristics       *prometheus.Desc
	DetectedCloudPlatform *prometheus.Desc
	DynamicWeight         *prometheus.Desc
	Flags                 *prometheus.Desc
	MajorVersion          *prometheus.Desc
	MinorVersion          *prometheus.Desc
	NeedsPreventQuorum    *prometheus.Desc
	NodeDrainStatus       *prometheus.Desc
	NodeHighestVersion    *prometheus.Desc
	NodeLowestVersion     *prometheus.Desc
	NodeWeight            *prometheus.Desc
	State                 *prometheus.Desc
	StatusInformation     *prometheus.Desc
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
	c.BuildNumber = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "build_number"),
		"Provides access to the node's BuildNumber property.",
		[]string{"name"},
		nil,
	)
	c.Characteristics = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "characteristics"),
		"Provides access to the characteristics set for the node.",
		[]string{"name"},
		nil,
	)
	c.DetectedCloudPlatform = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "detected_cloud_platform"),
		"(DetectedCloudPlatform)",
		[]string{"name"},
		nil,
	)
	c.DynamicWeight = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_weight"),
		"The dynamic vote weight of the node adjusted by dynamic quorum feature.",
		[]string{"name"},
		nil,
	)
	c.Flags = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "flags"),
		"Provides access to the flags set for the node.",
		[]string{"name"},
		nil,
	)
	c.MajorVersion = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "major_version"),
		"Provides access to the node's MajorVersion property, which specifies the major portion of the Windows version installed.",
		[]string{"name"},
		nil,
	)
	c.MinorVersion = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "minor_version"),
		"Provides access to the node's MinorVersion property, which specifies the minor portion of the Windows version installed.",
		[]string{"name"},
		nil,
	)
	c.NeedsPreventQuorum = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "needs_prevent_quorum"),
		"Whether the cluster service on that node should be started with prevent quorum flag.",
		[]string{"name"},
		nil,
	)
	c.NodeDrainStatus = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "node_drain_status"),
		"The current node drain status of a node. 0: Not Initiated; 1: In Progress; 2: Completed; 3: Failed",
		[]string{"name"},
		nil,
	)
	c.NodeHighestVersion = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "node_highest_version"),
		"Provides access to the node's NodeHighestVersion property, which specifies the highest possible version of the cluster service with which the node can join or communicate.",
		[]string{"name"},
		nil,
	)
	c.NodeLowestVersion = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "node_lowest_version"),
		"Provides access to the node's NodeLowestVersion property, which specifies the lowest possible version of the cluster service with which the node can join or communicate.",
		[]string{"name"},
		nil,
	)
	c.NodeWeight = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "node_weight"),
		"The vote weight of the node.",
		[]string{"name"},
		nil,
	)
	c.State = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "state"),
		"Returns the current state of a node. -1: Unknown; 0: Up; 1: Down; 2: Paused; 3: Joining",
		[]string{"name"},
		nil,
	)
	c.StatusInformation = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "status_information"),
		"The isolation or quarantine status of the node.",
		[]string{"name"},
		nil,
	)
	return nil
}

// MSCluster_Node docs:
// - https://docs.microsoft.com/en-us/previous-versions/windows/desktop/cluswmi/mscluster-node
type MSCluster_Node struct {
	Name string

	BuildNumber           uint
	Characteristics       uint
	DetectedCloudPlatform uint
	DynamicWeight         uint
	Flags                 uint
	MajorVersion          uint
	MinorVersion          uint
	NeedsPreventQuorum    uint
	NodeDrainStatus       uint
	NodeHighestVersion    uint
	NodeLowestVersion     uint
	NodeWeight            uint
	State                 uint
	StatusInformation     uint
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var dst []MSCluster_Node
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.QueryNamespace(q, &dst, "root/MSCluster"); err != nil {
		return err
	}

	NodeName = []string{}

	for _, v := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.BuildNumber,
			prometheus.GaugeValue,
			float64(v.BuildNumber),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Characteristics,
			prometheus.GaugeValue,
			float64(v.Characteristics),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DetectedCloudPlatform,
			prometheus.GaugeValue,
			float64(v.DetectedCloudPlatform),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DynamicWeight,
			prometheus.GaugeValue,
			float64(v.DynamicWeight),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Flags,
			prometheus.GaugeValue,
			float64(v.Flags),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MajorVersion,
			prometheus.GaugeValue,
			float64(v.MajorVersion),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MinorVersion,
			prometheus.GaugeValue,
			float64(v.MinorVersion),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NeedsPreventQuorum,
			prometheus.GaugeValue,
			float64(v.NeedsPreventQuorum),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NodeDrainStatus,
			prometheus.GaugeValue,
			float64(v.NodeDrainStatus),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NodeHighestVersion,
			prometheus.GaugeValue,
			float64(v.NodeHighestVersion),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NodeLowestVersion,
			prometheus.GaugeValue,
			float64(v.NodeLowestVersion),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NodeWeight,
			prometheus.GaugeValue,
			float64(v.NodeWeight),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.State,
			prometheus.GaugeValue,
			float64(v.State),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.StatusInformation,
			prometheus.GaugeValue,
			float64(v.StatusInformation),
			v.Name,
		)

		NodeName = append(NodeName, v.Name)
	}

	return nil
}

package mscluster_network

import (
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/wmi"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "mscluster_network"

type Config struct{}

var ConfigDefaults = Config{}

// A collector is a Prometheus collector for WMI MSCluster_Network metrics
type collector struct {
	logger log.Logger

	Characteristics *prometheus.Desc
	Flags           *prometheus.Desc
	Metric          *prometheus.Desc
	Role            *prometheus.Desc
	State           *prometheus.Desc
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
	c.Characteristics = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "characteristics"),
		"Provides the characteristics of the network.",
		[]string{"name"},
		nil,
	)
	c.Flags = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "flags"),
		"Provides access to the flags set for the node. ",
		[]string{"name"},
		nil,
	)
	c.Metric = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "metric"),
		"The metric of a cluster network (networks with lower values are used first). If this value is set, then the AutoMetric property is set to false.",
		[]string{"name"},
		nil,
	)
	c.Role = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "role"),
		"Provides access to the network's Role property. The Role property describes the role of the network in the cluster. 0: None; 1: Cluster; 2: Client; 3: Both ",
		[]string{"name"},
		nil,
	)
	c.State = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "state"),
		"Provides the current state of the network. 1-1: Unknown; 0: Unavailable; 1: Down; 2: Partitioned; 3: Up",
		[]string{"name"},
		nil,
	)
	return nil
}

// MSCluster_Network docs:
// - https://docs.microsoft.com/en-us/previous-versions/windows/desktop/cluswmi/mscluster-network
type MSCluster_Network struct {
	Name string

	Characteristics uint
	Flags           uint
	Metric          uint
	Role            uint
	State           uint
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var dst []MSCluster_Network
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.QueryNamespace(q, &dst, "root/MSCluster"); err != nil {
		return err
	}

	for _, v := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.Characteristics,
			prometheus.GaugeValue,
			float64(v.Characteristics),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Flags,
			prometheus.GaugeValue,
			float64(v.Flags),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Metric,
			prometheus.GaugeValue,
			float64(v.Metric),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Role,
			prometheus.GaugeValue,
			float64(v.Role),
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

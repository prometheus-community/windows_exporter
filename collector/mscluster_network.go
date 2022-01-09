package collector

import (
	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("mscluster_network", newMSCluster_NetworkCollector)
}

// A MSCluster_NetworkCollector is a Prometheus collector for WMI MSCluster_Network metrics
type MSCluster_NetworkCollector struct {
	Characteristics *prometheus.Desc
	Flags           *prometheus.Desc
	Metric          *prometheus.Desc
	Role            *prometheus.Desc
	State           *prometheus.Desc
}

func newMSCluster_NetworkCollector() (Collector, error) {
	const subsystem = "mscluster_network"
	return &MSCluster_NetworkCollector{
		Characteristics: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "characteristics"),
			"Provides the characteristics of the network.",
			[]string{"name"},
			nil,
		),
		Flags: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "flags"),
			"Provides access to the flags set for the node. ",
			[]string{"name"},
			nil,
		),
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "metric"),
			"The metric of a cluster network (networks with lower values are used first). If this value is set, then the AutoMetric property is set to false.",
			[]string{"name"},
			nil,
		),
		Role: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "role"),
			"Provides access to the network's Role property. The Role property describes the role of the network in the cluster. 0: None; 1: Cluster; 2: Client; 3: Both ",
			[]string{"name"},
			nil,
		),
		State: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "state"),
			"Provides the current state of the network. 1-1: Unknown; 0: Unavailable; 1: Down; 2: Partitioned; 3: Up",
			[]string{"name"},
			nil,
		),
	}, nil
}

// MSCluster_Network docs:
// - https://docs.microsoft.com/en-us/previous-versions/windows/desktop/cluswmi/mscluster-network
//
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
func (c *MSCluster_NetworkCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	var dst []MSCluster_Network
	q := queryAll(&dst)
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

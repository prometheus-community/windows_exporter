package mscluster_resource

import (
	"github.com/prometheus-community/windows_exporter/pkg/collector/mscluster_node"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/wmi"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "mscluster_resource"

type Config struct{}

var ConfigDefaults = Config{}

// A collector is a Prometheus collector for WMI MSCluster_Resource metrics
type collector struct {
	logger log.Logger

	Characteristics        *prometheus.Desc
	DeadlockTimeout        *prometheus.Desc
	EmbeddedFailureAction  *prometheus.Desc
	Flags                  *prometheus.Desc
	IsAlivePollInterval    *prometheus.Desc
	LooksAlivePollInterval *prometheus.Desc
	MonitorProcessId       *prometheus.Desc
	OwnerNode              *prometheus.Desc
	PendingTimeout         *prometheus.Desc
	ResourceClass          *prometheus.Desc
	RestartAction          *prometheus.Desc
	RestartDelay           *prometheus.Desc
	RestartPeriod          *prometheus.Desc
	RestartThreshold       *prometheus.Desc
	RetryPeriodOnFailure   *prometheus.Desc
	State                  *prometheus.Desc
	Subclass               *prometheus.Desc
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
		"Provides the characteristics of the object.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.DeadlockTimeout = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "deadlock_timeout"),
		"Indicates the length of time to wait, in milliseconds, before declaring a deadlock in any call into a resource.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.EmbeddedFailureAction = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "embedded_failure_action"),
		"The time, in milliseconds, that a resource should remain in a failed state before the Cluster service attempts to restart it.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.Flags = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "flags"),
		"Provides access to the flags set for the object.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.IsAlivePollInterval = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "is_alive_poll_interval"),
		"Provides access to the resource's IsAlivePollInterval property, which is the recommended interval in milliseconds at which the Cluster Service should poll the resource to determine whether it is operational. If the property is set to 0xFFFFFFFF, the Cluster Service uses the IsAlivePollInterval property for the resource type associated with the resource.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.LooksAlivePollInterval = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "looks_alive_poll_interval"),
		"Provides access to the resource's LooksAlivePollInterval property, which is the recommended interval in milliseconds at which the Cluster Service should poll the resource to determine whether it appears operational. If the property is set to 0xFFFFFFFF, the Cluster Service uses the LooksAlivePollInterval property for the resource type associated with the resource.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.MonitorProcessId = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "monitor_process_id"),
		"Provides the process ID of the resource host service that is currently hosting the resource.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.OwnerNode = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "owner_node"),
		"The node hosting the resource. 0: Not hosted; 1: Hosted",
		[]string{"type", "owner_group", "node_name", "name"},
		nil,
	)
	c.OwnerNode = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "owner_node"),
		"The node hosting the resource. 0: Not hosted; 1: Hosted",
		[]string{"type", "owner_group", "node_name", "name"},
		nil,
	)
	c.PendingTimeout = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "pending_timeout"),
		"Provides access to the resource's PendingTimeout property. If a resource cannot be brought online or taken offline in the number of milliseconds specified by the PendingTimeout property, the resource is forcibly terminated.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.ResourceClass = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "resource_class"),
		"Gets or sets the resource class of a resource. 0: Unknown; 1: Storage; 2: Network; 32768: Unknown ",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.RestartAction = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "restart_action"),
		"Provides access to the resource's RestartAction property, which is the action to be taken by the Cluster Service if the resource fails.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.RestartDelay = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "restart_delay"),
		"Indicates the time delay before a failed resource is restarted.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.RestartPeriod = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "restart_period"),
		"Provides access to the resource's RestartPeriod property, which is interval of time, in milliseconds, during which a specified number of restart attempts can be made on a nonresponsive resource.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.RestartThreshold = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "restart_threshold"),
		"Provides access to the resource's RestartThreshold property which is the maximum number of restart attempts that can be made on a resource within an interval defined by the RestartPeriod property before the Cluster Service initiates the action specified by the RestartAction property.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.RetryPeriodOnFailure = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "retry_period_on_failure"),
		"Provides access to the resource's RetryPeriodOnFailure property, which is the interval of time (in milliseconds) that a resource should remain in a failed state before the Cluster service attempts to restart it.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.State = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "state"),
		"The current state of the resource. -1: Unknown; 0: Inherited; 1: Initializing; 2: Online; 3: Offline; 4: Failed; 128: Pending; 129: Online Pending; 130: Offline Pending ",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.Subclass = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "subclass"),
		"Provides the list of references to nodes that can be the owner of this resource.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	return nil
}

// MSCluster_Resource docs:
// - https://docs.microsoft.com/en-us/previous-versions/windows/desktop/cluswmi/mscluster-resource
type MSCluster_Resource struct {
	Name       string
	Type       string
	OwnerGroup string
	OwnerNode  string

	Characteristics        uint
	DeadlockTimeout        uint
	EmbeddedFailureAction  uint
	Flags                  uint
	IsAlivePollInterval    uint
	LooksAlivePollInterval uint
	MonitorProcessId       uint
	PendingTimeout         uint
	ResourceClass          uint
	RestartAction          uint
	RestartDelay           uint
	RestartPeriod          uint
	RestartThreshold       uint
	RetryPeriodOnFailure   uint
	State                  uint
	Subclass               uint
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var dst []MSCluster_Resource
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.QueryNamespace(q, &dst, "root/MSCluster"); err != nil {
		return err
	}

	for _, v := range dst {

		ch <- prometheus.MustNewConstMetric(
			c.Characteristics,
			prometheus.GaugeValue,
			float64(v.Characteristics),
			v.Type, v.OwnerGroup, v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DeadlockTimeout,
			prometheus.GaugeValue,
			float64(v.DeadlockTimeout),
			v.Type, v.OwnerGroup, v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.EmbeddedFailureAction,
			prometheus.GaugeValue,
			float64(v.EmbeddedFailureAction),
			v.Type, v.OwnerGroup, v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Flags,
			prometheus.GaugeValue,
			float64(v.Flags),
			v.Type, v.OwnerGroup, v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.IsAlivePollInterval,
			prometheus.GaugeValue,
			float64(v.IsAlivePollInterval),
			v.Type, v.OwnerGroup, v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LooksAlivePollInterval,
			prometheus.GaugeValue,
			float64(v.LooksAlivePollInterval),
			v.Type, v.OwnerGroup, v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MonitorProcessId,
			prometheus.GaugeValue,
			float64(v.MonitorProcessId),
			v.Type, v.OwnerGroup, v.Name,
		)

		if mscluster_node.NodeName != nil {
			for _, node_name := range mscluster_node.NodeName {
				isCurrentState := 0.0
				if v.OwnerNode == node_name {
					isCurrentState = 1.0
				}
				ch <- prometheus.MustNewConstMetric(
					c.OwnerNode,
					prometheus.GaugeValue,
					isCurrentState,
					v.Type, v.OwnerGroup, node_name, v.Name,
				)
			}
		}

		ch <- prometheus.MustNewConstMetric(
			c.PendingTimeout,
			prometheus.GaugeValue,
			float64(v.PendingTimeout),
			v.Type, v.OwnerGroup, v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ResourceClass,
			prometheus.GaugeValue,
			float64(v.ResourceClass),
			v.Type, v.OwnerGroup, v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RestartAction,
			prometheus.GaugeValue,
			float64(v.RestartAction),
			v.Type, v.OwnerGroup, v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RestartDelay,
			prometheus.GaugeValue,
			float64(v.RestartDelay),
			v.Type, v.OwnerGroup, v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RestartPeriod,
			prometheus.GaugeValue,
			float64(v.RestartPeriod),
			v.Type, v.OwnerGroup, v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RestartThreshold,
			prometheus.GaugeValue,
			float64(v.RestartThreshold),
			v.Type, v.OwnerGroup, v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RetryPeriodOnFailure,
			prometheus.GaugeValue,
			float64(v.RetryPeriodOnFailure),
			v.Type, v.OwnerGroup, v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.State,
			prometheus.GaugeValue,
			float64(v.State),
			v.Type, v.OwnerGroup, v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Subclass,
			prometheus.GaugeValue,
			float64(v.Subclass),
			v.Type, v.OwnerGroup, v.Name,
		)
	}

	return nil
}

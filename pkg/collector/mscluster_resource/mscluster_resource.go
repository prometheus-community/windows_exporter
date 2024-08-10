package mscluster_resource

import (
	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/prometheus-community/windows_exporter/pkg/collector/mscluster_node"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "mscluster_resource"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI MSCluster_Resource metrics.
type Collector struct {
	logger log.Logger

	characteristics        *prometheus.Desc
	deadlockTimeout        *prometheus.Desc
	embeddedFailureAction  *prometheus.Desc
	flags                  *prometheus.Desc
	isAlivePollInterval    *prometheus.Desc
	looksAlivePollInterval *prometheus.Desc
	monitorProcessId       *prometheus.Desc
	ownerNode              *prometheus.Desc
	pendingTimeout         *prometheus.Desc
	resourceClass          *prometheus.Desc
	restartAction          *prometheus.Desc
	restartDelay           *prometheus.Desc
	restartPeriod          *prometheus.Desc
	restartThreshold       *prometheus.Desc
	retryPeriodOnFailure   *prometheus.Desc
	state                  *prometheus.Desc
	subclass               *prometheus.Desc
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
	c.characteristics = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "characteristics"),
		"Provides the characteristics of the object.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.deadlockTimeout = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "deadlock_timeout"),
		"Indicates the length of time to wait, in milliseconds, before declaring a deadlock in any call into a resource.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.embeddedFailureAction = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "embedded_failure_action"),
		"The time, in milliseconds, that a resource should remain in a failed state before the Cluster service attempts to restart it.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.flags = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "flags"),
		"Provides access to the flags set for the object.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.isAlivePollInterval = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "is_alive_poll_interval"),
		"Provides access to the resource's IsAlivePollInterval property, which is the recommended interval in milliseconds at which the Cluster Service should poll the resource to determine whether it is operational. If the property is set to 0xFFFFFFFF, the Cluster Service uses the IsAlivePollInterval property for the resource type associated with the resource.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.looksAlivePollInterval = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "looks_alive_poll_interval"),
		"Provides access to the resource's LooksAlivePollInterval property, which is the recommended interval in milliseconds at which the Cluster Service should poll the resource to determine whether it appears operational. If the property is set to 0xFFFFFFFF, the Cluster Service uses the LooksAlivePollInterval property for the resource type associated with the resource.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.monitorProcessId = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "monitor_process_id"),
		"Provides the process ID of the resource host service that is currently hosting the resource.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.ownerNode = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "owner_node"),
		"The node hosting the resource. 0: Not hosted; 1: Hosted",
		[]string{"type", "owner_group", "node_name", "name"},
		nil,
	)
	c.ownerNode = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "owner_node"),
		"The node hosting the resource. 0: Not hosted; 1: Hosted",
		[]string{"type", "owner_group", "node_name", "name"},
		nil,
	)
	c.pendingTimeout = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "pending_timeout"),
		"Provides access to the resource's PendingTimeout property. If a resource cannot be brought online or taken offline in the number of milliseconds specified by the PendingTimeout property, the resource is forcibly terminated.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.resourceClass = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "resource_class"),
		"Gets or sets the resource class of a resource. 0: Unknown; 1: Storage; 2: Network; 32768: Unknown ",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.restartAction = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "restart_action"),
		"Provides access to the resource's RestartAction property, which is the action to be taken by the Cluster Service if the resource fails.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.restartDelay = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "restart_delay"),
		"Indicates the time delay before a failed resource is restarted.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.restartPeriod = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "restart_period"),
		"Provides access to the resource's RestartPeriod property, which is interval of time, in milliseconds, during which a specified number of restart attempts can be made on a nonresponsive resource.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.restartThreshold = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "restart_threshold"),
		"Provides access to the resource's RestartThreshold property which is the maximum number of restart attempts that can be made on a resource within an interval defined by the RestartPeriod property before the Cluster Service initiates the action specified by the RestartAction property.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.retryPeriodOnFailure = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "retry_period_on_failure"),
		"Provides access to the resource's RetryPeriodOnFailure property, which is the interval of time (in milliseconds) that a resource should remain in a failed state before the Cluster service attempts to restart it.",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.state = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "state"),
		"The current state of the resource. -1: Unknown; 0: Inherited; 1: Initializing; 2: Online; 3: Offline; 4: Failed; 128: Pending; 129: Online Pending; 130: Offline Pending ",
		[]string{"type", "owner_group", "name"},
		nil,
	)
	c.subclass = prometheus.NewDesc(
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
func (c *Collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var dst []MSCluster_Resource
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.QueryNamespace(q, &dst, "root/MSCluster"); err != nil {
		return err
	}

	for _, v := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.characteristics,
			prometheus.GaugeValue,
			float64(v.Characteristics),
			v.Type, v.OwnerGroup, v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.deadlockTimeout,
			prometheus.GaugeValue,
			float64(v.DeadlockTimeout),
			v.Type, v.OwnerGroup, v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.embeddedFailureAction,
			prometheus.GaugeValue,
			float64(v.EmbeddedFailureAction),
			v.Type, v.OwnerGroup, v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.flags,
			prometheus.GaugeValue,
			float64(v.Flags),
			v.Type, v.OwnerGroup, v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.isAlivePollInterval,
			prometheus.GaugeValue,
			float64(v.IsAlivePollInterval),
			v.Type, v.OwnerGroup, v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.looksAlivePollInterval,
			prometheus.GaugeValue,
			float64(v.LooksAlivePollInterval),
			v.Type, v.OwnerGroup, v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.monitorProcessId,
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
					c.ownerNode,
					prometheus.GaugeValue,
					isCurrentState,
					v.Type, v.OwnerGroup, node_name, v.Name,
				)
			}
		}

		ch <- prometheus.MustNewConstMetric(
			c.pendingTimeout,
			prometheus.GaugeValue,
			float64(v.PendingTimeout),
			v.Type, v.OwnerGroup, v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.resourceClass,
			prometheus.GaugeValue,
			float64(v.ResourceClass),
			v.Type, v.OwnerGroup, v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.restartAction,
			prometheus.GaugeValue,
			float64(v.RestartAction),
			v.Type, v.OwnerGroup, v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.restartDelay,
			prometheus.GaugeValue,
			float64(v.RestartDelay),
			v.Type, v.OwnerGroup, v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.restartPeriod,
			prometheus.GaugeValue,
			float64(v.RestartPeriod),
			v.Type, v.OwnerGroup, v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.restartThreshold,
			prometheus.GaugeValue,
			float64(v.RestartThreshold),
			v.Type, v.OwnerGroup, v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.retryPeriodOnFailure,
			prometheus.GaugeValue,
			float64(v.RetryPeriodOnFailure),
			v.Type, v.OwnerGroup, v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.state,
			prometheus.GaugeValue,
			float64(v.State),
			v.Type, v.OwnerGroup, v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.subclass,
			prometheus.GaugeValue,
			float64(v.Subclass),
			v.Type, v.OwnerGroup, v.Name,
		)
	}

	return nil
}

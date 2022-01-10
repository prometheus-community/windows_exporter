package collector

import (
	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("mscluster_resource", newMSCluster_ResourceCollector)
}

// A MSCluster_ResourceCollector is a Prometheus collector for WMI MSCluster_Resource metrics
type MSCluster_ResourceCollector struct {
	Characteristics        *prometheus.Desc
	DeadlockTimeout        *prometheus.Desc
	EmbeddedFailureAction  *prometheus.Desc
	Flags                  *prometheus.Desc
	IsAlivePollInterval    *prometheus.Desc
	LooksAlivePollInterval *prometheus.Desc
	MonitorProcessId       *prometheus.Desc
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

func newMSCluster_ResourceCollector() (Collector, error) {
	const subsystem = "mscluster_resource"
	return &MSCluster_ResourceCollector{
		Characteristics: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "characteristics"),
			"Provides the characteristics of the object.",
			[]string{"type", "owner_group", "name"},
			nil,
		),
		DeadlockTimeout: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "deadlock_timeout"),
			"Indicates the length of time to wait, in milliseconds, before declaring a deadlock in any call into a resource.",
			[]string{"type", "owner_group", "name"},
			nil,
		),
		EmbeddedFailureAction: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "embedded_failure_action"),
			"The time, in milliseconds, that a resource should remain in a failed state before the Cluster service attempts to restart it.",
			[]string{"type", "owner_group", "name"},
			nil,
		),
		Flags: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "flags"),
			"Provides access to the flags set for the object.",
			[]string{"type", "owner_group", "name"},
			nil,
		),
		IsAlivePollInterval: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "is_alive_poll_interval"),
			"Provides access to the resource's IsAlivePollInterval property, which is the recommended interval in milliseconds at which the Cluster Service should poll the resource to determine whether it is operational. If the property is set to 0xFFFFFFFF, the Cluster Service uses the IsAlivePollInterval property for the resource type associated with the resource.",
			[]string{"type", "owner_group", "name"},
			nil,
		),
		LooksAlivePollInterval: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "looks_alive_poll_interval"),
			"Provides access to the resource's LooksAlivePollInterval property, which is the recommended interval in milliseconds at which the Cluster Service should poll the resource to determine whether it appears operational. If the property is set to 0xFFFFFFFF, the Cluster Service uses the LooksAlivePollInterval property for the resource type associated with the resource.",
			[]string{"type", "owner_group", "name"},
			nil,
		),
		MonitorProcessId: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "monitor_process_id"),
			"Provides the process ID of the resource host service that is currently hosting the resource.",
			[]string{"type", "owner_group", "name"},
			nil,
		),
		PendingTimeout: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pending_timeout"),
			"Provides access to the resource's PendingTimeout property. If a resource cannot be brought online or taken offline in the number of milliseconds specified by the PendingTimeout property, the resource is forcibly terminated.",
			[]string{"type", "owner_group", "name"},
			nil,
		),
		ResourceClass: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "resource_class"),
			"Gets or sets the resource class of a resource. 0: Unknown; 1: Storage; 2: Network; 32768: Unknown ",
			[]string{"type", "owner_group", "name"},
			nil,
		),
		RestartAction: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "restart_action"),
			"Provides access to the resource's RestartAction property, which is the action to be taken by the Cluster Service if the resource fails.",
			[]string{"type", "owner_group", "name"},
			nil,
		),
		RestartDelay: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "restart_delay"),
			"Indicates the time delay before a failed resource is restarted.",
			[]string{"type", "owner_group", "name"},
			nil,
		),
		RestartPeriod: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "restart_period"),
			"Provides access to the resource's RestartPeriod property, which is interval of time, in milliseconds, during which a specified number of restart attempts can be made on a nonresponsive resource.",
			[]string{"type", "owner_group", "name"},
			nil,
		),
		RestartThreshold: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "restart_threshold"),
			"Provides access to the resource's RestartThreshold property which is the maximum number of restart attempts that can be made on a resource within an interval defined by the RestartPeriod property before the Cluster Service initiates the action specified by the RestartAction property.",
			[]string{"type", "owner_group", "name"},
			nil,
		),
		RetryPeriodOnFailure: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "retry_period_on_failure"),
			"Provides access to the resource's RetryPeriodOnFailure property, which is the interval of time (in milliseconds) that a resource should remain in a failed state before the Cluster service attempts to restart it.",
			[]string{"type", "owner_group", "name"},
			nil,
		),
		State: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "state"),
			"The current state of the resource. -1: Unknown; 0: Inherited; 1: Initializing; 2: Online; 3: Offline; 4: Failed; 128: Pending; 129: Online Pending; 130: Offline Pending ",
			[]string{"type", "owner_group", "name"},
			nil,
		),
		Subclass: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "subclass"),
			"Provides the list of references to nodes that can be the owner of this resource.",
			[]string{"type", "owner_group", "name"},
			nil,
		),
	}, nil
}

// MSCluster_Resource docs:
// - https://docs.microsoft.com/en-us/previous-versions/windows/desktop/cluswmi/mscluster-resource
//
type MSCluster_Resource struct {
	Name       string
	Type       string
	OwnerGroup string

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
func (c *MSCluster_ResourceCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	var dst []MSCluster_Resource
	q := queryAll(&dst)
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

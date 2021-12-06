package collector

import (
	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("mscluster_resource", newMSCluster_ResourceCollector) // TODO: Add any perflib dependencies here
}

// A MSCluster_ResourceCollector is a Prometheus collector for WMI MSCluster_Resource metrics
type MSCluster_ResourceCollector struct {
	Characteristics           *prometheus.Desc
	DeadlockTimeout           *prometheus.Desc
	EmbeddedFailureAction     *prometheus.Desc
	Flags                     *prometheus.Desc
	IsAlivePollInterval       *prometheus.Desc
	LooksAlivePollInterval    *prometheus.Desc
	MonitorProcessId          *prometheus.Desc
	PendingTimeout            *prometheus.Desc
	RequiredDependencyClasses *prometheus.Desc
	ResourceClass             *prometheus.Desc
	RestartAction             *prometheus.Desc
	RestartDelay              *prometheus.Desc
	RestartPeriod             *prometheus.Desc
	RestartThreshold          *prometheus.Desc
	RetryPeriodOnFailure      *prometheus.Desc
	State                     *prometheus.Desc
	Subclass                  *prometheus.Desc
}

func newMSCluster_ResourceCollector() (Collector, error) {
	const subsystem = "mscluster_resource"
	return &MSCluster_ResourceCollector{
		Characteristics: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "characteristics"),
			"(Characteristics)",
			[]string{"type", "owner_group", "name"},
			nil,
		),
		DeadlockTimeout: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "deadlock_timeout"),
			"(DeadlockTimeout)",
			[]string{"type", "owner_group", "name"},
			nil,
		),
		EmbeddedFailureAction: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "embedded_failure_action"),
			"(EmbeddedFailureAction)",
			[]string{"type", "owner_group", "name"},
			nil,
		),
		Flags: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "flags"),
			"(Flags)",
			[]string{"type", "owner_group", "name"},
			nil,
		),
		IsAlivePollInterval: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "is_alive_poll_interval"),
			"(IsAlivePollInterval)",
			[]string{"type", "owner_group", "name"},
			nil,
		),
		LooksAlivePollInterval: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "looks_alive_poll_interval"),
			"(LooksAlivePollInterval)",
			[]string{"type", "owner_group", "name"},
			nil,
		),
		MonitorProcessId: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "monitor_process_id"),
			"(MonitorProcessId)",
			[]string{"type", "owner_group", "name"},
			nil,
		),
		PendingTimeout: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pending_timeout"),
			"(PendingTimeout)",
			[]string{"type", "owner_group", "name"},
			nil,
		),
		ResourceClass: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "resource_class"),
			"(ResourceClass)",
			[]string{"type", "owner_group", "name"},
			nil,
		),
		RestartAction: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "restart_action"),
			"(RestartAction)",
			[]string{"type", "owner_group", "name"},
			nil,
		),
		RestartDelay: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "restart_delay"),
			"(RestartDelay)",
			[]string{"type", "owner_group", "name"},
			nil,
		),
		RestartPeriod: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "restart_period"),
			"(RestartPeriod)",
			[]string{"type", "owner_group", "name"},
			nil,
		),
		RestartThreshold: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "restart_threshold"),
			"(RestartThreshold)",
			[]string{"type", "owner_group", "name"},
			nil,
		),
		RetryPeriodOnFailure: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "retry_period_on_failure"),
			"(RetryPeriodOnFailure)",
			[]string{"type", "owner_group", "name"},
			nil,
		),
		State: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "state"),
			"(State)",
			[]string{"type", "owner_group", "name"},
			nil,
		),
		Subclass: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "subclass"),
			"(Subclass)",
			[]string{"type", "owner_group", "name"},
			nil,
		),
	}, nil
}

// MSCluster_Resource docs:
// - <add link to documentation here>
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

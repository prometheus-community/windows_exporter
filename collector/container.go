//go:build windows
// +build windows

package collector

import (
	"github.com/Microsoft/hcsshim"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("container", NewContainerMetricsCollector)
}

// A ContainerMetricsCollector is a Prometheus collector for containers metrics
type ContainerMetricsCollector struct {
	// Presence
	ContainerAvailable *prometheus.Desc

	// Number of containers
	ContainersCount *prometheus.Desc
	// memory
	UsageCommitBytes            *prometheus.Desc
	UsageCommitPeakBytes        *prometheus.Desc
	UsagePrivateWorkingSetBytes *prometheus.Desc

	// CPU
	RuntimeTotal  *prometheus.Desc
	RuntimeUser   *prometheus.Desc
	RuntimeKernel *prometheus.Desc

	// Network
	BytesReceived          *prometheus.Desc
	BytesSent              *prometheus.Desc
	PacketsReceived        *prometheus.Desc
	PacketsSent            *prometheus.Desc
	DroppedPacketsIncoming *prometheus.Desc
	DroppedPacketsOutgoing *prometheus.Desc

	// Storage
	ReadCountNormalized  *prometheus.Desc
	ReadSizeBytes        *prometheus.Desc
	WriteCountNormalized *prometheus.Desc
	WriteSizeBytes       *prometheus.Desc
}

// NewContainerMetricsCollector constructs a new ContainerMetricsCollector
func NewContainerMetricsCollector() (Collector, error) {
	const subsystem = "container"
	return &ContainerMetricsCollector{
		ContainerAvailable: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "available"),
			"Available",
			[]string{"container_id"},
			nil,
		),
		ContainersCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "count"),
			"Number of containers",
			nil,
			nil,
		),
		UsageCommitBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memory_usage_commit_bytes"),
			"Memory Usage Commit Bytes",
			[]string{"container_id"},
			nil,
		),
		UsageCommitPeakBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memory_usage_commit_peak_bytes"),
			"Memory Usage Commit Peak Bytes",
			[]string{"container_id"},
			nil,
		),
		UsagePrivateWorkingSetBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memory_usage_private_working_set_bytes"),
			"Memory Usage Private Working Set Bytes",
			[]string{"container_id"},
			nil,
		),
		RuntimeTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cpu_usage_seconds_total"),
			"Total Run time in Seconds",
			[]string{"container_id"},
			nil,
		),
		RuntimeUser: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cpu_usage_seconds_usermode"),
			"Run Time in User mode in Seconds",
			[]string{"container_id"},
			nil,
		),
		RuntimeKernel: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cpu_usage_seconds_kernelmode"),
			"Run time in Kernel mode in Seconds",
			[]string{"container_id"},
			nil,
		),
		BytesReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "network_receive_bytes_total"),
			"Bytes Received on Interface",
			[]string{"container_id", "interface"},
			nil,
		),
		BytesSent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "network_transmit_bytes_total"),
			"Bytes Sent on Interface",
			[]string{"container_id", "interface"},
			nil,
		),
		PacketsReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "network_receive_packets_total"),
			"Packets Received on Interface",
			[]string{"container_id", "interface"},
			nil,
		),
		PacketsSent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "network_transmit_packets_total"),
			"Packets Sent on Interface",
			[]string{"container_id", "interface"},
			nil,
		),
		DroppedPacketsIncoming: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "network_receive_packets_dropped_total"),
			"Dropped Incoming Packets on Interface",
			[]string{"container_id", "interface"},
			nil,
		),
		DroppedPacketsOutgoing: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "network_transmit_packets_dropped_total"),
			"Dropped Outgoing Packets on Interface",
			[]string{"container_id", "interface"},
			nil,
		),
		ReadCountNormalized: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "storage_read_count_normalized_total"),
			"Read Count Normalized",
			[]string{"container_id"},
			nil,
		),
		ReadSizeBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "storage_read_size_bytes_total"),
			"Read Size Bytes",
			[]string{"container_id"},
			nil,
		),
		WriteCountNormalized: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "storage_write_count_normalized_total"),
			"Write Count Normalized",
			[]string{"container_id"},
			nil,
		),
		WriteSizeBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "storage_write_size_bytes_total"),
			"Write Size Bytes",
			[]string{"container_id"},
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *ContainerMetricsCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting ContainerMetricsCollector metrics:", desc, err)
		return err
	}
	return nil
}

// containerClose closes the container resource
func containerClose(c hcsshim.Container) {
	err := c.Close()
	if err != nil {
		log.Error(err)
	}
}

func (c *ContainerMetricsCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {

	// Types Container is passed to get the containers compute systems only
	containers, err := hcsshim.GetContainers(hcsshim.ComputeSystemQuery{Types: []string{"Container"}})
	if err != nil {
		log.Error("Err in Getting containers:", err)
		return nil, err
	}

	count := len(containers)

	ch <- prometheus.MustNewConstMetric(
		c.ContainersCount,
		prometheus.GaugeValue,
		float64(count),
	)
	if count == 0 {
		return nil, nil
	}

	for _, containerDetails := range containers {
		container, err := hcsshim.OpenContainer(containerDetails.ID)
		if container != nil {
			defer containerClose(container)
		}
		if err != nil {
			log.Error("err in opening container: ", containerDetails.ID, err)
			continue
		}

		cstats, err := container.Statistics()
		if err != nil {
			log.Error("err in fetching container Statistics: ", containerDetails.ID, err)
			continue
		}
		containerIdWithPrefix := getContainerIdWithPrefix(containerDetails)

		ch <- prometheus.MustNewConstMetric(
			c.ContainerAvailable,
			prometheus.CounterValue,
			1,
			containerIdWithPrefix,
		)
		ch <- prometheus.MustNewConstMetric(
			c.UsageCommitBytes,
			prometheus.GaugeValue,
			float64(cstats.Memory.UsageCommitBytes),
			containerIdWithPrefix,
		)
		ch <- prometheus.MustNewConstMetric(
			c.UsageCommitPeakBytes,
			prometheus.GaugeValue,
			float64(cstats.Memory.UsageCommitPeakBytes),
			containerIdWithPrefix,
		)
		ch <- prometheus.MustNewConstMetric(
			c.UsagePrivateWorkingSetBytes,
			prometheus.GaugeValue,
			float64(cstats.Memory.UsagePrivateWorkingSetBytes),
			containerIdWithPrefix,
		)
		ch <- prometheus.MustNewConstMetric(
			c.RuntimeTotal,
			prometheus.CounterValue,
			float64(cstats.Processor.TotalRuntime100ns)*ticksToSecondsScaleFactor,
			containerIdWithPrefix,
		)
		ch <- prometheus.MustNewConstMetric(
			c.RuntimeUser,
			prometheus.CounterValue,
			float64(cstats.Processor.RuntimeUser100ns)*ticksToSecondsScaleFactor,
			containerIdWithPrefix,
		)
		ch <- prometheus.MustNewConstMetric(
			c.RuntimeKernel,
			prometheus.CounterValue,
			float64(cstats.Processor.RuntimeKernel100ns)*ticksToSecondsScaleFactor,
			containerIdWithPrefix,
		)

		if len(cstats.Network) == 0 {
			log.Info("No Network Stats for container: ", containerDetails.ID)
			continue
		}

		networkStats := cstats.Network

		for _, networkInterface := range networkStats {
			ch <- prometheus.MustNewConstMetric(
				c.BytesReceived,
				prometheus.CounterValue,
				float64(networkInterface.BytesReceived),
				containerIdWithPrefix, networkInterface.EndpointId,
			)
			ch <- prometheus.MustNewConstMetric(
				c.BytesSent,
				prometheus.CounterValue,
				float64(networkInterface.BytesSent),
				containerIdWithPrefix, networkInterface.EndpointId,
			)
			ch <- prometheus.MustNewConstMetric(
				c.PacketsReceived,
				prometheus.CounterValue,
				float64(networkInterface.PacketsReceived),
				containerIdWithPrefix, networkInterface.EndpointId,
			)
			ch <- prometheus.MustNewConstMetric(
				c.PacketsSent,
				prometheus.CounterValue,
				float64(networkInterface.PacketsSent),
				containerIdWithPrefix, networkInterface.EndpointId,
			)
			ch <- prometheus.MustNewConstMetric(
				c.DroppedPacketsIncoming,
				prometheus.CounterValue,
				float64(networkInterface.DroppedPacketsIncoming),
				containerIdWithPrefix, networkInterface.EndpointId,
			)
			ch <- prometheus.MustNewConstMetric(
				c.DroppedPacketsOutgoing,
				prometheus.CounterValue,
				float64(networkInterface.DroppedPacketsOutgoing),
				containerIdWithPrefix, networkInterface.EndpointId,
			)
			break
		}

		ch <- prometheus.MustNewConstMetric(
			c.ReadCountNormalized,
			prometheus.CounterValue,
			float64(cstats.Storage.ReadCountNormalized),
			containerIdWithPrefix,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ReadSizeBytes,
			prometheus.CounterValue,
			float64(cstats.Storage.ReadSizeBytes),
			containerIdWithPrefix,
		)
		ch <- prometheus.MustNewConstMetric(
			c.WriteCountNormalized,
			prometheus.CounterValue,
			float64(cstats.Storage.WriteCountNormalized),
			containerIdWithPrefix,
		)
		ch <- prometheus.MustNewConstMetric(
			c.WriteSizeBytes,
			prometheus.CounterValue,
			float64(cstats.Storage.WriteSizeBytes),
			containerIdWithPrefix,
		)
	}

	return nil, nil
}

func getContainerIdWithPrefix(containerDetails hcsshim.ContainerProperties) string {
	switch containerDetails.Owner {
	case "containerd-shim-runhcs-v1.exe":
		return "containerd://" + containerDetails.ID
	default:
		// default to docker or if owner is not set
		return "docker://" + containerDetails.ID
	}
}

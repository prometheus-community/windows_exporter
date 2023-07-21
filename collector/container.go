//go:build windows
// +build windows

package collector

import (
	"fmt"
	"strings"

	"github.com/Microsoft/hcsshim"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// A ContainerMetricsCollector is a Prometheus collector for containers metrics
type ContainerMetricsCollector struct {
	logger log.Logger

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

// newContainerMetricsCollector constructs a new ContainerMetricsCollector
func newContainerMetricsCollector(logger log.Logger) (Collector, error) {
	const subsystem = "container"
	return &ContainerMetricsCollector{
		logger: log.With(logger, "collector", subsystem),

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
		_ = level.Error(c.logger).Log("msg", "failed collecting ContainerMetricsCollector metrics", "desc", desc, "err", err)
		return err
	}
	return nil
}

// containerClose closes the container resource
func (c *ContainerMetricsCollector) containerClose(container hcsshim.Container) {
	err := container.Close()
	if err != nil {
		_ = level.Error(c.logger).Log("err", err)
	}
}

func (c *ContainerMetricsCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {

	// Types Container is passed to get the containers compute systems only
	containers, err := hcsshim.GetContainers(hcsshim.ComputeSystemQuery{Types: []string{"Container"}})
	if err != nil {
		_ = level.Error(c.logger).Log("msg", "Err in Getting containers", "err", err)
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

	containerPrefixes := make(map[string]string)

	for _, containerDetails := range containers {
		container, err := hcsshim.OpenContainer(containerDetails.ID)
		if container != nil {
			defer c.containerClose(container)
		}
		if err != nil {
			_ = level.Error(c.logger).Log("msg", "err in opening container", "containerId", containerDetails.ID, "err", err)
			continue
		}

		cstats, err := container.Statistics()
		if err != nil {
			_ = level.Error(c.logger).Log("msg", "err in fetching container Statistics", "containerId", containerDetails.ID, "err", err)
			continue
		}

		containerIdWithPrefix := getContainerIdWithPrefix(containerDetails)
		containerPrefixes[containerDetails.ID] = containerIdWithPrefix

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

	hnsEndpoints, err := hcsshim.HNSListEndpointRequest()
	if err != nil {
		_ = level.Warn(c.logger).Log("msg", "Failed to collect network stats for containers")
		return nil, nil
	}

	if len(hnsEndpoints) == 0 {
		_ = level.Info(c.logger).Log("msg", fmt.Sprintf("No network stats for containers to collect"))
		return nil, nil
	}

	for _, endpoint := range hnsEndpoints {
		endpointStats, err := hcsshim.GetHNSEndpointStats(endpoint.Id)
		if err != nil {
			_ = level.Warn(c.logger).Log("msg", fmt.Sprintf("Failed to collect network stats for interface %s", endpoint.Id), "err", err)
			continue
		}

		for _, containerId := range endpoint.SharedContainers {
			containerIdWithPrefix, ok := containerPrefixes[containerId]
			endpointId := strings.ToUpper(endpoint.Id)

			if !ok {
				_ = level.Warn(c.logger).Log("msg", fmt.Sprintf("Failed to collect network stats for container %s", containerId))
				continue
			}

			ch <- prometheus.MustNewConstMetric(
				c.BytesReceived,
				prometheus.CounterValue,
				float64(endpointStats.BytesReceived),
				containerIdWithPrefix, endpointId,
			)

			ch <- prometheus.MustNewConstMetric(
				c.BytesSent,
				prometheus.CounterValue,
				float64(endpointStats.BytesSent),
				containerIdWithPrefix, endpointId,
			)
			ch <- prometheus.MustNewConstMetric(
				c.PacketsReceived,
				prometheus.CounterValue,
				float64(endpointStats.PacketsReceived),
				containerIdWithPrefix, endpointId,
			)
			ch <- prometheus.MustNewConstMetric(
				c.PacketsSent,
				prometheus.CounterValue,
				float64(endpointStats.PacketsSent),
				containerIdWithPrefix, endpointId,
			)
			ch <- prometheus.MustNewConstMetric(
				c.DroppedPacketsIncoming,
				prometheus.CounterValue,
				float64(endpointStats.DroppedPacketsIncoming),
				containerIdWithPrefix, endpointId,
			)
			ch <- prometheus.MustNewConstMetric(
				c.DroppedPacketsOutgoing,
				prometheus.CounterValue,
				float64(endpointStats.DroppedPacketsOutgoing),
				containerIdWithPrefix, endpointId,
			)
		}
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

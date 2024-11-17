//go:build windows

package container

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Microsoft/hcsshim"
	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "container"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for containers metrics.
type Collector struct {
	config Config

	logger *slog.Logger

	// Presence
	containerAvailable *prometheus.Desc

	// Number of containers
	containersCount *prometheus.Desc

	// Memory
	usageCommitBytes            *prometheus.Desc
	usageCommitPeakBytes        *prometheus.Desc
	usagePrivateWorkingSetBytes *prometheus.Desc

	// CPU
	runtimeTotal  *prometheus.Desc
	runtimeUser   *prometheus.Desc
	runtimeKernel *prometheus.Desc

	// Network
	bytesReceived          *prometheus.Desc
	bytesSent              *prometheus.Desc
	packetsReceived        *prometheus.Desc
	packetsSent            *prometheus.Desc
	droppedPacketsIncoming *prometheus.Desc
	droppedPacketsOutgoing *prometheus.Desc

	// Storage
	readCountNormalized  *prometheus.Desc
	readSizeBytes        *prometheus.Desc
	writeCountNormalized *prometheus.Desc
	writeSizeBytes       *prometheus.Desc
}

// New constructs a new Collector.
func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	c.logger = logger.With(slog.String("collector", Name))

	c.containerAvailable = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "available"),
		"Available",
		[]string{"container_id"},
		nil,
	)
	c.containersCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "count"),
		"Number of containers",
		nil,
		nil,
	)
	c.usageCommitBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memory_usage_commit_bytes"),
		"Memory Usage Commit Bytes",
		[]string{"container_id"},
		nil,
	)
	c.usageCommitPeakBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memory_usage_commit_peak_bytes"),
		"Memory Usage Commit Peak Bytes",
		[]string{"container_id"},
		nil,
	)
	c.usagePrivateWorkingSetBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memory_usage_private_working_set_bytes"),
		"Memory Usage Private Working Set Bytes",
		[]string{"container_id"},
		nil,
	)
	c.runtimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cpu_usage_seconds_total"),
		"Total Run time in Seconds",
		[]string{"container_id"},
		nil,
	)
	c.runtimeUser = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cpu_usage_seconds_usermode"),
		"Run Time in User mode in Seconds",
		[]string{"container_id"},
		nil,
	)
	c.runtimeKernel = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cpu_usage_seconds_kernelmode"),
		"Run time in Kernel mode in Seconds",
		[]string{"container_id"},
		nil,
	)
	c.bytesReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "network_receive_bytes_total"),
		"Bytes Received on Interface",
		[]string{"container_id", "interface"},
		nil,
	)
	c.bytesSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "network_transmit_bytes_total"),
		"Bytes Sent on Interface",
		[]string{"container_id", "interface"},
		nil,
	)
	c.packetsReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "network_receive_packets_total"),
		"Packets Received on Interface",
		[]string{"container_id", "interface"},
		nil,
	)
	c.packetsSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "network_transmit_packets_total"),
		"Packets Sent on Interface",
		[]string{"container_id", "interface"},
		nil,
	)
	c.droppedPacketsIncoming = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "network_receive_packets_dropped_total"),
		"Dropped Incoming Packets on Interface",
		[]string{"container_id", "interface"},
		nil,
	)
	c.droppedPacketsOutgoing = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "network_transmit_packets_dropped_total"),
		"Dropped Outgoing Packets on Interface",
		[]string{"container_id", "interface"},
		nil,
	)
	c.readCountNormalized = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "storage_read_count_normalized_total"),
		"Read Count Normalized",
		[]string{"container_id"},
		nil,
	)
	c.readSizeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "storage_read_size_bytes_total"),
		"Read Size Bytes",
		[]string{"container_id"},
		nil,
	)
	c.writeCountNormalized = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "storage_write_count_normalized_total"),
		"Write Count Normalized",
		[]string{"container_id"},
		nil,
	)
	c.writeSizeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "storage_write_size_bytes_total"),
		"Write Size Bytes",
		[]string{"container_id"},
		nil,
	)

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	// Types Container is passed to get the containers compute systems only
	containers, err := hcsshim.GetContainers(hcsshim.ComputeSystemQuery{Types: []string{"Container"}})
	if err != nil {
		return fmt.Errorf("error in fetching containers: %w", err)
	}

	count := len(containers)

	ch <- prometheus.MustNewConstMetric(
		c.containersCount,
		prometheus.GaugeValue,
		float64(count),
	)

	if count == 0 {
		return nil
	}

	containerPrefixes := make(map[string]string)
	collectErrors := make([]error, 0, len(containers))

	for _, containerDetails := range containers {
		containerIdWithPrefix := getContainerIdWithPrefix(containerDetails)

		if err = c.collectContainer(ch, containerDetails, containerIdWithPrefix); err != nil {
			if hcsshim.IsNotExist(err) {
				c.logger.Debug("err in fetching container statistics",
					slog.String("container_id", containerDetails.ID),
					slog.Any("err", err),
				)
			} else {
				c.logger.Error("err in fetching container statistics",
					slog.String("container_id", containerDetails.ID),
					slog.Any("err", err),
				)

				collectErrors = append(collectErrors, err)
			}

			continue
		}

		containerPrefixes[containerDetails.ID] = containerIdWithPrefix
	}

	if err = c.collectNetworkMetrics(ch, containerPrefixes); err != nil {
		return fmt.Errorf("error in fetching container network statistics: %w", err)
	}

	if len(collectErrors) > 0 {
		return fmt.Errorf("errors while fetching container statistics: %w", errors.Join(collectErrors...))
	}

	return nil
}

func (c *Collector) collectContainer(ch chan<- prometheus.Metric, containerDetails hcsshim.ContainerProperties, containerIdWithPrefix string) error {
	container, err := hcsshim.OpenContainer(containerDetails.ID)
	if err != nil {
		return fmt.Errorf("error in opening container: %w", err)
	}

	defer func() {
		if container == nil {
			return
		}

		if err := container.Close(); err != nil {
			c.logger.Error("error in closing container",
				slog.Any("err", err),
			)
		}
	}()

	containerStats, err := container.Statistics()
	if err != nil {
		return fmt.Errorf("error in fetching container statistics: %w", err)
	}

	ch <- prometheus.MustNewConstMetric(
		c.containerAvailable,
		prometheus.CounterValue,
		1,
		containerIdWithPrefix,
	)
	ch <- prometheus.MustNewConstMetric(
		c.usageCommitBytes,
		prometheus.GaugeValue,
		float64(containerStats.Memory.UsageCommitBytes),
		containerIdWithPrefix,
	)
	ch <- prometheus.MustNewConstMetric(
		c.usageCommitPeakBytes,
		prometheus.GaugeValue,
		float64(containerStats.Memory.UsageCommitPeakBytes),
		containerIdWithPrefix,
	)
	ch <- prometheus.MustNewConstMetric(
		c.usagePrivateWorkingSetBytes,
		prometheus.GaugeValue,
		float64(containerStats.Memory.UsagePrivateWorkingSetBytes),
		containerIdWithPrefix,
	)
	ch <- prometheus.MustNewConstMetric(
		c.runtimeTotal,
		prometheus.CounterValue,
		float64(containerStats.Processor.TotalRuntime100ns)*perfdata.TicksToSecondScaleFactor,
		containerIdWithPrefix,
	)
	ch <- prometheus.MustNewConstMetric(
		c.runtimeUser,
		prometheus.CounterValue,
		float64(containerStats.Processor.RuntimeUser100ns)*perfdata.TicksToSecondScaleFactor,
		containerIdWithPrefix,
	)
	ch <- prometheus.MustNewConstMetric(
		c.runtimeKernel,
		prometheus.CounterValue,
		float64(containerStats.Processor.RuntimeKernel100ns)*perfdata.TicksToSecondScaleFactor,
		containerIdWithPrefix,
	)
	ch <- prometheus.MustNewConstMetric(
		c.readCountNormalized,
		prometheus.CounterValue,
		float64(containerStats.Storage.ReadCountNormalized),
		containerIdWithPrefix,
	)
	ch <- prometheus.MustNewConstMetric(
		c.readSizeBytes,
		prometheus.CounterValue,
		float64(containerStats.Storage.ReadSizeBytes),
		containerIdWithPrefix,
	)
	ch <- prometheus.MustNewConstMetric(
		c.writeCountNormalized,
		prometheus.CounterValue,
		float64(containerStats.Storage.WriteCountNormalized),
		containerIdWithPrefix,
	)
	ch <- prometheus.MustNewConstMetric(
		c.writeSizeBytes,
		prometheus.CounterValue,
		float64(containerStats.Storage.WriteSizeBytes),
		containerIdWithPrefix,
	)

	return nil
}

// collectNetworkMetrics collects network metrics for containers.
// With HNSv2, the network stats must be collected from hcsshim.HNSListEndpointRequest.
// Network statistics from the container.Statistics() are providing data only, if HNSv1 is used.
// Ref: https://github.com/prometheus-community/windows_exporter/pull/1218
func (c *Collector) collectNetworkMetrics(ch chan<- prometheus.Metric, containerPrefixes map[string]string) error {
	hnsEndpoints, err := hcsshim.HNSListEndpointRequest()
	if err != nil {
		return fmt.Errorf("error in fetching HNS endpoints: %w", err)
	}

	if len(hnsEndpoints) == 0 {
		return errors.New("no network stats for containers to collect")
	}

	for _, endpoint := range hnsEndpoints {
		endpointStats, err := hcsshim.GetHNSEndpointStats(endpoint.Id)
		if err != nil {
			c.logger.Warn("Failed to collect network stats for interface "+endpoint.Id,
				slog.Any("err", err),
			)

			continue
		}

		for _, containerId := range endpoint.SharedContainers {
			containerIdWithPrefix, ok := containerPrefixes[containerId]

			if !ok {
				c.logger.Debug("Failed to collect network stats for container " + containerId)

				continue
			}

			endpointId := strings.ToUpper(endpoint.Id)

			ch <- prometheus.MustNewConstMetric(
				c.bytesReceived,
				prometheus.CounterValue,
				float64(endpointStats.BytesReceived),
				containerIdWithPrefix, endpointId,
			)

			ch <- prometheus.MustNewConstMetric(
				c.bytesSent,
				prometheus.CounterValue,
				float64(endpointStats.BytesSent),
				containerIdWithPrefix, endpointId,
			)
			ch <- prometheus.MustNewConstMetric(
				c.packetsReceived,
				prometheus.CounterValue,
				float64(endpointStats.PacketsReceived),
				containerIdWithPrefix, endpointId,
			)
			ch <- prometheus.MustNewConstMetric(
				c.packetsSent,
				prometheus.CounterValue,
				float64(endpointStats.PacketsSent),
				containerIdWithPrefix, endpointId,
			)
			ch <- prometheus.MustNewConstMetric(
				c.droppedPacketsIncoming,
				prometheus.CounterValue,
				float64(endpointStats.DroppedPacketsIncoming),
				containerIdWithPrefix, endpointId,
			)
			ch <- prometheus.MustNewConstMetric(
				c.droppedPacketsOutgoing,
				prometheus.CounterValue,
				float64(endpointStats.DroppedPacketsOutgoing),
				containerIdWithPrefix, endpointId,
			)
		}
	}

	return nil
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

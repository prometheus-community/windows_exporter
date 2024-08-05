//go:build windows

package container

import (
	"strings"

	"github.com/Microsoft/hcsshim"
	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "container"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for containers metrics
type Collector struct {
	logger log.Logger

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

// New constructs a new Collector
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
	return []string{}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build() error {
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
func (c *Collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.collect(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting collector metrics", "err", err)
		return err
	}
	return nil
}

// containerClose closes the container resource
func (c *Collector) containerClose(container hcsshim.Container) {
	err := container.Close()
	if err != nil {
		_ = level.Error(c.logger).Log("err", err)
	}
}

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	// Types Container is passed to get the containers compute systems only
	containers, err := hcsshim.GetContainers(hcsshim.ComputeSystemQuery{Types: []string{"Container"}})
	if err != nil {
		_ = level.Error(c.logger).Log("msg", "Err in Getting containers", "err", err)
		return err
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

	for _, containerDetails := range containers {
		// https://stackoverflow.com/questions/45617758/proper-way-to-release-resources-with-defer-in-a-loop
		func() {
			container, err := hcsshim.OpenContainer(containerDetails.ID)
			if container != nil {
				defer c.containerClose(container)
			}
			if err != nil {
				_ = level.Error(c.logger).Log("msg", "err in opening container", "containerId", containerDetails.ID, "err", err)
				return
			}

			cstats, err := container.Statistics()
			if err != nil {
				_ = level.Error(c.logger).Log("msg", "err in fetching container Statistics", "containerId", containerDetails.ID, "err", err)
				return
			}

			containerIdWithPrefix := getContainerIdWithPrefix(containerDetails)
			containerPrefixes[containerDetails.ID] = containerIdWithPrefix

			ch <- prometheus.MustNewConstMetric(
				c.containerAvailable,
				prometheus.CounterValue,
				1,
				containerIdWithPrefix,
			)
			ch <- prometheus.MustNewConstMetric(
				c.usageCommitBytes,
				prometheus.GaugeValue,
				float64(cstats.Memory.UsageCommitBytes),
				containerIdWithPrefix,
			)
			ch <- prometheus.MustNewConstMetric(
				c.usageCommitPeakBytes,
				prometheus.GaugeValue,
				float64(cstats.Memory.UsageCommitPeakBytes),
				containerIdWithPrefix,
			)
			ch <- prometheus.MustNewConstMetric(
				c.usagePrivateWorkingSetBytes,
				prometheus.GaugeValue,
				float64(cstats.Memory.UsagePrivateWorkingSetBytes),
				containerIdWithPrefix,
			)
			ch <- prometheus.MustNewConstMetric(
				c.runtimeTotal,
				prometheus.CounterValue,
				float64(cstats.Processor.TotalRuntime100ns)*perflib.TicksToSecondScaleFactor,
				containerIdWithPrefix,
			)
			ch <- prometheus.MustNewConstMetric(
				c.runtimeUser,
				prometheus.CounterValue,
				float64(cstats.Processor.RuntimeUser100ns)*perflib.TicksToSecondScaleFactor,
				containerIdWithPrefix,
			)
			ch <- prometheus.MustNewConstMetric(
				c.runtimeKernel,
				prometheus.CounterValue,
				float64(cstats.Processor.RuntimeKernel100ns)*perflib.TicksToSecondScaleFactor,
				containerIdWithPrefix,
			)
			ch <- prometheus.MustNewConstMetric(
				c.readCountNormalized,
				prometheus.CounterValue,
				float64(cstats.Storage.ReadCountNormalized),
				containerIdWithPrefix,
			)
			ch <- prometheus.MustNewConstMetric(
				c.readSizeBytes,
				prometheus.CounterValue,
				float64(cstats.Storage.ReadSizeBytes),
				containerIdWithPrefix,
			)
			ch <- prometheus.MustNewConstMetric(
				c.writeCountNormalized,
				prometheus.CounterValue,
				float64(cstats.Storage.WriteCountNormalized),
				containerIdWithPrefix,
			)
			ch <- prometheus.MustNewConstMetric(
				c.writeSizeBytes,
				prometheus.CounterValue,
				float64(cstats.Storage.WriteSizeBytes),
				containerIdWithPrefix,
			)
		}()
	}

	hnsEndpoints, err := hcsshim.HNSListEndpointRequest()
	if err != nil {
		_ = level.Warn(c.logger).Log("msg", "Failed to collect network stats for containers")
		return err
	}

	if len(hnsEndpoints) == 0 {
		_ = level.Info(c.logger).Log("msg", "No network stats for containers to collect")
		return nil
	}

	for _, endpoint := range hnsEndpoints {
		endpointStats, err := hcsshim.GetHNSEndpointStats(endpoint.Id)
		if err != nil {
			_ = level.Warn(c.logger).Log("msg", "Failed to collect network stats for interface "+endpoint.Id, "err", err)
			continue
		}

		for _, containerId := range endpoint.SharedContainers {
			containerIdWithPrefix, ok := containerPrefixes[containerId]
			endpointId := strings.ToUpper(endpoint.Id)

			if !ok {
				_ = level.Warn(c.logger).Log("msg", "Failed to collect network stats for container "+containerId)
				continue
			}

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

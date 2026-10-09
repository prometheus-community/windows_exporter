// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows

package container

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/headers/cri"
	"github.com/prometheus-community/windows_exporter/internal/headers/hcn"
	"github.com/prometheus-community/windows_exporter/internal/headers/hcs"
	"github.com/prometheus-community/windows_exporter/internal/headers/kernel32"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows"
)

const (
	Name = "container"

	subCollectorHCS         = "hcs"
	subCollectorHostprocess = "hostprocess"

	JobObjectMemoryUsageInformation = 28

	// defaultCRITimeout limits the CRI calls if the scrape has no timeout.
	defaultCRITimeout = 10 * time.Second

	// grpcUnimplemented is returned by containerd if its CRI plugin is disabled.
	grpcUnimplemented = 12
)

type Config struct {
	CollectorsEnabled []string `yaml:"enabled"`
	// CRIEndpoint is the Kubernetes Container Runtime Interface (CRI) endpoint.
	// It provides the Kubernetes metadata and the HostProcess containers.
	CRIEndpoint string `yaml:"cri-endpoint"`
}

//nolint:gochecknoglobals
var ConfigDefaults = Config{
	CollectorsEnabled: []string{
		subCollectorHCS,
		subCollectorHostprocess,
	},
	CRIEndpoint: "npipe:////./pipe/containerd-containerd",
}

// A Collector is a Prometheus Collector for containers metrics.
type Collector struct {
	config Config

	logger *slog.Logger

	criClient *cri.Client

	// mu serializes Collect, because the CRI state below is not safe for concurrent use.
	// Overlapping Collect calls, e.g. after a scrape timeout, would otherwise write it concurrently.
	mu sync.Mutex

	// criRuntimeName is the runtime name, e.g. containerd. Kubernetes uses it as container ID prefix.
	criRuntimeName string
	// criUnavailableLogged avoids logging an unavailable CRI endpoint on every scrape.
	criUnavailableLogged bool

	// Presence
	containerAvailable *prometheus.Desc

	// Number of containers
	containersCount *prometheus.Desc

	// Lifecycle
	startTime *prometheus.Desc
	processes *prometheus.Desc

	// Memory
	usageCommitBytes            *prometheus.Desc
	usageCommitPeakBytes        *prometheus.Desc
	usagePrivateWorkingSetBytes *prometheus.Desc
	pageFaults                  *prometheus.Desc

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
	readCountNormalized     *prometheus.Desc
	readSizeBytes           *prometheus.Desc
	writeCountNormalized    *prometheus.Desc
	writeSizeBytes          *prometheus.Desc
	writableLayerUsageBytes *prometheus.Desc
}

type containerInfo struct {
	id        string
	namespace string
	pod       string
	container string
	// createdAt is the creation time reported by the CRI endpoint. It is zero for containers not managed by Kubernetes.
	createdAt time.Time
}

// kubernetesContainers holds the Kubernetes metadata of the running containers, read from the CRI endpoint.
type kubernetesContainers struct {
	// containers maps the container IDs without runtime prefix to their metadata.
	containers map[string]containerInfo
	// sandboxes holds the IDs of all pod sandboxes. Their pause containers are not exported.
	sandboxes map[string]struct{}
}

func newKubernetesContainers(runtimeName string, sandboxes []cri.PodSandbox, containers []cri.Container) kubernetesContainers {
	k := kubernetesContainers{
		containers: make(map[string]containerInfo, len(containers)),
		sandboxes:  make(map[string]struct{}, len(sandboxes)),
	}

	pods := make(map[string]cri.PodSandbox, len(sandboxes))

	for _, sandbox := range sandboxes {
		k.sandboxes[sandbox.ID] = struct{}{}
		pods[sandbox.ID] = sandbox
	}

	for _, container := range containers {
		if container.State != cri.ContainerRunning {
			continue
		}

		pod := pods[container.PodSandboxID]

		k.containers[container.ID] = containerInfo{
			id:        runtimeName + "://" + container.ID,
			namespace: pod.Namespace,
			pod:       pod.Name,
			container: container.Name,
			createdAt: container.CreatedAt,
		}
	}

	return k
}

// New constructs a new Collector.
func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	if config.CollectorsEnabled == nil {
		config.CollectorsEnabled = ConfigDefaults.CollectorsEnabled
	}

	if config.CRIEndpoint == "" {
		config.CRIEndpoint = ConfigDefaults.CRIEndpoint
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(app *kingpin.Application) *Collector {
	c := &Collector{
		config: ConfigDefaults,
	}
	c.config.CollectorsEnabled = make([]string, 0)

	var collectorsEnabled string

	app.Flag(
		"collector.container.enabled",
		"Comma-separated list of collectors to use. Defaults to all, if not specified.",
	).Default(strings.Join(ConfigDefaults.CollectorsEnabled, ",")).StringVar(&collectorsEnabled)

	app.Flag(
		"collector.container.cri-endpoint",
		"Kubernetes CRI endpoint, used for Kubernetes labels and HostProcess containers.",
	).Default(ConfigDefaults.CRIEndpoint).StringVar(&c.config.CRIEndpoint)

	app.Action(func(*kingpin.ParseContext) error {
		c.config.CollectorsEnabled = strings.Split(collectorsEnabled, ",")

		return nil
	})

	return c
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	if c.criClient != nil {
		c.criClient.Close()
		c.criClient = nil
	}

	return nil
}

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	c.logger = logger.With(slog.String("collector", Name))

	for _, collector := range c.config.CollectorsEnabled {
		if !slices.Contains([]string{subCollectorHCS, subCollectorHostprocess}, collector) {
			return fmt.Errorf("unknown collector: %s", collector)
		}
	}

	// Without the Containers feature, the Host Compute Service is missing and
	// every scrape would fail. Report the collector as unsupported instead, so
	// it is skipped once at startup rather than logging a warning per scrape.
	if slices.Contains(c.config.CollectorsEnabled, subCollectorHCS) {
		if _, err := hcs.GetContainers(); errors.Is(err, hcs.ErrServiceNotAvailable) {
			return fmt.Errorf("host compute service not available, is the Containers feature installed? %w: %w",
				errors.ErrUnsupported, err)
		}
	}

	_ = c.Close()

	criClient, err := cri.NewClient(c.config.CRIEndpoint)
	if err != nil {
		return fmt.Errorf("invalid CRI endpoint: %w", err)
	}

	c.criClient = criClient
	c.criRuntimeName = ""
	c.criUnavailableLogged = false

	c.containerAvailable = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "available"),
		"Available",
		[]string{"container_id", "namespace", "pod", "container", "hostprocess"},
		nil,
	)
	c.containersCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "count"),
		"Number of containers",
		nil,
		nil,
	)
	c.startTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "start_time_seconds"),
		"Start time of the container since Unix epoch in seconds. HostProcess containers report their creation time.",
		[]string{"container_id", "namespace", "pod", "container"},
		nil,
	)
	c.processes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "processes"),
		"Number of processes running in the container",
		[]string{"container_id", "namespace", "pod", "container"},
		nil,
	)
	c.usageCommitBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memory_usage_commit_bytes"),
		"Memory Usage Commit Bytes",
		[]string{"container_id", "namespace", "pod", "container"},
		nil,
	)
	c.usageCommitPeakBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memory_usage_commit_peak_bytes"),
		"Memory Usage Commit Peak Bytes",
		[]string{"container_id", "namespace", "pod", "container"},
		nil,
	)
	c.usagePrivateWorkingSetBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memory_usage_private_working_set_bytes"),
		"Memory Usage Private Working Set Bytes",
		[]string{"container_id", "namespace", "pod", "container"},
		nil,
	)
	c.pageFaults = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memory_page_faults_total"),
		"Total number of page faults of the container processes. Only available for HostProcess containers.",
		[]string{"container_id", "namespace", "pod", "container"},
		nil,
	)
	c.runtimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cpu_usage_seconds_total"),
		"Total Run time in Seconds",
		[]string{"container_id", "namespace", "pod", "container"},
		nil,
	)
	c.runtimeUser = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cpu_usage_seconds_usermode"),
		"Run Time in User mode in Seconds",
		[]string{"container_id", "namespace", "pod", "container"},
		nil,
	)
	c.runtimeKernel = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cpu_usage_seconds_kernelmode"),
		"Run time in Kernel mode in Seconds",
		[]string{"container_id", "namespace", "pod", "container"},
		nil,
	)
	c.bytesReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "network_receive_bytes_total"),
		"Bytes Received on Interface",
		[]string{"container_id", "namespace", "pod", "container", "interface"},
		nil,
	)
	c.bytesSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "network_transmit_bytes_total"),
		"Bytes Sent on Interface",
		[]string{"container_id", "namespace", "pod", "container", "interface"},
		nil,
	)
	c.packetsReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "network_receive_packets_total"),
		"Packets Received on Interface",
		[]string{"container_id", "namespace", "pod", "container", "interface"},
		nil,
	)
	c.packetsSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "network_transmit_packets_total"),
		"Packets Sent on Interface",
		[]string{"container_id", "namespace", "pod", "container", "interface"},
		nil,
	)
	c.droppedPacketsIncoming = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "network_receive_packets_dropped_total"),
		"Dropped Incoming Packets on Interface",
		[]string{"container_id", "namespace", "pod", "container", "interface"},
		nil,
	)
	c.droppedPacketsOutgoing = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "network_transmit_packets_dropped_total"),
		"Dropped Outgoing Packets on Interface",
		[]string{"container_id", "namespace", "pod", "container", "interface"},
		nil,
	)
	c.readCountNormalized = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "storage_read_count_normalized_total"),
		"Read Count Normalized",
		[]string{"container_id", "namespace", "pod", "container"},
		nil,
	)
	c.readSizeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "storage_read_size_bytes_total"),
		"Read Size Bytes",
		[]string{"container_id", "namespace", "pod", "container"},
		nil,
	)
	c.writeCountNormalized = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "storage_write_count_normalized_total"),
		"Write Count Normalized",
		[]string{"container_id", "namespace", "pod", "container"},
		nil,
	)
	c.writeSizeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "storage_write_size_bytes_total"),
		"Write Size Bytes",
		[]string{"container_id", "namespace", "pod", "container"},
		nil,
	)
	c.writableLayerUsageBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "storage_writable_layer_usage_bytes"),
		"Used bytes of the writable layer, as reported by the CRI endpoint",
		[]string{"container_id", "namespace", "pod", "container"},
		nil,
	)

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric, maxScrapeDuration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if maxScrapeDuration <= 0 {
		maxScrapeDuration = defaultCRITimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), maxScrapeDuration)
	defer cancel()

	errs := make([]error, 0)

	// Without Kubernetes metadata, HCS containers are still exported, but without Kubernetes labels.
	kubernetes, err := c.getKubernetesContainers(ctx)
	if err != nil {
		errs = append(errs, err)
	}

	// hcsContainers and jobContainers stay nil if their collector is disabled.
	var (
		hcsContainers map[string]struct{}
		jobContainers map[string]struct{}
	)

	if slices.Contains(c.config.CollectorsEnabled, subCollectorHCS) {
		hcsContainers, err = c.collectHCS(ch, kubernetes)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, subCollectorHostprocess) {
		jobContainers, err = c.collectJobContainers(ch, kubernetes)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if err := c.collectCRIStats(ctx, ch, kubernetes, hcsContainers, jobContainers); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// getKubernetesContainers reads the running containers and pod sandboxes from the CRI endpoint.
// If the endpoint is not available, e.g. on hosts without Kubernetes, it returns no containers.
func (c *Collector) getKubernetesContainers(ctx context.Context) (kubernetesContainers, error) {
	if c.criRuntimeName == "" {
		version, err := c.criClient.Version(ctx)
		if err != nil {
			return kubernetesContainers{}, c.handleCRIError(err)
		}

		c.criRuntimeName = version.RuntimeName
		if c.criRuntimeName == "" {
			c.criRuntimeName = "containerd"
		}
	}

	sandboxes, err := c.criClient.ListPodSandboxes(ctx)
	if err != nil {
		return kubernetesContainers{}, c.handleCRIError(err)
	}

	containers, err := c.criClient.ListRunningContainers(ctx)
	if err != nil {
		return kubernetesContainers{}, c.handleCRIError(err)
	}

	if c.criUnavailableLogged {
		c.logger.InfoContext(ctx, "CRI endpoint is available", slog.String("cri_endpoint", c.config.CRIEndpoint))

		c.criUnavailableLogged = false
	}

	return newKubernetesContainers(c.criRuntimeName, sandboxes, containers), nil
}

// handleCRIError ignores errors of a missing CRI endpoint, which is expected on hosts without Kubernetes.
func (c *Collector) handleCRIError(err error) error {
	// The runtime may be replaced, so query its name again.
	c.criRuntimeName = ""

	var statusErr *cri.StatusError

	if !errors.Is(err, windows.ERROR_FILE_NOT_FOUND) && (!errors.As(err, &statusErr) || statusErr.Code != grpcUnimplemented) {
		return fmt.Errorf("error in fetching Kubernetes metadata from CRI endpoint %s: %w", c.config.CRIEndpoint, err)
	}

	if !c.criUnavailableLogged {
		c.logger.Info("CRI endpoint is not available, Kubernetes labels and HostProcess containers are not collected",
			slog.String("cri_endpoint", c.config.CRIEndpoint),
			slog.Any("err", err),
		)

		c.criUnavailableLogged = true
	}

	return nil
}

// collectHCS collects the metrics of the containers managed by HCS.
// It returns the IDs of all HCS containers in any state, or nil if HCS could not be queried.
func (c *Collector) collectHCS(ch chan<- prometheus.Metric, kubernetes kubernetesContainers) (map[string]struct{}, error) {
	// Types Container is passed to get the containers compute systems only
	containers, err := hcs.GetContainers()
	if err != nil {
		return nil, fmt.Errorf("error in fetching containers: %w", err)
	}

	count := len(containers)
	if count == 0 {
		ch <- prometheus.MustNewConstMetric(
			c.containersCount,
			prometheus.GaugeValue,
			0,
		)

		return map[string]struct{}{}, nil
	}

	var countersCount float64

	hcsIDs := make(map[string]struct{}, len(containers))
	hcsContainers := make(map[string]containerInfo, len(containers))
	collectErrors := make([]error, 0)

	for _, container := range containers {
		// Stopping or paused containers are still HCS containers, not Hyper-V isolated ones.
		hcsIDs[container.ID] = struct{}{}

		if container.State != "Running" {
			continue
		}

		countersCount++

		// Skip pause containers
		if _, ok := kubernetes.sandboxes[container.ID]; ok {
			continue
		}

		info, ok := kubernetes.containers[container.ID]
		if !ok {
			// Containers not managed by Kubernetes, e.g. Docker containers, have no Kubernetes labels.
			info = containerInfo{id: getContainerIdWithPrefix(container)}
		}

		hcsContainers[container.ID] = info

		if err = c.collectHCSContainer(ch, container, info); err != nil {
			if errors.Is(err, hcs.ErrIDNotFound) {
				c.logger.Debug("err in fetching container statistics",
					slog.String("container_id", container.ID),
					slog.String("container_name", info.container),
					slog.String("container_pod_name", info.pod),
					slog.String("container_namespace", info.namespace),
					slog.Any("err", err),
				)
			} else {
				c.logger.Error("err in fetching container statistics",
					slog.String("container_id", container.ID),
					slog.String("container_name", info.container),
					slog.String("container_pod_name", info.pod),
					slog.String("container_namespace", info.namespace),
					slog.Any("err", err),
				)

				collectErrors = append(collectErrors, err)
			}

			continue
		}
	}

	ch <- prometheus.MustNewConstMetric(
		c.containersCount,
		prometheus.GaugeValue,
		countersCount,
	)

	if err := c.collectNetworkMetrics(ch, hcsContainers, kubernetes.sandboxes); err != nil {
		return hcsIDs, fmt.Errorf("error in fetching container network statistics: %w", err)
	}

	if len(collectErrors) > 0 {
		return hcsIDs, fmt.Errorf("errors while fetching container statistics: %w", errors.Join(collectErrors...))
	}

	return hcsIDs, nil
}

func (c *Collector) collectHCSContainer(ch chan<- prometheus.Metric, containerDetails hcs.Properties, containerInfo containerInfo) error {
	properties, err := hcs.GetContainerStatistics(containerDetails.ID)
	if err != nil {
		return fmt.Errorf("error fetching container statistics: %w", err)
	}

	containerStats := properties.Statistics

	ch <- prometheus.MustNewConstMetric(
		c.containerAvailable,
		prometheus.GaugeValue,
		1,
		containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container, "false",
	)

	startTime := containerStats.ContainerStartTime
	if startTime.IsZero() {
		startTime = containerInfo.createdAt
	}

	if !startTime.IsZero() {
		ch <- prometheus.MustNewConstMetric(
			c.startTime,
			prometheus.GaugeValue,
			float64(startTime.UnixNano())/1e9,

			containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
		)
	}

	ch <- prometheus.MustNewConstMetric(
		c.processes,
		prometheus.GaugeValue,
		float64(len(properties.ProcessList)),

		containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
	)

	ch <- prometheus.MustNewConstMetric(
		c.usageCommitBytes,
		prometheus.GaugeValue,
		float64(containerStats.Memory.MemoryUsageCommitBytes),

		containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
	)

	ch <- prometheus.MustNewConstMetric(
		c.usageCommitPeakBytes,
		prometheus.GaugeValue,
		float64(containerStats.Memory.MemoryUsageCommitPeakBytes),

		containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
	)

	ch <- prometheus.MustNewConstMetric(
		c.usagePrivateWorkingSetBytes,
		prometheus.GaugeValue,
		float64(containerStats.Memory.MemoryUsagePrivateWorkingSetBytes),

		containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
	)

	ch <- prometheus.MustNewConstMetric(
		c.runtimeTotal,
		prometheus.CounterValue,
		float64(containerStats.Processor.TotalRuntime100ns)*pdh.TicksToSecondScaleFactor,

		containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
	)

	ch <- prometheus.MustNewConstMetric(
		c.runtimeUser,
		prometheus.CounterValue,
		float64(containerStats.Processor.RuntimeUser100ns)*pdh.TicksToSecondScaleFactor,

		containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
	)

	ch <- prometheus.MustNewConstMetric(
		c.runtimeKernel,
		prometheus.CounterValue,
		float64(containerStats.Processor.RuntimeKernel100ns)*pdh.TicksToSecondScaleFactor,

		containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
	)

	ch <- prometheus.MustNewConstMetric(
		c.readCountNormalized,
		prometheus.CounterValue,
		float64(containerStats.Storage.ReadCountNormalized),

		containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
	)

	ch <- prometheus.MustNewConstMetric(
		c.readSizeBytes,
		prometheus.CounterValue,
		float64(containerStats.Storage.ReadSizeBytes),

		containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
	)

	ch <- prometheus.MustNewConstMetric(
		c.writeCountNormalized,
		prometheus.CounterValue,
		float64(containerStats.Storage.WriteCountNormalized),

		containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
	)

	ch <- prometheus.MustNewConstMetric(
		c.writeSizeBytes,
		prometheus.CounterValue,
		float64(containerStats.Storage.WriteSizeBytes),

		containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
	)

	return nil
}

// collectNetworkMetrics collects network metrics for the HCS containers.
func (c *Collector) collectNetworkMetrics(ch chan<- prometheus.Metric, containers map[string]containerInfo, sandboxes map[string]struct{}) error {
	endpoints, err := hcn.ListEndpoints()
	if err != nil {
		return fmt.Errorf("error in fetching HCN endpoints: %w", err)
	}

	if len(endpoints) == 0 {
		return nil
	}

	for _, endpoint := range endpoints {
		if len(endpoint.SharedContainers) == 0 {
			continue
		}

		endpointStats, err := hcn.GetHNSEndpointStats(endpoint.ID)
		if err != nil {
			c.logger.Warn("Failed to collect network stats for interface "+endpoint.ID,
				slog.Any("err", err),
			)

			continue
		}

		for _, containerId := range endpoint.SharedContainers {
			containerInfo, ok := containers[containerId]
			if !ok {
				// Pause containers share the endpoint of their pod.
				if _, ok := sandboxes[containerId]; !ok {
					c.logger.Debug("Unknown container " + containerId + " for endpoint " + endpoint.ID)
				}

				continue
			}

			endpointId := strings.ToUpper(endpoint.ID)

			ch <- prometheus.MustNewConstMetric(
				c.bytesReceived,
				prometheus.CounterValue,
				float64(endpointStats.BytesReceived),
				containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container, endpointId,
			)

			ch <- prometheus.MustNewConstMetric(
				c.bytesSent,
				prometheus.CounterValue,
				float64(endpointStats.BytesSent),
				containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container, endpointId,
			)

			ch <- prometheus.MustNewConstMetric(
				c.packetsReceived,
				prometheus.CounterValue,
				float64(endpointStats.PacketsReceived),
				containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container, endpointId,
			)

			ch <- prometheus.MustNewConstMetric(
				c.packetsSent,
				prometheus.CounterValue,
				float64(endpointStats.PacketsSent),
				containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container, endpointId,
			)

			ch <- prometheus.MustNewConstMetric(
				c.droppedPacketsIncoming,
				prometheus.CounterValue,
				float64(endpointStats.DroppedPacketsIncoming),
				containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container, endpointId,
			)

			ch <- prometheus.MustNewConstMetric(
				c.droppedPacketsOutgoing,
				prometheus.CounterValue,
				float64(endpointStats.DroppedPacketsOutgoing),
				containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container, endpointId,
			)
		}
	}

	return nil
}

// collectJobContainers collects container metrics for job containers.
// Job container based on Win32 Job objects.
// https://learn.microsoft.com/en-us/windows/win32/procthread/job-objects
//
// Job containers are containers that aren't managed by HCS, e.g host process containers.
// They are found by the job object, which hcsshim names after the container ID.
// It returns the IDs of the job containers.
func (c *Collector) collectJobContainers(ch chan<- prometheus.Metric, kubernetes kubernetesContainers) (map[string]struct{}, error) {
	jobContainers := make(map[string]struct{})
	errs := make([]error, 0)

	for containerID, info := range kubernetes.containers {
		found, err := c.collectJobContainer(ch, containerID, info)
		if found {
			jobContainers[containerID] = struct{}{}
		}

		if err != nil {
			errs = append(errs, err)
		}
	}

	return jobContainers, errors.Join(errs...)
}

// collectJobContainer collects the metrics of a job container.
// It does nothing and returns false if the container has no job object, i.e. it is not a job container.
func (c *Collector) collectJobContainer(ch chan<- prometheus.Metric, containerID string, containerInfo containerInfo) (bool, error) {
	jobObjectHandle, err := openJobObject(containerID)
	if err != nil {
		if errors.Is(err, windows.ERROR_FILE_NOT_FOUND) {
			return false, nil
		}

		return true, fmt.Errorf("error in opening job object: %w", err)
	}

	defer func(fd windows.Handle) {
		_ = windows.Close(fd)
	}(jobObjectHandle)

	var jobInfo kernel32.JobObjectBasicAndIOAccountingInformation

	if err = windows.QueryInformationJobObject(
		jobObjectHandle,
		windows.JobObjectBasicAndIoAccountingInformation,
		uintptr(unsafe.Pointer(&jobInfo)),
		uint32(unsafe.Sizeof(jobInfo)),
		nil,
	); err != nil {
		return true, fmt.Errorf("error in querying job object information: %w", err)
	}

	var jobMemoryInfo kernel32.JobObjectMemoryUsageInformation

	// https://github.com/microsoft/hcsshim/blob/bfb2a106798d3765666f6e39ec6cf0117275eab4/internal/jobobject/jobobject.go#L410
	if err = windows.QueryInformationJobObject(
		jobObjectHandle,
		JobObjectMemoryUsageInformation,
		uintptr(unsafe.Pointer(&jobMemoryInfo)),
		uint32(unsafe.Sizeof(jobMemoryInfo)),
		nil,
	); err != nil {
		return true, fmt.Errorf("error in querying job object memory usage information: %w", err)
	}

	ch <- prometheus.MustNewConstMetric(
		c.containerAvailable,
		prometheus.GaugeValue,
		1,
		containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container, "true",
	)

	// Job objects have no start time.
	c.collectCreationTime(ch, containerInfo)

	ch <- prometheus.MustNewConstMetric(
		c.processes,
		prometheus.GaugeValue,
		float64(jobInfo.BasicInfo.ActiveProcesses),

		containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
	)

	ch <- prometheus.MustNewConstMetric(
		c.pageFaults,
		prometheus.CounterValue,
		float64(jobInfo.BasicInfo.TotalPageFaultCount),

		containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
	)

	ch <- prometheus.MustNewConstMetric(
		c.usageCommitBytes,
		prometheus.GaugeValue,
		float64(jobMemoryInfo.JobMemory),

		containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
	)

	ch <- prometheus.MustNewConstMetric(
		c.usageCommitPeakBytes,
		prometheus.GaugeValue,
		float64(jobMemoryInfo.PeakJobMemoryUsed),

		containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
	)

	if privateWorkingSetBytes, err := calculatePrivateWorkingSetBytes(jobObjectHandle); err != nil {
		c.logger.Debug("error in calculating private working set bytes",
			slog.String("container_id", containerID),
			slog.Any("err", err),
		)
	} else {
		ch <- prometheus.MustNewConstMetric(
			c.usagePrivateWorkingSetBytes,
			prometheus.GaugeValue,
			float64(privateWorkingSetBytes),

			containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
		)
	}

	// The ThisPeriod* times reset when a job time limit is set, so they are not monotonic.
	ch <- prometheus.MustNewConstMetric(
		c.runtimeTotal,
		prometheus.CounterValue,
		(float64(jobInfo.BasicInfo.TotalKernelTime)+float64(jobInfo.BasicInfo.TotalUserTime))*pdh.TicksToSecondScaleFactor,

		containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
	)

	ch <- prometheus.MustNewConstMetric(
		c.runtimeUser,
		prometheus.CounterValue,
		float64(jobInfo.BasicInfo.TotalUserTime)*pdh.TicksToSecondScaleFactor,

		containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
	)

	ch <- prometheus.MustNewConstMetric(
		c.runtimeKernel,
		prometheus.CounterValue,
		float64(jobInfo.BasicInfo.TotalKernelTime)*pdh.TicksToSecondScaleFactor,

		containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
	)

	ch <- prometheus.MustNewConstMetric(
		c.readCountNormalized,
		prometheus.CounterValue,
		float64(jobInfo.IoInfo.ReadOperationCount),

		containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
	)

	ch <- prometheus.MustNewConstMetric(
		c.readSizeBytes,
		prometheus.CounterValue,
		float64(jobInfo.IoInfo.ReadTransferCount),

		containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
	)

	ch <- prometheus.MustNewConstMetric(
		c.writeCountNormalized,
		prometheus.CounterValue,
		float64(jobInfo.IoInfo.WriteOperationCount),

		containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
	)

	ch <- prometheus.MustNewConstMetric(
		c.writeSizeBytes,
		prometheus.CounterValue,
		float64(jobInfo.IoInfo.WriteTransferCount),

		containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
	)

	return true, nil
}

// collectCreationTime collects the creation time reported by the CRI endpoint as start time,
// for containers whose start time HCS doesn't report.
func (c *Collector) collectCreationTime(ch chan<- prometheus.Metric, containerInfo containerInfo) {
	if containerInfo.createdAt.IsZero() {
		return
	}

	ch <- prometheus.MustNewConstMetric(
		c.startTime,
		prometheus.GaugeValue,
		float64(containerInfo.createdAt.UnixNano())/1e9,

		containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
	)
}

// openJobObject opens the job object of a job container, which hcsshim names after the container ID.
func openJobObject(containerID string) (windows.Handle, error) {
	return kernel32.OpenJobObject("Global\\JobContainer_" + containerID)
}

// isJobContainer reports whether a container has a job object.
// Errors other than a missing job object are treated as job container.
func isJobContainer(containerID string) bool {
	jobObjectHandle, err := openJobObject(containerID)
	if err != nil {
		return !errors.Is(err, windows.ERROR_FILE_NOT_FOUND)
	}

	_ = windows.Close(jobObjectHandle)

	return true
}

// collectCRIStats collects the container stats of the CRI endpoint for the running Kubernetes containers:
//   - the writable layer usage of all containers exported by the enabled collectors.
//   - CPU and memory usage of Hyper-V isolated containers. They run inside a utility VM,
//     so they are neither visible to HCS on the host nor job containers.
//     They belong to the hcs collector, which provides the list of HCS containers to exclude.
//
// The CRI stats lack the user and kernel mode CPU time, the commit peak, storage I/O and network statistics.
func (c *Collector) collectCRIStats(
	ctx context.Context,
	ch chan<- prometheus.Metric,
	kubernetes kubernetesContainers,
	hcsContainers map[string]struct{},
	jobContainers map[string]struct{},
) error {
	exported, hypervContainers := selectCRIStatsContainers(kubernetes, hcsContainers, jobContainers)
	if len(exported) == 0 {
		return nil
	}

	stats, err := c.criClient.ListContainerStats(ctx)
	if err != nil {
		err = fmt.Errorf("error in fetching container stats from CRI endpoint %s: %w", c.config.CRIEndpoint, err)

		// Without Hyper-V isolated containers, only the writable layer usage is missing.
		if len(hypervContainers) == 0 {
			c.logger.WarnContext(ctx, "writable layer usage is not collected", slog.Any("err", err))

			return nil
		}

		return err
	}

	c.collectCRIContainers(ch, stats, kubernetes, exported, hypervContainers)

	return nil
}

// selectCRIStatsContainers returns the running Kubernetes containers exported by the enabled collectors,
// and among them the Hyper-V isolated containers, which are neither HCS nor job containers.
// hcsContainers and jobContainers are nil if their collector is disabled.
func selectCRIStatsContainers(
	kubernetes kubernetesContainers,
	hcsContainers map[string]struct{},
	jobContainers map[string]struct{},
) (map[string]struct{}, map[string]struct{}) {
	exported := make(map[string]struct{}, len(kubernetes.containers))
	hypervContainers := make(map[string]struct{})

	for containerID := range kubernetes.containers {
		_, isHCS := hcsContainers[containerID]
		_, isJob := jobContainers[containerID]

		switch {
		case isHCS || isJob:
			exported[containerID] = struct{}{}
		case hcsContainers == nil:
			// Without the HCS containers, Hyper-V isolated containers are unknown.
		case jobContainers == nil && isJobContainer(containerID):
			// A job container while the hostprocess collector is disabled.
		default:
			exported[containerID] = struct{}{}
			hypervContainers[containerID] = struct{}{}
		}
	}

	return exported, hypervContainers
}

// collectCRIContainers collects the CRI stats of the exported containers.
func (c *Collector) collectCRIContainers(
	ch chan<- prometheus.Metric,
	stats []cri.ContainerStats,
	kubernetes kubernetesContainers,
	exported map[string]struct{},
	hypervContainers map[string]struct{},
) {
	for _, stat := range stats {
		if _, ok := exported[stat.ID]; !ok {
			continue
		}

		containerInfo := kubernetes.containers[stat.ID]

		// Containers without task metrics aren't running, e.g. they exited after they were listed.
		if _, ok := hypervContainers[stat.ID]; ok && (stat.CPU != nil || stat.Memory != nil) {
			c.collectCRIContainer(ch, stat, containerInfo)
		}

		// containerd reports 0 bytes without timestamp until it has measured the writable layer.
		if stat.WritableLayer != nil && stat.WritableLayer.UsedBytes != nil && stat.WritableLayer.Timestamp != 0 {
			ch <- prometheus.MustNewConstMetric(
				c.writableLayerUsageBytes,
				prometheus.GaugeValue,
				float64(*stat.WritableLayer.UsedBytes),

				containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
			)
		}
	}
}

// collectCRIContainer collects the CPU and memory usage of a container from its CRI stats.
// containerd reports the private working set as working set and the commit size as usage.
func (c *Collector) collectCRIContainer(ch chan<- prometheus.Metric, stat cri.ContainerStats, containerInfo containerInfo) {
	ch <- prometheus.MustNewConstMetric(
		c.containerAvailable,
		prometheus.GaugeValue,
		1,
		containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container, "false",
	)

	// HCS on the host doesn't report the start time.
	c.collectCreationTime(ch, containerInfo)

	if stat.CPU != nil && stat.CPU.UsageCoreNanoSeconds != nil {
		ch <- prometheus.MustNewConstMetric(
			c.runtimeTotal,
			prometheus.CounterValue,
			float64(*stat.CPU.UsageCoreNanoSeconds)/1e9,

			containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
		)
	}

	if stat.Memory == nil {
		return
	}

	if stat.Memory.WorkingSetBytes != nil {
		ch <- prometheus.MustNewConstMetric(
			c.usagePrivateWorkingSetBytes,
			prometheus.GaugeValue,
			float64(*stat.Memory.WorkingSetBytes),

			containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
		)
	}

	if stat.Memory.UsageBytes != nil {
		ch <- prometheus.MustNewConstMetric(
			c.usageCommitBytes,
			prometheus.GaugeValue,
			float64(*stat.Memory.UsageBytes),

			containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
		)
	}
}

func getContainerIdWithPrefix(container hcs.Properties) string {
	switch container.Owner {
	case "containerd-shim-runhcs-v1.exe":
		return "containerd://" + container.ID
	default:
		// default to docker or if owner is not set
		return "docker://" + container.ID
	}
}

func calculatePrivateWorkingSetBytes(jobObjectHandle windows.Handle) (uint64, error) {
	pids, err := kernel32.QueryJobObjectProcessIDs(jobObjectHandle)
	if err != nil {
		return 0, fmt.Errorf("error in querying job object process list: %w", err)
	}

	var privateWorkingSetBytes uint64

	getMemoryStats := func(pid uint32) (uint64, error) {
		processHandle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
		if err != nil {
			// The process exited after the process list was queried.
			if errors.Is(err, windows.ERROR_INVALID_PARAMETER) {
				return 0, nil
			}

			return 0, fmt.Errorf("error in opening process: %w", err)
		}

		defer func(fd windows.Handle) {
			_ = windows.Close(fd)
		}(processHandle)

		// Guard against PID reuse between querying the list and opening the process.
		isInJob, err := kernel32.IsProcessInJob(processHandle, jobObjectHandle)
		if err != nil {
			return 0, fmt.Errorf("error in checking if process is in job: %w", err)
		}

		if !isInJob {
			return 0, nil
		}

		var vmCounters kernel32.PROCESS_VM_COUNTERS

		retLen := uint32(unsafe.Sizeof(vmCounters))

		if err := windows.NtQueryInformationProcess(
			processHandle,
			windows.ProcessVmCounters,
			unsafe.Pointer(&vmCounters),
			retLen,
			&retLen,
		); err != nil {
			return 0, fmt.Errorf("error in querying process information: %w", err)
		}

		return uint64(vmCounters.PrivateWorkingSetSize), nil
	}

	for _, pid := range pids {
		privateWorkingSetSize, err := getMemoryStats(pid)
		if err != nil {
			return 0, fmt.Errorf("error in getting private working set bytes: %w", err)
		}

		privateWorkingSetBytes += privateWorkingSetSize
	}

	return privateWorkingSetBytes, nil
}

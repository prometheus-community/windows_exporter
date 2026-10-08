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
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/alecthomas/kingpin/v2"
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
)

type Config struct {
	CollectorsEnabled  []string `yaml:"enabled"`
	ContainerDStateDir string   `yaml:"containerd-state-dir"`
}

//nolint:gochecknoglobals
var ConfigDefaults = Config{
	CollectorsEnabled: []string{
		subCollectorHCS,
		subCollectorHostprocess,
	},
	ContainerDStateDir: `C:\ProgramData\containerd\state\io.containerd.runtime.v2.task\k8s.io\`,
}

// A Collector is a Prometheus Collector for containers metrics.
type Collector struct {
	config Config

	logger *slog.Logger

	// mu serializes Collect, because the annotation caches are not safe for concurrent use.
	// Overlapping Collect calls, e.g. after a scrape timeout, would otherwise write them concurrently.
	mu sync.Mutex

	// The annotation caches are keyed by the container ID without the runtime prefix.
	annotationsCacheHCS map[string]containerInfo
	// annotationsCacheJob holds every containerd bundle, so that each config.json is parsed only once.
	annotationsCacheJob map[string]jobContainer

	// stateDirMissingLogged avoids logging a missing containerd state directory on every scrape.
	stateDirMissingLogged bool

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

type containerInfo struct {
	id        string
	namespace string
	pod       string
	container string
}

type jobContainer struct {
	info        containerInfo
	hostProcess bool
}

type ociSpec struct {
	Annotations map[string]string `json:"annotations"`
}

// containerInfo returns the Kubernetes metadata of the container.
func (s ociSpec) containerInfo(id string) containerInfo {
	return containerInfo{
		id:        id,
		namespace: s.Annotations["io.kubernetes.cri.sandbox-namespace"],
		pod:       s.Annotations["io.kubernetes.cri.sandbox-name"],
		container: s.Annotations["io.kubernetes.cri.container-name"],
	}
}

// New constructs a new Collector.
func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	if config.CollectorsEnabled == nil {
		config.CollectorsEnabled = ConfigDefaults.CollectorsEnabled
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
		"collector.container.containerd-state-dir",
		"Path to the containerd state directory. Defaults to C:\\ProgramData\\containerd\\state\\io.containerd.runtime.v2.task\\k8s.io\\",
	).Default(ConfigDefaults.ContainerDStateDir).StringVar(&c.config.ContainerDStateDir)

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
	return nil
}

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	c.logger = logger.With(slog.String("collector", Name))

	for _, collector := range c.config.CollectorsEnabled {
		if !slices.Contains([]string{subCollectorHCS, subCollectorHostprocess}, collector) {
			return fmt.Errorf("unknown collector: %s", collector)
		}
	}

	c.annotationsCacheHCS = make(map[string]containerInfo)
	c.annotationsCacheJob = make(map[string]jobContainer)

	// Without the Containers feature, the Host Compute Service is missing and
	// every scrape would fail. Report the collector as unsupported instead, so
	// it is skipped once at startup rather than logging a warning per scrape.
	if slices.Contains(c.config.CollectorsEnabled, subCollectorHCS) {
		if _, err := hcs.GetContainers(); errors.Is(err, hcs.ErrServiceNotAvailable) {
			return fmt.Errorf("host compute service not available, is the Containers feature installed? %w: %w",
				errors.ErrUnsupported, err)
		}
	}

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

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric, _ time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	errs := make([]error, 0)

	if slices.Contains(c.config.CollectorsEnabled, subCollectorHCS) {
		if err := c.collectHCS(ch); err != nil {
			errs = append(errs, err)
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, subCollectorHostprocess) {
		if err := c.collectJobContainers(ch); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (c *Collector) collectHCS(ch chan<- prometheus.Metric) error {
	// Types Container is passed to get the containers compute systems only
	containers, err := hcs.GetContainers()
	if err != nil {
		return fmt.Errorf("error in fetching containers: %w", err)
	}

	count := len(containers)
	if count == 0 {
		clear(c.annotationsCacheHCS)

		ch <- prometheus.MustNewConstMetric(
			c.containersCount,
			prometheus.GaugeValue,
			0,
		)

		return nil
	}

	var countersCount float64

	containerIDs := make(map[string]struct{}, len(containers))
	collectErrors := make([]error, 0)

	for _, container := range containers {
		if container.State != "Running" {
			continue
		}

		containerIDs[container.ID] = struct{}{}

		countersCount++

		info, ok := c.annotationsCacheHCS[container.ID]
		if !ok {
			// Docker containers have no containerd bundle, so they keep empty Kubernetes labels.
			spec, _ := c.getContainerAnnotations(container.ID)
			info = spec.containerInfo(getContainerIdWithPrefix(container))
			c.annotationsCacheHCS[container.ID] = info
		}

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

	networkErr := c.collectNetworkMetrics(ch)

	// Remove containers that are no longer running
	for containerID := range c.annotationsCacheHCS {
		if _, ok := containerIDs[containerID]; !ok {
			delete(c.annotationsCacheHCS, containerID)
		}
	}

	if networkErr != nil {
		return fmt.Errorf("error in fetching container network statistics: %w", networkErr)
	}

	if len(collectErrors) > 0 {
		return fmt.Errorf("errors while fetching container statistics: %w", errors.Join(collectErrors...))
	}

	return nil
}

func (c *Collector) collectHCSContainer(ch chan<- prometheus.Metric, containerDetails hcs.Properties, containerInfo containerInfo) error {
	// Skip if the container is a pause container
	if containerInfo.pod != "" && containerInfo.container == "" {
		c.logger.Debug("skipping pause container",
			slog.String("container_id", containerDetails.ID),
			slog.String("container_name", containerInfo.container),
			slog.String("pod_name", containerInfo.pod),
			slog.String("namespace", containerInfo.namespace),
		)

		return nil
	}

	containerStats, err := hcs.GetContainerStatistics(containerDetails.ID)
	if err != nil {
		return fmt.Errorf("error fetching container statistics: %w", err)
	}

	ch <- prometheus.MustNewConstMetric(
		c.containerAvailable,
		prometheus.GaugeValue,
		1,
		containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container, "false",
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

// collectNetworkMetrics collects network metrics for containers.
func (c *Collector) collectNetworkMetrics(ch chan<- prometheus.Metric) error {
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
			containerInfo, ok := c.annotationsCacheHCS[containerId]

			if !ok {
				c.logger.Debug("Unknown container " + containerId + " for endpoint " + endpoint.ID)

				continue
			}

			// Skip if the container is a pause container
			if containerInfo.pod != "" && containerInfo.container == "" {
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
func (c *Collector) collectJobContainers(ch chan<- prometheus.Metric) error {
	// Each directory in the containerd state directory is the OCI bundle of a running task.
	entries, err := os.ReadDir(c.config.ContainerDStateDir)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("error in reading containerd state directory: %w", err)
		}

		if !c.stateDirMissingLogged {
			c.logger.Warn("containerd state directory does not exist",
				slog.String("path", c.config.ContainerDStateDir),
				slog.Any("err", err),
			)

			c.stateDirMissingLogged = true
		}

		clear(c.annotationsCacheJob)

		return nil
	}

	c.stateDirMissingLogged = false

	bundles := make(map[string]struct{}, len(entries))
	errs := make([]error, 0)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		containerID := entry.Name()
		bundles[containerID] = struct{}{}

		container, ok := c.annotationsCacheJob[containerID]
		if !ok {
			spec, err := c.getContainerAnnotations(containerID)
			if err != nil {
				// containerd may still be writing the bundle, so retry on the next scrape.
				c.logger.Debug("error in reading container annotations",
					slog.String("container_id", containerID),
					slog.Any("err", err),
				)

				continue
			}

			container = jobContainer{
				info:        spec.containerInfo("containerd://" + containerID),
				hostProcess: spec.Annotations["microsoft.com/hostprocess-container"] == "true",
			}

			c.annotationsCacheJob[containerID] = container
		}

		if !container.hostProcess {
			continue
		}

		if err := c.collectJobContainer(ch, containerID, container.info); err != nil {
			errs = append(errs, err)
		}
	}

	// Remove containers that are no longer running
	for containerID := range c.annotationsCacheJob {
		if _, ok := bundles[containerID]; !ok {
			delete(c.annotationsCacheJob, containerID)
		}
	}

	return errors.Join(errs...)
}

func (c *Collector) collectJobContainer(ch chan<- prometheus.Metric, containerID string, containerInfo containerInfo) error {
	jobObjectHandle, err := kernel32.OpenJobObject("Global\\JobContainer_" + containerID)
	if err != nil {
		if errors.Is(err, windows.ERROR_FILE_NOT_FOUND) {
			return nil
		}

		return fmt.Errorf("error in opening job object: %w", err)
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
		return fmt.Errorf("error in querying job object information: %w", err)
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
		return fmt.Errorf("error in querying job object memory usage information: %w", err)
	}

	ch <- prometheus.MustNewConstMetric(
		c.containerAvailable,
		prometheus.GaugeValue,
		1,
		containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container, "true",
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

	return nil
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

func (c *Collector) getContainerAnnotations(containerID string) (ociSpec, error) {
	// Close the file right away: the handle lacks FILE_SHARE_DELETE and would
	// block containerd from removing the bundle of a stopped container.
	configJSON, err := os.ReadFile(filepath.Join(c.config.ContainerDStateDir, containerID, "config.json"))
	if err != nil {
		return ociSpec{}, fmt.Errorf("error in reading config.json file: %w", err)
	}

	var annotations ociSpec

	if err = json.Unmarshal(configJSON, &annotations); err != nil {
		return ociSpec{}, fmt.Errorf("error in decoding config.json file: %w", err)
	}

	return annotations, nil
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

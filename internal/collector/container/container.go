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
	"slices"
	"strings"
	"unsafe"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-ole/go-ole"
	"github.com/prometheus-community/windows_exporter/internal/headers/hcn"
	"github.com/prometheus-community/windows_exporter/internal/headers/hcs"
	"github.com/prometheus-community/windows_exporter/internal/headers/iphlpapi"
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

	annotationsCacheHCS map[string]containerInfo
	annotationsCacheJob map[string]containerInfo

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

type ociSpec struct {
	Annotations map[string]string `json:"annotations"`
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
	c.annotationsCacheJob = make(map[string]containerInfo)

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
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
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
		ch <- prometheus.MustNewConstMetric(
			c.containersCount,
			prometheus.GaugeValue,
			0,
		)

		return nil
	}

	var countersCount float64

	containerIDs := make([]string, 0, len(containers))
	collectErrors := make([]error, 0)

	for _, container := range containers {
		if container.State != "Running" {
			continue
		}

		containerIDs = append(containerIDs, container.ID)

		countersCount++

		var (
			namespace     string
			podName       string
			containerName string
		)

		if _, ok := c.annotationsCacheHCS[container.ID]; !ok {
			if spec, err := c.getContainerAnnotations(container.ID); err == nil {
				namespace = spec.Annotations["io.kubernetes.cri.sandbox-namespace"]
				podName = spec.Annotations["io.kubernetes.cri.sandbox-name"]
				containerName = spec.Annotations["io.kubernetes.cri.container-name"]
			}

			c.annotationsCacheHCS[container.ID] = containerInfo{
				id:        getContainerIdWithPrefix(container),
				namespace: namespace,
				pod:       podName,
				container: containerName,
			}
		}

		if err = c.collectHCSContainer(ch, container, c.annotationsCacheHCS[container.ID]); err != nil {
			if errors.Is(err, hcs.ErrIDNotFound) {
				c.logger.Debug("err in fetching container statistics",
					slog.String("container_id", container.ID),
					slog.String("container_name", c.annotationsCacheHCS[container.ID].container),
					slog.String("container_pod_name", c.annotationsCacheHCS[container.ID].pod),
					slog.String("container_namespace", c.annotationsCacheHCS[container.ID].namespace),
					slog.Any("err", err),
				)
			} else {
				c.logger.Error("err in fetching container statistics",
					slog.String("container_id", container.ID),
					slog.String("container_name", c.annotationsCacheHCS[container.ID].container),
					slog.String("container_pod_name", c.annotationsCacheHCS[container.ID].pod),
					slog.String("container_namespace", c.annotationsCacheHCS[container.ID].namespace),
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

	if err = c.collectNetworkMetrics(ch); err != nil {
		return fmt.Errorf("error in fetching container network statistics: %w", err)
	}

	// Remove containers that are no longer running
	for _, containerID := range c.annotationsCacheHCS {
		if !slices.Contains(containerIDs, containerID.id) {
			delete(c.annotationsCacheHCS, containerID.id)
		}
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
	endpoints, err := hcn.EnumerateEndpoints()
	if err != nil {
		return fmt.Errorf("error in fetching HCN endpoints: %w", err)
	}

	if len(endpoints) == 0 {
		return nil
	}

	for _, endpoint := range endpoints {
		properties, err := hcn.GetEndpointProperties(endpoint)
		if err != nil {
			c.logger.Warn("Failed to collect properties for interface "+endpoint.String(),
				slog.Any("err", err),
			)

			continue
		}

		if len(properties.SharedContainers) == 0 {
			continue
		}

		var nicGUID *ole.GUID

		for _, allocator := range properties.Resources.Allocators {
			if allocator.AdapterNetCfgInstanceId != nil {
				nicGUID = allocator.AdapterNetCfgInstanceId

				break
			}
		}

		if nicGUID == nil {
			c.logger.Warn("Failed to get nic GUID for endpoint " + endpoint.String())

			continue
		}

		luid, err := iphlpapi.ConvertInterfaceGUIDToLUID(*nicGUID)
		if err != nil {
			return fmt.Errorf("error in converting interface GUID to LUID: %w", err)
		}

		var endpointStats iphlpapi.MIB_IF_ROW2
		endpointStats.InterfaceLuid = luid

		if err := iphlpapi.GetIfEntry2Ex(&endpointStats); err != nil {
			c.logger.Warn("Failed to get interface entry for endpoint "+endpoint.String(),
				slog.Any("err", err),
			)

			continue
		}

		for _, containerId := range properties.SharedContainers {
			containerInfo, ok := c.annotationsCacheHCS[containerId]

			if !ok {
				c.logger.Debug("Unknown container " + containerId + " for endpoint " + endpoint.String())

				continue
			}

			// Skip if the container is a pause container
			if containerInfo.pod != "" && containerInfo.container == "" {
				continue
			}

			endpointId := strings.ToUpper(endpoint.String())

			ch <- prometheus.MustNewConstMetric(
				c.bytesReceived,
				prometheus.CounterValue,
				float64(endpointStats.InOctets),
				containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container, endpointId,
			)

			ch <- prometheus.MustNewConstMetric(
				c.bytesSent,
				prometheus.CounterValue,
				float64(endpointStats.OutOctets),
				containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container, endpointId,
			)
			ch <- prometheus.MustNewConstMetric(
				c.packetsReceived,
				prometheus.CounterValue,
				float64(endpointStats.InUcastPkts+endpointStats.InNUcastPkts),
				containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container, endpointId,
			)
			ch <- prometheus.MustNewConstMetric(
				c.packetsSent,
				prometheus.CounterValue,
				float64(endpointStats.OutUcastPkts+endpointStats.OutNUcastPkts),
				containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container, endpointId,
			)
			ch <- prometheus.MustNewConstMetric(
				c.droppedPacketsIncoming,
				prometheus.CounterValue,
				float64(endpointStats.InDiscards+endpointStats.InErrors),
				containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container, endpointId,
			)
			ch <- prometheus.MustNewConstMetric(
				c.droppedPacketsOutgoing,
				prometheus.CounterValue,
				float64(endpointStats.OutDiscards+endpointStats.OutErrors),
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
	containerDStateFS := os.DirFS(c.config.ContainerDStateDir)

	allContainerIDs := make([]string, 0, len(c.annotationsCacheJob)+len(c.annotationsCacheHCS))
	jobContainerIDs := make([]string, 0, len(allContainerIDs))

	if err := fs.WalkDir(containerDStateFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				c.logger.Warn("containerd state directory does not exist",
					slog.String("path", c.config.ContainerDStateDir),
					slog.Any("err", err),
				)

				return nil
			}

			return err
		}

		if path == "." {
			return nil
		}

		if !d.IsDir() {
			return nil
		}

		if _, err := os.Stat(path + "\\config.json"); err != nil {
			containerID := strings.TrimPrefix(strings.Replace(path, c.config.ContainerDStateDir, "", 1), `\`)

			if spec, err := c.getContainerAnnotations(containerID); err == nil {
				isHostProcess, ok := spec.Annotations["microsoft.com/hostprocess-container"]
				if ok && isHostProcess == "true" {
					allContainerIDs = append(allContainerIDs, containerID)

					if _, ok := c.annotationsCacheJob[containerID]; !ok {
						var (
							namespace     string
							podName       string
							containerName string
						)

						namespace = spec.Annotations["io.kubernetes.cri.sandbox-namespace"]
						podName = spec.Annotations["io.kubernetes.cri.sandbox-name"]
						containerName = spec.Annotations["io.kubernetes.cri.container-name"]

						c.annotationsCacheJob[containerID] = containerInfo{
							id:        "containerd://" + containerID,
							namespace: namespace,
							pod:       podName,
							container: containerName,
						}
					}
				}
			}
		}

		// Skip the directory content
		return fs.SkipDir
	}); err != nil {
		return fmt.Errorf("error in walking containerd state directory: %w", err)
	}

	errs := make([]error, 0)

	for _, containerID := range allContainerIDs {
		if err := c.collectJobContainer(ch, containerID); err != nil {
			errs = append(errs, err)
		} else {
			jobContainerIDs = append(jobContainerIDs, containerID)
		}
	}

	// Remove containers that are no longer running
	for _, containerID := range c.annotationsCacheJob {
		if !slices.Contains(jobContainerIDs, containerID.id) {
			delete(c.annotationsCacheJob, containerID.id)
		}
	}

	return errors.Join(errs...)
}

func (c *Collector) collectJobContainer(ch chan<- prometheus.Metric, containerID string) error {
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

	privateWorkingSetBytes, err := calculatePrivateWorkingSetBytes(jobObjectHandle)
	if err != nil {
		c.logger.Debug("error in calculating private working set bytes", slog.Any("err", err))
	}

	containerInfo := c.annotationsCacheJob[containerID]

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
	ch <- prometheus.MustNewConstMetric(
		c.usagePrivateWorkingSetBytes,
		prometheus.GaugeValue,
		float64(privateWorkingSetBytes),

		containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
	)
	ch <- prometheus.MustNewConstMetric(
		c.runtimeTotal,
		prometheus.CounterValue,
		(float64(jobInfo.BasicInfo.ThisPeriodTotalKernelTime)+float64(jobInfo.BasicInfo.ThisPeriodTotalUserTime))*pdh.TicksToSecondScaleFactor,

		containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
	)
	ch <- prometheus.MustNewConstMetric(
		c.runtimeUser,
		prometheus.CounterValue,
		float64(jobInfo.BasicInfo.ThisPeriodTotalUserTime)*pdh.TicksToSecondScaleFactor,

		containerInfo.id, containerInfo.namespace, containerInfo.pod, containerInfo.container,
	)
	ch <- prometheus.MustNewConstMetric(
		c.runtimeKernel,
		prometheus.CounterValue,
		float64(jobInfo.BasicInfo.ThisPeriodTotalKernelTime)*pdh.TicksToSecondScaleFactor,

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
	configJSON, err := os.OpenFile(fmt.Sprintf(`%s%s\config.json`, c.config.ContainerDStateDir, containerID), os.O_RDONLY, 0)
	if err != nil {
		return ociSpec{}, fmt.Errorf("error in opening config.json file: %w", err)
	}

	var annotations ociSpec

	if err = json.NewDecoder(configJSON).Decode(&annotations); err != nil {
		return ociSpec{}, fmt.Errorf("error in decoding config.json file: %w", err)
	}

	return annotations, nil
}

func calculatePrivateWorkingSetBytes(jobObjectHandle windows.Handle) (uint64, error) {
	var pidList kernel32.JobObjectBasicProcessIDList

	retLen := uint32(unsafe.Sizeof(pidList))

	if err := windows.QueryInformationJobObject(
		jobObjectHandle,
		windows.JobObjectBasicProcessIdList,
		uintptr(unsafe.Pointer(&pidList)),
		retLen, &retLen); err != nil {
		return 0, err
	}

	var (
		privateWorkingSetBytes uint64
		vmCounters             kernel32.PROCESS_VM_COUNTERS
	)

	retLen = uint32(unsafe.Sizeof(vmCounters))

	getMemoryStats := func(pid uint32) (uint64, error) {
		processHandle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
		if err != nil {
			return 0, fmt.Errorf("error in opening process: %w", err)
		}

		defer func(fd windows.Handle) {
			_ = windows.Close(fd)
		}(processHandle)

		var isInJob bool

		if err := kernel32.IsProcessInJob(processHandle, jobObjectHandle, &isInJob); err != nil {
			return 0, fmt.Errorf("error in checking if process is in job: %w", err)
		}

		if !isInJob {
			return 0, nil
		}

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

	for _, pid := range pidList.PIDs() {
		privateWorkingSetSize, err := getMemoryStats(pid)
		if err != nil {
			return 0, fmt.Errorf("error in getting private working set bytes: %w", err)
		}

		privateWorkingSetBytes += privateWorkingSetSize
	}

	return privateWorkingSetBytes, nil
}

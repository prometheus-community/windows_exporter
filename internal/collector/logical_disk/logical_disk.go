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

package logical_disk

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/headers/fveapi"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows"
)

const (
	Name                  = "logical_disk"
	subCollectorMetrics   = "metrics"
	subCollectorBitlocker = "bitlocker_status"
)

// bitlockerStatusLabels lists the values of the status label of the bitlocker_status metric.
// The order and the values originate from the Shell property System.Volume.BitLockerProtection,
// which was used before fveapi.dll. The labels are kept to not break existing queries.
// "disabled" is never reported by fveapi.dll and always 0.
//
//nolint:gochecknoglobals
var bitlockerStatusLabels = []struct {
	state fveapi.State
	label string
}{
	{-1, "disabled"},
	{fveapi.StateOn, "on"},
	{fveapi.StateOff, "off"},
	{fveapi.StateEncrypting, "encrypting"},
	{fveapi.StateDecrypting, "decrypting"},
	{fveapi.StateSuspended, "suspended"},
	{fveapi.StateLocked, "locked"},
	{fveapi.StateUnknown, "unknown"},
	{fveapi.StateWaitingForActivation, "waiting_for_activation"},
}

type Config struct {
	CollectorsEnabled []string       `yaml:"enabled"`
	VolumeInclude     *regexp.Regexp `yaml:"volume-include"`
	VolumeExclude     *regexp.Regexp `yaml:"volume-exclude"`
}

//nolint:gochecknoglobals
var ConfigDefaults = Config{
	CollectorsEnabled: []string{
		subCollectorMetrics,
	},
	VolumeInclude: types.RegExpAny,
	VolumeExclude: types.RegExpEmpty,
}

// A Collector is a Prometheus Collector for perflib logicalDisk metrics.
type Collector struct {
	config Config
	logger *slog.Logger

	perfDataCollector *pdh.Collector[perfDataCounterValues]
	perfDataObject    []perfDataCounterValues

	metricsEnabled   bool
	bitlockerEnabled bool

	volumeInfo *volumeInfoReader
	// lastMountErr is the last logged error of getAllMountedVolumes.
	lastMountErr string

	avgReadQueue     *prometheus.Desc
	avgWriteQueue    *prometheus.Desc
	freeSpace        *prometheus.Desc
	idleTime         *prometheus.Desc
	information      *prometheus.Desc
	readBytesTotal   *prometheus.Desc
	readLatency      *prometheus.Desc
	readOnly         *prometheus.Desc
	readsTotal       *prometheus.Desc
	readTime         *prometheus.Desc
	readWriteLatency *prometheus.Desc
	requestsQueued   *prometheus.Desc
	splitIOs         *prometheus.Desc
	totalSpace       *prometheus.Desc
	writeBytesTotal  *prometheus.Desc
	writeLatency     *prometheus.Desc
	writesTotal      *prometheus.Desc
	writeTime        *prometheus.Desc

	bitlockerStatus *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	if config.VolumeExclude == nil {
		config.VolumeExclude = ConfigDefaults.VolumeExclude
	}

	if config.VolumeInclude == nil {
		config.VolumeInclude = ConfigDefaults.VolumeInclude
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

	var collectorsEnabled, volumeExclude, volumeInclude string

	app.Flag(
		"collector.logical_disk.volume-exclude",
		"Regexp of volumes to exclude. Volume name must both match include and not match exclude to be included.",
	).Default("").StringVar(&volumeExclude)

	app.Flag(
		"collector.logical_disk.volume-include",
		"Regexp of volumes to include. Volume name must both match include and not match exclude to be included.",
	).Default(".+").StringVar(&volumeInclude)

	app.Flag(
		"collector.logical_disk.enabled",
		fmt.Sprintf("Comma-separated list of collectors to use. Available collectors: %s, %s. Defaults to metrics, if not specified.",
			subCollectorMetrics,
			subCollectorBitlocker,
		),
	).Default(strings.Join(ConfigDefaults.CollectorsEnabled, ",")).StringVar(&collectorsEnabled)

	app.Action(func(*kingpin.ParseContext) error {
		c.config.CollectorsEnabled = strings.Split(collectorsEnabled, ",")

		var err error

		c.config.VolumeExclude, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", volumeExclude))
		if err != nil {
			return fmt.Errorf("collector.logical_disk.volume-exclude: %w", err)
		}

		c.config.VolumeInclude, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", volumeInclude))
		if err != nil {
			return fmt.Errorf("collector.logical_disk.volume-include: %w", err)
		}

		return nil
	})

	return c
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	c.perfDataCollector.Close()

	return nil
}

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	c.logger = logger.With(slog.String("collector", Name))

	for _, collector := range c.config.CollectorsEnabled {
		if !slices.Contains([]string{subCollectorMetrics, subCollectorBitlocker}, collector) {
			return fmt.Errorf("unknown sub collector: %s. Possible values: %s", collector,
				strings.Join([]string{subCollectorMetrics, subCollectorBitlocker}, ", "),
			)
		}
	}

	c.metricsEnabled = slices.Contains(c.config.CollectorsEnabled, subCollectorMetrics)
	c.bitlockerEnabled = slices.Contains(c.config.CollectorsEnabled, subCollectorBitlocker)
	c.volumeInfo = newVolumeInfoReader()

	c.information = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "info"),
		"A metric with a constant '1' value labeled with logical disk information",
		[]string{"disk", "type", "volume", "volume_name", "filesystem", "serial_number"},
		nil,
	)
	c.readOnly = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "readonly"),
		"Whether the logical disk is read-only",
		[]string{"volume"},
		nil,
	)
	c.requestsQueued = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "requests_queued"),
		"The number of requests queued to the disk (LogicalDisk.CurrentDiskQueueLength)",
		[]string{"volume"},
		nil,
	)

	c.avgReadQueue = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "avg_read_requests_queued"),
		"DEPRECATED: use rate(windows_logical_disk_read_seconds_total). "+
			"Cumulative read queue length in seconds, identical to windows_logical_disk_read_seconds_total (LogicalDisk.AvgDiskReadQueueLength)",
		[]string{"volume"},
		nil,
	)

	c.avgWriteQueue = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "avg_write_requests_queued"),
		"DEPRECATED: use rate(windows_logical_disk_write_seconds_total). "+
			"Cumulative write queue length in seconds, identical to windows_logical_disk_write_seconds_total (LogicalDisk.AvgDiskWriteQueueLength)",
		[]string{"volume"},
		nil,
	)

	c.readBytesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "read_bytes_total"),
		"The number of bytes transferred from the disk during read operations (LogicalDisk.DiskReadBytesPerSec)",
		[]string{"volume"},
		nil,
	)

	c.readsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "reads_total"),
		"The number of read operations on the disk (LogicalDisk.DiskReadsPerSec)",
		[]string{"volume"},
		nil,
	)

	c.writeBytesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "write_bytes_total"),
		"The number of bytes transferred to the disk during write operations (LogicalDisk.DiskWriteBytesPerSec)",
		[]string{"volume"},
		nil,
	)

	c.writesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "writes_total"),
		"The number of write operations on the disk (LogicalDisk.DiskWritesPerSec)",
		[]string{"volume"},
		nil,
	)

	c.readTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "read_seconds_total"),
		"Seconds that the disk was busy servicing read requests (LogicalDisk.PercentDiskReadTime)",
		[]string{"volume"},
		nil,
	)

	c.writeTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "write_seconds_total"),
		"Seconds that the disk was busy servicing write requests (LogicalDisk.PercentDiskWriteTime)",
		[]string{"volume"},
		nil,
	)

	c.freeSpace = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "free_bytes"),
		"Free space in bytes",
		[]string{"volume"},
		nil,
	)

	c.totalSpace = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "size_bytes"),
		"Total space in bytes",
		[]string{"volume"},
		nil,
	)

	c.idleTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "idle_seconds_total"),
		"Seconds that the disk was idle (LogicalDisk.PercentIdleTime)",
		[]string{"volume"},
		nil,
	)

	c.splitIOs = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "split_ios_total"),
		"The number of I/Os to the disk were split into multiple I/Os (LogicalDisk.SplitIOPerSec)",
		[]string{"volume"},
		nil,
	)

	c.readLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "read_latency_seconds_total"),
		"Shows the average time, in seconds, of a read operation from the disk (LogicalDisk.AvgDiskSecPerRead)",
		[]string{"volume"},
		nil,
	)

	c.writeLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "write_latency_seconds_total"),
		"Shows the average time, in seconds, of a write operation to the disk (LogicalDisk.AvgDiskSecPerWrite)",
		[]string{"volume"},
		nil,
	)

	c.readWriteLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "read_write_latency_seconds_total"),
		"Shows the time, in seconds, of the average disk transfer (LogicalDisk.AvgDiskSecPerTransfer)",
		[]string{"volume"},
		nil,
	)

	c.bitlockerStatus = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bitlocker_status"),
		"BitLocker status for the logical disk",
		[]string{"volume", "status"},
		nil,
	)

	var err error

	c.perfDataCollector, err = pdh.NewCollector[perfDataCounterValues](logger.With(slog.String("collector", Name)), pdh.CounterTypeRaw, "LogicalDisk", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create LogicalDisk collector: %w", err)
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric, _ time.Duration) error {
	err := c.perfDataCollector.Collect(&c.perfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect LogicalDisk metrics: %w", err)
	}

	// The performance counters don't depend on the mount points. Volumes that can't be
	// resolved fall back to their instance name.
	volumes, err := getAllMountedVolumes()
	c.logMountError(err)

	c.volumeInfo.beginScrape()
	defer c.volumeInfo.endScrape()

	for _, data := range c.perfDataObject {
		if c.config.VolumeExclude.MatchString(data.Name) || !c.config.VolumeInclude.MatchString(data.Name) {
			continue
		}

		info, err := c.volumeInfo.read(volumes, data.Name, c.metricsEnabled)
		if c.volumeInfo.shouldWarn(data.Name, err) {
			c.logger.Warn("failed to get volume information for "+data.Name,
				slog.Any("err", err),
			)
		} else if err != nil {
			c.logger.Debug("failed to get volume information for "+data.Name,
				slog.Any("err", err),
			)
		}

		if info.hasReadonly {
			ch <- prometheus.MustNewConstMetric(c.readOnly, prometheus.GaugeValue, info.readonly, data.Name)
		}

		ch <- prometheus.MustNewConstMetric(
			c.information,
			prometheus.GaugeValue,
			1,
			info.diskIDs,
			info.volumeType,
			data.Name,
			info.label,
			info.filesystem,
			info.serialNumber,
		)

		if c.metricsEnabled {
			// The Free Megabytes counter is cached by Windows and only updated every 10-15 minutes,
			// in whole megabytes. It's the fallback if the volume can't be queried directly.
			freeBytes, sizeBytes := data.FreeSpace*1024*1024, data.PercentFreeSpace*1024*1024
			if info.hasFreeBytes {
				freeBytes = info.freeBytes
			}

			if info.hasSizeBytes {
				sizeBytes = info.sizeBytes
			}

			ch <- prometheus.MustNewConstMetric(
				c.requestsQueued,
				prometheus.GaugeValue,
				data.CurrentDiskQueueLength,
				data.Name,
			)

			// The raw value of a PERF_COUNTER_100NS_QUEUELEN_TYPE counter is the queue length
			// integrated over time in 100ns ticks, not an average. It shares its source with
			// "% Disk Read/Write Time". Kept as gauge for compatibility: as a counter,
			// OpenMetrics exposition would rename the series to *_total.
			ch <- prometheus.MustNewConstMetric(
				c.avgReadQueue,
				prometheus.GaugeValue,
				data.AvgDiskReadQueueLength*pdh.TicksToSecondScaleFactor,
				data.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.avgWriteQueue,
				prometheus.GaugeValue,
				data.AvgDiskWriteQueueLength*pdh.TicksToSecondScaleFactor,
				data.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.readBytesTotal,
				prometheus.CounterValue,
				data.DiskReadBytesPerSec,
				data.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.readsTotal,
				prometheus.CounterValue,
				data.DiskReadsPerSec,
				data.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.writeBytesTotal,
				prometheus.CounterValue,
				data.DiskWriteBytesPerSec,
				data.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.writesTotal,
				prometheus.CounterValue,
				data.DiskWritesPerSec,
				data.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.readTime,
				prometheus.CounterValue,
				data.PercentDiskReadTime,
				data.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.writeTime,
				prometheus.CounterValue,
				data.PercentDiskWriteTime,
				data.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.freeSpace,
				prometheus.GaugeValue,
				freeBytes,
				data.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.totalSpace,
				prometheus.GaugeValue,
				sizeBytes,
				data.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.idleTime,
				prometheus.CounterValue,
				data.PercentIdleTime,
				data.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.splitIOs,
				prometheus.CounterValue,
				data.SplitIOPerSec,
				data.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.readLatency,
				prometheus.CounterValue,
				data.AvgDiskSecPerRead*pdh.TicksToSecondScaleFactor,
				data.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.writeLatency,
				prometheus.CounterValue,
				data.AvgDiskSecPerWrite*pdh.TicksToSecondScaleFactor,
				data.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.readWriteLatency,
				prometheus.CounterValue,
				data.AvgDiskSecPerTransfer*pdh.TicksToSecondScaleFactor,
				data.Name,
			)
		}

		if c.bitlockerEnabled {
			c.collectBitlocker(ch, volumes, data.Name, false)
		}
	}

	if c.bitlockerEnabled {
		// Locked BitLocker volumes have no LogicalDisk performance instance.
		seen := make(map[string]struct{}, len(c.perfDataObject))
		for _, data := range c.perfDataObject {
			seen[strings.ToUpper(data.Name)] = struct{}{}
		}

		for mountPoint := range volumes {
			if len(mountPoint) != 2 || mountPoint[1] != ':' {
				continue
			}

			if _, ok := seen[strings.ToUpper(mountPoint)]; ok {
				continue
			}

			if c.config.VolumeExclude.MatchString(mountPoint) || !c.config.VolumeInclude.MatchString(mountPoint) {
				continue
			}

			c.collectBitlocker(ch, volumes, mountPoint, true)
		}
	}

	return nil
}

// logMountError logs an error of getAllMountedVolumes. Repeats of the last error are logged at debug level.
func (c *Collector) logMountError(err error) {
	if err == nil {
		c.lastMountErr = ""

		return
	}

	level := slog.LevelWarn
	if msg := err.Error(); msg == c.lastMountErr {
		level = slog.LevelDebug
	} else {
		c.lastMountErr = msg
	}

	c.logger.Log(context.Background(), level, "failed to get the mount points of all volumes",
		slog.Any("err", err),
	)
}

// collectBitlocker emits the BitLocker status of a volume. With lockedOnly,
// volumes in any other state are skipped.
func (c *Collector) collectBitlocker(ch chan<- prometheus.Metric, volumes map[string]string, volume string, lockedOnly bool) {
	status, err := fveapi.GetStatus(getFveVolumeName(volumes, volume))
	if err != nil {
		level := slog.LevelWarn
		if lockedOnly {
			// Drive letters without performance data include empty media drives.
			level = slog.LevelDebug
		}

		c.logger.Log(context.Background(), level, "failed to get BitLocker status for "+volume,
			slog.Any("err", err),
		)

		return
	}

	state := status.State()

	if lockedOnly && state != fveapi.StateLocked {
		return
	}

	if state == fveapi.StateUnknown {
		c.logger.Debug("BitLocker status for "+volume+" is unknown",
			slog.String("flags", fmt.Sprintf("0x%08X", status.Flags)),
		)
	}

	for _, statusLabel := range bitlockerStatusLabels {
		val := 0.0
		if state == statusLabel.state {
			val = 1.0
		}

		ch <- prometheus.MustNewConstMetric(
			c.bitlockerStatus,
			prometheus.GaugeValue,
			val,
			volume,
			statusLabel.label,
		)
	}
}

// getFveVolumeName returns the name of a volume in the Win32 device namespace,
// as required by fveapi.dll.
func getFveVolumeName(volumes map[string]string, volume string) string {
	devicePath, _ := getVolumePaths(volumes, volume)

	return devicePath
}

// getAllMountedVolumes maps each mount point (e.g. C: or C:\mnt\data) to its volume GUID path.
// A volume whose mount points can't be read is skipped and reported in the returned error,
// together with the mount points of all other volumes.
func getAllMountedVolumes() (map[string]string, error) {
	guidBuf := make([]uint16, windows.MAX_PATH+1)
	guidBufLen := uint32(len(guidBuf))

	volumes := map[string]string{}

	hFindVolume, err := windows.FindFirstVolume(&guidBuf[0], guidBufLen)
	if err != nil {
		return volumes, fmt.Errorf("FindFirstVolume: %w", err)
	}

	defer func() {
		_ = windows.FindVolumeClose(hFindVolume)
	}()

	var errs []error

	for ; ; err = windows.FindNextVolume(hFindVolume, &guidBuf[0], guidBufLen) {
		if err != nil {
			if !errors.Is(err, windows.ERROR_NO_MORE_FILES) {
				errs = append(errs, fmt.Errorf("FindNextVolume: %w", err))
			}

			return volumes, errors.Join(errs...)
		}

		mountPoints, err := getVolumeMountPaths(&guidBuf[0], windows.GetVolumePathNamesForVolumeName)
		if err != nil {
			errs = append(errs, fmt.Errorf("GetVolumePathNamesForVolumeName(%s): %w", windows.UTF16ToString(guidBuf), err))

			continue
		}

		volumeName := strings.TrimSuffix(windows.UTF16ToString(guidBuf), `\`)
		for _, mountPoint := range mountPoints {
			volumes[strings.TrimSuffix(mountPoint, `\`)] = volumeName
		}
	}
}

// getVolumeMountPaths reads the MULTI_SZ mount paths. The API measures both
// buffer capacity and required capacity in UTF-16 characters, not bytes.
func getVolumeMountPaths(volumeName *uint16, query func(*uint16, *uint16, uint32, *uint32) error) ([]string, error) {
	buf := make([]uint16, windows.MAX_PATH+1)

	for {
		var required uint32

		err := query(volumeName, &buf[0], uint32(len(buf)), &required)
		if errors.Is(err, windows.ERROR_FILE_NOT_FOUND) {
			return nil, nil
		}

		if errors.Is(err, windows.ERROR_MORE_DATA) {
			if required <= uint32(len(buf)) {
				return nil, fmt.Errorf("invalid required mount path buffer size: %d", required)
			}

			buf = make([]uint16, required)

			continue
		}

		if err != nil {
			return nil, err
		}

		var paths []string

		for len(buf) > 0 && buf[0] != 0 {
			end := slices.Index(buf, uint16(0))
			if end < 0 {
				return nil, errors.New("unterminated mount path")
			}

			paths = append(paths, windows.UTF16ToString(buf[:end]))
			buf = buf[end+1:]
		}

		return paths, nil
	}
}
